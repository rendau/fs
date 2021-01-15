package util

import (
	"os"
	"path"
	"path/filepath"
	"strconv"
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

func NewInt(v int) *int {
	return &v
}

func NewInt64(v int64) *int64 {
	return &v
}

func NewFloat64(v float64) *float64 {
	return &v
}

func NewString(v string) *string {
	return &v
}

func NewBool(v bool) *bool {
	return &v
}

func NewTime(v time.Time) *time.Time {
	return &v
}

func NewSliceInt64(v ...int64) *[]int64 {
	res := make([]int64, 0, len(v))
	res = append(res, v...)
	return &res
}

func NewSliceString(v ...string) *[]string {
	res := make([]string, 0, len(v))
	res = append(res, v...)
	return &res
}

func Int64SliceToString(src []int64, delimiter, emptyV string) string {
	if len(src) == 0 {
		return emptyV
	}

	res := ""

	for _, v := range src {
		if res != "" {
			res += delimiter
		}
		res += strconv.FormatInt(v, 10)
	}

	return res
}
