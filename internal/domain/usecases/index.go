package usecases

import (
	"github.com/rendau/fs/internal/domain/core"
	"github.com/rendau/fs/internal/interfaces"
)

type St struct {
	lg interfaces.Logger

	cr *core.St
}

func New(
	lg interfaces.Logger,
	cr *core.St,
) *St {
	u := &St{
		lg: lg,
		cr: cr,
	}

	return u
}
