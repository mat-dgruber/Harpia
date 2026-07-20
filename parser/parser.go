package parser

import (
	"fmt"
	"strings"

	"github.com/mat-dgruber/Harpia/lexer"
)

// Parser representa o analisador sintático de descida recursiva manual do Harpia.
//
// O Parser consome a torrente de tokens lógicos fornecidos pelo Lexer e os organiza
// em estruturas hierárquicas e nós que formam a Árvore de Sintaxe Abstrata (AST).
type Parser struct {
	lex          *lexer.Lexer              // Instância ativa do analisador léxico.
	token        *lexer.Token              // O token corrente sob avaliação física no parser.
	proximoToken *lexer.Token              // Token de visualização antecipada (lookahead) para tomadas de decisão sintática.
	posicoes     map[BaseNode]*lexer.Token // Mapa unificador associando nós gerados aos tokens físicos (útil para tracebacks de erros).
	codigo       string                    // Cópia do código-fonte original.
	arquivo      string                    // Caminho físico ou identificação lógica do arquivo analisado.
}

// NewParser é o construtor padrão que inicializa e carrega os primeiros tokens de análise.
func NewParser(lex *lexer.Lexer) *Parser {
	parse := &Parser{
		lex:      lex,
		posicoes: make(map[BaseNode]*lexer.Token),
	}
	parse.avancar()
	return parse
}

// NewParserFromString é um construtor de conveniência que cria internamente o Lexer a partir de uma string.
func NewParserFromString(code string, filepath string) *Parser {
	p := NewParser(lexer.NewLexer(code))
	p.codigo = code
	p.arquivo = filepath
	return p
}

// fimDeArquivo verifica se o analisador sintático atingiu o token terminal de encerramento do script.
func (p *Parser) fimDeArquivo() bool {
	return p.token != nil && p.token.Tipo == lexer.TokenFimDeArquivo
}

// avancar move as referências de cursores sintáticos de tokens um passo adiante.
// Mantém as propriedades 'token' e 'proximoToken' (lookahead de tamanho 1) sempre atualizadas.
func (p *Parser) avancar() {
	if p.token == nil {
		p.token = p.lex.ProximoToken()
		p.proximoToken = p.lex.ProximoToken()
		return
	}

	p.token = p.proximoToken

	if p.token.Tipo != lexer.TokenFimDeArquivo {
		p.proximoToken = p.lex.ProximoToken()
	}
}

// consome valida se o valor textual do token corrente coincide com o esperado pelo Parser.
//
// Regras Especiais / Separação de Instruções:
// Se o token esperado for o ponto-e-vírgula (";"), a função executa um tratamento flexível inteligente:
//   - Aceita o caractere ";" explícito na sintaxe, consumindo-o.
//   - Aceita de forma opcional novas linhas ('TokenNovaLinha') ou o encerramento do arquivo ('EOF')
//     como delimitadores e terminadores implícitos de instrução, sem exigir o caractere ";" físico.
//
// Isso unifica o melhor dos mundos entre rigidez e flexibilidade sintática na escrita.
func (p *Parser) consome(token string) error {
	if token == ";" {
		if p.token.Valor == ";" {
			p.avancar()
			return nil
		}
		// Separador implícito: quebra de linha ou fim de arquivo satisfazem a instrução
		if p.token.Tipo == lexer.TokenNovaLinha || p.token.Tipo == lexer.TokenFimDeArquivo {
			return nil
		}
		return fmt.Errorf("era esperado o token ';' ou uma nova linha, mas no lugar foi encontrado '%v'", p.token.Valor)
	}

	if p.token.Valor != token {
		return fmt.Errorf("era esperado o token '%v', mas no lugar foi encontrado '%v'", token, p.token.Valor)
	}

	p.avancar()
	return nil
}

// registrar associa o nó da AST recém-criado ao token físico correspondente no mapa de posições do parser.
// Fundamental para que a VM trace mensagens de erro localizadas graficamente na linha exata do erro.
func (p *Parser) registrar(node BaseNode, tok *lexer.Token) BaseNode {
	if node != nil && tok != nil {
		p.posicoes[node] = tok
	}
	return node
}

// Parse inicia o ciclo completo de análise sintática e retorna o nó raiz 'Programa' contendo toda a AST.
func (p *Parser) Parse() (*Programa, error) {
	declaracoes, err := p.parseDeclaracoes()

	if err != nil {
		return nil, err
	}

	return &Programa{
		Declaracoes: declaracoes,
		Codigo:      p.codigo,
		Arquivo:     p.arquivo,
		Posicoes:    p.posicoes,
	}, nil
}

// parseDeclaracoes varre loops sequenciais de instruções e as acumula até que feche chaves '}' ou atinja EOF.
func (p *Parser) parseDeclaracoes() ([]BaseNode, error) {
	var declaracoes []BaseNode

	for !p.fimDeArquivo() && p.token.Tipo != lexer.TokenFechaChaves {
		if p.token.Tipo != lexer.TokenNovaLinha {
			declaracao, err := p.parseDeclaracao()

			if err != nil {
				return nil, err
			}

			declaracoes = append(declaracoes, declaracao)
		}

		p.avancar()
	}

	return declaracoes, nil
}

// parseDeclaracao processa uma única instrução lógica e registra seu token inicial no mapa de posições.
func (p *Parser) parseDeclaracao() (BaseNode, error) {
	tok := p.token
	res, err := p.parseDeclaracaoInterno()
	if err == nil && res != nil {
		p.registrar(res, tok)
	}
	return res, err
}

