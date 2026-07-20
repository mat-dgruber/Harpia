// Package colorize implementa suporte nativo para estilização e coloração de saídas de texto
// no console (terminal) utilizando sequências de escape ANSI de 24 bits (True Color).
//
// O pacote define objetos Foreground (texto) e Background (fundo) cujas propriedades de cores
// são mapeadas de forma dinâmica e expostas ao interpretador do Harpia.
package colorize

import (
	"os"

	"github.com/mat-dgruber/Harpia/hrp"
)

const (
	// InicioCodigo é o cabeçalho padrão de escape ANSI para início de sequência cromática.
	InicioCodigo = "\x1b["

	// FimCodigo finaliza o cabeçalho de modo de escape de renderização de cor.
	FimCodigo = "m"

	// ResetCodigo limpa todos os estilos e cores ativos no console, retornando ao padrão.
	ResetCodigo = "\x1b[0m"

	// TplFgRGB é o template ANSI para cor de primeiro plano (Foreground) em formato RGB de 24 bits.
	TplFgRGB = "38;2;%d;%d;%d"

	// TplBgRGB é o template ANSI para cor de plano de fundo (Background) em formato RGB de 24 bits.
	TplBgRGB = "48;2;%d;%d;%d"
)

// SuportaCores determina de forma reativa se o ambiente atual aceita coloração.
// Segue a especificação padrão da iniciativa "no-color.org", desabilitando cores se a variável
// de ambiente 'NO_COLOR' estiver presente e não vazia.
var SuportaCores = os.Getenv("NO_COLOR") == ""

func init() {
	constantes := hrp.Mapa{
		// FUNDO é uma instância do tipo Background, permitindo invocar cores de fundo (ex: colorize.FUNDO.vermelho("texto")).
		"FUNDO": &Background{},

		// TEXTO é uma instância do tipo Foreground, permitindo colorir o texto (ex: colorize.TEXTO.azul("texto")).
		"TEXTO": &Foreground{},

		// SUPORTA é uma propriedade booleana indicando se o console aceita estilização colorida.
		"SUPORTA": hrp.Booleano(SuportaCores),
	}

	metodos := []*hrp.Metodo{
		_color_converteRGB,
		_color_imprimac,
	}

	// Registra o módulo 'colorize' nativamente na VM do Harpia.
	hrp.RegistraModuloImpl(
		&hrp.ModuloImpl{
			Info: hrp.ModuloInfo{
				Nome:    "colorize",
				Arquivo: "stdlib/colorize",
			},
			Constantes: constantes,
			Metodos:    metodos,
		},
	)
}
