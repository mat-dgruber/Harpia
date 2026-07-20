package hrp

import (
	"plugin"
	"strings"
)

// ExecutarString compila uma string contendo comandos Harpia para AST, aloca um escopo virtual
// de módulo para a expressão, e executa-a retornando o objeto Modulo correspondente.
func ExecutarString(ctx *Contexto, codigo string) (*Modulo, error) {
	ast, err := ctx.StringParaAst(codigo, "<string>")
	if err != nil {
		return nil, err
	}

	impl := &ModuloImpl{
		Info: ModuloInfo{},
		Ast:  ast,
	}

	return ctx.InicializarModulo(impl)
}

// ExecutarArquivo localiza e interpreta um arquivo físico de script (extensão '.pt') no disco rígido.
//
// Diferencial de Conectividade Nativa:
// Se o arquivo resolvido e calculado possuir a extensão de objeto binário compartilhado (.so),
// o Harpia carrega dinamicamente o arquivo como um plug-in do Go (plugin.Open), resolve a função
// exportada global 'InicializaModulo()' via reflexão de símbolos, e aciona a inicialização nativa do módulo.
// Isto permite desenvolver extensões binárias de altíssima performance para o Harpia em Go ou C/C++.
func ExecutarArquivo(ctx *Contexto, nome, caminho, curDir string, useSysPaths bool) (*Modulo, error) {
	caminhoCalculado, ast, err := ctx.TransformarEmAst(caminho, useSysPaths, curDir)
	if err != nil {
		return nil, err
	}

	var impl *ModuloImpl

	// Carrega dinamicamente módulos binários compilados compartilhados (.so)
	if strings.HasSuffix(caminhoCalculado, "so") {
		plugin, err := plugin.Open(caminhoCalculado)
		if err != nil {
			return nil, NewErroF(ImportacaoErro, "Erro ao abrir plugin '%s': %s", caminhoCalculado, err)
		}

		inicializaModulo, err := plugin.Lookup("InicializaModulo")
		if err != nil {
			return nil, NewErroF(ImportacaoErro, "Símbolo 'InicializaModulo' não encontrado no módulo: %s", err)
		}

		impl = inicializaModulo.(func() *ModuloImpl)()
	} else {
		impl = &ModuloImpl{
			Info: ModuloInfo{
				Nome:    nome,
				Arquivo: caminhoCalculado,
			},
			Ast: ast,
		}
	}
	return ctx.InicializarModulo(impl)
}
