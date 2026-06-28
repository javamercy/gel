//go:build darwin

package domain

import (
	"fmt"
	"io/fs"
	"syscall"
	"time"
)

func readPlatformStatFields(info fs.FileInfo) (platformStatFields, error) {
	systemInfo := info.Sys()
	stat, ok := systemInfo.(*syscall.Stat_t)
	if !ok {
		return platformStatFields{}, fmt.Errorf("unexpected system stat type %T", systemInfo)
	}
	return platformStatFields{
		deviceID:   uint64(uint32(stat.Dev)),
		inode:      stat.Ino,
		userID:     stat.Uid,
		groupID:    stat.Gid,
		changeTime: time.Unix(stat.Ctimespec.Sec, stat.Ctimespec.Nsec).UTC(),
	}, nil
}
