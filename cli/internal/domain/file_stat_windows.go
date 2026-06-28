//go:build windows

package domain

func readPlatformStatFields(_ fs.FileInfo) (platformStatFields, error) {
	return platformStatFields{}, nil
}
