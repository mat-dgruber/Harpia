package hrp

// _Nulo representa a estrutura interna de dados que materializa a ausência de valor na VM.
type _Nulo struct{}

var (
	// Nulo é a instância estática unificada global do tipo _Nulo.
	Nulo = _Nulo(struct{}{})

	// TipoNulo especifica os metadados de classe e tipo para Nulo.
	TipoNulo = NewTipo("Nulo", "Tipo que referencia a algo sem valor definido")
)

// Tipo retorna a representação de classe (Tipo de Nulo).
func (n _Nulo) Tipo() *Tipo {
	return TipoNulo
}
