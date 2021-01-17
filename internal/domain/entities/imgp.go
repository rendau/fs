package entities

type ImgParsSt struct {
	Method string
	Width  int
	Height int
	WMark  bool
}

func (o *ImgParsSt) Reset() {
	o.Method = ""
	o.Width = 0
	o.Height = 0
	o.WMark = false
}

func (o *ImgParsSt) IsEmpty() bool {
	return o.Width == 0 && o.Height == 0 && !o.WMark
}
