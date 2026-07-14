package colorize

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/natanfeitosa/portuscript/ptst"
)

// regexCor é uma expressão regular usada para identificar e capturar sequências de escape ANSI cromáticas
// de forma a poder limpá-las caso o terminal hospedeiro não possua suporte a cores.
var regexCor = regexp.MustCompile(`\033\[[\d;?]+m`)

// met_color_imprimac implementa a lógica nativa para a função 'imprimac()'.
//
// Esta função opera de forma similar à função embutida 'imprima()', porém é otimizada para trabalhar
// com strings que contenham estilizações de cores do módulo colorize.
//
// Regra de Negócio:
// Se o terminal não suportar cores (ex: variável 'NO_COLOR' definida no ambiente), esta função
// intercepta a string final unificada, executa uma varredura via expressão regular regexCor
// e remove todas as sequências de escape ANSI de estilo e cores de forma limpa antes de realizar a impressão.
// Isso evita que códigos de escape sejam impressos como texto ilegível para o usuário em terminais limitados.
func met_color_imprimac(inst ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
	if len(args) == 0 {
		return nil, ptst.NewErro(ptst.TipagemErro, ptst.Texto("A função imprimac esperava receber ao menos 1 argumento"))
	}

	var junta, textoObj ptst.Objeto
	var err error

	// Obtém o método embutido 'junta' do tipo Texto para concatenar todos os argumentos com espaço.
	if junta, err = ptst.ObtemAtributoS(ptst.Texto(""), "junta"); err != nil {
		return nil, err
	}

	if textoObj, err = ptst.Chamar(junta, args); err != nil {
		return nil, err
	}

	saida := string(textoObj.(ptst.Texto))

	// Se o terminal não tiver suporte de cores ativo, remove as sequências de escapes cromáticos ANSI.
	if !SuportaCores && strings.Contains(saida, InicioCodigo) {
		saida = regexCor.ReplaceAllString(saida, "")
	}

	fmt.Println(saida)
	return nil, nil
}

// _color_imprimac cria e define a assinatura do método 'imprimac' exposto no módulo colorize.
var _color_imprimac = ptst.NewMetodoOuPanic(
	"imprimac",
	met_color_imprimac,
	"imprimac(...objeto) -> O mesmo que a função embutida imprima, porém mais apto a trabalhar com as cores",
)
