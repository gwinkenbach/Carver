package model

import (
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"

	carv "alvin.com/GoCarver/carving"
	"alvin.com/GoCarver/geom"
	"alvin.com/GoCarver/hmap"
	"alvin.com/GoCarver/qtui"
	"alvin.com/GoCarver/util"

	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/gui"
	"github.com/therecipe/qt/widgets"
)

// var _ qtui.MenuListener = (*Controller)(nil)

type Controller struct {
	uiManager  *qtui.UIManager
	model      *Model
	mb         *qtui.MenuBar
	mainWindow *widgets.QMainWindow
}

func NewController(um *qtui.UIManager, m *Model) *Controller {
	c := &Controller{
		uiManager: um,
		model:     m,
	}

	um.SetOnItemChanged(func(um *qtui.UIManager, tag string) {
		c.doOnItemChanged(tag)
	})

	return c
}

func (c *Controller) GetUIManager() *qtui.UIManager {
	return c.uiManager
}

func (c *Controller) GetModel() *Model {
	return c.model
}

func (c *Controller) ConnectUI(mainWindow *widgets.QMainWindow) {
	c.mainWindow = mainWindow
	c.updateAllUIItems()
}

func (c *Controller) SetMenuBar(mb *qtui.MenuBar) {
	c.mb = mb
	if mb != nil {
		mb.SetMenuListener(c)
	}
	c.updateMenuItems()
}

func (c *Controller) DoMenuChoice(menuID uint32) {
	switch menuID {
	case qtui.IDNewModel:
		break
	case qtui.IDOpenModel:
		c.doOpenModel()
		break
	case qtui.IDSaveModel:
		c.doSaveModel()
		break
	case qtui.IDSaveModelAs:
		c.doSaveModelAs()
		break
	case qtui.IDOpenImage:
		c.doOpenImageFile()
	case qtui.IDGenGRBL:
		c.doRunCarver()
	default:
	}
}

func (c *Controller) CheckShouldClose(event *gui.QCloseEvent) {
	if c.checkSaveOnDirty() == widgets.QMessageBox__Cancel {
		event.Ignore()
	} else {
		event.Accept()
	}
}

func (c *Controller) checkSaveOnDirty() widgets.QMessageBox__StandardButton {
	if c.model.dirty {
		choice := widgets.QMessageBox_Warning(c.mainWindow, "Unsaved changes",
			"There are unsaved changes. Would you like to save the model to file?",
			widgets.QMessageBox__Save|widgets.QMessageBox__Discard|widgets.QMessageBox__Cancel,
			widgets.QMessageBox__Save)
		if choice == widgets.QMessageBox__Save {
			if !c.doSaveModel() {
				return widgets.QMessageBox__Abort
			}
		}

		if choice == widgets.QMessageBox__Cancel {
			return widgets.QMessageBox__Cancel
		}
	}

	return widgets.QMessageBox__Ok
}

func (c *Controller) updateAllUIItems() {
	for _, tag := range qtui.AllUIItemTags {
		c.updateUIFromModel(tag)
	}

	c.updateMenuItems()
	c.uiManager.GetMainLayout().GetImagePanel().SetImage(c.model.GetHeightMap())
}

func (c *Controller) doOpenImageFile() {
	// TODO: save image file data/name to model.
	imageFile := widgets.QFileDialog_GetOpenFileName(c.mainWindow,
		"Choose an image file", "", "Image Files (*.png *.jpg)", "", 0)
	if imageFile != "" {
		img := gui.NewQImage9(imageFile, "")
		if img == nil {
			msg := fmt.Sprintf("Could not load image %s", imageFile)
			widgets.QMessageBox_Warning(c.mainWindow, "Open Image Error", msg,
				widgets.QMessageBox__Ok, widgets.QMessageBox__Ok)
			return
		}

		img = img.ConvertToFormat2(gui.QImage__Format_Grayscale16, core.Qt__AutoColor)
		c.model.root.HeightMap.Image = img
		c.model.dirty = true
		c.updateAllUIItems()
	}
}

