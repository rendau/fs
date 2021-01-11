package core

import (
	"context"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
)

func (c *St) Create(ctx context.Context, dir string, reader io.Reader) (string, error) {
	datePath := getDateUrlPath()

	absDirPath := filepath.Join(c.dirPath, normalizeFsPath(dir), normalizeFsPath(datePath))

	err := os.MkdirAll(absDirPath, os.ModePerm)
	if err != nil {
		return "", err
	}

	f, err := ioutil.TempFile(absDirPath, "asd_*.txt")
	if err != nil {
		return "", err
	}
	defer f.Close()

	_, fileName := filepath.Split(f.Name())

	_, err = io.Copy(f, reader)
	if err != nil {
		return "", err
	}

	return path.Join(normalizeUrlPath(dir), datePath, fileName), nil
}

func (c *St) Get(ctx context.Context, path string) ([]byte, error) {
	absPath := filepath.Join(c.dirPath, normalizeFsPath(path))

	content, err := ioutil.ReadFile(absPath)
	if err != nil {
		return nil, err
	}

	return content, nil
}
