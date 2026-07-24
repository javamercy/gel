package storage

import (
	"Gel/internal/domain"
	"compress/zlib"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

const (
	objectDirPermission  fs.FileMode = 0o755
	objectFilePermission fs.FileMode = 0o444
)

// ObjectStore persists content-addressed Gel objects.
type ObjectStore struct {
	objectsDir domain.AbsolutePath
}

// NewObjectStore creates an object store rooted at the .gel/objects directory.
func NewObjectStore(objectsDir domain.AbsolutePath) *ObjectStore {
	return &ObjectStore{
		objectsDir: objectsDir,
	}
}

// Write stores the given content-addressed object and returns its SHA-256 hash or an error if the operation fails.
func (s *ObjectStore) Write(object domain.Object) (domain.Hash, error) {
	encoded, err := domain.EncodeObject(object)
	if err != nil {
		return domain.Hash{}, fmt.Errorf(
			"encode object: %w",
			err,
		)
	}

	hash := domain.ComputeObjectHash(encoded)
	objectDir, objectPath, err := s.objectPaths(hash)
	if err != nil {
		return domain.Hash{}, fmt.Errorf(
			"resolve path for object %s: %w",
			hash,
			err,
		)
	}
	if err := os.MkdirAll(objectDir.String(), objectDirPermission); err != nil {
		return domain.Hash{}, fmt.Errorf(
			"create object directory %q: %w",
			objectDir,
			err,
		)
	}
	if err := writeObjectFile(objectPath.String(), encoded); err != nil {
		return domain.Hash{}, fmt.Errorf(
			"write object %s: %w",
			hash,
			err,
		)
	}
	return hash, nil
}

// Read retrieves and decodes a content-addressed object from storage using its hash.
//
// It ensures the object hash matches the provided hash and returns an error on mismatch or failure.
func (s *ObjectStore) Read(hash domain.Hash) (domain.Object, error) {
	_, objectPath, err := s.objectPaths(hash)
	if err != nil {
		return nil, fmt.Errorf(
			"resolve path for object %s: %w",
			hash,
			err,
		)
	}

	encoded, err := readCompressedObject(objectPath.String())
	if err != nil {
		return nil, fmt.Errorf(
			"read object %s from %q: %w",
			hash,
			objectPath,
			err,
		)
	}

	actualHash := domain.ComputeObjectHash(encoded)
	if actualHash != hash {
		return nil, fmt.Errorf(
			"object hash mismatch: requested=%s actual=%s",
			hash,
			actualHash,
		)
	}

	object, err := domain.DecodeObject(encoded)
	if err != nil {
		return nil, fmt.Errorf(
			"decode object %s: %w",
			hash,
			err,
		)
	}
	return object, nil
}

func (s *ObjectStore) objectPaths(hash domain.Hash) (
	objectDir domain.AbsolutePath,
	objectPath domain.AbsolutePath,
	err error,
) {
	hexHash := hash.Hex()
	objectDir, err = s.objectsDir.Join(hexHash[:2])
	if err != nil {
		return domain.AbsolutePath{}, domain.AbsolutePath{}, err
	}

	objectPath, err = objectDir.Join(hexHash[2:])
	if err != nil {
		return domain.AbsolutePath{}, domain.AbsolutePath{}, err
	}
	return objectDir, objectPath, nil
}

func writeObjectFile(path string, encoded []byte) (returnErr error) {
	dir := filepath.Dir(path)
	file, err := os.CreateTemp(dir, ".object.tmp-*")
	if err != nil {
		return fmt.Errorf(
			"create temporary object in %q: %w",
			dir,
			err,
		)
	}

	tempPath := file.Name()
	fileClosed := false
	compressor := zlib.NewWriter(file)
	compressorClosed := false

	defer func() {
		if !compressorClosed {
			if err := compressor.Close(); err != nil {
				returnErr = errors.Join(
					returnErr,
					fmt.Errorf(
						"close object compressor: %w",
						err,
					),
				)
			}
		}
		if !fileClosed {
			if err := file.Close(); err != nil {
				returnErr = errors.Join(
					returnErr,
					fmt.Errorf(
						"close temporary object %q: %w",
						tempPath,
						err,
					),
				)
			}
		}
		if err := os.Remove(tempPath); err != nil &&
			!errors.Is(err, os.ErrNotExist) {
			returnErr = errors.Join(
				returnErr,
				fmt.Errorf(
					"remove temporary object %q: %w",
					tempPath,
					err,
				),
			)
		}
	}()

	if _, err := compressor.Write(encoded); err != nil {
		return fmt.Errorf(
			"compress object: %w",
			err,
		)
	}

	err = compressor.Close()
	compressorClosed = true
	if err != nil {
		return fmt.Errorf(
			"finish object compression: %w",
			err,
		)
	}
	if err := file.Chmod(objectFilePermission); err != nil {
		return fmt.Errorf(
			"set permissions on temporary object %q: %w",
			tempPath,
			err,
		)
	}
	if err := file.Sync(); err != nil {
		return fmt.Errorf(
			"sync temporary object %q: %w",
			tempPath,
			err,
		)
	}

	err = file.Close()
	fileClosed = true
	if err != nil {
		return fmt.Errorf(
			"close temporary object %q: %w",
			tempPath,
			err,
		)
	}
	if err := os.Link(tempPath, path); err != nil {
		if errors.Is(err, fs.ErrExist) {
			info, statErr := os.Lstat(path)
			if statErr != nil {
				return fmt.Errorf(
					"inspect existing object %q: %w",
					path,
					statErr,
				)
			}
			if !info.Mode().IsRegular() {
				return fmt.Errorf(
					"object path %q exists but is not a regular file",
					path,
				)
			}
			return nil
		}
		return fmt.Errorf(
			"link object %q: %w",
			path,
			err,
		)
	}
	return nil
}

func readCompressedObject(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf(
			"open compressed object: %w",
			err,
		)
	}

	defer func() {
		_ = file.Close()
	}()

	decompressor, err := zlib.NewReader(file)
	if err != nil {
		return nil, fmt.Errorf(
			"initialize object decompressor: %w",
			err,
		)
	}

	defer func() {
		_ = decompressor.Close()
	}()

	// BUG: io.ReadAll has no size limit, which can lead to excessive memory usage if the object is very large.
	// Consider using a size-limited reader or streaming the data instead.
	encoded, err := io.ReadAll(decompressor)
	if err != nil {
		return nil, fmt.Errorf(
			"decompress object: %w",
			err,
		)
	}
	return encoded, nil
}