func (c *Controller) doRunCarver() {
	dir := ""
	if c.model.fromFilePath != "" {
		dir = filepath.Dir(c.model.fromFilePath)
	}
	outFile := c.getGrblOutputFile(dir)
	if outFile == nil {
		return
	}

	defer outFile.Close()
	carver := carv.NewCarver(outFile)

	// Test image conversion:
	// gr := util.QtImageToGray16Image(c.model.GetHeightMap())
	// util.WriteGray16ImageToPng(gr, "/Users/billy/test_gary.png")

	materialDim := geom.NewSize2FromFloat32(
		c.model.GetFloat32(qtui.ItemMaterialWidth),
		c.model.GetFloat32(qtui.ItemMaterialHeight))
	carvingOrigin := geom.NewPt2FromFloat32(
		c.model.GetFloat32(qtui.ItemCarvingAreaOffsetX),
		c.model.GetFloat32(qtui.ItemCarvingAreaOffsetY))
	carvingAreaDim := geom.NewSize2FromFloat32(
		c.model.GetFloat32(qtui.ItemCarvingAreaWidth),
		c.model.GetFloat32(qtui.ItemCarvingAreaHeight))
	materialTopZ := c.model.GetFloat32(qtui.ItemMaterialThickness)
	carver.ConfigureMaterial(materialDim, carvingOrigin, carvingAreaDim, float64(materialTopZ))

	toolType := carverToolTypeFromModelToolType(c.model.GetChoice(qtui.ItemToolType))
	toolDiameter := c.model.GetFloat32(qtui.ItemToolDiameter)
	horizFeedRate := c.model.GetFloat32(qtui.ItemHorizontalFeedRate)
	vertFeedRate := c.model.GetFloat32(qtui.ItemVerticalFeedRate)
	carver.ConfigureTool(
		toolType, float64(toolDiameter), float64(horizFeedRate), float64(vertFeedRate))

	topZ := materialTopZ + c.model.GetFloat32(qtui.ItemWhiteCarvingDepth)
	bottomZ := materialTopZ + c.model.GetFloat32(qtui.ItemBlackCarvingDepth)
	stepOverFraction := float64(c.model.GetFloat32(qtui.ItemToolStepOver)) * 0.01
	stepOverFraction = math.Max(0.05, math.Min(1.0, stepOverFraction))
	maxStepDown := c.model.GetFloat32(qtui.ItemMaxStepDownSize)
	mode := carverModeFromModelCarvingMode(c.model.GetChoice(qtui.ItemCarvingMode))
	carver.ConfigureCarvingProfile(
		c.getCarvingSampler(materialDim, carvingAreaDim, carvingOrigin),
		float64(topZ), float64(bottomZ),
		stepOverFraction, float64(maxStepDown),
		mode)

	carver.Run()
}

func (c *Controller) doSaveModel() bool {
	if c.model.fromFilePath == "" {
		return c.doSaveModelAs()
	}

	return c.doSaveModelToFile(c.model.fromFilePath)
}

func (c *Controller) doSaveModelAs() bool {
	dir := ""
	if c.model.fromFilePath != "" {
		dir = filepath.Dir(c.model.fromFilePath)
	}

	modelFile := widgets.QFileDialog_GetSaveFileName(c.mainWindow,
		"Save model as", dir, "Carving model (*.carv)", "", 0)
	if modelFile != "" {
		fmt.Printf("Do save to: %s\n", modelFile)
		return c.doSaveModelToFile(modelFile)
	}

	return false
}

func (c *Controller) doSaveModelToFile(filename string) bool {
	if filename != "" {
		mio := newModelIO(c.model)
		err := mio.writeToFile(filename)
		if err != nil {
			msg := fmt.Sprintf("Could not save emodel to %s: err = %s", filename,
				err.Error())
			widgets.QMessageBox_Warning(c.mainWindow, "Save error", msg,
				widgets.QMessageBox__Ok, widgets.QMessageBox__Ok)

			return false
		}

		c.model.SetDirty(false)
		c.model.fromFilePath = filename
		return true
	}

	return false
}

func (c *Controller) doOpenModel() {
	if c.checkSaveOnDirty() == widgets.QMessageBox__Cancel {
		return
	}

	dir := ""
	if c.model.fromFilePath != "" {
		dir = filepath.Dir(c.model.fromFilePath)
	}

	modelFile := widgets.QFileDialog_GetOpenFileName(c.mainWindow,
		"Open a carve file", dir, "Carving model (*.carv)", "", 0)
	if modelFile != "" {
		mio := newModelIO(c.model)
		if err := mio.readFromFile(modelFile); err != nil {
			msg := fmt.Sprintf("Errors while loading model %s: err = %s", modelFile,
				err.Error())
			widgets.QMessageBox_Warning(c.mainWindow, "Open error", msg,
				widgets.QMessageBox__Ok, widgets.QMessageBox__Ok)
		}

		c.model.fromFilePath = modelFile
		c.model.dirty = false
		c.updateAllUIItems()
	}
}

