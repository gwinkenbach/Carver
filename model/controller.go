package model

import (
	"image"
	"log"
	"math"
	"os"
	"path/filepath"

	carv "alvin.com/GoCarver/carving"
	"alvin.com/GoCarver/fui"
	"alvin.com/GoCarver/geom"
	"alvin.com/GoCarver/hmap"
	"alvin.com/GoCarver/mesh"
	"alvin.com/GoCarver/util"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/disintegration/imaging"
)

type Controller struct {
	uiManager      *UIManager
	model          *Model
	mainWindow     fyne.Window
	useMeshSampler bool
}

func NewController(m *Model) *Controller {
	c := &Controller{
		model:          m,
		useMeshSampler: true,
	}

	return c
}

func (c *Controller) GetUIManager() *UIManager {
	return c.uiManager
}

func (c *Controller) GetModel() *Model {
	return c.model
}

func (c *Controller) ConnectUI(uiManager *UIManager, mainWindow fyne.Window) {
	c.uiManager = uiManager
	c.mainWindow = mainWindow
	c.uiManager.BuildUI(mainWindow)

	c.uiManager.SetUIChangeListener(func(tag string) {
		c.doOnItemChanged(tag)
	})
	c.uiManager.SetMenuSelectedListener(func(tag string) {
		c.DoMenuChoice(tag)
	})

	c.updateAllUIItems()
}

func (c *Controller) DoMenuChoice(menuTag string) {
	switch menuTag {
	case MenuOpenImageTag:
		c.doOpenImageFile()
	case MenuNewModelTag:
		break
	case MenuOpenModelTag:
		c.doOpenModel()
	case MenuSaveModelTag:
		c.doSaveModel()
	case MenuSaveModelAsTag:
		c.doSaveModelAs()
	case MenuGenGrblTag:
		c.doRunCarver()
	default:
	}
}

func (c *Controller) CheckShouldClose() bool {
	return c.checkSaveOnDirty()
}

func (c *Controller) checkSaveOnDirty() bool {
	if c.model.dirty {
		dlg := fui.NewDialog("Unsaved Changes")
		if dlg.ShowYesNoDialog("There are unsaved changes. Would you like to save the model to file?") {
			return c.doSaveModel()
		} else {
			return true
		}
	}

	return true
}

func (c *Controller) updateAllUIItems() {
	for _, tag := range c.uiManager.GetAllUIItemTags() {
		c.updateUIFromModel(tag)
	}

	c.updateMenuItems()
	c.uiManager.SetImage(c.model.GetHeightMap())
}