// parseDeclaracaoInterno atua como a central de desvio do Parser, identificando palavras-chave
// estruturadas para redirecionar a análise às respectivas subfunções especializadas.
func (p *Parser) parseDeclaracaoInterno() (BaseNode, error) {
	switch p.token.Tipo {
	case lexer.TokenVar, lexer.TokenConst:
		return p.parseVariavel()
	case lexer.TokenRetorne:
		return p.parseRetorne()
	case lexer.TokenDe:
		return p.parseImporteDe()
	case lexer.TokenFunc:
		return p.parseFuncao()
	case lexer.TokenAssincrono:
		p.avancar() // Consome 'assincrono'
		if p.token.Tipo != lexer.TokenFunc {
			return nil, fmt.Errorf("era esperado o token 'func' ou 'funcao' após 'assincrono', mas no lugar foi encontrado '%s'", p.token.Valor)
		}
		fn, err := p.parseFuncao()
		if err != nil {
			return nil, err
		}
		fn.Assincrono = true
		return fn, nil
	case lexer.TokenClasse:
		return p.parseClasse()
	case lexer.TokenEstilo:
		return p.parseEstilo()
	case lexer.TokenTestar:
		return p.parseTeste()
	case lexer.TokenTente:
		return p.parseTenteCapture()
	case lexer.TokenExportar:
		return p.parseExportar()
	case lexer.TokenSe:
		return p.parseExpressaoSe()
	case lexer.TokenEnquanto:
		return p.parseEnquanto()
	case lexer.TokenAguarde:
		p.avancar()
		expr, err := p.parseExpressao()
		if err != nil {
			return nil, err
		}
		return &AguardeNode{Expressao: expr}, nil
	case lexer.TokenPare:
		p.avancar()
		return &PareNode{}, nil
	case lexer.TokenContinue:
		p.avancar()
		return &ContinueNode{}, nil
	case lexer.TokenAssegura:
		p.avancar()

		var condicao, mensagem BaseNode
		var err error

		if condicao, err = p.parseExpressao(); err != nil {
			return nil, err
		}

		if p.token.Tipo == lexer.TokenVirgula {
			p.avancar()
			if mensagem, err = p.parseExpressao(); err != nil {
				return nil, err
			}
		}
		return &AsseguraNode{condicao, mensagem}, nil
	case lexer.TokenPara:
		return p.parseBlocoPara()
	default:
		// Default desvia para expressões aritméticas e/ou possíveis reatribuições compostas (ex: x += 1)
		expressao, err := p.parseExpressao()
		if err != nil {
			return nil, err
		}

		token := p.token.Tipo
		if token >= lexer.TokenMaisIgual && token <= lexer.TokenBarraBarraIgual || token == lexer.TokenIgual {
			reatribuicao := &Reatribuicao{Objeto: expressao, Operador: p.token.Valor}
			p.avancar()

			expressao, err = p.parseExpressao()
			if err != nil {
				return nil, err
			}

			reatribuicao.Expressao = expressao
			if err := p.consome(";"); err != nil {
				return nil, err
			}

			return reatribuicao, nil
		}

		return expressao, nil
	}
}

// parseImporteDe analisa importações parciais (ex: de "matematica" importe PI, raiz;)
func (p *Parser) parseImporteDe() (*ImporteDe, error) {
	p.avancar()
	if p.token.Tipo != lexer.TokenTexto {
		return nil, fmt.Errorf("era esperado um texto após a palavra chave 'de'")
	}

	decl := &ImporteDe{Caminho: &TextoLiteral{p.token.Valor}}
	p.avancar()

	if err := p.consome("importe"); err != nil {
		return nil, err
	}

	for {
		token := p.token

		switch token.Tipo {
		case lexer.TokenIdentificador:
			if IsKeyword(token.Valor) {
				return nil, fmt.Errorf("'%s' é uma palavra-chave reservada e não pode ser importada", token.Valor)
			}
			decl.Nomes = append(decl.Nomes, token.Valor)
			p.avancar()
		case lexer.TokenVirgula:
			p.avancar()
			continue
		default:
			if len(decl.Nomes) == 0 {
				return nil, fmt.Errorf("esperava ao menos um identificador após 'importe', mas recebi '%s'", token.Valor)
			}
			return decl, nil
		}

		if p.token.Tipo == lexer.TokenVirgula {
			p.avancar()
			continue
		}

		return decl, nil
	}
}

// parseBlocoPara analisa laços iterativos: 'para (item em sequencia) { Bloco }'
func (p *Parser) parseBlocoPara() (*BlocoPara, error) {
	p.consome("para")
	if err := p.consome("("); err != nil {
		return nil, err
	}

	id := p.token.Valor
	p.avancar()

	if err := p.consome("em"); err != nil {
		return nil, err
	}

	iter, err := p.parsePrimario()
	if err != nil {
		return nil, err
	}
	if err := p.consome(")"); err != nil {
		return nil, err
	}

	corpo, err := p.parseBloco()
	if err != nil {
		return nil, err
	}

	return &BlocoPara{Identificador: id, Iterador: iter, Corpo: corpo}, nil
}

// parseExpressaoSe analisa blocos estruturados de desvios condicionais se/senao.
func (p *Parser) parseExpressaoSe() (*ExpressaoSe, error) {
	p.consome("se")
	if err := p.consome("("); err != nil {
		return nil, err
	}

	condicao, err := p.parseExpressao()
	if err != nil {
		return nil, err
	}

	expressaoSe := &ExpressaoSe{Condicao: condicao}
	if err := p.consome(")"); err != nil {
		return nil, err
	}

	corpo, err := p.parseBloco()
	if err != nil {
		return nil, err
	}

	expressaoSe.Corpo = corpo

	if p.token.Tipo == lexer.TokenSenao {
		p.avancar()
		var alternativa BaseNode

		switch p.token.Tipo {
		case lexer.TokenSe:
			if alternativa, err = p.parseExpressaoSe(); err != nil {
				return nil, err
			}
		case lexer.TokenAbreChaves:
			if alternativa, err = p.parseBloco(); err != nil {
				return nil, err
			}
		}

		expressaoSe.Alternativa = alternativa
	}

	return expressaoSe, nil
}

// parseEnquanto analisa laços condicionais enquanto: 'enquanto (condicao) { Bloco }'
func (p *Parser) parseEnquanto() (*Enquanto, error) {
	if err := p.consome("enquanto"); err != nil {
		return nil, err
	}

	if err := p.consome("("); err != nil {
		return nil, err
	}

	condicao, err := p.parseExpressao()
	if err != nil {
		return nil, err
	}

	if err := p.consome(")"); err != nil {
		return nil, err
	}

	corpo, err := p.parseBloco()
	if err != nil {
		return nil, err
	}

	return &Enquanto{Condicao: condicao, Corpo: corpo}, nil
}

