package rest

import (
	"bytes"
	"net/http"
	"path"

	"github.com/rendau/fs/internal/domain/entities"

	"github.com/rendau/fs/internal/domain/errs"
)

func (a *St) hSave(w http.ResponseWriter, r *http.Request) {
	var err error

	pDir := r.PostFormValue("dir")

	pFile, header, err := r.FormFile("file")
	if err != nil {
		a.uRespondJSON(w, ErrRepSt{ErrorCode: "bad_file"})
		return
	}
	defer pFile.Close()

	pFileName := header.Filename

	pNoCut := r.PostFormValue("no_cut") == "true"

	pExtractZip := r.PostFormValue("extract_zip") == "true"

	result, err := a.cr.Create(pDir, pFileName, pFile, pNoCut, pExtractZip)
	if err != nil {
		switch cErr := err.(type) {
		case errs.Err:
			a.uRespondJSON(w, ErrRepSt{ErrorCode: cErr.Error()})
		default:
			a.uRespondJSON(w, ErrRepSt{ErrorCode: errs.ServiceNA.Error()})
		}
		return
	}

	a.uRespondJSON(w, map[string]string{
		"path": result,
	})
}

func (a *St) hGet(w http.ResponseWriter, r *http.Request) {
	var err error

	urlPath := r.URL.Path

	urlQuery := r.URL.Query()

	imgPars := &entities.ImgParsSt{
		Method: a.uQpParseStringV(urlQuery, "m"),
		Width:  a.uQpParseIntV(urlQuery, "w"),
		Height: a.uQpParseIntV(urlQuery, "h"),
	}

	download := a.uQpParseStringV(urlQuery, "download")

	fName, fModTime, fData, err := a.cr.Get(urlPath, imgPars, download != "")
	if err != nil {
		switch cErr := err.(type) {
		case errs.Err:
			a.uRespondJSON(w, ErrRepSt{ErrorCode: cErr.Error()})
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

func (a *St) hClean(w http.ResponseWriter, r *http.Request) {
	a.cr.Clean(0)

	w.WriteHeader(200)
}
