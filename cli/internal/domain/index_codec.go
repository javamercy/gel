package domain

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math"
	"time"
)

// EncodeIndex encodes index using Gel Index Format version 1.
func EncodeIndex(index *Index) ([]byte, error) {
	if index == nil {
		return nil, fmt.Errorf("%w: index is nil", ErrInvalidIndex)
	}
	if len(index.entries) > math.MaxUint32 {
		return nil, fmt.Errorf(
			"%w: entry count %d exceeds maximum %d",
			ErrInvalidIndex,
			len(index.entries),
			uint64(math.MaxUint32),
		)
	}
	if err := validateIndexEntries(index.entries); err != nil {
		return nil, fmt.Errorf(
			"%w: validate entries: %w",
			ErrInvalidIndex,
			err,
		)
	}

	encoded := make([]byte, 0)
	encoded = encodeIndexHeader(encoded, uint32(len(index.entries)))
	encoded = encodeIndexEntries(encoded, index.entries)
	checksum := sha256.Sum256(encoded)
	return append(encoded, checksum[:]...), nil
}

// DecodeIndex decodes and validates Gel Index Format version 1 data.
func DecodeIndex(data []byte) (*Index, error) {
	if len(data) < minIndexSize {
		return nil, fmt.Errorf(
			"%w: got %d bytes, want at least %d",
			ErrInvalidIndex,
			len(data),
			minIndexSize,
		)
	}

	entryCount, err := decodeIndexHeader(data[:indexHeaderSize])
	if err != nil {
		return nil, fmt.Errorf(
			"decode index header: %w",
			err,
		)
	}

	payload := data[:len(data)-indexChecksumSize]
	storedChecksum := data[len(data)-indexChecksumSize:]
	computedChecksum := sha256.Sum256(payload)
	if !bytes.Equal(storedChecksum, computedChecksum[:]) {
		return nil, fmt.Errorf("%w: checksum mismatch", ErrInvalidIndex)
	}

	entriesData := data[indexHeaderSize : len(data)-indexChecksumSize]
	maxPossibleEntries := len(entriesData) / minIndexEntrySize
	if uint64(entryCount) > uint64(maxPossibleEntries) {
		return nil, fmt.Errorf("%w: entry count declares more entries than available data", ErrInvalidIndex)
	}

	entries := make([]IndexEntry, 0, entryCount)
	offset := 0
	for i := uint32(0); i < entryCount; i++ {
		if offset >= len(entriesData) {
			return nil, fmt.Errorf(
				"%w: entry count declares more entries than available data",
				ErrInvalidIndex,
			)
		}
		entry, totalSize, err := decodeIndexEntry(entriesData[offset:])
		if err != nil {
			return nil, fmt.Errorf(
				"%w: entry %d at byte offset %d: %w",
				ErrInvalidIndex,
				i,
				indexHeaderSize+offset,
				err,
			)
		}
		if len(entries) > 0 {
			if err := validateConsecutiveIndexEntries(entries[len(entries)-1], entry); err != nil {
				return nil, fmt.Errorf(
					"%w: entries %d and %d: %w",
					ErrInvalidIndex,
					i-1,
					i,
					err,
				)
			}
		}
		offset += totalSize
		entries = append(entries, entry)
	}
	if offset != len(entriesData) {
		remaining := len(entriesData) - offset
		return nil, fmt.Errorf(
			"%w: %d unexpected bytes after final entry",
			ErrInvalidIndex,
			remaining,
		)
	}
	return &Index{entries: entries}, nil
}

func encodeIndexHeader(dst []byte, entryCount uint32) []byte {
	dst = append(dst, indexSignature...)
	dst = binary.BigEndian.AppendUint32(dst, indexVersion)
	dst = binary.BigEndian.AppendUint32(dst, entryCount)
	return dst
}

func decodeIndexHeader(data []byte) (entryCount uint32, err error) {
	if len(data) != indexHeaderSize {
		return 0, fmt.Errorf(
			"header has %d bytes, want %d",
			len(data),
			indexHeaderSize,
		)
	}

	signature := string(data[:indexHeaderSignatureSize])
	if signature != indexSignature {
		return 0, fmt.Errorf(
			"signature %q, want %q",
			signature,
			indexSignature,
		)
	}

	version := binary.BigEndian.Uint32(data[indexHeaderSignatureSize : indexHeaderSignatureSize+indexHeaderVersionSize])
	if version != indexVersion {
		return 0, fmt.Errorf(
			"version %d, want %d",
			version,
			indexVersion,
		)
	}

	entryCount = binary.BigEndian.Uint32(data[indexHeaderSignatureSize+indexHeaderVersionSize:])
	return entryCount, nil
}

