package app

import (
	"Gel/internal/domain"
	"Gel/internal/storage"
	"fmt"
	"os"
)

// HashObjectInput specifies the files to hash and whether to store their
// resulting blob objects.
type HashObjectInput struct {
	// Paths are processed in order, with one hash returned for each path.
	Paths []domain.AbsolutePath
	// Write reports whether each blob should be persisted in the object store.
	Write bool
}

// HashObjectResult contains hashes for the files processed by HashObject.Run.
type HashObjectResult struct {
	// Hashes contains one hash per successfully processed path, in input order.
	Hashes []domain.Hash
}

// HashObject hashes files as Gel blob objects.
type HashObject struct {
	objectStore *storage.ObjectStore
}

// NewHashObject returns a HashObject configured to use objectStore when
// writing blobs.
//
// objectStore must not be nil when Run is called with HashObjectInput.Write
// set to true.
func NewHashObject(objectStore *storage.ObjectStore) *HashObject {
	return &HashObject{
		objectStore: objectStore,
	}
}

// Run reads each input path, hashes its contents as a Gel blob, and returns
// the hashes in input order.
//
// When input.Write is true, Run also stores each blob in the configured object
// store. When input.Write is false, it does not modify object storage. If an
// error occurs, the result contains hashes for paths processed before the
// failing path.
func (ho *HashObject) Run(input HashObjectInput) (HashObjectResult, error) {
	var result HashObjectResult

	for _, path := range input.Paths {
		data, err := os.ReadFile(path.String())
		if err != nil {
			return result, fmt.Errorf(
				"read file %s: %w",
				path,
				err,
			)
		}

		blob := domain.NewBlob(data)
		var hash domain.Hash

		if input.Write {
			hash, err = ho.objectStore.Write(blob)
			if err != nil {
				return result, fmt.Errorf(
					"write object %q: %w",
					path.String(),
					err,
				)
			}
		} else {
			encoded, err := domain.EncodeObject(blob)
			if err != nil {
				return result, fmt.Errorf(
					"encode object %q: %w",
					path.String(),
					err,
				)
			}

			hash = domain.ComputeObjectHash(encoded)
		}

		result.Hashes = append(result.Hashes, hash)
	}
	return result, nil
}
