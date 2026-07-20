package commit

import (
	"Gel/internal/core"
	"Gel/internal/domain"
	"fmt"
)

// CommitTreeService creates commit objects from an explicit tree hash.
type CommitTreeService struct {
	objectService *core.ObjectService
	configService *core.ConfigService
}

// NewCommitTreeService creates a commit-tree service.
func NewCommitTreeService(
	objectService *core.ObjectService,
	configService *core.ConfigService,
) *CommitTreeService {
	return &CommitTreeService{
		objectService: objectService,
		configService: configService,
	}
}

// CommitTree creates a commit object with the provided tree and parent hashes.
// Author/committer identity is loaded from config user.name and user.email.
func (c *CommitTreeService) CommitTree(
	hash domain.Hash,
	message string,
	parentHashes []domain.Hash,
) (
	domain.Hash, error,
) {
	_, err := c.objectService.ReadTree(hash)
	if err != nil {
		return domain.Hash{}, fmt.Errorf("commit-tree: %w", err)
	}

	// TODO: fix
	serializedData := []byte(nil)
	hexCommitHash := core.ComputeSHA256(serializedData)
	commitHash, err := domain.NewHashFromHex(hexCommitHash)
	if err != nil {
		return domain.Hash{}, fmt.Errorf("commit-tree: %w", err)
	}
	if err := c.objectService.Write(commitHash, serializedData); err != nil {
		return domain.Hash{}, fmt.Errorf("commit-tree: %w", err)
	}
	return commitHash, nil
}
