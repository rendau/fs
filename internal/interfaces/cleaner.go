package interfaces

type Cleaner interface {
	Check(pathList []string) ([]string, error)
}