// parseRetorne analisa saídas de retorno de funções: 'retorne expressao;'
func (p *Parser) parseRetorne() (*RetorneNode, error) {
	if err := p.consome("retorne"); err != nil {
		return nil, err
	}

	retorne := &RetorneNode{}

	if p.token.Tipo != lexer.TokenPontoEVirgula {
		expressao, err := p.parseExpressao()

		if err != nil {
			return nil, err
		}

		retorne.Expressao = expressao
		if err := p.consome(";"); err != nil {
			return nil, err
		}
	}
	return retorne, nil
}

// parseFuncao analisa declarações de funções normais.
func (p *Parser) parseFuncao() (*DeclFuncao, error) {
	if p.token.Tipo != lexer.TokenFunc {
		return nil, fmt.Errorf("era esperado o token 'func' ou 'funcao', mas no lugar foi encontrado '%s'", p.token.Valor)
	}
	p.avancar()

	funcao := &DeclFuncao{}

	if p.token.Tipo == lexer.TokenIdentificador {
		funcao.Nome = p.token.Valor
		p.avancar()
	} else if p.token.Tipo != lexer.TokenAbreParenteses {
		return nil, fmt.Errorf("era esperado o nome da função ou '(', mas no lugar foi encontrado '%s'", p.token.Valor)
	}

	if err := p.consome("("); err != nil {
		return nil, err
	}

	for {
		if p.token.Tipo == lexer.TokenFechaParenteses {
			break
		}

		params, err := p.parseDeclFuncaoParametro()

		if err != nil {
			return nil, err
		}

		funcao.Parametros = append(funcao.Parametros, params)

		if p.token.Tipo == lexer.TokenVirgula {
			p.avancar()
		}
	}

	if err := p.consome(")"); err != nil {
		return nil, err
	}

	if p.token.Tipo == lexer.TokenDoisPontos {
		if err := p.consome(":"); err != nil {
			return nil, err
		}
		funcao.TipoRetorno = p.token.Valor
		p.avancar()
	}

	corpo, err := p.parseBloco()

	if err != nil {
		return nil, err
	}

	funcao.Corpo = corpo

	return funcao, nil
}

// parseClasse analisa declarações de novas classes (Orientação a Objetos) e seus respectivos métodos.
func (p *Parser) parseClasse() (*DeclClasse, error) {
	if err := p.consome("classe"); err != nil {
		return nil, err
	}

	classe := &DeclClasse{}
	classe.Nome = p.token.Valor
	p.avancar()

	if p.token.Tipo == lexer.TokenEstende {
		p.avancar()
		classe.Heranca = p.token.Valor
		p.avancar()
	}

	if err := p.consome("{"); err != nil {
		return nil, err
	}

	for p.token.Tipo == lexer.TokenNovaLinha {
		p.avancar()
	}

	for p.token.Tipo != lexer.TokenFechaChaves && !p.fimDeArquivo() {
		if p.token.Tipo == lexer.TokenNovaLinha {
			p.avancar()
			continue
		}

		isEstatico := false
		if p.token.Tipo == lexer.TokenEstatico {
			p.avancar()
			isEstatico = true
		}

		metodo, err := p.parseFuncao()
		if err != nil {
			return nil, err
		}

		metodo.Estatico = isEstatico
		classe.Metodos = append(classe.Metodos, metodo)

		for p.token.Tipo == lexer.TokenNovaLinha {
			p.avancar()
		}
	}

	if err := p.consome("}"); err != nil {
		return nil, err
	}

	return classe, nil
}

func (p *Parser) parseTeste() (*DeclTeste, error) {
	if err := p.consome("testar"); err != nil {
		return nil, err
	}

	teste := &DeclTeste{}

	// Espera-se que seja um literal de texto contendo o nome do teste
	if p.token.Tipo != lexer.TokenTexto {
		return nil, fmt.Errorf("esperava um texto especificando o nome do teste, obtive '%s'", p.token.Valor)
	}

	// Remove as aspas do texto literal
	val := p.token.Valor
	teste.Nome = val[1 : len(val)-1]
	p.avancar()

	corpo, err := p.parseBloco()
	if err != nil {
		return nil, err
	}
	teste.Corpo = corpo

	return teste, nil
}

func (p *Parser) parseTenteCapture() (*TenteCaptureFinalmente, error) {
	if err := p.consome("tente"); err != nil {
		return nil, err
	}

	tenteNode := &TenteCaptureFinalmente{}
	tenteBlock, err := p.parseBloco()
	if err != nil {
		return nil, err
	}
	tenteNode.TenteBlock = tenteBlock

	// Verifica se há capture e/ou finalmente
	if p.token.Tipo != lexer.TokenCapture && p.token.Tipo != lexer.TokenFinalmente {
		return nil, fmt.Errorf("esperava 'capture' ou 'finalmente' após o bloco 'tente', mas recebi '%s'", p.token.Valor)
	}

	if p.token.Tipo == lexer.TokenCapture {
		if err := p.consome("capture"); err != nil {
			return nil, err
		}

		if err := p.consome("("); err != nil {
			return nil, err
		}

		if p.token.Tipo != lexer.TokenIdentificador {
			return nil, fmt.Errorf("esperava um identificador para o erro no bloco 'capture', mas recebi '%s'", p.token.Valor)
		}
		tenteNode.NomeErro = p.token.Valor
		p.avancar()

		if err := p.consome(")"); err != nil {
			return nil, err
		}

		captureBlock, err := p.parseBloco()
		if err != nil {
			return nil, err
		}
		tenteNode.CaptureBlock = captureBlock
	}

	if p.token.Tipo == lexer.TokenFinalmente {
		if err := p.consome("finalmente"); err != nil {
			return nil, err
		}

		finalmenteBlock, err := p.parseBloco()
		if err != nil {
			return nil, err
		}
		tenteNode.FinalmenteBlock = finalmenteBlock
	}

	return tenteNode, nil
}

