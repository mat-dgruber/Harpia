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
	// cacheMu protege o acesso simultâneo ao mapa cacheUnicode.
	// O uso de RWMutex permite leituras simultâneas por múltiplas goroutines,
	// bloqueando exclusivamente no momento de escrita de novas entradas.
	cacheMu      sync.RWMutex

	// cacheUnicode mapeia strings-fonte para sua tabela de offsets em bytes.
	// Evita recomputar repetidamente as posições de caracteres multibyte.
	cacheUnicode = make(map[string][]int)
)

// LimparCacheUnicode esvazia o cache estático global de índices de strings Unicode.
//
// Esta rotina libera memória ocupada por entradas previamente armazenadas. É indicada
// para uso ao final de uma compilação extensiva ou em ambientes que carregam e descartam
// scripts temporários em sequência.
func LimparCacheUnicode() {
	cacheMu.Lock()
	// ponytail: recriar o mapa é mais barato e imediato que percorrer todas as chaves para delete.
	cacheUnicode = make(map[string][]int)
	cacheMu.Unlock()
}

// IndiceBytePorCarater pré-calcula e mapeia a correspondência exata entre o índice sequencial de um
// caractere Unicode (runa) e a sua respectiva posição inicial em bytes (byte offset) dentro da string.
//
// Parâmetros:
//   - str: a string de entrada a ser mapeada (código-fonte do Harpia ou texto genérico).
//
// Retorna:
//   - []int: slice onde out[i] contém o byte offset da runa de índice conceitual i.
//     A última posição (out[RuneCount]) armazena len(str), permitindo operações
//     de fatiamento inclusivas (str[inicio:fim]) seguras no limite EOF.
//
// Decisão de Design / Por que isso existe:
// Em Go, strings são sequências de bytes UTF-8. Um único caractere Unicode (acentos, emojis,
// ideogramas) ocupa entre 1 e 4 bytes. Acessos aleatórios por índice em loops do lexer/parser
// seriam O(N) sem esta tabela, comprometendo o desempenho. O cache resultante reduz consultas
// subsequentes para O(1).
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
// Parâmetros:
//   - str: a string sob análise.
//   - indice: o índice conceitual (de runa) que se deseja resolver em bytes.
//   - cache: slice pré-calculado por IndiceBytePorCarater. Pode ser nil para modo sem cache.
//
// Retorna:
//   - int: o offset de byte referente ao caractere no índice. Se o índice for inválido
//     ou ultrapassar o limite, retorna len(str) (marca de EOF).
//
// Mecânica:
// Com cache válido a consulta é O(1). Sem cache, a função chama o caminho linear O(N).
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
// Parâmetros:
//   - str: a string sob análise.
//   - indice: o índice conceitual da runa alvo.
//
// Retorna:
//   - int: o byte offset correspondente após avançar sequencialmente pela decodificação de runas.
//
// Observação:
// Usada como fallback por IndiceCaraterParaByte quando nenhum cache é fornecido.
// ponytail: caminho linear simples; troque por tabela se um fluxo quente emergir.
func IndiceCaraterParaByteSemCache(str string, indice int) int {
	byteIndex := 0
	for i := 0; i < indice; i++ {
		_, tamanho := utf8.DecodeRuneInString(str[byteIndex:])
		byteIndex += tamanho
	}
	return byteIndex
}

// ObtemCaraterPorIndice extrai e devolve exatamente um único caractere Unicode presente no
// índice sequencial informado de uma string.
//
// Parâmetros:
//   - str: a string fonte completa.
//   - indice: a posição conceitual (runa) do caractere desejado.
//   - cache: slice pré-calculado contendo a tabela de byte offsets (opcional).
//
// Retorna:
//   - string: a substring de uma única runa localizada na posição solicitada.
//
// Segurança:
// A função determina os limites inicio/fim via IndiceCaraterParaByte e então executa
// str[inicio:fim]. Esse laço não pode causar panics de "slice bounds out of range", pois
// o fallback da função de índice retorna len(str) em limites inválidos.
func ObtemCaraterPorIndice(str string, indice int, cache []int) string {
	inicio := IndiceCaraterParaByte(str, indice, cache)
	fim := IndiceCaraterParaByte(str, indice+1, cache)
	return str[inicio:fim]
}

// ContemApenasLetras analisa se a totalidade da string informada é composta estritamente por letras Unicode.
//
// Parâmetros:
//   - str: a cadeia textual a ser inspecionada.
//
// Retorna:
//   - true caso a string não seja vazia e todas as runas sejam letras Unicode reconhecidas.
//   - false em caso de string vazia ou presença de espaços, símbolos, dígitos ou pontuações.
//
// Aplicação:
// Utilizada pelo Lexer para reconhecimento rápido de identificadores e palavras-chave,
// delegando a verificação para o pacote nativo `unicode` para cobrir alfabetos não latinos.
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

// ContemApenasDigitos analisa se todos os caracteres da string são dígitos numéricos decimais.
//
// Parâmetros:
//   - str: a string a ser validada como literal numérico.
//
// Retorna:
//   - true se todas as runas forem dígitos Unicode e a string não for vazia.
//   - false se a string for vazia ou contiver letras, espaços ou pontuações.
//
// Aplicação:
// Fundamental no fluxo do Lexer do Harpia para classificar eficientemente tokens numéricos
// inteiros, distinguindo-os de identificadores e outros símbolos.
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
