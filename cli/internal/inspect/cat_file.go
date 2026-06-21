package inspect

import (
	"Gel/internal/core"
	"Gel/internal/domain"
	"errors"
	"fmt"
	"io"
)

// CatFileOptions controls which cat-file outputs are produced.
type CatFileOptions struct {
	// ObjectType prints the object type (blob/tree/commit).
	ObjectType bool
	// Pretty prints object content in a human-readable form.
	Pretty bool
	// Size prints object body size in bytes.
	Size bool
	// Exists only checks whether the object exists.
	Exists bool
}

// CatFileService provides object inspection behavior for the cat-file command.
type CatFileService struct {
	objectService *core.ObjectService
}

// NewCatFileService creates a cat-file service backed by object storage reads.
func NewCatFileService(objectService *core.ObjectService) *CatFileService {
	return &CatFileService{
		objectService: objectService,
	}
}

// CatFile executes cat-file behavior for a given object hash.
//
// At least one option must be enabled. Multiple options can be combined
// (for example, printing type and size in sequence). When Exists is set, the
// function only performs existence checking and does not read object content.
func (c *CatFileService) CatFile(writer io.Writer, hash domain.Hash, options CatFileOptions) error {
	if !options.ObjectType && !options.Pretty && !options.Size && !options.Exists {
		return errors.New("cat-file: specify at least one of -t, -p, -s, -e")
	}
	if options.Exists {
		ok, err := c.objectService.Exists(hash)
		if err != nil {
			return fmt.Errorf("cat file: %w", err)
		}
		if !ok {
			return fmt.Errorf("cat file: object '%s' does not exist", hash)
		}
		return nil
	}

	object, err := c.objectService.Read(hash)
	if err != nil {
		return fmt.Errorf("cat file: %w", err)
	}
	if options.ObjectType {
		if _, err := fmt.Fprintf(writer, "%s\n", object.Type()); err != nil {
			return fmt.Errorf("cat file: %w", err)
		}
	}
	if options.Size {
		if _, err := fmt.Fprintf(writer, "%d\n", object.Size()); err != nil {
			return fmt.Errorf("cat file: %w", err)
		}
	}
	if options.Pretty {
		return c.catFileWithPretty(writer, object)
	}
	return nil
}

// catFileWithPretty writes object content in a format tailored to object type:
// tree entries for tree objects, raw body for blobs, and structured commit
// fields for commits.
func (c *CatFileService) catFileWithPretty(writer io.Writer, object domain.Object) error {
	switch object.Type() {
	case domain.ObjectTypeTree:
		tree, ok := object.(*domain.Tree)
		if !ok {
			// TODO: return proper error
			return fmt.Errorf("cat-file: object is not a tree")
		}

		for _, entry := range tree.Entries() {
			objectType, err := entry.Mode.ObjectType()
			if err != nil {
				return fmt.Errorf("cat file: %w", err)
			}
			if _, err := fmt.Fprintf(
				writer,
				"%s %s %s\t%s\n",
				entry.Mode,
				objectType,
				entry.Hash,
				entry.Name,
			); err != nil {
				return fmt.Errorf("cat file: %w", err)
			}
		}
	case domain.ObjectTypeBlob:
		blob, ok := object.(*domain.Blob)
		if !ok {
			// TODO: return proper error
			return fmt.Errorf("cat-file: object is not a blob")
		}
		if _, err := writer.Write(blob.Body()); err != nil {
			return fmt.Errorf("cat file: %w", err)
		}
	case domain.ObjectTypeCommit:
		commit, ok := object.(*domain.Commit)
		if !ok {
			// TODO: return proper error
			return fmt.Errorf("cat-file: object is not a commit")
		}
		if _, err := fmt.Fprintf(
			writer,
			"%s %s\n",
			domain.CommitFieldTree,
			commit.TreeHash,
		); err != nil {
			return fmt.Errorf("cat file: %w", err)
		}
		for _, parentHash := range commit.ParentHashes {
			if _, err := fmt.Fprintf(
				writer,
				"%s %s\n",
				domain.CommitFieldParent,
				parentHash,
			); err != nil {
				return fmt.Errorf("cat file: %w", err)
			}
		}
		if _, err := fmt.Fprintf(
			writer,
			"%s %s <%s> %s %s\n"+
				"%s %s <%s> %s %s\n"+
				"\n%s\n",
			domain.CommitFieldAuthor,
			commit.Author.Name,
			commit.Author.Email,
			commit.Author.Timestamp,
			commit.Author.Timezone,
			domain.CommitFieldCommitter,
			commit.Committer.Name,
			commit.Committer.Email,
			commit.Committer.Timestamp,
			commit.Committer.Timezone,
			commit.Message,
		); err != nil {
			return fmt.Errorf("cat file: %w", err)
		}
	}
	return nil
}
