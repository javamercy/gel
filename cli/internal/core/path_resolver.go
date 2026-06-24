package core

import (
	"Gel/internal/domain"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

const (
	PathspecTypeFile PathspecType = iota
	PathspecTypeDirectory
	PathspecTypeGlobPattern
	PathspecTypeNonExistent
)

const (
	globPatterns string = "*?[]"
)

type PathspecType int

// String returns a human-readable pathspec kind.
func (t PathspecType) String() string {
	switch t {
	case PathspecTypeFile:
		return "file"
	case PathspecTypeDirectory:
		return "directory"
	case PathspecTypeGlobPattern:
		return "glob pattern"
	case PathspecTypeNonExistent:
		return "non-existent"
	}
	return "unknown"
}

type ResolvedPath struct {
	Type            PathspecType
	NormalizedScope string
	NormalizedPaths map[domain.NormalizedPath]bool
}

// PathResolver expands pathspecs to normalized repository-relative paths.
type PathResolver struct {
	paths           *RepositoryPathResolver
	ignoredPatterns map[string]bool
}

// NewPathResolver creates a resolver rooted at repoDir.
// When ignoredPatterns is nil, a default set of VCS/editor directories is used.
func NewPathResolver(paths *RepositoryPathResolver, ignoredPatterns map[string]bool) *PathResolver {
	if ignoredPatterns == nil {
		// TODO: implement gelignore.
		ignoredPatterns = map[string]bool{
			".gel":  true,
			".git":  true,
			".idea": true,
		}
	}
	return &PathResolver{
		paths:           paths,
		ignoredPatterns: ignoredPatterns,
	}
}

// Resolve classifies and expands each pathspec and returns normalized results.
// NormalizedScope is repository-relative and used for scoped index reconciliation.
func (p *PathResolver) Resolve(pathspecs []string) ([]ResolvedPath, error) {
	resolvedPaths := make([]ResolvedPath, 0)
	for _, pathspec := range pathspecs {
		pathspecType, err := classifyPathspec(pathspec)
		if err != nil {
			return nil, err
		}

		paths, err := p.expandPathspec(pathspec, pathspecType)
		if err != nil {
			return nil, err
		}

		normalizedScope, err := p.normalizeScope(pathspec, pathspecType)
		if err != nil {
			return nil, err
		}

		normalizedPaths := make(map[domain.NormalizedPath]bool)
		for _, path := range paths {
			resolved, err := p.paths.Resolve(path)
			if err != nil {
				return nil, err
			}

			normPath := resolved.Normalized
			if p.shouldIgnore(normPath.String()) || normalizedPaths[normPath] {
				continue
			}
			normalizedPaths[normPath] = true
		}
		resolvedPaths = append(
			resolvedPaths, ResolvedPath{
				Type:            pathspecType,
				NormalizedScope: normalizedScope,
				NormalizedPaths: normalizedPaths,
			},
		)
	}
	return resolvedPaths, nil
}

// shouldIgnore checks whether any segment of a path matches ignored patterns.
func (p *PathResolver) shouldIgnore(path string) bool {
	segments := strings.Split(filepath.ToSlash(path), "/")
	for _, segment := range segments {
		if p.ignoredPatterns[segment] {
			return true
		}
	}
	return false
}

// expandPathspec resolves a pathspec according to its classified type.
func (p *PathResolver) expandPathspec(pathspec string, pathspecType PathspecType) ([]string, error) {
	switch pathspecType {
	case PathspecTypeFile:
		return []string{pathspec}, nil
	case PathspecTypeDirectory:
		return p.expandDirectory(pathspec)
	case PathspecTypeGlobPattern:
		return p.expandGlobPattern(pathspec)
	case PathspecTypeNonExistent:
		return []string{}, nil
	default:
		return nil, ErrUnknownPathspecType
	}
}

// normalizeScope converts the original pathspec into a repository-relative scope.
// For directory/glob pathspecs this ensures subdirectory invocations are scoped
// correctly instead of defaulting to repository root.
func (p *PathResolver) normalizeScope(pathspec string, pathspecType PathspecType) (string, error) {
	switch pathspecType {
	case PathspecTypeFile,
		PathspecTypeDirectory,
		PathspecTypeGlobPattern,
		PathspecTypeNonExistent:
		resolved, err := p.paths.Resolve(pathspec)
		if err != nil {
			return "", err
		}
		return resolved.Normalized.String(), nil
	default:
		return "", ErrUnknownPathspecType
	}
}

// classifyPathspec determines how a pathspec should be expanded.
func classifyPathspec(pathspec string) (PathspecType, error) {
	if strings.ContainsAny(pathspec, globPatterns) {
		return PathspecTypeGlobPattern, nil
	}

	fileInfo, err := os.Stat(pathspec)
	if errors.Is(err, os.ErrNotExist) {
		return PathspecTypeNonExistent, nil
	} else if err != nil {
		return PathspecTypeNonExistent, err
	}
	if fileInfo.IsDir() {
		return PathspecTypeDirectory, nil
	}
	return PathspecTypeFile, nil
}

// expandDirectory walks a directory pathspec and returns all file paths.
func (p *PathResolver) expandDirectory(path string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(
		path, func(p string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() {
				info, err := os.Stat(p)
				if err == nil && info.IsDir() {
					return nil
				}
				files = append(files, p)
			}
			return nil
		},
	)
	if err != nil {
		return nil, err
	}
	return files, nil
}

// expandGlobPattern expands a glob and recursively expands directory matches.
func (p *PathResolver) expandGlobPattern(pattern string) ([]string, error) {
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	var result []string
	for _, match := range matches {
		info, err := os.Stat(match)
		if err != nil {
			continue
		}
		if info.IsDir() {
			files, err := p.expandDirectory(match)
			if err != nil {
				return nil, err
			}
			result = append(result, files...)
		} else {
			result = append(result, match)
		}
	}
	return result, nil
}
