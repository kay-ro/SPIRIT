package custom_bindings

import (
	"fmt"
	"fyne.io/fyne/v2/data/binding"
	"math"
	"strconv"
)

// LazyFloatToString custom replacement for binging.FloatToString implementation, that does not interfere with user input to ensure convertibility
// calls referenced binding only when input parsable and does not return an error when not
type LazyFloatToString struct {
	refBinding   binding.Float
	defaultValue binding.Float
}

func (l *LazyFloatToString) AddListener(listener binding.DataListener) {
	l.refBinding.AddListener(listener)
}

func (l *LazyFloatToString) RemoveListener(listener binding.DataListener) {
	l.refBinding.RemoveListener(listener)
}

func NewLazyFloatToString(refBinding binding.Float, defaultValue binding.Float) *LazyFloatToString {
	return &LazyFloatToString{
		refBinding:   refBinding,
		defaultValue: defaultValue,
	}
}
func (l *LazyFloatToString) Get() (string, error) {
	if get, err := l.refBinding.Get(); err != nil {
		return "", err
	} else {
		if get == math.MaxFloat64 || get == -math.MaxFloat64 {
			return "", nil
		}
		return fmt.Sprint(get), nil
	}
}

func (l *LazyFloatToString) Set(s string) error {
	if float, err := strconv.ParseFloat(s, 64); err == nil {
		return l.refBinding.Set(float)
	} else {
		if l.defaultValue != nil {
			if get, err := l.defaultValue.Get(); err == nil {
				return l.refBinding.Set(get)
			} else {
				return err
			}
		}
		return nil
	}
}
