package core

import (
	"Gel/internal/domain"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type CommitResolver struct {
	refService    *RefService
	objectService *ObjectService
}

func NewCommitResolver(
	refService *RefService,
	objectService *ObjectService,
) *CommitResolver {
	return &CommitResolver{
		refService:    refService,
		objectService: objectService,
	}
}

func (r *CommitResolver) Resolve(target string) (domain.Hash, error) {
	target = strings.TrimSpace(target)
	if target == "" {
		return domain.Hash{}, errors.New("empty target")
	}

	base, steps, ok, err := parseTildeExpression(target)
	if err != nil {
		return domain.Hash{}, err
	}
	if !ok {
		return r.resolveBase(target)
	}
	if strings.TrimSpace(base) == "" {
		return domain.Hash{}, errors.New("invalid revision: missing base before '~'")
	}

	baseHash, err := r.resolveBase(base)
	if err != nil {
		return domain.Hash{}, err
	}
	return r.walkNParents(baseHash, steps)
}

func (r *CommitResolver) resolveBase(base string) (domain.Hash, error) {
	switch {
	case base == domain.HeadFileName:
		return r.refService.Resolve(domain.HeadFileName)

	case isFullHash(base):
		hash, err := domain.NewHashFromHex(base)
		if err != nil {
			return domain.Hash{}, err
		}
		if err := r.ensureCommit(hash); err != nil {
			return domain.Hash{}, err
		}
		return hash, nil

	case strings.HasPrefix(base, domain.RefsDirName+"/"):
		hash, err := r.refService.Read(base)
		if err != nil {
			return domain.Hash{}, err
		}
		if err := r.ensureCommit(hash); err != nil {
			return domain.Hash{}, err
		}
		return hash, nil
	}

	branchRef := filepath.Join(domain.RefsDirName, domain.HeadsDirName, base)
	hash, err := r.refService.Read(branchRef)
	if err != nil {
		return domain.Hash{}, err
	}
	if err := r.ensureCommit(hash); err != nil {
		return domain.Hash{}, err
	}
	return hash, nil
}

func (r *CommitResolver) ensureCommit(hash domain.Hash) error {
	if hash.IsZero() {
		return errors.New("empty hash")
	}
	if _, err := r.objectService.ReadCommit(hash); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("commit not found: %w", err)
		}
		// TODO: return proper error
		/* if errors.Is(err, domain.ErrObjectTypeMismatch) {
			return fmt.Errorf("object is not a commit: %w", err)
		} */
		return err
	}
	return nil
}

func (r *CommitResolver) walkNParents(hash domain.Hash, steps int) (domain.Hash, error) {
	curr := hash
	for i := 0; i < steps; i++ {
		commit, err := r.objectService.ReadCommit(curr)
		if err != nil {
			return domain.Hash{}, err
		}
		if len(commit.ParentHashes) == 0 {
			return domain.Hash{}, errors.New("commit has no parents")
		}
		curr = commit.ParentHashes[0]
	}
	return curr, nil
}

func parseTildeExpression(expression string) (base string, steps int, ok bool, err error) {
	tildeIndex := strings.LastIndex(expression, "~")
	if tildeIndex == -1 {
		return "", 0, false, nil
	}
	base = expression[:tildeIndex]
	steps, err = parsePositiveInteger(expression[tildeIndex+1:], 1)
	if err != nil {
		return "", 0, false, err
	}
	return base, steps, true, nil
}

func parsePositiveInteger(s string, defaultValue int) (int, error) {
	if s == "" {
		return defaultValue, nil
	}

	value, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}
	if value < 0 {
		return 0, errors.New("negative number")
	}
	return value, nil
}

func isFullHash(s string) bool {
	if len(s) != domain.HashHexLength {
		return false
	}

	_, err := hex.DecodeString(s)
	return err == nil
}
