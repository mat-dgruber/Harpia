# Parser Maturity Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Elevate `Harpia/parser` to production-grade maturity with zero-alloc AST node coordinate tracking, Panic Mode error recovery, logical/coalescing assignment operators, collection destructuring, and Go Fuzz testing.

**Architecture:** Update `BaseNode` interface to mandate `Pos() (lexer.PosicaoToken, lexer.PosicaoToken)` with `Inicio`/`Fim` fields on all AST structs. Add `erros []ErroSintatico` and `sincronizar()` method for statement boundary resynchronization. Implement destructuring and logical assignment AST nodes and parser routines.

**Tech Stack:** Go 1.24+ standard library (`testing`, `testing/fuzz`), `Harpia/lexer`, `Harpia/parser`.

## Global Constraints

- Must maintain 100% backward compatibility with existing AST node consumers (`vm/compilador.go`, `cmd/checar.go`).
- All tests in `go test ./...` must pass cleanly without regressions.
- All coordinate positions (`lexer.PosicaoToken`) must be stored by value to avoid heap allocations.

---

### Task 1: Zero-Alloc AST Node Coordinates & BaseNode Interface Refactoring

**Files:**
- Modify: `Harpia/parser/ast_nodes.go`
- Modify: `Harpia/parser/parser.go`
- Modify: `Harpia/tests/parser_test/parser_test.go`

**Interfaces:**
- Produces: `BaseNode` interface with `Pos() (lexer.PosicaoToken, lexer.PosicaoToken)`.
- Produces: `Inicio` and `Fim` fields (`lexer.PosicaoToken`) on all AST node structs.

- [ ] **Step 1: Write failing test for AST node coordinates**

```go
func TestNoPosicaoCoordenadas(t *testing.T) {
	p := NewParserFromString("var x = 42", "teste.hrp")
	ast, err := p.Parse()
	if err != nil {
		t.Fatalf("Erro inesperado: %v", err)
	}
	if len(ast.Instrucoes) == 0 {
		t.Fatalf("AST sem instruções")
	}
	decl := ast.Instrucoes[0]
	inicio, fim := decl.Pos()
	if inicio.Linha != 1 || inicio.Coluna != 1 {
		t.Errorf("Posição inicial incorreta: %+v", inicio)
	}
	if fim.Indice <= inicio.Indice {
		t.Errorf("Posição final deve ser maior que inicial: %d vs %d", fim.Indice, inicio.Indice)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test -v -run TestNoPosicaoCoordenadas ./tests/parser_test/...`
Expected: FAIL due to missing `Pos()` method on `BaseNode`.

- [ ] **Step 3: Update BaseNode and AST node structs in ast_nodes.go & parser.go**

In `Harpia/parser/ast_nodes.go`:
Add `Pos() (lexer.PosicaoToken, lexer.PosicaoToken)` to `BaseNode`.
Add `Inicio lexer.PosicaoToken` and `Fim lexer.PosicaoToken` to all node structs (`Programa`, `Bloco`, `DeclVar`, `DeclFuncao`, `Reatribuicao`, `OpBinaria`, `OpUnaria`, `InteiroLiteral`, `DecimalLiteral`, `TextoLiteral`, `LogicoLiteral`, `NuloLiteral`, `Identificador`, `Chamada`, `AcessoPropriedade`, `AcessoIndice`, `ListaLiteral`, `TuplaLiteral`, `MapaLiteral`, `Retorne`, `Se`, `Enquanto`, `Para`, `TenteCaptureFinalmente`).

