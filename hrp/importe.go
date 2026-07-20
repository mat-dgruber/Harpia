package hrp

import (
	"os"
	"path"
	"strings"
)

// MaquinarioImporteModulo centraliza e executa todo o fluxo de importação e carregamento
// dinâmico de dependências e módulos (diretiva 'importe').
//
// Algoritmo de Resolução de Dependências:
//  1. Realiza uma consulta prioritária no cache de módulos carregados do supervisor da VM (`ctx.ObterModulo`)
//     para evitar recarregamentos ou recompilações redundantes;
//  2. Se coincidir com um módulo nativo em Go (stdlib) registrado na inicialização (`ObtemImplModulo`),
//     aloca e inicializa o módulo diretamente no contexto da VM;
//  3. Valida as importações físicas: se o módulo não for nativo, seu caminho deve obrigatoriamente conter
//     os prefixos explícitos de busca local "./" ou "/" (ex: './meu_modulo');
//  4. Resolve de forma resiliente caminhos de importação relativa, recuperando a constante conceitual de escopo
//     '__arquivo__' no escopo léxico ativo para extrair a pasta base ('path.Dir');
//  5. Dispara 'ExecutarArquivo' para compilar e interpretar o arquivo físico, gerando a instância do Módulo.
func MaquinarioImporteModulo(ctx *Contexto, nome string, escopo *Escopo) (Objeto, error) {
	if strings.HasPrefix(nome, "@backend/") {
		return CarregarModuloRPC(ctx, nome)
	}

	if impl := ObtemImplModulo(nome); impl != nil {
		return ctx.InicializarModulo(impl)
	}

	if !(strings.HasPrefix(nome, "./") || strings.HasPrefix(nome, "/")) {
		return nil, NewErroF(ImportacaoErro, "Importações não relativas só estão disponíveis para módulos embutidos, corrija para './%s'", nome)
	}

	curDir := ""

	if strings.HasPrefix(nome, "/") {
		curDir, _ = os.Getwd()
	} else if strings.HasPrefix(nome, "./") {
		if escopo == nil {
			panic("Um escopo atual é necessário quando usar importação relativa do tipo './modulo'")
		}

		if arqAtual, err := escopo.ObterValor("__arquivo__"); err != nil {
			panic("O escopo atual precisa informar um `__arquivo__` para poder montar o caminho relativo")
		} else if arqAtual != nil {
			curDir = path.Dir(string(arqAtual.(Texto)))
		}
	}

	caminhos := ctx.Opcs.CaminhosPadrao
	caminhoAbsoluto, err := ResolveArquivohrp(nome, caminhos, curDir)
	if err != nil {
		return nil, NewErroF(ImportacaoErro, "Não foi possível resolver o caminho do módulo '%s': %v", nome, err)
	}

	for _, res := range ctx.ResolvendoModulos {
		if res == caminhoAbsoluto {
			return nil, NewErroF(ImportacaoErro, "Importação cíclica detectada: o módulo '%s' já está em processo de resolução na árvore de dependências", nome)
		}
	}

	if modulo, err := ctx.Modulos.ObterModulo(caminhoAbsoluto); err == nil {
		return modulo, nil
	}
	if modulo, err := ctx.Modulos.ObterModulo(nome); err == nil {
		return modulo, nil
	}

	ctx.ResolvendoModulos = append(ctx.ResolvendoModulos, caminhoAbsoluto)
	defer func() {
		if len(ctx.ResolvendoModulos) > 0 {
			ctx.ResolvendoModulos = ctx.ResolvendoModulos[:len(ctx.ResolvendoModulos)-1]
		}
	}()

	mod, err := ExecutarArquivo(ctx, nome, nome, curDir, true)
	if err != nil {
		return nil, err
	}

	return mod, nil
}

// MultiImporteModulo importa de forma consecutiva e variádica uma série de nomes de módulos na VM.
func MultiImporteModulo(ctx *Contexto, nomes ...string) error {
	for _, nome := range nomes {
		if _, err := MaquinarioImporteModulo(ctx, nome, nil); err != nil {
			return err
		}
	}

	return nil
}

// Importe é a referência global da VM para enlace de carregamento de dependências.
var Importe func(string, *Escopo) (Objeto, error) = func(s string, e *Escopo) (Objeto, error) {
	panic("Antes de usar a função `Importe` você precisa criar um contexto")
}
