package core

import (
	"sync"

	"github.com/rendau/fs/internal/domain/util"
	"github.com/rendau/fs/internal/interfaces"
)

type St struct {
	lg           interfaces.Logger
	dirPath      string
	imgMaxWidth  int
	imgMaxHeight int
	cleaner      interfaces.Cleaner
	testing      bool

	wg     sync.WaitGroup
	stop   bool
	stopMu sync.RWMutex
}

func New(
	lg interfaces.Logger,
	dirPath string,
	imgMaxWidth int,
	imgMaxHeight int,
	cleaner interfaces.Cleaner,
	testing bool,
) *St {
	c := &St{
		lg:           lg,
		dirPath:      util.NormalizeFsPath(dirPath),
		imgMaxWidth:  imgMaxWidth,
		imgMaxHeight: imgMaxHeight,
		cleaner:      cleaner,
		testing:      testing,
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
