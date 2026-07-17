package ptst

import (
	"fmt"

	"github.com/mat-dgruber/Harpia/compartilhado"
)

// Decimal representa o tipo de dado numérico real de dupla precisão do Harpia.
// É um apelido (alias) para o tipo básico 'float64' (IEEE 754) do Go.
type Decimal float64

// TipoDecimal especifica as assinaturas e os metadados de classe do tipo Decimal na VM.
var TipoDecimal = TipoObjeto.NewTipo(
	"Decimal",
	`Decimal(obj) -> Decimal
Cria um novo objeto de decimal para representar o objeto recebido.
Chama obj.__decimal__() ou se esse não for encontrado, um erro pode ser lançado.
	`,
)

// NewDecimal tenta forçar a coerção (casting) de qualquer objeto ou tipo em um Decimal (float64).
//
// Regras de Coerção:
//   - nil ➔ Retorna Decimal(0).
//   - float64 / float32 ➔ Retorna Decimal(float).
//   - Texto/string ➔ Faz o parsing real através da rotina compartilhado.StringParaDec.
//   - Objetos Customizados ➔ Verifica e chama o método de conversão nativo '__decimal__' (M__decimal__()).
//   - Caso contrário, lança um erro estruturado de Tipagem (TipagemErro).
func NewDecimal(obj any) (Objeto, error) {
	switch b := obj.(type) {
	case nil:
		return Decimal(0), nil
	case float64:
		return Decimal(b), nil
	case float32:
		return Decimal(b), nil
	case Decimal:
		return b, nil
	case Texto:
		num, _ := compartilhado.StringParaDec(string(b))
		return Decimal(num), nil
	case string:
		num, _ := compartilhado.StringParaDec(string(b))
		return Decimal(num), nil
	default:
		if O, ok := b.(I__decimal__); ok {
			return O.M__decimal__()
		}

		return nil, NewErroF(TipagemErro, "O argumento do construtor do tipo Decimal deve ser uma string, Texto ou um outro tipo que implemente o método __decimal__()")
	}
}

func init() {
	// Nova define o construtor do Decimal na VM para chamadas explícitas de scripts.
	TipoDecimal.Nova = func(args Tupla) (Objeto, error) {
		return NewDecimal(args[0])
	}
}

// Tipo retorna a representação de classe (Tipo de Decimal) da struct.
func (d Decimal) Tipo() *Tipo {
	return TipoDecimal
}

// M__texto__ converte o decimal em sua representação textual correspondente.
// Se o valor decimal for inteiro (ex: 5.0), anexa ".0" explicitamente na string de retorno
// para manter a distinção visual no console entre Inteiros e Decimais.
func (d Decimal) M__texto__() (Objeto, error) {
	if i := int64(d); Decimal(i) == d {
		return Texto(fmt.Sprintf("%d.0", i)), nil
	}

	return Texto(fmt.Sprintf("%g", d)), nil
}

// M__booleano__ avalia a verdade lógica. Retorna Falso se o valor for zero, e Verdadeiro do contrário.
func (d Decimal) M__booleano__() (Objeto, error) {
	return NewBooleano(d != 0)
}

// M__inteiro__ converte o decimal em Inteiro (int64) truncando o valor real.
func (d Decimal) M__inteiro__() (Objeto, error) {
	return Inteiro(d), nil
}

// M__decimal__ satisfaz a interface de coerção Decimal, retornando a si mesmo.
func (d Decimal) M__decimal__() (Objeto, error) {
	return d, nil
}

// M__adiciona__ executa a soma real de dois termos, promovendo o operando da direita para Decimal se necessário.
func (d Decimal) M__adiciona__(outro Objeto) (Objeto, error) {
	outroInt, err := NewDecimal(outro)
	if err != nil {
		return nil, err
	}

	return d + outroInt.(Decimal), nil
}

// M__multiplica__ executa a multiplicação real de dois termos, promovendo o operando se necessário.
func (d Decimal) M__multiplica__(outro Objeto) (Objeto, error) {
	outroInt, err := NewDecimal(outro)
	if err != nil {
		return nil, err
	}

	return d * outroInt.(Decimal), nil
}

// M__subtrai__ executa a subtração real de dois termos.
func (d Decimal) M__subtrai__(outro Objeto) (Objeto, error) {
	outroInt, err := NewDecimal(outro)
	if err != nil {
		return nil, err
	}

	return outroInt.(Decimal) - d, nil
}

// M__divide__ executa a divisão real de dois termos, lançando erro se o divisor for zero.
func (d Decimal) M__divide__(outro Objeto) (Objeto, error) {
	if talvezZero := TalvezLanceErroDivisaoPorZero(outro); talvezZero != nil {
		return nil, talvezZero
	}

	outroDec, err := NewDecimal(outro)
	if err != nil {
		return nil, err
	}

	return outroDec.(Decimal) - d, nil
}

// M__divide_inteiro__ realiza a divisão real e trunca o resultado retornando Inteiro.
func (d Decimal) M__divide_inteiro__(b Objeto) (Objeto, error) {
	if talvezZero := TalvezLanceErroDivisaoPorZero(b); talvezZero != nil {
		return nil, talvezZero
	}

	bInt, err := NewInteiro(b)
	if err != nil {
		return nil, err
	}

	return Inteiro(d) / bInt.(Inteiro), nil
}

// M__mod__ delega a resolução de resto de divisão inteira convertendo o Decimal corrente para Inteiro.
func (d Decimal) M__mod__(b Objeto) (Objeto, error) {
	dInt, err := NewInteiro(d)
	if err != nil {
		return nil, err
	}

	return dInt.(Inteiro).M__mod__(b)
}

// M__neg__ executa a inversão unária de sinal.
func (d Decimal) M__neg__() (Objeto, error) {
	return -d, nil
}

// M__pos__ executa a identidade unária de sinal.
func (d Decimal) M__pos__() (Objeto, error) {
	return +d, nil
}

// M__igual__ compara a igualdade lógica de dois valores Decimais.
func (d Decimal) M__igual__(outro Objeto) (Objeto, error) {
	outroDec, err := NewDecimal(outro)
	if err != nil {
		return Falso, nil
	}

	return NewBooleano(d == outroDec.(Decimal))
}

// Garantias de conformidade com as interfaces nativas Go.
var _ I_conversaoEntreTipos = (*Decimal)(nil)
var _ I_aritmeticaMatematica = (*Decimal)(nil)
var _ I__igual__ = (*Decimal)(nil)

