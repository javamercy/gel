package app

import (
	"Gel/internal/domain"
	"Gel/internal/storage"
	"bytes"
	"errors"
	"fmt"
)

type CatFileInput struct {
	Hash domain.Hash
}

type CatFileResult struct {
	Object domain.Object
	Body   []byte
}

type CatFile struct {
	objectStore *storage.ObjectStore
}

func NewCatFile(objectStore *storage.ObjectStore) *CatFile {
	return &CatFile{
		objectStore: objectStore,
	}
}

func (c *CatFile) Run(input CatFileInput) (CatFileResult, error) {
	var result CatFileResult

	if input.Hash.IsZero() {
		return result, errors.New("hash is zero")
	}

	object, err := c.objectStore.Read(input.Hash)
	if err != nil {
		return result, fmt.Errorf(
			"read object %q: %w",
			input.Hash,
			err,
		)
	}

	// TODO: encoding step does not look clean!
	encoded, err := domain.EncodeObject(object)
	if err != nil {
		return result, fmt.Errorf(
			"encode object %q: %w",
			input.Hash,
			err,
		)
	}

	_, body, _ := bytes.Cut(encoded, []byte{0})
	return CatFileResult{Object: object, Body: body}, nil
}