func (p *Parser) parseExportar() (*DeclExportar, error) {
	if err := p.consome("exportar"); err != nil {
		return nil, err
	}

	tok := p.token.Tipo
	if tok != lexer.TokenVar && tok != lexer.TokenConst && tok != lexer.TokenFunc && tok != lexer.TokenClasse {
		return nil, fmt.Errorf("esperava declaração de 'var', 'const', 'funcao' ou 'classe' após 'exportar', mas obtive '%s'", p.token.Valor)
	}

	expr, err := p.parseDeclaracaoInterno()
	if err != nil {
		return nil, err
	}

	return &DeclExportar{Expressao: expr}, nil
}

// parseBloco analisa blocos de código delimitados por chaves '{}'.
func (p *Parser) parseBloco() (*Bloco, error) {
	bloco := &Bloco{}

	if err := p.consome("{"); err != nil {
		return nil, err
	}

	decl, err := p.parseDeclaracoes()

	if err != nil {
		return nil, err
	}

	bloco.Declaracoes = decl

	if err := p.consome("}"); err != nil {
		return nil, err
	}

	return bloco, nil
}

// parseDeclFuncaoParametro analisa as assinaturas individuais de parâmetros de funções (como 'a: Inteiro = 10').
func (p *Parser) parseDeclFuncaoParametro() (*DeclFuncaoParametro, error) {
	parametro := &DeclFuncaoParametro{}

	parametro.Nome = p.token.Valor
	p.avancar()

	if p.token.Tipo == lexer.TokenDoisPontos {
		if err := p.consome(":"); err != nil {
			return nil, err
		}

		parametro.Tipo = p.token.Valor
		p.avancar()
	}

	if p.token.Tipo == lexer.TokenIgual {
		if err := p.consome("="); err != nil {
			return nil, err
		}

		expressao, err := p.parseExpressao()

		if err != nil {
			return nil, err
		}

		parametro.Padrao = expressao
	}

	return parametro, nil
}

// parseVariavel analisa a declaração de variáveis e constantes imutáveis: 'var x: Inteiro = 10;'
func (p *Parser) parseVariavel() (*DeclVar, error) {
	decl := &DeclVar{}
	decl.Constante = p.token.Valor == "const"

	p.avancar()

	decl.Nome = p.token.Valor
	p.avancar()

	if p.token.Tipo == lexer.TokenDoisPontos {
		if err := p.consome(":"); err != nil {
			return nil, err
		}

		decl.Tipo = p.token.Valor
		p.avancar()
	}

	if p.token.Tipo == lexer.TokenIgual {
		if err := p.consome("="); err != nil {
			return nil, err
		}

		expressao, err := p.parseExpressao()

		if err != nil {
			return nil, err
		}

		decl.Inicializador = expressao
	} else if decl.Constante {
		return nil, fmt.Errorf("a constante '%s' deve possuir um valor inicializador", decl.Nome)
	}

	if err := p.consome(";"); err != nil {
		return nil, err
	}

	return decl, nil
}

// parseExpressao analisa o escopo geral de expressões complexas.
func (p *Parser) parseExpressao() (BaseNode, error) {
	tok := p.token
	res, err := p.parseExpressaoInterno()
	if err == nil && res != nil {
		p.registrar(res, tok)
	}
	return res, err
}

func (p *Parser) parseExpressaoInterno() (BaseNode, error) {
	if p.token.Tipo == lexer.TokenNova {
		p.avancar()

		obj, err := p.parsePrimario()
		if err != nil {
			return nil, err
		}

		return &NovaNode{obj}, nil
	}

	if p.token.Tipo == lexer.TokenAguarde {
		p.avancar() // Consome 'aguarde'
		expr, err := p.parseExpressao()
		if err != nil {
			return nil, err
		}
		return &AguardeNode{Expressao: expr}, nil
	}

	return p.parsePipe()
}

// parsePipe resolve expressões de encadeamento pipe (|>).
func (p *Parser) parsePipe() (BaseNode, error) {
	esq, err := p.parseDisjuncao()
	if err != nil {
		return nil, err
	}

	for p.token.Tipo == lexer.TokenPipe {
		p.avancar() // Consome |>
		dir, err := p.parseDisjuncao()
		if err != nil {
			return nil, err
		}
		esq = &OpPipe{Esq: esq, Dir: dir}
	}

	return esq, nil
}

// parseEsqLst é a joia arquitetural do Parser.
//
// Monta operadores binários associativos à esquerda (left-associative).
// Recebe uma função 'proximo' de maior prioridade de precedência sintática e uma
// função 'proxOp' que consome e retorna o operador do nível corrente, se casar.
// Evita duplicações maciças de laços de precedência idênticos ao longo do analisador.
func (p *Parser) parseEsqLst(proximo func() (BaseNode, error), proxOp func() (string, bool)) (BaseNode, error) {
	esq, err := proximo()
	if err != nil {
		return nil, err
	}
	for {
		op, ok := proxOp()
		if !ok {
			return esq, nil
		}
		dir, err := proximo()
		if err != nil {
			return nil, err
		}
		esq = &OpBinaria{esq, op, dir}
	}
}

// parseDisjuncao resolve o operador de menor prioridade lógico 'ou'.
func (p *Parser) parseDisjuncao() (BaseNode, error) {
	return p.parseEsqLst(p.parseConjuncao, func() (string, bool) {
		if p.token.Tipo == lexer.TokenBoolOu {
			op := p.token.Valor
			p.avancar()
			return op, true
		}
		return "", false
	})
}

// parseConjuncao resolve o operador lógico 'e'.
func (p *Parser) parseConjuncao() (BaseNode, error) {
	return p.parseEsqLst(p.parseInversao, func() (string, bool) {
		if p.token.Tipo == lexer.TokenBoolE {
			op := p.token.Valor
			p.avancar()
			return op, true
		}
		return "", false
	})
}

