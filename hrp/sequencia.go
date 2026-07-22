package hrp

// Proximo é o wrapper de alto nível para invocar o método mágico '__proximo__' de um iterador.
//
// Esta função encapsula a asserção de tipo e chamada de M__proximo__(), padronizando e
// centralizando a captura de erros de controle como FimIteracao para as estruturas de controle da VM.
func Proximo(obj Objeto) (Objeto, error) {
	if iter, ok := obj.(I__proximo__); ok {
		return iter.M__proximo__()
	}

	return nil, NewErroF(TipagemErro, "O objeto do tipo '%s' não implementa a interface do iterador (método __proximo__)", obj.Tipo().Nome)
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

	return nil, NewErroF(TipagemErro, "O tipo '%s' não aceita a operação de pertinência 'em' (método __contem__ não implementado)", b.Tipo().Nome)
}
