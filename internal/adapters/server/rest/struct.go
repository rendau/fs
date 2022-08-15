package rest

import (
	"mime/multipart"
)

type SaveReqSt struct {
	Dir        string                `json:"dir" form:"dir" binding:"required"`
	File       *multipart.FileHeader `json:"file" form:"file" binding:"required"`
	NoCut      bool                  `json:"no_cut" form:"no_cut"`
	ExtractZip bool                  `json:"extract_zip" form:"extract_zip"`
}

type SaveRepSt struct {
	Path string `json:"path"`
}

type GetParamsSt struct {
	W        int    `json:"w" form:"w"`
	H        int    `json:"h" form:"h"`
	M        string `json:"m" form:"m"`
	Download string `json:"download" form:"download"`
}
