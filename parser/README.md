# Pacote `parser` (Analisador Sintático do Harpia)

> Documentação externa do pacote `parser` — companion do `CLAUDE.md` raiz.

---

## 1. Visão Geral

O pacote `parser` implementa o **Analisador Sintático** (_Parser_) do Harpia. Utilizando o algoritmo clássico de **Descida Recursiva Manual** (_Manual Recursive Descent Parser_), ele consome a torrente linear de tokens produzida pelo pacote `lexer` e a transforma em uma **Árvore de Sintaxe Abstrata** (AST) hierárquica, fortemente tipada e pronta para a máquina virtual (VM).

Diferente de geradores automáticos (yacc/bison/antlr), a escrita manual deste parser garante:

- **Velocidade de compilação**: zero indireções via tabelas geradas; cada nó é uma `struct` Go concreta.
- **Flexibilidade sintática**: omissão opcional de delimitadores (`;` é opcional quando houver quebra de linha ou EOF).
- **Diagnósticos ricos em Português**: mensagens de erro com código (`HRP-E101`), linha exata, indicador visual (`^^^^^^^`) e dica contextual.

---

## 2. Arquivos do Pacote

| Arquivo          | Responsabilidade                                                                                |
| ---------------- | ----------------------------------------------------------------------------------------------- |
| `ast_nodes.go`   | Definições de tipos da AST (BaseNode, NoBase, Programa, DeclVar, DeclFuncao, OpBinaria, JSX, …). |
| `helpers.go`     | Utilitários: `palavrasChave`, `IsKeyword`, `Ast2string` (serialização JSON para diagnóstico).   |
| `parser.go`      | Implementação do parser: struct `Parser`, métodos `Parse`, `consome`, `registrar`, `sincronizar`, todos os métodos `parse*`, e a struct `ErroSintatico`. |

---

## 3. Dependências

### Internas

- `github.com/mat-dgruber/Harpia/lexer` — fornece `Lexer`, `Token`, `PosicaoToken`, tipos de token e lookahead.

### Externas (stdlib Go)

- `encoding/json` — usado por `Ast2string` para serialização identada.
- `fmt`, `strings`, `regexp` — formatação de erros e dicas contextuais.

> **Sem dependências de terceiros**. O pacote é deliberadamente mantido enxuto.

---

## 4. Arquitetura e Fluxo

### 4.1. Estrutura da AST (`ast_nodes.go`)

Todos os nós implementam a interface marcadora:

```go
type BaseNode interface {
    isExpr()
    Pos() (lexer.PosicaoToken, lexer.PosicaoToken)
}
```

A interface vazia `isExpr()` permite que o compilador Go valide em tempo de compilação que apenas nós sintáticos válidos componham a árvore. A função `Pos` retorna coordenadas de início/fim para reportar erros de runtime com precisão.

Os nós mais relevantes são:

| Categoria           | Nós                                                                                        |
| ------------------- | ------------------------------------------------------------------------------------------ |
| Estruturais         | `Programa`, `Bloco`, `BlocoPara`, `CasoEscolha`                                             |
| Declarações         | `DeclVar`, `DeclFuncao`, `DeclFuncaoParametro`, `DeclClasse`, `DeclEnum`, `DeclInterface`, `DeclExportar`, `DeclImportar`, `DeclTeste`, `DeclEstilo`, `Decorador` |
| Controle de fluxo   | `ExpressaoSe`, `Enquanto`, `RetorneNode`, `TenteCaptureFinalmente`, `DeclEscolha`           |
| Expressões          | `OpBinaria`, `OpUnaria`, `Chamada`, `Atribuicao`, `AcessoPropriedade`, `Identificador`       |
| Literais            | `NumLiteral`, `TextoLiteral`, `LogicoLiteral`, `ListaLiteral`, `MapaLiteral`, `NoJSX`, `NoSeJSX`, `NoParaJSX` |
| Helpers             | `NoBase` (coordenadas), `ExprVetor` (vetor primitivo), `AtributoJSX`, `MetodoAssinatura`   |

### 4.2. O Parser (`parser.go`)

O `Parser` mantém:

- `token`, `próximoToken` — cursor duplo (lookahead de 1 token).
- `posicoes` — mapa `BaseNode → *lexer.Token` para tracebacks em runtime.
- `erros` — slice acumulador de erros sintáticos.
- `ultimoDocComment` — buffer temporário de comentários `///` ou `/** */` capturados entre tokens.

**Fluxo principal:**

```
NewParserFromString(code, file)
        │
        ▼
   Parse() ──► parseDeclaracoes()
                  │
                  ▼
              parseDeclaracao() ──► parseDeclaracaoInterno()
                                          │
                                          ├─► parseImporteDe
                                          ├─► parseBlocoPara
                                          ├─► parseExpressaoSe
                                          ├─► parseEnquanto
                                          ├─► parseRetorne
                                          ├─► parseFuncao
                                          ├─► parseClasse
                                          ├─► parseTeste
                                          ├─► parseTenteCapture
                                          ├─► parseExportar
                                          ├─► parseEnum
                                          ├─► parseEscolha
                                          ├─► parseInterface
                                          ├─► parseVariavel
                                          └─► parseExpressao
```