// parseInversao resolve o operador lógico de negação unária 'nao'.
func (p *Parser) parseInversao() (BaseNode, error) {
	if p.token.Tipo == lexer.TokenBoolNao {
		p.consome("nao")
		operacao, err := p.parseInversao()

		if err != nil {
			return nil, err
		}

		return &OpUnaria{"nao", operacao}, nil
	}

	return p.parseComparacao()
}

// parseComparacao resolve operadores de comparação relacional, pertencimento ('em') ou instância ('instancia de').
func (p *Parser) parseComparacao() (BaseNode, error) {
	return p.parseEsqLst(p.parseBitABitOu, func() (string, bool) {
		switch p.token.Tipo {
		case lexer.TokenIgualIgual,
			lexer.TokenDiferente,
			lexer.TokenMenorOuIgual,
			lexer.TokenMenorQue,
			lexer.TokenMaiorOuIgual,
			lexer.TokenMaiorQue,
			lexer.TokenEm:
			op := p.token.Valor
			p.avancar()
			return op, true
		case lexer.TokenIdentificador:
			// Resolve o operador composto de correspondência de classes 'instancia de'
			if p.token.Valor == "instancia" && p.proximoToken.Tipo == lexer.TokenDe {
				op := "instancia"
				p.avancar() // Consome "instancia"
				p.avancar() // Consome "de"
				return op, true
			}
		}
		return "", false
	})
}

// parseBitABitOu resolve operador bitwise OR (|).
func (p *Parser) parseBitABitOu() (BaseNode, error) {
	return p.parseEsqLst(p.parseBitABitExOu, func() (string, bool) {
		if p.token.Tipo == lexer.TokenBitABitOu {
			op := p.token.Valor
			p.avancar()
			return op, true
		}
		return "", false
	})
}

// parseBitABitExOu resolve operador bitwise XOR (^).
func (p *Parser) parseBitABitExOu() (BaseNode, error) {
	return p.parseEsqLst(p.parseBitABitE, func() (string, bool) {
		if p.token.Tipo == lexer.TokenBitABitExOu {
			op := p.token.Valor
			p.avancar()
			return op, true
		}
		return "", false
	})
}

// parseBitABitE resolve operador bitwise AND (&).
func (p *Parser) parseBitABitE() (BaseNode, error) {
	return p.parseEsqLst(p.parseDeslocamento, func() (string, bool) {
		if p.token.Tipo == lexer.TokenBitABitE {
			op := p.token.Valor
			p.avancar()
			return op, true
		}
		return "", false
	})
}

// parseDeslocamento resolve operadores de bit shift (<<, >>).
func (p *Parser) parseDeslocamento() (BaseNode, error) {
	return p.parseEsqLst(p.parseAritBasica, func() (string, bool) {
		switch p.token.Tipo {
		case lexer.TokenDeslocEsquerda, lexer.TokenDeslocDireita:
			op := p.token.Valor
			p.avancar()
			return op, true
		}
		return "", false
	})
}

// parseAritBasica resolve operadores de soma e subtração (+, -).
func (p *Parser) parseAritBasica() (BaseNode, error) {
	return p.parseEsqLst(p.parseTermo, func() (string, bool) {
		switch p.token.Tipo {
		case lexer.TokenMais, lexer.TokenMenos:
			op := p.token.Valor
			p.avancar()
			return op, true
		}
		return "", false
	})
}

// parseTermo resolve operadores de multiplicação, divisão, divisão inteira e resto (*, /, //, %).
func (p *Parser) parseTermo() (BaseNode, error) {
	return p.parseEsqLst(p.parseFator, func() (string, bool) {
		switch p.token.Tipo {
		case lexer.TokenAsterisco, lexer.TokenDivisao,
			lexer.TokenDivisaoInteira, lexer.TokenModulo:
			op := p.token.Valor
			p.avancar()
			return op, true
		}
		return "", false
	})
}

// parseFator resolve sinais unários (+, -, ~).
func (p *Parser) parseFator() (BaseNode, error) {
	token := p.token

	switch token.Tipo {
	case lexer.TokenMais, lexer.TokenMenos, lexer.TokenBitABitNao:
		p.avancar()
		expressao, err := p.parseFator()

		if err != nil {
			return nil, err
		}

		return &OpUnaria{token.Valor, expressao}, nil
	}

	return p.parsePotencia()
}

// parsePotencia resolve o operador aritmético de maior prioridade exponenciação (**).
func (p *Parser) parsePotencia() (BaseNode, error) {
	esquerda, err := p.parsePrimario()

	if err != nil {
		return nil, err
	}

	if p.token.Tipo == lexer.TokenPotencia {
		p.avancar()
		direita, err := p.parseFator()

		if err != nil {
			return nil, err
		}

		return &OpBinaria{esquerda, "**", direita}, nil
	}

	return esquerda, nil
}

// parsePrimario resolve acessos a membros (.), chamadas de funções e indexação de arrays com colchetes [].
func (p *Parser) parsePrimario() (BaseNode, error) {
	atom, err := p.parseAtomo()
	if err != nil {
		return nil, err
	}

	for p.token.Tipo == lexer.TokenPonto {
		p.avancar()
		membro, err := p.parseAtomo()
		if err != nil {
			return nil, err
		}

		atom = &AcessoMembro{atom, membro}
	}

	if p.token.Tipo == lexer.TokenAbreParenteses {
		chamada := &ChamadaFuncao{Identificador: atom}

		if err := p.consome("("); err != nil {
			return nil, err
		}

		for p.token.Tipo != lexer.TokenFechaParenteses {
			var expressao BaseNode
			var err error

			// Verifica se é uma atribuição nomeada (identificador seguido de '=')
			if p.token.Tipo == lexer.TokenIdentificador && p.proximoToken != nil && p.proximoToken.Tipo == lexer.TokenIgual {
				nome := p.token.Valor
				p.avancar() // consome o identificador
				p.avancar() // consome '='
				valor, err := p.parseExpressao()
				if err != nil {
					return nil, err
				}
				expressao = &ArgumentoNomeado{Nome: nome, Valor: valor}
			} else {
				expressao, err = p.parseExpressao()
				if err != nil {
					return nil, err
				}
			}

			chamada.Argumentos = append(chamada.Argumentos, expressao)

			if p.token.Tipo == lexer.TokenVirgula {
				p.avancar()
				continue
			}

			if p.token.Tipo != lexer.TokenFechaParenteses {
				return nil, fmt.Errorf("esperava ',' ou ')' na lista de argumentos, mas recebi '%s'", p.token.Valor)
			}
		}

		if err := p.consome(")"); err != nil {
			return nil, err
		}

		return chamada, nil
	}

	for p.token.Tipo == lexer.TokenAbreColchetes {
		p.avancar()

		arg, err := p.parseExpressao()
		if err != nil {
			return nil, err
		}

		if err := p.consome("]"); err != nil {
			return nil, err
		}

		atom = &Indexacao{atom, arg}
	}

	return atom, nil
}

