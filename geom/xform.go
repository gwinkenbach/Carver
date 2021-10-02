package geom

import (
	"log"
)

/*****
 *
 * About Coordinates.
 *
 * Material Coordinates (MC) refer to the original dimensions of the material with
 * (0, 0) at the bottom left and (material width, material height) at the top right.
 *
 * Carving coordinates are expressed in the Material Coordinates space.
 *
 * Normalized Image Coordinate  (NIC) are coordinates of the image with (0, 0) at
 * the bottom left and (1, 1) at the top right.
 *
 * In picture:
 *
 *                                                      MC: (width, height)
 *      Material                                            (200, 150)
 *      +-------------------------------------------------+
 *      |                                                 |
 *      |                                                 |
 *      |                                                 |
 *      |         Carving                (180, 135)       |
 *      |         +--------------------------+            |
 *      |         |                          |            |
 *      |         |                          | NIC:       |
 *      |         |##########################| (1, 1)     |
 *      |         |##########################|            |
 *      |    NIC: |##########################|            |
 *      |   (0, 0)|##########################|            |
 *      |         |                          |            |
 *      |         +--------------------------+            |
 *      |      (20, 15)                                   |
 *      |                                                 |
 *      |                                                 |
 *      |                                                 |
 *      |                                                 |
 *      +-------------------------------------------------+
 *    (0, 0)
 **/

// Image modes determines how the image fits within the carving area.
const (
	ImageModeFill = 0 // Stretch image to fill viewport
	ImageModeFit  = 1 // Whole image fits in viewport, keep aspect ratio
	ImageModeCrop = 2 // Crop image to fill viewport, keep aspect ratio
)

// XformCache compute and caches useful coordinate transformations.
type XformCache struct {
	materialWidth  float64
	materialHeight float64
	carvingWidth   float64
	carvingHeight  float64
	offsetX        float64
	offsetY        float64
	imgPixWidth    int
	imgPixHeight   int
	imgMode        int

	mc2Nic *Matrix33 // Machine coord. to Normalized Image Coord.
}

// NewXformCache create and returns a new XfromCache.
func NewXformCache(
	matWidth, matHeight float32,
	carvWidth, carvHeight, offsetX, offsetY float32,
	imgPixWidth, imgPixHeight, imgMode int) *XformCache {

	return &XformCache{
		materialHeight: float64(matHeight),
		materialWidth:  float64(matWidth),
		carvingHeight:  float64(carvHeight),
		carvingWidth:   float64(carvWidth),
		offsetX:        float64(offsetX),
		offsetY:        float64(offsetY),
		imgPixHeight:   imgPixHeight,
		imgPixWidth:    imgPixWidth,
		imgMode:        imgMode,
	}
}

// Clear clears all the cached tranformations.
func (xf *XformCache) Clear() {
	xf.mc2Nic = nil
}

// GetMc2NicXform returns the normalized-material-coordinates to normalize-image-coordinates
// transformation, creating it if necessary.
func (xf *XformCache) GetMc2NicXform() *Matrix33 {
	if xf.mc2Nic == nil {
		xf.mc2Nic = xf.makeMc2NicTransform()
	}

	return xf.mc2Nic
}

func (xf *XformCache) makeMc2NicTransform() *Matrix33 {

	// Normalized image dimensions.
	imgDim := NewVec2(float64(xf.imgPixWidth), float64(xf.imgPixHeight))

	// Calculate carving area in NMC space.
	carvBottomLeft := NewPt2(xf.offsetX, xf.offsetY)
	carvTopRight := NewPt2(xf.carvingWidth+xf.offsetX, xf.carvingHeight+xf.offsetY)

	var m *Matrix33
	switch xf.imgMode {
	case ImageModeFill:
		m = xf.calcFillModeMc2NicTransform(&carvBottomLeft, &carvTopRight)
	case ImageModeFit:
		m = xf.calcFitOrCropModeMc2NicTransform(&carvBottomLeft, &carvTopRight, &imgDim, ImageModeFit)
	case ImageModeCrop:
		m = xf.calcFitOrCropModeMc2NicTransform(&carvBottomLeft, &carvTopRight, &imgDim, ImageModeCrop)
	default:
		log.Fatalf("Invalid image mode: %d\n", xf.imgMode)
	}

	return m
}

func (xf *XformCache) calcFillModeMc2NicTransform(
	carvBottomLeft, carvTopRight *Pt2) *Matrix33 {

	m := NewTranslateMatrix33(-carvBottomLeft.X, -carvBottomLeft.Y)
	s := NewScaleMatrix33(
		1.0/(carvTopRight.X-carvBottomLeft.X),
		1.0/(carvTopRight.Y-carvBottomLeft.Y))
	m.Mul(&s)

	return &m
}

func (xf *XformCache) calcFitOrCropModeMc2NicTransform(
	carvBottomLeftMc, carvTopRightMc *Pt2, imgDim *Vec2, imageMode int) *Matrix33 {

	// Dimensions of carving area in material coordinates.
	carvDim := carvTopRightMc.Sub(*carvBottomLeftMc)

	// Let's figure whether to fit along X or Y.
	scaleX := carvDim.X / imgDim.X
	scaleY := carvDim.Y / imgDim.Y

	scale := scaleX
	if imageMode == ImageModeFit {
		// Pick the smallest of scaleX, scaleY to fit the image within the carving area.
		if scaleY < scaleX {
			scale = scaleY
		}
	} else if imageMode == ImageModeCrop {
		// Pick the largest of scaleX, scaleY to expand the image within the carving area.
		if scaleY > scaleX {
			scale = scaleY
		}
	} else {
		log.Fatal("Invalid image mode")
	}

	// Offset to corner of image within or outside carving area.
	offset := NewVec2(0.5*(carvDim.X-scale*imgDim.X), 0.5*(carvDim.Y-scale*imgDim.Y))

	// Dimension of image as scaled to fit within carving area, in MC. Point qBL is image bottom-left,
	// and qTR is image top-right.
	qBL := carvBottomLeftMc.Add(offset)
	qTR := carvTopRightMc.SubV(offset)

	// Translation by offset takes care of the image inset inside carving area.
	m := NewTranslateMatrix33(-qBL.X, -qBL.Y)
	// Scale by image dimensions within carving area to map to norm. image coord.
	s := NewScaleMatrix33(1.0/(qTR.X-qBL.X), 1.0/(qTR.Y-qBL.Y))
	m.Mul(&s)

	return &m
}
