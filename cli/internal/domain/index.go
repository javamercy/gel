package domain

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"time"
)

// Index format constants.
const (
	// IndexSignature is the 4-byte signature stored at the start of index files.
	IndexSignature = "DIRC"

	// IndexVersion is the supported on-disk index format version.
	IndexVersion = 2
)

// Index-specific errors.
var (
	ErrIndexTooShort         = errors.New("index file is too short: minimum 12 bytes required for header")
	ErrInvalidIndexSignature = errors.New("invalid index signature: expected 'DIRC', file may be corrupted")
	ErrTruncatedEntryData    = errors.New("index file truncated: not enough data to read all entries")
	ErrIncorrectChecksumSize = errors.New("invalid index checksum: expected 32 bytes at end of file")
	ErrChecksumMismatch      = errors.New("index checksum verification failed: file may be corrupted")
	ErrEntryDataTooShort     = errors.New("index entry is incomplete: minimum 74 bytes required")
	ErrPathNotNullTerminated = errors.New("index entry path is malformed: missing null terminator")
)

// Index file header and entry size constants.
const (
	IndexHeaderSignatureSize  = 4
	IndexHeaderVersionSize    = 4
	IndexHeaderNumEntriesSize = 4
	IndexHeaderSize           = IndexHeaderSignatureSize + IndexHeaderVersionSize + IndexHeaderNumEntriesSize
)

// Index checksum and padding constants.
const (
	IndexChecksumSize = 32
	PaddingAlignment  = 8
)

// Index entry field size constants (in bytes).
// Note: Device, Inode, and Size use uint64 (8 bytes) to support modern filesystems.
// Time fields use int64 (8 bytes) for seconds plus uint32 (4 bytes) for nanoseconds.
const (
	IndexEntryTimeSize              = 12 // 8 bytes (seconds) + 4 bytes (nanoseconds)
	IndexEntryDeviceSize            = 8  // uint64
	IndexEntryInodeSize             = 8  // uint64
	IndexEntryModeSize              = 4  // uint32
	IndexEntryUserIDSize            = 4  // uint32
	IndexEntryGroupIDSize           = 4  // uint32
	IndexEntrySizeFieldSize         = 8  // uint64
	IndexEntryHashSize              = HashByteLength
	IndexEntryFlagsSize             = 2
	IndexEntryPathNullTerminateSize = 1
	IndexEntryHashOffset            = 2*IndexEntryTimeSize + IndexEntryDeviceSize + IndexEntryInodeSize + IndexEntryModeSize + IndexEntryUserIDSize + IndexEntryGroupIDSize + IndexEntrySizeFieldSize
	IndexEntryFlagsOffset           = IndexEntryHashOffset + IndexEntryHashSize
	IndexEntryFixedSize             = IndexEntryFlagsOffset + IndexEntryFlagsSize
)

// Index flags constants.
const (
	MaxPathLength = 0xFFF // maximum path length encoded in flags (12 bits)
	StageMask     = 0x3   // mask for stage bits in flags
	StageShift    = 12    // bit offset for stage in flags
)

// IndexHeader stores the on-disk index header.
type IndexHeader struct {
	// Signature is the 4-byte file signature (e.g., "DIRC").
	Signature [4]byte
	// Version is the index format version.
	Version uint32
	// NumEntries is the number of entries in the index.
	NumEntries uint32
}

// NewIndexHeader returns an IndexHeader with the supplied values.
func NewIndexHeader(signature [4]byte, version uint32, numEntries uint32) IndexHeader {
	return IndexHeader{
		Signature:  signature,
		Version:    version,
		NumEntries: numEntries,
	}
}

// IndexEntry stores metadata for one tracked path.
type IndexEntry struct {
	// Path is the normalized relative path from the repository root.
	Path NormalizedPath
	// Hash is the SHA-256 content hash of the file/blob.
	Hash Hash
	// Size is the file size in bytes.
	Size uint64
	// Mode is the file mode (permissions and type).
	Mode uint32
	// Device is the device ID containing the file.
	Device uint64
	// Inode is the file's inode number.
	Inode uint64
	// UserID is the file owner's user ID.
	UserID uint32
	// GroupID is the file owner's group ID.
	GroupID uint32
	// Flags contains path length (lower 12 bits) and stage (upper 4 bits).
	Flags uint16
	// ChangedTime is the last metadata change time (ctime).
	ChangedTime time.Time
	// ModifiedTime is the last content modification time (mtime).
	ModifiedTime time.Time
}

