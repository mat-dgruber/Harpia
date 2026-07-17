package ptst

import (
	"bytes"
	"reflect"
)

// refsMapas gerencia de forma global as contagens de referências das instâncias do tipo Mapa (ponteiros implícitos).
var refsMapas = make(map[uintptr]int)

// Mapa representa a coleção do tipo dicionário ou mapa chave-valor associativo do Harpia (ex: { "a": 1 }).
//
// É um apelido (alias) para o tipo nativo hashmap `map[string]Objeto` do Go.
// No Harpia, as chaves de mapas são restritas estritamente ao tipo Texto (string).
type Mapa map[string]Objeto

// Métodos de gerenciamento explícito de memória (ObjetoGC) para o tipo Mapa.

func (m Mapa) Reter() {
	addr := reflect.ValueOf(m).Pointer()
	refsMapas[addr]++
}

func (m Mapa) Liberar() {
	addr := reflect.ValueOf(m).Pointer()
	if refsMapas[addr] > 0 {
		refsMapas[addr]--
		if refsMapas[addr] == 0 {
			delete(refsMapas, addr) // Limpa registro do cache
		}
	}
}

func (m Mapa) ObterRefs() int {
	addr := reflect.ValueOf(m).Pointer()
	return refsMapas[addr]
}

func (m Mapa) ObterFilhos() []Objeto {
	filhos := make([]Objeto, 0, len(m))
	for _, valor := range m {
		filhos = append(filhos, valor)
	}
	return filhos
}

// TipoMapa especifica as assinaturas e metadados de classe do tipo Mapa na VM.
var TipoMapa = NewTipo(
	"Mapa",
	"Objeto chave/valor",
)

// NewMapaVazio aloca e retorna uma nova instância de Mapa vazia.
func NewMapaVazio() Mapa {
	return make(Mapa)
}

// Tipo retorna a representação de classe (Tipo de Mapa).
func (m Mapa) Tipo() *Tipo {
	return TipoMapa
}

// M__texto__ converte os pares de chave-valor em sua representação string agregada, envolvida por chaves {}.
func (m Mapa) M__texto__() (Objeto, error) {
	var out bytes.Buffer
	out.WriteString("{ ")
	separar := false

	for chave, valor := range m {
		if separar {
			out.WriteString(", ")
		}

		var chaveT, valorT Objeto
		var err error

		if chaveT, err = NewTexto(chave); err != nil {
			return nil, err
		}

		if valorT, err = NewTexto(valor); err != nil {
			return nil, err
		}

		out.WriteString(string(chaveT.(Texto)))
		out.WriteString(": ")
		out.WriteString(string(valorT.(Texto)))
		separar = true
	}

	out.WriteString(" }")
	return NewTexto(out.String())
}

// M__iter__ satisfaz o protocolo de objetos iteráveis.
//
// Diferencial de Iteração:
// A iteração sobre mapas do Harpia varre e devolve consecutivamente uma Tupla
// contendo o par [Chave, Valor] de cada elemento, permitindo desestruturação fluida.
func (m Mapa) M__iter__() (Objeto, error) {
	entradas := make(Tupla, 0, len(m))

	for chave, valor := range m {
		entradas = append(entradas, Tupla{Texto(chave), valor})
	}

	return NewIterador(entradas)
}

// M__tamanho__ retorna a quantidade total de chaves registradas no dicionário.
func (m Mapa) M__tamanho__() (Objeto, error) {
	return NewInteiro(len(m))
}

// M__obtem_item__ lê e resolve o valor correspondente à chave fornecida. Lança ChaveErro se a chave não existir.
func (m Mapa) M__obtem_item__(obj Objeto) (Objeto, error) {
	chave, ok := obj.(Texto)

	if !ok {
		return nil, NewErroF(ChaveErro, "A chave para um '%s' deve ser do tipo '%s' e não '%s'", TipoMapa.Nome, TipoTexto.Nome, obj.Tipo().Nome)
	}

	if valor, ok := m[string(chave)]; ok {
		return valor, nil
	}

	return nil, NewErroF(ChaveErro, "O Mapa não tem um elemento com a chave '%s'", chave)
}

