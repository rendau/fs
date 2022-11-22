package core

import (
	"archive/zip"
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type Zip struct {
	r *St
}

func NewZip(r *St) *Zip {
	return &Zip{
		r: r,
	}
}

func (c *Zip) Extract(archive io.Reader, dstDirPath string) error {
	fileDataRaw, err := ioutil.ReadAll(archive)
	if err != nil {
		c.r.lg.Errorw("Fail to read archive", err)
		return err
	}

	reader, err := zip.NewReader(bytes.NewReader(fileDataRaw), int64(len(fileDataRaw)))
	if err != nil {
		c.r.lg.Errorw("Fail to create zip-reader", err)
		return err
	}

	fileHandler := func(f *zip.File, skipDirPrefix string) error {
		dstPath := filepath.Join(dstDirPath, strings.TrimLeft(f.Name, skipDirPrefix))

		if f.FileInfo().IsDir() {
			err = os.MkdirAll(dstPath, f.Mode())
			if err != nil {
				c.r.lg.Errorw("Fail to create dirs", err)
				return err
			}
		} else {
			err = os.MkdirAll(filepath.Dir(dstPath), os.ModePerm)
			if err != nil {
				c.r.lg.Errorw("Fail to create dirs for file", err)
				return err
			}

			srcFile, err := f.Open()
			if err != nil {
				c.r.lg.Errorw("Fail to open file in archive", err)
				return err
			}
			defer srcFile.Close()

			dstFile, err := os.Create(dstPath)
			if err != nil {
				c.r.lg.Errorw("Fail to create file", err)
				return err
			}
			defer dstFile.Close()

			_, err = io.Copy(dstFile, srcFile)
			if err != nil {
				c.r.lg.Errorw("Fail to copy data", err)
				return err
			}
		}

		return nil
	}

	filter := func(p string) bool {
		return strings.Contains("/"+p, "__MACOSX")
	}

	var skipDirPrefix string

	for _, f := range reader.File {
		if filter(f.Name) {
			continue
		}

		fPathSlice := strings.SplitN(f.Name, "/", 2)

		if len(fPathSlice) > 1 {
			if skipDirPrefix == "" {
				skipDirPrefix = fPathSlice[0]
			} else if fPathSlice[0] != skipDirPrefix {
				skipDirPrefix = ""
				break
			}
		} else {
			skipDirPrefix = ""
			break
		}
	}

	if skipDirPrefix != "" {
		skipDirPrefix += "/"
	}

	for _, f := range reader.File {
		if filter(f.Name) {
			continue
		}

		err = fileHandler(f, skipDirPrefix)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Zip) CompressDir(dirPath string) (*bytes.Buffer, error) {
	result := new(bytes.Buffer)

	zipWriter := zip.NewWriter(result)
	defer zipWriter.Close()

	err := filepath.Walk(dirPath, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info == nil {
			return nil
		}

		if p == dirPath || info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(dirPath, p)
		if err != nil {
			return err
		}

		srcF, err := os.Open(p)
		if err != nil {
			c.r.lg.Errorw("Fail to open file", err)
			return err
		}
		defer srcF.Close()

		dstF, err := zipWriter.Create(relPath)
		if err != nil {
			c.r.lg.Errorw("Fail to create file in zip", err)
			return err
		}

		_, err = io.Copy(dstF, srcF)
		if err != nil {
			c.r.lg.Errorw("Fail to copy file data", err)
			return err
		}

		return nil
	})
	if err != nil {
		c.r.lg.Errorw("Fail to walk dir", err, "dir_path", dirPath)
		return nil, err
	}

	return result, nil
}
