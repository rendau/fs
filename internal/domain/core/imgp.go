package core

import (
	"io"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/rendau/fs/internal/domain/entities"
)

var (
	imgpFileTypes = map[string]struct {
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

func (c *St) imgpHandle(fPath string, w io.Writer, pars *entities.ImgpParsSt) error {
	if pars.IsEmpty() {
		return nil
	}

	fileExt := strings.ToLower(filepath.Ext(fPath))

	imgFormat, ok := imgpFileTypes[fileExt]
	if !ok {
		return nil
	}

	hasChanges := false

	img, err := imaging.Open(fPath, imaging.AutoOrientation(true))
	if err != nil {
		return err
	}

	imgBounds := img.Bounds().Max

	pM := pars.Method
	pW := pars.Width
	pH := pars.Height

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
		}

		hasChanges = true
	}

	// if hasWmark {
	// 	imgBounds = img.Bounds().Max
	// 	if imgBounds.X > 100 {
	// 		for _, pr := range conf.Conf.WMarkServePathList {
	// 			if strings.HasPrefix(fPath, pr) {
	// 				wmarkResized := imaging.Resize(wmark, imgBounds.X/3, 0, imaging.Lanczos)
	// 				img = imaging.OverlayCenter(img, wmarkResized, conf.Conf.WMarkOpacity)
	// 				break
	// 			}
	// 		}
	// 	}
	// }

	if hasChanges {
		if w == nil {
			err = imaging.Save(img, fPath)
			if err != nil {
				c.lg.Errorw("Fail to save image", err)
				return err
			}
		} else {
			err = imaging.Encode(w, img, imgFormat.format)
			if err != nil {
				c.lg.Errorw("Fail to encode image", err)
				return err
			}
		}
	}

	return nil
}
