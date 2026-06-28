//go:build linux

package domain

func readPlatformStatFields(info fs.FileInfo) (platformStatFields, error) {
	systemInfo := info.Sys()
	stat, ok := systemInfo.(*syscall.Stat_t)
	if !ok {
		return platformStatFields{}, fmt.Errorf(
			"unexpected system stat type %T",
			systemInfo,
		)
	}
	return platformStatFields{
		deviceID: uint64(stat.Dev),
		inode:    uint64(stat.Ino),
		userID:   uint32(stat.Uid),
		groupID:  uint32(stat.Gid),
		changeTime: time.Unix(
			stat.Ctim.Sec,
			stat.Ctim.Nsec,
		).UTC(),
	}, nil
}
