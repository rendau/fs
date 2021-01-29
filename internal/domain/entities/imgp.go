package entities

import (
	"fmt"
)

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

func (o *ImgParsSt) String() string {
	return fmt.Sprintf("m=%s&w=%d&h=%dwm=%t", o.Method, o.Width, o.Height, o.WMark)
}
