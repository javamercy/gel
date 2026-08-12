package app

import (
	"Gel/internal/domain"
	"Gel/internal/storage"
	"fmt"
	"strings"
)

type fileNode struct {
	mode domain.FileMode
	hash domain.Hash
	name string
}

type treeNode struct {
	name     string
	children map[string]*treeNode
	files    []fileNode
}

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

	entries := index.Entries()
	for _, entry := range entries {
		if entry.Stage() != domain.IndexStageNormal {
			return WriteTreeResult{}, fmt.Errorf(
				"cannot write tree: path %q has merge stage %d",
				entry.Path(),
				entry.Stage(),
			)
		}
	}

	rootTree := buildRootTree(entries)
	rootHash, err := w.writeTreeRecursive(rootTree)
	if err != nil {
		return WriteTreeResult{}, fmt.Errorf(
			"write tree: %w",
			err,
		)
	}
	return WriteTreeResult{
		Hash: rootHash,
	}, nil
}

func (w *WriteTree) writeTreeRecursive(node *treeNode) (domain.Hash, error) {
	var entries []domain.TreeEntry

	for _, childTree := range node.children {
		childHash, err := w.writeTreeRecursive(childTree)
		if err != nil {
			return domain.Hash{}, fmt.Errorf(
				"write child tree %s: %w",
				childTree.name,
				err,
			)
		}

		entry, err := domain.NewTreeEntry(
			domain.FileModeDirectory,
			childHash,
			childTree.name,
		)
		if err != nil {
			return domain.Hash{}, fmt.Errorf(
				"create tree entry for child tree %s: %w",
				childTree.name,
				err,
			)
		}

		entries = append(entries, entry)
	}

	for _, file := range node.files {
		entry, err := domain.NewTreeEntry(
			file.mode,
			file.hash,
			file.name,
		)
		if err != nil {
			return domain.Hash{}, fmt.Errorf(
				"create tree entry for file %s: %w",
				file.name,
				err,
			)
		}

		entries = append(entries, entry)
	}

	tree, err := domain.NewTreeFromEntries(entries)
	if err != nil {
		return domain.Hash{}, fmt.Errorf(
			"create tree from entries: %w",
			err,
		)
	}

	hash, err := w.objectStore.Write(tree)
	if err != nil {
		return domain.Hash{}, fmt.Errorf(
			"write tree object: %w",
			err,
		)
	}
	return hash, nil
}

func buildRootTree(entries []domain.IndexEntry) *treeNode {
	root := &treeNode{
		children: make(map[string]*treeNode),
	}

	for _, entry := range entries {
		current := root
		names := strings.Split(entry.Path().String(), "/")

		for _, name := range names[:len(names)-1] {
			child := current.children[name]
			if child == nil {
				child = &treeNode{
					name:     name,
					children: make(map[string]*treeNode),
				}
				current.children[name] = child
			}
			current = child
		}
		current.files = append(
			current.files, fileNode{
				mode: entry.Mode(),
				hash: entry.Hash(),
				name: names[len(names)-1],
			},
		)
	}
	return root
}
