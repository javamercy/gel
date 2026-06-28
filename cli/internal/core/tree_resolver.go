package core

import (
	"Gel/internal/domain"
	"strings"
)

// PathHashes maps normalized repository-relative paths to content hashes.
type PathHashes map[domain.NormalizedPath]domain.Hash

func (p PathHashes) ExtractPaths() (paths []domain.NormalizedPath) {
	paths = make([]domain.NormalizedPath, 0, len(p))
	for path := range p {
		paths = append(paths, path)
	}
	return
}

func (p PathHashes) ExtractPathsSorted() (paths []domain.NormalizedPath) {
	paths = p.ExtractPaths()
	domain.SortPaths(paths)
	return
}

// TreeResolver resolves path->hash snapshots from repository trees, index, and working tree.
type TreeResolver struct {
	objectService  *ObjectService
	indexService   *IndexService
	refService     *RefService
	pathResolver   *PathResolver
	changeDetector *ChangeDetector
	workspace      *domain.Workspace
}

// NewTreeResolver creates a tree resolver.
func NewTreeResolver(
	objectService *ObjectService,
	indexService *IndexService,
	refService *RefService,
	pathResolver *PathResolver,
	changeDetector *ChangeDetector,
	workspace *domain.Workspace,
) *TreeResolver {
	return &TreeResolver{
		objectService:  objectService,
		indexService:   indexService,
		refService:     refService,
		pathResolver:   pathResolver,
		changeDetector: changeDetector,
		workspace:      workspace,
	}
}

// ResolveHEAD resolves the tree snapshot pointed to by HEAD.
func (t *TreeResolver) ResolveHEAD() (PathHashes, error) {
	return t.ResolveRef(domain.HeadFileName)
}

// ResolveRef resolves the tree snapshot pointed to by refName.
func (t *TreeResolver) ResolveRef(refName string) (PathHashes, error) {
	commitHash, err := t.refService.Resolve(refName)
	if err != nil {
		return nil, err
	}
	return t.ResolveCommit(commitHash)
}

// ResolveCommit walks a commit tree recursively and returns normalized path hashes.
func (t *TreeResolver) ResolveCommit(hash domain.Hash) (PathHashes, error) {
	commit, err := t.objectService.ReadCommit(hash)
	if err != nil {
		return nil, err
	}

	pathHashes := make(map[domain.NormalizedPath]domain.Hash)
	walker := NewTreeWalker(t.objectService, WalkOptions{Recursive: true})
	err = walker.Walk(
		commit.TreeHash, "", func(e domain.TreeEntry, relPath string) error {
			normalizedPath, err := domain.ParseNormalizedPath(relPath)
			if err != nil {
				return err
			}
			pathHashes[normalizedPath] = e.Hash()
			return nil
		},
	)
	return pathHashes, err
}

// ResolveIndex returns the current index snapshot as normalized path hashes.
func (t *TreeResolver) ResolveIndex() (PathHashes, error) {
	entries, err := t.indexService.GetEntries()
	if err != nil {
		return nil, err
	}

	pathHashes := make(PathHashes, len(entries))
	for _, entry := range entries {
		pathHashes[entry.Path] = entry.Hash
	}
	return pathHashes, nil
}

// ResolveWorkingTree returns repository-wide working tree path hashes.
// The scan is rooted at repository root so results are independent of current working directory.
func (t *TreeResolver) ResolveWorkingTree() (PathHashes, error) {
	resolvedPaths, err := t.pathResolver.Resolve([]string{t.workspace.RepoRoot().String()})
	if err != nil {
		return nil, err
	}

	index, err := t.indexService.Read()
	if err != nil {
		return nil, err
	}

	pathHashes := make(map[domain.NormalizedPath]domain.Hash)
	for _, resolved := range resolvedPaths {
		for path := range resolved.NormalizedPaths {
			entry, _ := index.FindEntry(path)
			if entry != nil {
				changeResult, err := t.changeDetector.DetectFileChange(entry)
				if err != nil {
					return nil, err
				}
				switch changeResult.FileState {
				case FileStateUnchanged:
					pathHashes[path] = entry.Hash
				case FileStateModified:
					pathHashes[path] = changeResult.NewHash
				case FileStateDeleted:
					delete(pathHashes, path)
				}
			} else {
				absolutePath, err := path.ToAbsolutePath(t.workspace.RepoRoot())
				if err != nil {
					return nil, err
				}

				hash, _, err := t.objectService.ComputeObjectHash(absolutePath)
				if err != nil {
					return nil, err
				}
				pathHashes[path] = hash
			}
		}
	}
	return pathHashes, nil
}

// LookupPathInTree traverses a tree hierarchy using the given tree hash and path, returning the matching tree entry.
func (t *TreeResolver) LookupPathInTree(treeHash domain.Hash, path domain.NormalizedPath) (domain.TreeEntry, error) {
	segments := strings.Split(path.String(), "/")
	return t.lookupPathInTreeRecursive(treeHash, segments)
}

// lookupPathInTreeRecursive traverses a tree hierarchy using the given tree hash and path segments, returning the matching tree entry.
func (t *TreeResolver) lookupPathInTreeRecursive(treeHash domain.Hash, segments []string) (domain.TreeEntry, error) {
	tree, err := t.objectService.ReadTree(treeHash)
	if err != nil {
		return domain.TreeEntry{}, err
	}
	for _, entry := range tree.Entries() {
		if entry.Name() == segments[0] {
			if len(segments) == 1 {
				return entry, nil
			}
			return t.lookupPathInTreeRecursive(entry.Hash(), segments[1:])
		}
	}
	return domain.TreeEntry{}, ErrPathNotFoundInTree
}