// NewEmptyIndexEntry creates an index entry with zero values for the given path, hash, and mode.
func NewEmptyIndexEntry(path NormalizedPath, hash Hash, mode uint32) *IndexEntry {
	return &IndexEntry{
		Path:         path,
		Hash:         hash,
		Size:         0,
		Mode:         mode,
		Device:       0,
		Inode:        0,
		UserID:       0,
		GroupID:      0,
		Flags:        0,
		ChangedTime:  time.Time{},
		ModifiedTime: time.Time{},
	}
}

// NewIndexEntry creates a fully populated index entry with all metadata.
func NewIndexEntry(
	path NormalizedPath,
	hash Hash,
	size uint64,
	mode uint32,
	device uint64,
	inode uint64,
	userID uint32,
	groupID uint32,
	flags uint16,
	createdTime time.Time,
	updatedTime time.Time,
) *IndexEntry {
	entry := IndexEntry{
		Path:         path,
		Hash:         hash,
		Size:         size,
		Mode:         mode,
		Device:       device,
		Inode:        inode,
		UserID:       userID,
		GroupID:      groupID,
		Flags:        flags,
		ChangedTime:  createdTime,
		ModifiedTime: updatedTime,
	}
	return &entry
}

// GetStage returns the staging stage encoded in the entry flags.
func (e *IndexEntry) GetStage() uint16 {
	return (e.Flags >> StageShift) & StageMask
}

