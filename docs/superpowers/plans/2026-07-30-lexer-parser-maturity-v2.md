# Lexer & Parser Maturity V2 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Elevate Harpia's lexer and parser to production maturity through ASCII lookup tables, hex/bin/octal literals, raw strings, rich IDE diagnostics, structural pattern matching (`escolha`), optional chaining (`?.()`, `?.[idx]`), and AST docstrings.

**Architecture:** Extended lexer state machine with $O(1)$ lookup table, enriched `ErroSintatico` with code snippet indicators, pattern matching AST nodes, optional chaining suffix parsers, and comment-to-node docstring association.

**Tech Stack:** Go (stdlib `testing`, `testing.F`, `testing.B`)

## Global Constraints

- Go stdlib only, no external third-party dependencies.
- Zero extra heap allocations during token scanning and position tracking.
- All existing tests across the repository (`go test ./...`) must pass without regressions.

---

### Task 1: Lexer ASCII Lookup Table, Alternative Bases, Visual Separators & Raw Strings

**Files:**
- Modify: `Harpia/lexer/tokens.go`
- Modify: `Harpia/lexer/helpers.go`
- Modify: `Harpia/lexer/lexer.go`
- Modify: `Harpia/tests/lexer_test/lexer_test.go`

**Interfaces:**
- Produces: Hexadecimal (`0xFF`), binary (`0b1010`), octal (`0o755`) token parsing.
- Produces: Digit separator `_` support in numbers (`1_000_000`).
- Produces: Raw string token parsing (`` `texto` ``).
- Produces: $O(1)$ ASCII lookup table in `lexer.go`.

- [ ] **Step 1: Write failing tests for lexer features**

In `Harpia/tests/lexer_test/lexer_test.go`:
```go
func TestLexerBasesENumerosSeparados(t *testing.T) {
	codigo := "0xFF 0b1010 0o755 1_000_000 `string bruta\ncom linha`"
	l := lexer.NovoLexer(codigo, "teste.hrp")
	tokens := l.ObterTokens()
	
	if len(tokens) < 5 {
		t.Fatalf("Esperava pelo menos 5 tokens, obteve %d", len(tokens))
	}
	if tokens[0].Valor != "0xFF" || tokens[1].Valor != "0b1010" || tokens[2].Valor != "0o755" {
		t.Errorf("Falha em literais de base: %+v", tokens[:3])
	}
	if tokens[3].Valor != "1_000_000" {
		t.Errorf("Falha em separador numérico: %s", tokens[3].Valor)
	}
	if tokens[4].Tipo != lexer.TokenTexto || !strings.Contains(tokens[4].Valor, "string bruta") {
		t.Errorf("Falha em raw string: %+v", tokens[4])
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test -v -run TestLexerBasesENumerosSeparados ./tests/lexer_test/...`
Expected: FAIL

- [ ] **Step 3: Implement ASCII lookup table and scanner additions in lexer.go**

