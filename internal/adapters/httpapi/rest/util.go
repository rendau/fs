package rest

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
)

func (a *St) uRespondJSON(w http.ResponseWriter, obj interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(obj); err != nil {
		a.lg.Infow("Fail to send response", "error", err)
	}
}

func (a *St) uLogErrorRequest(r *http.Request, err interface{}, msg string) {
	a.lg.Errorw(
		msg,
		err,
		"method", r.Method,
		"path", r.URL,
	)
}

func (a *St) uQpParseBool(values url.Values, key string) *bool {
	if qp, ok := values[key]; ok {
		if result, err := strconv.ParseBool(qp[0]); err == nil {
			return &result
		}
	}
	return nil
}

func (a *St) uQpParseBoolV(values url.Values, key string) bool {
	if v := a.uQpParseBool(values, key); v != nil {
		return *v
	}
	return false
}

func (a *St) uQpParseInt(values url.Values, key string) *int {
	if qp, ok := values[key]; ok {
		if result, err := strconv.Atoi(qp[0]); err == nil {
			return &result
		}
	}
	return nil
}

func (a *St) uQpParseIntV(values url.Values, key string) int {
	if v := a.uQpParseInt(values, key); v != nil {
		return *v
	}
	return 0
}

func (a *St) uQpParseString(values url.Values, key string) *string {
	if qp, ok := values[key]; ok {
		return &(qp[0])
	}
	return nil
}

func (a *St) uQpParseStringV(values url.Values, key string) string {
	if v := a.uQpParseString(values, key); v != nil {
		return *v
	}
	return ""
}
