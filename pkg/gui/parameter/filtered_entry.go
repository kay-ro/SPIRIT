package parameter

import (
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"slices"
	"strconv"
)

type FilteredEntry struct {
	widget.Entry
	allowedTyping []rune
}

// NewFilteredEntry creates a widget.Entry, that only accept the keys specified in allowedRunes
func NewFilteredEntry(allowedRunes ...rune) *FilteredEntry {
	entry := &FilteredEntry{
		allowedTyping: allowedRunes,
	}
	entry.Validator = parsableFloat
	entry.MultiLine = false
	entry.Scroll = container.ScrollNone
	entry.ExtendBaseWidget(entry)
	return entry
}

func (e *FilteredEntry) TypedRune(r rune) {
	if slices.Contains(e.allowedTyping, r) {
		e.Entry.TypedRune(r)
	}
}

func parsableFloat(s string) error {
	if s == "" {
		return nil
	}
	_, err := strconv.ParseFloat(s, 64)
	return err
}
