package core

import (
	"sync"

	"github.com/rendau/fs/internal/domain/util"
	"github.com/rendau/fs/internal/interfaces"
)

type St struct {
	lg            interfaces.Logger
	dirPath       string
	imgMaxWidth   int
	imgMaxHeight  int
	wMarkOpacity  float64
	wMarkDirPaths []string
	cleaner       interfaces.Cleaner
	testing       bool

	wg     sync.WaitGroup
	stop   bool
	stopMu sync.RWMutex
}

func New(
	lg interfaces.Logger,
	dirPath string,
	imgMaxWidth int,
	imgMaxHeight int,
	wMarkPath string,
	wMarkOpacity float64,
	wMarkDirPaths []string,
	cleaner interfaces.Cleaner,
	testing bool,
) *St {
	c := &St{
		lg:            lg,
		dirPath:       util.ToFsPath(dirPath),
		imgMaxWidth:   imgMaxWidth,
		imgMaxHeight:  imgMaxHeight,
		wMarkOpacity:  wMarkOpacity,
		wMarkDirPaths: wMarkDirPaths,
		cleaner:       cleaner,
		testing:       testing,
	}

	c.imgLoadWMark(wMarkPath)

	if c.wMarkOpacity == 0 {
		c.wMarkOpacity = 1
	}

	for i := range c.wMarkDirPaths {
		c.wMarkDirPaths[i] = util.ToFsPath(c.wMarkDirPaths[i])
	}

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
