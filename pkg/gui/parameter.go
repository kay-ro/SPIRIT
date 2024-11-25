package gui

import (
	"errors"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
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
	return &FilteredEntry{
		acceptContent: accept,
		allowedTyping: allowedRunes,
	}
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
	renderer     fyne.WidgetRenderer
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
		renderer:     nil,
	}
	p.renderer = newParameterRenderer(&p)
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
func (p *Parameter) Resize(size fyne.Size) {
	p.BaseWidget.Resize(size)
	p.renderer.Layout(size)
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
	return newParameterRenderer(this)
}

type parameterRenderer struct {
	fyne.WidgetRenderer
	parameter *Parameter
}

func newParameterRenderer(p *Parameter) fyne.WidgetRenderer {
	println("test")
	return &parameterRenderer{
		parameter: p,
	}
}
func (r *parameterRenderer) Layout(size fyne.Size) {
	bntW := float32(0.1)
	valueW := float32(0.6)
	extreamsW := float32(0.3)

	lockedSize := fyne.NewSize(15, 15)
	r.parameter.locked.Resize(lockedSize)
	r.parameter.locked.Move(fyne.NewPos(
		(size.Width*bntW)-(lockedSize.Width/2),
		(size.Height/2)-(lockedSize.Height/2)))

	lblSize := r.parameter.name.MinSize()
	r.parameter.name.Resize(lblSize)
	r.parameter.name.Move(fyne.NewPos(
		(size.Width/2)-(lblSize.Width/2),
		lblSize.Height))

	valEntrySize := fyne.NewSize(size.Width*valueW, size.Height-lblSize.Height)
	r.parameter.valEntry.Resize(valEntrySize)
	r.parameter.valEntry.Move(fyne.NewPos(
		size.Width*bntW+lockedSize.Width,
		size.Height-valEntrySize.Height))

	extEntrySize := fyne.NewSize(size.Width*extreamsW, (size.Height-lblSize.Height)/2)
	r.parameter.minEntry.Resize(extEntrySize)
	r.parameter.minEntry.Move(fyne.NewPos(
		size.Width-extEntrySize.Width,
		lblSize.Height+extEntrySize.Height))

	r.parameter.maxEntry.Resize(extEntrySize)
	r.parameter.maxEntry.Move(fyne.NewPos(
		size.Width-extEntrySize.Width,
		lblSize.Height+(extEntrySize.Height*2)))

	r.parameter.objects = []fyne.CanvasObject{
		r.parameter.name,
		r.parameter.locked,
		r.parameter.valEntry,
		r.parameter.minEntry,
		r.parameter.maxEntry,
	}
}
func (r *parameterRenderer) MinSize() fyne.Size {
	return r.parameter.MinSize()
}
func (r *parameterRenderer) Objects() []fyne.CanvasObject {
	return r.parameter.objects
}
func (r *parameterRenderer) Refresh() {
	r.Layout(r.parameter.Size())
	canvas.Refresh(r.parameter)
}
func (r *parameterRenderer) Destroy() {

}
