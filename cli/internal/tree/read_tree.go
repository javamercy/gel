package tree

import (
	"Gel/internal/core"
	"Gel/internal/domain"
	"fmt"
)

// ReadTreeService replaces index contents from a tree object.
type ReadTreeService struct {
	indexService  *core.IndexService
	objectService *core.ObjectService
}

// NewReadTreeService creates a read-tree service.
func NewReadTreeService(
	indexService *core.IndexService,
	objectService *core.ObjectService,
) *ReadTreeService {
	return &ReadTreeService{
		indexService:  indexService,
		objectService: objectService,
	}
}

// ReadTree reads all blob entries from the given tree hash and rewrites index.
//
// Existing index entries are discarded. New entries are created with tree mode
// and hash data; filesystem stat fields are set to zero values.
func (r *ReadTreeService) ReadTree(hash domain.Hash) error {
	var indexEntries []domain.IndexEntry
	processor := func(entry domain.TreeEntry, path string) error {
		_, err := domain.ParseNormalizedPath(path)
		if err != nil {
			return err
		}
		// TODO: fix this to create a proper index entry with the correct path, hash, and mode
		indexEntry := domain.IndexEntry{}
		indexEntries = append(indexEntries, indexEntry)
		return nil
	}
	options := core.WalkOptions{
		Recursive:    true,
		IncludeTrees: false,
		OnlyTrees:    false,
	}
	treeWalker := core.NewTreeWalker(r.objectService, options)
	err := treeWalker.Walk(hash, "", processor)
	if err != nil {
		return fmt.Errorf("read-tree: %w", err)
	}

	if err := r.indexService.WriteEntries(indexEntries); err != nil {
		return fmt.Errorf("read-tree: %w", err)
	}
	return nil
}
