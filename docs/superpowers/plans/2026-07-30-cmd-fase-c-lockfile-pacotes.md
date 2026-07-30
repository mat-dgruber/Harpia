# Harpia CLI Fase C (Lockfile & Gerenciamento de Pacotes) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement deterministic lockfile support (`harpia.lock`/`harpia.lock.json`), SHA-256 package checksum verification, and `harpia pacote` lifecycle subcommands (`adicionar`, `remover`, `listar`, `limpar`).

**Architecture:** Create `cmd/lockfile.go` for `Lockfile` data structure and dual JSON/HRP parser/writer, update `cmd/pacotes.go` with SHA-256 integrity calculation during package download/extraction, and implement `comandoPacote()` with subcommands registered in Cobra.

**Tech Stack:** Go 1.24+, `github.com/spf13/cobra`, stdlib `crypto/sha256`, `encoding/json`, `archive/zip`, `os`, `fmt`, `path/filepath`.

## Global Constraints

- Must compile and pass all tests via `go test ./cmd/...`.
- No new external third-party dependencies.
- Use Go stdlib `crypto/sha256` for integrity hashes.
- Lockfile verification failure must return error `HRP-LOCK-001`.

---

### Task 1: Lockfile Data Structure & Parser (`cmd/lockfile.go`)

**Files:**
- Create: `cmd/lockfile.go`
- Create: `cmd/lockfile_test.go`

**Interfaces:**
- Produces: `Lockfile` struct, `PacoteLockInfo` struct, `parseLockfile(conteudo []byte) (*Lockfile, error)`, `salvarLockfile(caminho string, lock *Lockfile) error`.

- [ ] **Step 1: Write failing test for parseLockfile**

Create `cmd/lockfile_test.go`:
```go
package cmd

import (
	"testing"
)

func TestParseLockfileJSON(t *testing.T) {
	jsonStr := `{
  "versao_lock": "1.0",
  "pacotes": {
    "matematica": {
      "versao": "1.2.0",
      "url": "https://example.com/matematica.zip",
      "sha256": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
    }
  }
}`
	lock, err := parseLockfile([]byte(jsonStr))
	if err != nil {
		t.Fatalf("erro ao analisar lockfile JSON: %v", err)
	}
	if lock.VersaoLock != "1.0" {
		t.Errorf("esperava versao_lock 1.0, obteve %s", lock.VersaoLock)
	}
	p, existe := lock.Pacotes["matematica"]
	if !existe || p.Versao != "1.2.0" {
		t.Errorf("pacote matematica ausente ou com versao incorreta: %+v", p)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./cmd -run TestParseLockfileJSON`
Expected: FAIL ("undefined: parseLockfile")

- [ ] **Step 3: Implement cmd/lockfile.go**

