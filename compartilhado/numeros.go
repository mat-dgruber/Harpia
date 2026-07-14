package compartilhado

import "strconv"

// StringParaInt converte uma representação textual de número em um inteiro de 64 bits (int64).
//
// A conversão utiliza a base decimal (base 10) e garante o tamanho máximo de precisão em 64 bits
// para evitar estouro de pilha (overflow) ou perda de sinal na representação dos números
// dentro da Máquina Virtual do Portuscript.
//
// Retorna o valor inteiro decodificado ou um erro do tipo *strconv.NumError caso a string
// não seja um número inteiro válido.
func StringParaInt(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

// StringParaDec converte a representação textual de um número de ponto flutuante em float64 (decimais).
//
// Esta função é fundamental para a correta tipagem do tipo 'Decimal' do Portuscript.
// Ela interpreta strings que possuem representação exponencial (científica) ou casas decimais
// separadas por ponto (padrão IEEE 754 de dupla precisão).
//
// Retorna o valor float64 decodificado ou um erro do tipo *strconv.NumError se a formatação
// textual estiver fora dos padrões aceitos.
func StringParaDec(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}