func (c *Controller) doOnItemChanged(tag string) {
	switch tag {
	case qtui.ItemMaterialWidth:
		c.model.root.Material.MaterialWidth = c.uiManager.GetValueFloat32(tag)
	case qtui.ItemMaterialHeight:
		c.model.root.Material.MaterialHeight = c.uiManager.GetValueFloat32(tag)
	case qtui.ItemMaterialThickness:
		c.model.root.Material.MaterialThickness = c.uiManager.GetValueFloat32(tag)
	case qtui.ItemCarvingAreaWidth:
		c.model.root.Material.CarvingAreaWidth = c.uiManager.GetValueFloat32(tag)
	case qtui.ItemCarvingAreaHeight:
		c.model.root.Material.CarvingAreaHeight = c.uiManager.GetValueFloat32(tag)
	case qtui.ItemCarvingAreaOffsetX:
		c.model.root.Material.CarvingAreaOffsetX = c.uiManager.GetValueFloat32(tag)
	case qtui.ItemCarvingAreaOffsetY:
		c.model.root.Material.CarvingAreaOffsetY = c.uiManager.GetValueFloat32(tag)
	case qtui.ItemBlackCarvingDepth:
		c.model.root.Material.BlackCarvingDepth = c.uiManager.GetValueFloat32(tag)
	case qtui.ItemWhiteCarvingDepth:
		c.model.root.Material.WhiteCarvingDepth = c.uiManager.GetValueFloat32(tag)
	case qtui.ItemToolDiameter:
		c.model.root.Carving.ToolDiameter = c.uiManager.GetValueFloat32(tag)
	case qtui.ItemToolType:
		c.model.root.Carving.ToolType = c.uiManager.GetChoice(tag)
	case qtui.ItemToolStepOver:
		c.model.root.Carving.StepOverPercent = c.uiManager.GetValueFloat32(tag)
	case qtui.ItemMaxStepDownSize:
		c.model.root.Carving.MaxStepDownSize = c.uiManager.GetValueFloat32(tag)
	case qtui.ItemHorizontalFeedRate:
		c.model.root.Carving.HorizontalFeedRate = c.uiManager.GetValueFloat32(tag)
	case qtui.ItemVerticalFeedRate:
		c.model.root.Carving.VerticalFeedRate = c.uiManager.GetValueFloat32(tag)
	case qtui.ItemCarvingMode:
		c.model.root.Carving.CarvingMode = c.uiManager.GetChoice(tag)
	case qtui.ItemImageMode:
		c.model.root.HeightMap.ImageMode = c.uiManager.GetChoice(tag)
	case qtui.ItemMirrorImageX:
		c.model.root.HeightMap.MirrorX = c.uiManager.GetBool(tag)
	case qtui.ItemMirrorImageY:
		c.model.root.HeightMap.MirrorY = c.uiManager.GetBool(tag)
	default:
		log.Fatalf("Controller: doOnItemChanged - unknown tag = %s", tag)
		return
	}

	c.model.SetDirty(true)
	c.updateMenuItems()
}

