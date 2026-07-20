package hrp

// Adiciona realiza a soma aritmética polimórfica ou concatenação de dois objetos.
// Verifica se o membro da esquerda 'a' satisfaz o protocolo 'I__adiciona__' e delega a ele.
func Adiciona(a, b Objeto) (Objeto, error) {
	if A, ok := a.(I__adiciona__); ok {
		res, err := A.M__adiciona__(b)

		if err != nil {
			return nil, err
		}

		return res, nil
	}

	return nil, NewErroF(TipagemErro, "A operação '+' não é suportada entre os tipos '%s' e '%s'", a.Tipo().Nome, b.Tipo().Nome)
}

// AdicionaEAtribui executa a operação acumulativa de soma (+=) no objeto.
// Se implementada a interface dedicada, delega a ela. Caso contrário, resolve reatribuindo por Adiciona(a, b).
func AdicionaEAtribui(a, b Objeto) (Objeto, error) {
	if A, ok := a.(I__adiciona_e_atribui__); ok {
		if res, err := A.M__adiciona_e_atribui__(b); err != nil {
			return nil, err
		} else if res != NaoImplementado {
			return res, nil
		}
	}

	return Adiciona(a, b)
}

// Multiplica realiza a multiplicação aritmética polimórfica ou replicação de strings de dois objetos.
func Multiplica(a, b Objeto) (Objeto, error) {
	if A, ok := a.(I__multiplica__); ok {
		res, err := A.M__multiplica__(b)

		if err != nil {
			return nil, err
		}

		return res, nil
	}

	return nil, NewErroF(TipagemErro, "A operação '*' não é suportada entre os tipos '%s' e '%s'", a.Tipo().Nome, b.Tipo().Nome)
}

// MultiplicaEAtribui executa a operação acumulativa de multiplicação (*=) no objeto.
func MultiplicaEAtribui(a, b Objeto) (Objeto, error) {
	if A, ok := a.(I__multiplica_e_atribui__); ok {
		if res, err := A.M__multiplica_e_atribui__(b); err != nil {
			return nil, err
		} else if res != NaoImplementado {
			return res, nil
		}
	}

	return Multiplica(a, b)
}

// Subtrai realiza a subtração aritmética de dois objetos.
func Subtrai(a, b Objeto) (Objeto, error) {
	if A, ok := a.(I__subtrai__); ok {
		res, err := A.M__subtrai__(b)

		if err != nil {
			return nil, err
		}

		return res, nil
	}

	return nil, NewErroF(TipagemErro, "A operação '-' não é suportada entre os tipos '%s' e '%s'", a.Tipo().Nome, b.Tipo().Nome)
}

// SubtraiEAtribui executa a operação acumulativa de subtração (-=) no objeto.
func SubtraiEAtribui(a, b Objeto) (Objeto, error) {
	if A, ok := a.(I__subtrai_e_atribui__); ok {
		if res, err := A.M__subtrai_e_atribui__(b); err != nil {
			return nil, err
		} else if res != NaoImplementado {
			return res, nil
		}
	}

	return Subtrai(a, b)
}

// Divide realiza a divisão real (/) de dois objetos, delegando ao protocolo 'I__divide__'.
func Divide(a, b Objeto) (Objeto, error) {
	if A, ok := a.(I__divide__); ok {
		res, err := A.M__divide__(b)

		if err != nil {
			return nil, err
		}

		return res, nil
	}

	return nil, NewErroF(TipagemErro, "A operação '/' não é suportada entre os tipos '%s' e '%s'", a.Tipo().Nome, b.Tipo().Nome)
}

// DivideEAtribui executa a operação acumulativa de divisão real (/=) no objeto.
func DivideEAtribui(a, b Objeto) (Objeto, error) {
	if A, ok := a.(I__divide_e_atribui__); ok {
		if res, err := A.M__divide_e_atribui__(b); err != nil {
			return nil, err
		} else if res != NaoImplementado {
			return res, nil
		}
	}

	return Divide(a, b)
}

// DivideInteiro realiza a divisão de piso (//) de dois objetos, delegando ao protocolo 'I__divide_inteiro__'.
func DivideInteiro(a, b Objeto) (Objeto, error) {
	if A, ok := a.(I__divide_inteiro__); ok {
		res, err := A.M__divide_inteiro__(b)

		if err != nil {
			return nil, err
		}

		return res, nil
	}

	return nil, NewErroF(TipagemErro, "A operação '//' não é suportada entre os tipos '%s' e '%s'", a.Tipo().Nome, b.Tipo().Nome)
}

// DivideInteiroEAtribui executa a operação acumulativa de divisão por piso (//=) no objeto.
func DivideInteiroEAtribui(a, b Objeto) (Objeto, error) {
	if A, ok := a.(I__divide_inteiro_e_atribui__); ok {
		if res, err := A.M__divide_inteiro_e_atribui__(b); err != nil {
			return nil, err
		} else if res != NaoImplementado {

			return res, nil
		}
	}

	return DivideInteiro(a, b)
}

