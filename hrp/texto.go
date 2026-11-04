package hrp

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

// Texto representa o tipo de dado textual (string) nativo do Harpia.
// É um apelido (alias) para o tipo básico 'string' (codificação UTF-8 por padrão) do Go.
type Texto string

// TipoTexto especifica as assinaturas e os metadados de classe do tipo Texto na VM.
var TipoTexto = TipoObjeto.NewTipo(
	"Texto",
	`Texto(obj) -> Texto
Cria um novo objeto de texto para representar o objeto recebido.
Chama obj.__texto__() ou obj.__repr__(), se nenhum dos dois for encontrado, um erro pode ser lançado.
	`,
)

// NewTexto tenta forçar a coerção (casting) de qualquer objeto ou interface Go em um Texto (string) do Harpia.
//
// Regras de Coerção:
//   - nil ➔ Retorna Texto("").
//   - string ➔ Realiza Unquote seguro através de 'strconv.Unquote' para resolver sequências de escape literais.
//   - Objetos Customizados ➔ Tenta buscar e chamar o atributo e método dinâmico '__texto__' de forma polimórfica.
//   - Caso contrário, lança erro de tipagem.
func NewTexto(arg any) (Objeto, error) {
	switch obj := arg.(type) {
	case nil:
		return Texto(""), nil
	case _Nulo:
		return Texto("Nulo"), nil
	case string:
		unquoted, err := strconv.Unquote(obj)
		if err == nil {
			return Texto(unquoted), nil
		}
		// Se falhar ou não estiver cotada, tenta embrulhar em aspas e dar unquote
		unquoted, err = strconv.Unquote(`"` + obj + `"`)
		if err == nil {
			return Texto(unquoted), nil
		}
		return Texto(obj), nil
	case Texto:
		return obj, nil
	}

	if met, _ := ObtemAtributoS(arg.(Objeto), "__texto__"); met != nil {
		if chame, ok := met.(I__chame__); ok {
			return chame.M__chame__(Tupla{})
		}
	}

	if O, ok := arg.(I__texto__); ok {
		return O.M__texto__()
	}

	return Texto(fmt.Sprintf("%v", arg)), nil
}

func init() {
	// Nova define o construtor do Texto na VM para chamadas explícitas de scripts.
	TipoTexto.Nova = func(args Tupla) (Objeto, error) {
		return NewTexto(args[0])
	}
}

// Tipo retorna a representação de classe (Tipo de Texto) da struct.
func (t Texto) Tipo() *Tipo {
	return TipoTexto
}

// M__texto__ satisfaz a interface de coerção Texto, retornando a si mesmo.
func (t Texto) M__texto__() (Objeto, error) {
	return t, nil
}

// M__bytes__ converte a string local em uma representação do tipo de dados Bytes (raw byte array).
func (t Texto) M__bytes__() (Objeto, error) {
	return NewBytes(string(t))
}

// M__booleano__ avalia a verdade lógica. Retorna Falso se a string for vazia (""), e Verdadeiro do contrário.
func (t Texto) M__booleano__() (Objeto, error) {
	return NewBooleano(len(t) != 0)
}

// M__igual__ compara a igualdade lógica de duas strings.
func (t Texto) M__igual__(outro Objeto) (Objeto, error) {
	if !MesmoTipo(t, outro) {
		return Falso, nil
	}

	return NewBooleano(t == outro.(Texto))
}

// M__menor_que__ compara lexicograficamente se o texto é menor que outro (<).
func (t Texto) M__menor_que__(outro Objeto) (Objeto, error) {
	if !MesmoTipo(t, outro) {
		return nil, NewErroF(TipagemErro, "A operação '<' não é suportada entre os tipos '%s' e '%s'", t.Tipo().Nome, outro.Tipo().Nome)
	}
	return NewBooleano(string(t) < string(outro.(Texto)))
}

// M__menor_ou_igual__ compara lexicograficamente se o texto é menor ou igual a outro (<=).
func (t Texto) M__menor_ou_igual__(outro Objeto) (Objeto, error) {
	if !MesmoTipo(t, outro) {
		return nil, NewErroF(TipagemErro, "A operação '<=' não é suportada entre os tipos '%s' e '%s'", t.Tipo().Nome, outro.Tipo().Nome)
	}
	return NewBooleano(string(t) <= string(outro.(Texto)))
}

// M__maior_que__ compara lexicograficamente se o texto é maior que outro (>).
func (t Texto) M__maior_que__(outro Objeto) (Objeto, error) {
	if !MesmoTipo(t, outro) {
		return nil, NewErroF(TipagemErro, "A operação '>' não é suportada entre os tipos '%s' e '%s'", t.Tipo().Nome, outro.Tipo().Nome)
	}
	return NewBooleano(string(t) > string(outro.(Texto)))
}

// M__maior_ou_igual__ compara lexicograficamente se o texto é maior ou igual a outro (>=).
func (t Texto) M__maior_ou_igual__(outro Objeto) (Objeto, error) {
	if !MesmoTipo(t, outro) {
		return nil, NewErroF(TipagemErro, "A operação '>=' não é suportada entre os tipos '%s' e '%s'", t.Tipo().Nome, outro.Tipo().Nome)
	}
	return NewBooleano(string(t) >= string(outro.(Texto)))
}

