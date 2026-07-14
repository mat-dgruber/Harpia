package embutidos

import (
	"github.com/natanfeitosa/portuscript/ptst"
)

// SequenciaNumerica representa o objeto gerador de intervalos numéricos (iterador) do Portuscript.
//
// Esta estrutura guarda o estado dinâmico da iteração ativa, definindo os limites,
// o tamanho do passo de incremento/decremento e o índice correspondente ao valor atual.
type SequenciaNumerica struct {
	Inicio, Fim, Passo, Atual ptst.Inteiro
}

// TipoSequenciaNumerica especifica a assinatura e metadados de classe da estrutura SequenciaNumerica.
var TipoSequenciaNumerica = ptst.NewTipo("SequenciaNumerica", "Gerador de numeros inteiro com ordem crescente")

// Tipo retorna a especificação de tipo da classe SequenciaNumerica para o interpretador.
func (sn *SequenciaNumerica) Tipo() *ptst.Tipo {
	return TipoSequenciaNumerica
}

// M__iter__ satisfaz a interface de objetos iteráveis do Portuscript (ptst.I_iterador).
// Ela retorna a própria estrutura como o iterador ativo a ser varrido pelo laço 'para'.
func (sn *SequenciaNumerica) M__iter__() (ptst.Objeto, error) {
	return sn, nil
}

// M__proximo__ avança e calcula a próxima iteração lógica do laço:
//
// Regras operacionais:
//   - Se o passo for positivo e o valor atual atingir ou passar o limite superior (Fim),
//     lança a exceção controlada 'ptst.FimIteracao' para encerrar o laço graciosamente.
//   - Se o passo for negativo e o valor atual atingir ou passar o limite inferior (Fim),
//     lança a exceção controlada 'ptst.FimIteracao'.
//   - Caso contrário, acumula o incremento ('Passo') no valor 'Atual' e retorna este número inteiro.
func (sn *SequenciaNumerica) M__proximo__() (ptst.Objeto, error) {
	if sn.Passo > 0 && sn.Atual >= sn.Fim {
		return nil, ptst.NewErro(ptst.FimIteracao, ptst.Nulo)
	}

	if sn.Passo < 0 && sn.Atual <= sn.Fim {
		return nil, ptst.NewErro(ptst.FimIteracao, ptst.Nulo)
	}

	sn.Atual += sn.Passo
	return sn.Atual, nil
}

// Garante que a estrutura SequenciaNumerica implemente a interface I_iterador em tempo de compilação.
var _ ptst.I_iterador = (*SequenciaNumerica)(nil)

var met_emb_sequencia_doc = `sequencia(fim) -> SequenciaNumerica
sequencia(inicio, fim, passo?) -> SequenciaNumerica

Gera uma lista de números de [inicio] a [fim] (exclusivos), com incremento de [passo]`

// met_emb_sequencia implementa a lógica nativa para a função global 'sequencia()'.
//
// Esta função aceita entre 1 e 3 argumentos decimais, interpretando-os de forma dinâmica:
//   - 1 Argumento: Interpretado como o limite superior ('fim'). Início assume 0 e passo assume 1.
//   - 2 Argumentos: Interpretados como o limite inferior ('inicio') e superior ('fim'). Passo assume 1.
//   - 3 Argumentos: Define explicitamente 'inicio', 'fim' e 'passo'. Lança erro se o passo for 0.
func met_emb_sequencia(mod ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
	if err := ptst.VerificaNumeroArgumentos("sequencia", false, args, 1, 3); err != nil {
		return nil, err
	}

	var inicio, fim, passo ptst.Objeto = ptst.Inteiro(0), ptst.Inteiro(1), ptst.Inteiro(1)
	var err error

	switch len(args) {
	case 3:
		if inicio, err = ptst.NewInteiro(args[0]); err != nil {
			return nil, err
		}

		if fim, err = ptst.NewInteiro(args[1]); err != nil {
			return nil, err
		}

		if passo, err = ptst.NewInteiro(args[2]); err != nil {
			return nil, err
		} else if passo.(ptst.Inteiro) == 0 {
			return nil, ptst.NewErroF(ptst.ValorErro, "O valor de passo da sequência deve ser diferente de zero")
		}

	case 2:
		if inicio, err = ptst.NewInteiro(args[0]); err != nil {
			return nil, err
		}

		if fim, err = ptst.NewInteiro(args[1]); err != nil {
			return nil, err
		}

	default:
		if fim, err = ptst.NewInteiro(args[1]); err != nil {
			return nil, err
		}
	}

	sn := &SequenciaNumerica{
		Inicio: inicio.(ptst.Inteiro),
		Fim:    fim.(ptst.Inteiro),
		Passo:  passo.(ptst.Inteiro),
		Atual:  0,
	}
	return sn, nil
}

// _emb_sequencia cria e define a assinatura do método 'sequencia' exposto globalmente.
var _emb_sequencia = ptst.NewMetodoOuPanic(
	"sequencia",
	met_emb_sequencia,
	met_emb_sequencia_doc,
)
