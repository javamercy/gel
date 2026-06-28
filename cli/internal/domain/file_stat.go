package domain

import (
	"fmt"
	"os"
	"time"
)

// FileStat contains normalized filesystem metadata used by the index stat cache.
// DeviceID, Inode, UserID, GroupID, and ChangeTime are zero when the operating
// system does not expose equivalent metadata.
type FileStat struct {
	DeviceID   uint64
	Inode      uint64
	UserID     uint32
	GroupID    uint32
	Mode       FileMode
	Size       uint64
	ChangeTime time.Time
	ModTime    time.Time
}

type platformStatFields struct {
	deviceID   uint64
	inode      uint64
	userID     uint32
	groupID    uint32
	changeTime time.Time
}

// ReadFileStat retrieves file metadata for the given absolute path.
func ReadFileStat(path AbsolutePath) (FileStat, error) {
	if path.value == "" {
		return FileStat{}, fmt.Errorf(
			"%w: file stat path is the zero value",
			ErrInvalidAbsolutePath,
		)
	}

	info, err := os.Lstat(path.value)
	if err != nil {
		return FileStat{}, fmt.Errorf(
			"stat %q: %w",
			path.value,
			err,
		)
	}

	mode, err := FileModeFromFS(info.Mode())
	if err != nil {
		return FileStat{}, fmt.Errorf(
			"stat %q: classify file mode: %w",
			path.value,
			err,
		)
	}

	size := info.Size()
	if size < 0 {
		return FileStat{}, fmt.Errorf(
			"stat %q: invalid negative size: %d",
			path.value,
			size,
		)
	}

	platform, err := readPlatformStatFields(info)
	if err != nil {
		return FileStat{}, fmt.Errorf(
			"stat %q: read platform metadata: %w",
			path.value,
			err,
		)
	}
	return FileStat{
		DeviceID:   platform.deviceID,
		Inode:      platform.inode,
		UserID:     platform.userID,
		GroupID:    platform.groupID,
		Mode:       mode,
		Size:       uint64(size),
		ChangeTime: platform.changeTime,
		ModTime:    info.ModTime().UTC(),
	}, nil
}
