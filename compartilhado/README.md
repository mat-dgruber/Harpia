# Pacote `compartilhado` (Utilitários Industriais do Harpia)

O pacote `compartilhado` reúne funções utilitárias e rotinas auxiliares fundamentais para a correta manipulação de strings, navegabilidade Unicode, validações de identificadores, cache thread-safe e conversões numéricas em múltiplos subsistemas do compilador, runtime e servidores web da linguagem **Harpia**.

O foco central deste pacote é fornecer **alta eficiência e zero contenção de CPU** no processamento e tratamento seguro de caracteres Unicode (UTF-8), com fast-paths para ASCII e cache LRU sharded de alta concorrência.

---

## 📖 Índice

1. [O Desafio de Strings UTF-8 em Go](#-o-desafio-de-strings-utf-8-em-go)
2. [Fast-Path ASCII e Sharded LRU Cache](#-fast-path-ascii-e-sharded-lru-cache)
3. [Análise e Navegação Unicode](#-análise-e-navegação-unicode)
4. [Validadores de Identificadores e Alfanuméricos](#-validadores-de-identificadores-e-alfanuméricos)
5. [Utilitários de Conversão e Bases Numéricas](#-utilitários-de-conversão-e-bases-numéricas)
6. [Arquitetura e Fluxo no Compilador](#-arquitetura-e-fluxo-no-compilador)
7. [Exemplos Práticos de Integração](#-exemplos-práticos-de-integração)

---

## ⚠️ O Desafio de Strings UTF-8 em Go

Na linguagem Go, as variáveis do tipo `string` são armazenadas internamente como fatias somente-leitura de bytes (`[]byte`), estruturadas nativamente sob a codificação **UTF-8**. 

Embora caracteres básicos da tabela ASCII ocupem apenas 1 byte, caracteres acentuados, letras de alfabetos não-latinos e símbolos complexos (como emojis) podem ocupar de **2 a 4 bytes**.

Isso gera três desafios principais em sistemas fullstack e no Lexer:
1. **Quebra de Caractere**: Tentar acessar uma string em uma posição de byte arbitrária (ex: `str[i]`) pode cortar um caractere multi-byte ao meio, gerando bytes inválidos ou caracteres corrompidos.
2. **Incompatibilidade de Índice**: O caractere visual de índice conceitual `5` pode estar localizado no offset de byte `8` se houver caracteres acentuados antes dele.
3. **Contenção e Desempenho O(N)**: Para descobrir onde o caractere `N` começa sem um cache eficiente, o programa é forçado a percorrer a string inteira decodificando runas.

O pacote `compartilhado` resolve esses gargalos combinando **Fast-Path ASCII O(1)** e um **Cache Sharded LRU Thread-Safe de 256 Shards**.

---

## ⚡ Fast-Path ASCII e Sharded LRU Cache

### Fast-Path ASCII
- **`EhASCII(s string) bool`**: Inspeciona se a string é composta estritamente por bytes ASCII (`b < 0x80`).
- Quando uma string é ASCII pura, funções como `IndiceCaraterParaByte` retornam a posição diretamente em **O(1)**, ignorando decodificação UTF-8 ou consultas ao cache.

### Sharded LRU Cache (`ShardedCache`)
- **Arquitetura**: O cache global é particionado em 256 shards independentes com travamento individual via `sync.Mutex`.
- **Mapeamento**: O shard de cada chave é determinado inline via algoritmo FNV-1a hash (`hash % 256`), garantindo distribuição uniforme.
- **Evicção LRU**: Cada shard possui limite de entradas (com lista duplamente encadeada). Ao atingir a capacidade, remove somente a entrada menos recentemente utilizada do shard, sem travar o cache global.
- **API Exportada**:
  - `NovoShardedCache(numShards int, capPorShard int) *ShardedCache`
  - `ObterDoCache(key string) ([]int, bool)`
  - `ArmazenarNoCache(key string, val []int)`
  - `LimparCache()`

---

## 🔤 Análise e Navegação Unicode

As funções a seguir gerenciam a tradução transparente entre índices conceituais de caracteres Unicode (runas) e as posições físicas em bytes.

#### 1. `IndiceBytePorCarater(str string) []int`
Gera e pré-calcula a tabela de correspondências de offsets de bytes. Para strings de até 4096 bytes, consulta e atualiza automaticamente o `ShardedCache` global.

#### 2. `IndiceCaraterParaByte(str string, indice int, cache []int) int`
Resolve a posição de byte para o índice de caractere informado. Utiliza Fast-Path ASCII quando aplicável, consulta o cache fornecido ou recorre a `IndiceCaraterParaByteSemCache`.

#### 3. `IndiceCaraterParaByteSemCache(str string, indice int) int`
Realiza busca linear com proteção contra limites out-of-bounds e prevenção de loops infinitos quando `indice >= len(str)`.

#### 4. `ObtemCaraterPorIndice(str string, indice int, cache []int) string`
Retorna a string contendo exatamente um único caractere Unicode válido na posição informada.

---

## 🛡️ Validadores de Identificadores e Alfanuméricos

#### 1. `ContemApenasAlfaNum(str string) bool`
Inspeciona individualmente cada runa da string, garantindo que seja uma letra Unicode (`unicode.IsLetter`) ou dígito Unicode (`unicode.IsDigit`). Retorna `true` para cadeias mistas como `"usuario123"`.

#### 2. `EIdentificadorValido(str string) bool`
Valida se a string é um identificador legal na linguagem Harpia (não-vazia, iniciando com letra Unicode ou `_`, seguida por letras, dígitos ou `_`).

#### 3. `ContemApenasLetras(str string) bool`
Garante que todas as runas sejam letras Unicode.

#### 4. `ContemApenasDigitos(str string) bool`
Garante que todas as runas sejam dígitos decimais Unicode.

---

## 🔢 Utilitários de Conversão e Bases Numéricas

O arquivo `numeros.go` fornece parsing robusto e suporte a bases numéricas com precisão de 64 bits.

#### 1. `StringParaIntBase(s string) (int64, error)`
Converte strings numéricas identificando automaticamente o prefixo de base:
- Hexadecimal: `0x` ou `0X` (base 16)
- Binário: `0b` ou `0B` (base 2)
- Octal: `0o` ou `0O` (base 8)
- Decimal: sem prefixo (base 10)

#### 2. `StringParaInt(s string) (int64, error)`
Parsing de inteiros decimais em 64 bits (`int64`).

#### 3. `StringParaDec(s string) (float64, error)`
Parsing de números decimais/ponto flutuante IEEE 754 de dupla precisão (`float64`).

#### 4. `IntParaString(v int64) string` e `DecParaString(v float64) string`
Utilitários de formatação textual rápida.

---

## 🏗️ Arquitetura e Fluxo no Compilador

```
[Código Fonte .hrp] ➔ [Lexer / Parser] 
                        │
                        ├─► Usa EhASCII() / IndiceBytePorCarater() + ShardedCache
                        │
                        ├─► Usa EIdentificadorValido() / ContemApenasAlfaNum()
                        │
                        ▼
                 [Tokens Gerados] ➔ [VM / Runtime]
                                            │
                                            ├─► Usa StringParaIntBase() / StringParaDec()
                                            ▼
                                     [Execução Fullstack]
```

---

## 🛠️ Exemplos Práticos de Integração

```go
package main

import (
	"fmt"
	"github.com/natanfeitosa/harpia/compartilhado"
)

func main() {
	// 1. Fast-Path ASCII
	fmt.Println(compartilhado.EhASCII("Harpia 123")) // true

	// 2. Validação de identificadores válidos
	fmt.Println(compartilhado.EIdentificadorValido("_contador")) // true
	fmt.Println(compartilhado.EIdentificadorValido("123var"))    // false

	// 3. Conversão de bases numéricas (hex, bin, oct)
	hexVal, _ := compartilhado.StringParaIntBase("0x1A")
	binVal, _ := compartilhado.StringParaIntBase("0b1010")
	fmt.Printf("Hex 0x1A = %d, Bin 0b1010 = %d\n", hexVal, binVal) // 26, 10

	// 4. Mapeamento Unicode com cache Sharded
	texto := "Harpia 🚀"
	cache := compartilhado.IndiceBytePorCarater(texto)
	emoji := compartilhado.ObtemCaraterPorIndice(texto, 7, cache)
	fmt.Printf("Caractere no índice 7: %s\n", emoji) // 🚀
}
```
