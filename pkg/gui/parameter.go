package gui

import (
	"errors"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"physicsGUI/pkg/util/option"
	"slices"
	"strconv"
)

type FilteredEntry struct {
	widget.Entry
	allowedTyping []rune
	acceptContent func(s string) bool
}

func NewFilteredEntry(allowedRunes []rune, accept func(s string) bool) *FilteredEntry {
	entry := &FilteredEntry{
		acceptContent: accept,
		allowedTyping: allowedRunes,
	}
	entry.ExtendBaseWidget(entry)
	return entry
}

func (e *FilteredEntry) TypedRune(r rune) {
	if slices.Contains(e.allowedTyping, r) {
		old := e.Text
		e.Entry.TypedRune(r)
		if !e.acceptContent(e.Text) {
			e.Text = old
		}
	}
}

type Parameter struct {
	widget.BaseWidget
	name         *widget.Label
	defaultValue float64
	valEntry     *FilteredEntry
	minEntry     *FilteredEntry
	maxEntry     *FilteredEntry
	locked       *widget.Check
	OnChanged    func()
	maxSize      fyne.Size
	minSize      fyne.Size
	objects      []fyne.CanvasObject // weil im renderer unterschiedliche Structs für rendern und Layout verwendet werden warum auch immer
}

func NewParameter(name string, defaultValue float64) *Parameter {
	numberRunes := []rune("0123456789.e-+")
	numberAccept := func(s string) bool {
		if s == "" {
			return true
		}
		_, err := strconv.ParseFloat(s, 64)
		return err == nil
	}

	p := Parameter{
		name:         widget.NewLabel(name),
		defaultValue: defaultValue,
		valEntry:     NewFilteredEntry(numberRunes, numberAccept),
		minEntry:     NewFilteredEntry(numberRunes, numberAccept),
		maxEntry:     NewFilteredEntry(numberRunes, numberAccept),
		locked:       widget.NewCheck("", nil),
		OnChanged:    func() {},
		minSize:      fyne.NewSize(240, 80),
		maxSize:      fyne.NewSize(400, 200),
	}
	p.locked.OnChanged = func(_ bool) {
		p.OnChanged()
	}
	p.valEntry.OnChanged = func(_ string) {
		p.OnChanged()
	}
	p.minEntry.OnChanged = func(_ string) {
		p.OnChanged()
	}
	p.maxEntry.OnChanged = func(_ string) {
		p.OnChanged()
	}
	return &p
}
func (this *Parameter) MinSize() fyne.Size {
	altMinX := max(this.name.MinSize().Width,
		this.valEntry.MinSize().Width+
			this.locked.MinSize().Width+
			max(this.minEntry.MinSize().Width, this.maxEntry.MinSize().Width))
	altMinY := this.name.MinSize().Height +
		max(this.valEntry.MinSize().Height,
			this.locked.MinSize().Height,
			max(this.minEntry.MinSize().Height+this.maxEntry.MinSize().Height))
	return fyne.NewSize(max(altMinX, this.minSize.Width), max(altMinY, this.minSize.Height))
}

func (this *Parameter) GetValue() float64 {
	if this.valEntry.Text == "" {
		this.valEntry.SetText(fmt.Sprintf("%f", this.defaultValue))
		return this.defaultValue
	}
	val, err := strconv.ParseFloat(this.valEntry.Text, 64)
	if err != nil {
		panic(errors.Join(errors.New("Float_Parsing_Error: Error while parsing Value input to float this should never happen because of rune filter"), err))
	}
	return val
}
func (this *Parameter) GetMin() option.Option[float64] {
	if this.valEntry.Text == "" {
		return option.None[float64]()
	}
	val, err := strconv.ParseFloat(this.valEntry.Text, 64)
	if err != nil {
		panic(errors.Join(errors.New("Float_Parsing_Error: Error while parsing Min input to float this should never happen because of rune filter"), err))
	}
	return option.Some[float64](&val) // Maybe change for better memory layout
}
func (this *Parameter) GetMax() option.Option[float64] {
	if this.valEntry.Text == "" {
		return option.None[float64]()
	}
	val, err := strconv.ParseFloat(this.valEntry.Text, 64)
	if err != nil {
		panic(errors.Join(errors.New("Float_Parsing_Error: Error while parsing Max input to float this should never happen because of rune filter"), err))
	}
	return option.Some[float64](&val) // Maybe change for better memory layout
}
func (this *Parameter) IsFixed() bool {
	return this.locked.Checked
}

func (this *Parameter) CreateRenderer() fyne.WidgetRenderer {
	this.valEntry.MultiLine = false
	this.valEntry.Validator = nil
	this.valEntry.PlaceHolder = fmt.Sprintf("%f", this.defaultValue)
	this.valEntry.Scroll = container.ScrollNone
	this.maxEntry.MultiLine = false
	this.maxEntry.Validator = nil
	this.maxEntry.PlaceHolder = "Max"
	this.maxEntry.Scroll = container.ScrollNone
	this.minEntry.MultiLine = false
	this.minEntry.Validator = nil
	this.minEntry.PlaceHolder = "Min"
	this.minEntry.Scroll = container.ScrollNone

	this.maxEntry.Refresh()
	this.minEntry.Refresh()
	this.valEntry.Refresh()
	cntRight := container.NewVBox(this.maxEntry, this.minEntry)
	return widget.NewSimpleRenderer(container.NewVBox(this.name, container.NewHBox(this.locked, this.valEntry, cntRight)))
}
