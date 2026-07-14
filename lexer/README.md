# Pacote `lexer` (Analisador Léxico do Portuscript)

O pacote `lexer` implementa o **Analisador Léxico** (também conhecido como *Scanner* ou *Tokenizer*) do **Portuscript**. Escrito inteiramente à mão em Go por questões de performance e design idiomático, ele realiza a varredura linear de caracteres Unicode UTF-8 do código-fonte e os transforma em uma sequência ordenada de tokens lógicos inteligíveis para o Parser.

O lexer também rastreia as coordenadas geográficas exatas de cada token, provendo as bases para a geração de tracebacks de depuração ricos em português.

---

## 📖 Índice

1. [Estrutura e Representação de Coordenadas](#-estrutura-e-representação-de-coordenadas)
2. [Tabelas de Decisão (O(1) Lookup)](#-tabelas-de-decisão-o1-lookup)
3. [Algoritmo Operacional e Máquina de Estados](#-algoritmo-operacional-e-máquina-de-estados)
   - [Cursor e Lookahead (Espreitadela)](#cursor-e-lookahead-espreitadela)
   - [Leitores de Cadeias Complexas](#leitores-de-cadeias-complexas)
4. [Exemplo de Integração em Go](#-exemplo-de-integração-en-go)

---

## 📍 Estrutura e Representação de Coordenadas

O arquivo `tokens.go` define a base atômica do processo de compilação:

### `PosicaoToken`
Guarda de forma estrita as coordenadas geométricas do caractere sob análise:
```go
type PosicaoToken struct {
    Coluna int // Índice a partir do início da linha ativa (base 1).
    Linha  int // Número da linha física corrente (base 1).
    Indice int // Offset absoluto em bytes desde o início do arquivo (base 0).
}
```

### `Token`
Representa o objeto unificado contendo a classificação (`Tipo`), o lexema textual exato (`Valor`), e ponteiros espaciais limitadores (`Inicio` e `Fim` do tipo `PosicaoToken`).
```go
type Token struct {
    Tipo   TokenType
    Valor  string
    Inicio *PosicaoToken
    Fim    *PosicaoToken
}
```

---

## 🔍 Tabelas de Decisão (O(1) Lookup)

O arquivo `helpers.go` centraliza as tabelas de símbolos estáticos para acelerar drasticamente a velocidade de categorização:

1. **`tokensSimples`**: Mapa chave-valor de operadores unários/binários e delimitadores físicos (como `+`, `//`, `==`, `+=`, `[`, `{`). Garante correspondências imediatas em tempo constante $O(1)$.
2. **`tokensIdentificadores`**: Mapa de palavras reservadas e estruturas de controle lógicas nativas da linguagem (como `var`, `func`, `se`, `para`, `Verdadeiro`). Atua como um classificador: se o identificador lido pelo lexer coincidir com alguma chave, o token é promovido à palavra-chave. Do contrário, permanece como um identificador ordinário de variável ou método.

---

## ⚙️ Algoritmo Operacional e Máquina de Estados

O arquivo principal `lexer.go` implementa a estrutura de processamento central:

```
                  [Código Fonte Textual]
                             │
                        NewLexer()
                             │
                             ▼
                +-------------------------+
                |    loop ProximoToken()  | <---------------+
                +-------------------------+                 |
                             │                              |
                             ├─► ignorarEspacos()           |
                             │                              |
                             ├─► Encontrou '#' ? ───────────┘
                             │     └─► ignorarComentario()  | (Reinicia busca)
                             │                              |
                             ▼                              |
                     [Decisão de Tipo]                      |
                             │                              |
      ┌──────────────────────┼──────────────────────┐       |
      ▼                      ▼                      ▼       |
Caractere Operador?    Aspas " ou '?          Letras ou _?  |
      │                      │                      │       |
Consome de forma       lerTexto()             lerIdentificador()
gulosa (composto?)           │                      │       |
      │                      │                 Consulta se  |
      │                      │                 é keyword    |
      │                      │                      │       |
      ▼                      ▼                      ▼       |
[Retorna Token]        [Retorna Token]        [Retorna Token]       |
      │                      │                      │       |
      └──────────────────────┴──────────────────────┴───────┘
```

### Cursor e Lookahead (Espreitadela)

- **`avancar()`**: Move o cursor de leitura em um caractere Unicode e atualiza a runa corrente no campo `carater`. Se encontrar um delimitador de quebra de linha `\n`, incrementa o contador global de linhas e reinicia a coluna do cursor.
- **`proximoCarater()`**: Executa uma operação de *lookahead* (espreitada à frente) para inspecionar a runa imediatamente seguinte, sem alterar o cursor físico de leitura do interpretador. Isso é crucial para que o loop identifique de forma "gulosa" operadores formados por múltiplos caracteres concatenados (como distinguir `=` de `==` ou `+=`).

### Leitores de Cadeias Complexas

- **`lerIdentificador()`**: Consome incrementalmente qualquer caractere alfanumérico ou sublinhado `_`, extraindo o lexema resultante e realizando a consulta classificatória em `tokensIdentificadores`.
- **`lerNumero()`**: Consome dígitos numéricos sequenciais decimais. Se encontrar um ponto `.`, promove o token para `TokenDecimal`. Caso contrário, mantém categorizado como `TokenInteiro`.
- **`lerTexto()`**: Varre cadeias textuais (strings) respeitando as aspas de delimitação correspondentes, identificando e escapando de forma transparente delimitadores internos (como `\"` ou `\'`).

---

## 💻 Exemplo de Integração em Go

Abaixo está um snippet demonstrando como instanciar e processar tokens a partir de um trecho de código em Portuscript usando Go:

```go
package main

import (
	"fmt"
	"github.com/natanfeitosa/portuscript/lexer"
)

func main() {
	codigo := "var total = 100 + 4.5; # Exemplo de código"
	
	// Inicializa o analisador léxico
	l := lexer.NewLexer(codigo)
	
	for {
		tok := l.ProximoToken()
		
		// Imprime as propriedades do token formatadas
		fmt.Printf(
			"Token: %-18s | Valor: %-8s | Linha: %d Coluna: %d\n",
			getEnumName(tok.Tipo), // utilitário conceitual de exibição
			tok.Valor,
			tok.Inicio.Linha,
			tok.Inicio.Coluna,
		)
		
		if tok.Tipo == lexer.TokenFimDeArquivo {
			break
		}
	}
}
```
