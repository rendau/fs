package core

import (
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/rendau/dop/dopErrs"
	"github.com/rendau/fs/internal/cns"
	"github.com/rendau/fs/internal/domain/util"
)

type Kvs struct {
	r  *St
	mu sync.RWMutex
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
	c.mu.Lock()
	defer c.mu.Unlock()

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

func (c *Kvs) Get(key string) ([]byte, time.Time, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

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

	fData, err := os.ReadFile(filePath)
	if err != nil {
		c.r.lg.Errorw("Fail to read file", err)
		return nil, fModTime, err
	}

	return fData, fModTime, nil
}

func (c *Kvs) Remove(key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

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