// parseAtomo resolve os elementos terminais de maior prioridade (valores diretos, identificadores e coleções).
func (p *Parser) parseAtomo() (BaseNode, error) {
	token := p.token
	switch token.Tipo {
	case lexer.TokenMenorQue:
		return p.parseJSX()
	case lexer.TokenVerdadeiro, lexer.TokenFalso, lexer.TokenNulo:
		p.avancar()
		return &Identificador{token.Valor}, nil
	case lexer.TokenTexto:
		val := token.Valor
		p.avancar()

		// Se tiver chaves de interpolação { ... } dentro da string (ex: "olá {nome}")
		if strings.Contains(val, "{") && strings.Contains(val, "}") && len(val) >= 2 {
			strContent := val[1 : len(val)-1]
			templateNode := &TemplateLiteral{}

			i := 0
			for i < len(strContent) {
				idxOpen := strings.Index(strContent[i:], "{")
				if idxOpen == -1 {
					// Adiciona o restante como texto literal
					rest := strContent[i:]
					templateNode.Partes = append(templateNode.Partes, &TextoLiteral{Valor: "\"" + rest + "\""})
					break
				}
				idxOpen += i

				// Adiciona a parte de texto antes de '{'
				if idxOpen > i {
					part := strContent[i:idxOpen]
					templateNode.Partes = append(templateNode.Partes, &TextoLiteral{Valor: "\"" + part + "\""})
				}

				idxClose := strings.Index(strContent[idxOpen:], "}")
				if idxClose == -1 {
					// Sem fechamento, trata o resto como texto literal
					templateNode.Partes = append(templateNode.Partes, &TextoLiteral{Valor: "\"" + strContent[idxOpen:] + "\""})
					break
				}
				idxClose += idxOpen

				exprStr := strContent[idxOpen+1 : idxClose]
				// ponytail: parser recursivo isolado por template. Evita implementar gramática de strings complexa com lookahead no parser principal.
				subParser := NewParserFromString(exprStr, p.arquivo)
				exprNode, err := subParser.parseExpressao()
				if err != nil {
					return nil, fmt.Errorf("erro ao analisar expressão de interpolação '%s': %v", exprStr, err)
				}

				templateNode.Partes = append(templateNode.Partes, &TemplateExpr{Expressao: exprNode})
				i = idxClose + 1
			}
			return templateNode, nil
		}
		return &TextoLiteral{token.Valor}, nil
	case lexer.TokenDecimal:
		p.avancar()
		return &DecimalLiteral{token.Valor}, nil
	case lexer.TokenInteiro:
		p.avancar()
		return &InteiroLiteral{token.Valor}, nil
	case lexer.TokenIdentificador, lexer.TokenSelf:
		if !IsKeyword(token.Valor) || token.Tipo == lexer.TokenSelf {
			p.avancar()
			return &Identificador{token.Valor}, nil
		}
	case lexer.TokenAbreParenteses:
		// Parentetização de escopo matemático ou definição de Tuplas literais imutáveis.
		tupla := &TuplaLiteral{}

		for p.token.Tipo != lexer.TokenFechaParenteses {
			p.avancar()
			exp, err := p.parseExpressao()

			if err != nil {
				return nil, err
			}

			if p.token.Tipo != lexer.TokenVirgula {
				if len(tupla.Elementos) == 0 {
					if err := p.consome(")"); err != nil {
						return nil, err
					}

					return exp, nil
				}
			}

			tupla.Elementos = append(tupla.Elementos, exp)
		}

		p.avancar()
		return tupla, nil
	case lexer.TokenAbreColchetes:
		// Análise de listas literais mutáveis: '[a, b, c]'
		literal := &ListaLiteral{}
		p.avancar()

		for p.token.Tipo != lexer.TokenFechaColchetes {
			exp, err := p.parseExpressao()

			if err != nil {
				return nil, err
			}

			literal.Elementos = append(literal.Elementos, exp)

			if p.token.Tipo == lexer.TokenVirgula {
				p.avancar()
			}
		}

		p.avancar()
		return literal, nil
	case lexer.TokenAbreChaves:
		return p.parseMapa()
	case lexer.TokenFunc:
		return p.parseFuncao()
	}

	return nil, fmt.Errorf("o token '%v' não é reconhecido", p.token.Valor)
}

// parseMapa analisa a declaração de dicionários lógicos chave-valor (mapas): '{ chave: valor }'
func (p *Parser) parseMapa() (*MapaLiteral, error) {
	mapa := &MapaLiteral{}
	if err := p.consome("{"); err != nil {
		return nil, err
	}

	for p.token.Tipo != lexer.TokenFechaChaves {
		chave, err := p.parseChaveMapa()
		if err != nil {
			return nil, err
		}

		valorImplicito := false
		valor := chave

		if _, ok := chave.(*Identificador); ok {
			if p.token.Tipo != lexer.TokenDoisPontos {
				valorImplicito = true
				valor = chave
			}
		}

		if !valorImplicito {
			if err := p.consome(":"); err != nil {
				return nil, err
			}
			valor, err = p.parseExpressao()
			if err != nil {
				return nil, err
			}
		}

		mapa.Entradas = append(mapa.Entradas, MapaPar{Chave: chave, Valor: valor, EhImplicito: valorImplicito})

		if p.token.Tipo == lexer.TokenVirgula {
			p.avancar()
		}
	}

	if err := p.consome("}"); err != nil {
		return nil, err
	}
	return mapa, nil
}

