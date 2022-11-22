package core

import (
	"sync"
	"time"

	"github.com/rendau/dop/adapters/logger"
	"github.com/rendau/fs/internal/adapters/cleaner"
	"github.com/rendau/fs/internal/domain/util"
)

type St struct {
	lg            logger.Lite
	dirPath       string
	imgMaxWidth   int
	imgMaxHeight  int
	wMarkDirPaths []string
	testing       bool

	Static *Static
	Img    *Img
	Zip    *Zip
	Cache  *Cache
	Clean  *Clean
	Kvs    *Kvs

	wg     sync.WaitGroup
	stop   bool
	stopMu sync.RWMutex
}

func New(
	lg logger.Lite,
	cleaner cleaner.Cleaner,
	dirPath string,
	imgMaxWidth int,
	imgMaxHeight int,
	wMarkPath string,
	wMarkOpacity float64,
	wMarkDirPaths []string,
	cacheCount int,
	cacheTtl time.Duration,
	testing bool,
) *St {
	c := &St{
		lg:            lg,
		dirPath:       dirPath,
		imgMaxWidth:   imgMaxWidth,
		imgMaxHeight:  imgMaxHeight,
		wMarkDirPaths: wMarkDirPaths,
		testing:       testing,
	}

	for i := range c.wMarkDirPaths {
		c.wMarkDirPaths[i] = util.ToFsPath(c.wMarkDirPaths[i])
	}

	c.Static = NewStatic(c)
	c.Img = NewImg(c, wMarkPath, wMarkOpacity)
	c.Zip = NewZip(c)
	c.Cache = NewCache(c, cacheCount, cacheTtl)
	c.Clean = NewClean(c, cleaner)
	c.Kvs = NewKvs(c)

	return c
}

func (c *St) Start() {
	c.Img.Start()
	c.Cache.Start()
	c.Kvs.Start()
}

func (c *St) StopAndWaitJobs() {
	c.stopMu.Lock()
	c.stop = true
	c.stopMu.Unlock()

	c.wg.Wait()
}

func (c *St) IsStopped() bool {
	c.stopMu.RLock()
	defer c.stopMu.RUnlock()

	return c.stop
}