func (c *Controller) updateUIFromModel(tag string) {
	switch tag {
	case qtui.ItemMaterialWidth:
		c.uiManager.SetValueFloat32(tag, c.model.root.Material.MaterialWidth)
	case qtui.ItemMaterialHeight:
		c.uiManager.SetValueFloat32(tag, c.model.root.Material.MaterialHeight)
	case qtui.ItemMaterialThickness:
		c.uiManager.SetValueFloat32(tag, c.model.root.Material.MaterialThickness)
	case qtui.ItemCarvingAreaWidth:
		c.uiManager.SetValueFloat32(tag, c.model.root.Material.CarvingAreaWidth)
	case qtui.ItemCarvingAreaHeight:
		c.uiManager.SetValueFloat32(tag, c.model.root.Material.CarvingAreaHeight)
	case qtui.ItemCarvingAreaOffsetX:
		c.uiManager.SetValueFloat32(tag, c.model.root.Material.CarvingAreaOffsetX)
	case qtui.ItemCarvingAreaOffsetY:
		c.uiManager.SetValueFloat32(tag, c.model.root.Material.CarvingAreaOffsetY)
	case qtui.ItemBlackCarvingDepth:
		c.uiManager.SetValueFloat32(tag, c.model.root.Material.BlackCarvingDepth)
	case qtui.ItemWhiteCarvingDepth:
		c.uiManager.SetValueFloat32(tag, c.model.root.Material.WhiteCarvingDepth)
	case qtui.ItemToolDiameter:
		c.uiManager.SetValueFloat32(tag, c.model.root.Carving.ToolDiameter)
	case qtui.ItemToolType:
		c.uiManager.SetChoice(tag, c.model.root.Carving.ToolType)
	case qtui.ItemToolStepOver:
		c.uiManager.SetValueFloat32(tag, c.model.root.Carving.StepOverPercent)
	case qtui.ItemMaxStepDownSize:
		c.uiManager.SetValueFloat32(tag, c.model.root.Carving.MaxStepDownSize)
	case qtui.ItemHorizontalFeedRate:
		c.uiManager.SetValueFloat32(tag, c.model.root.Carving.HorizontalFeedRate)
	case qtui.ItemVerticalFeedRate:
		c.uiManager.SetValueFloat32(tag, c.model.root.Carving.VerticalFeedRate)
	case qtui.ItemCarvingMode:
		c.uiManager.SetChoice(tag, c.model.root.Carving.CarvingMode)
	case qtui.ItemImageMode:
		c.uiManager.SetChoice(tag, c.model.root.HeightMap.ImageMode)
	case qtui.ItemMirrorImageX:
		c.uiManager.SetBool(tag, c.model.root.HeightMap.MirrorX)
	case qtui.ItemMirrorImageY:
		c.uiManager.SetBool(tag, c.model.root.HeightMap.MirrorY)
	default:
		log.Fatalf("Controller: updateUIFromModel - unknown tag = %s", tag)
		return
	}
}

func (c *Controller) updateMenuItems() {
	if c.mb != nil {
		c.mb.EnableMenuItem(qtui.IDSaveModel, c.model.dirty || c.model.fromFilePath == "")
		c.mb.EnableMenuItem(qtui.IDGenGRBL, c.model.GetHeightMap() != nil)
	}
}

func (c *Controller) getGrblOutputFile(dir string) *os.File {
	outFile := widgets.QFileDialog_GetSaveFileName(c.mainWindow,
		"Write Carving Code to", dir, "Grbl file (*.gcode)", "", 0)
	if outFile == "" {
		return nil
	}

	f, err := os.OpenFile(outFile, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		msg := fmt.Sprintf("Error opening GRBL output file %s: err = %s", outFile, err.Error())
		widgets.QMessageBox_Warning(c.mainWindow, "Open error", msg,
			widgets.QMessageBox__Ok, widgets.QMessageBox__Ok)
		return nil
	}

	return f
}

func (c *Controller) getCarvingSampler(
	matDim geom.Size2,
	carvDim geom.Size2,
	carvOrigin geom.Pt2) hmap.CarvingDepthSampler {

	heightMap := c.model.GetHeightMap()
	imgMode := c.model.GetChoice(qtui.ItemCarvingMode)

	xform := geom.NewXformCache(
		float32(matDim.W), float32(matDim.H),
		float32(carvDim.W), float32(carvDim.H),
		float32(carvOrigin.X), float32(carvOrigin.Y),
		heightMap.Width(), heightMap.Height(), imgMode)
	imgGray := util.QtImageToGray16Image(heightMap)
	sampler := hmap.NewPixelDepthSampler(xform.GetMc2NicXform(), carvOrigin, carvDim, imgGray)

	return sampler
}

func carverToolTypeFromModelToolType(modelToolType int) int {
	switch modelToolType {
	case ToolTypeBallNose:
		return carv.ToolTypeBall
	case ToolTypeStraight:
		return carv.ToolTypeFlat
	default:
		log.Fatalln("Unknown model tool type")
		return 0
	}
}

func carverModeFromModelCarvingMode(modelCarvingMode int) int {
	switch modelCarvingMode {
	case CarvingModeAlongX:
		return carv.CarveModeXOnly
	case CarvingModeAlongY:
		return carv.CarveModeYOnly
	case CarvingModeAlongXThenY:
		return carv.CarveModeXThenY
	default:
		log.Fatalln("Unknown model carving mode")
		return 0
	}
}
