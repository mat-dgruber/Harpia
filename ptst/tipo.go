package ptst

import (
	"fmt"
	"strings"
)

// NovaFunc define a assinatura da closure responsável pela instanciação básica de um objeto
// (corresponde conceitualmente ao método especial estático de alocação de memória '__nova_instancia__').
type NovaFunc func(args Tupla) (Objeto, error)

// InicializaFunc define a assinatura para a rotina de inicialização de valores e propriedades
// (corresponde ao método construtor '__inicializa__', similar ao '__init__' do Python).
type InicializaFunc func(args Tupla) error

// Tipo é a estrutura fundamental de representação de classes e metaclasses na VM do Portuscript.
//
// Cada instância de Tipo define uma classe, carregando seu dicionário de atributos, ponteiro
// para a classe pai (herança), manual de documentação técnica (Doc) e os ponteiros de funções
// construtoras e alocadoras em Go.
type Tipo struct {
	Nome       string         // Nome identificador textual da classe (ex: "Inteiro").
	Nova       NovaFunc       // Closure para alocação/instanciação (__nova_instancia__).
	Inicializa InicializaFunc // Função inicializadora do objeto (__inicializa__).
	Doc        string         // Bloco de documentação do tipo (Docstring).
	Base       *Tipo          // A classe base da qual esta herda propriedades (Herança).
	Mapa       Mapa           // Tabela hash contendo todos os atributos, métodos e propriedades da classe.
}

// NewTipo é o construtor padrão em Go para registrar novas classes simples.
// Ele aloca a estrutura, cria o dicionário de chaves e a enfileira para a montagem estrutural pré-runtime.
func NewTipo(nome string, doc string) *Tipo {
	t := &Tipo{Nome: nome, Doc: doc, Mapa: Mapa{}}
	EnfileiraMontagemDoTipo(t)
	return t
}

// Tipo satisfaz a interface Objeto, indicando que a própria estrutura 'Tipo' também é um objeto
// gerenciável pela VM (metaclasse).
func (b *Tipo) Tipo() *Tipo {
	return b
}

// ObtemDoc devolve o bloco de documentação (Docstring) registrado para a classe.
func (b *Tipo) ObtemDoc() string {
	return b.Doc
}

// NewTipo cria e retorna uma subclasse derivada a partir do tipo atual, estabelecendo a herança.
func (b *Tipo) NewTipo(nome string, doc string) *Tipo {
	t := &Tipo{Nome: nome, Doc: doc, Base: b, Mapa: Mapa{}}
	EnfileiraMontagemDoTipo(t)
	return t
}

// NewTipoX cria uma subclasse com injeções explícitas de closures construtoras e inicializadoras.
func (b *Tipo) NewTipoX(nome string, doc string, nova NovaFunc, inicializa InicializaFunc) *Tipo {
	t := &Tipo{Nome: nome, Doc: doc, Base: b, Nova: nova, Inicializa: inicializa, Mapa: Mapa{}}
	EnfileiraMontagemDoTipo(t)
	return t
}

// ObtemMapa retorna a tabela interna hash de chaves, propriedades e métodos do tipo.
func (b *Tipo) ObtemMapa() Mapa {
	return b.Mapa
}

// Monta realiza a compilação de herança, carregando métodos da classe pai para a atual,
// e estruturando os metadados de documentação na chave especial '__doc__'.
func (b *Tipo) Monta() error {
	if b.Mapa == nil {
		b.Mapa = Mapa{}
	}

	if _, ok := b.Mapa["__doc__"]; !ok {
		if b.Doc != "" {
			b.Mapa["__doc__"] = Texto(strings.Trim(b.Doc, "\r\n\t "))
		} else {
			b.Mapa["__doc__"] = Nulo
		}
	}

	return nil
}

// M__nova_instancia__ executa a alocação e criação física de uma nova instância da classe.
func (b *Tipo) M__nova_instancia__(meta *Tipo, args Tupla) (Objeto, error) {
	if b.Nova != nil {
		return b.Nova(args)
	}

	return nil, NewErroF(TipagemErro, "O objeto '%s' não é instanciável", b.Nome)
}

// TipoTipo é a metaclasse raiz de representação de todos os Tipos na VM.
var TipoTipo *Tipo = NewTipo(
	"Tipo",
	"Tipo raiz para todos os objetos (interno).",
)

// TipoObjeto é a classe de base universal de herança de onde todos os objetos descendem.
var TipoObjeto *Tipo = TipoTipo.NewTipo(
	"Objeto",
	"A classe base para todas as outras classes.",
)

func init() {
	TipoTipo.Monta()
	TipoObjeto.Monta()
}

// filaMontagem acumula temporariamente todos os tipos alocados em Go antes da inicialização do runtime.
var filaMontagem []*Tipo

// EnfileiraMontagemDoTipo enfileira um novo Tipo para sofrer o processo de montagem estrutural.
func EnfileiraMontagemDoTipo(tipo *Tipo) {
	filaMontagem = append(filaMontagem, tipo)
}

// MontaOsTipos varre os tipos enfileirados em filaMontagem, executando o método Monta() em cada um
// para consolidar heranças e tabelas de símbolos de forma consistente pré-execução.
func MontaOsTipos() error {
	for _, tipo := range filaMontagem {
		err := tipo.Monta()

		if err != nil {
			return fmt.Errorf("Erro ao montar o tipo %s: %v", tipo.Nome, err)
		}

		filaMontagem = nil
	}

	return nil
}

func init() {
	err := MontaOsTipos()

	if err != nil {
		panic(err)
	}

	// Injeta o método mágico especial de coerção textual '__texto__' na classe base de Tipos.
	// Retorna o nome identificador da classe quando impresso ou convertido para Texto.
	TipoTipo.Mapa["__texto__"] = NewMetodoOuPanic(
		"__texto__",
		func(inst Objeto) (Objeto, error) {
			if T, ok := inst.(*Tipo); ok {
				return NewTexto(T.Nome)
			}

			return NewTexto(inst.Tipo().Nome)
		},
		"Retorna o nome da classe",
	)
}

// Garantias de assinaturas estruturais em tempo de compilação Go.
var _ Objeto = (*Tipo)(nil)
var _ I_ObtemMapa = (*Tipo)(nil)
var _ I__nova_instancia__ = (*Tipo)(nil)