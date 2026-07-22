package hrp

// ArgumentoNomeadoObj representa um argumento nomeado avaliado (ex: `nome = valor`) no runtime.
//
// Esta estrutura é gerada dinamicamente pelo interpretador ao avaliar parâmetros nomeados
// em chamadas de função, permitindo o empacotamento, transporte e mapeamento flexível na execução de closures.
type ArgumentoNomeadoObj struct {
	Nome  string // Nome de amarração lógica do argumento (ex: "porta").
	Valor Objeto // Objeto resolvido contendo o valor real associado ao argumento.
}

// TipoArgumentoNomeadoObj define os metadados de classe do tipo ArgumentoNomeadoObj na VM.
var TipoArgumentoNomeadoObj = NewTipo("ArgumentoNomeadoObj", "Estrutura interna de transporte de argumentos nomeados de funções.")

// Tipo satisfaz a interface primordial Objeto, retornando a metaclasse de ArgumentoNomeadoObj.
func (a *ArgumentoNomeadoObj) Tipo() *Tipo {
	return TipoArgumentoNomeadoObj
}

// Garante conformidade de assinatura estrutural com a interface Objeto.
var _ Objeto = (*ArgumentoNomeadoObj)(nil)
