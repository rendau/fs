package rest

import (
	"bytes"
	"net/http"

	"github.com/gin-gonic/gin"
	dopHttps "github.com/rendau/dop/adapters/server/https"
	"github.com/rendau/dop/dopErrs"
)

// @Router  /kvs/:key [post]
// @Tags    kvs
// @Summary Set file.
// @Param   key path string true "key"
// @Success 200
// @Failure 400 {object} dopTypes.ErrRep
func (a *St) hKvsSet(c *gin.Context) {
	key := c.Param("key")

	err := a.core.Kvs.Set(key, c.Request.Body)
	if dopHttps.Error(c, err) {
		return
	}
}

// @Router  /kvs/:key [get]
// @Tags    kvs
// @Summary Get file.
// @Param   key path string true "key"
// @Param   query query bool   false "download"
// @Produce octet-stream
// @Success 200
// @Failure 400 {object} dopTypes.ErrRep
func (a *St) hKvsGet(c *gin.Context) {
	key := c.Param("key")

	download := c.Query("download")

	data, fModTime, err := a.core.Kvs.Get(key)
	if err != nil {
		if err == dopErrs.ObjectNotFound {
			c.Status(http.StatusNotFound)
		} else {
			dopHttps.Error(c, err)
		}
		return
	}

	fName := key

	if download != "" {
		c.Header("Content-Type", `application/octet-stream`)
		c.Header("Content-Disposition", `attachment; filename="`+download+`"`)

		fName = download
	}

	http.ServeContent(c.Writer, c.Request, fName, fModTime, bytes.NewReader(data))
}

// @Router  /kvs/:key [delete]
// @Tags    kvs
// @Summary Remove file.
// @Param   key   path  string true  "key"
// @Success 200
// @Failure 400 {object} dopTypes.ErrRep
func (a *St) hKvsRemove(c *gin.Context) {
	key := c.Param("key")

	err := a.core.Kvs.Remove(key)
	if err != nil {
		dopHttps.Error(c, err)
		return
	}
}
