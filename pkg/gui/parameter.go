package gui

import (
	"errors"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"physicsGUI/pkg/util/option"
	"slices"
	"strconv"
)

type FilteredEntry struct {
	widget.Entry
	allowedTyping []rune
}

func NewFilteredEntry(allowedRunes []rune) *FilteredEntry {
	entry := &FilteredEntry{
		allowedTyping: allowedRunes,
	}
	entry.ExtendBaseWidget(entry)
	return entry
}

func (e *FilteredEntry) TypedRune(r rune) {
	if slices.Contains(e.allowedTyping, r) {
		e.Entry.TypedRune(r)
	}
}

type Parameter struct {
	widget.BaseWidget
	window       fyne.Window
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

func NewParameter(name string, defaultValue float64, window fyne.Window) *Parameter {
	numberRunes := []rune("0123456789.e-+")

	p := Parameter{
		name:         widget.NewLabel(name),
		window:       window,
		defaultValue: defaultValue,
		valEntry:     NewFilteredEntry(numberRunes),
		minEntry:     NewFilteredEntry(numberRunes),
		maxEntry:     NewFilteredEntry(numberRunes),
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
func (p *Parameter) Refresh() {
	p.valEntry.Refresh()
	p.minEntry.Refresh()
	p.maxEntry.Refresh()
	p.locked.Refresh()
	p.BaseWidget.Refresh()
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
		dialog.NewError(errors.New("Float_Parsing_Error: Error while parsing Min input to float this should never happen because of rune filter"), this.window)
		return option.None[float64]()
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
	parsable := func(s string) error {
		if s == "" {
			return nil
		}
		_, err := strconv.ParseFloat(s, 64)
		return err
	}

	this.valEntry.MultiLine = false
	this.valEntry.Validator = parsable
	this.valEntry.PlaceHolder = fmt.Sprintf("%f", this.defaultValue)
	this.valEntry.Scroll = container.ScrollNone
	this.maxEntry.MultiLine = false
	this.maxEntry.Validator = parsable
	this.maxEntry.PlaceHolder = "Max"
	this.maxEntry.Scroll = container.ScrollNone
	this.minEntry.MultiLine = false
	this.minEntry.Validator = parsable
	this.minEntry.PlaceHolder = "Min"
	this.minEntry.Scroll = container.ScrollNone

	this.maxEntry.Refresh()
	this.minEntry.Refresh()
	this.valEntry.Refresh()
	cntRight := container.NewVBox(this.maxEntry, this.minEntry)
	return widget.NewSimpleRenderer(container.NewVBox(this.name, container.NewHBox(this.locked, this.valEntry, cntRight)))
}

type Profile struct {
	widget.BaseWidget
	name      *widget.Entry
	removeBtn *widget.Button
	Roughness *Parameter
	Edensity  *Parameter
	Thickness *Parameter
}

func NewProfile(name string, roughnessName string, defaultRoughness float64, edensityName string, defaultEdensity float64, thicknessName string, defaultThickness float64, window fyne.Window) *Profile {
	profile := &Profile{
		name:      widget.NewEntry(),
		removeBtn: widget.NewButton("🗑", func() {}),
		Roughness: NewParameter(roughnessName, defaultRoughness, window),
		Edensity:  NewParameter(edensityName, defaultEdensity, window),
		Thickness: NewParameter(thicknessName, defaultThickness, window),
	}
	profile.name.Text = name
	profile.ExtendBaseWidget(profile)
	return profile
}
func (this *Profile) Refresh() {
	this.Edensity.Refresh()
	this.Thickness.Refresh()
	this.Roughness.Refresh()
	this.BaseWidget.Refresh()
}

func (this *Profile) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(container.NewVBox(container.NewBorder(nil, nil, nil, this.removeBtn, this.name), this.Roughness, this.Edensity, this.Thickness))
}

type ProfilePanel struct {
	widget.BaseWidget
	base      *Profile
	bulk      *Profile
	profiles  []*Profile
	addButton *widget.Button
	renderer  *ProfilePanelRenderer
}

func (this *ProfilePanel) Resize(size fyne.Size) {
	if this.renderer != nil {
		this.renderer.Layout(size)
	}
	this.BaseWidget.Resize(size)
}

func NewProfilePanel(window fyne.Window, profiles ...*Profile) *ProfilePanel {
	//TODO read default values from some settings file
	defaultRoughness := 10.0
	defaultEdensity := 1.0
	defaultThickness := 1.0

	base := NewProfile("Base", "Roughness Base/Slab1", defaultRoughness, "Edensity Base", defaultEdensity, "Thickness Base", defaultThickness, window)
	base.removeBtn = widget.NewButton("🗑", func() {
		base.Edensity.valEntry.Text = ""
		base.Edensity.minEntry.Text = ""
		base.Edensity.maxEntry.Text = ""
		base.Roughness.valEntry.Text = ""
		base.Roughness.minEntry.Text = ""
		base.Roughness.maxEntry.Text = ""
		base.Thickness.valEntry.Text = ""
		base.Thickness.minEntry.Text = ""
		base.Thickness.maxEntry.Text = ""
		base.Refresh()
	})
	bulk := NewProfile("Bulk", "Roughness Bulk", 0.0, "Edensity Bulk", defaultEdensity, "Thickness Bulk", defaultThickness, window)
	bulk.removeBtn = widget.NewButton("🗑", func() {
		bulk.Edensity.valEntry.Text = ""
		bulk.Edensity.minEntry.Text = ""
		bulk.Edensity.maxEntry.Text = ""
		bulk.Thickness.valEntry.Text = ""
		bulk.Thickness.minEntry.Text = ""
		bulk.Thickness.maxEntry.Text = ""
		bulk.Refresh()
	})
	bulk.Roughness.valEntry.Disable()
	bulk.Roughness.minEntry.Disable()
	bulk.Roughness.maxEntry.Disable()
	bulk.Roughness.locked.Disable()

	p := &ProfilePanel{
		base:     base,
		bulk:     bulk,
		profiles: profiles,
	}
	if profiles == nil || len(profiles) == 0 {
		profileName := fmt.Sprintf("Slab 1")
		roughnessName := fmt.Sprintf("Rougthness Slab 1/Bulk")
		edensityName := fmt.Sprintf("Edensity Slab 1")
		thicknessName := fmt.Sprintf("Thickness Slab 1")
		baseProfile := NewProfile(profileName, roughnessName, defaultRoughness, edensityName, defaultEdensity, thicknessName, defaultThickness, window)
		p.AddProfile(baseProfile)
	}

	p.addButton = widget.NewButton("+", func() {
		profileName := fmt.Sprintf("Slab %d", len(p.profiles)+1)
		roughnessName := fmt.Sprintf("Rougthness Slab %d/Bulk", len(p.profiles)+1)
		edensityName := fmt.Sprintf("Edensity Slab %d", len(p.profiles)+1)
		thicknessName := fmt.Sprintf("Thickness Slab %d", len(p.profiles)+1)
		newP := NewProfile(profileName, roughnessName, defaultRoughness, edensityName, defaultEdensity, thicknessName, defaultThickness, window)
		p.profiles[len(p.profiles)-1].Roughness.name.SetText(fmt.Sprintf("Rougthness Slab %d/Slab %d", len(p.profiles), len(p.profiles)+1))
		p.AddProfile(newP) //TODO add Settings for default values
	})
	return p
}
func (this *ProfilePanel) AddProfile(profile *Profile) {
	this.profiles = append(this.profiles, profile)
	profile.removeBtn = widget.NewButton("🗑", func() {
		this.RemoveProfile(profile)
	})
	if this.renderer != nil {
		this.renderer.Update()
	}
}
func (this *ProfilePanel) RemoveProfile(profile *Profile) {
	i := slices.Index(this.profiles, profile)
	if i >= 0 {
		if len(this.profiles) > 1 {
			if i != len(this.profiles)-1 {
				for j := i + 1; j < len(this.profiles); j++ {
					profileName := fmt.Sprintf("Slab %d", j)
					roughnessName := fmt.Sprintf("Rougthness Slab %d/%d", j, j+1)
					edensityName := fmt.Sprintf("Edensity Slab %d", j)
					thicknessName := fmt.Sprintf("Thickness Slab %d", j)
					this.profiles[j].name.SetText(profileName)
					this.profiles[j].Roughness.name.SetText(roughnessName)
					this.profiles[j].Edensity.name.SetText(edensityName)
					this.profiles[j].Thickness.name.SetText(thicknessName)
				}
				this.profiles = append(this.profiles[:i], this.profiles[i+1:]...)
				this.profiles[len(this.profiles)-1].Roughness.name.SetText(fmt.Sprintf("Rougthness Slab %d/Bulk", len(this.profiles)))
			}
		} else {
			this.profiles[i].Edensity.valEntry.Text = ""
			this.profiles[i].Edensity.minEntry.Text = ""
			this.profiles[i].Edensity.maxEntry.Text = ""
			this.profiles[i].Roughness.valEntry.Text = ""
			this.profiles[i].Roughness.minEntry.Text = ""
			this.profiles[i].Roughness.maxEntry.Text = ""
			this.profiles[i].Thickness.valEntry.Text = ""
			this.profiles[i].Thickness.minEntry.Text = ""
			this.profiles[i].Thickness.maxEntry.Text = ""
			this.profiles[i].Refresh()
		}
	}
	this.renderer.Update()
}

func (this *ProfilePanel) CreateRenderer() fyne.WidgetRenderer {
	this.renderer = NewProfilePanelRenderer(this)
	return this.renderer
}

type ProfilePanelRenderer struct {
	obj    *ProfilePanel
	layout fyne.WidgetRenderer
}

func (p *ProfilePanelRenderer) Update() {
	objects := make([]fyne.CanvasObject, len(p.obj.profiles)+3)
	for i, profile := range p.obj.profiles {
		objects[i+1] = profile
	}
	objects[0] = p.obj.base
	objects[len(objects)-2] = p.obj.addButton
	objects[len(objects)-1] = p.obj.bulk

	center := container.NewHBox(objects...)
	cnt := container.NewHScroll(center)
	p.layout = widget.NewSimpleRenderer(cnt)
	p.Refresh()
}

func (p *ProfilePanelRenderer) Destroy() {}

func (p *ProfilePanelRenderer) Layout(size fyne.Size) {
	p.layout.Layout(size)
}

func (p *ProfilePanelRenderer) MinSize() fyne.Size {
	return p.layout.MinSize()
}

func (p *ProfilePanelRenderer) Objects() []fyne.CanvasObject {
	if p.layout == nil {
		p.Update()
	}
	return p.layout.Objects()
}

func (p *ProfilePanelRenderer) Refresh() {
	p.layout.Layout(p.obj.Size())
	p.layout.Refresh()
}

func NewProfilePanelRenderer(obj *ProfilePanel) *ProfilePanelRenderer {
	return &ProfilePanelRenderer{
		obj:    obj,
		layout: nil,
	}
}
