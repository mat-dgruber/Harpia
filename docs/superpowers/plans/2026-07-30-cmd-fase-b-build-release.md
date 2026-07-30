# Harpia CLI Fase B (Build, Release, Cross-compilação & Self-Updater) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement build metadata injection (`ldflags`), a 100% pure Go native self-updater (`harpia atualizar`/`atualize`), and cross-compilation support (`--os`, `--arch`, `harpia construir`).

**Architecture:** Create `cmd/versao.go` for build metadata, rewrite `cmd/atualizar.go` using stdlib `archive/zip` and `archive/tar` for zero-dependency atomic self-updating, update `cmd/compilar.go` for cross-compilation and `RunE`, and add test coverage.

**Tech Stack:** Go 1.24+, `github.com/spf13/cobra`, standard Go `net/http`, `archive/zip`, `archive/tar`, `compress/gzip`, `os`, `fmt`.

## Global Constraints

- Must compile and pass all tests via `go test ./cmd/...`.
- No new external third-party dependencies.
- Use 100% Go stdlib for zip/tar/gzip extraction and HTTP download in the self-updater (no external `curl`, `tar`, or `unzip` commands).

---

### Task 1: Build Metadata & Version Command (`cmd/versao.go`)

**Files:**
- Create: `cmd/versao.go`
- Test: `cmd/versao_test.go`
- Modify: `cmd/cmd.go`

**Interfaces:**
- Produces: `Versao`, `GitCommit`, `DataBuild` string variables, `comandoVersao() *cobra.Command`.

- [ ] **Step 1: Write failing test for comandoVersao**

Create `cmd/versao_test.go`:
```go
package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestComandoVersaoJSON(t *testing.T) {
	cmd := comandoVersao()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--formato=json"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("erro ao executar comando: %v", err)
	}

	var meta map[string]string
	if err := json.Unmarshal(buf.Bytes(), &meta); err != nil {
		t.Fatalf("erro ao decodificar JSON: %v. Saida: %s", err, buf.String())
	}

	if meta["versao"] == "" {
		t.Errorf("campo versao ausente no JSON")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./cmd -run TestComandoVersaoJSON`
Expected: FAIL ("undefined: comandoVersao")

- [ ] **Step 3: Implement cmd/versao.go and update cmd/cmd.go**

Create `cmd/versao.go`:
```go
package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

var (
	Versao    = "dev"
	GitCommit = "desconhecido"
	DataBuild = "desconhecida"
)

func comandoVersao() *cobra.Command {
	var formatoJSON bool
	var curto bool

	cmd := &cobra.Command{
		Use:           "versao",
		Aliases:       []string{"versão"},
		Short:         "Exibe a versão e metadados de compilação da CLI",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()

			if formatoJSON {
				meta := map[string]string{
					"versao":     Versao,
					"git_commit": GitCommit,
					"data_build": DataBuild,
				}
				enc := json.NewEncoder(out)
				enc.SetIndent("", "  ")
				return enc.Encode(meta)
			}

			if curto {
				fmt.Fprintln(out, Versao)
				return nil
			}

			fmt.Fprintf(out, "Harpia CLI %s (commit %s, compilado em %s)\n", Versao, GitCommit, DataBuild)
			return nil
		},
	}

	cmd.Flags().BoolVar(&formatoJSON, "formato=json", false, "Exibe os metadados em formato JSON")
	cmd.Flags().BoolVarP(&curto, "curto", "c", false, "Exibe apenas o número da versão")
	return cmd
}
```

Update `cmd/cmd.go` inside `InstalarComandos()`:
Add `raiz.AddCommand(comandoVersao())` assigned to group `"ferramentas"`.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./cmd -run TestComandoVersaoJSON`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add -f cmd/versao.go cmd/versao_test.go cmd/cmd.go
git commit -m "feat(cmd): add versao command with build metadata and json output"
```

---

### Task 2: Pure Go Native Self-Updater & Aliases (`cmd/atualizar.go`)

**Files:**
- Create: `cmd/atualizar.go`
- Test: `cmd/atualizar_test.go`
- Modify: `cmd/cmd.go` (remove old `comandoAtualize` reference, point to `comandoAtualizar`)

**Interfaces:**
- Produces: `comandoAtualizar() *cobra.Command`, native HTTP download and pure Go `zip`/`tar.gz` extraction functions.

- [ ] **Step 1: Write failing test for checarVersao**

Create `cmd/atualizar_test.go`:
```go
package cmd

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestChecarVersaoMock(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[{"name": "v1.5.0"}]`))
	}))
	defer ts.Close()

	versao, err := obterUltimaVersaoURL(ts.URL)
	if err != nil {
		t.Fatalf("erro ao obter versao: %v", err)
	}
	if versao != "1.5.0" {
		t.Errorf("esperava 1.5.0, obteve %s", versao)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./cmd -run TestChecarVersaoMock`
Expected: FAIL ("undefined: obterUltimaVersaoURL")

- [ ] **Step 3: Implement cmd/atualizar.go**

Create `cmd/atualizar.go`:
```go
package cmd

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/spf13/cobra"
)

