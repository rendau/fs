package rest

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

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

func (a *St) uQpParseInt64(values url.Values, key string) *int64 {
	if qp, ok := values[key]; ok {
		if result, err := strconv.ParseInt(qp[0], 10, 64); err == nil {
			return &result
		}
	}
	return nil
}

func (a *St) uQpParseFloat64(values url.Values, key string) *float64 {
	if qp, ok := values[key]; ok {
		if result, err := strconv.ParseFloat(qp[0], 64); err == nil {
			return &result
		}
	}
	return nil
}

func (a *St) uQpParseInt(values url.Values, key string) *int {
	if qp, ok := values[key]; ok {
		if result, err := strconv.Atoi(qp[0]); err == nil {
			return &result
		}
	}
	return nil
}

func (a *St) uQpParseString(values url.Values, key string) *string {
	if qp, ok := values[key]; ok {
		return &(qp[0])
	}
	return nil
}

func (a *St) uQpParseTime(values url.Values, key string) *time.Time {
	if qp, ok := values[key]; ok {
		if result, err := time.Parse(time.RFC3339, qp[0]); err == nil {
			return &result
		} else {
			fmt.Println(err)
		}
	}
	return nil
}

func (a *St) uQpParseInt64Slice(values url.Values, key string) *[]int64 {
	if _, ok := values[key]; ok {
		items := strings.Split(values.Get(key), ",")

		result := make([]int64, 0, len(items))

		for _, vStr := range items {
			if v, err := strconv.ParseInt(vStr, 10, 64); err == nil {
				result = append(result, v)
			}
		}

		return &result
	}

	return nil
}
