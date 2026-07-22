# Pacote `parser` (Analisador Sintático do Harpia)

O pacote `parser` implementa o **Analisador Sintático** (Parser) do **Harpia**. Utilizando o consagrado algoritmo de **Parser de Descida Recursiva Manual** (_Manual Recursive Descent Parser_), ele traduz a torrente de tokens lineares produzida pelo Lexer e constrói a **Árvore de Sintaxe Abstrata** (AST) hierárquica e tipada do script.

Diferente de geradores automáticos, a escrita manual deste parser garante máxima velocidade de compilação, flexibilidade sintática na omissão opcional de delimitadores e mapeamentos detalhados de tokens físicos para tracebacks de erros ricos inteiramente em Português.

---

## 📖 Índice

1. [A Base Sintática: `BaseNode` e Polimorfismo](#-a-base-sintática-basenode-e-polimorfismo)
2. [Principais Nós da Árvore Sintática (AST)](#-principais-nós-da-árvore-sintática-ast)
3. [Mecânica de Análise e Precedência de Operadores](#-mecânica-de-análise-e-precedência-de-operadores)
   - [O Padrão Elegante `parseEsqLst`](#o-padrão-elegante-parseesqlst)
   - [Delimitação Flexível de Instruções](#delimitação-flexível-de-instruções)
4. [Associação e Rastreamento de Erros (Traceback)](#-associação-e-rastreamento-de-erros-traceback)
5. [Exemplo de Visualização de AST](#-exemplo-de-visualização-de-ast)

---

## 🔌 A Base Sintática: `BaseNode` e Polimorfismo

O arquivo `ast_nodes.go` unifica toda a modelagem da AST usando os recursos nativos de polimorfismo estrutural do Go:

- **A Interface `BaseNode`**:
  ```go
  type BaseNode interface { isExpr() }
  ```
  Qualquer struct que represente um nó sintático (ex: `DeclVar`, `OpBinaria`, `TextoLiteral`) implementa a assinatura vazia de marcação `isExpr()`. Isso permite ao compilador Go validar em tempo de compilação que apenas nós sintáticos válidos componham a árvore hierárquica, mantendo o Parser extremamente seguro e fortemente tipado.

---

## 🌳 Principais Nós da Árvore Sintática (AST)

Os nós da AST representam as estruturas semânticas da linguagem. Eles são divididos por categorias:

### 1. Estruturas Raiz e Blocos:

- **`Programa`**: O nó raiz de toda a AST, encapsulando a lista de instruções, o arquivo de origem e as posições de depuração.
- **`Bloco`**: Uma lista de instruções pertencentes a um escopo físico (delimitado por chaves `{}`).

### 2. Declarações e Atribuições:

- **`DeclVar`**: Representa declarações de variáveis mutáveis (`var x`) ou constantes imutáveis (`const y`), suportando anotações estáticas de tipo (ex: `var x: Inteiro`). O tipo é armazenado no campo `Tipo string` (vazio se nenhuma anotação for informada).
- **`Reatribuicao`**: Modificações de variáveis ou reatribuições com acumuladores aritméticos (ex: `x += 1`, `y //= 2`).

### 3. Operações lógicas, de bit e matemáticas:

- **`OpBinaria`**: Cálculos aritméticos ou lógicos envolvendo dois operandos (esquerda e direita).
- **`OpUnaria`**: Inversões lógicas (`nao x`), bitwise (`~x`) ou aritméticas (`-x`).
- **`OpPipe`**: Encadeamentos expressivos utilizando o operador pipe (`|>`).

### 4. Coleções Literais Primitivas:

- **`ListaLiteral`**: Vetores ordenados mutáveis (ex: `[1, 2]`).
- **`TuplaLiteral`**: Vetores ordenados imutáveis (ex: `(1, 2)`).
- **`MapaLiteral` / `MapaPar`**: Estruturas de chaves e valores (dicionários), aceitando chaves estáticas ou expressões dinâmicas (ex: `{[expressao]: valor}`).

---

## ⚙️ Mecânica de Análise e Precedência de Operadores

O parser do Harpia analisa os tokens em descida recursiva, onde a ordem em que as funções chamam umas às outras dita a precedência matemática e lógica.

```
       [Expressão]
            │
            ▼
         [Pipe]
            │
            ▼
       [Disjunção]  ➔ (ou)
            │
            ▼
       [Conjunção]  ➔ (e)
            │
            ▼
       [Inversão]   ➔ (nao)
            │
            ▼
       [Comparação] ➔ (==, !=, <, <=, >, >=, em, instancia de)
            │
            ▼
      [Bitwise OR]  ➔ (|)
            │
            ▼
     [Bitwise XOR]  ➔ (^)
            │
            ▼
     [Bitwise AND]  ➔ (&)
            │
            ▼
     [Deslocamento] ➔ (<<, >>)
            │
            ▼
    [Aritmética +/-]➔ (+, -)
            │
            ▼
   [Multiplicativa] ➔ (*, /, //, %)
            │
            ▼
    [Unários Fator] ➔ (+, -, ~)
            │
            ▼
       [Potência]   ➔ (**)
            │
            ▼
       [Primário]   ➔ (Acesso com ponto ., chamadas ( ), indexações [ ])
            │
            ▼
        [Átomo]     ➔ (Identificadores, literais básicos, listas, mapas)
```

### O Padrão Elegante `parseEsqLst`

Para reduzir centenas de linhas redundantes e loops repetitivos de precedência de operadores binários associativos à esquerda, o arquivo `parser.go` introduz um método genérico excepcional baseado em closures:

```go
func (p *Parser) parseEsqLst(
    proximo func() (BaseNode, error),
    proxOp func() (string, bool),
) (BaseNode, error) {
    esq, err := proximo()
    if err != nil { return nil, err }
    for {
        op, ok := proxOp()
        if !ok { return esq, nil }
        dir, err := proximo()
        if err != nil { return nil, err }
        esq = &OpBinaria{esq, op, dir}
    }
}
```

Cada nível de precedência (como soma, multiplicação e operações bitwise) simplesmente invoca `parseEsqLst` passando como parâmetros:

1. O método correspondente ao nível de precedência superior subsequente.
2. Uma closure rápida que valida e consome o operador do nível atual.

### Delimitação Flexível de Instruções

O método `consome(";")` implementa a flexibilidade sintática do Harpia. Ele aceita o caractere `;` se estiver explícito, mas aceita de forma transparente o token `TokenNovaLinha` (\n) ou o fim do arquivo (EOF) como divisores e terminadores lógicos automáticos de comandos, permitindo uma escrita limpa e livre de ponto-e-vírgula obrigatório.

---

## ⚠️ Nó de Tratamento de Exceções (Sprint 5)

A partir do Sprint 5, o parser constrói o nó composto `TenteCaptureFinalmente` para suportar `tente { ... } capture (erro) { ... } finalmente { ... }`:

```go
type TenteCaptureFinalmente struct {
    TenteBlock      *Bloco
    CaptureBlock    *Bloco  // opcional
    FinalmenteBlock *Bloco  // opcional
    NomeErro        string // identificador exposto pelo capture
}
```

`Capture`/`Finalmente` são opcionais e independentes: as três variantes sintáticas abaixo são aceitas via `parseTenteCapture`:

| Forma                                                     | Captura erro? | Finally roda sempre? |
| :-------------------------------------------------------- | :-----------: | :------------------: |
| `tente { ... } capture (erro) { ... }`                    |      sim      |         não          |
| `tente { ... } finalmente { ... }`                        | não, propaga  |         sim          |
| `tente { ... } capture (erro) { ... } finalmente { ... }` |      sim      |         sim          |

> ⚠️ O lexer promove os identificadores curtos `e`, `ou`, `nao`, `de`, `em` a palavras-chave (operadores lógicos). Por isso, **evite** usar esses nomes para a binding do erro em `capture (...)` — prefira nomes compostos como `erro`.

---

## 🎯 Associação e Rastreamento de Erros (Traceback)

Um dos maiores diferenciais do parser escrito à mão do Harpia é o registrador geográfico de coordenadas.

### `registrar(node BaseNode, tok *lexer.Token)`

Toda vez que o parser processa e inst instancia um nó sintático da AST (`parseDeclaracao()` e `parseExpressao()`), ele invoca a função `registrar()`, salvando a correspondência do nó físico com o respectivo token inicial do lexer no mapa global `Parser.posicoes`:

```go
p.posicoes[node] = tok
```

Durante a execução na Máquina Virtual, se um nó sintático causar uma exceção em tempo de execução (como uma divisão por zero ou tentativa de somar tipos incompatíveis), a VM recupera o token original a partir deste mapa. Com isso, o interpretador consegue imprimir o arquivo de origem, a linha exata e a coluna da instrução causadora do erro, sublinhando-a graficamente para o desenvolvedor.

---

## 💻 Exemplo de Uso e Visualização de AST em Go

Abaixo está um snippet demonstrando como instanciar o Parser a partir de um trecho de código em Harpia e exportar sua AST formatada em JSON:

```go
package main

import (
	"fmt"
	"log"

	"github.com/mat-dgruber/Harpia/parser"
)

func main() {
	codigo := `
		var x = 10 + 20 * 3
		se x > 50 {
			imprimir("Valor alto!")
		}
	`

	// Instancia o analisador sintático a partir do código-fonte
	p := parser.NewParserFromString(codigo, "exemplo.hrp")

	// Constrói a Árvore de Sintaxe Abstrata (AST)
	ast, err := p.Parse()
	if err != nil {
		log.Fatalf("Erro de compilação sintática: %v", err)
	}

	// Converte a AST em JSON formatado para inspeção
	jsonAst, err := parser.Ast2string(ast)
	if err != nil {
		log.Fatalf("Erro ao serializar AST: %v", err)
	}

	fmt.Println(string(jsonAst))
}
```

Ao submeter a expressão `10 + 20 * 3` para análise sintática, a função de depuração `Ast2string` serializa o nó gerando a seguinte árvore hierárquica estruturada:

```json
{
  "Esq": {
    "Valor": "10"
  },
  "Operador": "+",
  "Dir": {
    "Esq": {
      "Valor": "20"
    },
    "Operador": "*",
    "Dir": {
      "Valor": "3"
    }
  }
}
```

Como observado no JSON acima, o nó de multiplicação `*` foi corretamente aninhado como filho direito do nó de soma `+`. Isso prova que a ordem de precedência gramatical do Harpia foi perfeitamente executada, garantindo que `20 * 3` seja avaliado prioritariamente antes da adição de `10`.

