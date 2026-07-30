// Package compartilhado reúne funções utilitárias e rotinas auxiliares comuns que são consumidas
// por múltiplos subsistemas do Harpia (como o lexer, parser, compilador e o próprio interpretador).
//
// O pacote serve como uma caixa de ferramentas desacoplada para evitar duplicação de lógica,
// com foco especial na manipulação eficiente de strings Unicode (UTF-8) e conversões numéricas.
package compartilhado

import (
	"unicode"
	"unicode/utf8"
)

// EhASCII verifica se a string é composta estritamente por caracteres ASCII (bytes < 0x80).
func EhASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] >= 0x80 {
			return false
		}
	}
	return true
}

// LimparCacheUnicode esvazia o cache de índices Unicode.
func LimparCacheUnicode() {
	globalCache.Limpar()
}

// IndiceBytePorCarater mapeia o byte offset de cada runa na string.
func IndiceBytePorCarater(str string) []int {
	if str == "" {
		return []int{0}
	}

	usarCache := len(str) < 4096
	if usarCache {
		if c, ok := globalCache.Obter(str); ok {
			return c
		}
	}

	out := make([]int, 0, utf8.RuneCountInString(str)+1)
	byteIdx := 0
	for byteIdx < len(str) {
		out = append(out, byteIdx)
		_, tamanho := utf8.DecodeRuneInString(str[byteIdx:])
		if tamanho == 0 {
			break
		}
		byteIdx += tamanho
	}
	out = append(out, len(str))

	if usarCache {
		globalCache.Armazenar(str, out)
	}

	return out
}

// IndiceCaraterParaByte resolve a posição de byte correspondente ao índice conceitual de caractere.
func IndiceCaraterParaByte(str string, indice int, cache []int) int {
	if indice <= 0 {
		return 0
	}
	if EhASCII(str) {
		if indice >= len(str) {
			return len(str)
		}
		return indice
	}
	if cache != nil {
		if indice < len(cache) {
			return cache[indice]
		}
		return len(str)
	}
	return IndiceCaraterParaByteSemCache(str, indice)
}

// IndiceCaraterParaByteSemCache calcula o byte offset sem o uso de cache.
func IndiceCaraterParaByteSemCache(str string, indice int) int {
	if indice <= 0 {
		return 0
	}
	byteIndex := 0
	for i := 0; i < indice; i++ {
		if byteIndex >= len(str) {
			return len(str)
		}
		_, tamanho := utf8.DecodeRuneInString(str[byteIndex:])
		if tamanho == 0 {
			return len(str)
		}
		byteIndex += tamanho
	}
	return byteIndex
}

// ObtemCaraterPorIndice extrai o caractere Unicode no índice especificado.
func ObtemCaraterPorIndice(str string, indice int, cache []int) string {
	inicio := IndiceCaraterParaByte(str, indice, cache)
	fim := IndiceCaraterParaByte(str, indice+1, cache)
	if inicio >= len(str) || inicio >= fim {
		return ""
	}
	return str[inicio:fim]
}

// ContemApenasLetras verifica se todas as runas da string são letras Unicode.
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

// ContemApenasDigitos verifica se todas as runas da string são dígitos decimais Unicode.
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
