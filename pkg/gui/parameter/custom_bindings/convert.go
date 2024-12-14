package custom_bindings

import (
	"fmt"
	"fyne.io/fyne/v2/data/binding"
	"math"
	"slices"
	"strconv"
)

// LazyFloatToString custom replacement for binging.FloatToString implementation, that does not interfere with user input to ensure convertibility
// calls referenced binding only when input parsable and does not return an error when not
type LazyFloatToString struct {
	refBinding   binding.Float
	defaultValue binding.Float
	listener     []binding.DataListener
	lastSet      float64
}

func (l *LazyFloatToString) AddListener(listener binding.DataListener) {
	l.listener = append(l.listener, listener)
}

func (l *LazyFloatToString) RemoveListener(listener binding.DataListener) {
	listenerIndex := slices.Index(l.listener, listener)
	if listenerIndex != -1 {
		l.listener = append(l.listener[:listenerIndex], l.listener[listenerIndex+1:]...)
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
		old, err := l.refBinding.Get()
		if err == nil && float == old {
			return nil
		}
		l.lastSet = float
		return l.refBinding.Set(float)
	} else {
		if l.defaultValue != nil {
			if get, err := l.defaultValue.Get(); err == nil {
				_ = l.refBinding.Set(get)
				l.lastSet = get
			} else {
				return err
			}
		}
		if s == "" {
			return nil
		} else {
			return err
		}
	}
}

// DataChanged does not use async que like binging.FloatToString implementation
func (l *LazyFloatToString) DataChanged() {
	if f, err := l.refBinding.Get(); err == nil && f != l.lastSet {
		for _, listener := range l.listener {
			listener.DataChanged()
		}
	}
}

func NewLazyFloatToString(refBinding binding.Float, defaultValue binding.Float) *LazyFloatToString {
	conv := &LazyFloatToString{
		refBinding:   refBinding,
		defaultValue: defaultValue,
	}
	refBinding.AddListener(conv)
	return conv
}
