package entities

type ImgParsSt struct {
	Method string
	Width  int
	Height int
}

func (o *ImgParsSt) IsEmpty() bool {
	return o.Width == 0 && o.Height == 0
}
