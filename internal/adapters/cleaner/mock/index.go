package cleaner

type St struct {
	handler func(pathList []string) []string
}

func New() *St {
	return &St{}
}

func (m *St) SetHandler(v func(pathList []string) []string) {
	m.handler = v
}

func (m *St) Check(pathList []string) ([]string, error) {
	if m.handler != nil {
		return m.handler(pathList), nil
	}

	return []string{}, nil
}
