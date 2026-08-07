package app

import (
	"Gel/internal/domain"
	"Gel/internal/storage"
	"bytes"
	"errors"
	"fmt"
)

// CatFileInput identifies the object to read.
type CatFileInput struct {
	// Hash is the SHA-256 identifier of the object to read.
	Hash domain.Hash
}

// CatFileResult contains the decoded object and its canonical payload body.
type CatFileResult struct {
	// Object is the decoded object identified by the input hash.
	Object domain.Object
	// Body is the object's payload without its type and size header.
	Body []byte
}

// CatFile reads objects from the Gel object store.
type CatFile struct {
	objectStore *storage.ObjectStore
}

// NewCatFile returns a CatFile that reads from objectStore.
//
// objectStore must not be nil when the returned CatFile is used.
func NewCatFile(objectStore *storage.ObjectStore) *CatFile {
	return &CatFile{
		objectStore: objectStore,
	}
}

// Run reads and decodes the object identified by input.Hash.
//
// It returns an error when input.Hash is zero or the object cannot be read or
// encoded. The returned Body contains the object payload without its canonical
// type and size header.
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
