package core

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/rendau/fs/internal/domain/entities"
	"github.com/rendau/fs/internal/domain/util"
)

func (c *St) Create(ctx context.Context, reqDir string, reqFileName string, reqFile io.Reader) (string, error) {
	dateUrlPath := util.GetDateUrlPath()

	absFsDirPath := filepath.Join(c.dirPath, util.NormalizeFsPath(reqDir), util.NormalizeFsPath(dateUrlPath))

	err := os.MkdirAll(absFsDirPath, os.ModePerm)
	if err != nil {
		return "", err
	}

	reqFileExt := strings.ToLower(filepath.Ext(reqFileName))

	fileName, fileFsPath, err := func() (string, string, error) {
		f, err := ioutil.TempFile(absFsDirPath, "*"+reqFileExt)
		if err != nil {
			return "", "", err
		}
		defer f.Close()

		fileFsPath := f.Name()
		_, fileName := filepath.Split(fileFsPath)

		_, err = io.Copy(f, reqFile)
		if err != nil {
			return "", "", err
		}

		return fileName, fileFsPath, nil
	}()
	if err != nil {
		return "", err
	}

	err = c.imgHandle(fileFsPath, nil, &entities.ImgParsSt{
		Method: "fit",
		Width:  c.imgMaxWidth,
		Height: c.imgMaxHeight,
	})

	return path.Join(util.NormalizeUrlPath(reqDir), dateUrlPath, fileName), nil
}

func (c *St) Get(ctx context.Context, path string, imgPars *entities.ImgParsSt) (string, []byte, error) {
	var err error

	absFsPath := filepath.Join(c.dirPath, util.NormalizeFsPath(path))

	var name string
	var content = make([]byte, 0)

	if util.PathIsDir(absFsPath) {
		return "", nil, nil
	} else {
		_, name = filepath.Split(absFsPath)
	}

	if !imgPars.IsEmpty() {
		buffer := bytes.Buffer{}

		err = c.imgHandle(absFsPath, &buffer, imgPars)
		if err != nil {
			return "", nil, err
		}

		content = buffer.Bytes()
	} else {
		content, err = ioutil.ReadFile(absFsPath)
		if err != nil {
			return "", nil, err
		}
	}

	return name, content, nil
}
