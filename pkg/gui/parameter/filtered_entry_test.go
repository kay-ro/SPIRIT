package parameter

import (
	"fyne.io/fyne/v2/test"
	"testing"
)

var uutFilteredEntry *FilteredEntry

func init() {
	uutFilteredEntry = NewFilteredEntry([]rune("1234567890e+-.")...)
}
func TestFilterEntryFilter(t *testing.T) {
	uutFilteredEntry.Text = ""
	test.Type(uutFilteredEntry, "7.8901234.e.123857213+-2e23e")
	if uutFilteredEntry.Text != "7.8901234.e.123857213+-2e23e" {
		t.Errorf("TestFilterEntryFilter() failed. Filtered Entry also removed allowed runes: Expected %s, got %s", "7.8901234.e.123857213+-2e23e", uutFilteredEntry.Text)
	}
	uutFilteredEntry.Text = ""
	test.Type(uutFilteredEntry, "!Da$ 15t e1n Test s@tz ohne jed3 bed3u7un9.")
	if uutFilteredEntry.Text != "15e1eee3e379." {
		t.Errorf("TestFilterEntryFilter() failed. Filtered Entry also removed allowed runes: Expected %s, got %s", "15e1eee3e379.", uutFilteredEntry.Text)
	}
}
func TestFilterEntryValidator(t *testing.T) {
	uutFilteredEntry.Text = ""
	test.Type(uutFilteredEntry, "+7.890e-10")
	if err := uutFilteredEntry.Validate(); err != nil {
		t.Errorf("TestFilterEntryValidator() failed. Filtered Entry did not validate valid input: Expected %s, got %s for %s", "", err.Error(), uutFilteredEntry.Text)
	}
	uutFilteredEntry.Text = ""
	test.Type(uutFilteredEntry, "+7.89.0e-1e0+-")
	if err := uutFilteredEntry.Validate(); err == nil {
		t.Errorf("TestFilterEntryFilter() failed. Filtered Entry did validate invalid input: Expected %t, got %t for %s", false, true, uutFilteredEntry.Text)
	}
}
