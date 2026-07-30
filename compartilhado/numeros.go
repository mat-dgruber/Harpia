package compartilhado

import (
	"strconv"
	"strings"
)

// StringParaInt converte uma representação textual de número em um inteiro de 64 bits (int64).
//
// Parâmetros:
//   - s: literal textual do número (somente dígitos, com sinal opcional).
//
// Retorna:
//   - int64: valor numérico inteiro convertido em base 10.
//   - error: *strconv.NumError caso a string não seja um inteiro decimal válido
//     (contendo pontos, letras ou formatação inválida) ou caso ocorra overflow de 64 bits.
//
// Decisão de Design:
// A Harpia adota precisão de 64 bits em seus tipos numéricos para evitar estouro de
// limite (overflow) e garantir compatibilidade cruzada entre plataformas (x86/ARM).
func StringParaInt(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

// StringParaIntBase converte uma string numerica (decimal, hex 0x, binario 0b, octal 0o, com sinal opcional) em int64.
func StringParaIntBase(s string) (int64, error) {
	s = strings.TrimSpace(s)
	if len(s) == 0 {
		return 0, strconv.ErrSyntax
	}

	sinal := int64(1)
	if strings.HasPrefix(s, "-") {
		sinal = -1
		s = s[1:]
	} else if strings.HasPrefix(s, "+") {
		s = s[1:]
	}

	base := 10
	sLimpa := s

	if len(s) > 2 {
		prefixo := strings.ToLower(s[:2])
		switch prefixo {
		case "0x":
			base = 16
			sLimpa = s[2:]
		case "0b":
			base = 2
			sLimpa = s[2:]
		case "0o":
			base = 8
			sLimpa = s[2:]
		}
	}

	val, err := strconv.ParseInt(sLimpa, base, 64)
	if err != nil {
		return 0, err
	}
	return val * sinal, nil
}

// StringParaDec converte a representação textual de um número de ponto flutuante em float64.
//
// Parâmetros:
//   - s: literal textual contendo casas decimais separadas por ponto, com ou sem
//     notação científica (ex: "1.23", "1.23e-4").
//
// Retorna:
//   - float64: valor decimal convertido seguindo o padrão IEEE 754 de dupla precisão.
//   - error: *strconv.NumError quando o formato estiver fora dos padrões aceitos ou
//     o valor não couber na precisão float64.
//
// Aplicação:
// Fundamental para a correta tipagem e instanciação do tipo 'Decimal' da linguagem Harpia.
func StringParaDec(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}

// IntParaString converte um valor int64 em sua representacao textual.
func IntParaString(v int64) string {
	return strconv.FormatInt(v, 10)
}

// DecParaString converte um valor float64 em sua representacao textual.
func DecParaString(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}
