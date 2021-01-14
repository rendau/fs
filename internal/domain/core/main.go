package core

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/rendau/fs/internal/domain/errs"

	"github.com/rendau/fs/internal/cns"

	"github.com/rendau/fs/internal/domain/entities"
	"github.com/rendau/fs/internal/domain/util"
)

func (c *St) Create(ctx context.Context, reqDir string, reqFileName string, reqFile io.Reader, unZip bool) (string, error) {
	if strings.Contains("/"+util.NormalizeUrlPath(reqDir), cns.ZipDirNamePrefix) {
		return "", errs.BadDirName
	}

	dateUrlPath := util.GetDateUrlPath()

	absFsDirPath := filepath.Join(c.dirPath, util.NormalizeFsPath(reqDir), util.NormalizeFsPath(dateUrlPath))

	err := os.MkdirAll(absFsDirPath, os.ModePerm)
	if err != nil {
		c.lg.Errorw("Fail to create dirs", err)
		return "", err
	}

	reqFileExt := strings.ToLower(filepath.Ext(reqFileName))

	var targetFsPath string
	var isZipDir bool

	if unZip && reqFileExt == ".zip" {
		targetFsPath, err = ioutil.TempDir(absFsDirPath, cns.ZipDirNamePrefix+"*")
		if err != nil {
			c.lg.Errorw("Fail to create temp-dir", err)
			return "", err
		}

		err = c.zipExtract(reqFile, targetFsPath)
		if err != nil {
			return "", err
		}

		isZipDir = true
	} else {
		targetFsPath, err = func() (string, error) {
			f, err := ioutil.TempFile(absFsDirPath, "*"+reqFileExt)
			if err != nil {
				c.lg.Errorw("Fail to create temp-file", err)
				return "", err
			}
			defer f.Close()

			_, err = io.Copy(f, reqFile)
			if err != nil {
				c.lg.Errorw("Fail to copy data", err)
				return "", err
			}

			return f.Name(), nil
		}()
		if err != nil {
			return "", err
		}

		err = c.imgHandle(targetFsPath, nil, &entities.ImgParsSt{
			Method: "fit",
			Width:  c.imgMaxWidth,
			Height: c.imgMaxHeight,
		})
		if err != nil {
			return "", err
		}
	}

	fileFsRelPath, err := filepath.Rel(c.dirPath, targetFsPath)
	if err != nil {
		c.lg.Errorw("Fail to get relative path", err, "path", targetFsPath, "base", c.dirPath)
		return "", err
	}

	fileUrlRelPath := util.NormalizeUrlPath(fileFsRelPath)

	if isZipDir {
		fileUrlRelPath += "/"
	}

	return fileUrlRelPath, nil
}

func (c *St) Get(ctx context.Context, reqPath string, imgPars *entities.ImgParsSt, download bool) (string, []byte, error) {
	var err error

	absFsPath := filepath.Join(c.dirPath, util.NormalizeFsPath(reqPath))

	var name string
	var content = make([]byte, 0)

	if util.PathIsDir(absFsPath) {
		dirName := filepath.Base(absFsPath)

		if strings.HasPrefix(dirName, cns.ZipDirNamePrefix) {
			if download {
				archiveBuffer, err := c.zipCompressDir(absFsPath)
				if err != nil {
					return "", nil, err
				}

				return "archive.zip", archiveBuffer.Bytes(), nil
			} else if strings.HasSuffix(reqPath, "/") {
				absFsPath = filepath.Join(absFsPath, "index.html")
				name = "index.html"
				imgPars.Reset()
			} else {
				return "", nil, errs.NotFound
			}
		} else {
			return "", nil, errs.NotFound
		}
	} else {
		_, name = filepath.Split(absFsPath)
	}

	fInfo, err := os.Stat(absFsPath)
	if err != nil {
		if !os.IsNotExist(err) {
			c.lg.Errorw("Fail to get stat of file", err, "f_path", absFsPath)
		}
		return "", nil, errs.NotFound
	}
	if fInfo.IsDir() { // if "index.html" is dir
		return "", nil, errs.NotFound
	}

	if !imgPars.IsEmpty() {
		buffer := new(bytes.Buffer)

		err = c.imgHandle(absFsPath, buffer, imgPars)
		if err != nil {
			return "", nil, err
		}

		content = buffer.Bytes()
	} else {
		content, err = ioutil.ReadFile(absFsPath)
		if err != nil {
			c.lg.Errorw("Fail to read file", err, "f_path", absFsPath)
			return "", nil, err
		}
	}

	return name, content, nil
}
