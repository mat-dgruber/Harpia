package compartilhado

import "strconv"

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
