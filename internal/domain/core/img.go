package core

import (
	"image"
	"io"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/rendau/fs/internal/domain/types"
)

var (
	imgFileTypes = map[string]struct {
		format      imaging.Format
		contentType string
	}{
		".jpg":  {imaging.JPEG, "image/jpeg"},
		".jpeg": {imaging.JPEG, "image/jpeg"},
		".png":  {imaging.PNG, "image/png"},
		".tif":  {imaging.TIFF, "image/tiff"},
		".tiff": {imaging.TIFF, "image/tiff"},
		".bmp":  {imaging.BMP, "image/bmp"},
		// ".gif":  {imaging.GIF, "image/gif"},
	}
)

type Img struct {
	r            *St
	wMarkPath    string
	wMark        image.Image
	wMarkOpacity float64
}

func NewImg(r *St, wMarkPath string, wMarkOpacity float64) *Img {
	if wMarkOpacity == 0 {
		wMarkOpacity = 1
	}

	return &Img{
		r:            r,
		wMarkPath:    wMarkPath,
		wMarkOpacity: wMarkOpacity,
	}
}

func (c *Img) Start() {
	// load w-mark
	if c.wMarkPath != "" {
		wMark, err := imaging.Open(c.wMarkPath)
		if err == nil {
			c.wMark = wMark
		} else {
			c.r.lg.Errorw("Fail to load w-mark", err)
		}
	}
}

func (c *Img) Handle(fPath string, w io.Writer, pars *types.ImgParsSt) error {
	if pars.IsEmpty() {
		return nil
	}

	fileExt := strings.ToLower(filepath.Ext(fPath))

	imgFormat, ok := imgFileTypes[fileExt]
	if !ok {
		return nil
	}

	pM := pars.Method
	pW := pars.Width
	pH := pars.Height
	pBlur := pars.Blur
	pGrayscale := pars.Grayscale
	pWMark := pars.WMark

	hasChanges := false

	img, err := imaging.Open(fPath, imaging.AutoOrientation(true))
	if err != nil {
		// c.lg.Errorw("Fail to open img", err)
		return nil
	}

	imgBounds := img.Bounds().Max

	if pW > 0 || pH > 0 {
		if pW == 0 {
			if imgBounds.Y > 0 {
				pW = imgBounds.X * pH / imgBounds.Y
			}
			if pW == 0 {
				pW = imgBounds.X
			}
		} else if pH == 0 {
			if imgBounds.X > 0 {
				pH = imgBounds.Y * pW / imgBounds.X
			}
			if pH == 0 {
				pH = imgBounds.Y
			}
		}

		if imgBounds.X > pW || imgBounds.Y > pH {
			if pM == "fit" {
				img = imaging.Fit(img, pW, pH, imaging.Lanczos)
			} else {
				img = imaging.Fill(img, pW, pH, imaging.Center, imaging.Lanczos)
			}

			imgBounds = img.Bounds().Max
		}

		hasChanges = true
	}

	if pBlur != 0 {
		img = imaging.Blur(img, pBlur)
	}

	if pGrayscale {
		img = imaging.Grayscale(img)
	}

	if pWMark && c.wMark != nil {
		if imgBounds.X > 100 {
			wMarkResized := imaging.Resize(c.wMark, imgBounds.X/3, 0, imaging.Lanczos)

			img = imaging.OverlayCenter(img, wMarkResized, c.wMarkOpacity)

			hasChanges = true
		}
	}

	if hasChanges {
		if w == nil {
			err = imaging.Save(img, fPath)
			if err != nil {
				c.r.lg.Errorw("Fail to save image", err)
				return err
			}
		} else {
			err = imaging.Encode(w, img, imgFormat.format)
			if err != nil {
				c.r.lg.Errorw("Fail to encode image", err)
				return err
			}
		}
	}

	return nil
}
