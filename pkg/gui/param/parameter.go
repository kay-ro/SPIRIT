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
	relatives []*Parameter[T]

	// use for fit checkbox
	checkbox *widget.Check
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
		binding: binding.NewString(),
		config:  config,
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
func (f *Parameter[T]) GetRelatives() []*Parameter[T] {
	return f.relatives
}

// set relative parameters (min, max, checkboxes, ...)
func (f *Parameter[T]) SetRelatives(relatives ...*Parameter[T]) {
	f.relatives = relatives
}

// Widget returns the widget (drawable element for the gui) of the parameter
func (f *Parameter[T]) Widget() *widget.Entry {
	return f.widget
}

// SetCheckbox sets the checkbox of the parameter
//
// returns always true if checkbox isn't set
func (f *Parameter[T]) IsChecked() bool {
	if f.checkbox == nil {
		return true
	}

	return f.checkbox.Checked
}
