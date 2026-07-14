// Package embutidos reúne e expõe os tipos primitivos, constantes universais e funções essenciais
// que compõem o escopo global implícito do Portuscript.
//
// Diferente de outros módulos da biblioteca padrão (como 'matematica' ou 'sistema'), os elementos
// deste pacote não requerem uma declaração de importação explícita; eles são injetados diretamente
// no escopo primordial de execução de qualquer script no momento de inicialização da máquina virtual.
package embutidos

import (
	"github.com/natanfeitosa/portuscript/ptst"
)

// registrarTipos é um utilitário interno para automatizar o mapeamento de tipos estruturais
// e classes do Portuscript diretamente no dicionário global de constantes do pacote.
func registrarTipos(tipos []*ptst.Tipo, mapa ptst.Mapa) {
	for _, tipo := range tipos {
		mapa[tipo.Nome] = tipo
	}
}

func init() {
	// constantes define o escopo de constantes primordiais e classes de tipos mapeados globalmente.
	constantes := ptst.Mapa{
		"Verdadeiro": ptst.Verdadeiro, // O valor booleano positivo padrão
		"Falso":      ptst.Falso,      // O valor booleano negativo padrão
		"Nulo":       ptst.Nulo,       // A representação padrão de ausência de valor
	}

	// Registra todos os tipos primitivos básicos e classes de exceções padrão na tabela global de símbolos.
	registrarTipos(
		[]*ptst.Tipo{
			ptst.TipoInteiro,
			ptst.TipoDecimal,
			ptst.TipoTexto,
			// ptst.TipoLista,
			// ptst.TipoTupla,
			// ptst.TipoMapa,
			ptst.TipoBooleano,
			ptst.TipoBytes,

			// Erros e Exceções estruturadas da VM
			ptst.TipoErro,
			ptst.SintaxeErro,
			ptst.AtributoErro,
			ptst.TipagemErro,
			ptst.NomeErro,
			ptst.ImportacaoErro,
			ptst.ValorErro,
			ptst.IndiceErro,
			ptst.RuntimeErro,
			ptst.FimIteracao,
			ptst.ErroDeAsseguracao,
			ptst.ConsultaErro,
			ptst.ChaveErro,
			ptst.ErroDeSistema,
			ptst.ArquivoNaoEncontradoErro,
		},
		constantes,
	)

	// Cria e registra o alias "imprimir" direcionado para o mesmo ponteiro da função nativa "imprima",
	// provendo maior flexibilidade léxica para os desenvolvedores.
	_emb_imprimir := *_emb_imprima
	_emb_imprimir.Nome = "imprimir"

	// metodos é o catálogo de funções utilitárias que residem no ambiente de execução global.
	metodos := []*ptst.Metodo{
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
		ptst.NewMetodoOuPanic(
			"tipo",
			func(_ ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
				if err := ptst.VerificaNumeroArgumentos("tipo", false, args, 1, 1); err != nil {
					return nil, err
				}

				return args[0].Tipo(), nil
			},
			"Obtem o tipo de um objeto",
		),
	}

	// Registra o escopo agregador de embutidos na lista central de inicializações do interpretador.
	ptst.RegistraModuloImpl(
		&ptst.ModuloImpl{
			Info: ptst.ModuloInfo{
				Nome: "embutidos",
			},
			Constantes: constantes,
			Metodos:    metodos,
		},
	)
}