func (c *Controller) doOpenImageFile() {
	d := fui.NewDialog("Choose Image")
	if img, filename, err := d.OpenAndLoadImageFile(); err == nil {
		c.model.root.HeightMap.Image = img
		c.model.root.HeightMap.ImageFileName = filename
		c.model.SetDirty(true)
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

	var mc carv.MachiningConfig
	mc.Material.MaterialDim = geom.NewSize2FromFloat32(
		c.model.GetFloat32Value(MatWidthTag), c.model.GetFloat32Value(MatHeightTag))
	mc.Material.CarvingAreaOrigin = geom.NewPt2FromFloat32(
		c.model.GetFloat32Value(CarvOffsetXTag), c.model.GetFloat32Value(CarvOffsetYTag))
	mc.Material.CarvingAreaDim = geom.NewSize2FromFloat32(
		c.model.GetFloat32Value(CarvWidthTag), c.model.GetFloat32Value(CarvHeightTag))
	mc.Material.MaterialThickness = float64(c.model.GetFloat32Value(MatThicknessTag))

	mc.Carving.Tool.ToolType = carverToolTypeFromModelToolType(c.model.GetIntValue(ToolTypeTag))
	mc.Carving.Tool.ToolDiameter = float64(c.model.GetFloat32Value(ToolDiamTag))
	mc.Carving.Tool.HorizFeedRate = float64(c.model.GetFloat32Value(HorizFeedRateTag))
	mc.Carving.Tool.VertFeedRate = float64(c.model.GetFloat32Value(VertFeedRateTag))

	stepOverFraction := float64(c.model.GetFloat32Value(StepOverTag)) * 0.01
	mc.Carving.StepOverFraction = math.Max(0.05, math.Min(1.0, stepOverFraction))
	mc.Carving.Tool.MaxStepDown = float64(c.model.GetFloat32Value(MaxStepDownTag))
	mc.Carving.CarvingMode = carverModeFromModelCarvingMode(c.model.GetIntValue(CarvDirectionTag))

	topZ := mc.Material.MaterialThickness + float64(c.model.GetFloat32Value(CarvWhiteDepthTag))
	bottomZ := mc.Material.MaterialThickness + float64(c.model.GetFloat32Value(CarvBlackDepthTag))
	invertImage := false
	if bottomZ > topZ {
		invertImage = true
		bottomZ, topZ = topZ, bottomZ
	}

	mc.Carving.CarvingTopZ = topZ
	mc.Carving.CarvingBottomZ = bottomZ
	mc.Carving.Sampler = c.getCarvingSampler(mc.Material.MaterialDim, mc.Material.CarvingAreaDim,
		mc.Material.CarvingAreaOrigin, invertImage, topZ, bottomZ,
		float64(mc.Carving.Tool.ToolDiameter))

	mc.Carving.FinishStepFraction =
		float64(c.model.GetFloat32Value(FinishPassReductionTag)) * 0.01 * stepOverFraction
	mc.Carving.EnableFinishing = c.model.GetBoolValue(UseFinishPassTag)
	mc.Carving.FinishMode =
		carverFinishModeFromModelFinishMode(c.model.GetIntValue(FinishPassModeTag))
	mc.Carving.FinishHorizFeedRate = float64(c.model.GetFloat32Value(FinishPassHorizFeedRateTag))

	title := "Generating carving code"
	progress := c.showProgressDialog(title, filepath.Base(c.model.fromFilePath))

	carv.DoMachining(&mc, outFile)

	progress.Hide()
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

	dlg := fui.NewDialog("Save Model As")
	filename, err := dlg.SaveToCarverFile(dir)
	if err == nil && filename != "" {
		return c.doSaveModelToFile(filename)
	}

	return false
}

func (c *Controller) doSaveModelToFile(filename string) bool {
	if filename != "" {
		mio := newModelIO(c.model)
		err := mio.writeToFile(filename)
		if err != nil {
			dlg := fui.NewDialog("Save Error")
			dlg.ShowErrorDialog("Could not save emodel to %s: err = %s", filename, err.Error())
			return false
		}

		c.model.SetDirty(false)
		c.model.fromFilePath = filename
		return true
	}

	return false
}

func (c *Controller) doOpenModel() {
	if !c.checkSaveOnDirty() {
		return
	}

	dir := ""
	if c.model.fromFilePath != "" {
		dir = filepath.Dir(c.model.fromFilePath)
	}

	dlg := fui.NewDialog("Open a Carver File")
	filename, err := dlg.OpenCarverFile(dir)
	if err == nil && filename != "" {
		c.uiManager.DisableListeners()
		defer c.uiManager.EnableListeners()

		mio := newModelIO(c.model)
		if err := mio.readFromFile(filename); err != nil {
			dlg := fui.NewDialog("Open Error")
			dlg.ShowErrorDialog("Errors while loading model %s: err = %s", filename, err.Error())
			return
		}

		c.model.fromFilePath = filename
		c.model.SetDirty(false)
		c.updateAllUIItems()
	}
}

func (c *Controller) doOnItemChanged(tag string) {
	switch {
	case c.uiManager.IsNumEntryUIItem(tag):
		SetModelValueByTag(c.model, tag, GetUiValueByTag[float32](c.uiManager, tag))
	case c.uiManager.IsSelectorUIItem(tag):
		SetModelValueByTag(c.model, tag, GetUiValueByTag[int](c.uiManager, tag))
	case c.uiManager.IsCheckboxUIItem(tag):
		SetModelValueByTag(c.model, tag, GetUiValueByTag[bool](c.uiManager, tag))
	default:
		log.Fatalf("Controller: doOnItemChanged - unknown tag = %s", tag)
		return
	}

	c.model.SetDirty(true)
	c.updateMenuItems()
}

func (c *Controller) updateUIFromModel(tag string) {
	switch {
	case c.uiManager.IsNumEntryUIItem(tag):
		SetUiValueByTag(c.uiManager, tag, GetModelValueByTag[float32](c.model, tag))
	case c.uiManager.IsSelectorUIItem(tag):
		SetUiValueByTag(c.uiManager, tag, GetModelValueByTag[int](c.model, tag))
	case c.uiManager.IsCheckboxUIItem(tag):
		SetUiValueByTag(c.uiManager, tag, GetModelValueByTag[bool](c.model, tag))
	default:
		log.Fatalf("Controller: updateUIFromModel - unknown tag = %s", tag)
		return
	}
}

func (c *Controller) updateMenuItems() {
	c.uiManager.SetMenuItemEnabledState(MenuSaveModelTag, c.model.dirty || c.model.fromFilePath == "")
	c.uiManager.SetMenuItemEnabledState(MenuGenGrblTag, c.model.GetHeightMap() != nil)
}

func (c *Controller) getGrblOutputFile(dir string) *os.File {
	dlg := fui.NewDialog("Export GRBL Code")
	filename, _ := dlg.SaveToGrblFile(dir)

	f, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		dlg := fui.NewDialog("GRBL Error")
		dlg.ShowErrorDialog("Error opening GRBL output file %s: err = %s", filename, err.Error())
		return nil
	}

	f.Truncate(0) // Just in case we're overwriting an existing file.
	return f
}

