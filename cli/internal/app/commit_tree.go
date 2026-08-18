package app

import (
	"Gel/internal/domain"
	"Gel/internal/storage"
	"fmt"
	"time"
)

// CommitTreeInput specifies the tree, parents, and message for a new commit.
type CommitTreeInput struct {
	// TreeHash identifies the root tree object and must not be zero.
	TreeHash domain.Hash
	// ParentHashes identifies existing parent commits in semantic order. An
	// empty slice creates a root commit.
	ParentHashes []domain.Hash
	// Message is the commit message.
	Message string
}

// CommitTreeResult contains the identifier of the written commit object.
type CommitTreeResult struct {
	// CommitHash is the SHA-256 identifier of the new commit.
	CommitHash domain.Hash
}

// CommitTree creates commit objects from tree and parent objects.
type CommitTree struct {
	configStore *storage.ConfigStore
	objectStore *storage.ObjectStore
	now         func() time.Time
}

// NewCommitTree returns a CommitTree backed by configStore and objectStore.
//
// Both stores must not be nil when Run is called.
func NewCommitTree(
	configStore *storage.ConfigStore,
	objectStore *storage.ObjectStore,
) *CommitTree {
	return &CommitTree{
		configStore: configStore,
		objectStore: objectStore,
		now:         time.Now,
	}
}

// Run validates the input, creates a commit, and writes it to object storage.
//
// Run verifies that TreeHash identifies a tree and that every parent hash
// identifies a commit. It reads the author name and email from the user.name
// and user.email configuration entries, and uses the current local time for
// both the author and committer identities. Run returns an error when either
// identity is not configured or any input, object, configuration, or write
// operation is invalid. Errors return a zero CommitTreeResult.
func (c *CommitTree) Run(input CommitTreeInput) (CommitTreeResult, error) {
	commitInput := commitObjectInput{
		treeHash:     input.TreeHash,
		parentHashes: input.ParentHashes,
		message:      input.Message,
	}

	commitHash, err := writeCommitObject(
		commitInput,
		c.configStore,
		c.objectStore,
		c.now(),
	)
	if err != nil {
		return CommitTreeResult{}, fmt.Errorf(
			"write commit object: %w",
			err,
		)
	}
	return CommitTreeResult{CommitHash: commitHash}, nil
}
