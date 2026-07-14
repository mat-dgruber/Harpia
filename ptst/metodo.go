package ptst

import "fmt"

// Metodo representa uma função ou rotina nativa escrita em Go que é exposta e integrada
// de forma chamável no interpretador do Portuscript.
type Metodo struct {
	Nome     string  // Nome identificador do método (ex: "adiciona").
	Doc      string  // Bloco de documentação explicativo (Docstring).
	Modulo   *Modulo // Ponteiro de referência para o módulo ao qual o método pertence.
	chamavel any     // A closure ou função Go subjacente correspondente ao método.
}

// TipoMetodo especifica as assinaturas e metadados de classe do tipo Metodo na VM.
var TipoMetodo = NewTipo("Metodo", "Um metodo Portuscript")

// Tipo retorna a representação de classe (Tipo de Metodo).
func (o *Metodo) Tipo() *Tipo {
	return TipoMetodo
}

// Chamar executa de forma reflexiva a closure Go do método, tratando as assinaturas polimórficas suportadas.
//
// Assinaturas de Métodos Go Suportadas:
//   - `func(inst Objeto, args Tupla) (Objeto, error)`: Assinatura padrão para métodos variádicos dinâmicos.
//   - `func(Objeto) (Objeto, error)`: Assinatura estrita sem parâmetros. Lança erro se argumentos forem passados.
//   - `func(Objeto, Objeto) (Objeto, error)`: Assinatura de um único parâmetro.
func (m *Metodo) Chamar(inst Objeto, args Tupla) (Objeto, error) {
	switch f := m.chamavel.(type) {
	case func(inst Objeto, args Tupla) (Objeto, error):
		return f(inst, args)
	case func(Objeto) (Objeto, error):
		if len(args) != 0 {
			return nil, NewErroF(TipagemErro, "%s() não aceita argumentos, %d foram recebidos", m.Nome, len(args))
		}
		return f(inst)
	case func(Objeto, Objeto) (Objeto, error):
		return f(inst, Objeto(args))
	}

	panic(fmt.Sprintf("Tipo de método desconhecido: %T", m.chamavel))
}

// ObtemDoc devolve o bloco de documentação (Docstring) registrado para o método.
func (m *Metodo) ObtemDoc() string {
	return m.Doc
}

// M__chame__ satisfaz o protocolo de chamabilidade da VM (I__chame__), invocando o método
// com o escopo associado do seu próprio módulo de vinculação.
func (m *Metodo) M__chame__(args Tupla) (Objeto, error) {
	return m.Chamar(Objeto(m.Modulo), args)
}

// M__obtem__ implementa o protocolo descriptor (Descriptor, semelhante ao __get__ do Python).
//
// Quando um método nativo é acessado a partir de uma instância, esta rotina intercepta o acesso e retorna
// um objeto 'MetodoProxy' que realiza o enlace dinâmico de 'self' (a instância local), passando-a
// de forma oculta como primeiro argumento de qualquer chamada de forma automática.
func (m *Metodo) M__obtem__(inst Objeto, dono *Tipo) (Objeto, error) {
	if inst != Nulo {
		return NewMetodoProxy(inst, m), nil
	}

	return m, nil
}

// Garantias de assinaturas estruturais em Go.
var _ I__chame__ = (*Metodo)(nil)
var _ I_Chamar = (*Metodo)(nil)
var _ I_ObtemDoc = (*Metodo)(nil)
var _ I__obtem__ = (*Metodo)(nil)

// NewMetodo aloca e inicializa uma nova instância de Metodo associando o identificador Go.
func NewMetodo(nome string, chamavel any, doc string) (*Metodo, error) {
	return &Metodo{
		Nome:     nome,
		Doc:      doc,
		chamavel: chamavel,
	}, nil
}

// NewMetodoOuPanic é um atalho que aloca o Metodo e dispara panic em caso de falha de instanciação.
func NewMetodoOuPanic(nome string, chamavel any, doc string) *Metodo {
	m, err := NewMetodo(nome, chamavel, doc)

	if err != nil {
		panic(err)
	}

	return m
}

// NewMetodoProxyDeNativo monta adaptadores e wrappers rápidos para envelopar interfaces Go nativas
// (como __iter__, __tamanho__, __texto__) para serem chamáveis de forma unificada no interpretador.
func NewMetodoProxyDeNativo(nome string, chamavel any) (*Metodo, error) {
	metodo := &Metodo{
		Nome: nome,
	}

	switch fn := chamavel.(type) {
	case func() (Objeto, error):
		metodo.chamavel = func(_ Objeto) (Objeto, error) {
			return fn()
		}
	case func(Objeto) (Objeto, error):
		metodo.chamavel = func(_ Objeto, arg Objeto) (Objeto, error) {
			return fn(arg)
		}
	case func(*Tipo, Tupla) (Objeto, error):
		metodo.chamavel = func(inst Objeto, args Tupla) (Objeto, error) {
			meta := args[0].(*Tipo)
			args = args[1:]

			return fn(meta, args)
		}
	default:
		return nil, fmt.Errorf("não foi possível criar um proxy para o método %T", fn)
	}

	return metodo, nil
}