// parseChaveMapa lê as chaves de mapas.
// Suporta tanto chaves declaradas de átomos simples (identificadores, strings)
// quanto chaves declaradas por expressões dinâmicas delimitadas por colchetes (ex: {[expressao]: valor}).
func (p *Parser) parseChaveMapa() (BaseNode, error) {
	if p.token.Tipo == lexer.TokenAbreColchetes {
		p.avancar()
		chave, err := p.parseExpressao()
		if err != nil {
			return nil, err
		}
		if err := p.consome("]"); err != nil {
			return nil, err
		}
		return chave, nil
	}
	return p.parseAtomo()
}

// parseEstilo analisa declarações de blocos de estilização: estilo MeuComponente { ... }
func (p *Parser) parseEstilo() (*DeclEstilo, error) {
	if err := p.consome("estilo"); err != nil {
		return nil, err
	}
	if p.token.Tipo != lexer.TokenIdentificador {
		return nil, fmt.Errorf("esperava um identificador após 'estilo', mas recebi '%s'", p.token.Valor)
	}
	nome := p.token.Valor
	p.avancar()

	if err := p.consome("{"); err != nil {
		return nil, err
	}

	// Como a gramática de CSS/estilos é livre e aninhada, e não queremos reinventar o parser de CSS em Go (YAGNI),
	// consumimos tudo até o fechamento correspondente de chaves '}' mantendo chaves aninhadas!
	var regras strings.Builder
	nivelChaves := 1

	for nivelChaves > 0 && !p.fimDeArquivo() {
		switch p.token.Valor {
		case "{":
			nivelChaves++
			regras.WriteString("{")
			p.avancar()
		case "}":
			nivelChaves--
			if nivelChaves > 0 {
				regras.WriteString("}")
				p.avancar()
			}
		default:
			regras.WriteString(p.token.Valor)
			regras.WriteString(" ")
			p.avancar()
		}
	}

	if err := p.consome("}"); err != nil {
		return nil, err
	}

	return &DeclEstilo{Nome: nome, Regras: strings.TrimSpace(regras.String())}, nil
}

// parseJSX analisa um elemento JSX ou uma tag de controle (se/para)
func (p *Parser) parseJSX() (BaseNode, error) {
	if err := p.consome("<"); err != nil {
		return nil, err
	}

	tag := p.token.Valor
	p.avancar()

	if tag == "se" {
		return p.parseSeJSX()
	} else if tag == "para" {
		return p.parseParaJSX()
	}

	node := &NoJSX{Tag: tag}

	// Analisa atributos/propriedades da tag
	for p.token.Tipo != lexer.TokenMaiorQue && !(p.token.Valor == "/" && p.proximoToken != nil && p.proximoToken.Tipo == lexer.TokenMaiorQue) && !p.fimDeArquivo() {
		if p.token.Tipo != lexer.TokenMaiorQue && p.token.Valor != "=" {
			nomeAttr := p.token.Valor
			p.avancar()

			var valorAttr BaseNode
			if p.token.Tipo == lexer.TokenIgual {
				p.avancar()
				if p.token.Tipo == lexer.TokenAbreChaves {
					p.avancar() // consome '{'
					expr, err := p.parseExpressao()
					if err != nil {
						return nil, err
					}
					if err := p.consome("}"); err != nil {
						return nil, err
					}
					valorAttr = expr
				} else if p.token.Tipo == lexer.TokenTexto {
					valorAttr = &TextoLiteral{p.token.Valor}
					p.avancar()
				} else {
					return nil, fmt.Errorf("esperava string ou expressão em '{}' para o atributo '%s', mas recebi '%s'", nomeAttr, p.token.Valor)
				}
			} else {
				// Atributo booleano implícito (ex: <input desabilitado />)
				valorAttr = &Identificador{"Verdadeiro"}
			}
			node.Atributos = append(node.Atributos, &AtributoJSX{Nome: nomeAttr, Valor: valorAttr})
		} else {
			return nil, fmt.Errorf("esperava nome de atributo, mas recebi '%s'", p.token.Valor)
		}
	}

	if p.token.Valor == "/" && p.proximoToken != nil && p.proximoToken.Tipo == lexer.TokenMaiorQue {
		p.avancar() // consome '/'
		p.avancar() // consome '>'
		node.AutoFechado = true
		return node, nil
	}

	if err := p.consome(">"); err != nil {
		return nil, err
	}

	fechouTag := false
	for !p.fimDeArquivo() {
		// Se encontrarmos "</", é o início do fechamento da tag
		if p.token.Tipo == lexer.TokenMenorQue && p.proximoToken != nil && p.proximoToken.Valor == "/" {
			p.avancar() // '<'
			p.avancar() // '/'
			if p.token.Valor != tag {
				return nil, fmt.Errorf("tag de fechamento incorreta: esperava '</%s>', mas recebi '</%s>'", tag, p.token.Valor)
			}
			p.avancar() // tag
			if err := p.consome(">"); err != nil {
				return nil, err
			}
			fechouTag = true
			break
		}

		// Se for outra tag filha
		if p.token.Tipo == lexer.TokenMenorQue {
			filho, err := p.parseJSX()
			if err != nil {
				return nil, err
			}
			node.Filhos = append(node.Filhos, filho)
			continue
		}

		// Se for uma expressão em chaves { ... }
		if p.token.Tipo == lexer.TokenAbreChaves {
			p.avancar() // '{'
			expr, err := p.parseExpressao()
			if err != nil {
				return nil, err
			}
			if err := p.consome("}"); err != nil {
				return nil, err
			}
			node.Filhos = append(node.Filhos, expr)
			continue
		}

		// Caso contrário, é texto literal livre
		var textoLivre strings.Builder
		for p.token.Tipo != lexer.TokenMenorQue && p.token.Tipo != lexer.TokenAbreChaves && !p.fimDeArquivo() {
			textoLivre.WriteString(p.token.Valor)
			textoLivre.WriteString(" ")
			p.avancar()
		}
		txtTrim := strings.TrimSpace(textoLivre.String())
		if len(txtTrim) > 0 {
			node.Filhos = append(node.Filhos, &TextoLiteral{Valor: "\"" + txtTrim + "\""})
		}
	}

	if !fechouTag {
		return nil, fmt.Errorf("tag '<%s>' não foi fechada correspondente", tag)
	}

	return node, nil
}

