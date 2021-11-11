package fui

import (
	"image"
	"os"

	"github.com/sqweek/dialog"
)

// Dialog is a helper for showing and handling various modal dialogs. 
type Dialog struct {
	title string
}

// NewDialog v=creates and return a new dialog object with the given title.
func NewDialog(title string) Dialog {
	return Dialog{title: title}
}

// OpenAndLoadImageFile shows a file-open dialog for the user to select either a JPG or PNG
// file. When the user select a valid image file, the image is loaded into an image.Image object
// and returned, together with its filename. Otherwise and error is returned, which may be 
// dialog.ErrCancelled.
func (d Dialog) OpenAndLoadImageFile() (img image.Image, filename string, err error) {
	img = nil

	dlg := dialog.File()
	dlg.Title(d.title)
	dlg.Filter("Image File", "png", "jpg")
	filename, err = dlg.Load()
	if err != nil {
		return
	}

	var f *os.File
	f, err = os.Open(filename)
	if err != nil {
		dialog.Message("Error opening image: %v", err).Error()
		return
	}
	defer f.Close()

	img, _, err = image.Decode(f)
	if err != nil {
		dialog.Message("Error loading image: %v", err).Error()
		return
	}

	return
}

// SaveToCarverFile shows the save-file dialog for saving a carver file with extension "carv".
// Returns the filename (full path to the file to save into) and an error. When the user cancels
// the dialog the error returned is dialog.ErrCancelled.
func (d Dialog) SaveToCarverFile(startFromDir string) (filename string, err error) {
	dlg := dialog.File()
	dlg.Title(d.title)
	dlg.Filter("Carver File", "carv")
	if startFromDir != "" {
		dlg.SetStartDir(startFromDir)
	}

	return dlg.Save()
}

// SaveToGrblFile shows the save-file dialog for saving a GRBL file with extension "gcode".
// Returns the filename (full path to the file to save into) and an error. When the user cancels
// the dialog the error returned is dialog.ErrCancelled.
func (d Dialog) SaveToGrblFile(startFromDir string) (filename string, err error) {
	dlg := dialog.File()
	dlg.Title(d.title)
	dlg.Filter("Gcode file", "gcode")
	if startFromDir != "" {
		dlg.SetStartDir(startFromDir)
	}

	return dlg.Save()
}

// OpenCarverFile shows the open-file dialog for the user to select a carver file (with extension
// "carv"). Returns the full path to the selected file as filename or an error, which may be
// dialog.ErrCancelled.
func (d Dialog) OpenCarverFile(startFromDir string) (filename string, err error) {
	dlg := dialog.File()
	dlg.Title(d.title)
	dlg.Filter("Carver File", "carv")
	if startFromDir != "" {
		dlg.SetStartDir(startFromDir)
	}

	return dlg.Load()
}

// ShowYesNoDialog shows an alert with the given message and "yes" and "no" buttons. Returns true
// when the user clicks on the "yes" button.
func (d Dialog) ShowYesNoDialog(message string) (yes bool) {
	dlg := dialog.Message(message)
	dlg.Title(d.title)
	return dlg.YesNo()
}

// ShowErrorDialog shows an error dialog with the given formatted message and an OK button.
func (d Dialog) ShowErrorDialog(format string, args ...interface{}) {
	dlg := dialog.Message(format, args...)
	dlg.Title(d.title)
	dlg.Error()
}
