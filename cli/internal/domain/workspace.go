package domain

import (
	"fmt"
	"path/filepath"
)

const (
	// GelDirName is the repository metadata directory name.
	GelDirName = ".gel"

	// ObjectsDirName is the object database directory name.
	ObjectsDirName = "objects"

	// RefsDirName is the references directory name.
	RefsDirName = "refs"

	// HeadsDirName is the branch-head references directory name.
	HeadsDirName = "heads"

	// HeadFileName is the symbolic HEAD file name.
	HeadFileName = "HEAD"

	// IndexFileName is the staging index file name.
	IndexFileName = "index"
)

// Workspace describes the standard paths of a Gel repository.
type Workspace struct {
	repoRoot   AbsolutePath
	gelDir     AbsolutePath
	objectsDir AbsolutePath
	refsDir    AbsolutePath
	headsDir   AbsolutePath
	headPath   AbsolutePath
	indexPath  AbsolutePath
	configPath AbsolutePath
}

// NewWorkspace constructs a repository layout rooted at repoRoot.
//
// It derives paths only and does not verify that the repository or its
// metadata files exist.
func NewWorkspace(repoRoot AbsolutePath) (*Workspace, error) {
	if repoRoot.value == "" {
		return nil, fmt.Errorf(
			"%w: repository root is the zero value",
			ErrInvalidAbsolutePath,
		)
	}

	join := func(elements ...string) AbsolutePath {
		parts := append([]string{repoRoot.value}, elements...)
		return AbsolutePath{
			value: filepath.Join(parts...),
		}
	}

	return &Workspace{
		repoRoot:   repoRoot,
		gelDir:     join(GelDirName),
		objectsDir: join(GelDirName, ObjectsDirName),
		refsDir:    join(GelDirName, RefsDirName),
		headsDir:   join(GelDirName, RefsDirName, HeadsDirName),
		headPath:   join(GelDirName, HeadFileName),
		indexPath:  join(GelDirName, IndexFileName),
		configPath: join(GelDirName, ConfigFileName),
	}, nil
}

// RepoRoot returns the repository root directory.
func (w *Workspace) RepoRoot() AbsolutePath {
	return w.repoRoot
}

// GelDir returns the repository metadata directory.
func (w *Workspace) GelDir() AbsolutePath {
	return w.gelDir
}

// ObjectsDir returns the object database directory.
func (w *Workspace) ObjectsDir() AbsolutePath {
	return w.objectsDir
}

// RefsDir returns the directory containing references.
func (w *Workspace) RefsDir() AbsolutePath {
	return w.refsDir
}

// HeadsDir returns the directory containing branch heads.
func (w *Workspace) HeadsDir() AbsolutePath {
	return w.headsDir
}

// HeadPath returns the path to the HEAD file.
func (w *Workspace) HeadPath() AbsolutePath {
	return w.headPath
}

// IndexPath returns the path to the index file.
func (w *Workspace) IndexPath() AbsolutePath {
	return w.indexPath
}

// ConfigPath returns the path to the configuration file.
func (w *Workspace) ConfigPath() AbsolutePath {
	return w.configPath
}
