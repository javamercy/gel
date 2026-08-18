package app

import (
	"Gel/internal/domain"
	"Gel/internal/storage"
	"fmt"
	"strings"
)

type indexTreeFile struct {
	mode domain.FileMode
	hash domain.Hash
	name string
}

type indexTreeNode struct {
	name     string
	children map[string]*indexTreeNode
	files    []indexTreeFile
}

func writeIndexTree(
	entries []domain.IndexEntry,
	objectStore *storage.ObjectStore,
) (domain.Hash, error) {
	for _, entry := range entries {
		if entry.Stage() != domain.IndexStageNormal {
			return domain.Hash{}, fmt.Errorf(
				"cannot write tree: path %q has merge stage %d",
				entry.Path(),
				entry.Stage(),
			)
		}
	}

	return writeIndexTreeNode(
		buildIndexTree(entries),
		objectStore,
	)
}

func buildIndexTree(entries []domain.IndexEntry) *indexTreeNode {
	root := &indexTreeNode{
		children: make(map[string]*indexTreeNode),
	}

	for _, entry := range entries {
		current := root
		names := strings.Split(entry.Path().String(), "/")

		for _, name := range names[:len(names)-1] {
			child := current.children[name]
			if child == nil {
				child = &indexTreeNode{
					name:     name,
					children: make(map[string]*indexTreeNode),
				}
				current.children[name] = child
			}
			current = child
		}
		current.files = append(
			current.files, indexTreeFile{
				mode: entry.Mode(),
				hash: entry.Hash(),
				name: names[len(names)-1],
			},
		)
	}
	return root
}

func writeIndexTreeNode(root *indexTreeNode, objectStore *storage.ObjectStore) (domain.Hash, error) {
	var entries []domain.TreeEntry

	for _, childTree := range root.children {
		childHash, err := writeIndexTreeNode(childTree, objectStore)
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

	for _, file := range root.files {
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

	hash, err := objectStore.Write(tree)
	if err != nil {
		return domain.Hash{}, fmt.Errorf(
			"write tree object: %w",
			err,
		)
	}
	return hash, nil
}
