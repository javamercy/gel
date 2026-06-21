package domain

import (
	"bytes"
	"errors"
)

// ErrInvalidCommitFormat is returned when commit body parsing fails.
var ErrInvalidCommitFormat = errors.New("invalid commit format")

// ErrInvalidCommitField is returned when an unknown commit header is encountered.
var ErrInvalidCommitField = errors.New("invalid commit field")

// CommitFieldTree is the commit header key that stores the root tree hash.
const CommitFieldTree string = "tree"

// CommitFieldParent is the commit header key for a parent commit hash.
// A commit may contain zero or more parent headers.
const CommitFieldParent string = "parent"

// CommitFieldAuthor is the commit header key for author identity metadata.
const CommitFieldAuthor string = "author"

// CommitFieldCommitter is the commit header key for committer identity metadata.
const CommitFieldCommitter string = "committer"

// CommitFields contains the semantic fields represented by a commit object body.
// It is the normalized in-memory form used by serialization and parsing code.
type CommitFields struct {
	// TreeHash points to the root tree object for this commit snapshot.
	TreeHash Hash

	// ParentHashes contains zero or more parent commit hashes.
	// ParentHashes[0], when present, is the first parent used by first-parent traversal.
	ParentHashes []Hash

	// Author describes who originally authored the change and when.
	Author Identity

	// Committer describes who created this commit object and when.
	Committer Identity

	// Message is the full commit message body and may contain multiple lines.
	Message string
}

// Commit represents a parsed commit domain object.
// body stores the raw commit payload (without object header), while CommitFields
// stores parsed structured fields from the same payload.
type Commit struct {
	body []byte
	CommitFields
}

// Body returns a defensive copy of the raw commit body bytes.
func (commit *Commit) Body() []byte {
	return append([]byte(nil), commit.body...)
}

// Type returns the domain object type for Commit.
func (commit *Commit) Type() ObjectType {
	return ObjectTypeCommit
}

// Size returns the byte length of the raw commit body.
func (commit *Commit) Size() int {
	return len(commit.body)
}

// Serialize returns the full object serialization in "<type> <size>\\x00<body>" format.
func (commit *Commit) Serialize() []byte {
	return serializeObject(ObjectTypeCommit, commit.body)
}

// NewCommit parses a raw commit body and returns a validated Commit.
// The input bytes are copied so subsequent caller mutations do not affect the Commit.
func NewCommit(body []byte) (*Commit, error) {
	bodyCopy := append([]byte(nil), body...)
	fields, err := deserializeFields(bodyCopy)
	if err != nil {
		return nil, err
	}
	return &Commit{
		body:         bodyCopy,
		CommitFields: cloneCommitFields(fields),
	}, nil
}

// DeserializeCommit is an alias for NewCommit for compatibility with older call sites.
func DeserializeCommit(data []byte) (*Commit, error) {
	return NewCommit(data)
}

// NewCommitFromFields validates commit fields, serializes them into canonical commit-body
// format, and returns a new Commit value.
func NewCommitFromFields(commitFields CommitFields) (*Commit, error) {
	if err := validateCommitFields(commitFields); err != nil {
		return nil, err
	}
	clonedFields := cloneCommitFields(commitFields)
	return &Commit{
		body:         serializeBody(clonedFields),
		CommitFields: clonedFields,
	}, nil
}

// validateCommitFields validates logical commit invariants before serialization:
// non-empty tree hash, no empty parent hashes, and valid author/committer identities.
func validateCommitFields(fields CommitFields) error {
	if fields.TreeHash.IsZero() {
		return ErrInvalidCommitFormat
	}
	for _, parentHash := range fields.ParentHashes {
		if parentHash.IsZero() {
			return ErrInvalidCommitFormat
		}
	}
	if _, err := NewIdentity(
		fields.Author.Name,
		fields.Author.Email,
		fields.Author.Timestamp,
		fields.Author.Timezone,
	); err != nil {
		return err
	}
	if _, err := NewIdentity(
		fields.Committer.Name,
		fields.Committer.Email,
		fields.Committer.Timestamp,
		fields.Committer.Timezone,
	); err != nil {
		return err
	}
	return nil
}

// serializeBody serializes commit fields into raw commit-body format:
//
//	tree <hash>\n
//	parent <hash>\n (zero or more)
//	author <name> <email> <timestamp> <timezone>\n
//	committer <name> <email> <timestamp> <timezone>\n
//	\n
//	<message>
func serializeBody(fields CommitFields) []byte {
	var buffer bytes.Buffer
	buffer.WriteString(CommitFieldTree)
	buffer.WriteString(" ")
	buffer.WriteString(fields.TreeHash.Hex())
	buffer.WriteString("\n")

	for _, parentHash := range fields.ParentHashes {
		buffer.WriteString(CommitFieldParent)
		buffer.WriteString(" ")
		buffer.WriteString(parentHash.Hex())
		buffer.WriteString("\n")
	}
	buffer.WriteString(CommitFieldAuthor)
	buffer.WriteString(" ")
	buffer.Write(fields.Author.Serialize())
	buffer.WriteString("\n")
	buffer.WriteString(CommitFieldCommitter)
	buffer.WriteString(" ")
	buffer.Write(fields.Committer.Serialize())
	buffer.WriteString("\n")
	buffer.WriteString("\n")
	buffer.WriteString(fields.Message)

	return buffer.Bytes()
}