// parseSeJSX analisa <se condicao={...}>...</se>
func (p *Parser) parseSeJSX() (*NoSeJSX, error) {
	var condicao BaseNode
	var err error

	if p.token.Valor == "condicao" {
		p.avancar()
		if err := p.consome("="); err != nil {
			return nil, err
		}
		if err := p.consome("{"); err != nil {
			return nil, err
		}
		condicao, err = p.parseExpressao()
		if err != nil {
			return nil, err
		}
		if err := p.consome("}"); err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("esperava atributo 'condicao' na tag <se>, mas recebi '%s'", p.token.Valor)
	}

	if err := p.consome(">"); err != nil {
		return nil, err
	}

	node := &NoSeJSX{Condicao: condicao}

	fechouTag := false
	for !p.fimDeArquivo() {
		if p.token.Tipo == lexer.TokenMenorQue && p.proximoToken != nil && p.proximoToken.Valor == "/" {
			p.avancar() // '<'
			p.avancar() // '/'
			if p.token.Valor != "se" {
				return nil, fmt.Errorf("tag de fechamento incorreta para <se>: recebido '</%s>'", p.token.Valor)
			}
			p.avancar()
			if err := p.consome(">"); err != nil {
				return nil, err
			}
			fechouTag = true
			break
		}

		if p.token.Tipo == lexer.TokenMenorQue {
			filho, err := p.parseJSX()
			if err != nil {
				return nil, err
			}
			node.Filhos = append(node.Filhos, filho)
			continue
		}

		if p.token.Tipo == lexer.TokenAbreChaves {
			p.avancar()
			expr, err := p.parseExpressao()
			if err != nil {
				return nil, err
			}
			if err := p.consome("}"); err != nil {
				return nil, err
			}
			node.Filhos = append(node.Filhos, expr)
			continue
		}

		// Texto
		var textoLivre strings.Builder
		for p.token.Tipo != lexer.TokenMenorQue && p.token.Tipo != lexer.TokenAbreChaves && !p.fimDeArquivo() {
			textoLivre.WriteString(p.token.Valor)
			textoLivre.WriteString(" ")
			p.avancar()
		}
		txtTrim := strings.TrimSpace(textoLivre.String())
		if len(txtTrim) > 0 {
			node.Filhos = append(node.Filhos, &TextoLiteral{Valor: "\"" + txtTrim + "\""})
		}
	}

	if !fechouTag {
		return nil, fmt.Errorf("tag '<se>' não foi fechada correspondente")
	}

	return node, nil
}

// parseParaJSX analisa <para item em lista={...}>...</para>
func (p *Parser) parseParaJSX() (*NoParaJSX, error) {
	var item string
	var lista BaseNode
	var err error

	if p.token.Tipo == lexer.TokenIdentificador {
		item = p.token.Valor
		p.avancar()
	} else {
		return nil, fmt.Errorf("esperava nome de variável após '<para', mas recebi '%s'", p.token.Valor)
	}

	if err := p.consome("em"); err != nil {
		return nil, err
	}

	if p.token.Valor == "lista" {
		p.avancar()
		if err := p.consome("="); err != nil {
			return nil, err
		}
		if err := p.consome("{"); err != nil {
			return nil, err
		}
		lista, err = p.parseExpressao()
		if err != nil {
			return nil, err
		}
		if err := p.consome("}"); err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("esperava atributo 'lista' na tag <para>, mas recebi '%s'", p.token.Valor)
	}

	if err := p.consome(">"); err != nil {
		return nil, err
	}

	node := &NoParaJSX{Item: item, Lista: lista}

	fechouTag := false
	for !p.fimDeArquivo() {
		if p.token.Tipo == lexer.TokenMenorQue && p.proximoToken != nil && p.proximoToken.Valor == "/" {
			p.avancar() // '<'
			p.avancar() // '/'
			if p.token.Valor != "para" {
				return nil, fmt.Errorf("tag de fechamento incorreta para <para>: recebido '</%s>'", p.token.Valor)
			}
			p.avancar()
			if err := p.consome(">"); err != nil {
				return nil, err
			}
			fechouTag = true
			break
		}

		if p.token.Tipo == lexer.TokenMenorQue {
			filho, err := p.parseJSX()
			if err != nil {
				return nil, err
			}
			node.Filhos = append(node.Filhos, filho)
			continue
		}

		if p.token.Tipo == lexer.TokenAbreChaves {
			p.avancar()
			expr, err := p.parseExpressao()
			if err != nil {
				return nil, err
			}
			if err := p.consome("}"); err != nil {
				return nil, err
			}
			node.Filhos = append(node.Filhos, expr)
			continue
		}

		// Texto
		var textoLivre strings.Builder
		for p.token.Tipo != lexer.TokenMenorQue && p.token.Tipo != lexer.TokenAbreChaves && !p.fimDeArquivo() {
			textoLivre.WriteString(p.token.Valor)
			textoLivre.WriteString(" ")
			p.avancar()
		}
		txtTrim := strings.TrimSpace(textoLivre.String())
		if len(txtTrim) > 0 {
			node.Filhos = append(node.Filhos, &TextoLiteral{Valor: "\"" + txtTrim + "\""})
		}
	}

	if !fechouTag {
		return nil, fmt.Errorf("tag '<para>' não foi fechada correspondente")
	}

	return node, nil
}