// M__define_item__ registra ou sobrescreve uma associação chave-valor mutável no dicionário local.
func (m Mapa) M__define_item__(obj, valor Objeto) (Objeto, error) {
	chave, ok := obj.(Texto)

	if !ok {
		return nil, NewErroF(ChaveErro, "A chave para um '%s' deve ser do tipo '%s' e não '%s'", TipoMapa.Nome, TipoTexto.Nome, obj.Tipo().Nome)
	}

	m[string(chave)] = valor
	return nil, nil
}

// Garantias de assinaturas estruturais em Go.
var _ I__iter__ = (*Mapa)(nil)
var _ I__texto__ = (*Mapa)(nil)
var _ I__tamanho__ = (*Mapa)(nil)
var _ I__obtem_item__ = (*Mapa)(nil)
var _ I__define_item__ = (*Mapa)(nil)

// Chaves retorna uma tupla imutável contendo todas as chaves (Texto) registradas no mapa.
func (m Mapa) Chaves() (Tupla, error) {
	if len(m) == 0 {
		return Tupla(nil), nil
	}

	chaves := make(Tupla, 0)

	for chave := range m {
		chaves = append(chaves, Texto(chave))
	}

	return chaves, nil
}

// Valores retorna uma tupla contendo todos os valores dos objetos cadastrados no mapa.
func (m Mapa) Valores() (Tupla, error) {
	if len(m) == 0 {
		return Tupla(nil), nil
	}

	valores := make(Tupla, 0)

	for _, valor := range m {
		valores = append(valores, valor)
	}

	return valores, nil
}

// Atualizar mescla os dados de outro mapa na instância corrente (operação merge mutável).
// Se ignoreExistentes for Verdadeiro, chaves preexistentes repetidas não sofrerão sobreposição.
func (m Mapa) Atualizar(outro Mapa, ignoreExistentes Booleano) (Mapa, error) {
	for c, v := range outro {
		if ignoreExistentes {
			if _, existe := m[c]; existe {
				continue
			}
		}

		m[c] = v
	}

	return m, nil
}

func init() {
	// Injeção de métodos estáticos do tipo Mapa no interpretador.

	TipoMapa.Mapa["chaves"] = NewMetodoOuPanic("chaves", func(inst Objeto) (Objeto, error) {
		return inst.(Mapa).Chaves()
	}, `Retorna uma tupla contendo todos as chaves do mapa`)

	TipoMapa.Mapa["valores"] = NewMetodoOuPanic("valores", func(inst Objeto) (Objeto, error) {
		return inst.(Mapa).Valores()
	}, `Retorna uma tupla contendo todos os valores do mapa`)

	TipoMapa.Mapa["atualizar"] = NewMetodoOuPanic("atualizar", func(inst Objeto, args Tupla) (Objeto, error) {
		if err := VerificaNumeroArgumentos("atualizar", true, args, 1, 2); err != nil {
			return nil, err
		}

		ignoreExistentes := Falso
		novoMapa := args[0]

		if _, ok := novoMapa.(Mapa); !ok {
			return nil, NewErroF(TipagemErro, "Era esperado o tipo 'Mapa', mas ao invés disso foi recebido o tipo '%s'", novoMapa.Tipo().Nome)
		}

		if len(args) == 2 {
			if ignore, ok := args[1].(Booleano); ok {
				ignoreExistentes = ignore
			} else if !ok {
				return nil, NewErroF(TipagemErro, "Era esperado o tipo 'Booleano', mas ao invés disso foi recebido o tipo '%s'", args[1].Tipo().Nome)
			}
		}

		return inst.(Mapa).Atualizar(novoMapa.(Mapa), ignoreExistentes)
	}, `mapa.atualizar(outroMapa, ignoreExistentes?) -> Mapa
Atualiza o mapa atual com as chaves/valores do outro, se o parâmetro ignoreExistentes for Verdadeiro,
as chaves que se repetem serão mantidas com o valor atual do mapa que chamou o método.`)
}
