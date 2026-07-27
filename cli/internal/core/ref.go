package core

import (
	"Gel/internal/domain"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// RefService manages operations related to symbolic and direct references within a repository workspace.
type RefService struct {
	workspace *domain.Workspace
}

// NewRefService initializes and returns a new RefService instance for managing references in the specified workspace.
func NewRefService(workspace *domain.Workspace) *RefService {
	return &RefService{
		workspace: workspace,
	}
}

// ReadSymbolic reads a symbolic reference file (for example HEAD) and returns its target ref path.
// The file content must be in "ref: <target>" format.
func (r *RefService) ReadSymbolic(name string) (string, error) {
	refPath, err := r.symbolicPath(name)
	if err != nil {
		return "", err
	}

	contentBytes, err := os.ReadFile(refPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("'%s': %w", name, ErrRefNotFound)
		}
		return "", fmt.Errorf("ref: failed to read '%s': %w", name, err)
	}

	contentStr := strings.TrimSpace(string(contentBytes))
	if !strings.HasPrefix(contentStr, "ref: ") {
		return "", fmt.Errorf("'%s': %w", contentStr, ErrInvalidSymbolicRef)
	}
	return strings.TrimPrefix(contentStr, "ref: "), nil
}

// WriteSymbolic writes a symbolic reference file in "ref: <target>" format.
// Name must resolve inside .gel and ref must start with "refs/".
func (r *RefService) WriteSymbolic(name, ref string) error {
	if name == "" || ref == "" {
		return ErrInvalidSymbolicRef
	}
	if err := validateRefPrefix(ref); err != nil {
		return err
	}

	path, err := r.symbolicPath(name)
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, domain.DefaultDirPermission); err != nil {
		return fmt.Errorf("ref: failed to create directory '%s': %w", dir, err)
	}

	contentStr := fmt.Sprintf("ref: %s\n", ref)
	if err := os.WriteFile(path, []byte(contentStr), domain.DefaultFilePermission); err != nil {
		return fmt.Errorf("ref: failed to write symbolic ref '%s': %w", name, err)
	}
	return nil
}

// Read resolves a direct ref file (for example refs/heads/main) to its commit hash.
// Empty ref files are treated as zero hashes.
func (r *RefService) Read(ref string) (domain.Hash, error) {
	if err := validateRefPrefix(ref); err != nil {
		return domain.Hash{}, err
	}

	absPath := filepath.Join(r.workspace.GelDir().String(), ref)
	contentBytes, err := os.ReadFile(absPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return domain.Hash{}, fmt.Errorf("'%s': %w", ref, ErrRefNotFound)
		}
		return domain.Hash{}, fmt.Errorf("ref: failed to read '%s': %w", ref, err)
	}
	if len(contentBytes) == 0 {
		return domain.Hash{}, nil
	}

	hexHash := strings.TrimSpace(string(contentBytes))
	hash, err := domain.ParseHash(hexHash)
	if err != nil {
		return domain.Hash{}, fmt.Errorf("ref: %w", err)
	}
	return hash, nil
}

// Write updates a direct ref file with hash plus trailing newline.
// Missing parent directories are created.
func (r *RefService) Write(ref string, hash domain.Hash) error {
	if err := validateRefPrefix(ref); err != nil {
		return err
	}

	absPath := filepath.Join(r.workspace.GelDir().String(), ref)
	dir := filepath.Dir(absPath)
	if err := os.MkdirAll(dir, domain.DefaultDirPermission); err != nil {
		return fmt.Errorf("ref: failed to create directory '%s': %w", dir, err)
	}

	contentStr := fmt.Sprintf("%s\n", hash)
	if err := os.WriteFile(absPath, []byte(contentStr), domain.DefaultFilePermission); err != nil {
		return fmt.Errorf("ref: failed to write '%s': %w", ref, err)
	}
	return nil
}

// Delete removes a direct ref path under .gel/refs.
func (r *RefService) Delete(ref string) error {
	if err := validateRefPrefix(ref); err != nil {
		return err
	}

	path := filepath.Join(r.workspace.GelDir().String(), ref)
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("ref: failed to delete '%s': %w", ref, err)
	}
	return nil
}

// Exists reports whether a direct ref path exists.
// It validates ref prefix and returns wrapped stat errors.
func (r *RefService) Exists(ref string) (bool, error) {
	if err := validateRefPrefix(ref); err != nil {
		return false, err
	}

	path := filepath.Join(r.workspace.GelDir().String(), ref)
	ok, err := Exists(path)
	if err != nil {
		return false, fmt.Errorf("ref: %w", err)
	}
	return ok, nil
}

// Resolve reads a symbolic name (for example HEAD) and then reads the direct ref it points to.
func (r *RefService) Resolve(name string) (domain.Hash, error) {
	ref, err := r.ReadSymbolic(name)
	if err != nil {
		return domain.Hash{}, err
	}
	return r.Read(ref)
}

// symbolicPath sanitizes symbolic-ref file names and maps them under .gel.
// Absolute and traversal paths are rejected.
func (r *RefService) symbolicPath(name string) (string, error) {
	cleanName := filepath.Clean(name)
	if cleanName == "." || filepath.IsAbs(cleanName) || cleanName == ".." ||
		strings.HasPrefix(cleanName, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("'%s': %w", name, ErrInvalidSymbolicRef)
	}
	return filepath.Join(r.workspace.GelDir().String(), cleanName), nil
}

// validateRefPrefix ensures a direct reference path uses the refs/ namespace.
func validateRefPrefix(ref string) error {
	if !strings.HasPrefix(ref, "refs/") {
		return fmt.Errorf("'%s': %w", ref, ErrInvalidRef)
	}
	return nil
}
