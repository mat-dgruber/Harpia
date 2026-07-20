package hrp

// ArgumentoNomeadoObj representa um argumento nomeado avaliado (nome = valor) em runtime.
type ArgumentoNomeadoObj struct {
	Nome  string
	Valor Objeto
}

var TipoArgumentoNomeadoObj = NewTipo("ArgumentoNomeadoObj", "Argumento nomeado avaliado")

func (a *ArgumentoNomeadoObj) Tipo() *Tipo {
	return TipoArgumentoNomeadoObj
}

var _ Objeto = (*ArgumentoNomeadoObj)(nil)
