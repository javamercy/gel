package domain

import (
	"crypto/sha256"
	"errors"
)

// ErrInvalidIndex indicates invalid index state or malformed encoded index data.
var ErrInvalidIndex = errors.New("invalid index")

const (
	indexSignature        = "GIDX"
	indexVersion   uint32 = 1
)

const (
	indexHeaderSignatureSize  = len(indexSignature)
	indexHeaderVersionSize    = 4
	indexHeaderEntryCountSize = 4

	indexHeaderSize = indexHeaderSignatureSize +
		indexHeaderVersionSize +
		indexHeaderEntryCountSize
)

const (
	indexChecksumSize = sha256.Size
	minIndexSize      = indexHeaderSize + indexChecksumSize
)

const nanosecondsPerSecond = 1_000_000_000

const (
	indexEntryChangeTimeSecondsSize = 8
	indexEntryChangeTimeNanosSize   = 4
	indexEntryModTimeSecondsSize    = 8
	indexEntryModTimeNanosSize      = 4
	indexEntryDeviceIDSize          = 8
	indexEntryInodeSize             = 8
	indexEntryModeSize              = 4
	indexEntryUserIDSize            = 4
	indexEntryGroupIDSize           = 4
	indexEntryFileSizeSize          = 8
	indexEntryHashSize              = sha256.Size
	indexEntryFlagsSize             = 2

	indexEntryFixedSize = indexEntryChangeTimeSecondsSize +
		indexEntryChangeTimeNanosSize +
		indexEntryModTimeSecondsSize +
		indexEntryModTimeNanosSize +
		indexEntryDeviceIDSize +
		indexEntryInodeSize +
		indexEntryModeSize +
		indexEntryUserIDSize +
		indexEntryGroupIDSize +
		indexEntryFileSizeSize +
		indexEntryHashSize +
		indexEntryFlagsSize
)

const (
	indexEntryPathNULTerminatorSize = 1
	indexEntryAlignment             = 8
	minIndexEntryPathSize           = 1

	minIndexEntrySize = indexEntryFixedSize +
		minIndexEntryPathSize +
		indexEntryPathNULTerminatorSize
)

const (
	indexPathLengthMask uint16 = 0x0FFF
	indexStageMask      uint16 = 0x3000
	indexReservedMask   uint16 = 0xC000

	indexStageShift           = 12
	maxEncodedIndexPathLength = 0x0FFF
)
