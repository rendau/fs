package util

import (
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

func ToFsPath(v string) string {
	v = strings.ReplaceAll(v, "..", "")
	v = strings.ReplaceAll(v, "./", "")
	v = strings.ReplaceAll(v, "/.", "")
	v = strings.TrimPrefix(strings.TrimSuffix(v, "/"), "/")
	return filepath.Join(strings.Split(v, "/")...)
}

func ToUrlPath(v string) string {
	return path.Join(strings.Split(strings.TrimPrefix(strings.TrimSuffix(v, "/"), "/"), "/")...)
}

func GetDateUrlPath() string {
	return time.Now().Format("2006/01/02")
}

func FsPathIsDir(p string) bool {
	fileInfo, err := os.Stat(p)

	return err == nil && fileInfo.IsDir()
}
