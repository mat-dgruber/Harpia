package cmd

import (
	"bytes"
	"strings"
	"testing"
)

// A release injeta cmd.Version/Commit/Datetime via ldflags (main.go); `harpia versao` precisa
// mostrar esses valores, e não placeholders.
func TestVersaoUsaMetadadosDeBuild(t *testing.T) {
	antV, antC, antD := Version, Commit, Datetime
	Version, Commit, Datetime = "1.2.3", "abc1234", "2026-11-04T00:00:00Z"
	t.Cleanup(func() { Version, Commit, Datetime = antV, antC, antD })

	var out bytes.Buffer
	c := comandoVersao()
	c.SetOut(&out)
	c.SetArgs([]string{})
	if err := c.Execute(); err != nil {
		t.Fatal(err)
	}
	for _, esperado := range []string{"1.2.3", "abc1234", "2026-11-04T00:00:00Z"} {
		if !strings.Contains(out.String(), esperado) {
			t.Fatalf("saída %q não contém %q", out.String(), esperado)
		}
	}
}