func encodeIndexEntries(dst []byte, entries []IndexEntry) []byte {
	for _, entry := range entries {
		dst = encodeIndexEntry(dst, entry)
	}
	return dst
}

func encodeIndexEntry(dst []byte, entry IndexEntry) []byte {
	pathText := entry.path.String()
	flags := encodeIndexFlags(len(pathText), entry.stage)
	size := indexEntryFixedSize + len(pathText) + indexEntryPathNULTerminatorSize
	padding := computeIndexPadding(size)

	changeTimeSeconds, changeTimeNanoseconds := encodeIndexTime(entry.changeTime)
	dst = binary.BigEndian.AppendUint64(dst, uint64(changeTimeSeconds))
	dst = binary.BigEndian.AppendUint32(dst, changeTimeNanoseconds)

	modTimeSeconds, modTimeNanoseconds := encodeIndexTime(entry.modTime)
	dst = binary.BigEndian.AppendUint64(dst, uint64(modTimeSeconds))
	dst = binary.BigEndian.AppendUint32(dst, modTimeNanoseconds)

	dst = binary.BigEndian.AppendUint64(dst, entry.deviceID)
	dst = binary.BigEndian.AppendUint64(dst, entry.inode)
	dst = binary.BigEndian.AppendUint32(dst, uint32(entry.mode))
	dst = binary.BigEndian.AppendUint32(dst, entry.userID)
	dst = binary.BigEndian.AppendUint32(dst, entry.groupID)
	dst = binary.BigEndian.AppendUint64(dst, entry.size)
	dst = append(dst, entry.hash[:]...)
	dst = binary.BigEndian.AppendUint16(dst, flags)
	dst = append(dst, pathText...)
	dst = append(dst, 0)
	dst = appendIndexPadding(dst, padding)
	return dst
}