// M__adiciona__ executa a concatenação de duas strings. Lança erro de tipagem caso tente somar tipos distintos.
func (t Texto) M__adiciona__(outro Objeto) (Objeto, error) {
	if !MesmoTipo(t, outro) {
		return nil, NewErroF(TipagemErro, "Não é possível concatenar o tipo '%s' com '%s'", t.Tipo().Nome, outro.Tipo().Nome)
	}

	outroTexto, err := NewTexto(outro)

	if err != nil {
		return nil, err
	}

	return t + outroTexto.(Texto), nil
}

// M__multiplica__ repete a string atual N vezes, onde N é um número Inteiro positivo (ex: "A" * 3 ➔ "AAA").
func (t Texto) M__multiplica__(outro Objeto) (Objeto, error) {
	switch obj := outro.(type) {
	case Inteiro:
		qtd := int64(obj)
		if qtd <= 0 {
			return Texto(""), nil
		}
		if len(t) > 0 && int64(len(t))*qtd > (64<<20) {
			return nil, NewErroF(ErroDeLimite, "Texto resultante excede o limite de memória")
		}
		return Texto(strings.Repeat(string(t), int(qtd))), nil
	default:
		return nil, NewErroF(TipagemErro, "A operação '*' não é suportada entre os tipos '%s' e '%s'", t.Tipo().Nome, obj.Tipo().Nome)
	}
}

// M__iter__ satisfaz o protocolo de objetos iteráveis, retornando um iterador que percorre o texto caractere a caractere.
// Preserva corretamente caracteres Unicode multibyte (runas).
func (t Texto) M__iter__() (Objeto, error) {
	return NewIterador(t)
}

// M__tamanho__ retorna o comprimento real de caracteres Unicode (runas) e não a quantidade de bytes raw de strings.
// Garante o tratamento correto de strings multibyte em loops.
func (t Texto) M__tamanho__() (Objeto, error) {
	return Inteiro(utf8.RuneCountInString(string(t))), nil
}

// String devolve a representação de string nativa do Go.
func (t Texto) String() string {
	return string(t)
}

// M__contem__ verifica se uma substring existe contida na string atual utilizando strings.Contains do Go.
func (t Texto) M__contem__(obj Objeto) (Objeto, error) {
	if other, err := NewTexto(obj); err != nil {
		return nil, err
	} else if other != nil {
		if strings.Contains(string(t), string(other.(Texto))) {
			return Verdadeiro, nil
		}
	}

	return Falso, nil
}

var reFormatoTexto = regexp.MustCompile(`%(\w)`)

// M__mod__ implementa a interpolação de strings no Harpia usando o operador modulo % (semelhante ao Python).
//
// Suporta formatações:
//   - %i: Formata valores de Inteiros.
//   - %d: Formata valores Decimais.
//   - %b: Formata valores Booleanos.
//   - %s (default): Formata qualquer objeto chamando sua representação textual.
func (t Texto) M__mod__(obj Objeto) (res Objeto, err error) {
	defer func() {
		if erroRecuperado := recover(); erroRecuperado != nil {
			if e, ok := erroRecuperado.(error); ok {
				err = e
			} else {
				err = fmt.Errorf("%v", erroRecuperado)
			}
			res = nil
		}
	}()

	copia := string(t)
	var args Tupla

	if tupla, ok := obj.(Tupla); ok {
		args = tupla
	} else {
		args = Tupla{obj}
	}

	maximo := len(args)
	atual := 0

	copia = reFormatoTexto.ReplaceAllStringFunc(copia, func(s string) string {
		if atual >= maximo {
			return s
		}

		var conversor FuncaoComErro[any] = NewTexto

		switch s[1] {
		case 'i':
			conversor = NewInteiro
		case 'd':
			conversor = NewDecimal
		case 'b':
			conversor = NewBooleano
		}

		arg := args[atual]
		if arg == nil {
			arg = Nulo
		}
		atual += 1

		return string(
			RetornaOuPanic(
				NewTexto,
				RetornaOuPanic(conversor, arg.(any)).(any),
			).(Texto),
		)
	})

	res, err = NewTexto(copia)
	return
}

// Interfaces Go satisfeitas pela struct Texto.
var _ I__texto__ = Texto("")
var _ I__bytes__ = Texto("")
var _ I__booleano__ = Texto("")
var _ I__igual__ = Texto("")
var _ I__menor_que__ = Texto("")
var _ I__menor_ou_igual__ = Texto("")
var _ I__maior_que__ = Texto("")
var _ I__maior_ou_igual__ = Texto("")
var _ I__adiciona__ = Texto("")
var _ I__multiplica__ = Texto("")
var _ I__tamanho__ = Texto("")
var _ I__contem__ = Texto("")
var _ I__iter__ = Texto("")

