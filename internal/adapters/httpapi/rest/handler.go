package rest

import (
	"bytes"
	"net/http"
	"path"

	"github.com/rendau/fs/internal/domain/entities"

	"github.com/rendau/fs/internal/domain/errs"
)

// swagger:route POST / main hSave
// Upload and save file.
// Responses:
//   200: saveRep
//   400: errRep
func (a *St) hSave(w http.ResponseWriter, r *http.Request) {
	// swagger:parameters hSave
	type docReqSt struct {
		// in:formData
		// required:true
		// swagger:file
		File bytes.Buffer `json:"file"`

		// Directory name on a server, file will be saved to
		// in:formData
		// required: true
		Dir string `json:"dir"`
	}

	// swagger:response saveRep
	type docRepSt struct {
		// in:body
		Body struct {
			Path string `json:"path"`
		}
	}

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

// swagger:route GET /{path} main hGet
// Get or download file.
// Produces:
// - application/octet-stream
// - image/jpeg
// - image/png
// Responses:
//   200: getRep
//   404:
func (a *St) hGet(w http.ResponseWriter, r *http.Request) {
	// swagger:parameters hGet
	type docReqSt struct {
		// Value from `POST` API
		// in:path
		Path string `json:"path"`

		// Width of image (in pixels) to resize *(optional)*
		// in:query
		W string `json:"w"`

		// Height of image (in pixels) to resize *(optional)*
		// in:query
		H string `json:"h"`

		// Method of resizing image. Works only with `w` or `h`
		// Possible values:
		// <ul>
		//   <li>
		//     <strong>fit</strong> - image will fit to <code>w</code> and(or) <code>h</code>. Will not crop image, just resizes with aspect ratio
		//   </li>
		//   <li>
		//     <strong>fill</strong> - image will fill <code>w</code> and(or) <code>h</code>. Might crop edges, resizes with aspect ratio
		//   </li>
		// </ul>
		// in:query
		M string `json:"m"`

		// Name of file. File will be downloaded with this name (optional)
		// in:query
		Download string `json:"download"`
	}

	// swagger:response getRep
	type docRepSt struct {
		// in:body
		// type:file
		Body string
	}

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

func (a *St) hClean(w http.ResponseWriter, r *http.Request) {
	a.cr.Clean(0)

	w.WriteHeader(200)
}
