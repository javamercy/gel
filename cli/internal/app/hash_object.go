package app

import (
	"Gel/internal/domain"
	"Gel/internal/storage"
	"fmt"
	"os"
)

type HashObjectInput struct {
	Paths []domain.AbsolutePath
	Write bool
}
type HashObjectResult struct {
	Hashes []domain.Hash
}

type HashObject struct {
	objectStore *storage.ObjectStore
}

func NewHashObject(objectStore *storage.ObjectStore) *HashObject {
	return &HashObject{
		objectStore: objectStore,
	}
}

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
