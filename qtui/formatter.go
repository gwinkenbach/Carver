package qtui

import "fmt"

type formatter interface {
	Format(value interface{}) string
}

// Declare conformity with formatter interface
var _ formatter = (*Float32Formatter)(nil)
var _ formatter = (*SelectFormatter)(nil)
var _ formatter = (*BoolFormatter)(nil)

// Float32Formatter maps float32 value to strings through printf.
type Float32Formatter struct {
	format string
}

func NewFloat32Formatter(format string) *Float32Formatter {
	return &Float32Formatter{
		format: format,
	}
}

func (f *Float32Formatter) Format(value interface{}) string {
	v, ok := value.(float32)
	if ok {
		return fmt.Sprintf(f.format, v)
	}
	return "error: wrong type"
}

// SelectFormatter maps integer choice indices to strings.
type SelectFormatter struct {
	choices []string
}

func NewSelectFormatter(choice []string) *SelectFormatter {
	return &SelectFormatter{
		choices: choice,
	}
}

func (f *SelectFormatter) Format(value interface{}) string {
	v, ok := value.(int)
	if !ok {
		return "error: wrong type"
	}
	if v < 0 && v >= len(f.choices) {
		return "error: out of range"
	}

	return f.choices[v]
}

// BoolFormatter is a trivial true/false formatter.
type BoolFormatter struct {
}

func NewBoolFormatter() *BoolFormatter {
	return &BoolFormatter{}
}

func (f *BoolFormatter) Format(value interface{}) string {
	v, ok := value.(bool)
	if !ok {
		return "error: wrong type"
	}

	s := "false"
	if v {
		s = "true"
	}
	return s
}
