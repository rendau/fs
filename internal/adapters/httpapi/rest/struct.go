package rest

// swagger:response errRep
type docErrRepSt struct {
	// in:body
	Body ErrRepSt
}

type ErrRepSt struct {
	ErrorCode string `json:"error_code"`
}
