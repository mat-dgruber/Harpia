// Package embutidos reúne e expõe os tipos primitivos, constantes universais e funções essenciais
// que compõem o escopo global implícito do Harpia.
//
// Diferente de outros módulos da biblioteca padrão (como 'matematica' ou 'sistema'), os elementos
// deste pacote não requerem uma declaração de importação explícita; eles são injetados diretamente
// no escopo primordial de execução de qualquer script no momento de inicialização da máquina virtual.
package embutidos

import (
	"github.com/mat-dgruber/Harpia/hrp"
)

// registrarTipos é um utilitário interno para automatizar o mapeamento de tipos estruturais
// e classes do Harpia diretamente no dicionário global de constantes do pacote.
func registrarTipos(tipos []*hrp.Tipo, mapa hrp.Mapa) {
	for _, tipo := range tipos {
		mapa[tipo.Nome] = tipo
	}
}

type ObjetoPromise struct{}

func (p *ObjetoPromise) Tipo() *hrp.Tipo {
	return hrp.NewTipo("Promise", "Promise JavaScript")
}

func (p *ObjetoPromise) M__obtem_attributo__(nome string) (hrp.Objeto, error) {
	if nome == "resolve" {
		return hrp.NewMetodoOuPanic("resolve", func(_ hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
			if len(args) == 0 {
				return hrp.Nulo, nil
			}
			val := args[0]
			return &ObjetoPromiseResult{val: val}, nil
		}, "Resolve Promise"), nil
	}
	return nil, hrp.NewErroF(hrp.AtributoErro, "Atributo '%s' não existe no tipo 'Promise'", nome)
}

type ObjetoPromiseResult struct {
	val hrp.Objeto
}

func (pr *ObjetoPromiseResult) Tipo() *hrp.Tipo {
	return hrp.NewTipo("PromiseResult", "Promise Result")
}

func (pr *ObjetoPromiseResult) M__obtem_attributo__(nome string) (hrp.Objeto, error) {
	if nome == "then" {
		return hrp.NewMetodoOuPanic("then", func(_ hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
			if len(args) > 0 {
				return hrp.Chamar(args[0], hrp.Tupla{pr.val})
			}
			return pr.val, nil
		}, "Executa o callback then"), nil
	}
	return nil, hrp.NewErroF(hrp.AtributoErro, "Atributo '%s' não existe no tipo 'PromiseResult'", nome)
}

