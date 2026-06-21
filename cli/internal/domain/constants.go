package domain

import "os"

const (
	// GelDirName is the repository metadata directory name.
	GelDirName string = ".gel"

	// ObjectsDirName is the object storage directory name.
	ObjectsDirName string = "objects"

	// RefsDirName is the references directory name.
	RefsDirName string = "refs"

	// HeadFileName is the symbolic HEAD reference filename.
	HeadFileName string = "HEAD"

	// HeadsDirName is the refs/heads directory name.
	HeadsDirName string = "heads"

	// IndexFileName is the index file name.
	IndexFileName string = "index"

	// ConfigFileName is the repository config file name.
	ConfigFileName string = "config.toml"
)

const (
	// DefaultBranchName is the default branch name.
	DefaultBranchName string = "main"

	// DefaultBranchRef is the full ref path for the default branch.
	DefaultBranchRef string = "refs/heads/main"
)

const (
	// DefaultFilePermission is the default permission for regular files written by Gel.
	DefaultFilePermission os.FileMode = 0o644

	// DefaultDirPermission is the default permission for directories created by Gel.
	DefaultDirPermission os.FileMode = 0o755
)
