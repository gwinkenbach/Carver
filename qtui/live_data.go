package qtui

import (
	"fmt"
	"log"
)

type onDataChanged func(val interface{})

type liveDataObserver interface {
	notifyDataChanged(label string)
}

type baseLiveData interface {
	updateValue(val interface{})
}

type liveData struct {
	baseLiveData

	o liveDataObserver

	label string
	ds    dataSource
	ft    formatter
	vl    validator

	onDataChangeDisableCount int

	value interface{}
}

func NewLiveData(label string, f formatter, v validator, d dataSource) *liveData {
	ld := &liveData{
		label:                    label,
		ft:                       f,
		vl:                       v,
		onDataChangeDisableCount: 0,
	}

	ld.SetDataSource(d)
	ld.value = v.GetDefaultValue()
	return ld
}

func (ld *liveData) SetDataSource(ds dataSource) {
	if ld.ds != nil {
		ld.ds.SetOnChange(nil)
	}

	ld.ds = ds
	if ld.ds != nil {
		ld.ds.SetOnChange(func(val interface{}) {
			ld.onDataChanged(val)
		})
	}
}

func (ld *liveData) SetObserver(o liveDataObserver) {
	ld.o = o
}

func (ld *liveData) updateValue(val interface{}) {
	ld.disableOnDataChangedResponse()
	defer ld.enableOnDataChangedResponse()

	if ld.ft != nil {
		ld.ds.UpdateValue(ld.ft.Format(val))
	} else {
		ld.ds.UpdateValue(val)
	}
	ld.value = val
}

func (ld *liveData) getValue() interface{} {
	return ld.value
}

func (ld *liveData) onDataChanged(val interface{}) {
	show := false
	if ld.label == ItemBlackCarvingDepth {
		fmt.Printf("On data changed: %s\n", ld.label)
		show = true
	}

	if ld.vl != nil && ld.onDataChangeDisableCount == 0 {
		// A validator implies string type for the value.
		str, ok := val.(string)
		if !ok {
			log.Fatal("liveData: onDataChanged - Invalid type")
		}

		result := ld.vl.Validate(str)
		ld.ds.OnValidationResult(result)
		if result != ValidationAcceptable {
			if show {
				fmt.Printf("*** Validation failed\n")
			}
			return
		}

		ld.value = ld.vl.GetValidatedValue()
		if ld.o != nil {
			if show {
				fmt.Printf("*** Notify\n")
			}
			ld.o.notifyDataChanged(ld.label)
		}
	} else {
		ld.value = val
	}
}

func (ld *liveData) disableOnDataChangedResponse() {
	ld.onDataChangeDisableCount++
}

func (ld *liveData) enableOnDataChangedResponse() {
	if ld.onDataChangeDisableCount > 0 {
		ld.onDataChangeDisableCount--
	}
}
