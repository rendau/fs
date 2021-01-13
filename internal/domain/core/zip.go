package core

import (
	"archive/zip"
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

func (c *St) zipExtract(archive io.Reader, dstDirPath string) error {
	fileDataRaw, err := ioutil.ReadAll(archive)
	if err != nil {
		c.lg.Errorw("Fail to read archive", err)
		return err
	}

	reader, err := zip.NewReader(bytes.NewReader(fileDataRaw), int64(len(fileDataRaw)))
	if err != nil {
		c.lg.Errorw("Fail to create zip-reader", err)
		return err
	}

	fileHandler := func(f *zip.File) error {
		frc, err := f.Open()
		if err != nil {
			c.lg.Errorw("Fail to open file in archive", err)
			return err
		}
		defer frc.Close()

		fPath := filepath.Join(dstDirPath, f.Name)

		if f.FileInfo().IsDir() {
			err = os.MkdirAll(fPath, f.Mode())
			if err != nil {
				c.lg.Errorw("Fail to create dirs", err)
				return err
			}
		} else {
			err = os.MkdirAll(filepath.Dir(fPath), os.ModePerm)
			if err != nil {
				c.lg.Errorw("Fail to create dirs for file", err)
				return err
			}

			dstFile, err := os.Create(fPath)
			if err != nil {
				c.lg.Errorw("Fail to create file", err)
				return err
			}
			defer dstFile.Close()

			_, err = io.Copy(dstFile, frc)
			if err != nil {
				c.lg.Errorw("Fail to copy data", err)
				return err
			}
		}

		return nil
	}

	for _, f := range reader.File {
		err = fileHandler(f)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *St) zipCompressDir(dirPath string) (*bytes.Buffer, error) {
	items := make([][2]string, 0)

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info == nil {
			return nil
		}

		relPath, err := filepath.Rel(dirPath, path)
		if err != nil {
			return err
		}

		items = append(items, [2]string{relPath, path})

		return nil
	})
	if err != nil {
		c.lg.Errorw("Fail to walk dir", err, "dir_path", dirPath)
		return nil, err
	}

	result := new(bytes.Buffer)

	zipWriter := zip.NewWriter(result)
	defer zipWriter.Close()

	for _, item := range items {
		dstF, err := zipWriter.Create(item[0])
		if err != nil {
			c.lg.Errorw("Fail to create file in zip", err)
			return nil, err
		}

		srcF, err := os.Open(item[1])
		if err != nil {
			c.lg.Errorw("Fail to open file", err)
			return nil, err
		}

		_, err = io.Copy(dstF, srcF)
		if err != nil {
			c.lg.Errorw("Fail to copy file data", err)
			return nil, err
		}
	}

	return result, nil
}