### 4.3. Precedência de Operadores

A precedência é implementada em camadas recursivas, do menos para o mais precedence:

```
parseExpressaoInterno
    └─► parseTernario
            └─► parsePipe
                    └─► parseCoalescenciaNula
                            └─► parseDisjuncao       (ou lógico)
                                    └─► parseConjuncao (e lógico)
                                            └─► parseInversao
                                                    └─► parseComparacao
                                                            └─► parseBitABitOu
                                                                    └─► parseBitABitExOu
                                                                            └─► parseBitABitE
                                                                                    └─► parseDeslocamento
                                                                                            └─► parseAritBasica
                                                                                                    └─► parseTermo
                                                                                                            └─► parseFator
                                                                                                                    └─► parsePotencia
                                                                                                                            └─► parsePrimario
                                                                                                                                    └─► parseAtomo
```

O helper genérico **`parseEsqLst`** (a "joia arquitetural" do parser) elimina duplicação de laços left-associative:

```go
expr, _ := p.parseEsqLst(p.parseConjuncao, func() (string, bool) {
    if p.token.Tipo == lexer.TokenOuLogico {
        op := p.token.Valor
        p.avancar()
        return op, true
    }
    return "", false
})
```

### 4.4. Recuperação de Erros

`sincronizar()` é o coração da recuperação. Quando um `parse*` falha, o parser consome tokens até uma fronteira segura (newline, `;`, palavra-chave de início de instrução, ou EOF), permitindo reportar **múltiplos erros** em uma única execução.

### 4.5. Renderização de Erros (`ErroSintatico`)

Cada erro é capturado com:

1. **Código**: `HRP-E101` (sintaxe) ou código custom.
2. **Mensagem**: descrição legível.
3. **Dica**: sugestão em português obtida via `ObterDicaSintatica`.
4. **LinhaFonte + Indicador**: linha do código + espaços + `^^^^^^^`.

Saída típica:

```
HRP-E101  script.hrp:linha 12, coluna 5
  const x 10
      ^^^^^^
  dica: Esperava-se '=' após o nome da constante em declaração 'const'.
```

---

## 5. Exemplo de Uso

### 5.1. Parsing a partir de uma string

```go
package main

import (
    "fmt"
    "github.com/mat-dgruber/Harpia/parser"
)

func main() {
    codigo := `
        const PI: Decimal = 3.14159
        var raio: Decimal = 5.0

        funcao area(): Decimal {
            retorne PI * raio * raio
        }

        imprimir(area())
    `

    p := parser.NewParserFromString(codigo, "exemplo.hrp")
    prog, err := p.Parse()
    if err != nil {
        fmt.Println("Erro de parsing:", err)
        return
    }

    for _, decl := range prog.Declaracoes {
        fmt.Printf("- %T\n", decl)
    }
}
```

### 5.2. Serialização em JSON (para inspeção)

```go
import (
    "fmt"
    "github.com/mat-dgruber/Harpia/parser"
)

p := parser.NewParserFromString(codigo, "exemplo.hrp")
prog, _ := p.Parse()

data, _ := parser.Ast2string(prog)
fmt.Println(string(data))
```

Saída (parcial):

```json
{
    "Inicio": { "Linha": 2, "Coluna": 8 },
    "Fim":    { "Linha": 9, "Coluna": 1 },
    "Declaracoes": [
        {
            "Constante": true,
            "Nome": "PI",
            "Tipo": "Decimal",
            "Inicializador": {
                "Valor": "3.14159",
                ...
            }
        },
        ...
    ]
}
```

### 5.3. Verificando palavras-chave

```go
if parser.IsKeyword("se") {
    fmt.Println("'se' é palavra reservada")
}
```

---

## 6. Thread-Safety

O `Parser` **não é thread-safe**. Cada arquivo deve ser analisado por sua própria instância `*Parser`. A estrutura `Programa` resultante é imutável do ponto de vista lógico e pode ser compartilhada entre goroutines pela VM.

---

## 7. Performance

- **Lookahead fixo de 1 token**: mantém complexidade linear (`O(n)`) no tamanho da entrada.
- **Alocações reduzidas**: o embedding de `NoBase` evita indireções via interface no caminho quente.
- **`parseEsqLst`**: elimina duplicação massiva de loops, mantendo o binário compacto.

---

## 8. Nota de Segurança

Este pacote é parte do **compilador oficial do Harpia**. Sua única finalidade é transformar texto-fonte em uma estrutura de dados (AST). **Não é código malicioso** — é software de infraestrutura de compilação, aberto e auditável. Não há execução de código do usuário no parser; toda interpretação é feita pela VM em uma fase separada.

---

## 9. Roadmap

- [ ] Streaming parser para arquivos muito grandes (>10MB).
- [ ] Melhoria na recuperação de erros em JSX aninhado.
- [ ] Relatórios de métricas de qualidade (gramática formal medida).
