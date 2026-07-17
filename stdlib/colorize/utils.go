package colorize

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/mat-dgruber/Harpia/ptst"
)

// cor define a estrutura de dados interna para mapeamento estático de cores hexadecimais.
type cor struct {
	Nome string
	Hex  string
}

// criaRenderizadorDeCores constrói dinamicamente uma função com a assinatura compatível do interpretador
// que atua como o renderizador de cores para as chamadas de método nos objetos Foreground e Background.
//
// Retorna uma closure que monta a string final adicionando a sequência ANSI de início de cor,
// concatena todos os argumentos fornecidos convertidos para texto e anexa o código ANSI de redefinição (Reset).
func criaRenderizadorDeCores(r, g, b ptst.Inteiro, background bool) func(inst ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
	codigo := InicioCodigo + RgbParaAnsi(r, g, b, background) + FimCodigo

	return func(inst ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
		saida := codigo
		for _, item := range args {
			itemObj, err := ptst.NewTexto(item)
			if err != nil {
				return nil, err
			}

			saida = fmt.Sprintf("%s%s", saida, itemObj)
		}

		return ptst.Texto(saida + ResetCodigo), nil
	}
}

// HexParaRgb faz o parsing de uma string contendo cor no formato Hexadecimal (CSS, ex: "#ff0000" ou "fff")
// e a decodifica de forma segura para os seus respectivos componentes de canal Vermelho, Verde e Azul (RGB)
// representados como ptst.Inteiro.
//
// Suporta variações comuns de entrada:
//   - Strings vazias (retorna erro)
//   - Prefixos "#" ou "0x" (removidos de forma transparente)
//   - Abreviações de 3 caracteres (ex: "ccc" é expandido internamente para "cccccc")
func HexParaRgb(hex string) (r, g, b ptst.Inteiro, err error) {
	hex = strings.TrimSpace(hex)
	if hex == "" {
		err = fmt.Errorf("O código hex não pode ser vazio")
		return
	}

	// Remove prefixos como de regras de CSS padrão
	if hex[0] == '#' {
		hex = hex[1:]
	}

	hex = strings.ToLower(hex)
	switch len(hex) {
	case 3: // "ccc" ➔ "cccccc"
		hex = string([]byte{hex[0], hex[0], hex[1], hex[1], hex[2], hex[2]})
	case 8: // "0xad99c0"
		hex = strings.TrimPrefix(hex, "0x")
	}

	// Valida se possui exatamente 6 caracteres após as limpezas
	if len(hex) != 6 {
		err = fmt.Errorf("O código '%s' não segue um formato válido de cor hex", hex)
		return
	}

	// Converte a string hexadecimal em um inteiro de 32 bits
	if i64, err := strconv.ParseInt(hex, 16, 32); err == nil {
		color := int(i64)
		// Realiza operações bitwise de máscara de bits (deslocamento) para obter r, g e b
		r = ptst.Inteiro(color >> 16)
		g = ptst.Inteiro((color & 0x00FF00) >> 8)
		b = ptst.Inteiro(color & 0x0000FF)
	}
	return
}

// RgbParaAnsi formata os três canais de cor e constrói a string numérica de controle ANSI correspondente
// (38;2;r;g;b para texto / 48;2;r;g;b para fundo).
func RgbParaAnsi(r, g, b ptst.Inteiro, background bool) string {
	if background {
		return fmt.Sprintf(TplBgRGB, r, g, b)
	}

	return fmt.Sprintf(TplFgRGB, r, g, b)
}