func decodeIndexEntry(data []byte) (entry IndexEntry, totalSize int, err error) {
	if len(data) < minIndexEntrySize {
		return IndexEntry{}, 0, fmt.Errorf(
			"entry has %d bytes, want at least %d",
			len(data),
			minIndexEntrySize,
		)
	}

	offset := 0

	changeTimeSeconds := int64(binary.BigEndian.Uint64(data[offset : offset+indexEntryChangeTimeSecondsSize]))
	offset += indexEntryChangeTimeSecondsSize

	changeTimeNanoseconds := binary.BigEndian.Uint32(data[offset : offset+indexEntryChangeTimeNanosSize])
	if changeTimeNanoseconds >= nanosecondsPerSecond {
		return IndexEntry{}, 0, fmt.Errorf(
			"change-time nanoseconds %d must be less than %d",
			changeTimeNanoseconds,
			nanosecondsPerSecond,
		)
	}

	offset += indexEntryChangeTimeNanosSize

	changeTime := decodeIndexTime(changeTimeSeconds, changeTimeNanoseconds)

	modTimeSeconds := int64(binary.BigEndian.Uint64(data[offset : offset+indexEntryModTimeSecondsSize]))
	offset += indexEntryModTimeSecondsSize

	modTimeNanoseconds := binary.BigEndian.Uint32(data[offset : offset+indexEntryModTimeNanosSize])
	if modTimeNanoseconds >= nanosecondsPerSecond {
		return IndexEntry{}, 0, fmt.Errorf(
			"modification-time nanoseconds %d must be less than %d",
			modTimeNanoseconds,
			nanosecondsPerSecond,
		)
	}

	offset += indexEntryModTimeNanosSize

	modTime := decodeIndexTime(modTimeSeconds, modTimeNanoseconds)

	deviceID := binary.BigEndian.Uint64(data[offset : offset+indexEntryDeviceIDSize])
	offset += indexEntryDeviceIDSize

	inode := binary.BigEndian.Uint64(data[offset : offset+indexEntryInodeSize])
	offset += indexEntryInodeSize

	mode := binary.BigEndian.Uint32(data[offset : offset+indexEntryModeSize])
	offset += indexEntryModeSize

	fileMode, err := NewFileMode(mode)
	if err != nil {
		return IndexEntry{}, 0, fmt.Errorf(
			"decode index entry mode %#o: %w",
			mode,
			err,
		)
	}

	userID := binary.BigEndian.Uint32(data[offset : offset+indexEntryUserIDSize])
	offset += indexEntryUserIDSize

	groupID := binary.BigEndian.Uint32(data[offset : offset+indexEntryGroupIDSize])
	offset += indexEntryGroupIDSize

	fileSize := binary.BigEndian.Uint64(data[offset : offset+indexEntryFileSizeSize])
	offset += indexEntryFileSizeSize

	hashBytes := data[offset : offset+indexEntryHashSize]
	hash, err := NewHash(hashBytes)
	if err != nil {
		return IndexEntry{}, 0, err
	}

	offset += indexEntryHashSize

	flags := binary.BigEndian.Uint16(data[offset : offset+indexEntryFlagsSize])
	encodedPathLength, _, err := decodeIndexFlags(flags)
	if err != nil {
		return IndexEntry{}, 0, err
	}

	offset += indexEntryFlagsSize

	// BUG: For encoded path lengths below 0xFFF,
	// scanning all remaining entry data for NUL can consume bytes from a later entry;
	// use the encoded length as the exact terminator position and scan only when the encoded length is capped at 0xFFF.
	nulIndex := bytes.IndexByte(data[offset:], 0)
	if nulIndex == -1 {
		return IndexEntry{}, 0, fmt.Errorf("missing NUL terminator")
	}

	pathBytes := data[offset : offset+nulIndex]
	if encodedPathLength != min(maxEncodedIndexPathLength, len(pathBytes)) {
		return IndexEntry{}, 0, fmt.Errorf("invalid path length")
	}

	normPath, err := ParseNormalizedPath(string(pathBytes))
	if err != nil {
		return IndexEntry{}, 0, fmt.Errorf(
			"decode index-entry path: %w",
			err,
		)
	}

	offset += nulIndex + 1

	padding := computeIndexPadding(indexEntryFixedSize + len(pathBytes) + indexEntryPathNULTerminatorSize)
	if len(data[offset:]) < padding {
		return IndexEntry{}, 0, fmt.Errorf("not enough data for padding")
	}
	for paddingIndex, value := range data[offset : offset+padding] {
		if value != 0 {
			return IndexEntry{}, 0, fmt.Errorf(
				"padding byte at entry offset %d is %#02x, want zero",
				offset+paddingIndex,
				value,
			)
		}
	}

	offset += padding

	stat := NewFileStat(deviceID, inode, userID, groupID, fileMode, fileSize, changeTime, modTime)
	entry, err = NewIndexEntryFromFileStat(
		normPath,
		hash,
		stat,
	)
	if err != nil {
		return IndexEntry{}, 0, fmt.Errorf(
			"construct decoded index entry: %w",
			err,
		)
	}
	return entry, offset, nil
}

func encodeIndexFlags(pathLength int, stage IndexStage) uint16 {
	encodedPathLength := uint16(min(pathLength, maxEncodedIndexPathLength))
	return encodedPathLength | uint16(stage)<<indexStageShift
}

func decodeIndexFlags(flags uint16) (pathLength int, stage IndexStage, err error) {
	if flags&indexReservedMask != 0 {
		return 0, 0, fmt.Errorf(
			"reserved flag bits are set: %#04x",
			flags&indexReservedMask,
		)
	}

	pathLength = int(flags & indexPathLengthMask)
	stage = IndexStage((flags & indexStageMask) >> indexStageShift)
	return pathLength, stage, nil
}

func encodeIndexTime(t time.Time) (seconds int64, nanoseconds uint32) {
	if t.IsZero() {
		return 0, 0
	}

	t = t.UTC()
	return t.Unix(), uint32(t.Nanosecond())
}

func decodeIndexTime(seconds int64, nanoseconds uint32) time.Time {
	if seconds == 0 && nanoseconds == 0 {
		return time.Time{}
	}
	return time.Unix(seconds, int64(nanoseconds)).UTC()
}

func computeIndexPadding(size int) int {
	return (indexEntryAlignment - (size % indexEntryAlignment)) % indexEntryAlignment
}

func appendIndexPadding(dst []byte, count int) []byte {
	for range count {
		dst = append(dst, 0)
	}
	return dst
}
