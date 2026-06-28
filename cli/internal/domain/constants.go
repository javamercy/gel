package domain

import "os"

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
