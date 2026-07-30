# Especificação de Design: Reformulação e Maturidade do Parser do Harpia

**Data:** 2026-07-30  
**Módulo Target:** `Harpia/parser`  
**Objetivo:** Elevar o analisador sintático (`Harpia/parser`) do Harpia ao nível de maturidade e robustez de linguagens de produção (Go, Rust, TypeScript, Python), implementando recuperação de erros sintáticos (Panic Mode com ressincronização), coordenadas geográficas zero-alloc integradas nos nós da AST, operadores de atribuição lógica/coalescente, desestruturação de coleções e resiliência via Go Fuzz testing.

---

## 1. Contexto e Objetivos

O analisador sintático do Harpia é um **Parser de Descida Recursiva Manual** (*Manual Recursive Descent Parser*). Atualmente ele constrói a AST corretamente, mas apresenta limitações de maturidade:
1. **Interrupção no Primeiro Erro**: Interrompe o processamento na primeira falha sintática encontrada.
2. **Coordenadas de Nós via Mapa Externo**: Associa nós a tokens por um mapa global em memória (`p.posicoes`), em vez de armazenar o intervalo de posições diretamente em cada nó.
3. **Falta de Atribuição Lógica/Coalescente**: Operadores como `??=`, `||=`, `&&=` não têm suporte sintático nativo.
4. **Falta de Desestruturação**: Declarações `var [a, b] = lista` ou `const { x, y } = ponto` não possuem parsing dedicado.

---

## 2. Arquitetura e Alterações Detalhadas

### 2.1. Interface `BaseNode` e Coordenadas Zero-Alloc
No arquivo `Harpia/parser/ast_nodes.go`:

```go
type BaseNode interface {
    isExpr()
    Pos() (lexer.PosicaoToken, lexer.PosicaoToken)
}
```

Cada struct de nó da AST (ex: `DeclVar`, `OpBinaria`, `Bloco`, `Programa`, `Chamada`, `TextoLiteral`, etc.) ganha os campos:
```go
Inicio lexer.PosicaoToken
Fim    lexer.PosicaoToken
```
E implementa `Pos() (lexer.PosicaoToken, lexer.PosicaoToken)` retornando `(n.Inicio, n.Fim)`.

---

### 2.2. Recuperação de Erros Sintáticos (Panic Mode)
No arquivo `Harpia/parser/parser.go`:

```go
type ErroSintatico struct {
    Mensagem string
    Inicio   lexer.PosicaoToken
    Fim      lexer.PosicaoToken
}

type Parser struct {
    // ...
    erros []ErroSintatico
}
```

Ao detectar um token inesperado em comandos/declarações:
1. Adiciona um `ErroSintatico` ao slice `p.erros`.
2. Executa o método `sincronizar()`:
   - Descarta os tokens correntes até encontrar um separador de comando (ponto-e-vírgula `;`, quebra de linha `\n`) ou o início de uma declaração segura (`var`, `const`, `funcao`, `se`, `enquanto`, `para`, `retorne`, `tente`).
3. Retorna da instrução atual com `nil` e continua o parsing do resto da aplicação.

---

### 2.3. Operadores de Atribuição Lógica e Coalescente
Extensão em `parseReatribuicao()` / `parseDeclaracao()` no parser:
- Suporte aos operadores `??=`, `||=`, `&&=`.
- Representados no nó `Reatribuicao` existente (`Operador: "??="`, `||=`, `&&=`).

---

### 2.4. Desestruturação de Listas e Objetos
Novo nó em `Harpia/parser/ast_nodes.go`:

```go
type ElementoDesestruturacao struct {
    Nome        string
    ValorPadrao BaseNode // opcional
}

type DeclDesestruturacao struct {
    Constante   bool
    TipoEstrut  string // "lista" ou "mapa"
    Campos      []ElementoDesestruturacao
    Resto       string // identificador rest (...resto)
    Valor       BaseNode
    Inicio      lexer.PosicaoToken
    Fim         lexer.PosicaoToken
}
```

Sintaxe aceita em `parseDeclaracao()`:
- `var [a, b, ...resto] = lista`
- `const { nome, idade } = usuario`

---

### 2.5. Suíte de Fuzzing & Testes
Arquivo `Harpia/tests/parser_test/parser_fuzz_test.go`:
- Implementa `FuzzParser(f *testing.F)` com entradas variadas de código Harpia.
- Garante que a combinação `Lexer + Parser` nunca entre em loop infinito ou cause panic com entradas aleatórias de bytes.

---

## 3. Plano de Testes e Validação

1. **Testes Unitários de Novas Sintaxes**:
   - `TestRecuperacaoErrosSintaticos`: Testa código com múltiplos erros e verifica se todos são capturados sem panic.
   - `TestAtribucionCoalescenteELogica`: Testa `??=`, `||=`, `&&=`.
   - `TestDesestruturacao`: Testa parsing de desestruturação de listas e mapas.
2. **Fuzz Testing**:
   - `go test -fuzz=FuzzParser -fuzztime=10s ./tests/parser_test/...`
3. **Regressão Global**:
   - `go test ./...` para garantir que o repositório todo continua compilando e passando em 100% dos testes.
