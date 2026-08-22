package app

import (
	"Gel/internal/domain"
	"Gel/internal/storage"
	"fmt"
)

// ReadTreeInput specifies the tree to load into the index.
type ReadTreeInput struct {
	// Hash identifies the root tree object and must not be zero unless Empty is
	// true.
	Hash domain.Hash
	// Empty reports whether to replace the index with an empty index. When true,
	// Hash is ignored.
	Empty bool
}

// ReadTree replaces the repository index with the contents of a tree object.

type ReadTree struct {
	indexStore  *storage.IndexStore
	objectStore *storage.ObjectStore
}

// NewReadTree returns a ReadTree backed by indexStore and objectStore.
//
// Both stores must not be nil when Run is called.
func NewReadTree(
	indexStore *storage.IndexStore,
	objectStore *storage.ObjectStore,
) *ReadTree {
	return &ReadTree{
		indexStore:  indexStore,
		objectStore: objectStore,
	}
}

// Run replaces the index with the flattened content of input.Hash, or with an
// empty index when input.Empty is true.
//
// When not empty, Run reads input.Hash and verifies it is a tree, recursing into
// each subtree and verifying it is also a tree. Blob entries are flattened into
// normal-stage index entries preserving their repository-relative paths. The
// flattened entries are canonicalized, sorted, and atomically written to
// storage, replacing any existing index contents. Blob file modes are preserved,
// but other stat fields are not. Any read, validation, or write error returns
// without saving a new index.
func (r *ReadTree) Run(input ReadTreeInput) error {
	var entries []domain.IndexEntry

	if !input.Empty {
		if input.Hash.IsZero() {
			return fmt.Errorf("tree hash cannot be zero")
		}

		tree, err := r.objectStore.ReadAs[*domain.Tree](input.Hash)
		if err != nil {
			return fmt.Errorf(
				"read tree object: %w",
				err,
			)
		}

		entries, err = r.flattenTreeRecursive(tree, nil, "")
		if err != nil {
			return fmt.Errorf(
				"flatten tree %s: %w",
				input.Hash,
				err,
			)
		}
	}

	index, err := domain.NewIndexFromEntries(entries)
	if err != nil {
		return fmt.Errorf(
			"create index: %w",
			err,
		)
	}

	if err := r.indexStore.Save(index); err != nil {
		return fmt.Errorf(
			"save index %w",
			err,
		)
	}
	return nil
}

func (r *ReadTree) flattenTreeRecursive(
	tree *domain.Tree,
	entries []domain.IndexEntry,
	prefix string,
) ([]domain.IndexEntry, error) {
	for _, treeEntry := range tree.Entries() {
		pathText := treeEntry.Name()
		if prefix != "" {
			pathText = prefix + "/" + treeEntry.Name()
		}

		if treeEntry.Mode().IsDirectory() {
			childTree, err := r.objectStore.ReadAs[*domain.Tree](treeEntry.Hash())
			if err != nil {
				return nil, fmt.Errorf(
					"read child tree %s: %w",
					pathText,
					err,
				)
			}

			entries, err = r.flattenTreeRecursive(childTree, entries, pathText)
			if err != nil {
				return nil, err
			}
			continue
		}

		path, err := domain.ParseNormalizedPath(pathText)
		if err != nil {
			return nil, fmt.Errorf(
				"parse tree path %q: %w",
				pathText,
				err,
			)
		}

		entry, err := domain.NewIndexEntryFromFileStat(
			path,
			treeEntry.Hash(),
			domain.FileStat{Mode: treeEntry.Mode()},
		)
		if err != nil {
			return nil, fmt.Errorf(
				"create index entry for %q: %w",
				pathText,
				err,
			)
		}
		entries = append(entries, entry)
	}
	return entries, nil
}
