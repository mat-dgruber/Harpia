package hrp

// Iterador é o cursor de percurso unificado para loops no Harpia.
//
// Ele armazena o índice da posição corrente do percurso e uma referência para o contêiner
// de dados que está sendo iterado (como sequências, tuplas, textos ou bytes).
type Iterador struct {
	Posicao   int    // Índice do ponteiro de leitura do cursor atual de percurso.
	Conteiner Objeto // Referência ao objeto colecionável subjacente (ex: Tupla, Texto, Bytes).
}

// TipoIterador define os metadados de classe do tipo Iterador na máquina virtual.
var TipoIterador = NewTipo("Iterador", "Objeto cursor de percurso para sequências e coleções.")

// NewIterador aloca e inicializa uma nova instância de Iterador apontando para o início da coleção.
func NewIterador(seq Objeto) (*Iterador, error) {
	return &Iterador{Posicao: 0, Conteiner: seq}, nil
}

// Tipo satisfaz a interface Objeto, devolvendo o ponteiro de metadados da classe Iterador.
func (it *Iterador) Tipo() *Tipo {
	return TipoIterador
}

// M__iter__ satisfaz o protocolo de loops (I__iter__), retornando a si próprio como iterador ativo.
func (it *Iterador) M__iter__() (Objeto, error) {
	return it, nil
}

// M__proximo__ avança o cursor e retorna o próximo item disponível na coleção.
//
// Esta função é o coração do loop "para" e "enquanto" no Harpia. Ela analisa dinamicamente
// o tipo do contêiner e aplica o avanço do cursor:
//   - Tupla (e Lista): Retorna o item no índice atual. Dispara FimIteracao se estourar o limite.
//   - Texto: Itera de forma segura sobre runas Unicode (multibyte), prevenindo quebra de caracteres.
//   - Bytes: Retorna o valor de cada byte codificado como Inteiro.
//
// Quando a coleção é totalmente percorrida, dispara o erro especial de controle FimIteracao.
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
		// Texto: itera sobre runas (caracteres Unicode), preservando multibyte de forma segura.
		runas := []rune(string(c))
		if it.Posicao >= len(runas) {
			return nil, NewErro(FimIteracao, Nulo)
		}
		item := Texto(string(runas[it.Posicao]))
		it.Posicao++
		return item, nil

	case *Bytes:
		// Bytes: itera byte a byte, retornando cada byte individual como Inteiro de 64 bits.
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
