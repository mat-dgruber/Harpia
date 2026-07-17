// Package playground implementa a Interface de Usuário de Terminal (TUI) e o ambiente de REPL
// (Read-Eval-Print Loop) interativo do Harpia.
//
// O pacote é responsável por gerenciar a entrada histórica de comandos, renderizar o prompt,
// controlar o buffer de escrita do desenvolvedor em tempo de digitação, compilar e avaliar os
// trechos de código executados linha a linha em um escopo isolado de REPL.
package playground

import (
	"strings"
)

// indicador define a string visual exibida no início de cada linha do terminal REPL.
type indicador string

const (
	// Normal é o prompt de entrada principal para novos comandos de linha única (">>> ").
	Normal   indicador = ">>> "

	// Continua é o prompt secundário ("... ") que indica que o comando anterior possui
	// delimitadores abertos (ex: parênteses, colchetes ou chaves) e requer continuação na próxima linha.
	Continua indicador = "... "
)

// Estado armazena e acompanha as informações contextuais sobre a entrada corrente do REPL.
type Estado struct {
	// Indicador é o prompt visual corrente (Normal ou Continua).
	Indicador indicador

	// Continua é uma flag que sinaliza se a VM deve continuar lendo mais linhas de código do terminal
	// antes de enviar o buffer acumulado para a compilação e execução física.
	Continua bool

	// Codigo armazena a cadeia de texto (buffer) acumulada pelo usuário ao longo de múltiplas linhas.
	Codigo string
}

// NewEstado é o construtor padrão para a estrutura Estado do playground, inicializando-a em modo normal.
func NewEstado() *Estado {
	return &Estado{Normal, false, ""}
}

// RecalcularEstado atualiza o buffer de código com o novo trecho fornecido pelo usuário e analisa se
// a expressão está sintaticamente completa ou se possui delimitadores não pareados (abertos).
//
// Regra de Negócio:
// A função varre o buffer em busca de colchetes, parênteses e chaves. Se o número de símbolos de abertura
// for estritamente maior do que o de fechamento, a flag 'Continua' é ativada e o indicador visual
// muda para "... ", sinalizando ao REPL que a instrução lógica continua em uma nova linha.
// Assim que todos os pares são devidamente fechados, o estado retorna ao modo Normal e o acumulado
// fica elegível para execução pela VM.
func (e *Estado) RecalcularEstado(cod string) {
	e.Codigo += cod
	continua := e.continuaEmNovaLinha("[", "]") || e.continuaEmNovaLinha("(", ")") || e.continuaEmNovaLinha("{", "}")

	if continua == e.Continua {
		return
	}

	e.Continua = continua
	if e.Continua {
		e.Indicador = Continua
		return
	}

	e.Indicador = Normal
}

// continuaEmNovaLinha faz uma comparação simples e rápida usando strings.Count para verificar se o delimitador
// de abertura ('abre') possui mais ocorrências ativas do que o delimitador correspondente de fechamento ('fecha').
func (e *Estado) continuaEmNovaLinha(abre, fecha string) bool {
	return strings.Count(e.Codigo, abre) > strings.Count(e.Codigo, fecha)
}