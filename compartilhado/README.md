# Pacote `compartilhado` (Utilitários do Portuscript)

O pacote `compartilhado` reúne funções utilitárias e rotinas auxiliares fundamentais para a correta manipulação de strings, tokens, identificadores e conversões numéricas em múltiplos subsistemas do compilador e máquina virtual do **Portuscript** (como lexer, parser, compilador e interpretador).

O foco central deste pacote é fornecer **alta eficiência** no processamento e tratamento seguro de caracteres Unicode (UTF-8), mitigando gargalos comuns na análise de strings em Go.

---

## 📖 Índice

1. [O Desafio de Strings UTF-8 em Go](#-o-desafio-de-strings-utf-8-em-go)
2. [Análise e Manipulação de Strings](#-análise-e-manipulação-de-strings)
   - [Tratamento de Índices e Cache](#tratamento-de-índices-e-cache)
   - [Predicados de Validação de Caracteres](#predicados-de-validação-de-caracteres)
3. [Utilitários de Conversão Numérica](#-utilitários-de-conversão-numérica)
4. [Predicado Alfanumérico Composto](#-predicado-alfanumérico-composto)
5. [Arquitetura e Fluxo no Compilador](#-arquitetura-e-fluxo-no-compilador)
6. [Exemplos Práticos de Integração](#-exemplos-práticos-de-integração)

---

## ⚠️ O Desafio de Strings UTF-8 em Go

Na linguagem Go, as variáveis do tipo `string` são armazenadas internamente como fatias somente-leitura de bytes (`[]byte`), estruturadas nativamente sob a codificação **UTF-8**. 

Embora caracteres básicos da tabela ASCII ocupem apenas 1 byte, caracteres acentuados, letras de alfabetos não-latinos e símbolos complexos (como emojis) podem ocupar de **2 a 4 bytes**.

Isso gera dois problemas principais durante a construção de um Lexer:
1. **Quebra de Caractere**: Tentar acessar uma string em uma posição de byte arbitrária (ex: `str[i]`) pode cortar um caractere multi-byte ao meio, gerando bytes inválidos ou caracteres corrompidos.
2. **Incompatibilidade de Índice**: O caractere visual de índice conceitual `5` pode estar localizado, na verdade, no offset de byte `8` se houver caracteres acentuados antes dele.
3. **Incompatibilidade de Desempenho O(N)**: Para descobrir onde o caractere `N` começa sem um cache, o programa é forçado a percorrer a string inteira a partir do início decodificando runas. Em arquivos de código extensos, isso causaria séria degradação de performance.

O pacote `compartilhado` resolve esses gargalos implementando uma **estrutura de mapeamento por indexação e cache de tempo constante O(1)**.

---

## 🔤 Análise e Manipulação de Strings

### Tratamento de Índices e Cache

As funções a seguir gerenciam a tradução transparente entre índices conceituais de caracteres Unicode (runas) e as posições físicas em bytes.

#### 1. `IndiceBytePorCarater(str string) []int`
Gera e pré-calcula a tabela de correspondências onde cada posição do slice de retorno aponta para o offset inicial do byte daquela runa.
- **Entrada**: `str string`
- **Retorno**: `[]int` (tabela de offsets de bytes)
- **Detalhe do Limite**: O slice resultante possui tamanho `RuneCount + 1`, onde a última posição contém o valor exato de `len(str)`. Isso serve para facilitar operações de slice inclusivas (ex: `str[inicio:fim]`) em que o limite superior aponta para o caractere final do arquivo (EOF).

#### 2. `IndiceCaraterParaByte(str string, indice int, cache []int) int`
Resolve a posição de byte para o índice de caractere informado.
- **Entrada**: `str string`, `indice int` (índice conceitual), `cache []int` (opcional)
- **Mecânica**: Se `cache` não for nulo (`nil`), resolve em **O(1)** consultando diretamente o índice na tabela. Caso contrário, invoca o método alternativo sequencial sem cache (fallback **O(N)**).

#### 3. `IndiceCaraterParaByteSemCache(str string, indice int) int`
Faz a busca linear decodificando runas uma a uma de forma incremental utilizando a biblioteca nativa `unicode/utf8`.
- **Entrada**: `str string`, `indice int`
- **Retorno**: `int` (offset de bytes correspondente)

#### 4. `ObtemCaraterPorIndice(str string, indice int, cache []int) string`
Retorna uma string contendo exatamente um único caractere Unicode válido na posição informada.
- **Entrada**: `str string`, `indice int`, `cache []int`
- **Funcionamento**: Obtém com segurança o offset de byte `inicio` e o offset do caractere seguinte `fim`. Em seguida, extrai a fatia exata `str[inicio:fim]`.

---

### Predicados de Validação de Caracteres

Usados pelo Lexer do Portuscript para análise léxica em tempo de varredura.

#### 5. `ContemApenasLetras(str string) bool`
Inspeciona cada runa da string e verifica se ela é classificada como uma letra válida segundo a tabela Unicode.
- **Entrada**: `str string`
- **Retorno**: `bool` (retorna `false` se a string for vazia ou se contiver espaços, símbolos, números ou pontuações).

#### 6. `ContemApenasDigitos(str string) bool`
Verifica se todas as runas são classificadas como dígitos decimais Unicode.
- **Entrada**: `str string`
- **Retorno**: `bool` (retorna `false` se a string for vazia ou se contiver letras, espaços ou pontuações).

---

## 🔢 Utilitários de Conversão Numérica

O arquivo `numeros.go` centraliza a conversão de literais numéricos reconhecidos pelo analisador sintático para os formatos nativos interpretados pela VM do Portuscript.

#### 1. `StringParaInt(s string) (int64, error)`
Faz o parsing de um literal numérico inteiro em formato de string para um inteiro com sinal de 64 bits (`int64`).
- **Base utilizada**: Base 10 (decimal).
- **Tratamento**: Retorna um erro caso contenha pontos decimais, letras ou caracteres especiais de formatação inválidos.
- **Por que int64?** O Portuscript adota precisão de 64 bits para seus tipos numéricos internos para evitar estouro de limite (overflow) e garantir compatibilidade cruzada estável entre diferentes plataformas (como compilador Go rodando em x86 ou ARM).

#### 2. `StringParaDec(s string) (float64, error)`
Converte strings literais contendo ponto flutuante para a representação decimal nativa do Go de dupla precisão (`float64`).
- **Padrão interpretado**: Padrão IEEE 754 de dupla precisão, aceitando notações científicas normatizadas (ex: `1.23e-4`).
- **Retorno**: `(float64, error)`

---

## 🔀 Predicado Alfanumérico Composto

#### `ContemApenasAlfaNum(str string) bool`
Função de utilidade definida em `compartilhado.go` para verificar se uma cadeia textual é estritamente composta por elementos alfanuméricos (combinação de letras e dígitos).

- **Estrutura lógica**:
  ```go
  func ContemApenasAlfaNum(str string) bool {
      return ContemApenasDigitos(str) || ContemApenasLetras(str)
  }
  ```
- **Por que funciona assim?** Em vez de reescrever um loop customizado, ela delega e unifica as regras para os validadores dedicados, reduzindo o risco de discrepâncias nas verificações de símbolos complexos do Lexer.

---

## 🏗️ Arquitetura e Fluxo no Compilador

Durante o processo de compilação de um script Portuscript, o pacote `compartilhado` é consumido principalmente no início da pipeline:

```
[Código Fonte .pt] ➔ [Lexer] 
                        │
                        ├─► Usa IndiceBytePorCarater() para mapear o arquivo
                        │
                        ├─► Usa ContemApenasLetras() / ContemApenasDigitos()
                        │   para identificar identificadores e palavras-chave
                        │
                        ▼
                 [Tokens Gerados] ➔ [Parser / Compilador]
                                            │
                                            ├─► Usa StringParaInt() / StringParaDec()
                                            │   para converter os literais numéricos
                                            ▼
                                     [AST / Bytecode]
```

---

## 🛠️ Exemplos Práticos de Integração

Abaixo estão alguns snippets de exemplo demonstrando como importar e aplicar de forma limpa as funções do pacote `compartilhado` em Go:

```go
package main

import (
	"fmt"
	"github.com/natanfeitosa/portuscript/compartilhado"
)

func main() {
	// Exemplo de manipulação de string Unicode complexa
	texto := "PortuScript é incrível! 🚀"
	
	// 1. Pré-calculando o cache de índices
	cache := compartilhado.IndiceBytePorCarater(texto)
	
	// 2. Acesso seguro O(1) sem corromper runas multibyte
	// O emoji de foguete '🚀' ocupa 4 bytes em UTF-8
	caracterRocket := compartilhado.ObtemCaraterPorIndice(texto, 24, cache)
	fmt.Printf("Caractere no índice 24: %s\n", caracterRocket) // Saída: 🚀
	
	// 3. Validação léxica de identificadores e palavras reservadas
	fmt.Println(compartilhado.ContemApenasLetras("Funcao")) // Saída: true
	fmt.Println(compartilhado.ContemApenasLetras("se_entao")) // Saída: false (contém underline)
	
	// 4. Conversão numérica de literais da AST
	inteiro, err := compartilhado.StringParaInt("42195")
	if err == nil {
		fmt.Printf("Inteiro convertido com sucesso: %d\n", inteiro)
	}
}
```
