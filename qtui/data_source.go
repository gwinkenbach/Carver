package qtui

import (
	"log"

	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/widgets"
)

type dataSource interface {
	SetOnChange(oc onDataChanged)
	UpdateValue(val interface{})
	OnValidationResult(result int)
	HasValidInput() bool
}

// Declare conformity with validator interface
var _ dataSource = (*Float32LineEditDataSource)(nil)
var _ dataSource = (*SelectDataSource)(nil)
var _ dataSource = (*BoolDataSource)(nil)

// Float32LineEditDataSource wraps an QT LineEdit widget as a data source.
type Float32LineEditDataSource struct {
	entry *widgets.QLineEdit
}

func NewFloat32LineEditDataSource(entry *widgets.QLineEdit) *Float32LineEditDataSource {
	return &Float32LineEditDataSource{
		entry: entry,
	}
}

func (fed *Float32LineEditDataSource) SetOnChange(oc onDataChanged) {
	if oc == nil {
		fed.entry.ConnectTextEdited(nil)
	} else {
		fed.entry.ConnectTextEdited(func(val string) {
			oc(val)
		})
	}
}

func (fed *Float32LineEditDataSource) UpdateValue(val interface{}) {
	str, ok := val.(string)
	if !ok {
		log.Fatal("Float32LineEditDataSource: UpdateValue - invalid type")
	}

	fed.entry.SetText(str)
}

func (fed *Float32LineEditDataSource) OnValidationResult(result int) {
	if result == ValidationInvalid {
		fed.entry.SetStyleSheet("color: red;")
	} else {
		fed.entry.SetStyleSheet("color: black;")
	}
}

func (fed *Float32LineEditDataSource) HasValidInput() bool {
	return fed.entry.HasAcceptableInput()
}

// SelectDataSource wraps ComboBox widget as a data source.
type SelectDataSource struct {
	sel *widgets.QComboBox
}

func NewSelectDataSource(sel *widgets.QComboBox) *SelectDataSource {
	return &SelectDataSource{
		sel: sel,
	}
}

func (sd *SelectDataSource) SetOnChange(oc onDataChanged) {
	if oc == nil {
		sd.sel.ConnectCurrentTextChanged(nil)
	} else {
		sd.sel.ConnectCurrentTextChanged(func(txt string) {
			oc(txt)
		})
	}
}

func (sd *SelectDataSource) UpdateValue(val interface{}) {
	str, ok := val.(string)
	if !ok {
		log.Fatal("SelectDataSource: UpdateValue - invalid type")
	}

	sd.sel.SetCurrentText(str)
}

func (sd *SelectDataSource) OnValidationResult(result int) {
}

func (sd *SelectDataSource) HasValidInput() bool {
	return true
}

// BoolDataSource wraps on-off widget, such as a check box, as a data source.
type BoolDataSource struct {
	cb *widgets.QCheckBox
}

func NewBoolDataSource(cb *widgets.QCheckBox) *BoolDataSource {
	return &BoolDataSource{
		cb: cb,
	}
}

func (bd *BoolDataSource) SetOnChange(oc onDataChanged) {
	if oc == nil {
		bd.cb.ConnectStateChanged(nil)
	} else {
		bd.cb.ConnectStateChanged(func(s int) {
			val := "false"
			if core.Qt__CheckState(s) == core.Qt__Checked {
				val = "true"
			}
			oc(val)
		})
	}
}

func (bd *BoolDataSource) UpdateValue(val interface{}) {
	checked, ok := val.(string)
	if !ok {
		log.Fatal("BoolDataSource: UpdateValue - invalid type")
	}

	triState := core.Qt__Unchecked
	if checked == "true" {
		triState = core.Qt__Checked
	}
	bd.cb.SetCheckState(triState)
}

func (bd *BoolDataSource) OnValidationResult(result int) {
}

func (bd *BoolDataSource) HasValidInput() bool {
	return true
}