Update `Harpia/lexer/lexer.go`:
1. Define ASCII lookup table `[256]byte` with bit flags `flagDigito`, `flagLetra`, `flagEspaco`, `flagHex`.
2. Update `lerNumero()` to support `0x`/`0X`, `0b`/`0B`, `0o`/`0O` and skips `_`.
3. Add `lerStringBruta()` triggered by backtick (`` ` ``).

- [ ] **Step 4: Run test to verify it passes**

Run: `go test -v -run TestLexerBasesENumerosSeparados ./tests/lexer_test/...`
Expected: PASS

- [ ] **Step 5: Commit changes**

```bash
git add lexer/tokens.go lexer/helpers.go lexer/lexer.go tests/lexer_test/lexer_test.go
git commit -m "feat(lexer): add ASCII lookup table, hex/bin/octal literals, numeric separators and raw strings"
```

---

### Task 2: Rich Error Diagnostics (`ErroSintatico`)

**Files:**
- Modify: `Harpia/parser/parser.go`
- Modify: `Harpia/tests/parser_test/parser_test.go`

**Interfaces:**
- Produces: `ErroSintatico` with `Codigo`, `LinhaFonte`, `Indicador`, and `Dica`.
- Produces: `Parser.formatarErro(codigo, msg, dica)`.

- [ ] **Step 1: Write failing test for rich error diagnostics**

In `Harpia/tests/parser_test/parser_test.go`:
```go
func TestDiagnosticosRicos(t *testing.T) {
	codigo := "var x = 10 +"
	p := NewParserFromString(codigo, "teste_diag.hrp")
	_, err := p.Parse()
	if err == nil {
		t.Fatalf("Esperava erro sintático")
	}
	erros := p.Erros()
	if len(erros) == 0 {
		t.Fatalf("Nenhum ErroSintatico acumulado")
	}
	e := erros[0]
	if e.Codigo == "" || e.LinhaFonte == "" || e.Indicador == "" {
		t.Errorf("ErroSintatico sem campos de diagnóstico ricos: %+v", e)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test -v -run TestDiagnosticosRicos ./tests/parser_test/...`
Expected: FAIL

- [ ] **Step 3: Update ErroSintatico and error construction in parser.go**

In `Harpia/parser/parser.go`:
Expand `ErroSintatico` struct fields (`Codigo`, `LinhaFonte`, `Indicador`, `Dica`).
Update `errof()` to extract source line snippet from input, generate caret string `^^^^`, and assign diagnostic error codes (e.g. `HRP-E101`).

- [ ] **Step 4: Run test to verify it passes**

Run: `go test -v -run TestDiagnosticosRicos ./tests/parser_test/...`
Expected: PASS

- [ ] **Step 5: Commit changes**

```bash
git add parser/parser.go tests/parser_test/parser_test.go
git commit -m "feat(parser): add IDE-grade rich error diagnostics with source snippets and tips"
```

---

### Task 3: Complete Optional Chaining (`?.()`, `?.[idx]`)

**Files:**
- Modify: `Harpia/parser/ast_nodes.go`
- Modify: `Harpia/parser/parser.go`
- Modify: `Harpia/tests/parser_test/parser_test.go`

**Interfaces:**
- Produces: `ChamadaOpcional` and `IndexacaoOpcional` AST node structs.
- Produces: Suffix parsing for `?.(` and `?.[`.

- [ ] **Step 1: Write failing test for optional chaining**

In `Harpia/tests/parser_test/parser_test.go`:
```go
func TestEncadeamentoOpcionalCompleto(t *testing.T) {
	codigo := `
		var res1 = fn?.(1, 2)
		var res2 = lista?.[0]
	`
	p := NewParserFromString(codigo, "teste_opt.hrp")
	ast, err := p.Parse()
	if err != nil {
		t.Fatalf("Erro sintático: %v", err)
	}
	if len(ast.Declaracoes) != 2 {
		t.Fatalf("Esperava 2 declarações, obteve %d", len(ast.Declaracoes))
	}
	decl1 := ast.Declaracoes[0].(*DeclVar)
	if _, ok := decl1.Inicializador.(*ChamadaOpcional); !ok {
		t.Errorf("Esperava *ChamadaOpcional, obteve %T", decl1.Inicializador)
	}
	decl2 := ast.Declaracoes[1].(*DeclVar)
	if _, ok := decl2.Inicializador.(*IndexacaoOpcional); !ok {
		t.Errorf("Esperava *IndexacaoOpcional, obteve %T", decl2.Inicializador)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test -v -run TestEncadeamentoOpcionalCompleto ./tests/parser_test/...`
Expected: FAIL

- [ ] **Step 3: Define AST nodes & parser logic in ast_nodes.go and parser.go**

In `Harpia/parser/ast_nodes.go`:
Define `ChamadaOpcional` and `IndexacaoOpcional` (embed `NoBase`).

In `Harpia/parser/parser.go`:
In `parseExpressaoSufixo()`, when encountering `TokenInterrogacaoPonto` (`?.`), inspect next token:
- If `(`, parse as `ChamadaOpcional`.
- If `[`, parse as `IndexacaoOpcional`.
- Otherwise parse as existing `AcessoMembroOpcional`.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test -v -run TestEncadeamentoOpcionalCompleto ./tests/parser_test/...`
Expected: PASS

- [ ] **Step 5: Commit changes**

```bash
git add parser/ast_nodes.go parser/parser.go tests/parser_test/parser_test.go
git commit -m "feat(parser): add complete optional chaining support for calls ?.() and indexing ?.[idx]"
```

---

### Task 4: Structural Pattern Matching (`escolha`)

**Files:**
- Modify: `Harpia/lexer/tokens.go`
- Modify: `Harpia/lexer/helpers.go`
- Modify: `Harpia/parser/ast_nodes.go`
- Modify: `Harpia/parser/parser.go`
- Modify: `Harpia/tests/parser_test/parser_test.go`

**Interfaces:**
- Produces: `TokenEscolha`, `TokenCaso`, `TokenQuando` tokens.
- Produces: `CasoEscolha` and `DeclEscolha` AST node structs.
- Produces: `parseEscolha()` in `parser.go`.

- [ ] **Step 1: Write failing test for structural pattern matching**

In `Harpia/tests/parser_test/parser_test.go`:
```go
func TestPatternMatchingEscolha(t *testing.T) {
	codigo := `
		escolha resposta {
			caso [a, b, ...resto] quando a > 0 {
				imprimir(a)
			}
			caso { status: 200 } {
				imprimir("ok")
			}
			caso _ {
				imprimir("padrao")
			}
		}
	`
	p := NewParserFromString(codigo, "teste_pm.hrp")
	ast, err := p.Parse()
	if err != nil {
		t.Fatalf("Erro sintático em escolha: %v", err)
	}
	if len(ast.Declaracoes) != 1 {
		t.Fatalf("Esperava 1 declaração, obteve %d", len(ast.Declaracoes))
	}
	esc, ok := ast.Declaracoes[0].(*DeclEscolha)
	if !ok || len(esc.Casos) != 3 {
		t.Errorf("Falha no parse de DeclEscolha: %+v", ast.Declaracoes[0])
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test -v -run TestPatternMatchingEscolha ./tests/parser_test/...`
Expected: FAIL

- [ ] **Step 3: Add tokens and parseEscolha logic**

In `Harpia/lexer/tokens.go` & `helpers.go`:
Add keywords `escolha`, `caso`, `quando`.

In `Harpia/parser/ast_nodes.go`:
Define `CasoEscolha` and `DeclEscolha` (embed `NoBase`).

In `Harpia/parser/parser.go`:
Implement `parseEscolha()`, parsing expression, cases, patterns (literals, destructuring, wildcard `_`), optional `quando` guard, and body blocks.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test -v -run TestPatternMatchingEscolha ./tests/parser_test/...`
Expected: PASS

- [ ] **Step 5: Commit changes**

```bash
git add lexer/tokens.go lexer/helpers.go parser/ast_nodes.go parser/parser.go tests/parser_test/parser_test.go
git commit -m "feat(parser): add structural pattern matching escolha syntax parsing"
```

---

### Task 5: AST Docstrings Preservation & Performance Benchmarks

**Files:**
- Modify: `Harpia/parser/ast_nodes.go`
- Modify: `Harpia/parser/parser.go`
- Create: `Harpia/tests/parser_test/parser_bench_test.go`
- Modify: `Harpia/tests/parser_test/parser_test.go`

**Interfaces:**
- Produces: `DocComments string` field on declaration AST nodes (`DeclFuncao`, `DeclClasse`, `DeclVar`).
- Produces: `BenchmarkLexer` and `BenchmarkParser` suites.

- [ ] **Step 1: Write failing test for docstrings and benchmark file**

In `Harpia/tests/parser_test/parser_test.go`:
```go
func TestASTDocstrings(t *testing.T) {
	codigo := `
		/// Documentação da função soma
		funcao soma(a, b) {
			retorne a + b
		}
	`
	p := NewParserFromString(codigo, "teste_doc.hrp")
	ast, err := p.Parse()
	if err != nil {
		t.Fatalf("Erro sintático: %v", err)
	}
	if len(ast.Declaracoes) != 1 {
		t.Fatalf("Esperava 1 declaração")
	}
	fn, ok := ast.Declaracoes[0].(*DeclFuncao)
	if !ok || !strings.Contains(fn.DocComments, "Documentação da função soma") {
		t.Errorf("Falha ao capturar DocComments na AST: %+v", fn)
	}
}
```

Create `Harpia/tests/parser_test/parser_bench_test.go`:
```go
package parser_test

import (
	"testing"

	"github.com/mat-dgruber/Harpia/lexer"
	"github.com/mat-dgruber/Harpia/parser"
)

func BenchmarkLexer(b *testing.B) {
	codigo := "var [a, b, ...resto] = [1, 2, 3]; escolha x { caso 1 { retorne 10 } }"
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		l := lexer.NovoLexer(codigo, "bench.hrp")
		_ = l.ObterTokens()
	}
}

func BenchmarkParser(b *testing.B) {
	codigo := "var [a, b, ...resto] = [1, 2, 3]; escolha x { caso 1 { retorne 10 } }"
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		p := parser.NewParserFromString(codigo, "bench.hrp")
		_, _ = p.Parse()
	}
}
```

- [ ] **Step 2: Run test to verify failure**

Run: `go test -v -run TestASTDocstrings ./tests/parser_test/...`
Expected: FAIL

- [ ] **Step 3: Implement DocComments attachment in parser.go**

In `Harpia/parser/ast_nodes.go`: Add `DocComments string` to `DeclFuncao`, `DeclClasse`, `DeclVar`, `DeclEnum`, `DeclInterface`.
In `Harpia/parser/parser.go`: Collect preceding `///` or `/** */` comments during statement parsing and assign to declaration node.

- [ ] **Step 4: Run test and benchmarks to verify pass**

Run: `go test -v -run TestASTDocstrings ./tests/parser_test/...`
Expected: PASS

Run: `go test -bench=. -benchmem ./tests/parser_test/...`
Expected: Benchmark outputs performance metrics and 0 unnecessary allocations.

- [ ] **Step 5: Commit changes**

```bash
git add parser/ast_nodes.go parser/parser.go tests/parser_test/parser_bench_test.go tests/parser_test/parser_test.go
git commit -m "feat(parser): add AST docstring preservation and lexer/parser performance benchmarks"
```
