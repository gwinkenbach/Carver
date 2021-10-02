package qtui

import (
	"fmt"
	"regexp"
	"strconv"
)

type validator interface {
	Validate(string) int
	GetValidatedValue() interface{}
	GetDefaultValue() interface{}
}

// Validate returns one of these constants.
const (
	ValidationInvalid      = 0
	ValidationIntermediate = 1
	ValidationAcceptable   = 2
)

// Declare conformity with validator interface
var _ validator = (*Float32Validator)(nil)
var _ validator = (*SelectValidator)(nil)
var _ validator = (*BoolValidator)(nil)

// Float32Validator checks that a string maps to a float32 via regex and that
// the resulting value is within a given range.
type Float32Validator struct {
	min   float32
	max   float32
	regex *regexp.Regexp

	val float32
}

func NewFloat32Validator(reg string, min, max float32) *Float32Validator {
	return &Float32Validator{
		min:   min,
		max:   max,
		regex: regexp.MustCompile(reg),
	}
}

func (f *Float32Validator) Validate(s string) int {
	fmt.Printf("Validate: %s - ", s)

	match := f.regex.FindStringSubmatch(s)
	if match == nil || len(match) < 2 {
		fmt.Printf("invalid\n")
		return ValidationInvalid
	}

	val64, err := strconv.ParseFloat(match[1], 32)
	if err != nil {
		fmt.Printf("interm\n")
		return ValidationIntermediate
	}

	f.val = float32(val64)
	if f.val < f.min {
		fmt.Printf("interm, too small\n")
		return ValidationIntermediate
	}
	if f.val > f.max {
		fmt.Printf("interm, to big\n")
		return ValidationInvalid
	}

	fmt.Printf("validated\n")
	return ValidationAcceptable
}

func (f *Float32Validator) GetValidatedValue() interface{} {
	return f.val
}

func (f *Float32Validator) GetDefaultValue() interface{} {
	return f.min
}

// SelectValidator validator for a selector widget.
type SelectValidator struct {
	choices []string
	val     int
}

func NewSelectValidator(choices []string) *SelectValidator {
	return &SelectValidator{
		choices: choices,
	}
}

func (f *SelectValidator) Validate(s string) int {
	for i, c := range f.choices {
		if c == s {
			f.val = i
			return ValidationAcceptable
		}
	}

	return ValidationInvalid
}

func (f *SelectValidator) GetValidatedValue() interface{} {
	return f.val
}

func (f *SelectValidator) GetDefaultValue() interface{} {
	return 0
}

// BoolValidator is a trivial true/false validator.
type BoolValidator struct {
	val bool
}

func NewBoolValidator() *BoolValidator {
	return &BoolValidator{}
}

func (f *BoolValidator) Validate(s string) int {
	f.val = (s == "true")
	return ValidationAcceptable
}

func (f *BoolValidator) GetValidatedValue() interface{} {
	return f.val
}

func (f *BoolValidator) GetDefaultValue() interface{} {
	return false
}
