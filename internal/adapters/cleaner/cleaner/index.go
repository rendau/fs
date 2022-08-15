package cleaner

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/rendau/dop/dopErrs"
	"github.com/rendau/fs/internal/interfaces"
)

const conTimeout = 20 * time.Second

type St struct {
	lg       interfaces.Logger
	checkUrl string

	httpClient *http.Client
}

func New(lg interfaces.Logger, checkUrl string) *St {
	return &St{
		lg:         lg,
		checkUrl:   checkUrl,
		httpClient: &http.Client{Timeout: conTimeout},
	}
}

func (c *St) Check(pathList []string) ([]string, error) {
	reqBytes, err := json.Marshal(pathList)
	if err != nil {
		c.lg.Errorw("Fail to marshal json", err)
		return nil, dopErrs.ServiceNA
	}

	req, err := http.NewRequest("PUT", c.checkUrl, bytes.NewBuffer(reqBytes))
	if err != nil {
		c.lg.Errorw("Fail to create http-request", err)
		return nil, dopErrs.ServiceNA
	}

	rep, err := c.httpClient.Do(req)
	if err != nil {
		c.lg.Errorw("Fail to send http-request", err)
		return nil, dopErrs.ServiceNA
	}
	defer rep.Body.Close()

	repBytes, err := ioutil.ReadAll(rep.Body)
	if err != nil {
		c.lg.Errorw("Fail to read body", err)
		return nil, dopErrs.ServiceNA
	}

	if rep.StatusCode < 200 || rep.StatusCode >= 300 {
		c.lg.Errorw(
			"Fail to send http-request, bad status code",
			nil,
			"status_code", rep.StatusCode,
			"body", string(repBytes),
		)
		return nil, dopErrs.ServiceNA
	}

	result := make([]string, 0)

	if err = json.Unmarshal(repBytes, &result); err != nil {
		c.lg.Errorw("Fail to parse json", err, "body", string(repBytes))
		return nil, dopErrs.ServiceNA
	}

	return result, nil
}
