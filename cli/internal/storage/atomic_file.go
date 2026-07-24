package storage

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

func replaceFileAtomically(
	path string,
	data []byte,
	permission fs.FileMode,
) (returnErr error) {
	dir := filepath.Dir(path)
	pattern := "." + filepath.Base(path) + ".temp-*"
	file, err := os.CreateTemp(dir, pattern)
	if err != nil {
		return fmt.Errorf(
			"create temporary file in %q: %w",
			dir,
			err,
		)
	}

	tempPath := file.Name()
	closed := false
	committed := false

	defer func() {
		if !closed {
			if err := file.Close(); err != nil {
				returnErr = errors.Join(
					returnErr, fmt.Errorf(
						"close temporary file %q: %w",
						tempPath,
						err,
					),
				)
			}
		}
		if !committed {
			if err := os.Remove(tempPath); err != nil &&
				!errors.Is(err, os.ErrNotExist) {
				returnErr = errors.Join(
					returnErr, fmt.Errorf(
						"remove temporary file %q: %w",
						tempPath,
						err,
					),
				)
			}
		}
	}()

	if _, err := file.Write(data); err != nil {
		return fmt.Errorf(
			"write temporary file %q: %w",
			tempPath,
			err,
		)
	}
	if err := file.Chmod(permission); err != nil {
		return fmt.Errorf(
			"set permissions on temporary file %q: %w",
			tempPath,
			err,
		)
	}
	if err := file.Sync(); err != nil {
		return fmt.Errorf(
			"sync temporary file %q: %w",
			tempPath,
			err,
		)
	}

	err = file.Close()
	closed = true
	if err != nil {
		return fmt.Errorf(
			"close temporary file %q: %w",
			tempPath,
			err,
		)
	}
	if err := os.Rename(tempPath, path); err != nil {
		return fmt.Errorf(
			"replace %q: %w",
			path,
			err,
		)
	}

	committed = true
	return nil
}
