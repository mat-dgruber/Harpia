# Harpia CLI Fase A (DX, Formatação AST & Stream UI) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement a central UI/TTY helper module, an AST-aware code formatter (`harpia formatar`), unified git diff output (`--diff`), and CI check mode (`--verificar`).

**Architecture:** Create `cmd/ui.go` for terminal capability detection (`isatty`, `NO_COLOR`, `FORCE_COLOR`), update `cmd/formatar.go` with AST-driven block indentation, diff generation and exit status handling, and add comprehensive test coverage in `cmd/formatar_test.go`.

**Tech Stack:** Go 1.24+, `github.com/spf13/cobra`, standard Go `os`, `fmt`, `strings`, `path/filepath`.

## Global Constraints

- Must compile and pass all tests via `go test ./cmd/...`.
- No new external third-party dependencies.
- Terminal output colors must be disabled if `NO_COLOR` environment variable is set.

---

### Task 1: Terminal UI & Stream Helper (`cmd/ui.go`)

**Files:**
- Create: `cmd/ui.go`
- Test: `cmd/ui_test.go`

**Interfaces:**
- Produces: `UI` struct with `IsTerminal(w io.Writer) bool`, `ColorEnabled() bool`, `Verde(s string) string`, `Vermelho(s string) string`, `Amarelo(s string) string`.

- [ ] **Step 1: Write failing test for UI helper**

Create `cmd/ui_test.go`:
```go
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
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./cmd -run TestColorEnabledNoColor`
Expected: FAIL ("undefined: UI")

- [ ] **Step 3: Write minimal implementation of UI helper**

Create `cmd/ui.go`:
```go
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
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./cmd -run TestColorEnabledNoColor`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add -f cmd/ui.go cmd/ui_test.go
git commit -m "feat(cmd): add terminal UI helper with NO_COLOR support"
```

---

### Task 2: AST Code Formatter & Unified Diff (`cmd/formatar.go`)

**Files:**
- Modify: `cmd/formatar.go`
- Test: `cmd/formatar_test.go`

**Interfaces:**
- Consumes: `UI` from `cmd/ui.go`
- Produces: `FormatarCodigoHarpia(codigo string) string`, `GerarUnifiedDiff(caminho, orig, form string, ui *UI) string`, updated `comandoFormatar()` supporting `--diff`, `--verificar`, and `-w` (`--escrever`).

- [ ] **Step 1: Write failing test for GerarUnifiedDiff**

Add to `cmd/formatar_test.go`:
```go
package cmd

import (
	"strings"
	"testing"
)

func TestGerarUnifiedDiff(t *testing.T) {
	ui := &UI{}
	orig := "se x {\n  imprimir(1)\n}"
	form := "se x {\n    imprimir(1)\n}\n"

	diff := GerarUnifiedDiff("teste.hrp", orig, form, ui)
	if !strings.Contains(diff, "--- a/teste.hrp") || !strings.Contains(diff, "+++ b/teste.hrp") {
		t.Errorf("diff ausente de cabecalho esperado: %s", diff)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./cmd -run TestGerarUnifiedDiff`
Expected: FAIL ("undefined: GerarUnifiedDiff")

- [ ] **Step 3: Implement GerarUnifiedDiff and update comandoFormatar in cmd/formatar.go**

Update `cmd/formatar.go`:
```go
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func GerarUnifiedDiff(caminho, orig, form string, ui *UI) string {
	linhasOrig := strings.Split(orig, "\n")
	linhasForm := strings.Split(form, "\n")

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("--- a/%s\n", caminho))
	sb.WriteString(fmt.Sprintf("+++ b/%s\n", caminho))

	max := len(linhasOrig)
	if len(linhasForm) > max {
		max = len(linhasForm)
	}

	for i := 0; i < max; i++ {
		lO := ""
		lF := ""
		if i < len(linhasOrig) {
			lO = linhasOrig[i]
		}
		if i < len(linhasForm) {
			lF = linhasForm[i]
		}

		if lO != lF {
			if i < len(linhasOrig) {
				sb.WriteString(ui.Vermelho("- " + lO + "\n"))
			}
			if i < len(linhasForm) {
				sb.WriteString(ui.Verde("+ " + lF + "\n"))
			}
		} else {
			sb.WriteString("  " + lO + "\n")
		}
	}
	return sb.String()
}

func comandoFormatar() *cobra.Command {
	var escrever bool
	var verificar bool
	var diff bool

	ui := &UI{}

	cmdFormatar := &cobra.Command{
		Use:           "formatar [arquivo.hrp | diretorio]",
		Short:         "Formata a indentação e o estilo visual de arquivos Harpia",
		Args:          cobra.MinimumNArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			errOut := cmd.ErrOrStderr()

			var arquivos []string
			for _, arg := range args {
				if arg == "./..." || arg == "..." {
					arg = "."
				}
				info, err := os.Stat(arg)
				if err != nil {
					fmt.Fprintf(errOut, "Erro ao acessar caminho '%s': %v\n", arg, err)
					return err
				}
				if info.IsDir() {
					_ = filepath.Walk(arg, func(path string, f os.FileInfo, err error) error {
						if err == nil && !f.IsDir() && (strings.HasSuffix(path, ".hrp") || strings.HasSuffix(path, ".pt")) {
							arquivos = append(arquivos, path)
						}
						return nil
					})
				} else {
					arquivos = append(arquivos, arg)
				}
			}

			desformatados := 0
			for _, caminho := range arquivos {
				conteudo, err := os.ReadFile(caminho)
				if err != nil {
					fmt.Fprintf(errOut, "Erro ao ler arquivo %s: %v\n", caminho, err)
					continue
				}

				origStr := string(conteudo)
				formatado := FormatarCodigoHarpia(origStr)

				if origStr != formatado {
					desformatados++
					if diff {
						fmt.Fprint(out, GerarUnifiedDiff(caminho, origStr, formatado, ui))
					} else if verificar {
						fmt.Fprintf(errOut, "HRP-FMT-001: %s precisa de formatação.\n", caminho)
					}
				}

				if escrever && origStr != formatado {
					err = os.WriteFile(caminho, []byte(formatado), 0644)
					if err != nil {
						fmt.Fprintf(errOut, "Erro ao salvar arquivo %s: %v\n", caminho, err)
					} else {
						fmt.Fprintf(out, "Formatado: %s\n", caminho)
					}
				} else if !verificar && !diff && !escrever {
					fmt.Fprint(out, formatado)
				}
			}

			if verificar && desformatados > 0 {
				fmt.Fprintf(errOut, "\n%d arquivo(s) precisam de formatação. Execute 'harpia formatar -w [caminho]'.\n", desformatados)
				return fmt.Errorf("%d arquivo(s) desformatados", desformatados)
			}

			return nil
		},
	}

	cmdFormatar.Flags().BoolVarP(&escrever, "escrever", "w", false, "Salva as alterações de volta nos arquivos originais")
	cmdFormatar.Flags().BoolVar(&verificar, "verificar", false, "Sai com erro se algum arquivo não estiver formatado (para CI)")
	cmdFormatar.Flags().BoolVar(&diff, "diff", false, "Exibe as diferenças (diff) entre o arquivo atual e o formatado")
	return cmdFormatar
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./cmd -run TestGerarUnifiedDiff`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add -f cmd/formatar.go cmd/formatar_test.go
git commit -m "feat(cmd): add unified diff and CI check mode to harpia formatar"
```
