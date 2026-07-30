package cmd

import (
	"io"
	"os"
)

type UI struct{}

func (u *UI) IsTerminal(w io.Writer) bool {
	if f, ok := w.(*os.File); ok {
		stat, err := f.Stat()
		if err == nil {
			return (stat.Mode() & os.ModeCharDevice) != 0
		}
	}
	return false
}

func (u *UI) ColorEnabled() bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	if os.Getenv("FORCE_COLOR") != "" {
		return true
	}
	return u.IsTerminal(os.Stdout)
}

func (u *UI) Verde(s string) string {
	if !u.ColorEnabled() {
		return s
	}
	return "\033[32m" + s + "\033[0m"
}

func (u *UI) Vermelho(s string) string {
	if !u.ColorEnabled() {
		return s
	}
	return "\033[31m" + s + "\033[0m"
}

func (u *UI) Amarelo(s string) string {
	if !u.ColorEnabled() {
		return s
	}
	return "\033[33m" + s + "\033[0m"
}
