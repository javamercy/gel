package domain

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"unicode/utf8"
)

// ErrInvalidCommit indicates invalid commit state or malformed commit body data.
var ErrInvalidCommit = errors.New("invalid commit")

const (
	commitHeaderTree      = "tree"
	commitHeaderParent    = "parent"
	commitHeaderAuthor    = "author"
	commitHeaderCommitter = "committer"
)

// EncodeCommit encodes commit using the canonical Gel commit-body format.
func EncodeCommit(commit *Commit) ([]byte, error) {
	if err := validateCommit(commit); err != nil {
		return nil, fmt.Errorf(
			"%w: encode commit: %w",
			ErrInvalidCommit,
			err,
		)
	}

	encoded := make([]byte, 0)

	encoded = appendCommitHashHeader(encoded, commitHeaderTree, commit.treeHash)

	for _, parentHash := range commit.parentHashes {
		encoded = appendCommitHashHeader(encoded, commitHeaderParent, parentHash)
	}

	encoded = append(encoded, commitHeaderAuthor...)
	encoded = append(encoded, ' ')
	encoded = encodeCommitIdentity(encoded, commit.author)
	encoded = append(encoded, '\n')

	encoded = append(encoded, commitHeaderCommitter...)
	encoded = append(encoded, ' ')
	encoded = encodeCommitIdentity(encoded, commit.committer)

	encoded = append(encoded, '\n', '\n')

	encoded = append(encoded, commit.message...)
	return encoded, nil
}

// DecodeCommit decodes and validates a canonical Gel commit body.
func DecodeCommit(data []byte) (*Commit, error) {
	if !utf8.Valid(data) {
		return nil, fmt.Errorf(
			"%w: body is not valid UTF-8",
			ErrInvalidCommit,
		)
	}
	if bytes.IndexByte(data, 0) >= 0 {
		return nil, fmt.Errorf(
			"%w: body contains NUL",
			ErrInvalidCommit,
		)
	}
	if bytes.IndexByte(data, '\r') >= 0 {
		return nil, fmt.Errorf(
			"%w: body contains carriage return",
			ErrInvalidCommit,
		)
	}

	headerData, messageData, found := bytes.Cut(data, []byte("\n\n"))
	if !found {
		return nil, fmt.Errorf(
			"%w: missing message separator",
			ErrInvalidCommit,
		)
	}

	lines := bytes.Split(headerData, []byte("\n"))
	if len(lines) < 3 {
		return nil, fmt.Errorf(
			"%w: incomplete commit header",
			ErrInvalidCommit,
		)
	}

	position := 0
	treeValue, err := decodeCommitHeader(lines[position], commitHeaderTree)
	if err != nil {
		return nil, fmt.Errorf(
			"%w: tree header: %w",
			ErrInvalidCommit,
			err,
		)
	}

	treeHash, err := decodeCommitHash(treeValue)
	if err != nil {
		return nil, fmt.Errorf(
			"%w: tree header: %w",
			ErrInvalidCommit,
			err,
		)
	}

	position++

	var parentHashes []Hash
	seenParents := make(map[Hash]struct{})
	for position < len(lines) &&
		bytes.HasPrefix(lines[position], []byte(commitHeaderParent+" ")) {
		parentValue, err := decodeCommitHeader(lines[position], commitHeaderParent)
		if err != nil {
			return nil, fmt.Errorf(
				"%w: parent %d: %w",
				ErrInvalidCommit,
				len(parentHashes),
				err,
			)
		}

		parentHash, err := decodeCommitHash(parentValue)
		if err != nil {
			return nil, fmt.Errorf(
				"%w: parent %d: %w",
				ErrInvalidCommit,
				len(parentHashes),
				err,
			)
		}
		if _, exists := seenParents[parentHash]; exists {
			return nil, fmt.Errorf(
				"%w: duplicate parent hash %q",
				ErrInvalidCommit,
				parentHash.Hex(),
			)
		}

		seenParents[parentHash] = struct{}{}
		parentHashes = append(parentHashes, parentHash)
		position++
	}
	if position >= len(lines)-1 {
		return nil, fmt.Errorf(
			"%w: missing author header",
			ErrInvalidCommit,
		)
	}

	authorValue, err := decodeCommitHeader(lines[position], commitHeaderAuthor)
	if err != nil {
		return nil, fmt.Errorf(
			"%w: author header: %w",
			ErrInvalidCommit,
			err,
		)
	}

	author, err := decodeCommitIdentity(authorValue)
	if err != nil {
		return nil, fmt.Errorf(
			"%w: author: %w",
			ErrInvalidCommit,
			err,
		)
	}

	position++

	if position >= len(lines) {
		return nil, fmt.Errorf(
			"%w: missing committer header",
			ErrInvalidCommit,
		)
	}

	committerValue, err := decodeCommitHeader(
		lines[position],
		commitHeaderCommitter,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"%w: committer header: %w",
			ErrInvalidCommit,
			err,
		)
	}

	committer, err := decodeCommitIdentity(committerValue)
	if err != nil {
		return nil, fmt.Errorf(
			"%w: committer: %w",
			ErrInvalidCommit,
			err,
		)
	}

	position++

	if position != len(lines) {
		return nil, fmt.Errorf(
			"%w: unexpected header %q after committer",
			ErrInvalidCommit,
			lines[position],
		)
	}

	commit := &Commit{
		treeHash:     treeHash,
		parentHashes: parentHashes,
		author:       author,
		committer:    committer,
		message:      string(messageData),
	}
	if err := validateCommit(commit); err != nil {
		return nil, fmt.Errorf(
			"%w: decoded commit: %w",
			ErrInvalidCommit,
			err,
		)
	}
	return commit, nil
}

