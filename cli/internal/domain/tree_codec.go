package domain

import (
	"bytes"
	"fmt"
	"strings"
)

// EncodeTree encodes tree using the canonical Gel tree-body format.
func EncodeTree(tree *Tree) ([]byte, error) {
	if tree == nil {
		return nil, fmt.Errorf(
			"%w: tree is nil",
			ErrInvalidTree,
		)
	}
	if err := validateTreeEntries(tree.entries); err != nil {
		return nil, fmt.Errorf(
			"%w: validate entries: %w",
			ErrInvalidTree,
			err,
		)
	}
	return encodeTreeEntries(tree.entries), nil
}

// DecodeTree decodes and validates a canonical Gel tree body.
//
// Encoded entries must already be in canonical order.
func DecodeTree(data []byte) (*Tree, error) {
	entries, err := decodeTreeEntries(data)
	if err != nil {
		return nil, fmt.Errorf(
			"%w: decode entries: %w",
			ErrInvalidTree,
			err,
		)
	}
	return &Tree{
		entries: entries,
	}, nil
}

func encodeTreeEntries(entries []TreeEntry) []byte {
	encoded := make([]byte, 0)
	for _, entry := range entries {
		encoded = append(
			encoded,
			[]byte(fmt.Sprintf(
				"%s %s\x00%s",
				entry.mode.String(),
				entry.name,
				entry.hash.String(),
			))...,
		)
	}
	return encoded
}

func decodeTreeEntries(body []byte) ([]TreeEntry, error) {
	var entries []TreeEntry
	seenNames := make(map[string]struct{})
	previousSortKey := ""
	hasPrevious := false

	for offset := 0; offset < len(body); {
		entryOffset := offset
		spaceIndex := bytes.IndexByte(body[offset:], ' ')
		if spaceIndex == -1 {
			return nil, fmt.Errorf(
				"entry at offset %d: missing mode separator",
				entryOffset,
			)
		}

		modeEnd := offset + spaceIndex
		modeText := string(body[offset:modeEnd])
		mode, err := ParseFileMode(modeText)
		if err != nil {
			return nil, fmt.Errorf(
				"entry at offset %d: %w",
				entryOffset,
				err,
			)
		}

		offset = modeEnd + 1
		nameStart := offset
		nulIndex := bytes.IndexByte(body[offset:], 0)
		if nulIndex == -1 {
			return nil, fmt.Errorf(
				"entry at offset %d: missing name terminator",
				entryOffset,
			)
		}

		nameEnd := offset + nulIndex
		name := string(body[nameStart:nameEnd])
		offset = nameEnd + 1

		if len(body)-offset < HashByteLength {
			return nil, fmt.Errorf(
				"entry %q at offset %d: truncated hash: got %d bytes, want %d",
				name,
				entryOffset,
				len(body)-offset,
				HashByteLength,
			)
		}

		hash, err := NewHash(body[offset : offset+HashByteLength])
		if err != nil {
			return nil, fmt.Errorf(
				"entry %q at offset %d: parse hash: %w",
				name,
				entryOffset,
				err,
			)
		}

		offset += HashByteLength

		entry := TreeEntry{
			mode: mode,
			hash: hash,
			name: name,
		}
		if err := validateTreeEntry(entry); err != nil {
			return nil, fmt.Errorf(
				"entry at offset %d: %w",
				entryOffset,
				err,
			)
		}
		if _, exists := seenNames[name]; exists {
			return nil, fmt.Errorf(
				"duplicate entry name %q",
				name,
			)
		}

		seenNames[name] = struct{}{}

		sortKey := treeEntrySortKey(entry.name, entry.mode.IsDirectory())
		if hasPrevious && strings.Compare(previousSortKey, sortKey) > 0 {
			return nil, fmt.Errorf(
				"entries are not in canonical order: %q before %q",
				previousSortKey,
				sortKey,
			)
		}

		previousSortKey = sortKey
		hasPrevious = true
		entries = append(entries, entry)
	}
	return entries, nil
}
