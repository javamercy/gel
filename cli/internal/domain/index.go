package domain

import (
	"cmp"
	"fmt"
	"slices"
	"sort"
	"strings"
)

// Index stores staged entries in canonical path-and-stage order.
// The zero value is a valid empty index.
type Index struct {
	entries []IndexEntry
}

// NewIndexFromEntries constructs an index from entries.
//
// The supplied slice is copied, sorted into canonical order, and validated.
func NewIndexFromEntries(entries []IndexEntry) (*Index, error) {
	entries = slices.Clone(entries)
	slices.SortFunc(entries, compareIndexEntries)

	if err := validateIndexEntries(entries); err != nil {
		return nil, fmt.Errorf(
			"create index from entries: %w",
			err,
		)
	}
	return &Index{entries: entries}, nil
}

// Entries returns a copy of all entries in canonical order.
func (i *Index) Entries() []IndexEntry {
	return slices.Clone(i.entries)
}

// AddEntry inserts a new entry in canonical path-and-stage order.
//
// It returns an error if entry is invalid, its path-and-stage identity already
// exists, or inserting it would violate the index stage-group invariants.
func (i *Index) AddEntry(entry IndexEntry) error {
	if err := validateIndexEntry(entry); err != nil {
		return fmt.Errorf(
			"add entry: %w",
			err,
		)
	}

	position, found := i.findEntryPosition(entry)
	if found {
		return fmt.Errorf(
			"add entry: entry for path %q at stage %d already exists",
			entry.path.String(),
			entry.stage,
		)
	}
	if err := i.validateIndexEntryInsertion(entry, position); err != nil {
		return fmt.Errorf(
			"add entry: %w",
			err,
		)
	}

	i.entries = slices.Insert(i.entries, position, entry)
	return nil
}

// UpdateEntry replaces an existing entry with the same path and stage.
//
// The returned boolean reports whether the entry existed.
func (i *Index) UpdateEntry(entry IndexEntry) (bool, error) {
	if err := validateIndexEntry(entry); err != nil {
		return false, fmt.Errorf(
			"update entry: %w",
			err,
		)
	}

	position, found := i.findEntryPosition(entry)
	if !found {
		return false, nil
	}

	i.entries[position] = entry
	return true, nil
}

// SetEntry replaces an existing entry with the same path and stage,
// or inserts a new entry if it does not exist.
func (i *Index) SetEntry(entry IndexEntry) error {
	if err := validateIndexEntry(entry); err != nil {
		return fmt.Errorf(
			"set entry: %w",
			err,
		)
	}

	position, found := i.findEntryPosition(entry)
	if found {
		i.entries[position] = entry
		return nil
	}

	if err := i.validateIndexEntryInsertion(entry, position); err != nil {
		return fmt.Errorf(
			"set entry: %w",
			err,
		)
	}

	i.entries = slices.Insert(i.entries, position, entry)
	return nil
}

// RemoveEntry removes an existing entry with the same identity.
//
// The returned boolean reports whether the entry existed.
func (i *Index) RemoveEntry(entry IndexEntry) bool {
	position, found := i.findEntryPosition(entry)
	if !found {
		return false
	}

	i.entries = slices.Delete(i.entries, position, position+1)
	return true
}

// FindEntry returns the entry with the same identity if it exists.
func (i *Index) FindEntry(entry IndexEntry) (IndexEntry, bool) {
	position, found := i.findEntryPosition(entry)
	if found {
		return i.entries[position], true
	}
	return IndexEntry{}, false
}

func (i *Index) findEntryPosition(entry IndexEntry) (position int, found bool) {
	position = sort.Search(
		len(i.entries),
		func(j int) bool {
			return compareIndexEntries(i.entries[j], entry) >= 0
		},
	)
	found = position < len(i.entries) &&
		compareIndexEntries(i.entries[position], entry) == 0
	return position, found
}

func (i *Index) validateIndexEntryInsertion(entry IndexEntry, position int) error {
	if position > 0 {
		if err := validateConsecutiveIndexEntries(
			i.entries[position-1],
			entry,
		); err != nil {
			return err
		}
	}
	if position < len(i.entries) {
		if err := validateConsecutiveIndexEntries(
			entry,
			i.entries[position],
		); err != nil {
			return err
		}
	}
	return nil
}

func validateConsecutiveIndexEntries(prev, curr IndexEntry) error {
	result := strings.Compare(
		prev.path.String(),
		curr.path.String(),
	)
	if result == 0 {
		if prev.stage == curr.stage {
			return fmt.Errorf(
				"duplicate entry for path %q at stage %d",
				prev.path.String(),
				prev.stage,
			)
		}
		if prev.stage == IndexStageNormal || curr.stage == IndexStageNormal {
			return fmt.Errorf(
				"path %q mixes stage zero with non-zero stages",
				prev.path.String(),
			)
		}
		if prev.stage > curr.stage {
			return fmt.Errorf(
				"path %q has stages %d and %d out of order",
				prev.path.String(),
				prev.stage,
				curr.stage,
			)
		}
	} else if result > 0 {
		return fmt.Errorf(
			"path %q appears before earlier path %q",
			prev.path.String(),
			curr.path.String(),
		)
	}
	return nil
}

func compareIndexEntries(a, b IndexEntry) int {
	if result := strings.Compare(
		a.path.String(),
		b.path.String(),
	); result != 0 {
		return result
	}
	return cmp.Compare(a.stage, b.stage)
}
