package ptst

// Tupla representa a coleção indexada ordenada e imutável de dados do Harpia (ex: (1, 2, 3)).
//
// É um apelido (alias) para um slice de Objetos Go (`[]Objeto`). Diferente da Lista, uma vez instanciada,
// a Tupla não expõe métodos mutáveis para alteração, adição ou deleção de dados físicos em tempo de execução.
type Tupla []Objeto

// TipoTupla especifica as assinaturas e metadados de classe do tipo Tupla na VM.
var TipoTupla = TipoObjeto.NewTipo(
	"Tupla",
	"Tupla(obj) -> Tupla",
)

// Tipo retorna a representação de classe (Tipo de Tupla).
func (t Tupla) Tipo() *Tipo {
	return TipoTupla
}

// GRepr é um gerador auxiliar de formatação de coleções textuais em colchetes ou parênteses.
func (t Tupla) GRepr(inicio, fim string) (Objeto, error) {
	junta, err := ObtemAtributoS(Texto(","), "junta")
	if err != nil {
		return nil, err
	}

	res, err := Chamar(junta, t)
	if err != nil {
		return nil, err
	}

	return (Texto(inicio) + res.(Texto) + Texto(fim)), nil
}

// M__iter__ satisfaz o protocolo de objetos iteráveis, retornando um iterador para varredura em loops.
func (t Tupla) M__iter__() (Objeto, error) {
	return NewIterador(t)
}

// M__texto__ converte os itens internos em sua representação string agregada, envolvida por parênteses ().
func (t Tupla) M__texto__() (Objeto, error) {
	return t.GRepr("(", ")")
}

// M__tamanho__ retorna a contagem de elementos na tupla.
func (t Tupla) M__tamanho__() (Objeto, error) {
	return Inteiro(len(t)), nil
}

// ObtemItem resolve o fatiamento e leitura de itens indexados. Lança erro de tipagem caso o índice não seja Inteiro.
func (t Tupla) ObtemItem(i Objeto, nomeTipo string) (Objeto, error) {
	if I, ok := i.(Inteiro); ok {
		return t[I], nil
	}

	return nil, NewErroF(TipagemErro, "O tipo '%s' não é aceito para indexação no tipo '%s'. Use um 'Inteiro'.", i.Tipo().Nome, nomeTipo)
}

// DefineItem executa a escrita no slice físico.
// Embora exposta como interface Go, o Harpia lança erros ou impede mutabilidades em scripts de usuário.
func (t Tupla) DefineItem(chave, valor Objeto, nomeTipo string) (Objeto, error) {
	if I, ok := chave.(Inteiro); ok {
		t[I] = valor
		return t, nil
	}

	return nil, NewErroF(TipagemErro, "O tipo '%s' não é aceito para indexação no tipo '%s'. Use um 'Inteiro'.", chave.Tipo().Nome, nomeTipo)
}

// M__obtem_item__ lê um item localizado no índice especificado.
func (t Tupla) M__obtem_item__(obj Objeto) (Objeto, error) {
	return t.ObtemItem(obj, t.Tipo().Nome)
}

// M__define_item__ tenta escrever dados na tupla.
func (t Tupla) M__define_item__(chave, valor Objeto) (Objeto, error) {
	return t.DefineItem(chave, valor, t.Tipo().Nome)
}

// M__contem__ verifica se o objeto reside na coleção.
func (t Tupla) M__contem__(obj Objeto) (Objeto, error) {
	for _, item := range t {
		if igual, _ := Igual(item, obj); igual == Verdadeiro {
			return Verdadeiro, nil
		}
	}

	return Falso, nil
}

// Garantias de conformidade com as interfaces nativas Go.
var _ I__iter__ = Tupla(nil)
var _ I__texto__ = Tupla(nil)
var _ I__tamanho__ = Tupla(nil)
var _ I__obtem_item__ = Tupla(nil)
var _ I__define_item__ = Tupla(nil)
var _ I__contem__ = Tupla(nil)
