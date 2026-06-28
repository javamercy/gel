package inspect

import (
	"Gel/internal/core"
	"Gel/internal/diff"
	"Gel/internal/domain"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

// ShowOptions controls optional show output variants.
type ShowOptions struct {
	// NoPatch suppresses commit patch output.
	NoPatch bool
	// Oneline requests compact commit header output.
	Oneline bool
	// Stat requests diffStat output instead of full hunks.
	Stat bool
	// NameOnly requests changed paths only.
	NameOnly bool
	// NameStatus requests changed paths with change status labels.
	NameStatus bool
}

// ShowMode identifies which object representation a ShowResult contains.
type ShowMode int

const (
	// ShowModeCommit indicates a commit object result.
	ShowModeCommit ShowMode = iota
	// ShowModeTree indicates a tree object result.
	ShowModeTree
	// ShowModeBlob indicates a blob object result.
	ShowModeBlob
)

// ShowResult is a tagged union holding exactly one object-specific show result.
type ShowResult struct {
	Mode   ShowMode
	Commit *ShowCommitResult
	Tree   *ShowTreeResult
	Blob   *ShowBlobResult
}

// ShowCommitResult contains commit metadata plus diff output for display.
type ShowCommitResult struct {
	Hash   domain.Hash
	Branch string // e.g. "HEAD -> main" (only when true)
	Commit *domain.Commit
	Diff   []*diff.DiffResult // nil when NoPatch
}

// ShowTreeResult contains a tree hash and its sorted direct entries.
type ShowTreeResult struct {
	Hash        domain.Hash
	TreeEntries []domain.TreeEntry
}

// ShowBlobResult contains blob hash and raw blob bytes.
type ShowBlobResult struct {
	Hash domain.Hash
	Body []byte
}

// resolvedObjectRef stores a resolved object hash and optional decoration text.
type resolvedObjectRef struct {
	Hash       domain.Hash
	Decoration string
}

// ShowService resolves object references and builds object-typed show results.
type ShowService struct {
	objectService *core.ObjectService
	refService    *core.RefService
	diffService   *diff.DiffService
}

// NewShowService creates a show service.
func NewShowService(
	objectService *core.ObjectService,
	refService *core.RefService,
	diffService *diff.DiffService,
) *ShowService {
	return &ShowService{
		objectService: objectService,
		refService:    refService,
		diffService:   diffService,
	}
}

// Show resolves objectRef (HEAD, branch, or hash) and returns a typed show result.
func (s *ShowService) Show(objectRef string, options ShowOptions) (*ShowResult, error) {
	// TODO: implement ShowOptions
	_ = options

	resolved, err := s.resolveObjectRef(objectRef)
	if err != nil {
		return nil, fmt.Errorf("show: failed to resolve object reference: %w", err)
	}

	object, err := s.objectService.Read(resolved.Hash)
	if err != nil {
		return nil, fmt.Errorf("show: %w", err)
	}

	switch obj := object.(type) {
	case *domain.Blob:
		return &ShowResult{
			Mode: ShowModeBlob,
			Blob: &ShowBlobResult{Hash: resolved.Hash, Body: obj.Body()},
		}, nil
	case *domain.Tree:
		domain.SortTreeEntriesByName(obj.Entries())
		return &ShowResult{
			Mode: ShowModeTree,
			Tree: &ShowTreeResult{Hash: resolved.Hash, TreeEntries: obj.Entries()},
		}, nil
	case *domain.Commit:
		commitResult, err := s.buildShowCommitResult(resolved, obj)
		if err != nil {
			return nil, fmt.Errorf("show: %w", err)
		}
		return &ShowResult{Mode: ShowModeCommit, Commit: commitResult}, nil
	default:
		return nil, fmt.Errorf("'%s': %w", object.Type(), ErrUnsupportedObjectType)
	}
}

// resolveObjectRef resolves a user-provided reference to a concrete object hash.
func (s *ShowService) resolveObjectRef(objectRef string) (resolvedObjectRef, error) {
	if objectRef == "" || objectRef == domain.HeadFileName {
		hash, err := s.refService.Resolve(domain.HeadFileName)
		if err != nil {
			if errors.Is(err, core.ErrRefNotFound) {
				return resolvedObjectRef{}, fmt.Errorf("no commits yet")
			}
			return resolvedObjectRef{}, err
		}

		ref, err := s.refService.ReadSymbolic(domain.HeadFileName)
		if err != nil {
			return resolvedObjectRef{}, err
		}

		branch := strings.TrimPrefix(ref, filepath.Join(domain.RefsDirName, domain.HeadsDirName)+"/")
		return resolvedObjectRef{Hash: hash, Decoration: "HEAD -> " + branch}, nil
	}

	ref := filepath.Join(domain.RefsDirName, domain.HeadsDirName, objectRef)
	exists, err := s.refService.Exists(ref)
	if err != nil {
		return resolvedObjectRef{}, err
	}
	if exists {
		hash, err := s.refService.Read(ref)
		if err != nil {
			return resolvedObjectRef{}, err
		}
		return resolvedObjectRef{Hash: hash, Decoration: objectRef}, nil
	}

	hash, err := domain.NewHashFromHex(objectRef)
	if err != nil {
		return resolvedObjectRef{}, fmt.Errorf("'%s': %w", objectRef, core.ErrRefNotFound)
	}

	ok, err := s.objectService.Exists(hash)
	if err != nil {
		return resolvedObjectRef{}, err
	}
	if !ok {
		return resolvedObjectRef{}, fmt.Errorf("'%s': %w", objectRef, ErrObjectNotFound)
	}
	return resolvedObjectRef{Hash: hash}, nil
}

// buildShowCommitResult builds commit metadata and parent-vs-commit diff results.
func (s *ShowService) buildShowCommitResult(resolved resolvedObjectRef, commit *domain.Commit) (
	*ShowCommitResult, error,
) {
	var parentHash domain.Hash
	if len(commit.ParentHashes) > 0 {
		parentHash = commit.ParentHashes[0]
	}

	diffResults, err := s.diffService.Diff(
		diff.DiffOptions{
			Mode:             diff.DiffModeCommitVsCommit,
			BaseCommitHash:   parentHash,
			TargetCommitHash: resolved.Hash,
		},
	)
	if err != nil {
		return nil, err
	}
	return &ShowCommitResult{
		Hash:   resolved.Hash,
		Branch: resolved.Decoration,
		Commit: commit,
		Diff:   diffResults,
	}, nil
}
