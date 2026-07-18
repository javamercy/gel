package tree

import (
	"Gel/internal/core"
	"Gel/internal/domain"
	"strings"
)

// fileNode represents one staged file leaf in a directory tree.
type fileNode struct {
	mode domain.FileMode
	hash domain.Hash
	name string
}

// directoryNode represents one directory in the temporary tree builder graph.
type directoryNode struct {
	name     string
	children map[string]*directoryNode
	files    []*fileNode
}

// WriteTreeService writes tree objects from the current index snapshot.
type WriteTreeService struct {
	indexService  *core.IndexService
	objectService *core.ObjectService
}

// NewWriteTreeService creates a write-tree service.
func NewWriteTreeService(indexService *core.IndexService, objectService *core.ObjectService) *WriteTreeService {
	return &WriteTreeService{
		indexService:  indexService,
		objectService: objectService,
	}
}

// WriteTree converts all current index entries into a root tree object.
// It returns the root tree hash and writes any missing tree objects to storage.
func (w *WriteTreeService) WriteTree() (domain.Hash, error) {
	entries, err := w.indexService.GetEntries()
	if err != nil {
		return domain.Hash{}, err
	}

	root := buildRootTree(entries)
	rootHash, err := w.writeTreeRecursive(root)
	if err != nil {
		return domain.Hash{}, err
	}
	return rootHash, nil
}

// writeTreeRecursive materializes tree objects for a directory subtree.
// Child trees are written first so parent tree entries can reference their hashes.
func (w *WriteTreeService) writeTreeRecursive(root *directoryNode) (domain.Hash, error) {
	var entries []domain.TreeEntry

	for _, childDir := range root.children {
		subTreeHash, err := w.writeTreeRecursive(childDir)
		if err != nil {
			return domain.Hash{}, err
		}

		entry, err := domain.NewTreeEntry(domain.FileModeDirectory, subTreeHash, childDir.name)
		if err != nil {
			return domain.Hash{}, err
		}
		entries = append(entries, entry)
	}
	for _, file := range root.files {
		entry, err := domain.NewTreeEntry(file.mode, file.hash, file.name)
		if err != nil {
			return domain.Hash{}, err
		}
		entries = append(entries, entry)
	}

	domain.SortTreeEntries(entries)

	tree, err := domain.NewTreeFromEntries(entries)
	if err != nil {
		return domain.Hash{}, err
	}

	data := tree.Serialize()
	hexHash := core.ComputeSHA256(data)
	hash, err := domain.NewHashFromHex(hexHash)
	if err != nil {
		return domain.Hash{}, err
	}

	ok, err := w.objectService.Exists(hash)
	if err != nil {
		return domain.Hash{}, err
	}
	if ok {
		return hash, nil
	}

	err = w.objectService.Write(hash, data)
	if err != nil {
		return domain.Hash{}, err
	}
	return hash, nil
}

// buildRootTree groups flat index entries into an in-memory directory tree.
func buildRootTree(entries []domain.IndexEntry) *directoryNode {
	root := &directoryNode{
		name:     "",
		children: make(map[string]*directoryNode),
		files:    make([]*fileNode, 0),
	}

	for _, entry := range entries {
		parentDir := root
		names := strings.Split(entry.Path().String(), "/")

		for i, name := range names {
			if i == len(names)-1 {
				fileNode := &fileNode{
					mode: entry.Mode(),
					hash: entry.Hash(),
					name: name,
				}
				parentDir.files = append(parentDir.files, fileNode)
			} else {
				var childDir *directoryNode
				if existingChild, exists := parentDir.children[name]; exists {
					childDir = existingChild
				} else {
					childDir = &directoryNode{
						name:     name,
						children: make(map[string]*directoryNode),
						files:    make([]*fileNode, 0),
					}
					parentDir.children[name] = childDir
				}
				parentDir = childDir
			}
		}
	}
	return root
}
