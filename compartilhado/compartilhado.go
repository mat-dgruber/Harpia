package compartilhado

// ContemApenasAlfaNum analisa se a string informada é estritamente alfanumérica.
//
// Parâmetros:
//   - str: a cadeia textual a ser avaliada.
//
// Retorna:
//   - true se a string contiver apenas dígitos Unicode (0-9) OU apenas letras Unicode
//     (incluindo caracteres acentuados de qualquer alfabeto suportado pelo padrão Unicode).
//   - false se a string for vazia ou contiver qualquer outro tipo de caractere (símbolos,
//     espaços, pontuações ou mistura de letras e números).
//
// Decisão de Design:
// A função delega a lógica para ContemApenasDigitos e ContemApenasLetras para garantir
// reaproveitamento de código e consistência nas regras de classificação do Lexer do Harpia,
// evitando caminhos de validação divergentes.
func ContemApenasAlfaNum(str string) bool {
	return ContemApenasDigitos(str) || ContemApenasLetras(str)
}
