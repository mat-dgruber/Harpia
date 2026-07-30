package cmd

import (
	"os"
	"testing"
)

func TestColorEnabledNoColor(t *testing.T) {
	os.Setenv("NO_COLOR", "1")
	defer os.Unsetenv("NO_COLOR")

	ui := &UI{}
	if ui.ColorEnabled() {
		t.Errorf("esperava ColorEnabled = false quando NO_COLOR esta ativo")
	}
}
