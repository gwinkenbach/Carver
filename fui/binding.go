package fui

import (
	"errors"
	"regexp"
	"strconv"

	"fyne.io/fyne/v2/data/binding"
)

var _ binding.String = (*floatToString)(nil)

// Float-to-string binding with min-max range validation.
type floatToString struct {
	innerBinding binding.String
	regex        *regexp.Regexp
	minVal       float64
	maxVal       float64
}

// NewFloatRangeBinding creates and returns a new float-to-string binding with min-max validation.
func NewFloatRangeBinding(
	bindVal binding.Float,
	minVal, maxVal float64,
	reg string) binding.String {

	b := &floatToString{
		innerBinding: binding.FloatToString(bindVal),
		minVal:       minVal,
		maxVal:       maxVal,
	}

	if reg != "" {
		b.regex = regexp.MustCompile(reg)
	}

	return b
}

// NewFloatRangeBinding creates and returns a new float-to-string binding with min-max validation.
// The format string is used for float-to-string conversions.
func NewFloatRangeBindingWithFormat(
	bindVal binding.Float,
	minVal, maxVal float64,
	format string,
	reg string) binding.String {

	b := &floatToString{
		innerBinding: binding.FloatToStringWithFormat(bindVal, format),
		minVal:       minVal,
		maxVal:       maxVal,
	}

	if reg != "" {
		b.regex = regexp.MustCompile(reg)
	}

	return b
}

func (f *floatToString) AddListener(dl binding.DataListener) {
	f.innerBinding.AddListener(dl)
}

func (f *floatToString) RemoveListener(dl binding.DataListener) {
	f.innerBinding.RemoveListener(dl)
}

func (f *floatToString) Get() (string, error) {
	return f.innerBinding.Get()
}

func (f *floatToString) Set(s string) error {
	var val64 float64 = 0.0
	var err error

	// Map string to float, using the regex if available.
	if f.regex != nil {
		match := f.regex.FindStringSubmatch(s)
		if match == nil || len(match) < 2 {
			return errors.New("Invalid input")
		}

		val64, err = strconv.ParseFloat(match[1], 32)
		if err != nil {
			return errors.New("Invalid input")
		}
	} else {
		val64, err = strconv.ParseFloat(s, 32)
		if err != nil {
			return errors.New("Invalid input")
		}
	}

	// Validate value against range.
	if f.minVal < f.maxVal {
		if val64 < f.minVal {
			return errors.New("Too small")
		}
		if val64 > f.maxVal {
			return errors.New("Too big")
		}
	}

	return f.innerBinding.Set(s)
}
