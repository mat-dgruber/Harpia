package hrp

import (
	"fmt"
)

// ClasseObj representa a modelagem de uma nova classe declarada em código pelo usuário do Harpia.
//
// Ela armazena o nome da classe, as referências de herança para sua classe base correspondente
// e o mapa estruturado de métodos e funções internas.
type ClasseObj struct {
	Nome    string             // Nome identificador da classe (ex: "Pessoa").
	Base    *ClasseObj         // Ponteiro para a classe pai direta (Herança simples).
	Metodos map[string]*Funcao // Catálogo mapeando chaves textuais aos métodos lógicos da classe.
}

// TipoClasseObj especifica as assinaturas e metadados de metaclasse para classes do Harpia na VM.
var TipoClasseObj = NewTipo("ClasseObj", "Metaclasse para classes Harpia")

// Tipo retorna a representação de Tipo de ClasseObj.
func (c *ClasseObj) Tipo() *Tipo {
	return TipoClasseObj
}

// M__obtem_attributo__ resolve a busca de métodos estáticos ou normais na classe, varrendo recursivamente a cadeia de herança.
func (c *ClasseObj) M__obtem_attributo__(nome string) (Objeto, error) {
	classeAtual := c
	for classeAtual != nil {
		if metodo, ok := classeAtual.Metodos[nome]; ok {
			return metodo, nil
		}
		classeAtual = classeAtual.Base
	}

	return nil, NewErroF(AtributoErro, "O atributo '%s' não existe na classe '%s'", nome, c.Nome)
}

// M__nova_instancia__ realiza a instanciação física da classe, retornando um objeto Instancia correspondente.
//
// Algoritmo de Instanciação:
//  1. Aloca a estrutura da Instancia configurando a classe de amarração e inicializando seu dicionário de atributos;
//  2. Varre as tabelas de símbolos em busca do método construtor inicializador '.inicializar()' da classe;
//  3. Se localizar o construtor, executa a chamada deste método injetando os argumentos passados;
//  4. Se o construtor causar falhas, aborta e propaga o erro; do contrário, retorna a instância populada.
func (c *ClasseObj) M__nova_instancia__(meta *Tipo, args Tupla) (Objeto, error) {
	instancia := &Instancia{
		Classe:    c,
		Atributos: make(map[string]Objeto),
	}

	inicializador, err := instancia.M__obtem_attributo__("inicializar")
	if err == nil && inicializador != nil {
		_, err = Chamar(inicializador, args)
		if err != nil {
			return nil, err
		}
	}

	return instancia, nil
}

// Instancia representa o objeto físico instanciado a partir de uma classe customizada do Harpia (ClasseObj).
type Instancia struct {
	GCMixin
	Classe    *ClasseObj        // Ponteiro de amarração que referencia a classe de origem.
	Atributos map[string]Objeto // Tabela de símbolos local contendo as propriedades e atributos dinâmicos do objeto.
}

// Tipo retorna a representação dinâmica e sintética de classe correspondente à instância, mapeando seu nome correto.
func (inst *Instancia) Tipo() *Tipo {
	return NewTipo(inst.Classe.Nome, fmt.Sprintf("Instância da classe %s", inst.Classe.Nome))
}

// M__obtem_attributo__ resolve e extrai o valor de uma propriedade ou método da instância em tempo de execução.
//
// Regras e Enlace de 'self':
//  1. Executa uma busca prioritária local na tabela de 'Atributos' da instância. Se encontrar, devolve;
//  2. Se não constar localmente, sobe para o catálogo de métodos da ClasseObj e de suas classes base (herança);
//  3. Se encontrar o método e ele for 'Estatico', devolve-o diretamente como uma função ordinária;
//  4. Se for um método comum, instancia e retorna um 'MetodoProxy' vinculando a instância atual (self)
//     ao primeiro parâmetro de chamada, automatizando a amarração de forma elegante e transparente.
func (inst *Instancia) M__obtem_attributo__(nome string) (Objeto, error) {
	if val, ok := inst.Atributos[nome]; ok {
		return val, nil
	}

	classeAtual := inst.Classe
	for classeAtual != nil {
		if metodo, ok := classeAtual.Metodos[nome]; ok {
			if metodo.Estatico {
				return metodo, nil
			}
			return &MetodoProxy{Inst: inst, Metodo: metodo}, nil
		}
		classeAtual = classeAtual.Base
	}

	return nil, NewErroF(AtributoErro, "O atributo '%s' não existe na instância de '%s'", nome, inst.Classe.Nome)
}

// M__define_item__ gerencia a atribuição dinâmica de propriedades locais através do operador colchetes [].
func (inst *Instancia) M__define_item__(chave, valor Objeto) (Objeto, error) {
	nomeStr, ok := chave.(Texto)
	if !ok {
		return nil, NewErroF(TipagemErro, "Nome do atributo deve ser texto")
	}
	inst.Atributos[string(nomeStr)] = valor
	return valor, nil
}

// DefinirAtributo é o atalho Go para escrita direta no dicionário de propriedades locais do objeto.
func (inst *Instancia) DefinirAtributo(nome string, valor Objeto) {
	inst.Atributos[nome] = valor
}

// M__define_atributo__ implementa o protocolo de escrita e atribuição de propriedades por ponto (ex: inst.x = 10).
func (inst *Instancia) M__define_atributo__(nome string, valor Objeto) error {
	inst.DefinirAtributo(nome, valor)
	return nil
}

// ObterFilhos retorna todos os atributos ativos da instância para rastreamento de ciclos pelo GC.
func (inst *Instancia) ObterFilhos() []Objeto {
	filhos := make([]Objeto, 0, len(inst.Atributos))
	for _, v := range inst.Atributos {
		if v != nil {
			filhos = append(filhos, v)
		}
	}
	return filhos
}

// Garantias de assinaturas estruturais em Go.
var _ Objeto = (*ClasseObj)(nil)
var _ I__nova_instancia__ = (*ClasseObj)(nil)
var _ I__obtem_attributo__ = (*ClasseObj)(nil)
var _ Objeto = (*Instancia)(nil)
var _ I__obtem_attributo__ = (*Instancia)(nil)
var _ I__define_atributo__ = (*Instancia)(nil)