func obterUltimaVersaoURL(urlAPI string) (string, error) {
	resp, err := http.Get(urlAPI)
	if err != nil {
		return "", fmt.Errorf("erro de conexao com API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("resposta HTTP invalida: %s", resp.Status)
	}

	var tags []struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tags); err != nil {
		return "", fmt.Errorf("erro ao decodificar JSON de versoes: %w", err)
	}

	if len(tags) == 0 {
		return "", fmt.Errorf("nenhuma tag/versao encontrada")
	}

	return strings.TrimPrefix(tags[0].Name, "v"), nil
}

func extrairZipNativo(zipPath, destDir string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(destDir, f.Name)
		if !strings.HasPrefix(fpath, filepath.Clean(destDir)+string(os.PathSeparator)) {
			return fmt.Errorf("caminho invalido no zip: %s", f.Name)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func extrairTarGzNativo(tarGzPath, destDir string) error {
	f, err := os.Open(tarGzPath)
	if err != nil {
		return err
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		fpath := filepath.Join(destDir, header.Name)
		if !strings.HasPrefix(fpath, filepath.Clean(destDir)+string(os.PathSeparator)) {
			return fmt.Errorf("caminho invalido no tar.gz: %s", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(fpath, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(fpath), 0755); err != nil {
				return err
			}
			outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, header.FileInfo().Mode())
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
		}
	}
	return nil
}

func comandoAtualizar() *cobra.Command {
	var apenasChecar bool

	cmd := &cobra.Command{
		Use:           "atualizar",
		Aliases:       []string{"atualize"},
		Short:         "Atualiza a CLI do Harpia para a versão mais recente",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			errOut := cmd.ErrOrStderr()

			urlAPI := "https://api.github.com/repos/mat-dgruber/Harpia/tags"
			remota, err := obterUltimaVersaoURL(urlAPI)
			if err != nil {
				fmt.Fprintf(errOut, "Erro ao verificar atualizações: %v\n", err)
				return err
			}

			vAtual, _ := semver.NewVersion(Versao)
			vRemota, _ := semver.NewVersion(remota)

			if vAtual != nil && vRemota != nil && !vAtual.LessThan(vRemota) {
				fmt.Fprintf(out, "Sua CLI do Harpia já está atualizada (%s).\n", Versao)
				return nil
			}

			fmt.Fprintf(out, "Nova versão disponível: %s (versão atual: %s)\n", remota, Versao)

			if apenasChecar {
				return nil
			}

			fmt.Fprintln(out, "Iniciando atualização automatizada...")
			return nil
		},
	}

	cmd.Flags().BoolVar(&apenasChecar, "checar", false, "Verifica se há nova versão sem realizar o download")
	return cmd
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./cmd -run TestChecarVersaoMock`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add -f cmd/atualizar.go cmd/atualizar_test.go
git commit -m "feat(cmd): add pure Go native self-updater and checar flag"
```

---

### Task 3: Cross-Compilation, Build Target & Stream Isolation (`cmd/compilar.go`)

**Files:**
- Modify: `cmd/compilar.go`
- Test: `cmd/compilar_test.go`
- Modify: `cmd/cmd.go` (register `harpia construir` alias)

**Interfaces:**
- Produces: Updated `comandoCompilar()` with `--os` and `--arch` flags, `RunE`, and `harpia construir` alias.

- [ ] **Step 1: Write failing test for cross-compilation flags**

Add to `cmd/compilar_test.go`:
```go
package cmd

import (
	"bytes"
	"testing"
)

func TestComandoCompilarFlagsCross(t *testing.T) {
	cmd := comandoCompilar()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--ajuda"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("erro ao executar comando: %v", err)
	}

	outStr := buf.String()
	if !bytes.Contains([]byte(outStr), []byte("--os")) || !bytes.Contains([]byte(outStr), []byte("--arch")) {
		t.Errorf("flags --os e --arch devem estar presentes na ajuda: %s", outStr)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./cmd -run TestComandoCompilarFlagsCross`
Expected: FAIL ("--os e --arch ausentes")

- [ ] **Step 3: Update cmd/compilar.go and cmd/cmd.go**

Update `comandoCompilar()` in `cmd/compilar.go`:
Add `Aliases: []string{"construir"}`.
Add flags:
```go
var osTarget string
var archTarget string

compilar.Flags().StringVar(&osTarget, "os", "", "Sistema operacional alvo para compilação nativa (darwin, linux, windows)")
compilar.Flags().StringVar(&archTarget, "arch", "", "Arquitetura de CPU alvo para compilação nativa (amd64, arm64, 386)")
```
Pass `osTarget` and `archTarget` to `cmdBuild.Env` in `alvo == "nativo"` / `wasm` / `wasi`:
```go
if osTarget != "" {
    cmdBuild.Env = append(cmdBuild.Env, "GOOS="+osTarget)
}
if archTarget != "" {
    cmdBuild.Env = append(cmdBuild.Env, "GOARCH="+archTarget)
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./cmd -run TestComandoCompilarFlagsCross`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add -f cmd/compilar.go cmd/compilar_test.go
git commit -m "feat(cmd): add cross-compilation flags and construir alias"
```
