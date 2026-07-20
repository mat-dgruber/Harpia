package hrp

import (
	"fmt"

	"github.com/mat-dgruber/Harpia/compartilhado"
)

// Inteiro representa o tipo de dado numérico inteiro com sinal de 64 bits do Harpia.
// É um apelido (alias) para o tipo básico 'int64' do Go.
type Inteiro int64

var (
	cacheInteirosMin = -100
	cacheInteirosMax = 2000
	cacheInteiros    [2101]Objeto // ponytail: pool de alocação rápida Eden Space para inteiros curtos
)

func init() {
	for i := cacheInteirosMin; i <= cacheInteirosMax; i++ {
		cacheInteiros[i-cacheInteirosMin] = Inteiro(i)
	}

	// Nova define o construtor do Inteiro na VM para chamadas explícitas de scripts.
	TipoInteiro.Nova = func(args Tupla) (Objeto, error) {
		return NewInteiro(args[0])
	}
}

// TipoInteiro especifica as assinaturas e metadados de classe do tipo Inteiro na VM.
var TipoInteiro = TipoObjeto.NewTipo(
	"Inteiro",
	`Inteiro(obj) -> Inteiro
Cria um novo objeto de inteiro para representar o objeto recebido.
Chama obj.__inteiro__() ou se esse não for encontrado, um erro pode ser lançado.
	`,
)

// NewInteiro tenta forçar a coerção (casting) de qualquer objeto ou tipo primitivo Go em um Inteiro de 64 bits.
//
// Regras de Coerção:
//   - nil ➔ Retorna Inteiro(0).
//   - int ➔ Retorna Inteiro(int).
//   - Texto/string ➔ Faz o parsing decimal através da rotina compartilhado.StringParaInt.
//   - Objetos Customizados ➔ Verifica e chama o método de conversão nativo '__inteiro__' (M__inteiro__()).
//   - Caso contrário, lança um erro estruturado de Tipagem (TipagemErro).
func NewInteiro(obj any) (Objeto, error) {
	switch b := obj.(type) {
	case nil:
		return cacheInteiros[0-cacheInteirosMin], nil
	case int:
		if b >= cacheInteirosMin && b <= cacheInteirosMax {
			return cacheInteiros[b-cacheInteirosMin], nil
		}
		return Inteiro(b), nil
	case int64:
		if b >= int64(cacheInteirosMin) && b <= int64(cacheInteirosMax) {
			bInt := int(b)
			return cacheInteiros[bInt-cacheInteirosMin], nil
		}
		return Inteiro(b), nil
	case Inteiro:
		return b, nil
	case Texto:
		num, _ := compartilhado.StringParaInt(string(b))
		return Inteiro(num), nil
	case string:
		num, _ := compartilhado.StringParaInt(string(b))
		return Inteiro(num), nil
	default:
		if O, ok := b.(I__inteiro__); ok {
			return O.M__inteiro__()
		}

		return nil, NewErroF(TipagemErro, "O argumento do construtor do tipo Inteiro deve ser uma string, Texto ou um outro tipo que implemente o método __inteiro__()")
	}
}

func init() {
	// Nova define o construtor do Inteiro na VM para chamadas explícitas de scripts.
	TipoInteiro.Nova = func(args Tupla) (Objeto, error) {
		return NewInteiro(args[0])
	}
}

// Tipo retorna a representação de classe (Tipo de Inteiro) da struct.
func (i Inteiro) Tipo() *Tipo {
	return TipoInteiro
}

// M__texto__ converte o número inteiro em sua representação de string de texto decodificada (ex: 123 ➔ "123").
func (i Inteiro) M__texto__() (Objeto, error) {
	return Texto(fmt.Sprintf("%d", i)), nil
}

// M__booleano__ avalia a verdade lógica. Retorna Falso se o valor for zero, e Verdadeiro do contrário.
func (i Inteiro) M__booleano__() (Objeto, error) {
	return NewBooleano(i != 0)
}

// M__inteiro__ satisfaz a interface de coerção Inteiro, retornando a si mesmo.
func (i Inteiro) M__inteiro__() (Objeto, error) {
	return i, nil
}

// M__decimal__ converte o inteiro em Decimal (float64) correspondente.
func (i Inteiro) M__decimal__() (Objeto, error) {
	return Decimal(i), nil
}

// M__adiciona__ executa a operação aritmética de soma.
//
// Regras e Desvios de Coerção:
//   - Se o operando da direita for Inteiro, executa a soma simples retornando um Inteiro.
//   - Se for um Decimal, promove o inteiro local para Decimal e realiza cálculo real.
//   - Se for um tipo incompatível, retorna NaoImplementado para que a VM tente desvios.
func (i Inteiro) M__adiciona__(b Objeto) (Objeto, error) {
	if ok, err := InstanciaDe(b, Tupla{TipoInteiro, TipoDecimal}); !ok {
		return NaoImplementado, nil
	} else if err != nil {
		return nil, err
	}

	if bi, ok := b.(Inteiro); ok {
		return NewInteiro(i + bi)
	}

	return NewDecimal(Decimal(i) + b.(Decimal))
}

// M__multiplica__ executa a multiplicação aritmética de dois membros, promovendo para decimal se necessário.
func (i Inteiro) M__multiplica__(b Objeto) (Objeto, error) {
	if ok, err := InstanciaDe(b, Tupla{TipoInteiro, TipoDecimal}); !ok {
		return NaoImplementado, nil
	} else if err != nil {
		return nil, err
	}

	if bi, ok := b.(Inteiro); ok {
		return NewInteiro(i * bi)
	}

	return NewDecimal(Decimal(i) * b.(Decimal))
}

