package cleaner

type Cleaner interface {
	Check(pathList []string) ([]string, error)
}
