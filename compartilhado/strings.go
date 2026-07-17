// Package compartilhado reúne funções utilitárias e rotinas auxiliares comuns que são consumidas
// por múltiplos subsistemas do Harpia (como o lexer, parser, compilador e o próprio interpretador).
//
// O pacote serve como uma caixa de ferramentas desacoplada para evitar duplicação de lógica,
// com foco especial na manipulação eficiente de strings Unicode (UTF-8) e conversões numéricas.
package compartilhado

import (
	"sync"
	"unicode"
	"unicode/utf8"
)

var (
	cacheMu      sync.RWMutex
	cacheUnicode = make(map[string][]int)
)

// LimparCacheUnicode esvazia de forma síncrona o cache estático global de índices de strings Unicode.
func LimparCacheUnicode() {
	cacheMu.Lock()
	cacheUnicode = make(map[string][]int)
	cacheMu.Unlock()
}

// IndiceBytePorCarater pré-calcula e mapeia a correspondência exata entre o índice sequencial de um
// caractere Unicode (runa) e a sua respectiva posição inicial em bytes (byte offset) dentro da string.
//
// Retorna um array de inteiros onde out[indice_caractere] = offset em bytes da runa.
// O array inclui uma entrada adicional na posição final (out[rune_count] == len(str)) para satisfazer
// leituras inclusivas de limites superior (ex: fatiamento de strings e verificação de fim de arquivo).
//
// Decisão de Design / Por que isso existe:
// Em Go, as strings são sequências de bytes formatados em UTF-8. Um único caractere Unicode
// (como acentos ou emojis) pode ocupar entre 1 e 4 bytes. Fazer acessos aleatórios frequentes baseados no índice
// conceitual de caractere em loops de lexer/parser seria uma operação O(N) muito custosa. Este mapeamento
// atua como uma tabela de consulta rápida (cache) que reduz futuros acessos ao arquivo ou string para O(1).
func IndiceBytePorCarater(str string) []int {
	// Se a string for muito longa, evitamos colocar no cache para não gastar memória excessiva.
	usarCache := len(str) < 4096

	if usarCache {
		cacheMu.RLock()
		if c, ok := cacheUnicode[str]; ok {
			cacheMu.RUnlock()
			return c
		}
		cacheMu.RUnlock()
	}

	out := make([]int, 0, len(str)+2)
	byteIdx := 0
	for byteIdx < len(str) {
		out = append(out, byteIdx)
		_, tamanho := utf8.DecodeRuneInString(str[byteIdx:])
		byteIdx += tamanho
	}
	out = append(out, len(str))

	if usarCache {
		cacheMu.Lock()
		// Evita o crescimento infinito e vazamento de memória da tabela hash do cache
		if len(cacheUnicode) > 2048 {
			cacheUnicode = make(map[string][]int)
		}
		cacheUnicode[str] = out
		cacheMu.Unlock()
	}

	return out
}

// IndiceCaraterParaByte resolve o offset de byte correspondente ao índice conceitual do caractere informado.
//
// Se um cache pré-calculado (gerado previamente por IndiceBytePorCarater) for fornecido, a conversão
// é realizada instantaneamente em tempo constante O(1). Caso contrário, a função executa uma busca
// sequencial incremental baseada na decodificação de runas, degradando graciosamente para O(N) como fallback.
func IndiceCaraterParaByte(str string, indice int, cache []int) int {
	if cache != nil {
		if indice >= 0 && indice < len(cache) {
			return cache[indice]
		}
		return len(str)
	}
	return IndiceCaraterParaByteSemCache(str, indice)
}

// IndiceCaraterParaByteSemCache calcula o offset de bytes de um caractere sem o uso de cache (busca O(N)).
//
// Esta função percorre sequencialmente os caracteres da string decodificando runas uma a uma
// e acumulando os tamanhos em bytes até atingir o índice conceitual desejado. É chamada como fallback
// por IndiceCaraterParaByte.
func IndiceCaraterParaByteSemCache(str string, indice int) int {
	byteIndex := 0
	for i := 0; i < indice; i++ {
		_, tamanho := utf8.DecodeRuneInString(str[byteIndex:])
		byteIndex += tamanho
	}
	return byteIndex
}

// ObtemCaraterPorIndice extrai e devolve a substring exata correspondente a um único caractere Unicode
// localizado no índice sequencial especificado na string fornecida.
//
// Utiliza de forma inteligente a resolução de bytes com cache (via IndiceCaraterParaByte) para delimitar as posições
// de início e fim da fatia em bytes (slice) e extraí-la de forma segura e rápida sem quebrar caracteres multi-byte.
func ObtemCaraterPorIndice(str string, indice int, cache []int) string {
	inicio := IndiceCaraterParaByte(str, indice, cache)
	fim := IndiceCaraterParaByte(str, indice+1, cache)
	return str[inicio:fim]
}

// ContemApenasLetras analisa se a totalidade da string informada é composta estritamente por letras Unicode.
//
// Retorna false se a string for vazia ou se possuir qualquer caractere que não seja considerado
// uma letra válida segundo a tabela Unicode (como espaços, pontuações, dígitos ou caracteres de controle).
//
// Utilizado pelo Lexer para otimizar a identificação inicial de tokens e identificadores de 1 caractere,
// mantendo a capacidade genérica para validação de termos maiores.
func ContemApenasLetras(str string) bool {
	if str == "" {
		return false
	}
	for _, char := range str {
		if !unicode.IsLetter(char) {
			return false
		}
	}
	return true
}

// ContemApenasDigitos analisa se todos os caracteres presentes na string são dígitos numéricos decimais.
//
// Retorna false se a string for vazia ou se possuir qualquer caractere não classificado como dígito
// de acordo com a tabela de classificação Unicode (incluindo letras, acentos, pontuação ou espaços).
//
// Essencial para o reconhecimento rápido e classificação sintática de números inteiros pelo Lexer do Harpia.
func ContemApenasDigitos(str string) bool {
	if str == "" {
		return false
	}
	for _, char := range str {
		if !unicode.IsDigit(char) {
			return false
		}
	}
	return true
}
