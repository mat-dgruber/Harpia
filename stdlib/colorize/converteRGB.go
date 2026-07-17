package colorize

import "github.com/mat-dgruber/Harpia/ptst"

// met_color_converteRGB implementa a lógica nativa para a função 'converteRGB()'.
//
// Esta função recebe três inteiros representando as cores Vermelho, Verde e Azul (RGB)
// e opcionalmente um valor booleano para 'background'. Ela valida e converte esses
// argumentos em uma sequência de escape ANSI de cores compatível com terminais.
func met_color_converteRGB(inst ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
	if err := ptst.VerificaNumeroArgumentos("converteRGB", false, args, 3, 4); err != nil {
		return nil, err
	}

	var vermelho, verde, azul = args[0], args[1], args[2]
	background := false

	if len(args) > 3 {
		background = bool(args[3].(ptst.Booleano))
	}

	return ptst.Texto(RgbParaAnsi(vermelho.(ptst.Inteiro), verde.(ptst.Inteiro), azul.(ptst.Inteiro), background)), nil
}

// _color_converteRGB cria e define a assinatura do método 'converteRGB' exposto no módulo colorize.
var _color_converteRGB = ptst.NewMetodoOuPanic(
	"converteRGB",
	met_color_converteRGB,
	"converteRGB(vermelho, verde, azul, background?) -> Retorna a cor em string no formato ANSI. Se um valor para `background` não for definido, ele será igual a Falso e retornará uma cor de foreground (texto)",
)
