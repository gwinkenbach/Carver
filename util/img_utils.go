package util

import (
	"image"
	"image/png"
	"os"
	"path"
)

// ImageToGrayImage converts an image to a Go Gray image an returns the result.
func ImageToGrayImage(img image.Image) *image.Gray {
	if img == nil {
		return nil
	}

	bounds := img.Bounds()
	grayImg := image.NewGray(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			var rgba = img.At(x, y)
			grayImg.Set(x, y, rgba)
		}
	}

	return grayImg
}

// LoadGray8Image loads a Gray16 image from a file, converting the pixel format as necessary.
func LoadGray8Image(imgPath string) *image.Gray {
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

		grayImg := image.NewGray(img.Bounds())
		for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
			for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
				grayImg.Set(x, y, img.At(x, y))
			}
		}

		return grayImg
	}

	return nil
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
