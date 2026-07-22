package hrp

import (
	"fmt"
	"os"
	"reflect"
	"strings"
)

// LancarErro realiza a interrupção abrupta e imediata da execução da VM, exibindo o relatório
// formatado da exceção no canal de erros padrão (Stderr) e finalizando o processo com código de saída 1.
func LancarErro(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// adicionaContextoSeNaoTiver vincula de forma segura as informações geográficas (linha, coluna, token, arquivo)
// do supervisor Contexto ao Erro levantado. Isso garante que a renderização visual do traceback
// aponte de forma cirúrgica o trecho exato de código culpado em caso de exceções no runtime.
func adicionaContextoSeNaoTiver(err error, context *Contexto) {
	if err == nil {
		return
	}

	erro, ok := err.(*Erro)
	if !ok {
		return
	}

	if erro.Contexto == nil {
		erro.Contexto = context
	}

	if erro.Linha == -1 {
		if context.LinhaAtual > 0 {
			erro.Linha = context.LinhaAtual - 1
		} else {
			erro.Linha = context.LinhaAtual
		}
		erro.Coluna = context.ColunaAtual
		erro.Token = context.TokenAtual
		erro.Arquivo = context.ArquivoAtual
		erro.Codigo = context.CodigoAtual
	}

	// Double check that if Codigo is still empty, we fallback to context
	if erro.Codigo == "" && context.CodigoAtual != "" {
		erro.Codigo = context.CodigoAtual
	}
}

// Chamar é o despachante universal que aciona e executa o protocolo de chamabilidade da VM (I__chame__).
// Converte os argumentos dinâmicos para o formato Tupla de forma resiliente e executa a chamada do invocável.
func Chamar(obj Objeto, args Objeto) (Objeto, error) {
	var argsTupla Tupla

	if t, ok := args.(Tupla); ok {
		argsTupla = t
	} else {
		argsTupla = Tupla{args}
	}

	// Se o invocável for uma função assíncrona, interceptamos e rodamos concorrentemente em uma goroutine Go
	if fn, ok := obj.(*Funcao); ok && fn.Assincrono {
		prom := NewPromessa()
		if fn.contexto != nil {
			fn.contexto.AdicionarTrabalho()
		}
		go func() {
			defer func() {
				if fn.contexto != nil {
					fn.contexto.EncerrarTrabalho()
				}
			}()
			res, err := fn.M__chame__(argsTupla)
			if err != nil {
				prom.Rejeitar(err)
			} else {
				prom.Resolver(res)
			}
		}()
		return prom, nil
	}

	if I, ok := obj.(I__chame__); ok {
		return I.M__chame__(argsTupla)
	}

	return nil, NewErroF(TipagemErro, "O objeto '%s' não é do tipo chamável.", obj.Tipo().Nome)
}

// NomeAtributo extrai e valida se o nome do atributo acessado corresponde a uma string válida (tipo Texto).
func NomeAtributo(obj Objeto) (string, error) {
	if nome, ok := obj.(Texto); ok {
		return string(nome), nil
	}

	return "", NewErroF(AtributoErro, "O nome do atributo deve ser do tipo texto, não '%s'", obj.Tipo().Nome)
}

// ObtemAtributoRecursivamente varre recursivamente tabelas hash de tipos (classes) e heranças
// em busca de atributos ou métodos.
//
// O Recurso de Otimização e Reflexão Nativa (Reflection):
// Se o atributo solicitado for uma interface de método mágico nativo do Harpia (ex: iniciada e finalizada com "__",
// como "__texto__"), a função executa um desvio de inteligência via 'reflect' em Go. Ela busca se o struct Go do objeto
// implementa um método físico precedido de "M" (ex: "M__texto__"). Se implementado, monta, cria e retorna
// um MetodoProxy de forma instantânea e automatizada, reduzindo dezenas de mapeamentos manuais e redundâncias!
func ObtemAtributoRecursivamente(classe Objeto, nome string) Objeto {
	if classe == nil {
		return nil
	}

	if I, ok := classe.(I_ObtemMapa); ok {
		mapa := I.ObtemMapa()
		if res, ok := mapa[nome]; ok {
			return res
		}
	}

	// Reflexão inteligente de métodos mágicos baseada em convenções de nomenclatura
	if len(nome) > 4 && (strings.HasPrefix(nome, "__") && strings.HasSuffix(nome, "__")) {
		ref := reflect.ValueOf(classe)
		m := ref.MethodByName("M" + nome)
		if m.IsValid() {
			metodo, err := NewMetodoProxyDeNativo(nome, m.Interface())
			if err != nil {
				panic(err)
			}
			return metodo
		}
	}

	if tipo, ok := classe.(*Tipo); ok {
		if tipo.Base != nil {
			obj := ObtemAtributoRecursivamente(tipo.Base, nome)
			if obj != nil {
				return obj
			}
		}
	}

	if tipo := classe.Tipo(); tipo != classe {
		return ObtemAtributoRecursivamente(tipo, nome)
	}

	return nil
}

// ObtemAtributoS executa a resolução dinâmica de atributos e métodos associados em qualquer objeto.
// Respeita a assinatura de protocolo 'I__obtem_attributo__' ou recorre à busca recursiva de herança e descritores.
func ObtemAtributoS(inst Objeto, nome string) (Objeto, error) {
	if I, ok := inst.(I__obtem_attributo__); ok {
		return I.M__obtem_attributo__(nome)
	}

	if obj := ObtemAtributoRecursivamente(inst, nome); obj != nil {
		if desc, ok := obj.(I__obtem__); ok {
			return desc.M__obtem__(inst, inst.Tipo())
		}

		return obj, nil
	}

	return nil, NewErroF(AtributoErro, "O atributo '%s' não existe no tipo '%s'", nome, inst.Tipo().Nome)
}

// Nao executa a negação lógica unária booleana (NOT lógico).
func Nao(obj Objeto) (Objeto, error) {
	booleano, err := NewBooleano(obj)
	if err != nil {
		return nil, err
	}

	switch booleano.(Booleano) {
	case Falso:
		return Verdadeiro, nil
	case Verdadeiro:
		return Falso, nil
	}

	return nil, nil
}

// InstanciaDe valida se o objeto informado é herdeiro ou instância direta de um tipo (classe) ou de uma tupla de tipos.
func InstanciaDe(obj Objeto, tipos any) (Booleano, error) {
	switch tipo_tupla := tipos.(type) {
	case Tupla:
		for _, tipo := range tipo_tupla {
			if ok, err := InstanciaDe(obj, tipo); ok {
				return ok, nil
			} else if err != nil {
				return false, err
			}
		}

		return false, nil
	default:
		return obj.Tipo() == tipos.(*Tipo), nil
	}
}

// ObtemItem é a ponte unificadora que aciona o protocolo de acesso indexado por colchetes [] (I__obtem_item__).
func ObtemItem(inst, arg Objeto) (Objeto, error) {
	if I, ok := inst.(I__obtem_item__); ok {
		return I.M__obtem_item__(arg)
	}

	return nil, NewErroF(TipagemErro, "O tipo '%s' não suporta o uso de indices", inst.Tipo().Nome)
}

// DefineItem é a ponte unificadora que aciona o protocolo de escrita indexada por colchetes [] (I__define_item__).
func DefineItem(inst, chave, valor Objeto) (Objeto, error) {
	if I, ok := inst.(I__define_item__); ok {
		return I.M__define_item__(chave, valor)
	}

	return nil, NewErroF(TipagemErro, "O tipo '%s' não suporta a atribuição por indice", inst.Tipo().Nome)
}

// DefineAtributo aciona e executa a atribuição de propriedades lógicas por meio de ponto (I__define_atributo__).
func DefineAtributo(obj Objeto, nome string, valor Objeto) error {
	if def, ok := obj.(I__define_atributo__); ok {
		return def.M__define_atributo__(nome, valor)
	}

	if inst, ok := obj.(*Instancia); ok {
		return inst.M__define_atributo__(nome, valor)
	}

	return NewErroF(TipagemErro, "O tipo '%s' não suporta a definição de atributos", obj.Tipo().Nome)
}

// NovaInstancia aciona o protocolo de instanciação de novas classes de usuário ou tipos em Go (I__nova_instancia__).
func NovaInstancia(obj Objeto, args Tupla) (Objeto, error) {
	nova, err := ObtemAtributoS(obj, "__nova_instancia__")
	if err == nil {
		var isTipo bool
		if len(args) > 0 {
			_, isTipo = args[0].(*Tipo)
		}

		if !isTipo {
			args = append(Tupla{obj}, args...)
		}

		return Chamar(nova, args)
	}

	if I, ok := obj.(I__nova_instancia__); ok {
		return I.M__nova_instancia__(obj.(*Tipo), args)
	}

	return nil, NewErroF(TipagemErro, "O objeto '%s' não é instanciável", obj.Tipo().Nome)
}

// Tamanho resolve a contagem de elementos de coleções e sequências acionando a interface I__tamanho__.
func Tamanho(obj Objeto) (Objeto, error) {
	if I, ok := obj.(I__tamanho__); ok {
		return I.M__tamanho__()
	}

	return nil, NewErroF(TipagemErro, "Objeto do tipo '%s' não implementa a interface '__tamanho__'.", obj.Tipo().Nome)
}
