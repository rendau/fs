package rest

import (
	"net/http"

	"github.com/rendau/fs/internal/domain/errs"
)

func (a *St) hSaveFile(w http.ResponseWriter, r *http.Request) {
	var err error

	pDir := r.PostFormValue("dir")

	pFile, header, err := r.FormFile("file")
	if err != nil {
		a.uRespondJSON(w, ErrRepSt{ErrorCode: "bad_file"})
		return
	}
	defer pFile.Close()

	pFileName := header.Filename

	pExtractZip := r.PostFormValue("extract_zip") == "true"

	result, err := a.cr.Create(pDir, pFileName, pFile, pExtractZip)
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