// M__subtrai__ executa a subtração aritmética de dois membros.
func (i Inteiro) M__subtrai__(b Objeto) (Objeto, error) {
	if ok, err := InstanciaDe(b, Tupla{TipoInteiro, TipoDecimal}); !ok {
		return NaoImplementado, nil
	} else if err != nil {
		return nil, err
	}

	if bi, ok := b.(Inteiro); ok {
		return NewInteiro(i - bi)
	}

	return NewDecimal(Decimal(i) - b.(Decimal))
}

// M__divide__ executa a divisão real de dois membros.
// Garante o tratamento e lançamento de DivisaoPorZeroErro caso o divisor seja igual a zero.
func (i Inteiro) M__divide__(b Objeto) (Objeto, error) {
	if talvezZero := TalvezLanceErroDivisaoPorZero(b); talvezZero != nil {
		return nil, talvezZero
	}

	bInt, err := NewDecimal(b)
	if err != nil {
		return nil, err
	}

	return Decimal(i) / bInt.(Decimal), nil
}

// M__divide_inteiro__ executa a divisão por piso de dois membros, truncando o resultado.
func (i Inteiro) M__divide_inteiro__(b Objeto) (Objeto, error) {
	if talvezZero := TalvezLanceErroDivisaoPorZero(b); talvezZero != nil {
		return nil, talvezZero
	}

	bInt, err := NewInteiro(b)
	if err != nil {
		return nil, err
	}

	return i / bInt.(Inteiro), nil
}

// M__mod__ calcula o resto de divisão inteira (módulo) entre os dois termos.
func (i Inteiro) M__mod__(b Objeto) (Objeto, error) {
	bInt, err := NewInteiro(b)
	if err != nil {
		return nil, err
	}

	return i % bInt.(Inteiro), nil
}

// M__neg__ executa a inversão unária de sinal aritmético (ex: 5 ➔ -5).
func (i Inteiro) M__neg__() (Objeto, error) {
	return -i, nil
}

// M__pos__ executa a identidade unária de sinal.
func (i Inteiro) M__pos__() (Objeto, error) {
	return +i, nil
}

// M__menor_que__ compara se o inteiro atual é estritamente menor do que o operando informado.
func (i Inteiro) M__menor_que__(b Objeto) (Objeto, error) {
	if !MesmoTipo(i, b) {
		return nil, NewErroF(TipagemErro, "A operação '<' não é suportada entre os tipos '%s' e '%s'", i.Tipo().Nome, b.Tipo().Nome)
	}

	return NewBooleano(i < b.(Inteiro))
}

// M__menor_ou_igual__ compara se o inteiro atual é menor ou igual ao operando informado.
func (i Inteiro) M__menor_ou_igual__(b Objeto) (Objeto, error) {
	if !MesmoTipo(i, b) {
		return nil, NewErroF(TipagemErro, "A operação '<=' não é suportada entre os tipos '%s' e '%s'", i.Tipo().Nome, b.Tipo().Nome)
	}

	return NewBooleano(i <= b.(Inteiro))
}

// M__igual__ compara a igualdade lógica de valor numérico.
func (i Inteiro) M__igual__(b Objeto) (Objeto, error) {
	if !MesmoTipo(i, b) {
		return Falso, nil
	}

	return NewBooleano(i == b.(Inteiro))
}

// M__diferente__ compara a desigualdade lógica de valor numérico.
func (i Inteiro) M__diferente__(b Objeto) (Objeto, error) {
	if !MesmoTipo(i, b) {
		return Verdadeiro, nil
	}

	return NewBooleano(i != b.(Inteiro))
}

// M__maior_que__ compara se o inteiro é maior que o operando.
func (i Inteiro) M__maior_que__(b Objeto) (Objeto, error) {
	if !MesmoTipo(i, b) {
		return nil, NewErroF(TipagemErro, "A operação '>' não é suportada entre os tipos '%s' e '%s'", i.Tipo().Nome, b.Tipo().Nome)
	}

	return NewBooleano(i > b.(Inteiro))
}

// M__maior_ou_igual__ compara se o inteiro é maior ou igual ao operando.
func (i Inteiro) M__maior_ou_igual__(b Objeto) (Objeto, error) {
	if !MesmoTipo(i, b) {
		return nil, NewErroF(TipagemErro, "A operação '>=' não é suportada entre os tipos '%s' e '%s'", i.Tipo().Nome, b.Tipo().Nome)
	}

	return NewBooleano(i >= b.(Inteiro))
}

// M__ou__ executa a operação lógica ou bitwise OR (|) entre dois inteiros.
func (i Inteiro) M__ou__(b Objeto) (Objeto, error) {
	if MesmoTipo(i, b) {
		return NewInteiro(i | b.(Inteiro))
	}

	return NaoImplementado, nil
}

// M__e__ executa a operação lógica ou bitwise AND (&) entre dois inteiros.
func (i Inteiro) M__e__(b Objeto) (Objeto, error) {
	if MesmoTipo(i, b) {
		return NewInteiro(i & b.(Inteiro))
	}

	return NaoImplementado, nil
}

// Garantias de assinaturas e conformidade com as interfaces nativas Go.
var _ I_conversaoEntreTipos = (*Inteiro)(nil)
var _ I_aritmeticaMatematica = (*Inteiro)(nil)
var _ I_comparacaoRica = (*Inteiro)(nil)
var _ I_aritmeticaBooleana = (*Inteiro)(nil)