// deserializeFields parses a raw commit body into CommitFields.
// It requires exactly one tree, one author, one committer header, and a message section.
// Duplicate tree/author/committer headers are rejected.
func deserializeFields(data []byte) (CommitFields, error) {
	var fields CommitFields
	i := 0
	hasTree := false
	hasAuthor := false
	hasCommitter := false
	hasMessage := false

	for i < len(data) {
		if data[i] == '\n' {
			i++
			fields.Message = string(data[i:])
			hasMessage = true
			break
		}

		fieldStr, nextI, err := deserializeFieldStr(data, i)
		if err != nil {
			return fields, err
		}
		i = nextI

		switch fieldStr {
		case CommitFieldTree:
			if hasTree {
				return fields, ErrInvalidCommitFormat
			}
			hash, nextI, err := deserializeTreeOrParent(data, i)
			if err != nil {
				return fields, err
			}
			fields.TreeHash = hash
			hasTree = true
			i = nextI

		case CommitFieldParent:
			hash, nextI, err := deserializeTreeOrParent(data, i)
			if err != nil {
				return fields, err
			}
			fields.ParentHashes = append(fields.ParentHashes, hash)
			i = nextI

		case CommitFieldAuthor:
			if hasAuthor {
				return fields, ErrInvalidCommitFormat
			}
			author, nextI, err := deserializeIdentity(data, i)
			if err != nil {
				return fields, err
			}
			fields.Author = author
			hasAuthor = true
			i = nextI

		case CommitFieldCommitter:
			if hasCommitter {
				return fields, ErrInvalidCommitFormat
			}
			committer, nextI, err := deserializeIdentity(data, i)
			if err != nil {
				return fields, err
			}
			fields.Committer = committer
			hasCommitter = true
			i = nextI
		}
	}

	if !hasTree || !hasAuthor || !hasCommitter || !hasMessage {
		return fields, ErrInvalidCommitFormat
	}
	return fields, nil
}

// deserializeFieldStr parses a commit header key from data[start:] and returns
// the field name and the next index after the separating space.
func deserializeFieldStr(data []byte, start int) (string, int, error) {
	i := start
	for i < len(data) && data[i] != ' ' {
		i++
	}

	if i >= len(data) {
		return "", i, ErrInvalidCommitFormat
	}

	fieldStr := string(data[start:i])
	if ok := isValidCommitField(fieldStr); !ok {
		return "", i, ErrInvalidCommitField
	}
	return fieldStr, i + 1, nil
}

// deserializeTreeOrParent parses a hash value terminated by newline from data[start:].
func deserializeTreeOrParent(data []byte, start int) (Hash, int, error) {
	i := start
	for i < len(data) && data[i] != '\n' {
		i++
	}
	if i >= len(data) {
		return Hash{}, i, ErrInvalidCommitFormat
	}

	hexHash := string(data[start:i])
	hash, err := NewHashFromHex(hexHash)
	if err != nil {
		return Hash{}, i, err
	}
	return hash, i + 1, nil
}

// deserializeIdentity parses an identity line segment in
// "<name> <email> <timestamp> <timezone>" form and returns a validated Identity.
func deserializeIdentity(data []byte, start int) (Identity, int, error) {
	i := start
	lineEnd := i
	for lineEnd < len(data) && data[lineEnd] != '\n' {
		lineEnd++
	}
	if lineEnd >= len(data) {
		return Identity{}, i, ErrInvalidCommitFormat
	}

	emailStart := i
	for emailStart < lineEnd && data[emailStart] != '<' {
		emailStart++
	}
	if emailStart >= lineEnd {
		return Identity{}, i, ErrInvalidCommitFormat
	}

	nameBytes := bytes.TrimSpace(data[i:emailStart])
	name := string(nameBytes)

	emailEnd := emailStart + 1
	for emailEnd < lineEnd && data[emailEnd] != '>' {
		emailEnd++
	}
	if emailEnd >= lineEnd {
		return Identity{}, i, ErrInvalidCommitFormat
	}

	email := string(data[emailStart+1 : emailEnd])

	i = emailEnd + 1
	for i < lineEnd && data[i] == ' ' {
		i++
	}

	timestampStart := i
	for i < lineEnd && data[i] != ' ' {
		i++
	}
	if i >= lineEnd {
		return Identity{}, i, ErrInvalidCommitFormat
	}
	timestamp := string(data[timestampStart:i])

	i++
	if i >= lineEnd {
		return Identity{}, i, ErrInvalidCommitFormat
	}
	timezone := string(data[i:lineEnd])

	identity, err := NewIdentity(name, email, timestamp, timezone)
	if err != nil {
		return Identity{}, i, ErrInvalidCommitFormat
	}
	return identity, lineEnd + 1, nil
}

// cloneCommitFields deep-copies mutable fields to avoid aliasing shared slices.
func cloneCommitFields(fields CommitFields) CommitFields {
	cloned := fields
	cloned.ParentHashes = append([]Hash(nil), fields.ParentHashes...)
	return cloned
}

// isValidCommitField reports whether field is a supported commit header key.
func isValidCommitField(field string) bool {
	switch field {
	case CommitFieldTree,
		CommitFieldParent,
		CommitFieldAuthor,
		CommitFieldCommitter:
		return true
	}
	return false
}
