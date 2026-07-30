# Harpia Lexer & Parser Maturity V2 — Design Specification

**Date:** 2026-07-30  
**Status:** Approved  
**Scope:** `Harpia/lexer` & `Harpia/parser`  

---

## 1. Overview & Goals

This specification defines the second major wave of compiler maturity for Harpia's frontend (`lexer` and `parser`), advancing the language to the performance, expressiveness, and error diagnostics standards of production systems such as Go (`go/parser`), TypeScript, and Rust (`rustc`).

### Primary Objectives:
1. **Lexer Optimization & Modern Literals**: $O(1)$ ASCII lookup tables for zero-alloc character classification, visual numeric separators (`1_000_000`), alternative numeric bases (`0x`, `0b`, `0o`), and raw multi-line string literals (`` `text` ``).
2. **Rich IDE-Grade Error Diagnostics**: Expanded `ErroSintatico` with error codes (`HRP-E101`), source line snippets with column indicators (`^^^^^`), and actionable Portuguese resolution tips.
3. **Advanced Syntactic Constructs**: Structural pattern matching (`escolha`), complete null-safe optional chaining (`?.()`, `?.[idx]`), and AST docstring preservation (`///` or `/** */`).

---

## 2. Architecture & Detailed Component Design

### Phase 1: Lexer Optimizations, Modern Literals & Rich Diagnostics

#### 2.1 Lexer ASCII Lookup Table & Tokenization
- **Lookup Table**: A static `[256]byte` lookup array classifying ASCII byte flags:
  - `flagDigito (1 << 0)`
  - `flagLetra (1 << 1)`
  - `flagEspaco (1 << 2)`
  - `flagDelimitador (1 << 3)`
  - `flagHex (1 << 4)`
- **Binary, Hexadecimal, Octal & Numeric Separators**:
  - `0x` or `0X` followed by hex digits (`0-9`, `a-f`, `A-F`, `_`).
  - `0b` or `0B` followed by binary digits (`0-1`, `_`).
  - `0o` or `0O` followed by octal digits (`0-7`, `_`).
  - Underscores `_` allowed between digits (e.g. `1_000_000`, `0xFF_FF`), ignored during numeric evaluation.
- **Raw Strings**:
  - Delimited by backticks (`` ` ``).
  - Preserves verbatim newlines and internal quotes without escape sequences (`\n`, `\"`).

#### 2.2 Rich Error Diagnostics (`ErroSintatico`)
- Structure updated in `parser/parser.go`:
```go
type ErroSintatico struct {
    Codigo     string             // ex: "HRP-E001"
    Mensagem   string             // ex: "era esperado o token ';'"
    Inicio     lexer.PosicaoToken
    Fim        lexer.PosicaoToken
    LinhaFonte string             // a linha de código onde o erro ocorreu
    Indicador  string             // ex: "             ^^^^^"
    Dica       string             // ex: "Dica: adicione ';' ao final da instrução"
}
```
- `Parser.formatarErro(codigo, msg, dica)` constructs formatted diagnostic output for terminal and LSP consumers.

---

### Phase 2: Advanced Language Syntax & AST Docstrings

#### 2.3 Structural Pattern Matching (`escolha`)
- AST Nodes:
```go
type CasoEscolha struct {
    NoBase
    Padrao     BaseNode   // Literal, Identificador, DeclDesestruturacao, "_"
    Guarda     BaseNode   // Expressao condicional "quando <expr>" (opcional)
    Corpo      BaseNode   // Bloco de instrucoes
}

type DeclEscolha struct {
    NoBase
    Expressao BaseNode
    Casos     []CasoEscolha
}
```
- Grammar:
```harpia
escolha expressao {
    caso [a, b, ...resto] quando a > 0 {
        imprimir(a)
    }
    caso { status: 200 } {
        imprimir("sucesso")
    }
    caso _ {
        imprimir("padrão")
    }
}
```

#### 2.4 Complete Optional Chaining (`?.()`, `?.[idx]`)
- AST Nodes:
```go
type ChamadaOpcional struct {
    NoBase
    Funcao     BaseNode
    Argumentos []BaseNode
}

type IndexacaoOpcional struct {
    NoBase
    Objeto    BaseNode
    Argumento BaseNode
}
```
- Parser logic in `parseExpressaoSufixo()` / `parseAcessoOuChamada()`:
  - `funcao?.(arg)` -> `ChamadaOpcional`
  - `objeto?.[indice]` -> `IndexacaoOpcional`

#### 2.5 AST Docstrings Preservation
- Fields added to AST nodes:
  - `DocComments string` on `DeclFuncao`, `DeclClasse`, `DeclVar`, `DeclEnum`, `DeclInterface`.
- Lexer/Parser integration:
  - Comments starting with `///` or `/**` immediately preceding a declaration are attached to `DocComments`.

---

## 3. Verification & Testing Strategy

1. **Unit Tests**:
   - `TestLiteraisNumericosETextuais`: Hex, binary, octal, underscore separators, raw strings.
   - `TestDiagnosticosRicos`: Formatting of `ErroSintatico` with line snippets and Portuguese tips.
   - `TestPatternMatchingEscolha`: Pattern matching grammar parsing and AST representation.
   - `TestEncadeamentoOpcionalCompleto`: Optional calls `?.()` and optional indexing `?.[i]`.
   - `TestDocCommentsAST`: Capturing `///` comments onto AST nodes.
2. **Performance Benchmarks**:
   - `BenchmarkLexer`: MB/s throughput test on large Harpia scripts.
   - `BenchmarkParser`: Operations/sec and allocations/op on complex AST structures.
3. **Fuzz Testing**:
   - Extended `FuzzParser` covering raw strings, alternative bases, `escolha`, and optional chaining.
