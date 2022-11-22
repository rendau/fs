package types

import (
	"fmt"
)

type ImgParsSt struct {
	WMark     bool
	Method    string
	Width     int
	Height    int
	Blur      float64
	Grayscale bool
}

func (o *ImgParsSt) Reset() {
	o.Method = ""
	o.Width = 0
	o.Height = 0
	o.Blur = 0
	o.Grayscale = false
	o.WMark = false
}

func (o *ImgParsSt) IsEmpty() bool {
	return o.Width == 0 && o.Height == 0 && o.Blur == 0 && !o.Grayscale && !o.WMark
}

func (o *ImgParsSt) String() string {
	return fmt.Sprintf("m=%s&w=%d&h=%d&blur=%fgrayscale=%vwm=%v", o.Method, o.Width, o.Height, o.Blur, o.Grayscale, o.WMark)
}
