package util

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path"

	"github.com/therecipe/qt/gui"
)

// QtImageToGray16Image converts a QT QImage to a Go Gray16 image an retiurns the result.
func QtImageToGray16Image(qtImg *gui.QImage) *image.Gray16 {
	if qtImg == nil || !qtImg.IsGrayscale() {
		return nil
	}

	width := qtImg.Width()
	height := qtImg.Height()
	bounds := image.Rect(0, 0, width, height)
	grayImg := image.NewGray16(bounds)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			pix := qtImg.Pixel2(x, y)
			// Cheat alert: we're dealing with a grayscale image where R = G = B, so we just grab
			// the 8-bit G & B channels to make a 16-bit value.
			gray := pix & 0xffff
			grayImg.SetGray16(x, y, color.Gray16{Y: uint16(gray)})
		}
	}

	return grayImg
}

// LoadGray16Image loads a Gray16 image from a file, converting the pixel format as necessary.
func LoadGray16Image(imgPath string) *image.Gray16 {
	if imgPath != "" {
		dir, _ := os.Getwd()
		fullPath := path.Join(dir, imgPath)

		reader, err := os.Open(fullPath)
		if err != nil {
			return nil
		}
		defer reader.Close()

		img, _, err := image.Decode(reader)
		if err != nil {
			return nil
		}

		if img == nil {
			return nil
		}

		grayImg := image.NewGray16(img.Bounds())
		for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
			for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
				grayImg.Set(x, y, img.At(x, y))
			}
		}

		return grayImg
	}

	return nil
}

// WriteGray16ImageToPng writes a Gary16 image to a PNG file with the given file path.
func WriteGray16ImageToPng(img *image.Gray16, filepath string) error {
	f, err := os.Create(filepath + ".png")
	if err != nil {
		return err
	}
	defer f.Close()

	return png.Encode(f, img)

}
