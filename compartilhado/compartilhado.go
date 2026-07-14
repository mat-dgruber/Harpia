package compartilhado

// ContemApenasAlfaNum analisa se a string informada é estritamente alfanumérica.
//
// Esta função é uma composição lógica conveniente que retorna true se a string
// contiver apenas dígitos Unicode (0-9) OU apenas letras Unicode (incluindo caracteres acentuados
// de qualquer alfabeto suportado pelo padrão Unicode).
//
// Decisão de Design:
// A função delega a lógica para ContemApenasDigitos e ContemApenasLetras para garantir
// reaproveitamento de código e consistência nas regras de classificação do Lexer do Portuscript,
// evitando caminhos de validação divergentes.
func ContemApenasAlfaNum(str string) bool {
	return ContemApenasDigitos(str) || ContemApenasLetras(str)
}