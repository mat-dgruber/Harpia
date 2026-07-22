package hrp

type Iterador struct {
	Posicao   int
	Conteiner Objeto
}

var TipoIterador = NewTipo("Iterador", "Objeto abstrato que representa um iterador nativo")

func NewIterador(seq Objeto) (*Iterador, error) {
	return &Iterador{Posicao: 0, Conteiner: seq}, nil
}

func (it *Iterador) Tipo() *Tipo {
	return TipoIterador
}

func (it *Iterador) M__iter__() (Objeto, error) {
	return it, nil
}

func (it *Iterador) M__proximo__() (Objeto, error) {
	switch c := it.Conteiner.(type) {
	case Tupla:
		// Tupla (e Lista — pois Lista.M__iter__ já passa l.Itens como Tupla)
		if it.Posicao >= len(c) {
			return nil, NewErro(FimIteracao, Nulo)
		}
		item := c[it.Posicao]
		it.Posicao++
		return item, nil

	case Texto:
		// Texto: itera sobre runas (caracteres Unicode), preservando multibyte.
		runas := []rune(string(c))
		if it.Posicao >= len(runas) {
			return nil, NewErro(FimIteracao, Nulo)
		}
		item := Texto(string(runas[it.Posicao]))
		it.Posicao++
		return item, nil

	case *Bytes:
		// Bytes: itera byte a byte, retornando cada byte como Inteiro.
		if it.Posicao >= len(c.Itens) {
			return nil, NewErro(FimIteracao, Nulo)
		}
		item := Inteiro(c.Itens[it.Posicao])
		it.Posicao++
		return item, nil
	}

	return nil, NewErroF(TipagemErro, "o tipo '%s' não é iterável", it.Conteiner.Tipo().Nome)
}

var _ I_iterador = (*Iterador)(nil)
