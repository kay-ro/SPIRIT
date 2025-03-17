package param

import (
	"physicsGUI/pkg/trigger"

	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
)

type Parameter[T any] struct {
	binding binding.String
	widget  *widget.Entry
	config  *Config[T]

	// optional relative parameters
	relatives map[string]*Parameter[T]

	// use for fit checkbox
	checkbox *widget.Check
}

type Parameters[T any] []*Parameter[T]

type ParameterType interface {
	Get() (any, error)
}

type Config[T any] struct {
	InitialValue T
	Validator    func(string) error
	Format       func(T) string
	Parser       func(string) (T, error)
}

// New creates a new parameter with the given configuration
func New[T any](config *Config[T]) *Parameter[T] {
	f := &Parameter[T]{
		binding:   binding.NewString(),
		config:    config,
		relatives: make(map[string]*Parameter[T]),
	}

	// creates the widget with the binding
	f.widget = widget.NewEntryWithData(f.binding)
	f.widget.Validator = config.Validator
	f.widget.OnChanged = func(s string) {
		trigger.Recalc()
	}

	f.Set(config.InitialValue)

	return f
}

// Set sets the value of the parameter
func (f *Parameter[T]) Set(data T) error {
	return f.binding.Set(f.config.Format(data))
}

// Get returns the value of the parameter
func (f *Parameter[T]) Get() (T, error) {
	var null T

	s, err := f.binding.Get()
	if err != nil {
		return null, err
	}

	return f.config.Parser(s)
}

// return relative parameters (min, max, checkboxes, ...)
func (f *Parameter[T]) GetRelatives() map[string]*Parameter[T] {
	return f.relatives
}

// return relative parameter by name
func (f *Parameter[T]) GetRelative(key string) *Parameter[T] {
	return f.relatives[key]
}

// set relative parameters (min, max, checkboxes, ...)
func (f *Parameter[T]) SetRelative(key string, relatives *Parameter[T]) {
	f.relatives[key] = relatives
}

// Widget returns the widget (drawable element for the gui) of the parameter
func (f *Parameter[T]) Widget() *widget.Entry {
	return f.widget
}

// SetCheckbox sets the checkbox of the parameter
// returns true if checkbox isn't set
func (f *Parameter[T]) IsChecked() bool {
	if f.checkbox == nil {
		return true
	}

	return f.checkbox.Checked
}

func (f *Parameter[T]) SetCheck(checked bool) {
	if f.checkbox == nil {
	}
	f.checkbox.SetChecked(checked)
}
