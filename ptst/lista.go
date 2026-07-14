package ptst

// Lista representa a coleção indexada mutável ordenada de dados do Portuscript (ex: [1, 2, 3]).
//
// Ela atua encapsulando uma estrutura interna do tipo Tupla, estendendo-a com capacidades de mutabilidade,
// inserção, deleção e redimensionamento dinâmico.
type Lista struct {
	Itens Tupla // O slice físico de dados subjacentes representados por Tupla.
}

// TipoLista especifica as assinaturas e metadados de classe do tipo Lista na VM.
var TipoLista = TipoObjeto.NewTipo(
	"Lista",
	"Lista(obj) -> Lista",
)

// Tipo retorna a representação de classe (Tipo de Lista).
func (l *Lista) Tipo() *Tipo {
	return TipoLista
}

// M__iter__ satisfaz o protocolo de objetos iteráveis, retornando um iterador encapsulado sobre a lista.
func (l *Lista) M__iter__() (Objeto, error) {
	return NewIterador(l.Itens)
}

// M__texto__ converte os itens internos em sua representação string agregada, envolvida por colchetes [].
func (l *Lista) M__texto__() (Objeto, error) {
	return l.Itens.GRepr("[", "]")
}

// M__tamanho__ retorna a contagem de elementos atualmente armazenados na lista.
func (l *Lista) M__tamanho__() (Objeto, error) {
	return l.Itens.M__tamanho__()
}

// M__obtem_item__ lê e resolve o item localizado no índice numérico fornecido. Lança erro se o índice estiver fora dos limites.
func (l *Lista) M__obtem_item__(obj Objeto) (Objeto, error) {
	return l.Itens.ObtemItem(obj, "Lista")
}

// M__define_item__ executa a atribuição/escrita mutável de um valor em um índice existente da lista (ex: lista[0] = "valor").
func (l *Lista) M__define_item__(chave, valor Objeto) (Objeto, error) {
	if _, err := l.Itens.DefineItem(chave, valor, l.Tipo().Nome); err != nil {
		return nil, err
	}

	return l, nil
}

// M__contem__ verifica o pertencimento. Retorna Verdadeiro se o objeto existir na lista, e Falso caso contrário.
func (l *Lista) M__contem__(obj Objeto) (Objeto, error) {
	return l.Itens.M__contem__(obj)
}

// Garantias de assinaturas estruturais em Go.
var _ I__iter__ = (*Lista)(nil)
var _ I__texto__ = (*Lista)(nil)
var _ I__tamanho__ = (*Lista)(nil)
var _ I__obtem_item__ = (*Lista)(nil)
var _ I__define_item__ = (*Lista)(nil)
var _ I__contem__ = (*Lista)(nil)

// Adiciona insere um novo objeto ao término da lista (operação push/append mutável).
func (l *Lista) Adiciona(item Objeto) (Objeto, error) {
	l.Itens = append(l.Itens, item)
	return nil, nil
}

// Indice busca e retorna o índice numérico da primeira ocorrência do objeto na lista.
// Lança erro estruturado ValorErro se o item não estiver contido.
func (l *Lista) Indice(obj Objeto) (Objeto, error) {
	for indice, item := range l.Itens {
		if ok, _ := Igual(item, obj); ok.(Booleano) {
			return Inteiro(indice), nil
		}
	}

	objTexto, err := NewTexto(obj)
	if err != nil {
		return nil, err
	}

	return nil, NewErroF(ValorErro, "O item '%s' não está na lista", objTexto)
}

// Pop extrai e remove o item localizado na posição do índice informado, retornando-o e reorganizando o slice físico.
// Lança erro estruturado IndiceErro se o índice estiver fora do intervalo permitido.
func (l *Lista) Pop(indice Inteiro) (Objeto, error) {
	tamanho, err := l.M__tamanho__()
	if err != nil {
		return nil, err
	}

	if indice > tamanho.(Inteiro) || indice < 0 {
		return nil, NewErroF(IndiceErro, "O range é de %d indice(s), %d está fora dele", tamanho.(Inteiro), indice)
	}

	var removido Objeto
	var novaTupla Tupla

	for idx, item := range l.Itens {
		if idx == int(indice) {
			removido = item
			continue
		}

		novaTupla = append(novaTupla, item)
	}

	l.Itens = novaTupla
	return removido, nil
}

func init() {
	// Registro de métodos de classe e métodos mágicos de Lista.

	TipoLista.Mapa["adiciona"] = NewMetodoOuPanic("adiciona", func(inst Objeto, args Tupla) (Objeto, error) {
		if err := VerificaNumeroArgumentos("adiciona", true, args, 1, 1); err != nil {
			return nil, err
		}

		inst.(*Lista).Adiciona(args[0])
		return nil, nil
	}, "O método recebe um objeto e adiciona ao fim da lista")

	TipoLista.Mapa["extende"] = NewMetodoOuPanic("extende", func(inst Objeto, args Tupla) (Objeto, error) {
		if err := VerificaNumeroArgumentos("extende", true, args, 1, 1); err != nil {
			return nil, err
		}

		inst.(*Lista).Itens = append(inst.(*Lista).Itens, (args[0].(Tupla))...)
		return nil, nil
	}, "Adiciona os elementos da lista recebida ao fim da lista atual")

	TipoLista.Mapa["remove"] = NewMetodoOuPanic("remove", func(inst Objeto, args Tupla) (Objeto, error) {
		if err := VerificaNumeroArgumentos("remove", true, args, 1, 1); err != nil {
			return nil, err
		}

		instancia := inst.(*Lista)
		idx, err := instancia.Indice(args[0])
		if err != nil {
			return nil, err
		}

		return instancia.Pop(idx.(Inteiro))
	}, "Remove um elemento da lista e o retorna, se existir")

	TipoLista.Mapa["pop"] = NewMetodoOuPanic("pop", func(inst Objeto, args Tupla) (Objeto, error) {
		if err := VerificaNumeroArgumentos("pop", true, args, 0, 1); err != nil {
			return nil, err
		}

		idx := Inteiro(0)

		if len(args) == 1 {
			idx = args[0].(Inteiro)
		}

		return inst.(*Lista).Pop(idx)
	}, "Remove um item da lista com base no seu índice")

	TipoLista.Mapa["indice"] = NewMetodoOuPanic("indice", func(inst Objeto, args Tupla) (Objeto, error) {
		if err := VerificaNumeroArgumentos("indice", true, args, 1, 1); err != nil {
			return nil, err
		}

		return inst.(*Lista).Indice(args[0])
	}, "Retorna o índice de um elemento se ele existir na lista")

	TipoLista.Mapa["limpa"] = NewMetodoOuPanic("limpa", func(inst Objeto) (Objeto, error) {
		inst.(*Lista).Itens = Tupla(nil)
		return nil, nil
	}, "Limpa completamente a lista")
}
