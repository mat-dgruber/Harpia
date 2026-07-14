package ptst

// MetodoProxy atua como o vinculador dinâmico de métodos (Bound Methods, semelhante ao do Python).
//
// Esta estrutura guarda a referência de uma instância ('Inst', correspondente ao self) e o 'Metodo'
// correspondente acessado. Ao ser acionada, ela intercepta a chamada, anexa e injeta a instância local
// como primeiro parâmetro ('self') de forma transparente antes de repassar a execução para a closure.
type MetodoProxy struct {
	Inst   Objeto // A instância do objeto à qual o método está associado (self).
	Metodo Objeto // O método ou função correspondente a ser invocado.
}

// TipoMetodoProxy especifica as assinaturas e metadados de classe do tipo MetodoProxy na VM.
var TipoMetodoProxy = NewTipo("MetodoProxy", "Um método vinculado a uma instância física que injeta a si mesma como primeiro argumento (self).")

// NewMetodoProxy aloca um novo vinculador de instância para o método correspondente.
func NewMetodoProxy(inst, metodo Objeto) *MetodoProxy {
	return &MetodoProxy{inst, metodo}
}

// Tipo retorna a representação de classe (Tipo de MetodoProxy).
func (mp *MetodoProxy) Tipo() *Tipo {
	return TipoMetodoProxy
}

// M__chame__ satisfaz o protocolo de chamabilidade da VM (I__chame__).
//
// Se o método for de classe Go nativa (*Metodo), invoca-o repassando a instância como o operando principal.
// Se for uma função de script (*Funcao), realiza o append do 'self' como o primeiro argumento da Tupla e aciona.
func (mp *MetodoProxy) M__chame__(args Tupla) (Objeto, error) {
	if m, ok := mp.Metodo.(*Metodo); ok {
		return m.Chamar(mp.Inst, args)
	}

	if f, ok := mp.Metodo.(*Funcao); ok {
		novaTupla := append(Tupla{mp.Inst}, args...)
		return f.M__chame__(novaTupla)
	}

	return Chamar(mp.Metodo, args)
}

// ObtemDoc repassa e devolve o docstring do método subjacente.
func (mp *MetodoProxy) ObtemDoc() string {
	return mp.Metodo.(*Metodo).ObtemDoc()
}

// Garantias de assinaturas estruturais em Go.
var _ I__chame__ = (*MetodoProxy)(nil)
var _ I_ObtemDoc = (*MetodoProxy)(nil)
