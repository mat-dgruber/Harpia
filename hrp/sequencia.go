package hrp

// próximo é o wrapper de alto nível para invocar o método mágico '__próximo__' de um iterador.
//
// Esta função encapsula a asserção de tipo e chamada de M__próximo__(), padronizando e
// centralizando a captura de erros de controle como FimIteracao para as estruturas de controle da VM.
func próximo(obj Objeto) (Objeto, error) {
	if iter, ok := obj.(I__próximo__); ok {
		return iter.M__próximo__()
	}

	return nil, NewErroF(TipagemErro, "O objeto do tipo '%s' não implementa a interface do iterador (método __próximo__)", obj.Tipo().Nome)
}

// Iter é o wrapper de alto nível para obter o iterador ativo de um objeto colecionável.
//
// Ela resolve e aciona o protocolo mágico '__iter__', obtendo a estrutura de cursor (Iterador)
// que permite que a instrução "para" percorra coleções de forma segura e padronizada.
func Iter(obj Objeto) (Objeto, error) {
	if iter, ok := obj.(I__iter__); ok {
		return iter.M__iter__()
	}

	return nil, NewErroF(TipagemErro, "O objeto do tipo '%s' não implementa a interface de coleção iterável (método __iter__)", obj.Tipo().Nome)
}

// Em avalia a pertinência lógica de um elemento dentro de um contêiner através do operador "em".
//
// Esta rotina resolve o operador "em" (ex: "valor em lista") acionando o método mágico '__contem__' (M__contem__).
// Garante o isolamento contra erros de tipagem caso o contêiner não suporte buscas de pertinência.
func Em(a, b Objeto) (Objeto, error) {
	if A, ok := a.(I__contem__); ok {
		res, err := A.M__contem__(b)

		if err != nil {
			return nil, err
		}

		return res, nil
	}

	return nil, NewErroF(TipagemErro, "O tipo '%s' não aceita a operação de pertinência 'em' (método __contem__ não implementado)", a.Tipo().Nome)
}

// Materializar devolve um iterável indexável para o laço `para ... em` da VM, que percorre
// por índice: Lista e Tupla passam direto (a Lista segue viva durante o laço); os demais
// (Mapa, SequenciaNumerica, Texto, Bytes) são consumidos via __iter__ numa Tupla.
// ponytail: materializa tudo; iteradores infinitos/preguiçosos pediriam um opcode de iteração.
func Materializar(obj Objeto) (Objeto, error) {
	switch obj.(type) {
	case *Lista, Tupla:
		return obj, nil
	}
	it, err := Iter(obj)
	if err != nil {
		return nil, err
	}
	itens := Tupla{}
	for {
		item, err := próximo(it)
		if err != nil {
			if e, ok := err.(*Erro); ok && e.Tipo() == FimIteracao {
				return itens, nil
			}
			return nil, err
		}
		itens = append(itens, item)
	}
}
