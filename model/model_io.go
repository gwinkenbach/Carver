package model

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"io"
	"io/ioutil"
	"strings"
)

const (
	fileExtension = "carv"
	modelFilename = "model.json"
	imageFilename = "img.png"
)

type modelIO struct {
	model    *Model
	filename string
}

func newModelIO(m *Model) *modelIO {
	return &modelIO{
		model: m,
	}
}

func (mio *modelIO) readFromFile(filename string) error {
	reader, err := zip.OpenReader(filename)
	if err != nil {
		return err
	}
	defer reader.Close()

	for _, f := range reader.File {
		if err = mio.readFileFromZip(f); err != nil {
			return err
		}
	}

	return nil
}

func (mio *modelIO) readFileFromZip(f *zip.File) error {
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	if strings.HasSuffix(strings.ToLower(f.Name), ".json") {
		return mio.readJSON(rc, f.UncompressedSize64)
	}
	return mio.readImage(rc, f.UncompressedSize64)
}

func (mio *modelIO) readJSON(rc io.ReadCloser, numBytes uint64) error {
	buf := new(bytes.Buffer)
	_, err := io.Copy(buf, rc)
	if err != nil {
		return err
	}

	return json.Unmarshal(buf.Bytes(), &mio.model.root)
}

func (mio *modelIO) readImage(rc io.ReadCloser, numBytes uint64) error {
	img, _, err := image.Decode(rc)
	if img == nil || err != nil {
		return fmt.Errorf("Failed to read image from file")
	}

	mio.model.root.HeightMap.Image = img
	return nil
}

func (mio *modelIO) writeToFile(filename string) error {
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)

	if err := mio.writeJSONModel(w); err != nil {
		return err
	}

	if err := mio.writeImage(w); err != nil {
		return err
	}

	if err := w.Close(); err != nil {
		return err
	}

	return ioutil.WriteFile(filename, buf.Bytes(), 0644)
}

func (mio *modelIO) writeJSONModel(w *zip.Writer) error {
	f, err := w.Create(modelFilename)
	if err != nil {
		return err
	}

	jsonData, err := json.Marshal(&mio.model.root)
	if err != nil {
		return err
	}

	_, err = f.Write(jsonData)
	return err
}

func (mio *modelIO) writeImage(w *zip.Writer) error {
	img := mio.model.GetHeightMap()
	if img != nil {
		f, err := w.Create(imageFilename)
		if err != nil {
			return err
		}

		buffer := new(bytes.Buffer)
		err = png.Encode(buffer, img)
		if err != nil {
			return fmt.Errorf("Could not save image to png")
		}

		const bufferSize = 4096
		bytesBuffer := make([]byte, bufferSize)
		for {
			n, err := buffer.Read(bytesBuffer)
			if n <= 0 {
				break
			}

			_, err = f.Write(bytesBuffer[:n])
			if err != nil {
				return err
			}
		}
	}
	return nil
}
