package core

import (
	"github.com/rendau/fs/internal/interfaces"
)

type St struct {
	lg interfaces.Logger
}

func New(
	lg interfaces.Logger,
) *St {
	c := &St{
		lg: lg,
	}

	return c
}
