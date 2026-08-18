package app

import (
	"Gel/internal/domain"
	"Gel/internal/storage"
	"fmt"
)

// WriteTreeResult contains the hash of the tree written from the index.
type WriteTreeResult struct {
	// Hash is the object identifier of the root tree.
	Hash domain.Hash
}

// WriteTree writes the staged index as a hierarchy of tree objects.
type WriteTree struct {
	indexStore  *storage.IndexStore
	objectStore *storage.ObjectStore
}

// NewWriteTree returns a WriteTree backed by indexStore and objectStore.
//
// Both stores must not be nil when Run is called.
func NewWriteTree(
	indexStore *storage.IndexStore,
	objectStore *storage.ObjectStore,
) *WriteTree {
	return &WriteTree{
		indexStore:  indexStore,
		objectStore: objectStore,
	}
}

// Run loads the index, writes its tree hierarchy to object storage, and
// returns the root tree hash.
//
// The index must contain only normal-stage entries. An index with no entries
// produces and stores an empty root tree. If loading, validation, tree
// construction, or object storage fails, Run returns an error and a zero
// result.
func (w *WriteTree) Run() (WriteTreeResult, error) {
	index, err := w.indexStore.Load()
	if err != nil {
		return WriteTreeResult{}, fmt.Errorf(
			"load index: %w",
			err,
		)
	}

	rootHash, err := writeIndexTree(
		index.Entries(),
		w.objectStore,
	)
	if err != nil {
		return WriteTreeResult{}, fmt.Errorf(
			"write index tree: %w",
			err,
		)
	}
	return WriteTreeResult{
		Hash: rootHash,
	}, nil
}
