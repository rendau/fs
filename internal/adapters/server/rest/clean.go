package rest

import (
	"github.com/gin-gonic/gin"
)

func (a *St) hClean(c *gin.Context) {
	a.core.Clean.Clean(0)
}
