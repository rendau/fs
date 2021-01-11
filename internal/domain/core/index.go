package core

import (
	"github.com/rendau/fs/internal/interfaces"
)

type St struct {
	lg      interfaces.Logger
	dirPath string
}

func New(
	lg interfaces.Logger,
	dirPath string,
) *St {
	c := &St{
		lg:      lg,
		dirPath: dirPath,
	}

	return c
}
