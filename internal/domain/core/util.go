package core

import (
	"path"
	"path/filepath"
	"strings"
	"time"
)

func normalizeFsPath(v string) string {
	v = strings.ReplaceAll(v, "..", "")
	v = strings.ReplaceAll(v, "./", "")
	v = strings.ReplaceAll(v, "/.", "")
	v = strings.TrimPrefix(strings.TrimSuffix(v, "/"), "/")
	return filepath.Join(strings.Split(v, "/")...)
}

func normalizeUrlPath(v string) string {
	return path.Join(strings.Split(strings.TrimPrefix(strings.TrimSuffix(v, "/"), "/"), "/")...)
}

func getDateUrlPath() string {
	return time.Now().Format("2006/01/02")
}
