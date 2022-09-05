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
	cleaner       cleaner.Cleaner
	dirPath       string
	imgMaxWidth   int
	imgMaxHeight  int
	wMarkOpacity  float64
	wMarkDirPaths []string
	testing       bool

	wg     sync.WaitGroup
	stop   bool
	stopMu sync.RWMutex

	Cache *Cache
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
		cleaner:       cleaner,
		dirPath:       dirPath,
		imgMaxWidth:   imgMaxWidth,
		imgMaxHeight:  imgMaxHeight,
		wMarkOpacity:  wMarkOpacity,
		wMarkDirPaths: wMarkDirPaths,
		testing:       testing,
	}

	c.imgLoadWMark(wMarkPath)

	if c.wMarkOpacity == 0 {
		c.wMarkOpacity = 1
	}

	for i := range c.wMarkDirPaths {
		c.wMarkDirPaths[i] = util.ToFsPath(c.wMarkDirPaths[i])
	}

	c.Cache = NewCache(c, cacheCount, cacheTtl)

	return c
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
