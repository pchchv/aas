package locales

import "testing"

func TestUniqueIds(t *testing.T) {
	result := Get()
	idMap := make(map[string]bool)
	for _, locale := range result {
		if idMap[locale.Id] {
			t.Errorf("Duplicate Id found: %s", locale.Id)
		}
		idMap[locale.Id] = true
	}
}

func TestNonEmptyFields(t *testing.T) {
	result := Get()
	for _, locale := range result {
		if locale.Id == "" {
			t.Error("Found a Locale with empty Id")
		}
		if locale.Value == "" {
			t.Error("Found a Locale with empty Value")
		}
	}
}
