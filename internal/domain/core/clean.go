package core

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rendau/fs/internal/adapters/cleaner"
	"github.com/rendau/fs/internal/cns"
)

type Clean struct {
	r *St

	cleaner cleaner.Cleaner
}

func NewClean(r *St, cleaner cleaner.Cleaner) *Clean {
	return &Clean{
		r:       r,
		cleaner: cleaner,
	}
}

func (c *Clean) Clean(checkChunkSize int) {
	if checkChunkSize == 0 {
		checkChunkSize = cns.DefaultCleanChunkSize
	}

	c.r.wg.Add(1)
	if c.r.testing {
		c.routine(checkChunkSize)
	} else {
		go c.routine(checkChunkSize)
	}
}

func (c *Clean) routine(checkChunkSize int) {
	defer c.r.wg.Done()

	stop := false

	rootDirPath := c.r.dirPath

	var pathList []string

	var totalCount uint64
	var removedCount uint64

	startTime := time.Now()

	err := filepath.Walk(rootDirPath, func(p string, info os.FileInfo, err error) error {
		if stop {
			return filepath.SkipDir
		}

		if err != nil {
			c.r.lg.Errorw("Fail to walk", err, "path", p)
			return err
		}

		if info == nil {
			return nil
		}

		if p == rootDirPath {
			return nil
		}

		mtIsAllowed := info.ModTime().AddDate(0, 0, cns.CleanFileNotCheckPeriodDays).Before(time.Now())

		if len(pathList) >= checkChunkSize {
			removedCount += c.pathListRoutine(pathList)

			pathList = nil

			if stop = c.r.IsStopped(); stop {
				return filepath.SkipDir
			}
		}

		relPath, err := filepath.Rel(rootDirPath, p)
		if err != nil {
			c.r.lg.Errorw("Fail to get rel p", err, "path", p, "root_dir_path", rootDirPath)
			return err
		}

		if info.IsDir() {
			if !strings.HasPrefix(info.Name(), cns.ZipDirNamePrefix) {
				return nil
			}

			if !mtIsAllowed {
				return filepath.SkipDir
			}

			pathList = append(pathList, relPath+"/")

			totalCount++

			return filepath.SkipDir
		}

		if !mtIsAllowed {
			return nil
		}

		pathList = append(pathList, relPath)

		totalCount++

		return nil
	})
	if err != nil {
		c.r.lg.Errorw("Fail to walk dir", err)
		return
	}

	removedCount += c.pathListRoutine(pathList)

	err = c.removeEmptyDirs(rootDirPath)
	if err != nil {
		c.r.lg.Errorw("Fail to remove empty dirs", err)
		return
	}

	c.r.lg.Infow(
		"Cleaned",
		"total_count", totalCount,
		"removed_count", removedCount,
		"duration", time.Now().Sub(startTime).String(),
	)
}

func (c *Clean) pathListRoutine(pathList []string) uint64 {
	if len(pathList) == 0 {
		return 0
	}

	if c.r.IsStopped() {
		return 0
	}

	rmPathList, err := c.cleaner.Check(pathList)
	if err != nil {
		return 0
	}

	for _, p := range rmPathList {
		// c.r.lg.Infow("Want to remove", "f_path", p)

		err = os.RemoveAll(filepath.Join(c.r.dirPath, p))
		if err != nil {
			c.r.lg.Errorw("Fail to remove path", err, "path", p)
		}
	}

	return uint64(len(rmPathList))
}

func (c *Clean) removeEmptyDirs(rootDirPath string) error {
	if c.r.IsStopped() {
		return nil
	}

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
