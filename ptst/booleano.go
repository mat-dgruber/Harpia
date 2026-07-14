package ptst

// Booleano representa a representação nativa de valores lógicos de verdadeiro ou falso no Portuscript.
// É um apelido (alias) para o tipo básico 'bool' do Go.
type Booleano bool

var (
	// TipoBooleano especifica a assinatura e os metadados de classe do tipo Booleano.
	TipoBooleano = NewTipo(
		"Booleano",
		"Verdadeiro ou Falso",
	)

	// Verdadeiro é a constante nativa para estados lógicos positivos.
	Verdadeiro = Booleano(true)

	// Falso é a constante nativa para estados lógicos negativos.
	Falso = Booleano(false)
)

// boolParaBooleano converte um bool standard de Go na constante estruturada Booleano correspondente.
func boolParaBooleano(obj bool) Booleano {
	if obj {
		return Verdadeiro
	}

	return Falso
}

// NewBooleano tenta forçar a coerção (casting) de qualquer objeto ou interface Go em um Booleano nativo.
//
// Regras de Resolução:
//   - Se o argumento implementar o método de conversão '__booleano__' (I__booleano__), delega a este método.
//   - Se for um bool puro do Go, chama boolParaBooleano.
//   - Fallback: Retorna Falso.
func NewBooleano(obj any) (Objeto, error) {
	switch b := obj.(type) {
	case I__booleano__:
		return b.M__booleano__()
	case bool:
		return boolParaBooleano(b), nil
	}

	return Falso, nil
}

func init() {
	// Nova define a função construtora para instanciação explícita de Booleanos em scripts Portuscript.
	TipoBooleano.Nova = func(args Tupla) (Objeto, error) {
		return NewBooleano(args[0])
	}
}

// Tipo retorna a representação de Tipo de Booleano para a VM.
func (b Booleano) Tipo() *Tipo {
	return TipoBooleano
}

// M__texto__ converte o valor lógico em sua respectiva representação de string ("Verdadeiro" ou "Falso").
func (b Booleano) M__texto__() (Objeto, error) {
	if b {
		return Texto("Verdadeiro"), nil
	}

	return Texto("Falso"), nil
}

// M__booleano__ satisfaz a interface de coerção booleana interna da VM, retornando a si mesmo.
func (b Booleano) M__booleano__() (Objeto, error) {
	return b, nil
}

// M__igual__ compara a igualdade lógica de valor entre dois booleanos.
func (b Booleano) M__igual__(a Objeto) (Objeto, error) {
	if !MesmoTipo(b, a) {
		return Falso, nil
	}

	return NewBooleano(b == a.(Booleano))
}

// M__diferente__ compara a desigualdade lógica de valor. Retorna Verdadeiro se forem de classes
// ou valores lógicos distintos.
func (b Booleano) M__diferente__(a Objeto) (Objeto, error) {
	if !MesmoTipo(b, a) {
		return Verdadeiro, nil
	}

	igual, err := b.M__igual__(a)
	if err != nil {
		return nil, err
	}

	return Nao(igual)
}

// M__decimal__ converte o booleano em Decimal (1.0 para Verdadeiro e 0.0 para Falso).
func (b Booleano) M__decimal__() (Objeto, error) {
	if b == Verdadeiro {
		return Decimal(1.0), nil
	}

	return Decimal(0), nil
}

// M__inteiro__ converte o booleano em Inteiro (1 para Verdadeiro e 0 para Falso).
func (b Booleano) M__inteiro__() (Objeto, error) {
	if b == Verdadeiro {
		return Inteiro(1), nil
	}

	return Inteiro(0), nil
}

// M__ou__ executa a operação lógica OU em nível de bits (Bitwise OR) entre operandos booleanos.
func (b Booleano) M__ou__(a Objeto) (Objeto, error) {
	if MesmoTipo(b, a) {
		var bi, ai Objeto
		var err error

		if bi, err = NewInteiro(b); err != nil {
			return nil, err
		}

		if ai, err = NewInteiro(a); err != nil {
			return nil, err
		}

		return NewBooleano(bi.(Inteiro) | ai.(Inteiro))
	}

	return NaoImplementado, nil
}

// M__e__ executa a operação lógica E em nível de bits (Bitwise AND) entre operandos booleanos.
func (b Booleano) M__e__(a Objeto) (Objeto, error) {
	if MesmoTipo(b, a) {
		var bi, ai Objeto
		var err error

		if bi, err = NewInteiro(b); err != nil {
			return nil, err
		}

		if ai, err = NewInteiro(a); err != nil {
			return nil, err
		}

		return NewBooleano(bi.(Inteiro) & ai.(Inteiro))
	}

	return NaoImplementado, nil
}

// Garantias de conformidade com os protocolos de interfaces nativos Go.
var _ I_conversaoEntreTipos = (*Booleano)(nil)
var _ I__igual__ = (*Booleano)(nil)
var _ I__diferente__ = (*Booleano)(nil)
var _ I_aritmeticaBooleana = (*Booleano)(nil)