func (c *Controller) getCarvingSampler(
	matDim, carvDim geom.Size2,
	carvOrigin geom.Pt2,
	invertImage bool,
	topZ, bottomZ, toolDiameter float64) hmap.ScalarGridSampler {

	imgGray := c.getHeightMapImageForSampler()
	imgMode := c.model.GetIntValue(CarvDirectionTag)

	xform := geom.NewXformCache(
		float32(matDim.W), float32(matDim.H),
		float32(carvDim.W), float32(carvDim.H),
		float32(carvOrigin.X), float32(carvOrigin.Y),
		imgGray.Bounds().Dx(), imgGray.Bounds().Dy(), imgMode)
	sampler := hmap.NewPixelDepthSampler(xform.GetMc2NicXform(), carvOrigin, carvDim, imgGray)
	sampler.EnableInvertImage(invertImage)

	if c.useMeshSampler {
		tmesh := mesh.NewTriangleMesh(carvOrigin, carvOrigin.Add(geom.NewVec2(carvDim.W, carvDim.H)),
			bottomZ, topZ, sampler)
		sampler = mesh.NewMeshSamplerWithBallCutter(tmesh, 0.5*toolDiameter)
	}

	return sampler
}

// Return a gray-scale image for the current model height map, mirroring along X and Y
// as needed.
func (c *Controller) getHeightMapImageForSampler() *image.Gray {
	heightMap := c.model.GetHeightMap()
	mirrorX := c.model.GetBoolValue(ImgMirrorXTag)
	mirrorY := c.model.GetBoolValue(ImgMirrorYTag)
	if mirrorX && mirrorY {
		heightMap = imaging.Rotate180(heightMap)
	} else if mirrorX {
		heightMap = imaging.FlipH(heightMap)
	} else if mirrorY {
		heightMap = imaging.FlipV(heightMap)
	}

	return util.ImageToGrayImage(heightMap)
}

func (c *Controller) showProgressDialog(title, subtitle string) *widget.PopUp {
	progress := widget.NewProgressBarInfinite()
	popup := widget.NewModalPopUp(
		container.NewVBox(widget.NewCard(title, subtitle, progress)),
		c.mainWindow.Canvas())
	popup.Show()
	return popup
}

func carverToolTypeFromModelToolType(modelToolType int) int {
	switch modelToolType {
	case ToolTypeBallNose:
		return carv.ToolTypeBallPoint
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

func carverFinishModeFromModelFinishMode(modelFinishMode int) int {
	switch modelFinishMode {
	case FinishModeFirstDirectionOnly:
		return carv.FinishPassModeAlongFirstDirOnly
	case FinishModeLastDirectionOnly:
		return carv.FinishPassModeAlongLastDirOnly
	case FinishModeInAllDirections:
		return carv.FinishPassModeAlongAllDirs
	default:
		log.Fatalln("Unknown model finish-pass mode")
		return 0
	}
}
