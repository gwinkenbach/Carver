package fui

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
)

const (
	borderSize  = 1      // 1 pixel border size.
	ImgModeFill = "fill" // Fills canvas without maintaining aspect ratio.
	ImgModeFit  = "fit"  // Maintain aspect ratio and stretch image into the canvas.
	ImgModeCrop = "crop" // Maintain aspect ratio and stretch image to completely fill the canvas.
)

func round(v float32) float32 {
	return float32(math.Round(float64(v)))
}

var materialFrameBackground = color.NRGBA{R: 240, G: 240, B: 220, A: 255}
var transparent = color.NRGBA{R: 255, G: 255, B: 255, A: 0}

type ImagePanel struct {
	lock sync.Mutex

	// Used to feed a new UI image to update goroutine.
	newImageQueue chan image.Image

	root      *fyne.Container
	panelSize float32

	materialFrame *canvas.Rectangle

	uiImage     *canvas.Image
	originalImg image.Image

	imgMirrorX  bool
	imgMirrorY  bool
	imgFillMode string

	materialWidth  float32
	materialHeight float32
	carvingWidth   float32
	carvingHeight  float32
	carvingOffsetX float32
	carvingOffsetY float32
}

func NewImagePanel(panelSize float32) *ImagePanel {
	ip := &ImagePanel{
		panelSize:      panelSize,
		imgFillMode:    ImgModeFill,
		materialWidth:  100,
		materialHeight: 100,
		carvingWidth:   90,
		carvingHeight:  90,
		carvingOffsetX: 5,
		carvingOffsetY: 5,

		newImageQueue: make(chan image.Image, 1),
	}

	// Use a transparent background to enforce the size of the canvas.
	background := canvas.NewRectangle(transparent)
	background.SetMinSize(fyne.NewSize(ip.panelSize, ip.panelSize))

	// The material frame draws a black border around the material area.
	ip.materialFrame = canvas.NewRectangle(materialFrameBackground)
	ip.materialFrame.Move(fyne.NewPos(borderSize, borderSize))
	ip.materialFrame.StrokeWidth = borderSize
	ip.materialFrame.StrokeColor = color.Black

	// For now the image is just a placeholder.
	ip.uiImage = &canvas.Image{}
	//	g := canvas.NewVerticalGradient(color.Black, color.White)
	//	ip.uiImage.Image = g.Generate(200, 200)
	ip.uiImage.Move(fyne.NewPos(borderSize, borderSize))
	ip.uiImage.Resize(fyne.NewSize(200, 200))

	ip.root = container.NewWithoutLayout(background, ip.materialFrame, ip.uiImage)
	ip.root.Resize(fyne.NewSize(ip.panelSize, ip.panelSize))

	ip.Refresh() // I.e. setup the frames per material/canvas dimensions.
	return ip
}

func (ip *ImagePanel) Refresh() {
	ip.lock.Lock()
	defer ip.lock.Unlock()

	ip.internalRefresh(ip.materialWidth, ip.materialHeight,
		ip.carvingWidth, ip.carvingHeight, ip.carvingOffsetX, ip.carvingOffsetY,
		ip.imgMirrorX, ip.imgMirrorY, ip.imgFillMode)
}

func (ip *ImagePanel) SetMaterialDimensions(width, height float64) {
	ip.materialWidth = float32(width)
	ip.materialHeight = float32(height)
}

func (ip *ImagePanel) SetCarvingAreaDimensions(width, height float64) {
	ip.carvingWidth = float32(width)
	ip.carvingHeight = float32(height)
}

func (ip *ImagePanel) SetCarvingAreaOffsets(dx, dy float64) {
	ip.carvingOffsetX = float32(dx)
	ip.carvingOffsetY = float32(dy)
}

func (ip *ImagePanel) SetImageOptions(mode string, mirrorX, mirrorY bool) {
	ip.imgFillMode = mode
	ip.imgMirrorX = mirrorX
	ip.imgMirrorY = mirrorY

	fmt.Printf("Image options: %s, mirrorX = %v, mirrorY = %v\n", mode, mirrorX, mirrorY)
}

func (ip *ImagePanel) getRootContainer() *fyne.Container {
	return ip.root
}

func (ip *ImagePanel) getCanvasSize() float32 {
	return ip.panelSize - 2*borderSize
}

func (ip *ImagePanel) internalRefresh(
	matW, matH, carvW, carvH, offsetX, offsetY float32,
	imgMirrorX, imgMirrorY bool, imgFillMode string) {

	// Adjust the material frame.
	if matW < 10. || matH < 10.0 {
		return
	}

	canvasSize := ip.getCanvasSize()
	h, w := canvasSize, canvasSize
	if matW > matH {
		h = round((matH / matW) * canvasSize)
	} else {
		w = round((matW / matH) * canvasSize)
	}
	ip.materialFrame.Resize(fyne.NewSize(float32(w), float32(h)))

	// Now adjust the carving area frame.
	if carvW < 5.0 || carvH < 5.0 {
		return
	}

	ref := matH
	if matW > matH {
		ref = matW
	}
	h = round((carvH / ref) * canvasSize)
	w = round((carvW / ref) * canvasSize)
	posX := round((offsetX / ref) * canvasSize)
	posY := round((offsetY / ref) * canvasSize)

	// Temporary image:
	g := canvas.NewLinearGradient(color.Black, color.White, 45)
	ip.uiImage.Image = g.Generate(int(w), int(h))
	ip.uiImage.SetMinSize(fyne.NewSize(w, h))
	ip.uiImage.Resize(fyne.NewSize(w, h))
	ip.uiImage.Move(fyne.NewPos(posX, posY))
	ip.uiImage.Refresh()

	// If a new image is available, update the pixmap in the carving frame.
	// var img = ip.getUpdatedUIImage()
	// if img != nil {
	// 	if imgMirrorX || imgMirrorY {
	// 		img = img.Mirrored2(imgMirrorX, imgMirrorY)
	// 	}
	// 	pixmap := gui.QPixmap_FromImage(img, core.Qt__AutoColor)
	// 	if pixmap != nil {
	// 		ip.setCarvingPixmap(pixmap)
	// 	}
	// } else {
	// 	ip.carvingFrame.SetPixmap(nil)
	// }

	//ip.root.Refresh()
}