Create `cmd/lockfile.go`:
```go
//go:build !js && !wasm

package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type PacoteLockInfo struct {
	Versao string `json:"versao"`
	URL    string `json:"url"`
	SHA256 string `json:"sha256"`
}

type Lockfile struct {
	VersaoLock string                    `json:"versao_lock"`
	Pacotes    map[string]PacoteLockInfo `json:"pacotes"`
}

func parseLockfile(conteudo []byte) (*Lockfile, error) {
	var lock Lockfile
	if err := json.Unmarshal(conteudo, &lock); err == nil {
		if lock.Pacotes == nil {
			lock.Pacotes = make(map[string]PacoteLockInfo)
		}
		return &lock, nil
	}

	// Fallback para parse linha-a-linha de formato HRP key-value
	lock.VersaoLock = "1.0"
	lock.Pacotes = make(map[string]PacoteLockInfo)

	linhas := strings.Split(string(conteudo), "\n")
	var pacoteAtual string
	var infoAtual PacoteLockInfo

	for _, linha := range linhas {
		linha = strings.TrimSpace(linha)
		if linha == "" || strings.HasPrefix(linha, "#") || strings.HasPrefix(linha, "//") {
			continue
		}

		if strings.HasPrefix(linha, "[") && strings.HasSuffix(linha, "]") {
			if pacoteAtual != "" {
				lock.Pacotes[pacoteAtual] = infoAtual
			}
			pacoteAtual = strings.Trim(linha, "[] ")
			infoAtual = PacoteLockInfo{}
			continue
		}

		if strings.Contains(linha, "=") {
			partes := strings.SplitN(linha, "=", 2)
			k := strings.TrimSpace(partes[0])
			v := strings.Trim(strings.TrimSpace(partes[1]), "\"'")

			switch k {
			case "versao_lock":
				lock.VersaoLock = v
			case "versao":
				infoAtual.Versao = v
			case "url":
				infoAtual.URL = v
			case "sha256":
				infoAtual.SHA256 = v
			}
		}
	}
	if pacoteAtual != "" {
		lock.Pacotes[pacoteAtual] = infoAtual
	}

	return &lock, nil
}

func salvarLockfile(caminho string, lock *Lockfile) error {
	if lock.VersaoLock == "" {
		lock.VersaoLock = "1.0"
	}
	if lock.Pacotes == nil {
		lock.Pacotes = make(map[string]PacoteLockInfo)
	}

	bytes, err := json.MarshalIndent(lock, "", "  ")
	if err != nil {
		return fmt.Errorf("erro ao serializar lockfile: %w", err)
	}

	return os.WriteFile(caminho, bytes, 0644)
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./cmd -run TestParseLockfileJSON`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add -f cmd/lockfile.go cmd/lockfile_test.go
git commit -m "feat(cmd): add Lockfile struct and parser with dual JSON and HRP support"
```

---

### Task 2: SHA-256 Checksum Integrity & Deterministic Install (`cmd/pacotes.go`)

**Files:**
- Modify: `cmd/pacotes.go`
- Test: `cmd/pacotes_test.go`

**Interfaces:**
- Consumes: `Lockfile`, `parseLockfile`, `salvarLockfile` from `cmd/lockfile.go`.
- Produces: `calcularSHA256(dados []byte) string`, updated `baixarEExtrairPacote` with SHA-256 verification and lockfile integration.

- [ ] **Step 1: Write failing test for SHA-256 calculation and integrity error**

Add to `cmd/pacotes_test.go`:
```go
package cmd

import (
	"testing"
)

func TestCalcularSHA256(t *testing.T) {
	dados := []byte("harpia-pacote-teste")
	hash := calcularSHA256(dados)
	if hash == "" || len(hash) != 64 {
		t.Errorf("hash SHA-256 invalido: %s", hash)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./cmd -run TestCalcularSHA256`
Expected: FAIL ("undefined: calcularSHA256")

- [ ] **Step 3: Update cmd/pacotes.go**

In `cmd/pacotes.go`:
1. Add `calcularSHA256(dados []byte) string` using `crypto/sha256`.
2. Update `baixarEExtrairPacote(nome, url string, sha256Esperado string) (string, error)` to calculate hash, check against `sha256Esperado` if non-empty, and return hash or error `HRP-LOCK-001`.
3. Update `comandoInstalar` to check for `harpia.lock` / `harpia.lock.json` and enforce deterministic version/hash installation.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./cmd -run TestCalcularSHA256`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add -f cmd/pacotes.go cmd/pacotes_test.go
git commit -m "feat(cmd): add SHA-256 package integrity validation and lockfile checking"
```

---

### Task 3: Package Lifecycle Subcommands (`harpia pacote`)

**Files:**
- Modify: `cmd/pacotes.go`
- Test: `cmd/pacotes_test.go`
- Modify: `cmd/cmd.go`

**Interfaces:**
- Produces: `comandoPacote() *cobra.Command` with subcommands `adicionar`, `remover`, `listar`, `limpar`.

- [ ] **Step 1: Write failing test for comandoPacote Listar**

Add to `cmd/pacotes_test.go`:
```go
package cmd

import (
	"bytes"
	"testing"
)

func TestComandoPacoteListarVazio(t *testing.T) {
	cmd := comandoPacote()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"listar"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("erro ao executar pacote listar: %v", err)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./cmd -run TestComandoPacoteListarVazio`
Expected: FAIL ("undefined: comandoPacote")

- [ ] **Step 3: Implement comandoPacote and register in cmd/cmd.go**

In `cmd/pacotes.go`:
Create `comandoPacote() *cobra.Command` with subcommands:
- `adicionar` (`add`)
- `remover` (`remove`/`rm`)
- `listar` (`list`/`ls`)
- `limpar` (`clean`)

In `cmd/cmd.go`:
Register `raiz.AddCommand(comandoPacote())` under group `"ferramentas"`.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./cmd -run TestComandoPacoteListarVazio`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add -f cmd/pacotes.go cmd/pacotes_test.go cmd/cmd.go
git commit -m "feat(cmd): add harpia pacote lifecycle subcommands (adicionar, remover, listar, limpar)"
```
