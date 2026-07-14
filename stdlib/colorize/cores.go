package colorize

import (
	"fmt"

	"github.com/natanfeitosa/portuscript/ptst"
)

// Background representa o tipo de objeto nativo do Portuscript usado para colorir o fundo do texto.
type Background struct{}

// TipoBackground define as propriedades e métodos associados à estrutura Background no interpretador.
var TipoBackground = ptst.TipoObjeto.NewTipo(
	"Background",
	"",
)

// Tipo retorna a especificação de tipo da struct Background para a VM.
func (b *Background) Tipo() *ptst.Tipo {
	return TipoBackground
}

// Foreground representa o tipo de objeto nativo do Portuscript usado para colorir a frente (letra) do texto.
type Foreground struct{}

// TipoForeground define as propriedades e métodos associados à estrutura Foreground no interpretador.
var TipoForeground = ptst.TipoObjeto.NewTipo(
	"Foreground",
	"",
)

// Tipo retorna a especificação de tipo da struct Foreground para a VM.
func (f *Foreground) Tipo() *ptst.Tipo {
	return TipoForeground
}

// cores define uma lista estática de cores conhecidas mapeadas de CSS padrão.
// Estas cores são usadas para popular dinamicamente os métodos dos tipos Foreground e Background.
var cores = []*cor{
	{"vermelho", "ff0000"},
	{"lima", "00ff00"},
	{"azul", "0000ff"},
	{"amarelo", "ffff00"},
	{"agua", "00ffff"},
	{"fuchsia", "ff00ff"},
	{"branco", "fff"},
	{"preto", "000"},
}

func init() {
	// Popula dinamicamente os mapas de propriedades de TipoBackground e TipoForeground.
	// Cada entrada na lista de cores se torna um método chamável de conveniência no Portuscript.
	// Por exemplo, colorize.TEXTO.vermelho("Olá") retornará a string formatada em vermelho.
	for _, cor := range cores {
		r, g, b, err := HexParaRgb(cor.Hex)
		if err != nil {
			panic(err)
		}

		TipoBackground.Mapa[cor.Nome] = ptst.NewMetodoOuPanic(
			cor.Nome,
			criaRenderizadorDeCores(
				r, g, b,
				true,
			),
			fmt.Sprintf("Define a cor %s ao fundo do texto", cor.Nome),
		)

		TipoForeground.Mapa[cor.Nome] = ptst.NewMetodoOuPanic(
			cor.Nome,
			criaRenderizadorDeCores(
				r, g, b,
				false,
			),
			fmt.Sprintf("Define a cor %s ao texto", cor.Nome),
		)
	}
}