func appendCommitHashHeader(dst []byte, header string, hash Hash) []byte {
	dst = append(dst, header...)
	dst = append(dst, ' ')
	dst = append(dst, hash.Hex()...)
	dst = append(dst, '\n')
	return dst
}

func encodeCommitIdentity(dst []byte, identity CommitIdentity) []byte {
	dst = append(dst, identity.name...)
	dst = append(dst, ' ', '<')
	dst = append(dst, identity.email...)
	dst = append(dst, '>', ' ')
	dst = strconv.AppendInt(dst, identity.timestamp, 10)
	dst = append(dst, ' ')
	dst = appendCommitTimezone(dst, identity.timezoneOffsetMinutes)
	return dst
}

func appendCommitTimezone(dst []byte, offsetMinutes int) []byte {
	sign := byte('+')
	if offsetMinutes < 0 {
		sign = byte('-')
		offsetMinutes = -offsetMinutes
	}

	hours := offsetMinutes / 60
	minutes := offsetMinutes % 60
	return append(
		dst,
		sign,
		byte('0'+hours/10),
		byte('0'+hours%10),
		byte('0'+minutes/10),
		byte('0'+minutes%10),
	)
}

func decodeCommitHeader(line []byte, expectedHeader string) ([]byte, error) {
	prefix := expectedHeader + " "
	if !bytes.HasPrefix(line, []byte(prefix)) {
		return nil, fmt.Errorf(
			"got %q, want %q header",
			line,
			expectedHeader,
		)
	}

	value := line[len(prefix):]
	if len(value) == 0 {
		return nil, fmt.Errorf(
			"%s value is empty",
			expectedHeader,
		)
	}
	return value, nil
}

func decodeCommitHash(data []byte) (Hash, error) {
	expectedLength := HashByteLength * 2
	if len(data) != expectedLength {
		return Hash{}, fmt.Errorf(
			"hash has %d bytes, want %d",
			len(data),
			expectedLength,
		)
	}

	hash, err := NewHashFromHex(string(data))
	if err != nil {
		return Hash{}, fmt.Errorf(
			"decode hash %q: %w",
			data,
			err,
		)
	}
	if string(data) != hash.Hex() {
		return Hash{}, fmt.Errorf(
			"hash %q is not canonical lowercase hexadecimal",
			data,
		)
	}
	if hash.IsZero() {
		return Hash{}, fmt.Errorf(
			"hash cannot be zero",
		)
	}
	return hash, nil
}

