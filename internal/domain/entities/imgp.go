package entities

type ImgpParsSt struct {
	Method string
	Width  int
	Height int
}

func (o *ImgpParsSt) IsEmpty() bool {
	return o.Width == 0 && o.Height == 0
}
