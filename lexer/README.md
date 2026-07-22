# Pacote `lexer` (Analisador Léxico do Harpia)

O pacote `lexer` implementa o **Analisador Léxico** (também conhecido como *Scanner* ou *Tokenizer*) do **Harpia**. Escrito inteiramente à mão em Go por questões de performance e design idiomático, ele realiza a varredura linear de caracteres Unicode UTF-8 do código-fonte e os transforma em uma sequência ordenada de tokens lógicos inteligíveis para o Parser.

O lexer também rastreia as coordenadas geográficas exatas de cada token, provendo as bases para a geração de tracebacks de depuração ricos em português.

---

## 📖 Índice

1. [Estrutura e Representação de Coordenadas](#-estrutura-e-representação-de-coordenadas)
2. [Tabelas de Decisão (O(1) Lookup)](#-tabelas-de-decisão-o1-lookup)
3. [Algoritmo Operacional e Máquina de Estados](#-algoritmo-operacional-e-máquina-de-estados)
   - [Cursor e Lookahead (Espreitadela)](#cursor-e-lookahead-espreitadela)
   - [Leitores de Cadeias Complexas](#leitores-de-cadeias-complexas)
4. [Documentação de Métodos e Funções](#-documentação-de-métodos-e-funções)
5. [Exemplo de Integração em Go](#-exemplo-de-integração-em-go)

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
2. **`tokensIdentificadores`**: Mapa de palavras reservadas e estruturas de controle lógicas nativas da linguagem (como `var`, `func`, `se`, `para`, `Verdadeiro`, `tente`, `capture`, `finalmente`). Atua como um classificador: se o identificador lido pelo lexer coincidir com alguma chave, o token é promovido à palavra-chave. Do contrário, permanece como um identificador ordinário de variável ou método.

> **Limitação conhecida**: identificadores curtos como `e`, `ou`, `nao`, `de`, `em` colidem com os operadores lógicos (`e`/`ou`) reservados do lexer e não podem ser usados como nomes de variáveis (incluindo a binding do erro em `capture`). Use nomes compostos (`erro`, `minhaVariavel`). *Tracking: refator de contextual keywords em sprint futuro.*

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
      │                      │                      │       |
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

## 🛠️ Documentação de Métodos e Funções

Abaixo estão descritos todos os métodos implementados na estrutura `Lexer`:

### `NewLexer(entrada string) *Lexer`
Aloca e prepara uma nova instância operacional de `Lexer` para o código-fonte fornecido. Normaliza as quebras de linha substituindo retornos de carro (`\r`), calcula o tamanho total em runas do texto e cria o cache indexado de bytes para permitir saltos rápidos e seguros sem vazamentos ou estouro de inteiros.

### `fimDeArquivo() bool`
Verifica se o cursor de leitura atingiu ou ultrapassou os limites do código-fonte, garantindo a prevenção de estouros de array ao tentar ler além do EOF.

### `proximoCarater() string`
Espreita o próximo caractere sem deslocar o cursor físico do lexer. É a base de suporte para algoritmos gulosos de correspondência de caracteres (lookahead 1).

### `caraterRelativo(offset int) string`
Permite um lookahead arbitrário (positivo ou de recuo) para dar suporte à análise de sequências complexas como comentários JSX (`<!-- -->`).

### `ignorarComentarioHTML()`
Varre e consome silenciosamente todo o conteúdo de comentários HTML (`<!-- ... -->`) encontrados em arquivos JSX do ecossistema do Harpia.

### `avancar()`
Incrementa o cursor físico de leitura em exatamente uma posição Unicode e atualiza os campos `linha` e `coluna` com base na presença de caracteres especiais como `\n`.

### `posicaoAtual() *PosicaoToken`
Retorna as coordenadas geométricas instantâneas onde o caractere sob análise se encontra no documento.

### `ignorarEspacos()`
Avança o cursor descartando caracteres irrelevantes de formatação como espaço simples (` `) e tabulações (`\t`), sem consumir a quebra de linha `\n` que delimita instruções.

### `ignorarComentario()`
Consome caracteres sequencialmente após encontrar um indicador de comentário (`#` ou `//`), parando ao atingir uma quebra de linha ou o EOF.

### `subString(inicio, fim int) string`
Extrai de forma ultra rápida a substring Unicode delimitada pelos índices de início e fim conceitual usando o `byteCache` para conversão imediata de runas para offsets de bytes.

### `lerIdentificador() *Token`
Lê uma sequência consecutiva de caracteres permitidos para identificadores. Faz o lookup automático contra a tabela de palavras reservadas `tokensIdentificadores` e promove o token para a constante do compilador se coincidir.

### `lerNumero() *Token`
Consome dígitos e pontos para discernir e instanciar tokens de tipo `TokenInteiro` ou `TokenDecimal`.

### `lerTexto() *Token`
Instancia literais textuais escapando delimitadores correspondentes de forma segura.

### `ProximoToken() *Token`
O coração do lexer. Executa a máquina de estados consumindo a entrada e retornando um objeto `Token` por chamada, que é consumido sequencialmente pelo Parser do compilador Harpia.

---

## 💻 Exemplo de Integração em Go

Abaixo está um snippet demonstrando como instanciar e processar tokens a partir de um trecho de código em Harpia usando Go:

```go
package main

import (
	"fmt"
	"github.com/mat-dgruber/Harpia/lexer"
)

func main() {
	codigo := "var total = 100 + 4.5; # Exemplo de código"
	
	// Inicializa o analisador léxico
	l := lexer.NewLexer(codigo)
	
	for {
		tok := l.ProximoToken()
		
		// Imprime as propriedades do token formatadas
		fmt.Printf(
			"Token Tipo: %-4d | Valor: %-15s | Linha: %d Coluna: %d\n",
			tok.Tipo,
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
