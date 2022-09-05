package cleaner

import (
	"github.com/rendau/dop/adapters/client/httpc"
	"github.com/rendau/dop/adapters/logger"
)

type St struct {
	lg    logger.Lite
	httpc httpc.HttpC
}

func New(lg logger.Lite, httpc httpc.HttpC) *St {
	return &St{
		lg:    lg,
		httpc: httpc,
	}
}

func (c *St) Check(pathList []string) ([]string, error) {
	result := make([]string, 0, len(pathList))

	_, err := c.httpc.SendJsonRecvJson(pathList, &result, httpc.OptionsSt{})

	return result, err
}
