package errs

type Err string

func (e Err) Error() string {
	return string(e)
}

const (
	ServiceNA = Err("server_not_available")
	NotFound  = Err("not_found")
)
