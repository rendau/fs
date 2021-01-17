package core

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rendau/fs/internal/cns"
	"github.com/rendau/fs/internal/domain/entities"
	"github.com/rendau/fs/internal/domain/errs"
	"github.com/rendau/fs/internal/domain/util"
)

func (c *St) Create(reqDir string, reqFileName string, reqFile io.Reader, unZip bool) (string, error) {
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

	fileUrlRelPath := util.ToUrlPath(fileFsRelPath)

	if isZipDir {
		fileUrlRelPath += "/"
	}

	return fileUrlRelPath, nil
}

func (c *St) Get(reqPath string, imgPars *entities.ImgParsSt, download bool) (string, time.Time, []byte, error) {
	var err error

	reqFsPath := util.ToFsPath(reqPath)
	absFsPath := filepath.Join(c.dirPath, reqFsPath)

	name := ""
	modTime := time.Now()
	content := make([]byte, 0)

	if util.FsPathIsDir(absFsPath) {
		dirName := filepath.Base(absFsPath)

		if strings.HasPrefix(dirName, cns.ZipDirNamePrefix) {
			if download {
				archiveBuffer, err := c.zipCompressDir(absFsPath)
				if err != nil {
					return "", modTime, nil, err
				}

				return "archive.zip", modTime, archiveBuffer.Bytes(), nil
			} else if strings.HasSuffix(reqPath, "/") {
				absFsPath = filepath.Join(absFsPath, "index.html")
				name = "index.html"
				imgPars.Reset()
			} else {
				return "", modTime, nil, errs.NotFound
			}
		} else {
			return "", modTime, nil, errs.NotFound
		}
	} else {
		_, name = filepath.Split(absFsPath)
	}

	fInfo, err := os.Stat(absFsPath)
	if err != nil {
		if !os.IsNotExist(err) {
			c.lg.Errorw("Fail to get stat of file", err, "f_path", absFsPath)
		}
		return "", modTime, nil, errs.NotFound
	}
	if fInfo.IsDir() { // if "index.html" is dir
		return "", modTime, nil, errs.NotFound
	}

	if !strings.Contains("/"+reqFsPath, "/"+cns.ZipDirNamePrefix) {
		for _, p := range c.wMarkDirPaths {
			if strings.HasPrefix(reqFsPath, p) {
				imgPars.WMark = true
				break
			}
		}
	}

	if !imgPars.IsEmpty() {
		buffer := new(bytes.Buffer)

		err = c.imgHandle(absFsPath, buffer, imgPars)
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

		if !download {
			modTime = fInfo.ModTime()
		}
	}

	return name, modTime, content, nil
}

func (c *St) Clean(checkChunkSize int) {
	if checkChunkSize == 0 {
		checkChunkSize = cns.DefaultCleanChunkSize
	}

	c.wg.Add(1)
	if c.testing {
		c.cleanRoutine(checkChunkSize)
	} else {
		go c.cleanRoutine(checkChunkSize)
	}
}

func (c *St) cleanRoutine(checkChunkSize int) {
	defer c.wg.Done()

	stop := false

	rootDirPath := c.dirPath

	var pathList []string

	var totalCount uint64
	var removedCount uint64

	startTime := time.Now()

	err := filepath.Walk(rootDirPath, func(p string, info os.FileInfo, err error) error {
		if stop {
			return filepath.SkipDir
		}

		if err != nil {
			c.lg.Errorw("Fail to walk", err, "path", p)
			return err
		}

		if info == nil {
			return nil
		}

		if p == rootDirPath {
			return nil
		}

		if len(pathList) >= checkChunkSize {
			removedCount += c.cleanPathListRoutine(pathList)

			pathList = nil

			if stop = c.IsStopped(); stop {
				return filepath.SkipDir
			}
		}

		relPath, err := filepath.Rel(rootDirPath, p)
		if err != nil {
			c.lg.Errorw("Fail to get rel p", err, "path", p, "root_dir_path", rootDirPath)
			return err
		}

		if info.IsDir() {
			if !strings.HasPrefix(info.Name(), cns.ZipDirNamePrefix) {
				return nil
			}

			pathList = append(pathList, relPath+"/")

			totalCount++

			return filepath.SkipDir
		}

		pathList = append(pathList, relPath)

		totalCount++

		return nil
	})
	if err != nil {
		c.lg.Errorw("Fail to walk dir", err)
		return
	}

	removedCount += c.cleanPathListRoutine(pathList)

	err = c.cleanRemoveEmptyDirs(rootDirPath)
	if err != nil {
		c.lg.Errorw("Fail to remove empty dirs", err)
		return
	}

	c.lg.Infow(
		"Cleaned",
		"total_count", totalCount,
		"removed_count", removedCount,
		"duration", time.Now().Sub(startTime).String(),
	)
}

func (c *St) cleanRemoveEmptyDirs(rootDirPath string) error {
	dirs := map[string]uint64{}

	err := filepath.Walk(rootDirPath, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info == nil {
			return nil
		}

		if p == rootDirPath {
			return nil
		}

		if parentPath := filepath.Dir(p); parentPath != rootDirPath {
			dirs[parentPath]++
		}

		if info.IsDir() {
			if _, ok := dirs[p]; !ok {
				dirs[p] = 0
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	var rr func() error

	rr = func() error {
		for k, v := range dirs {
			if k == rootDirPath {
				continue
			}

			if v <= 0 {
				parentPath := filepath.Dir(k)
				if _, ok := dirs[parentPath]; ok {
					dirs[parentPath]--
				}

				err = os.RemoveAll(k)
				if err != nil {
					return err
				}

				delete(dirs, k)

				err = rr()
				if err != nil {
					return err
				}

				break
			}
		}

		return nil
	}

	err = rr()
	if err != nil {
		return err
	}

	return nil
}

func (c *St) cleanPathListRoutine(pathList []string) uint64 {
	if len(pathList) == 0 {
		return 0
	}

	if c.IsStopped() {
		return 0
	}

	rmPathList, err := c.cleaner.Check(pathList)
	if err != nil {
		return 0
	}

	for _, p := range rmPathList {
		err = os.RemoveAll(filepath.Join(c.dirPath, p))
		if err != nil {
			c.lg.Errorw("Fail to remove path", err, "path", p)
		}
	}

	return uint64(len(rmPathList))
}
