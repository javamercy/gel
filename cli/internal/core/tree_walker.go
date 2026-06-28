package core

import (
	"Gel/internal/domain"
	"path"
)

// Processor is invoked for each traversed entry that matches walk options.
type Processor = func(entry domain.TreeEntry, relPath string) error

// WalkOptions controls which entries are yielded during traversal.
type WalkOptions struct {
	// Recursive descends into child trees.
	Recursive bool
	// IncludeTrees yields tree entries in addition to non-tree entries.
	IncludeTrees bool
	// OnlyTrees yields only tree entries.
	OnlyTrees bool
}

// TreeWalker traverses tree objects depth-first.
type TreeWalker struct {
	objectService *ObjectService
	options       WalkOptions
}

// NewTreeWalker creates a tree walker for the given options.
func NewTreeWalker(objectService *ObjectService, options WalkOptions) *TreeWalker {
	return &TreeWalker{
		objectService: objectService,
		options:       options,
	}
}

// Walk traverses the tree identified by hash and calls processor for matching entries.
// prefix is joined with each entry name to produce a repository-relative path.
func (w *TreeWalker) Walk(hash domain.Hash, prefix string, processor Processor) error {
	tree, err := w.objectService.ReadTree(hash)
	if err != nil {
		return err
	}

	for _, entry := range tree.Entries() {
		relPath := path.Join(prefix, entry.Name())
		shouldProcess := w.shouldProcess(entry)
		if shouldProcess {
			if err := processor(entry, relPath); err != nil {
				return err
			}
		}
		if entry.Mode().IsDirectory() && w.options.Recursive {
			if err := w.Walk(entry.Hash(), relPath, processor); err != nil {
				return err
			}
		}
	}
	return nil
}

// shouldProcess determines if the given entry, based on its type, should be processed according to the walk options.
func (w *TreeWalker) shouldProcess(entry domain.TreeEntry) bool {
	isTree := entry.Mode().IsDirectory()
	if isTree {
		return w.options.IncludeTrees || w.options.OnlyTrees
	}
	return !w.options.OnlyTrees
}