func decodeCommitIdentity(data []byte) (CommitIdentity, error) {
	timezoneSeparator := bytes.LastIndexByte(data, ' ')
	if timezoneSeparator <= 0 {
		return CommitIdentity{}, fmt.Errorf("missing timestamp-timezone separator")
	}

	timestampSeparator := bytes.LastIndexByte(data[:timezoneSeparator], ' ')
	if timestampSeparator <= 0 {
		return CommitIdentity{}, fmt.Errorf("missing identity-timestamp separator")
	}

	nameAndEmail := data[:timestampSeparator]
	timestampData := data[timestampSeparator+1 : timezoneSeparator]
	timezoneData := data[timezoneSeparator+1:]

	if len(nameAndEmail) == 0 || nameAndEmail[len(nameAndEmail)-1] != '>' {
		return CommitIdentity{}, fmt.Errorf("missing closing email delimiter")
	}

	emailSeparator := bytes.LastIndex(nameAndEmail, []byte(" <"))
	if emailSeparator <= 0 {
		return CommitIdentity{}, fmt.Errorf("missing name-email separator")
	}

	name := string(nameAndEmail[:emailSeparator])
	email := string(nameAndEmail[emailSeparator+2 : len(nameAndEmail)-1])

	timestamp, err := decodeCommitTimestamp(timestampData)
	if err != nil {
		return CommitIdentity{}, fmt.Errorf(
			"timestamp: %w",
			err,
		)
	}

	timezoneOffsetMinutes, err := decodeCommitTimezone(timezoneData)
	if err != nil {
		return CommitIdentity{}, fmt.Errorf(
			"timezone: %w",
			err,
		)
	}

	identity := CommitIdentity{
		name:                  name,
		email:                 email,
		timestamp:             timestamp,
		timezoneOffsetMinutes: timezoneOffsetMinutes,
	}
	if err := validateCommitIdentity(identity); err != nil {
		return CommitIdentity{}, err
	}
	return identity, nil
}

func decodeCommitTimestamp(data []byte) (int64, error) {
	if len(data) == 0 {
		return 0, fmt.Errorf(
			"timestamp is empty",
		)
	}

	timestamp, err := strconv.ParseInt(string(data), 10, 64)
	if err != nil {
		return 0, fmt.Errorf(
			"invalid timestamp %q",
			data,
		)
	}
	if strconv.FormatInt(timestamp, 10) != string(data) {
		return 0, fmt.Errorf(
			"timestamp %q is not canonical",
			data,
		)
	}
	return timestamp, nil
}

func decodeCommitTimezone(data []byte) (int, error) {
	if len(data) != 5 {
		return 0, fmt.Errorf(
			"timezone must contain exactly five bytes",
		)
	}

	sign := data[0]
	if sign != '+' && sign != '-' {
		return 0, fmt.Errorf(
			"timezone must begin with '+' or '-'",
		)
	}
	for _, value := range data[1:] {
		if value < '0' || value > '9' {
			return 0, fmt.Errorf(
				"timezone %q contains non-decimal digits",
				data,
			)
		}
	}

	hours := int(data[1]-'0')*10 + int(data[2]-'0')
	minutes := int(data[3]-'0')*10 + int(data[4]-'0')

	if hours > 23 {
		return 0, fmt.Errorf(
			"timezone hour %d is outside range 00-23",
			hours,
		)
	}

	if minutes > 59 {
		return 0, fmt.Errorf(
			"timezone minute %d is outside range 00-59",
			minutes,
		)
	}

	offset := hours*60 + minutes
	if sign == '-' {
		if offset == 0 {
			return 0, fmt.Errorf(
				"timezone -0000 is not canonical",
			)
		}
		offset = -offset
	}
	return offset, nil
}
