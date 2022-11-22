package core

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rendau/dop/dopErrs"
	"github.com/rendau/fs/internal/cns"
	"github.com/rendau/fs/internal/domain/errs"
	"github.com/rendau/fs/internal/domain/types"
	"github.com/rendau/fs/internal/domain/util"
)

func (c *St) Create(reqDir string, reqFileName string, reqFile io.Reader, noCut bool, unZip bool) (string, error) {
	if strings.Contains("/"+util.ToUrlPath(reqDir), "/"+cns.ZipDirNamePrefix) {
		return "", errs.BadDirName
	}

	dateUrlPath := util.GetDateUrlPath()

	absFsDirPath := filepath.Join(c.dirPath, util.ToFsPath(reqDir), util.ToFsPath(dateUrlPath))

	err := os.MkdirAll(absFsDirPath, os.ModePerm)
	if err != nil {
		c.lg.Errorw("Fail to create dirs", err)
		return "", err
	}

	reqFileExt := strings.ToLower(filepath.Ext(reqFileName))

	var targetFsPath string
	var isZipDir bool

	if unZip && reqFileExt == ".zip" {
		targetFsPath, err = os.MkdirTemp(absFsDirPath, cns.ZipDirNamePrefix+"*")
		if err != nil {
			c.lg.Errorw("Fail to create temp-dir", err)
			return "", err
		}

		err = c.Zip.Extract(reqFile, targetFsPath)
		if err != nil {
			return "", err
		}

		isZipDir = true
	} else {
		targetFsPath, err = func() (string, error) {
			f, err := os.CreateTemp(absFsDirPath, "*"+reqFileExt)
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

		if !noCut {
			err = c.Img.Handle(targetFsPath, nil, &types.ImgParsSt{
				Method: "fit",
				Width:  c.imgMaxWidth,
				Height: c.imgMaxHeight,
			})
			if err != nil {
				return "", err
			}
		}
	}

	fileFsRelPath, err := filepath.Rel(c.dirPath, targetFsPath)
	if err != nil {
		c.lg.Errorw("Fail to get relative path", err, "path", targetFsPath, "base", c.dirPath)
		return "", err
	}

	fileUrlRelPath := util.ToUrlPath(fileFsRelPath)

	if isZipDir {
		fileUrlRelPath += "/"
	}

	return fileUrlRelPath, nil
}

func (c *St) Get(reqPath string, imgPars *types.ImgParsSt, download bool) (string, time.Time, []byte, error) {
	var err error

	cKey := c.Cache.GenerateKey(reqPath, imgPars, download)

	if name, modTime, content := c.Cache.GetAndRefresh(cKey); content != nil {
		return name, modTime, content, nil
	}

	reqFsPath := util.ToFsPath(reqPath)
	absFsPath := filepath.Join(c.dirPath, reqFsPath)

	name := ""
	modTime := time.Now()
	content := make([]byte, 0)

	fInfo, err := os.Stat(absFsPath)
	if err != nil {
		if !os.IsNotExist(err) {
			c.lg.Errorw("Fail to get stat of file", err, "f_path", absFsPath)
		}
		return "", modTime, nil, dopErrs.ObjectNotFound
	}

	if !download {
		modTime = fInfo.ModTime()
	}

	if fInfo.IsDir() {
		dirName := filepath.Base(absFsPath)

		if strings.HasPrefix(dirName, cns.ZipDirNamePrefix) {
			if download {
				archiveBuffer, err := c.Zip.CompressDir(absFsPath)
				if err != nil {
					return "", modTime, nil, err
				}

				return "archive.zip", modTime, archiveBuffer.Bytes(), nil
			} else if strings.HasSuffix(reqPath, "/") {
				absFsPath = filepath.Join(absFsPath, "index.html")
				name = "index.html"
				imgPars.Reset()
			} else {
				return "", modTime, nil, dopErrs.ObjectNotFound
			}
		} else {
			return "", modTime, nil, dopErrs.ObjectNotFound
		}
	} else {
		_, name = filepath.Split(absFsPath)
	}

	for _, p := range c.wMarkDirPaths {
		if strings.HasPrefix(reqFsPath, p) {
			imgPars.WMark = true
			break
		}
	}

	if !imgPars.IsEmpty() {
		buffer := new(bytes.Buffer)

		err = c.Img.Handle(absFsPath, buffer, imgPars)
		if err != nil {
			return "", modTime, nil, err
		}

		if buffer.Len() > 0 {
			content = buffer.Bytes()
		}
	}

	if len(content) == 0 {
		content, err = ioutil.ReadFile(absFsPath)
		if err != nil {
			c.lg.Errorw("Fail to read file", err, "f_path", absFsPath)
			return "", modTime, nil, err
		}
	}

	c.Cache.Set(cKey, name, modTime, content)

	return name, modTime, content, nil
}
