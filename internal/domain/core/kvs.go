package core

import (
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/rendau/dop/dopErrs"
	"github.com/rendau/fs/internal/cns"
	"github.com/rendau/fs/internal/domain/util"
)

type Kvs struct {
	r *St
}

func NewKvs(r *St) *Kvs {
	return &Kvs{
		r: r,
	}
}

func (c *Kvs) Start() {
	err := os.MkdirAll(filepath.Join(c.r.dirPath, cns.KvsDirNamePrefix), os.ModePerm)
	if err != nil {
		c.r.lg.Errorw("Fail to create kvs-dir", err)
	}
}

func (c *Kvs) Set(key string, file io.Reader) error {
	f, err := os.Create(c.generateAbsFilePath(key))
	if err != nil {
		c.r.lg.Errorw("Fail to create file", err)
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, file)
	if err != nil {
		c.r.lg.Errorw("Fail to copy data", err)
		return err
	}

	return nil
}

func (c *Kvs) Get(key string) (io.ReadSeekCloser, time.Time, error) {
	filePath := c.generateAbsFilePath(key)

	fModTime := time.Now()

	fStat, err := os.Stat(filePath)
	if err != nil {
		if !os.IsNotExist(err) {
			c.r.lg.Errorw("Fail to get stat of file", err, "f_path", filePath)
		}
		return nil, fModTime, dopErrs.ObjectNotFound
	}

	fModTime = fStat.ModTime()

	f, err := os.Open(filePath)
	if err != nil {
		c.r.lg.Errorw("Fail to open file", err)
		return nil, fModTime, err
	}

	return f, fModTime, nil
}

func (c *Kvs) Remove(key string) error {
	err := os.RemoveAll(c.generateAbsFilePath(key))
	if err != nil {
		c.r.lg.Errorw("Fail to create file", err)
		return err
	}

	return nil
}

func (c *Kvs) generateAbsFilePath(key string) string {
	return filepath.Join(c.r.dirPath, cns.KvsDirNamePrefix, util.ToFsPath(key))
}