// Mod calcula o resto de divisão inteira ou realiza formatação/interpolação textual (%).
func Mod(a, b Objeto) (Objeto, error) {
	if A, ok := a.(I__mod__); ok {
		res, err := A.M__mod__(b)

		if err != nil {
			return nil, err
		}

		return res, nil
	}

	return nil, NewErroF(TipagemErro, "A operação '%%' não é suportada entre os tipos '%s' e '%s'", a.Tipo().Nome, b.Tipo().Nome)
}

// MenorQue compara se o objeto 'a' é estritamente menor que 'b' (<).
func MenorQue(a, b Objeto) (Objeto, error) {
	if A, ok := a.(I__menor_que__); ok {
		res, err := A.M__menor_que__(b)

		if err != nil {
			return nil, err
		}

		return res, nil
	}

	return nil, NewErroF(TipagemErro, "A operação '<' não é suportada entre os tipos '%s' e '%s'", a.Tipo().Nome, b.Tipo().Nome)
}

// MenorOuIgual compara se 'a' é menor ou igual a 'b' (<=).
func MenorOuIgual(a, b Objeto) (Objeto, error) {
	if A, ok := a.(I__menor_ou_igual__); ok {
		res, err := A.M__menor_ou_igual__(b)

		if err != nil {
			return nil, err
		}

		return res, nil
	}

	return nil, NewErroF(TipagemErro, "A operação '<=' não é suportada entre os tipos '%s' e '%s'", a.Tipo().Nome, b.Tipo().Nome)
}

// Igual compara a igualdade de valores semânticos de dois objetos (==).
func Igual(a, b Objeto) (Objeto, error) {
	if A, ok := a.(I__igual__); ok {
		res, err := A.M__igual__(b)

		if err != nil {
			return nil, err
		}

		return res, nil
	}

	return nil, NewErroF(TipagemErro, "A operação '==' não é suportada entre os tipos '%s' e '%s'", a.Tipo().Nome, b.Tipo().Nome)
}

// Diferente compara a desigualdade de valores lógicos de dois objetos (!=).
func Diferente(a, b Objeto) (Objeto, error) {
	if A, ok := a.(I__diferente__); ok {
		res, err := A.M__diferente__(b)

		if err != nil {
			return nil, err
		}

		return res, nil
	}

	return nil, NewErroF(TipagemErro, "A operação '!=' não é suportada entre os tipos '%s' e '%s'", a.Tipo().Nome, b.Tipo().Nome)
}

// MaiorQue compara se 'a' é estritamente maior que 'b' (>).
func MaiorQue(a, b Objeto) (Objeto, error) {
	if A, ok := a.(I__maior_que__); ok {
		res, err := A.M__maior_que__(b)

		if err != nil {
			return nil, err
		}

		return res, nil
	}

	return nil, NewErroF(TipagemErro, "A operação '>' não é suportada entre os tipos '%s' e '%s'", a.Tipo().Nome, b.Tipo().Nome)
}

// MaiorOuIgual compara se 'a' é maior ou igual a 'b' (>=).
func MaiorOuIgual(a, b Objeto) (Objeto, error) {
	if A, ok := a.(I__maior_ou_igual__); ok {
		res, err := A.M__maior_ou_igual__(b)

		if err != nil {
			return nil, err
		}

		return res, nil
	}

	return nil, NewErroF(TipagemErro, "A operação '>=' não é suportada entre os tipos '%s' e '%s'", a.Tipo().Nome, b.Tipo().Nome)
}

// Ou realiza a operação lógica OR bitwise (|) entre dois objetos.
func Ou(a, b Objeto) (Objeto, error) {
	if A, ok := a.(I__ou__); ok {
		res, err := A.M__ou__(b)

		if err != nil {
			return nil, err
		}

		if res != NaoImplementado {
			return res, nil
		}
	}

	return nil, NewErroF(TipagemErro, "A operação '|' não é suportada entre os tipos '%s' e '%s'", a.Tipo().Nome, b.Tipo().Nome)
}

// E realiza a operação lógica AND bitwise (&) entre dois objetos.
func E(a, b Objeto) (Objeto, error) {
	if A, ok := a.(I__e__); ok {
		res, err := A.M__e__(b)

		if err != nil {
			return nil, err
		}

		if res != NaoImplementado {
			return res, nil
		}
	}

	return nil, NewErroF(TipagemErro, "A operação '&' não é suportada entre os tipos '%s' e '%s'", a.Tipo().Nome, b.Tipo().Nome)
}

// Neg realiza o sinal e inversão unária aritmética (-).
func Neg(a Objeto) (Objeto, error) {
	if A, ok := a.(I__neg__); ok {
		res, err := A.M__neg__()
		if err != nil {
			return nil, err
		}

		return res, nil
	}

	return nil, NewErroF(TipagemErro, "A operação '-' não é suportada para o tipo '%s'", a.Tipo().Nome)
}

// Pos realiza a identidade de sinal unário (+).
func Pos(a Objeto) (Objeto, error) {
	if A, ok := a.(I__pos__); ok {
		res, err := A.M__pos__()
		if err != nil {
			return nil, err
		}

		return res, nil
	}

	return nil, NewErroF(TipagemErro, "A operação '+' não é suportada para o tipo '%s'", a.Tipo().Nome)
}
