package tree

import (
	"Gel/internal/core"
	"Gel/internal/domain"
	"fmt"
)

// LsTreeOptions controls tree listing behavior.
type LsTreeOptions struct {
	// Recursive descends into subtrees.
	Recursive bool
	// ShowTrees includes tree entries in output when traversing.
	ShowTrees bool
	// NameOnly prints only relative paths.
	NameOnly bool
}

// LsTreeService lists tree object contents.
type LsTreeService struct {
	objectService *core.ObjectService
}

// NewLsTreeService creates an ls-tree service.
func NewLsTreeService(objectService *core.ObjectService) *LsTreeService {
	return &LsTreeService{
		objectService: objectService,
	}
}

// LsTree resolves a tree hash and returns formatted listing lines.
//
// In NameOnly mode it returns relative paths only; otherwise it returns
// "<mode> <type> <hash>\t<path>" entries.
func (l *LsTreeService) LsTree(hash domain.Hash, options LsTreeOptions) ([]string, error) {
	var contents []string
	processor := func(entry domain.TreeEntry, relPath string) error {
		objectType, err := entry.Mode().ObjectType()
		if err != nil {
			return err
		}
		if options.NameOnly {
			contents = append(contents, relPath)
		} else {
			result := fmt.Sprintf(
				"%s %s %s\t%s",
				entry.Mode, objectType, entry.Hash, relPath,
			)
			contents = append(contents, result)
		}
		return nil
	}

	walkOptions := core.WalkOptions{
		Recursive:    options.Recursive,
		IncludeTrees: options.ShowTrees,
		OnlyTrees:    false,
	}
	treeWalker := core.NewTreeWalker(l.objectService, walkOptions)
	if err := treeWalker.Walk(hash, "", processor); err != nil {
		return nil, err
	}
	return contents, nil
}