func init() {
	// constantes define o escopo de constantes primordiais e classes de tipos mapeados globalmente.
	constantes := hrp.Mapa{
		"Verdadeiro": hrp.Verdadeiro, // O valor booleano positivo padrão
		"Falso":      hrp.Falso,      // O valor booleano negativo padrão
		"Nulo":       hrp.Nulo,       // A representação padrão de ausência de valor
	}

	// Registra todos os tipos primitivos básicos e classes de exceções padrão na tabela global de símbolos.
	registrarTipos(
		[]*hrp.Tipo{
			hrp.TipoInteiro,
			hrp.TipoDecimal,
			hrp.TipoTexto,
			// hrp.TipoLista,
			// hrp.TipoTupla,
			// hrp.TipoMapa,
			hrp.TipoBooleano,
			hrp.TipoBytes,
			hrp.TipoCanal, // Primitiva de Concorrência CSP por Canais (Fase B)

			// Erros e Exceções estruturadas da VM
			hrp.TipoErro,
			hrp.SintaxeErro,
			hrp.AtributoErro,
			hrp.TipagemErro,
			hrp.NomeErro,
			hrp.ImportacaoErro,
			hrp.ValorErro,
			hrp.IndiceErro,
			hrp.RuntimeErro,
			hrp.FimIteracao,
			hrp.ErroDeAsseguracao,
			hrp.ConsultaErro,
			hrp.ChaveErro,
			hrp.ErroDeSistema,
			hrp.ArquivoNaoEncontradoErro,
		},
		constantes,
	)

	// Cria e registra o alias "imprimir" direcionado para o mesmo ponteiro da função nativa "imprima",
	// provendo maior flexibilidade léxica para os desenvolvedores.
	_emb_imprimir := *_emb_imprima
	_emb_imprimir.Nome = "imprimir"

	// metodos é o catálogo de funções utilitárias que residem no ambiente de execução global.
	metodos := []*hrp.Metodo{
		_emb_imprima,
		&_emb_imprimir,
		_emb_leia,
		_emb_doc,
		_emb_int,
		_emb_texto,
		_emb_tamanho,
		_emb_instanciaDe,
		_emb_sequencia,
		_emb_mesmoTipo,
		// tipo(objeto) retorna o ponteiro correspondente ao tipo de classe do objeto informado.
		hrp.NewMetodoOuPanic(
			"tipo",
			func(_ hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
				if err := hrp.VerificaNumeroArgumentos("tipo", false, args, 1, 1); err != nil {
					return nil, err
				}

				return args[0].Tipo(), nil
			},
			"Obtem o tipo de um objeto",
		),
		// sinal(valorInicial) cria um sinal síncrono estático no backend para SSR
		hrp.NewMetodoOuPanic(
			"sinal",
			func(_ hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
				if err := hrp.VerificaNumeroArgumentos("sinal", false, args, 1, 1); err != nil {
					return nil, err
				}
				valor := args[0]
				// getter apenas retorna o valor estático
				getter := hrp.NewMetodoOuPanic("getter", func(_ hrp.Objeto, _ hrp.Tupla) (hrp.Objeto, error) {
					return valor, nil
				}, "")
				// setter é um stub vazio
				setter := hrp.NewMetodoOuPanic("setter", func(_ hrp.Objeto, _ hrp.Tupla) (hrp.Objeto, error) {
					return hrp.Nulo, nil
				}, "")
				return hrp.Tupla{getter, setter}, nil
			},
			"Cria um sinal reativo estático para SSR no backend",
		),
		// sinalPersistente(chave, valorInicial) stub estático para SSR — no browser, sincroniza com localStorage
		hrp.NewMetodoOuPanic(
			"sinalPersistente",
			func(_ hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
				if err := hrp.VerificaNumeroArgumentos("sinalPersistente", false, args, 2, 2); err != nil {
					return nil, err
				}
				valor := args[1] // ignora a chave, usa valorInicial
				getter := hrp.NewMetodoOuPanic("getter", func(_ hrp.Objeto, _ hrp.Tupla) (hrp.Objeto, error) {
					return valor, nil
				}, "")
				setter := hrp.NewMetodoOuPanic("setter", func(_ hrp.Objeto, _ hrp.Tupla) (hrp.Objeto, error) {
					return hrp.Nulo, nil
				}, "")
				return hrp.Tupla{getter, setter}, nil
			},
			"Cria um sinal reativo persistente (stub SSR)",
		),
		// efeito(funcao) executa imediatamente a função informada
		hrp.NewMetodoOuPanic(
			"efeito",
			func(_ hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
				if err := hrp.VerificaNumeroArgumentos("efeito", false, args, 1, 1); err != nil {
					return nil, err
				}
				return hrp.Chamar(args[0], nil)
			},
			"Executa um efeito reativo imediatamente no backend",
		),
		// derivado(funcao) executa e retorna o valor memoizado estático para SSR
		hrp.NewMetodoOuPanic(
			"derivado",
			func(_ hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
				if err := hrp.VerificaNumeroArgumentos("derivado", false, args, 1, 1); err != nil {
					return nil, err
				}
				res, err := hrp.Chamar(args[0], nil)
				if err != nil {
					return nil, err
				}
				getter := hrp.NewMetodoOuPanic("getter", func(_ hrp.Objeto, _ hrp.Tupla) (hrp.Objeto, error) {
					return res, nil
				}, "")
				return getter, nil
			},
			"Cria um valor derivado estático no backend",
		),
		// armazem(objeto) simplesmente retorna o objeto no backend
		hrp.NewMetodoOuPanic(
			"armazem",
			func(_ hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
				if err := hrp.VerificaNumeroArgumentos("armazem", false, args, 1, 1); err != nil {
					return nil, err
				}
				return args[0], nil
			},
			"Cria um armazem de estado global no backend",
		),
	}

	// Adiciona objeto Roteador SPA nativo no escopo de embutidos
	roteadorObj := hrp.NewMapaVazio()
	rotasCadastradas := hrp.NewMapaVazio()
	caminhoAtualSignal, _ := hrp.NewTexto("/")
	
	roteadorObj.M__define_item__(hrp.Texto("definir"), hrp.NewMetodoOuPanic("definir", func(_ hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
		if err := hrp.VerificaNumeroArgumentos("definir", false, args, 2, 2); err != nil {
			return nil, err
		}
		rotasCadastradas.M__define_item__(args[0], args[1])
		return hrp.Nulo, nil
	}, "Cadastra uma rota no Roteador SPA (caminho, componente)"))

	roteadorObj.M__define_item__(hrp.Texto("navegar"), hrp.NewMetodoOuPanic("navegar", func(_ hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
		if err := hrp.VerificaNumeroArgumentos("navegar", false, args, 1, 1); err != nil {
			return nil, err
		}
		caminhoAtualSignal = args[0].(hrp.Texto)
		return hrp.Nulo, nil
	}, "Navega para uma URL no SPA sem recarregar a página"))

	roteadorObj.M__define_item__(hrp.Texto("obterRota"), hrp.NewMetodoOuPanic("obterRota", func(_ hrp.Objeto, _ hrp.Tupla) (hrp.Objeto, error) {
		return caminhoAtualSignal, nil
	}, "Retorna a rota ativa atual"))

	constantes["roteador"] = roteadorObj

	constantes["Promise"] = &ObjetoPromise{}

	// Registra o escopo agregador de embutidos na lista central de inicializações do interpretador.
	hrp.RegistraModuloImpl(
		&hrp.ModuloImpl{
			Info: hrp.ModuloInfo{
				Nome: "embutidos",
			},
			Constantes: constantes,
			Metodos:    metodos,
		},
	)
}
