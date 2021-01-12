package core

import (
	"github.com/rendau/fs/internal/domain/util"
	"github.com/rendau/fs/internal/interfaces"
)

type St struct {
	lg           interfaces.Logger
	dirPath      string
	imgMaxWidth  int
	imgMaxHeight int
}

func New(
	lg interfaces.Logger,
	dirPath string,
	imgMaxWidth int,
	imgMaxHeight int,
) *St {
	c := &St{
		lg:           lg,
		dirPath:      util.NormalizeFsPath(dirPath),
		imgMaxWidth:  imgMaxWidth,
		imgMaxHeight: imgMaxHeight,
	}

	return c
}
