package rest

import (
	"bytes"
	"net/http"
	"path"
	"strings"

	"github.com/gin-gonic/gin"
	dopHttps "github.com/rendau/dop/adapters/server/https"
	"github.com/rendau/dop/dopErrs"
	"github.com/rendau/fs/internal/domain/entities"

	"github.com/rendau/fs/internal/domain/errs"
)

// @Router   / [post]
// @Tags     main
// @Summary  Upload and save file.
// @Accept   mpfd
// @Param    body  body      SaveReqSt  false  "body"
// @Success  200   {object}  SaveRepSt
// @Failure  400   {object}  dopTypes.ErrRep
func (a *St) hSave(c *gin.Context) {
	var err error

	reqObj := &SaveReqSt{}
	err = c.ShouldBind(reqObj)
	if err != nil {
		dopHttps.Error(c, dopErrs.ErrWithDesc{Err: errs.BadFormData, Desc: err.Error()})
		return
	}
	if reqObj.File == nil {
		dopHttps.Error(c, dopErrs.ErrWithDesc{Err: errs.BadFile})
		return
	}
	f, err := reqObj.File.Open()
	if err != nil {
		a.lg.Errorw("Fail to open file", err)
		dopHttps.Error(c, dopErrs.ErrWithDesc{Err: errs.BadFile})
		return
	}
	defer f.Close()

	result, err := a.core.Create(
		reqObj.Dir,
		reqObj.File.Filename,
		f,
		reqObj.NoCut,
		reqObj.ExtractZip,
	)
	if dopHttps.Error(c, err) {
		return
	}

	c.JSON(http.StatusOK, SaveRepSt{Path: result})
}

// @Router   /:path [get]
// @Tags     main
// @Summary  Get or download file.
// @Param    path   path   string       true   "path"
// @Param    query  query  GetParamsSt  false  "query"
// @Produce  octet-stream
// @Success  200
// @Failure  400    {object}  dopTypes.ErrRep
func (a *St) hGet(c *gin.Context) {
	var err error

	urlPath := c.Request.URL.Path

	if strings.HasPrefix(urlPath, "/static") {
		urlPath = urlPath[7:]
	}

	pars := &GetParamsSt{}
	if !dopHttps.BindQuery(c, pars) {
		return
	}

	fName, fModTime, fData, err := a.core.Get(urlPath, &entities.ImgParsSt{
		Method: pars.M,
		Width:  pars.W,
		Height: pars.H,
	}, pars.Download != "")
	if err != nil {
		if err == dopErrs.ObjectNotFound {
			c.Status(http.StatusNotFound)
		} else {
			dopHttps.Error(c, err)
		}
		return
	}

	if pars.Download != "" {
		pars.Download += path.Ext(fName)
		c.Header("Content-Type", `application/octet-stream`)
		c.Header("Content-Disposition", `attachment; filename="`+pars.Download+`"`)
	}

	http.ServeContent(c.Writer, c.Request, fName, fModTime, bytes.NewReader(fData))
}

func (a *St) hClean(c *gin.Context) {
	a.core.Clean(0)
}