func init() {
	// Injeção de métodos de instância de Texto no mapa da classe.

	TipoTexto.Mapa["junta"] = NewMetodoOuPanic("junta", func(inst Objeto, iter Objeto) (Objeto, error) {
		tupla, ok := iter.(Tupla)
		if !ok {
			return nil, NewErroF(TipagemErro, "O argumento de 'junta' deve ser uma Tupla")
		}

		var sb strings.Builder
		sep := string(inst.(Texto))

		for i, arg := range tupla {
			texto, err := NewTexto(arg)
			if err != nil {
				return nil, err
			}

			sb.WriteString(string(texto.(Texto)))
			if i != len(tupla)-1 {
				sb.WriteString(sep)
			}
		}

		return Texto(sb.String()), nil
	}, `concatena o iterável recebido com o texto da instancia`)

	TipoTexto.Mapa["titulo"] = NewMetodoOuPanic("titulo", func(inst Objeto) (Objeto, error) {
		titularizado := strings.Title(strings.ToLower(string(inst.(Texto))))
		return Texto(titularizado), nil
	}, "retorna uma cópia do texto com a primeira letra da frase em maiúsculo")

	TipoTexto.Mapa["maiusculas"] = NewMetodoOuPanic("maiusculas", func(inst Objeto) (Objeto, error) {
		return Texto(strings.ToUpper(string(inst.(Texto)))), nil
	}, "retorna uma cópia do texto com todas as letras em maiúsculas")

	TipoTexto.Mapa["minusculas"] = NewMetodoOuPanic("minusculas", func(inst Objeto) (Objeto, error) {
		return Texto(strings.ToLower(string(inst.(Texto)))), nil
	}, "retorna uma cópia do texto com todas as letras em minúsculas")

	TipoTexto.Mapa["trocar"] = NewMetodoOuPanic("trocar", func(inst Objeto, args Tupla) (Objeto, error) {
		if err := VerificaNumeroArgumentos("trocar", true, args, 2, 2); err != nil {
			return nil, err
		}
		antigo, ok1 := args[0].(Texto)
		novo, ok2 := args[1].(Texto)
		if !ok1 || !ok2 {
			return nil, NewErroF(TipagemErro, "Os argumentos de 'trocar' devem ser do tipo 'Texto'")
		}
		original := string(inst.(Texto))
		resultado := strings.ReplaceAll(original, string(antigo), string(novo))
		return Texto(resultado), nil
	}, "retorna uma cópia do texto substituindo as ocorrências de um termo por outro")

	TipoTexto.Mapa["contem"] = NewMetodoOuPanic("contem", func(inst Objeto, args Tupla) (Objeto, error) {
		if err := VerificaNumeroArgumentos("contem", true, args, 1, 1); err != nil {
			return nil, err
		}
		return inst.(Texto).M__contem__(args[0])
	}, `Verifica se a substring está contida no texto`)

	TipoTexto.Mapa["dividir"] = NewMetodoOuPanic("dividir", func(inst Objeto, args Tupla) (Objeto, error) {
		if err := VerificaNumeroArgumentos("dividir", true, args, 1, 1); err != nil {
			return nil, err
		}
		separador, ok := args[0].(Texto)
		if !ok {
			return nil, NewErroF(TipagemErro, "O argumento de 'dividir' deve ser do tipo 'Texto'")
		}
		original := string(inst.(Texto))
		partes := strings.Split(original, string(separador))
		saida := make(Tupla, 0, len(partes))
		for _, p := range partes {
			saida = append(saida, Texto(p))
		}
		return saida, nil
	}, "divide o texto em uma tupla de substrings usando o separador fornecido")

	fnComeca := NewMetodoOuPanic("comecaCom", func(inst Objeto, args Tupla) (Objeto, error) {
		if err := VerificaNumeroArgumentos("comecaCom", true, args, 1, 1); err != nil {
			return nil, err
		}
		prefixo, ok := args[0].(Texto)
		if !ok {
			return nil, NewErroF(TipagemErro, "O argumento de 'comecaCom' deve ser do tipo 'Texto'")
		}
		if strings.HasPrefix(string(inst.(Texto)), string(prefixo)) {
			return Verdadeiro, nil
		}
		return Falso, nil
	}, "verifica se o texto começa com o prefixo informado")

	TipoTexto.Mapa["comecaCom"] = fnComeca
	TipoTexto.Mapa["comeca_com"] = fnComeca

	fnTermina := NewMetodoOuPanic("terminaCom", func(inst Objeto, args Tupla) (Objeto, error) {
		if err := VerificaNumeroArgumentos("terminaCom", true, args, 1, 1); err != nil {
			return nil, err
		}
		sufixo, ok := args[0].(Texto)
		if !ok {
			return nil, NewErroF(TipagemErro, "O argumento de 'terminaCom' deve ser do tipo 'Texto'")
		}
		if strings.HasSuffix(string(inst.(Texto)), string(sufixo)) {
			return Verdadeiro, nil
		}
		return Falso, nil
	}, "verifica se o texto termina com o sufixo informado")

	TipoTexto.Mapa["terminaCom"] = fnTermina
	TipoTexto.Mapa["termina_com"] = fnTermina
}
