package colorize

import "github.com/mat-dgruber/Harpia/hrp"

// met_color_converteRGB implementa a lógica nativa para a função 'converteRGB()'.
//
// Esta função recebe três inteiros representando as cores Vermelho, Verde e Azul (RGB)
// e opcionalmente um valor booleano para 'background'. Ela valida e converte esses
// argumentos em uma sequência de escape ANSI de cores compatível com terminais.
func met_color_converteRGB(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if err := hrp.VerificaNumeroArgumentos("converteRGB", false, args, 3, 4); err != nil {
		return nil, err
	}

	rObj, err := hrp.NewInteiro(args[0])
	if err != nil {
		return nil, err
	}
	gObj, err := hrp.NewInteiro(args[1])
	if err != nil {
		return nil, err
	}
	bObj, err := hrp.NewInteiro(args[2])
	if err != nil {
		return nil, err
	}

	rInt := int64(rObj.(hrp.Inteiro))
	gInt := int64(gObj.(hrp.Inteiro))
	bInt := int64(bObj.(hrp.Inteiro))

	if rInt < 0 || rInt > 255 || gInt < 0 || gInt > 255 || bInt < 0 || bInt > 255 {
		return nil, hrp.NewErroF(hrp.ValorErro, "valores RGB devem estar no intervalo [0, 255]")
	}

	background := false
	if len(args) > 3 {
		if bgBool, ok := args[3].(hrp.Booleano); ok {
			background = bool(bgBool)
		} else if b, err := hrp.NewBooleano(args[3]); err == nil {
			background = bool(b.(hrp.Booleano))
		}
	}

	return hrp.Texto(RgbParaAnsi(hrp.Inteiro(rInt), hrp.Inteiro(gInt), hrp.Inteiro(bInt), background)), nil
}

// _color_converteRGB cria e define a assinatura do método 'converteRGB' exposto no módulo colorize.
var _color_converteRGB = hrp.NewMetodoOuPanic(
	"converteRGB",
	met_color_converteRGB,
	"converteRGB(vermelho, verde, azul, background?) -> Retorna a cor em string no formato ANSI. Se um valor para `background` não for definido, ele será igual a Falso e retornará uma cor de foreground (texto)",
)
