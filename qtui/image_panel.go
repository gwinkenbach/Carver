package qtui

import (
	"math"
	"time"

	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/gui"
	"github.com/therecipe/qt/widgets"
)

const (
	panelSize   = 512
	imgModeFill = 0
	imgModeFit  = 1
	imgModeCrop = 2
)

type ImagePanel struct {
	root          *widgets.QGridLayout
	materialFrame *widgets.QLabel
	carvingFrame  *widgets.QLabel

	// Cache image, to avoid unecessary updates.
	cachedImage *gui.QImage

	// Used to feed a new UI image to update goroutine.
	uiImageFeed chan *gui.QImage

	materialWidth  float32
	materialHeight float32
	carvingWidth   float32
	carvingHeight  float32
	carvingOffsetX float32
	carvingOffsetY float32

	imgMirrorX  bool
	imgMirrorY  bool
	imgFillMode int

	updateTimer *time.Timer
}

func NewImagePanel(parent *widgets.QHBoxLayout) *ImagePanel {
	ip := ImagePanel{
		root:        widgets.NewQGridLayout(nil),
		uiImageFeed: make(chan *gui.QImage, 1),
	}

	ip.materialFrame = widgets.NewQLabel(nil, core.Qt__Widget)
	ip.carvingFrame = widgets.NewQLabel2("Load an image", ip.materialFrame, core.Qt__Widget)

	ip.materialFrame.SetStyleSheet("border: 1px solid black; background-color: darkgray;")
	ip.materialFrame.SetAlignment(core.Qt__AlignCenter)

	ip.carvingFrame.SetStyleSheet("border: 1px solid rgba(0, 0, 128, 0.3); background-color: rgba(200, 200, 255, 0.3);")
	ip.carvingFrame.Move2(20, 20)

	ip.root.AddWidget2(ip.materialFrame, 0, 0, core.Qt__AlignCenter)

	ip.root.SetColumnMinimumWidth(0, panelSize)
	ip.root.SetRowMinimumHeight(0, panelSize)

	parent.AddLayout(ip.root, 0)

	return &ip
}

func (ip *ImagePanel) UpdateMaterialSize(width, height float32) {
	if width <= 10. || height <= 10. {
		return
	}

	ip.materialWidth, ip.materialHeight = width, height

	ip.updateUI()
}

func (ip *ImagePanel) UpdateCarvingArea(width, height, offsetX, offsetY float32) {
	if width <= 5. || height <= 5. {
		return
	}

	ip.carvingWidth, ip.carvingHeight = width, height
	ip.carvingOffsetX, ip.carvingOffsetY = offsetX, offsetY

	ip.updateUI()
}

func (ip *ImagePanel) UpdateImageParameters(mode int, mirrorX, mirrorY bool) {
	if mode < 0 || mode >= len(imageFillModeChoices) {
		return
	}

	ip.imgMirrorX = mirrorX
	ip.imgMirrorY = mirrorY
	ip.imgFillMode = mode

	ip.updateUI()
}

func (ip *ImagePanel) SetImage(img *gui.QImage) {
	if ip.cachedImage != img {
		ip.cachedImage = img
		ip.updateUI()
	}
}

func (ip *ImagePanel) updateUI() {
	ip.scheduleUIImageUpdate(ip.cachedImage)

	if ip.updateTimer == nil {
		ip.updateTimer = time.NewTimer(15 * time.Millisecond)
		go func() {
			for {
				<-ip.updateTimer.C
				ip.updateUIAfterDelay(ip.materialWidth, ip.materialHeight, ip.carvingWidth,
					ip.carvingHeight, ip.carvingOffsetX, ip.carvingOffsetY,
					ip.imgMirrorX, ip.imgMirrorY, ip.imgFillMode)
			}
		}()

	} else {
		ip.updateTimer.Stop()
		ip.updateTimer.Reset(10 * time.Millisecond)
	}
}

func (ip *ImagePanel) updateUIAfterDelay(
	matW, matH, carvW, carvH, offsetX, offsetY float32,
	imgMirrorX, imgMirrorY bool, imgFillMode int) {

	// Adjust the material frame.
	if matW < 10. || matH < 10.0 {
		return
	}

	h := panelSize
	w := panelSize
	if matW > matH {
		h = int(math.Round(float64((matH / matW) * float32(panelSize))))
	} else {
		w = int(math.Round(float64((matW / matH) * float32(panelSize))))
	}
	ip.materialFrame.SetFixedSize2(w, h)

	// Now adjust the carving area frame.
	if carvW < 5.0 || carvH < 5.0 {
		return
	}

	ref := matH
	if matW > matH {
		ref = matW
	}
	h = int(math.Round(float64((carvH / ref) * float32(panelSize))))
	w = int(math.Round(float64((carvW / ref) * float32(panelSize))))
	ip.carvingFrame.SetFixedSize2(w, h)

	posX := int(math.Round(float64((offsetX / ref) * float32(panelSize))))
	posY := int(math.Round(float64((offsetY / ref) * float32(panelSize))))
	ip.carvingFrame.Move2(posX, posY)

	// If a new image is available, update the pixmap in the carving frame.
	var img = ip.getUpdatedUIImage()
	if img != nil {
		if imgMirrorX || imgMirrorY {
			img = img.Mirrored2(imgMirrorX, imgMirrorY)
		}
		pixmap := gui.QPixmap_FromImage(img, core.Qt__AutoColor)
		if pixmap != nil {
			ip.setCarvingPixmap(pixmap)
		}
	} else {
		ip.carvingFrame.SetPixmap(nil)
	}
}

func (ip *ImagePanel) setCarvingPixmap(pix *gui.QPixmap) {
	ip.carvingFrame.SetScaledContents(ip.imgFillMode == imgModeFill)
	if ip.imgFillMode == imgModeCrop {
		dx := float32(pix.Width()) / float32(ip.carvingFrame.Width())
		dy := float32(pix.Height()) / float32(ip.carvingFrame.Height())
		if dx > dy {
			pix = pix.ScaledToHeight(ip.carvingFrame.Height(), core.Qt__SmoothTransformation)
		} else {
			pix = pix.ScaledToWidth(ip.carvingFrame.Width(), core.Qt__SmoothTransformation)
		}
	} else if ip.imgFillMode == imgModeFit {
		dx := float32(pix.Width()) / float32(ip.carvingFrame.Width())
		dy := float32(pix.Height()) / float32(ip.carvingFrame.Height())
		if dx < dy {
			pix = pix.ScaledToHeight(ip.carvingFrame.Height(), core.Qt__SmoothTransformation)
		} else {
			pix = pix.ScaledToWidth(ip.carvingFrame.Width(), core.Qt__SmoothTransformation)
		}
	}

	ip.carvingFrame.SetAlignment(core.Qt__AlignHCenter | core.Qt__AlignVCenter)
	ip.carvingFrame.SetPixmap(pix)
}

// Tell the UI-update goroutine to update the image display with img.
func (ip *ImagePanel) scheduleUIImageUpdate(img *gui.QImage) {
	if img != nil {
		// Ensure the channel is empty.
		for ip.getUpdatedUIImage() != nil {
		}

		ip.uiImageFeed <- ip.cachedImage
	}
}

// Return the image to update the UI with, or nil if none is available. Non-blocking.
func (ip *ImagePanel) getUpdatedUIImage() *gui.QImage {
	select {
	case img := <-ip.uiImageFeed:
		return img
	default:
		return nil
	}
}