// serialize converts the index entry to its binary representation.
func (e *IndexEntry) serialize() ([]byte, error) {
	pathLen := len(e.Path.String())
	totalBytes := IndexEntryFixedSize + pathLen + IndexEntryPathNullTerminateSize
	padding := (PaddingAlignment - (totalBytes % PaddingAlignment)) % PaddingAlignment
	totalBytes += padding
	var buffer bytes.Buffer
	buffer.Grow(totalBytes)
	if err := writeIndexEntryFields(&buffer, e); err != nil {
		return nil, err
	}
	if _, err := buffer.Write(e.Hash[:]); err != nil {
		return nil, err
	}
	if err := binary.Write(&buffer, binary.BigEndian, e.Flags); err != nil {
		return nil, err
	}
	if _, err := buffer.WriteString(e.Path.String()); err != nil {
		return nil, err
	}
	buffer.WriteByte(0)
	paddingBytes := make([]byte, padding)
	if _, err := buffer.Write(paddingBytes); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

// MatchesStat compares the entry's stored metadata against a fresh stat call.
// Returns true if the file appears unchanged (same device, inode, times, size, mode).
func (e *IndexEntry) MatchesStat(stat *FileStat) bool {
	if e.ChangedTime != stat.ChangedTime {
		return false
	}
	if e.ModifiedTime != stat.ModifiedTime {
		return false
	}
	if e.Size != stat.Size {
		return false
	}
	if e.Device != stat.Device || e.Inode != stat.Inode {
		return false
	}
	// TODO: compare filemode vs os stat mode
	return true
}

// Index stores the staging area entries and checksum.
type Index struct {
	// Header stores the parsed index header.
	Header IndexHeader
	// Entries stores the tracked paths in sorted order.
	Entries []*IndexEntry
	// Checksum stores the hex-encoded SHA-256 checksum of the serialized data.
	Checksum string
}

// NewIndex returns an Index backed by the supplied header, entries, and checksum.
func NewIndex(header IndexHeader, entries []*IndexEntry, checksum string) *Index {
	return &Index{
		Header:   header,
		Entries:  entries,
		Checksum: checksum,
	}
}

// NewEmptyIndex returns an empty index with the default signature and version.
func NewEmptyIndex() *Index {
	signatureBytes := [4]byte([]byte(IndexSignature))
	header := NewIndexHeader(signatureBytes, IndexVersion, 0)
	return NewIndex(header, []*IndexEntry{}, "")
}

func (idx *Index) Clone() *Index {
	if idx == nil {
		return nil
	}

	cloned := &Index{
		Header:   idx.Header,
		Checksum: idx.Checksum,
		Entries:  make([]*IndexEntry, len(idx.Entries)),
	}
	for i, entry := range idx.Entries {
		if entry == nil {
			continue
		}
		entryCopy := *entry
		cloned.Entries[i] = &entryCopy
	}
	return cloned
}

// AddEntry inserts an entry into the index, maintaining sorted order.
// If an entry with the same path exists, it is replaced.
func (idx *Index) AddEntry(entry *IndexEntry) {
	i := sort.Search(
		len(idx.Entries), func(i int) bool {
			return idx.Entries[i].Path.String() >= entry.Path.String()
		},
	)
	if i < len(idx.Entries) && idx.Entries[i].Path == entry.Path {
		idx.Entries[i] = entry
		return
	}

	idx.Entries = append(idx.Entries, nil)
	copy(idx.Entries[i+1:], idx.Entries[i:])
	idx.Entries[i] = entry
	idx.Header.NumEntries = uint32(len(idx.Entries))
}

func (idx *Index) ReplaceEntries(entries []*IndexEntry) {
	idx.Entries = entries
	idx.Header.NumEntries = uint32(len(entries))
}

// UpdateEntry replaces an existing entry with the same path. Returns true if updated, false if not found.
func (idx *Index) UpdateEntry(entry *IndexEntry) bool {
	prevEntry, i := idx.FindEntry(entry.Path)
	if prevEntry == nil {
		return false
	}
	idx.Entries[i] = entry
	return true
}

// SetEntry updates an existing entry or adds it if not found.
func (idx *Index) SetEntry(entry *IndexEntry) {
	if !idx.UpdateEntry(entry) {
		idx.AddEntry(entry)
	}
}

// RemoveEntry removes the entry with the given path if it exists.
func (idx *Index) RemoveEntry(path NormalizedPath) {
	if entry, i := idx.FindEntry(path); entry != nil {
		idx.Entries = append(idx.Entries[:i], idx.Entries[i+1:]...)
		idx.Header.NumEntries = uint32(len(idx.Entries))
	}
}

// FindEntry looks up an entry by path. Returns the entry and its index, or nil if not found.
func (idx *Index) FindEntry(path NormalizedPath) (*IndexEntry, int) {
	i := sort.Search(
		len(idx.Entries), func(i int) bool {
			return idx.Entries[i].Path.String() >= path.String()
		},
	)
	if i < len(idx.Entries) && idx.Entries[i].Path.String() == path.String() {
		return idx.Entries[i], i
	}
	return nil, 0
}

// FindEntriesByPathPrefix returns all entries whose path starts with the given prefix.
func (idx *Index) FindEntriesByPathPrefix(prefix string) []*IndexEntry {
	var result []*IndexEntry
	for _, entry := range idx.Entries {
		if strings.HasPrefix(entry.Path.String(), prefix) {
			result = append(result, entry)
		}
	}
	return result
}

// FindEntriesByPathPattern returns all entries whose path matches the given glob pattern.
func (idx *Index) FindEntriesByPathPattern(pattern string) []*IndexEntry {
	var result []*IndexEntry
	for _, entry := range idx.Entries {
		if match, _ := filepath.Match(pattern, entry.Path.String()); match {
			result = append(result, entry)
		}
	}
	return result
}

// HasEntry reports whether an entry with the given path exists in the index.
func (idx *Index) HasEntry(path NormalizedPath) bool {
	entry, _ := idx.FindEntry(path)
	return entry != nil
}

// Serialize serializes the entire index to bytes, including header, entries, and checksum.
func (idx *Index) Serialize() ([]byte, error) {
	serializedHeader := idx.serializeHeader()

	slices.SortFunc(
		idx.Entries, func(a, b *IndexEntry) int {
			return strings.Compare(a.Path.String(), b.Path.String())
		},
	)

	serializedEntries, err := idx.serializeEntries()
	if err != nil {
		return nil, err
	}

	data := append(serializedHeader, serializedEntries...)
	checksum := ComputeSHA256(data)
	checksumBytes, err := hex.DecodeString(checksum)
	if err != nil {
		return nil, err
	}

	result := make([]byte, 0, len(data)+len(checksumBytes))
	result = append(result, data...)
	result = append(result, checksumBytes...)

	return result, nil
}

// serializeHeader converts the index header to bytes.
func (idx *Index) serializeHeader() []byte {
	serializedHeader := make([]byte, IndexHeaderSize)
	header := idx.Header

	copy(serializedHeader[0:IndexHeaderSignatureSize], header.Signature[:])
	binary.BigEndian.PutUint32(
		serializedHeader[IndexHeaderSignatureSize:IndexHeaderSignatureSize+IndexHeaderVersionSize], header.Version,
	)
	binary.BigEndian.PutUint32(serializedHeader[IndexHeaderSignatureSize+IndexHeaderVersionSize:], header.NumEntries)

	return serializedHeader
}

// serializeEntries converts all index entries to bytes.
func (idx *Index) serializeEntries() ([]byte, error) {
	var serializedEntries []byte
	for _, entry := range idx.Entries {
		serializedEntry, err := entry.serialize()
		if err != nil {
			return nil, err
		}
		serializedEntries = append(serializedEntries, serializedEntry...)
	}
	return serializedEntries, nil
}

// DeserializeIndex parses index data from bytes.
// It validates the header signature, version, entry count, and checksum.
func DeserializeIndex(data []byte) (*Index, error) {
	if len(data) == 0 {
		return NewEmptyIndex(), nil
	}
	if len(data) < IndexHeaderSize {
		return nil, ErrIndexTooShort
	}

	var index Index
	offset := 0

	header, err := deserializeHeader(data[:IndexHeaderSize])
	if err != nil {
		return nil, err
	}
	index.Header = header

	numEntries := header.NumEntries
	offset += IndexHeaderSize

	for i := uint32(0); i < numEntries; i++ {
		if offset >= len(data)-IndexChecksumSize {
			return nil, ErrTruncatedEntryData
		}

		entry, entrySize, err := deserializeIndexEntry(data[offset:])
		if err != nil {
			return nil, err
		}

		index.AddEntry(entry)
		offset += entrySize
	}

	if len(data)-offset != IndexChecksumSize {
		return nil, ErrIncorrectChecksumSize
	}

	expectedChecksumBytes := data[len(data)-IndexChecksumSize:]
	actualChecksum := ComputeSHA256(data[:len(data)-IndexChecksumSize])
	actualChecksumBytes, err := hex.DecodeString(actualChecksum)
	if err != nil {
		return nil, err
	}

	if !bytes.Equal(expectedChecksumBytes, actualChecksumBytes) {
		return nil, ErrChecksumMismatch
	}

	index.Checksum = actualChecksum
	return &index, nil
}

// ComputeIndexFlags encodes the path length and stage into a 16-bit flags field.
func ComputeIndexFlags(path string, stage uint16) uint16 {
	pathLength := min(len(path), MaxPathLength)
	flags := uint16(pathLength) | (stage << StageShift)
	return flags
}

// deserializeHeader parses the index header from bytes.
func deserializeHeader(data []byte) (IndexHeader, error) {
	var header IndexHeader

	copy(header.Signature[:], data[0:IndexHeaderSignatureSize])
	if !bytes.Equal(header.Signature[:], []byte(IndexSignature)) {
		return header, ErrInvalidIndexSignature
	}

	header.Version = binary.BigEndian.Uint32(data[IndexHeaderSignatureSize : IndexHeaderSignatureSize+IndexHeaderVersionSize])
	header.NumEntries = binary.BigEndian.Uint32(data[IndexHeaderSignatureSize+IndexHeaderVersionSize:])

	return header, nil
}

// deserializeIndexEntry parses a single index entry from bytes.
// Returns the entry and the total number of bytes consumed (including padding).
func deserializeIndexEntry(data []byte) (*IndexEntry, int, error) {
	if len(data) < IndexEntryFixedSize {
		return nil, 0, ErrEntryDataTooShort
	}

	var entry IndexEntry
	reader := bytes.NewReader(data)

	if err := readIndexEntryFields(reader, &entry); err != nil {
		return nil, 0, err
	}

	hashBytes := make([]byte, IndexEntryHashSize)
	if _, err := reader.Read(hashBytes); err != nil {
		return nil, 0, err
	}

	hash, err := NewHashFromHex(hex.EncodeToString(hashBytes))
	if err != nil {
		return nil, 0, err
	}
	entry.Hash = hash

	if err := binary.Read(reader, binary.BigEndian, &entry.Flags); err != nil {
		return nil, 0, err
	}

	offset := IndexEntryFixedSize
	pathEnd := bytes.IndexByte(data[offset:], 0)
	if pathEnd == -1 {
		return nil, 0, ErrPathNotNullTerminated
	}

	normPath, err := ParseNormalizedPath(string(data[offset : offset+pathEnd]))
	if err != nil {
		return nil, 0, fmt.Errorf("invalid path in index entry: %w", err)
	}

	entry.Path = normPath
	offset += pathEnd + IndexEntryPathNullTerminateSize
	padding := (PaddingAlignment - (offset % PaddingAlignment)) % PaddingAlignment
	totalSize := offset + padding

	return &entry, totalSize, nil
}

// writeIndexEntryFields writes the fixed-size fields of an index entry to the buffer.
func writeIndexEntryFields(buffer *bytes.Buffer, entry *IndexEntry) error {
	if err := binary.Write(buffer, binary.BigEndian, entry.ChangedTime.Unix()); err != nil {
		return err
	}
	if err := binary.Write(buffer, binary.BigEndian, uint32(entry.ChangedTime.Nanosecond())); err != nil {
		return err
	}
	if err := binary.Write(buffer, binary.BigEndian, entry.ModifiedTime.Unix()); err != nil {
		return err
	}
	if err := binary.Write(buffer, binary.BigEndian, uint32(entry.ModifiedTime.Nanosecond())); err != nil {
		return err
	}
	if err := binary.Write(buffer, binary.BigEndian, entry.Device); err != nil {
		return err
	}
	if err := binary.Write(buffer, binary.BigEndian, entry.Inode); err != nil {
		return err
	}
	if err := binary.Write(buffer, binary.BigEndian, entry.Mode); err != nil {
		return err
	}
	if err := binary.Write(buffer, binary.BigEndian, entry.UserID); err != nil {
		return err
	}
	if err := binary.Write(buffer, binary.BigEndian, entry.GroupID); err != nil {
		return err
	}
	if err := binary.Write(buffer, binary.BigEndian, entry.Size); err != nil {
		return err
	}
	return nil
}

// readIndexEntryFields reads the fixed-size fields of an index entry from the reader.
func readIndexEntryFields(reader *bytes.Reader, entry *IndexEntry) error {
	var changedTimeUnix int64
	var changedTimeNanoseconds uint32
	if err := binary.Read(reader, binary.BigEndian, &changedTimeUnix); err != nil {
		return err
	}
	if err := binary.Read(reader, binary.BigEndian, &changedTimeNanoseconds); err != nil {
		return err
	}
	entry.ChangedTime = time.Unix(changedTimeUnix, int64(changedTimeNanoseconds))

	var modifiedTimeUnix int64
	var modifiedTimeNanoseconds uint32
	if err := binary.Read(reader, binary.BigEndian, &modifiedTimeUnix); err != nil {
		return err
	}
	if err := binary.Read(reader, binary.BigEndian, &modifiedTimeNanoseconds); err != nil {
		return err
	}
	entry.ModifiedTime = time.Unix(modifiedTimeUnix, int64(modifiedTimeNanoseconds))

	if err := binary.Read(reader, binary.BigEndian, &entry.Device); err != nil {
		return err
	}
	if err := binary.Read(reader, binary.BigEndian, &entry.Inode); err != nil {
		return err
	}
	if err := binary.Read(reader, binary.BigEndian, &entry.Mode); err != nil {
		return err
	}
	if err := binary.Read(reader, binary.BigEndian, &entry.UserID); err != nil {
		return err
	}
	if err := binary.Read(reader, binary.BigEndian, &entry.GroupID); err != nil {
		return err
	}
	if err := binary.Read(reader, binary.BigEndian, &entry.Size); err != nil {
		return err
	}
	return nil
}
