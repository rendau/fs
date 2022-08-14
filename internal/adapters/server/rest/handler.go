package rest

import (
	"bytes"
	"net/http"
	"path"

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
// @Param    body  body  SaveReqSt  false  "body"
// @Success  200 {object} SaveRepSt
// @Failure  400  {object}  dopTypes.ErrRep
func (a *St) hSave(c *gin.Context) {
	var err error

	reqObj := &SaveReqSt{}
	err = c.ShouldBind(reqObj)
	if err != nil {
		dopHttps.Error(c, dopErrs.ErrWithDesc{Err: errs.BadFormData})

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

	c.JSON(http.StatusOK, SaveRepSt{
		Path: result,
	})
}

// swagger:route GET /{path} main hGet
// Get or download file.
// Produces:
// - application/octet-stream
// - image/jpeg
// - image/png
// Responses:
//   200: getRep
//   404:

// @Router   /:path [get]
// @Tags     main
// @Summary  Get or download file.
// @Param    path     path   string                true   "path"
// @Param    query  query  GetParamsSt  false  "query"
// @Success  200  octet-stream
// @Failure  400  {object}  dopTypes.ErrRep
func (a *St) hGet(c *gin.Context) {
	var err error

	urlPath := c.Request.URL.Path

	pars := &GetParamsSt{}
	if !dopHttps.BindQuery(c, pars) {
		return
	}

	imgPars := &entities.ImgParsSt{
		Method: pars.M,
		Width:  pars.W,
		Height: pars.H,
	}

	fName, fModTime, fData, err := a.core.Get(urlPath, imgPars, pars.Download != "")
	if err != nil {
		switch cErr := err.(type) {
		case errs.Err:
			if cErr == errs.NotFound {
				http.NotFound(w, r)
			} else {
				a.uRespondJSON(w, ErrRepSt{ErrorCode: cErr.Error()})
			}
		default:
			a.uRespondJSON(w, ErrRepSt{ErrorCode: errs.ServiceNA.Error()})
		}
		return
	}

	if download != "" {
		download += path.Ext(fName)
		w.Header().Set("Content-Type", `application/octet-stream`)
		w.Header().Set("Content-Disposition", `attachment; filename="`+download+`"`)
	}

	http.ServeContent(w, r, fName, fModTime, bytes.NewReader(fData))
}

func (a *St) hClean(c *gin.Context) {
	a.core.Clean(0)
}