In `Harpia/parser/parser.go`:
Update `registrar(node BaseNode, tok *lexer.Token)` to set `n.Inicio` and `n.Fim` when `tok` is non-nil.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test -v -run TestNoPosicaoCoordenadas ./tests/parser_test/...`
Expected: PASS

- [ ] **Step 5: Commit changes**

```bash
git add parser/ast_nodes.go parser/parser.go tests/parser_test/parser_test.go
git commit -m "refactor(parser): add zero-alloc Pos() coordinates to AST nodes"
```

---

### Task 2: Error Recovery & Synchronization (Panic Mode)

**Files:**
- Modify: `Harpia/parser/parser.go`
- Modify: `Harpia/tests/parser_test/parser_test.go`

**Interfaces:**
- Produces: `ErroSintatico` struct.
- Produces: `Parser.Erros() []ErroSintatico`.
- Produces: `p.sincronizar()` method for statement boundary resynchronization.

- [ ] **Step 1: Write failing test for multiple syntax errors**

```go
func TestRecuperacaoErrosSintaticos(t *testing.T) {
	codigo := `
		var a = 1 + ;
		var b = 2;
		se { ;
		var c = 3;
	`
	p := NewParserFromString(codigo, "teste_erros.hrp")
	_, err := p.Parse()
	if err == nil {
		t.Fatalf("Esperava erros de compilação")
	}
	erros := p.Erros()
	if len(erros) < 2 {
		t.Errorf("Esperava pelo menos 2 erros acumulados, obteve %d", len(erros))
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test -v -run TestRecuperacaoErrosSintaticos ./tests/parser_test/...`
Expected: FAIL (only 1 error returned, no error recovery).

- [ ] **Step 3: Implement ErroSintatico, erros slice and sincronizar()**

In `Harpia/parser/parser.go`:
```go
type ErroSintatico struct {
	Mensagem string
	Inicio   lexer.PosicaoToken
	Fim      lexer.PosicaoToken
}

func (p *Parser) Erros() []ErroSintatico {
	return p.erros
}

func (p *Parser) sincronizar() {
	p.avancar()
	for !p.noFim() {
		if p.atual.Tipo == lexer.TokenPontoEVirgula || p.atual.Tipo == lexer.TokenNovaLinha {
			p.avancar()
			return
		}
		switch p.atual.Tipo {
		case lexer.TokenVar, lexer.TokenConst, lexer.TokenFuncao, lexer.TokenSe,
			lexer.TokenEnquanto, lexer.TokenPara, lexer.TokenRetorne, lexer.TokenTente:
			return
		}
		p.avancar()
	}
}
```

In `parseDeclaracao()`: when `erroSintatico()` occurs, record `ErroSintatico`, call `p.sincronizar()` and return `nil`.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test -v -run TestRecuperacaoErrosSintaticos ./tests/parser_test/...`
Expected: PASS

- [ ] **Step 5: Commit changes**

```bash
git add parser/parser.go tests/parser_test/parser_test.go
git commit -m "feat(parser): implement Panic Mode error recovery and multi-error collection"
```

---

### Task 3: Logical & Coalescing Assignment Operators (`??=`, `||=`, `&&=`)

**Files:**
- Modify: `Harpia/parser/parser.go`
- Modify: `Harpia/tests/parser_test/parser_test.go`

**Interfaces:**
- Produces: Support for `??=`, `||=`, `&&=` assignment operators in `parseReatribuicao()`.

- [ ] **Step 1: Write failing test for logical assignment operators**

```go
func TestAtribucaoLogicaECoalescente(t *testing.T) {
	codigo := `
		var x = Nulo
		x ??= 10
		var y = Falso
		y ||= Verdadeiro
		var z = Verdadeiro
		z &&= Falso
	`
	p := NewParserFromString(codigo, "teste_op.hrp")
	ast, err := p.Parse()
	if err != nil {
		t.Fatalf("Erro sintático: %v", err)
	}
	if len(ast.Instrucoes) != 6 {
		t.Fatalf("Esperava 6 instruções, obteve %d", len(ast.Instrucoes))
	}
	reatr, ok := ast.Instrucoes[1].(*Reatribuicao)
	if !ok || reatr.Operador != "??=" {
		t.Errorf("Esperava reatribuição com '??=', obteve %+v", ast.Instrucoes[1])
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test -v -run TestAtribucaoLogicaECoalescente ./tests/parser_test/...`
Expected: FAIL with syntax error on `??=`.

- [ ] **Step 3: Implement logical/coalescing assignment parsing in parser.go**

In `Harpia/parser/parser.go`:
Update `parseReatribuicao()` to recognize `??=`, `||=`, `&&=`.
Check for `TokenCoalescenciaNula` followed by `=`, `TokenOu` followed by `=`, `TokenE` followed by `=`.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test -v -run TestAtribucaoLogicaECoalescente ./tests/parser_test/...`
Expected: PASS

- [ ] **Step 5: Commit changes**

```bash
git add parser/parser.go tests/parser_test/parser_test.go
git commit -m "feat(parser): add support for logical and coalescing assignment operators"
```

---

### Task 4: Collection Destructuring (`var [a, b] = lista`, `const { x, y } = ponto`)

**Files:**
- Modify: `Harpia/parser/ast_nodes.go`
- Modify: `Harpia/parser/parser.go`
- Modify: `Harpia/tests/parser_test/parser_test.go`

**Interfaces:**
- Produces: `ElementoDesestruturacao` struct.
- Produces: `DeclDesestruturacao` struct.

- [ ] **Step 1: Write failing test for collection destructuring**

```go
func TestDesestruturacao(t *testing.T) {
	codigo := `
		var [a, b, ...resto] = [1, 2, 3, 4]
		const { nome, idade } = usuario
	`
	p := NewParserFromString(codigo, "teste_desest.hrp")
	ast, err := p.Parse()
	if err != nil {
		t.Fatalf("Erro sintático em desestruturação: %v", err)
	}
	if len(ast.Instrucoes) != 2 {
		t.Fatalf("Esperava 2 instruções, obteve %d", len(ast.Instrucoes))
	}
	decl1, ok := ast.Instrucoes[0].(*DeclDesestruturacao)
	if !ok || decl1.TipoEstrut != "lista" || decl1.Resto != "resto" {
		t.Errorf("Falha na desestruturação de lista: %+v", ast.Instrucoes[0])
	}
	decl2, ok := ast.Instrucoes[1].(*DeclDesestruturacao)
	if !ok || decl2.TipoEstrut != "mapa" || !decl2.Constante {
		t.Errorf("Falha na desestruturação de mapa: %+v", ast.Instrucoes[1])
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test -v -run TestDesestruturacao ./tests/parser_test/...`
Expected: FAIL with syntax error on `[` or `{` after `var`/`const`.

- [ ] **Step 3: Define DeclDesestruturacao & implement parseDesestruturacao**

In `Harpia/parser/ast_nodes.go`:
```go
type ElementoDesestruturacao struct {
	Nome        string
	ValorPadrao BaseNode
}

type DeclDesestruturacao struct {
	Constante   bool
	TipoEstrut  string // "lista" ou "mapa"
	Campos      []ElementoDesestruturacao
	Resto       string
	Valor       BaseNode
	Inicio      lexer.PosicaoToken
	Fim         lexer.PosicaoToken
}

func (n *DeclDesestruturacao) isExpr() {}
func (n *DeclDesestruturacao) Pos() (lexer.PosicaoToken, lexer.PosicaoToken) { return n.Inicio, n.Fim }
```

In `Harpia/parser/parser.go`:
In `parseDeclaracao()`: check if `p.atual` after `var`/`const` is `TokenColcheteEsq` or `TokenChaveEsq`. If so, invoke `parseDesestruturacao(constante)`.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test -v -run TestDesestruturacao ./tests/parser_test/...`
Expected: PASS

- [ ] **Step 5: Commit changes**

```bash
git add parser/ast_nodes.go parser/parser.go tests/parser_test/parser_test.go
git commit -m "feat(parser): add list and map destructuring syntax parsing"
```

---

### Task 5: Native Go Fuzz Suite & Full Repository Verification

**Files:**
- Create: `Harpia/tests/parser_test/parser_fuzz_test.go`
- Modify: `Harpia/tests/parser_test/...`

**Interfaces:**
- Produces: `FuzzParser(f *testing.F)` fuzzing suite.

- [ ] **Step 1: Create parser_fuzz_test.go with FuzzParser**

```go
package parser_test

import (
	"testing"

	"github.com/mat-dgruber/Harpia/parser"
)

func FuzzParser(f *testing.F) {
	sementes := []string{
		"var x = 10 + 20 * 3",
		"se x > 5 { imprimir(x) } senao { imprimir(0) }",
		"var [a, b, ...resto] = [1, 2, 3]",
		"const { nome, idade } = usuario",
		"x ??= 10; y ||= Falso",
		"tente { risco() } capture (e) { log(e) } finalmente { limpa() }",
	}

	for _, semente := range sementes {
		f.Add([]byte(semente))
	}

	f.Fuzz(func(t *testing.T, dados []byte) {
		p := parser.NewParserFromString(string(dados), "fuzz.hrp")
		// Must never panic or hang
		_, _ = p.Parse()
	})
}
```

- [ ] **Step 2: Run fuzzing for 10 seconds**

Run: `go test -fuzz=FuzzParser -fuzztime=10s ./tests/parser_test/...`
Expected: PASS with 0 crashes or hangs.

- [ ] **Step 3: Run full repository test suite**

Run: `go test ./...`
Expected: ALL PASS

- [ ] **Step 4: Commit fuzz testing suite**

```bash
git add tests/parser_test/parser_fuzz_test.go
git commit -m "test(parser): add native Go fuzz testing suite for parser resilience"
```
