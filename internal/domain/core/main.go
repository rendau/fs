package core

import (
	"context"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func (c *St) Create(ctx context.Context, reqDir string, reqFileName string, reqFile io.Reader) (string, error) {
	datePath := getDateUrlPath()

	absDirPath := filepath.Join(c.dirPath, normalizeFsPath(reqDir), normalizeFsPath(datePath))

	err := os.MkdirAll(absDirPath, os.ModePerm)
	if err != nil {
		return "", err
	}

	reqFileExt := strings.ToLower(filepath.Ext(reqFileName))

	f, err := ioutil.TempFile(absDirPath, "*"+reqFileExt)
	if err != nil {
		return "", err
	}
	defer f.Close()

	_, fileName := filepath.Split(f.Name())

	_, err = io.Copy(f, reqFile)
	if err != nil {
		return "", err
	}

	return path.Join(normalizeUrlPath(reqDir), datePath, fileName), nil
}

func (c *St) Get(ctx context.Context, path string) ([]byte, error) {
	absPath := filepath.Join(c.dirPath, normalizeFsPath(path))

	content, err := ioutil.ReadFile(absPath)
	if err != nil {
		return nil, err
	}

	return content, nil
}
