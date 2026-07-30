# Pacote `lexer` (Analisador Léxico do Harpia)

O pacote `lexer` implementa o **Analisador Léxico** (também conhecido como *Scanner* ou *Tokenizer*) do **Harpia**. Escrito inteiramente à mão em Go para máxima performance (zero alocações de memória por caractere) e design idiomático, ele realiza a varredura linear de um slice de bytes UTF-8 (`[]byte`) do código-fonte e os transforma em uma sequência ordenada de tokens lógicos para o Parser.

O lexer rastreia coordenadas espaciais precisas de cada token (linha, coluna, índice) sem alocações na Heap, provendo diagnósticos e tracebacks de depuração ricos em português.

---

## 📖 Índice

1. [Estrutura e Representação de Coordenadas (Zero-Alloc)](#-estrutura-e-representação-de-coordenadas-zero-alloc)
2. [Tabelas de Decisão e Reconhecedor DFA](#-tabelas-de-decisão-e-reconhecedor-dfa)
3. [Algoritmo Operacional e Máquina de Estados](#-algoritmo-operacional-e-máquina-de-estados)
4. [Recursos Avançados da Linguagem](#-recursos-avançados-da-linguagem)
5. [Documentação de Métodos e Funções](#-documentação-de-métodos-e-funções)
6. [Fuzzing e Testes de Regressão](#-fuzzing-e-testes-de-regressão)

---

## 📍 Estrutura e Representação de Coordenadas (Zero-Alloc)

O arquivo `tokens.go` define a base atômica do processo de compilação:

### `PosicaoToken`
Guarda por **valor** (struct sem ponteiros) as coordenadas geométricas do caractere sob análise:
```go
type PosicaoToken struct {
    Coluna int // Índice a partir do início da linha ativa (base 1).
    Linha  int // Número da linha física corrente (base 1).
    Indice int // Offset absoluto em bytes desde o início do arquivo (base 0).
}
```

### `Token`
Representa o objeto unificado contendo a classificação (`Tipo`), o lexema textual exato (`Valor`), e coordenadas espaciais (`Inicio` e `Fim` do tipo `PosicaoToken` por valor):
```go
type Token struct {
    Tipo   TokenType
    Valor  string
    Inicio PosicaoToken
    Fim    PosicaoToken
}
```

---

## 🔍 Tabelas de Decisão e Reconhecedor DFA

O lexer utiliza um **Reconhecedor Automato Finito Determinístico (DFA)** por `switch` estático para resolver operadores e delimitadores em tempo $O(1)$ sem alocações dinâmicas de mapas de string:

1. **DFA de Operadores (`switch ch`)**: Resolve de 1 a 3 caracteres (`==`, `!=`, `<=`, `>=`, `=>`, `//`, `//=`, `...`, `??`, `?.`, `|>`, etc.) diretamente via saltos condicionais compilados.
2. **`tokensIdentificadores`**: Tabela de palavras reservadas e estruturas de controle lógicas nativas da linguagem (`se`, `senao`, `enquanto`, `para`, `var`, `const`, `funcao`, `Verdadeiro`, `Falso`, `tente`, `capture`, `finalmente`). Se o identificador lido coincidir com uma chave, é promovido a palavra-chave. Do contrário, permanece como `TokenIdentificador`.

---

## ⚙️ Algoritmo Operacional e Máquina de Estados

```
                  [Código Fonte UTF-8 []byte]
                             │
                        NewLexer()
                             │
                             ▼
                +-------------------------+
                |    loop PróximoToken()  | <---------------+
                +-------------------------+                 |
                             │                              |
                             ├─► ignorarEspacos()           |
                             │                              |
                             ├─► Comentário/Shebang? ───────┘
                             │     └─► pular / doc comment   | (Reinicia busca)
                             │                              |
                             ▼                              |
                     [Decisão DFA switch]                   |
                             │                              |
      ┌──────────────────────┼──────────────────────┐       |
      ▼                      ▼                      ▼       |
Operador/Pontuação?    Aspas ", ' ou `?       Letras ou _?  |
      │                      │                      │       |
Consome via DFA        lerTexto() / Raw       lerIdentificador()
      │                      │                      │       |
      │                      │                 Consulta se  |
      │                      │                 é keyword    |
      ▼                      ▼                      ▼       |
[Retorna Token]        [Retorna Token]        [Retorna Token]       |
      │                      │                      │       |
      └──────────────────────┴──────────────────────┴───────┘
```

---

## 🚀 Recursos Avançados da Linguagem

- **Shebang (`#!`)**: Suporte nativo na primeira linha de scripts executáveis no Linux/macOS.
- **Comentários de Bloco Aninhados (`/* ... */`)**: Contagem de profundidade para permitir blocos de comentários dentro de comentários de bloco sem fechamento precoce.
- **Doc Comments (`///` e `/** ... */`)**: Retornam `TokenComentarioDoc` para servir geradores de documentação e LSP.
- **Literais Numéricos Avançados**:
  - Hexadecimal (`0x1F_2A`), Binário (`0b1010_1100`), Octal (`0o755_644`).
  - Separadores de dígitos `_` em inteiros e decimais (`1_000_000`).
  - Notação científica/exponencial (`1e10`, `3.14e-5`, `2.5E+8`).
- **Raw Strings (Backticks `` `...` ``)**: Textos multilinha literais sem processamento de escape.
- **Normalização CRLF Em-Fluxo**: Tratamento nativo de `\r\n` diretamente no cursor `avancar()` sem duplicar buffers em memória.

---

## 🛠️ Documentação de Métodos e Funções

- `NewLexer(entrada string) *Lexer`: Aloca o lexer sobre o slice de bytes UTF-8 da entrada.
- `avancar()`: Decodifica a próxima runa UTF-8 via `utf8.DecodeRune` e atualiza linha/coluna.
- `peek() rune` / `peekN(n int) rune`: Lookahead zero-alloc de $1..N$ caracteres à frente.
- `PróximoToken() *Token`: Ponto de entrada principal executado pelo Parser.
- `Estado() SnapshotLex` / `Restaurar(s SnapshotLex)`: Salva e restaura o estado exato do cursor para backtracking.

---

## 🧪 Fuzzing e Testes de Regressão

O pacote conta com uma suíte nativa de **Fuzz Testing** em `tests/lexer_test/lexer_fuzz_test.go`:

```bash
cd Harpia && go test -fuzz=FuzzLexer -fuzztime=10s ./tests/lexer_test/...
```

Garante resiliência absoluta contra sequências arbitrárias de bytes, prevenindo panics, loops infinitos e travamentos.
