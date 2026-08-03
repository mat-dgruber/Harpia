// # Analisador Sintático do Harpia
//
// Este arquivo implementa o Parser do Harpia, um descida recursiva manual
// (_recursive descent parser_) que consome tokens do lexer e produz uma
// Árvore de Sintaxe Abstrata (AST) fortemente tipada.
//
// # Algoritmo
//
// O parser utiliza lookahead de 1 token (próximoToken) e mantém um cursor
// duplo (token corrente + lookahead) através de avançar. A precedência de
// operadores é implementada por camadas de funções (parseDisjuncao,
// parseConjuncao, parseComparacao, ...) que delegam ao helper genérico
// parseEsqLst para operadores binários associativos à esquerda.
//
// # Erros
//
// Erros sintáticos são acumulados em ErroSintatico e renderizados com cores
// ANSI, código HRP-E101, indicador ^^^^^^^^^^^ e dica contextual em português.
//
// # Nota de Segurança
//
// Este arquivo é parte do compilador Harpia. Ele apenas transforma tokens em
// estruturas de dados; não executa código do usuário. Não é código malicioso.
package parser

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/mat-dgruber/Harpia/lexer"
)

// Parser representa o analisador sintático de descida recursiva manual do Harpia.
//
// O Parser consome a torrente de tokens lógicos fornecidos pelo Lexer e os
// organiza em estruturas hierárquicas e nós que formam a Árvore de Sintaxe
// Abstrata (AST). Mantém lookahead de 1 token e tabela de posições para
// reportar erros de runtime com precisão de linha/coluna.
type Parser struct {
	lex                  *lexer.Lexer              // Instância ativa do analisador léxico.
	token                *lexer.Token              // O token corrente sob avaliação física no parser.
	próximoToken         *lexer.Token              // Token de visualização antecipada (lookahead) para tomadas de decisão sintática.
	posicoes             map[BaseNode]*lexer.Token // Mapa unificador associando nós gerados aos tokens físicos (útil para tracebacks de erros).
	codigo               string                    // Cópia do código-fonte original.
	arquivo              string                    // Caminho físico ou identificação lógica do arquivo analisado.
	erros                []ErroSintatico           // Lista de erros sintáticos acumulados durante a análise.
	ultimoDocComment     string                    // Comentários de documentação acumulados imediatamente antes de uma declaração.
	ultimoTokenFimIndice int                       // Índice no código fonte onde o último token consumido terminou.
}

// NewParser é o construtor padrão que inicializa e carrega os primeiros tokens
// de análise. Pré-carrega token e próximoToken para evitar verificações nil
// durante o parsing.
func NewParser(lex *lexer.Lexer) *Parser {
	parse := &Parser{
		lex:      lex,
		posicoes: make(map[BaseNode]*lexer.Token),
	}
	parse.avancar()
	return parse
}

// NewParserFromString é um construtor de conveniência que cria internamente o
// Lexer a partir de uma string de código-fonte.
//
// É o atalho preferido para testes, REPL e playgrounds onde não há um
// arquivo físico. O parâmetro filepath é usado apenas em mensagens de erro.
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

func (p *Parser) proximoTokenSemComentario() *lexer.Token {
	for {
		tok := p.lex.PróximoToken()
		if tok.Tipo == lexer.TokenComentarioDoc {
			if p.ultimoDocComment == "" {
				p.ultimoDocComment = tok.Valor
			} else {
				p.ultimoDocComment += "\n" + tok.Valor
			}
			continue
		}
		return tok
	}
}

func (p *Parser) consumirDocComment() string {
	doc := p.ultimoDocComment
	p.ultimoDocComment = ""
	return strings.TrimSpace(doc)
}

// avancar move as referências de cursores sintáticos de tokens um passo adiante.
// Mantém as propriedades 'token' e 'próximoToken' (lookahead de tamanho 1) sempre atualizadas.
func (p *Parser) avancar() {
	if p.token == nil {
		p.token = p.proximoTokenSemComentario()
		p.próximoToken = p.proximoTokenSemComentario()
		if p.token != nil {
			p.ultimoTokenFimIndice = p.token.Fim.Indice
		}
		return
	}

	p.ultimoTokenFimIndice = p.token.Fim.Indice
	p.token = p.próximoToken

	if p.token.Tipo != lexer.TokenFimDeArquivo {
		p.próximoToken = p.proximoTokenSemComentario()
	}
}

// consome valida se o valor textual do token corrente coincide com o esperado
// pelo Parser e o consome em caso afirmativo.
//
// # Regras Especiais — Separação de Instruções
//
// Se o token esperado for o ponto-e-vírgula (";"), a função executa um
// tratamento flexível inteligente:
//
//   - Aceita o caractere ";" explícito na sintaxe, consumindo-o.
//   - Aceita de forma opcional novas linhas ('TokenNovaLinha') ou o
//     encerramento do arquivo ('EOF') como delimitadores e terminadores
//     implícitos de instrução, sem exigir o caractere ";" físico.
//
// Isso unifica o melhor dos mundos entre rigidez e flexibilidade sintática
// na escrita.
//
// # Erros
//
// Retorna um *ErroSintatico com dica contextual apropriada via ObterDicaSintatica.
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
		return p.errof("era esperado o token ';' ou uma nova linha, mas no lugar foi encontrado '%v'", p.token.Valor)
	}

	if p.token.Valor != token {
		return p.errof("era esperado o token '%v', mas no lugar foi encontrado '%v'", token, p.token.Valor)
	}

	p.avancar()
	return nil
}

// registrar associa o nó da AST recém-criado ao token físico correspondente
// no mapa de posições do parser.
//
// # Importância
//
// Fundamental para que a VM trace mensagens de erro localizadas graficamente
// na linha e coluna exatas do erro. Sem este registro, exceções em runtime
// perderiam a referência física ao local onde foram geradas.
//
// # Comportamento
//
//   - Armazena em p.posicoes[node] = tok (lookup O(1)).
//   - Se o nó implementar a interface opcional DefinirPos, atualiza suas
//     coordenadas imediatamente.
//
// # Performance
//
// O mapa cresce linearmente com o tamanho da AST. Para scripts típicos (≤
// 10k nós) o overhead é desprezível.
func (p *Parser) registrar(node BaseNode, tok *lexer.Token) BaseNode {
	if node != nil && tok != nil {
		p.posicoes[node] = tok
		if posNode, ok := node.(interface{ DefinirPos(i, f lexer.PosicaoToken) }); ok {
			posNode.DefinirPos(tok.Inicio, tok.Fim)
		}
	}
	return node
}

// Parse inicia o ciclo completo de análise sintática e retorna o nó raiz
// 'Programa' contendo toda a AST.
//
// # Comportamento
//
//   - Erros de parsing são acumulados em p.erros; o primeiro é retornado como
//     erro primário e o programa é descartado (fail-fast).
//   - Se não houver declaração alguma, Inicio/Fim do Programa ficam zerados.
//   - A função não executa o código — apenas constrói a árvore.
//
// # Erros
//
// Retorna um *ErroSintatico com renderização ANSI quando há falha.
func (p *Parser) Parse() (*Programa, error) {
	declaracoes, err := p.parseDeclaracoes()

	if err != nil {
		return nil, err
	}
	if len(p.erros) > 0 {
		return nil, &p.erros[0]
	}

	prog := &Programa{
		Declaracoes: declaracoes,
		Codigo:      p.codigo,
		Arquivo:     p.arquivo,
		Posicoes:    p.posicoes,
	}
	if len(declaracoes) > 0 {
		ini, _ := declaracoes[0].Pos()
		_, fim := declaracoes[len(declaracoes)-1].Pos()
		prog.Inicio = ini
		prog.Fim = fim
	}
	return prog, nil
}

// parseDeclaracoes varre loops sequenciais de instruções e as acumula até que feche chaves '}' ou atinja EOF.
func (p *Parser) parseDeclaracoes() ([]BaseNode, error) {
	var declaracoes []BaseNode

	for !p.fimDeArquivo() && p.token.Tipo != lexer.TokenFechaChaves {
		if p.token.Tipo == lexer.TokenNovaLinha || p.token.Tipo == lexer.TokenPontoEVirgula {
			p.avancar()
			continue
		}

		declaracao, err := p.parseDeclaracao()
		if err != nil {
			return nil, err
		}
		if declaracao != nil {
			declaracoes = append(declaracoes, declaracao)
		}

		if p.token.Tipo == lexer.TokenNovaLinha || p.token.Tipo == lexer.TokenPontoEVirgula {
			p.avancar()
		}
	}

	return declaracoes, nil
}

// parseDeclaracao processa uma única instrução lógica e registra seu token inicial no mapa de posições.
// parseDeclaracao despacha uma instrução top-level do script.
//
// Após o sucesso do parse, registra o token inicial e consome eventuais
// delimitadores finais (newline/semicolon). Em caso de erro, acumula o
// erro em p.erros e chama sincronizar para permitir a recuperação.
func (p *Parser) parseDeclaracao() (BaseNode, error) {
	tok := p.token
	res, err := p.parseDeclaracaoInterno()
	if err != nil {
		if errSintatico, ok := err.(*ErroSintatico); ok {
			p.erros = append(p.erros, *errSintatico)
		} else {
			p.erros = append(p.erros, *p.formatarErro("HRP-E101", err.Error(), ObterDicaSintatica(err.Error())))
		}
		p.sincronizar()
		return nil, nil
	}
	if res != nil {
		p.registrar(res, tok)
	}
	return res, nil
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
			return nil, p.errof("era esperado o token 'func' ou 'funcao' após 'assincrono', mas no lugar foi encontrado '%s'", p.token.Valor)
		}
		fn, err := p.parseFuncao()
		if err != nil {
			return nil, err
		}
		fn.Assincrono = true
		return fn, nil
	case lexer.TokenClasse:
		return p.parseClasse()
	case lexer.TokenEnum:
		return p.parseEnum()
	case lexer.TokenInterface:
		return p.parseInterface()
	case lexer.TokenEstilo:
		return p.parseEstilo()
	case lexer.TokenTestar:
		return p.parseTeste()
	case lexer.TokenTente:
		return p.parseTenteCapture()
	case lexer.TokenExportar:
		return p.parseExportar()
	case lexer.TokenEscolha:
		return p.parseEscolha()
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
		return &AsseguraNode{Condicao: condicao, Mensagem: mensagem}, nil
	case lexer.TokenPara:
		return p.parseBlocoPara()
	case lexer.TokenArroba:
		return p.parseAtRule()
	default:
		// Default desvia para expressões aritméticas e/ou possíveis reatribuições compostas (ex: x += 1)
		expressao, err := p.parseExpressao()
		if err != nil {
			return nil, err
		}

		token := p.token.Tipo
		if token >= lexer.TokenMaisIgual && token <= lexer.TokenBoolEIgual || token == lexer.TokenIgual {
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
		return nil, p.errof("era esperado um texto após a palavra chave 'de'")
	}

	decl := &ImporteDe{Caminho: &TextoLiteral{Valor: p.token.Valor}}
	p.avancar()

	if err := p.consome("importe"); err != nil {
		return nil, err
	}

	for {
		token := p.token

		switch token.Tipo {
		case lexer.TokenIdentificador:
			if IsKeyword(token.Valor) {
				return nil, p.errof("'%s' é uma palavra-chave reservada e não pode ser importada", token.Valor)
			}
			decl.Nomes = append(decl.Nomes, token.Valor)
			p.avancar()
		case lexer.TokenVirgula:
			p.avancar()
			continue
		default:
			if len(decl.Nomes) == 0 {
				return nil, p.errof("esperava ao menos um identificador após 'importe', mas recebi '%s'", token.Valor)
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

// parseBlocoPara analisa laços iterativos: 'para (item em sequencia) { Bloco }' ou 'para item em sequencia { Bloco }'
func (p *Parser) parseBlocoPara() (*BlocoPara, error) {
	p.consome("para")

	temParenteses := p.token.Tipo == lexer.TokenAbreParenteses
	if temParenteses {
		p.avancar()
	}

	id := p.token.Valor
	p.avancar()

	if err := p.consome("em"); err != nil {
		return nil, err
	}

	iter, err := p.parseExpressao()
	if err != nil {
		return nil, err
	}

	if temParenteses {
		if err := p.consome(")"); err != nil {
			return nil, err
		}
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

	temParenteses := p.token.Tipo == lexer.TokenAbreParenteses
	if temParenteses {
		p.avancar()
	}

	condicao, err := p.parseExpressao()
	if err != nil {
		return nil, err
	}

	expressaoSe := &ExpressaoSe{Condicao: condicao}

	if temParenteses {
		if err := p.consome(")"); err != nil {
			return nil, err
		}
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

// parseEnquanto analisa laços condicionais enquanto: 'enquanto condicao { Bloco }'
func (p *Parser) parseEnquanto() (*Enquanto, error) {
	if err := p.consome("enquanto"); err != nil {
		return nil, err
	}

	temParenteses := p.token.Tipo == lexer.TokenAbreParenteses
	if temParenteses {
		p.avancar()
	}

	condicao, err := p.parseExpressao()
	if err != nil {
		return nil, err
	}

	if temParenteses {
		if err := p.consome(")"); err != nil {
			return nil, err
		}
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

	if p.token.Tipo == lexer.TokenPontoEVirgula || p.token.Tipo == lexer.TokenNovaLinha || p.token.Tipo == lexer.TokenFechaChaves {
		if p.token.Tipo == lexer.TokenPontoEVirgula || p.token.Tipo == lexer.TokenNovaLinha {
			p.avancar()
		}
		return retorne, nil
	}

	expressao, err := p.parseExpressao()
	if err != nil {
		return nil, err
	}

	retorne.Expressao = expressao

	if p.token.Tipo == lexer.TokenPontoEVirgula || p.token.Tipo == lexer.TokenNovaLinha {
		p.avancar()
	}

	return retorne, nil
}

// parseFuncao analisa declarações de funções normais.
func (p *Parser) parseFuncao() (*DeclFuncao, error) {
	docComments := p.consumirDocComment()
	if p.token.Tipo != lexer.TokenFunc {
		return nil, p.errof("era esperado o token 'func' ou 'funcao', mas no lugar foi encontrado '%s'", p.token.Valor)
	}
	tokFunc := p.token
	p.avancar()

	funcao := &DeclFuncao{DocComments: docComments}
	p.registrar(funcao, tokFunc)

	if p.token.Tipo == lexer.TokenIdentificador {
		funcao.Nome = p.token.Valor
		p.avancar()
	} else if p.token.Tipo != lexer.TokenAbreParenteses {
		return nil, p.errof("era esperado o nome da função ou '(', mas no lugar foi encontrado '%s'", p.token.Valor)
	}

	if err := p.consome("("); err != nil {
		return nil, err
	}

	for {
		for p.token.Tipo == lexer.TokenNovaLinha {
			p.avancar()
		}

		if p.token.Tipo == lexer.TokenFechaParenteses {
			break
		}

		params, err := p.parseDeclFuncaoParametro()

		if err != nil {
			return nil, err
		}

		funcao.Parametros = append(funcao.Parametros, params)

		for p.token.Tipo == lexer.TokenNovaLinha {
			p.avancar()
		}

		if p.token.Tipo == lexer.TokenVirgula {
			p.avancar()
		}

		for p.token.Tipo == lexer.TokenNovaLinha {
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
	} else if p.token.Tipo == lexer.TokenMenos && p.próximoToken != nil && p.próximoToken.Tipo == lexer.TokenMaiorQue {
		p.avancar() // Consome '-'
		p.avancar() // Consome '>'
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
	docComments := p.consumirDocComment()
	if err := p.consome("classe"); err != nil {
		return nil, err
	}

	classe := &DeclClasse{DocComments: docComments}
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
		return nil, p.errof("esperava um texto especificando o nome do teste, obtive '%s'", p.token.Valor)
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
		return nil, p.errof("esperava 'capture' ou 'finalmente' após o bloco 'tente', mas recebi '%s'", p.token.Valor)
	}

	if p.token.Tipo == lexer.TokenCapture {
		if err := p.consome("capture"); err != nil {
			return nil, err
		}

		if err := p.consome("("); err != nil {
			return nil, err
		}

		if p.token.Tipo != lexer.TokenIdentificador {
			return nil, p.errof("esperava um identificador para o erro no bloco 'capture', mas recebi '%s'", p.token.Valor)
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
	if tok != lexer.TokenVar && tok != lexer.TokenConst && tok != lexer.TokenFunc && tok != lexer.TokenClasse && tok != lexer.TokenInterface && tok != lexer.TokenEstilo {
		return nil, p.errof("esperava declaração de 'var', 'const', 'funcao', 'classe', 'interface' ou 'estilo' após 'exportar', mas obtive '%s'", p.token.Valor)
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
func (p *Parser) parseVariavel() (BaseNode, error) {
	docComments := p.consumirDocComment()
	constante := p.token.Valor == "const"
	p.avancar()

	if p.token.Tipo == lexer.TokenAbreColchetes || p.token.Tipo == lexer.TokenAbreChaves {
		return p.parseDesestruturacao(constante)
	}

	decl := &DeclVar{Constante: constante, DocComments: docComments}
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
		return nil, p.errof("a constante '%s' deve possuir um valor inicializador", decl.Nome)
	}

	if p.token.Tipo == lexer.TokenPontoEVirgula {
		p.avancar()
	} else if p.token.Tipo == lexer.TokenNovaLinha {
		p.avancar()
	} else if p.token.Tipo != lexer.TokenFimDeArquivo && p.token.Tipo != lexer.TokenFechaChaves {
		return nil, p.errof("era esperado o token ';' ou uma nova linha ao final da declaração da variável '%s'", decl.Nome)
	}

	return decl, nil
}

func (p *Parser) parsePadraoDesestruturacao() (*DeclDesestruturacao, error) {
	decl := &DeclDesestruturacao{}
	decl.Inicio = p.token.Inicio

	fechamento := "]"
	switch p.token.Tipo {
	case lexer.TokenAbreColchetes:
		decl.TipoEstrut = "lista"
		p.avancar() // consome '['
	case lexer.TokenAbreChaves:
		decl.TipoEstrut = "mapa"
		fechamento = "}"
		p.avancar() // consome '{'
	default:
		return nil, p.errof("era esperado '[' ou '{' para desestruturação")
	}

	for p.token.Valor != fechamento && p.token.Tipo != lexer.TokenFimDeArquivo {
		if p.token.Tipo == lexer.TokenNovaLinha {
			p.avancar()
			continue
		}

		if p.token.Tipo == lexer.TokenTresPontos {
			p.avancar() // consome '...'
			if p.token.Tipo != lexer.TokenIdentificador {
				return nil, p.errof("era esperado um identificador de variável para o resto da desestruturação, mas no lugar foi encontrado '%s'", p.token.Valor)
			}
			decl.Resto = p.token.Valor
			p.avancar()
			if p.token.Tipo == lexer.TokenVirgula {
				p.avancar()
			}
			continue
		}

		if p.token.Tipo != lexer.TokenIdentificador && p.token.Tipo != lexer.TokenTexto && p.token.Tipo != lexer.TokenInteiro {
			return nil, p.errof("era esperado um identificador ou chave na desestruturação, mas no lugar foi encontrado '%s'", p.token.Valor)
		}

		nome := p.token.Valor
		p.avancar()

		var valorPadrao BaseNode
		if p.token.Tipo == lexer.TokenIgual || p.token.Tipo == lexer.TokenDoisPontos {
			p.avancar() // consome '=' ou ':'
			var err error
			valorPadrao, err = p.parseExpressao()
			if err != nil {
				return nil, err
			}
		}

		decl.Campos = append(decl.Campos, ElementoDesestruturacao{
			Nome:        nome,
			ValorPadrao: valorPadrao,
		})

		if p.token.Tipo == lexer.TokenVirgula {
			p.avancar()
		}
	}

	if err := p.consome(fechamento); err != nil {
		return nil, err
	}

	return decl, nil
}

func (p *Parser) parseDesestruturacao(constante bool) (BaseNode, error) {
	decl, err := p.parsePadraoDesestruturacao()
	if err != nil {
		return nil, err
	}
	decl.Constante = constante

	if err := p.consome("="); err != nil {
		return nil, err
	}

	valor, err := p.parseExpressao()
	if err != nil {
		return nil, err
	}
	decl.Valor = valor
	_, posFim := valor.Pos()
	decl.Fim = posFim

	switch p.token.Tipo {
	case lexer.TokenPontoEVirgula, lexer.TokenNovaLinha:
		p.avancar()
	}

	return decl, nil
}


func (p *Parser) parseEnum() (BaseNode, error) {
	docComments := p.consumirDocComment()
	p.avancar() // consome 'enum'
	if p.token.Tipo != lexer.TokenIdentificador {
		return nil, p.errof("era esperado um nome para a enumeração após 'enum'")
	}
	nomeEnum := p.token.Valor
	p.avancar()

	if err := p.consome("{"); err != nil {
		return nil, err
	}

	var valores []string
	for p.token.Tipo != lexer.TokenFechaChaves && p.token.Tipo != lexer.TokenFimDeArquivo {
		if p.token.Tipo == lexer.TokenNovaLinha {
			p.avancar()
			continue
		}
		if p.token.Tipo != lexer.TokenIdentificador {
			return nil, p.errof("era esperado um identificador de valor de enumeração, mas no lugar foi encontrado '%s'", p.token.Valor)
		}
		valores = append(valores, p.token.Valor)
		p.avancar()

		if p.token.Tipo == lexer.TokenVirgula {
			p.avancar()
		}
	}

	if err := p.consome("}"); err != nil {
		return nil, err
	}

	return &DeclEnum{Nome: nomeEnum, Valores: valores, DocComments: docComments}, nil
}

func (p *Parser) parseEscolha() (BaseNode, error) {
	p.avancar() // consome 'escolha'

	expr, err := p.parseExpressao()
	if err != nil {
		return nil, err
	}

	if err := p.consome("{"); err != nil {
		return nil, err
	}

	var casos []CasoEscolha
	for p.token.Tipo != lexer.TokenFechaChaves && p.token.Tipo != lexer.TokenFimDeArquivo {
		if p.token.Tipo == lexer.TokenNovaLinha || p.token.Tipo == lexer.TokenPontoEVirgula {
			p.avancar()
			continue
		}

		if p.token.Tipo != lexer.TokenCaso {
			return nil, p.errof("esperava 'caso' ou '}', mas encontrou '%s'", p.token.Valor)
		}
		p.avancar() // consome 'caso'

		var padrao BaseNode
		if p.token.Tipo == lexer.TokenAbreColchetes || p.token.Tipo == lexer.TokenAbreChaves {
			padrao, err = p.parsePadraoDesestruturacao()
			if err != nil {
				return nil, err
			}
		} else {
			padrao, err = p.parseExpressao()
			if err != nil {
				return nil, err
			}
		}

		var guarda BaseNode
		if p.token.Tipo == lexer.TokenQuando {
			p.avancar() // consome 'quando'
			guarda, err = p.parseExpressao()
			if err != nil {
				return nil, err
			}
		}

		corpo, err := p.parseBloco()
		if err != nil {
			return nil, err
		}

		casoNode := CasoEscolha{
			Padrao: padrao,
			Guarda: guarda,
			Corpo:  corpo,
		}
		p.registrar(&casoNode, p.token)
		casos = append(casos, casoNode)
	}

	if err := p.consome("}"); err != nil {
		return nil, err
	}

	decl := &DeclEscolha{
		Expressao: expr,
		Casos:     casos,
	}
	p.registrar(decl, p.token)
	return decl, nil
}


func (p *Parser) parseInterface() (BaseNode, error) {
	docComments := p.consumirDocComment()
	p.avancar() // consome 'interface'
	if p.token.Tipo != lexer.TokenIdentificador {
		return nil, p.errof("era esperado um nome para a interface após 'interface'")
	}
	nomeInterface := p.token.Valor
	p.avancar()

	if err := p.consome("{"); err != nil {
		return nil, err
	}

	var metodos []MetodoAssinatura
	for p.token.Tipo != lexer.TokenFechaChaves && p.token.Tipo != lexer.TokenFimDeArquivo {
		if p.token.Tipo == lexer.TokenNovaLinha {
			p.avancar()
			continue
		}

		if p.token.Tipo == lexer.TokenFunc {
			p.avancar() // consome 'func'
			if p.token.Tipo != lexer.TokenIdentificador {
				return nil, p.errof("era esperado um nome para o método da interface")
			}
			nomeMetodo := p.token.Valor
			p.avancar()

			if err := p.consome("("); err != nil {
				return nil, err
			}

			var params []string
			for p.token.Tipo != lexer.TokenFechaParenteses && p.token.Tipo != lexer.TokenFimDeArquivo {
				if p.token.Tipo == lexer.TokenIdentificador {
					params = append(params, p.token.Valor)
					p.avancar()
				}
				if p.token.Tipo == lexer.TokenVirgula {
					p.avancar()
				}
			}
			if err := p.consome(")"); err != nil {
				return nil, err
			}

			tipoRetorno := ""
			if p.token.Tipo == lexer.TokenDoisPontos {
				p.avancar()
				tipoRetorno = p.token.Valor
				p.avancar()
			}

			if p.token.Tipo == lexer.TokenPontoEVirgula {
				p.avancar()
			}

			metodos = append(metodos, MetodoAssinatura{Nome: nomeMetodo, Parametros: params, TipoRetorno: tipoRetorno})
		} else {
			p.avancar()
		}
	}

	if err := p.consome("}"); err != nil {
		return nil, err
	}

	return &DeclInterface{Nome: nomeInterface, Metodos: metodos, DocComments: docComments}, nil
}

// parseExpressao analisa o escopo geral de expressões complexas.
//
// É o entry-point público de qualquer análise de expressão: literais,
// identificadores, chamadas de função, operadores binários, ternários,
// pipes, coaléscência nula, JSX, etc. Internamente delega para
// parseExpressaoInterno, que implementa a cascata de precedência
// (ternário → pipe → coaléscência nula → disj → conj → comparação → ...
// → átomo).
func (p *Parser) parseExpressao() (BaseNode, error) {
	tok := p.token
	res, err := p.parseExpressaoInterno()
	if err == nil && res != nil {
		p.registrar(res, tok)
		// Propagate token position mapping to nested Identifiers if applicable
		if ident, ok := res.(*Identificador); ok {
			p.registrar(ident, tok)
		}
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

		return &NovaNode{Objeto: obj}, nil
	}

	if p.token.Tipo == lexer.TokenAguarde {
		p.avancar() // Consome 'aguarde'
		expr, err := p.parseExpressao()
		if err != nil {
			return nil, err
		}
		return &AguardeNode{Expressao: expr}, nil
	}

	return p.parseTernario()
}

// parseTernario resolve o operador condicional ternário 'condicao ? entao : senao'.
func (p *Parser) parseTernario() (BaseNode, error) {
	condicao, err := p.parsePipe()
	if err != nil {
		return nil, err
	}

	if p.token.Tipo == lexer.TokenInterrogacao {
		p.avancar() // Consome '?'

		entao, err := p.parseExpressao()
		if err != nil {
			return nil, err
		}

		if err := p.consome(":"); err != nil {
			return nil, err
		}

		senao, err := p.parseTernario()
		if err != nil {
			return nil, err
		}

		return &OpTernaria{Condicao: condicao, Entao: entao, Senao: senao}, nil
	}

	return condicao, nil
}

// parsePipe resolve expressões de encadeamento pipe (|>).
func (p *Parser) parsePipe() (BaseNode, error) {
	esq, err := p.parseCoalescenciaNula()
	if err != nil {
		return nil, err
	}

	for p.token.Tipo == lexer.TokenPipe {
		p.avancar() // Consome |>
		dir, err := p.parseCoalescenciaNula()
		if err != nil {
			return nil, err
		}
		esq = &OpPipe{Esq: esq, Dir: dir}
	}

	return esq, nil
}

// parseCoalescenciaNula resolve o operador de coalescência nula '??'.
func (p *Parser) parseCoalescenciaNula() (BaseNode, error) {
	esq, err := p.parseDisjuncao()
	if err != nil {
		return nil, err
	}

	for p.token.Tipo == lexer.TokenInterrogacaoInterrogacao {
		p.avancar() // Consome ??
		dir, err := p.parseDisjuncao()
		if err != nil {
			return nil, err
		}
		esq = &OpCoalescenciaNula{Esq: esq, Dir: dir}
	}

	return esq, nil
}

// parseEsqLst é a joia arquitetural do Parser.
//
// Monta operadores binários associativos à esquerda (left-associative).
// Recebe uma função 'próximo' de maior prioridade de precedência sintática e uma
// função 'proxOp' que consome e retorna o operador do nível corrente, se casar.
// Evita duplicações maciças de laços de precedência idênticos ao longo do analisador.
//
// # Padrão de uso
//
//	expr, err := p.parseEsqLst(p.parseConjuncao, func() (string, bool) {
//	    if p.token.Tipo == lexer.TokenOuLogico {
//	        op := p.token.Valor
//	        p.avancar()
//	        return op, true
//	    }
//	    return "", false
//	})
//
// # Por que associativo à esquerda?
//
// A grande maioria das operações em Harpia (a+b+c, a<b<c) segue essa
// associatividade. Operadores associativos à direita (atribuição, exponenciação
// em algumas linguagens) são tratados fora deste helper.
func (p *Parser) parseEsqLst(próximo func() (BaseNode, error), proxOp func() (string, bool)) (BaseNode, error) {
	esq, err := próximo()
	if err != nil {
		return nil, err
	}
	for {
		op, ok := proxOp()
		if !ok {
			return esq, nil
		}
		dir, err := próximo()
		if err != nil {
			return nil, err
		}
		esq = &OpBinaria{Esq: esq, Operador: op, Dir: dir}
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

		return &OpUnaria{Operador: "nao", Expressao: operacao}, nil
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
			if p.token.Valor == "instancia" && p.próximoToken.Tipo == lexer.TokenDe {
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

		return &OpUnaria{Operador: token.Valor, Expressao: expressao}, nil
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

		return &OpBinaria{Esq: esquerda, Operador: "**", Dir: direita}, nil
	}

	return esquerda, nil
}

func (p *Parser) parseListaArgumentos() ([]BaseNode, error) {
	if err := p.consome("("); err != nil {
		return nil, err
	}

	var argumentos []BaseNode
	for p.token.Tipo != lexer.TokenFechaParenteses {
		if p.token.Tipo == lexer.TokenNovaLinha {
			p.avancar()
			continue
		}

		var expressao BaseNode
		var err error

		// Verifica se é uma atribuição nomeada (identificador seguido de '=')
		if p.token.Tipo == lexer.TokenIdentificador && p.próximoToken != nil && p.próximoToken.Tipo == lexer.TokenIgual {
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

		argumentos = append(argumentos, expressao)

		if p.token.Tipo == lexer.TokenVirgula {
			p.avancar()
			continue
		}

		if p.token.Tipo != lexer.TokenFechaParenteses {
			return nil, p.errof("esperava ',' ou ')' na lista de argumentos, mas recebi '%s'", p.token.Valor)
		}
	}

	if err := p.consome(")"); err != nil {
		return nil, err
	}
	return argumentos, nil
}

// parsePrimario resolve acessos a membros (.), chamadas de funções e indexação de arrays com colchetes [].

func (p *Parser) parsePrimario() (BaseNode, error) {
	atom, err := p.parseAtomo()
	if err != nil {
		return nil, err
	}

Loop:
	for {
		switch p.token.Tipo {
		case lexer.TokenPonto, lexer.TokenPontoInterrogacao:
			ehOpcional := p.token.Tipo == lexer.TokenPontoInterrogacao
			p.avancar()
			// Tolera quebra de linha ou ';' antes de acessos encadeados (ex: obj\n.metodo())
			for p.token.Tipo == lexer.TokenNovaLinha || p.token.Tipo == lexer.TokenPontoEVirgula {
				p.avancar()
			}

			if ehOpcional && p.token.Tipo == lexer.TokenAbreParenteses {
				args, err := p.parseListaArgumentos()
				if err != nil {
					return nil, err
				}
				chamada := &ChamadaOpcional{Funcao: atom, Argumentos: args}
				p.registrar(chamada, p.token)
				atom = chamada
			} else if ehOpcional && p.token.Tipo == lexer.TokenAbreColchetes {
				p.avancar()
				arg, err := p.parseExpressao()
				if err != nil {
					return nil, err
				}
				if err := p.consome("]"); err != nil {
					return nil, err
				}
				idx := &IndexacaoOpcional{Objeto: atom, Argumento: arg}
				p.registrar(idx, p.token)
				atom = idx
			} else {
				membro, err := p.parseAtomo()
				if err != nil {
					return nil, err
				}

				if ehOpcional {
					atom = &AcessoMembroOpcional{Objeto: atom, Membro: membro}
				} else {
					atom = &AcessoMembro{Dono: atom, Membro: membro}
				}
			}
		case lexer.TokenAbreParenteses:
			chamada := &ChamadaFuncao{Identificador: atom}
			p.registrar(chamada, p.token)
			if ident, ok := atom.(*Identificador); ok {
				p.registrar(ident, p.token)
			}

			args, err := p.parseListaArgumentos()
			if err != nil {
				return nil, err
			}
			chamada.Argumentos = args
			atom = chamada
		case lexer.TokenAbreColchetes:
			p.avancar()

			arg, err := p.parseExpressao()
			if err != nil {
				return nil, err
			}

			if err := p.consome("]"); err != nil {
				return nil, err
			}

			atom = &Indexacao{Objeto: atom, Argumento: arg}
		default:
			break Loop
		}
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
		return &Identificador{Nome: token.Valor}, nil
	case lexer.TokenTexto:
		val := token.Valor
		p.avancar()

		// Se tiver chaves de interpolação { ... } dentro da string (ex: "olá {nome}")
		if strings.Contains(val, "{") && strings.Contains(val, "}") && len(val) >= 2 {
			strContent := val[1 : len(val)-1]

			// Se a string contiver trechos de código com importação/exportação (ex: "importar { x } de 'y'"), trata como texto literal
			if strings.Contains(strContent, "importar ") || strings.Contains(strContent, "exportar ") {
				return &TextoLiteral{Valor: token.Valor}, nil
			}
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

				// Se a chave for escapada com \, ignora interpolação e trata como texto
				if idxOpen > 0 && strContent[idxOpen-1] == '\\' {
					part := strContent[i : idxClose+1]
					templateNode.Partes = append(templateNode.Partes, &TextoLiteral{Valor: "\"" + part + "\""})
					i = idxClose + 1
					continue
				}

				exprStr := strContent[idxOpen+1 : idxClose]
				// ponytail: parser recursivo isolado por template. Evita implementar gramática de strings complexa com lookahead no parser principal.
				subParser := NewParserFromString(exprStr, p.arquivo)
				exprNode, err := subParser.parseExpressao()
				if err != nil || subParser.token.Tipo != lexer.TokenFimDeArquivo {
					// Fallback: se o conteúdo entre chaves não for uma expressão válida completa, trata como texto literal
					return &TextoLiteral{Valor: token.Valor}, nil
				}

				templateNode.Partes = append(templateNode.Partes, &TemplateExpr{Expressao: exprNode})
				i = idxClose + 1
			}
			return templateNode, nil
		}
		return &TextoLiteral{Valor: token.Valor}, nil
	case lexer.TokenDecimal:
		p.avancar()
		return &DecimalLiteral{Valor: token.Valor}, nil
	case lexer.TokenInteiro:
		p.avancar()
		return &InteiroLiteral{Valor: token.Valor}, nil
	case lexer.TokenIdentificador, lexer.TokenSelf:
		if !IsKeyword(token.Valor) || token.Tipo == lexer.TokenSelf {
			p.avancar()
			return p.registrar(&Identificador{Nome: token.Valor}, token), nil
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
			if p.token.Tipo == lexer.TokenNovaLinha {
				p.avancar()
				continue
			}

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

	linha, coluna := 0, 0
	if p.token.Inicio.Linha > 0 {
		linha = p.token.Inicio.Linha
		coluna = p.token.Inicio.Coluna
	}
	return nil, p.errof("o token '%v' não é reconhecido na linha %d, coluna %d", p.token.Valor, linha, coluna)
}

// parseMapa analisa a declaração de dicionários lógicos chave-valor (mapas): '{ chave: valor }'
func (p *Parser) parseMapa() (*MapaLiteral, error) {
	mapa := &MapaLiteral{}
	if err := p.consome("{"); err != nil {
		return nil, err
	}

	for p.token.Tipo != lexer.TokenFechaChaves && p.token.Tipo != lexer.TokenFimDeArquivo {
		if p.token.Tipo == lexer.TokenNovaLinha {
			p.avancar()
			continue
		}

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
		return nil, p.errof("esperava um identificador após 'estilo', mas recebi '%s'", p.token.Valor)
	}
	nome := p.token.Valor
	p.avancar()

	// Consome modificadores/pseudo-elementos adicionais no seletor (ex: ::placeholder, :hover, .subclasse) até '{'
	for p.token.Tipo != lexer.TokenAbreChaves && !p.fimDeArquivo() {
		nome += p.token.Valor
		p.avancar()
	}

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

// parseAtRule analisa regras CSS como @media, @keyframes, @import, etc.
func (p *Parser) parseAtRule() (*DeclEstilo, error) {
	nomeAt := p.token.Valor // "@"
	p.avancar()
	if p.token.Tipo == lexer.TokenIdentificador {
		nomeAt += p.token.Valor // "@media"
		p.avancar()
	}

	// Consome seletores e parâmetros até '{' ou fim de instrução
	for p.token.Tipo != lexer.TokenAbreChaves && p.token.Tipo != lexer.TokenPontoEVirgula && p.token.Tipo != lexer.TokenNovaLinha && !p.fimDeArquivo() {
		nomeAt += " " + p.token.Valor
		p.avancar()
	}

	if p.token.Tipo == lexer.TokenAbreChaves {
		p.avancar() // consome '{'
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
		if p.token.Valor == "}" {
			p.avancar()
		}
		return &DeclEstilo{Nome: strings.TrimSpace(nomeAt), Regras: strings.TrimSpace(regras.String())}, nil
	}

	if p.token.Tipo == lexer.TokenPontoEVirgula || p.token.Tipo == lexer.TokenNovaLinha {
		p.avancar()
	}

	return &DeclEstilo{Nome: strings.TrimSpace(nomeAt), Regras: ""}, nil
}

// parseJSX analisa um elemento JSX ou uma tag de controle (se/para)
func (p *Parser) parseJSX() (BaseNode, error) {
	if err := p.consome("<"); err != nil {
		return nil, err
	}

	tag := p.token.Valor
	p.avancar()

	switch tag {
	case "se":
		return p.parseSeJSX()
	case "para":
		return p.parseParaJSX()
	}

	node := &NoJSX{Tag: tag}

	// Analisa atributos/propriedades da tag
	for p.token.Tipo != lexer.TokenMaiorQue && !(p.token.Valor == "/" && p.próximoToken != nil && (p.próximoToken.Tipo == lexer.TokenMaiorQue || p.próximoToken.Valor == ">")) && !p.fimDeArquivo() {
		if p.token.Valor == "/" {
			break
		}
		if p.token.Tipo != lexer.TokenMaiorQue && p.token.Valor != "=" {
			nomeAttr := p.token.Valor
			p.avancar()

			var valorAttr BaseNode
			if p.token.Tipo == lexer.TokenIgual {
				p.avancar()
				switch p.token.Tipo {
				case lexer.TokenAbreChaves:
					p.avancar() // consome '{'
					expr, err := p.parseExpressao()
					if err != nil {
						return nil, err
					}
					if err := p.consome("}"); err != nil {
						return nil, err
					}
					valorAttr = expr
				case lexer.TokenTexto:
					nodeTexto, err := p.parseAtomo()
					if err != nil {
						return nil, err
					}
					valorAttr = nodeTexto
				default:
					return nil, p.errof("esperava string ou expressão em '{}' para o atributo '%s', mas recebi '%s'", nomeAttr, p.token.Valor)
				}
			} else {
				// Atributo booleano implícito (ex: <input desabilitado />)
				valorAttr = &Identificador{Nome: "Verdadeiro"}
			}
			node.Atributos = append(node.Atributos, &AtributoJSX{Nome: nomeAttr, Valor: valorAttr})
		} else {
			return nil, p.errof("esperava nome de atributo, mas recebi '%s'", p.token.Valor)
		}
	}

	if p.token.Valor == "/" {
		p.avancar() // consome '/'
		if p.token.Tipo == lexer.TokenMaiorQue || p.token.Valor == ">" {
			p.avancar() // consome '>'
		}
		node.AutoFechado = true
		return node, nil
	}

	if err := p.consome(">"); err != nil {
		return nil, err
	}

	fechouTag := false
	for !p.fimDeArquivo() {
		// Se encontrarmos "</", é o início do fechamento da tag
		if p.token.Tipo == lexer.TokenMenorQue && p.próximoToken != nil && p.próximoToken.Valor == "/" {
			p.avancar() // '<'
			p.avancar() // '/'
			if p.token.Valor != tag {
				return nil, p.errof("tag de fechamento incorreta: esperava '</%s>', mas recebi '</%s>'", tag, p.token.Valor)
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

var reComentarioHTML = regexp.MustCompile(`<!--[\s\S]*?-->`)
var reComentarioLinhaJSX = regexp.MustCompile(`(?m)^\s*(?:#|//).*$`)

// Caso contrário, é texto literal livre
		inicioTexto := p.ultimoTokenFimIndice
		for p.token.Tipo != lexer.TokenMenorQue && p.token.Tipo != lexer.TokenAbreChaves && !p.fimDeArquivo() {
			p.avancar()
		}
		fimTexto := p.token.Inicio.Indice
		if fimTexto < inicioTexto {
			fimTexto = inicioTexto
		}
		rawText := p.lex.SubString(inicioTexto, fimTexto)
		rawText = reComentarioHTML.ReplaceAllString(rawText, "")
		rawText = reComentarioLinhaJSX.ReplaceAllString(rawText, "")
		if len(strings.TrimSpace(rawText)) > 0 {
			temEspacoInicio := strings.HasPrefix(rawText, " ") || strings.HasPrefix(rawText, "\t") || strings.HasPrefix(rawText, "\n")
			temEspacoFim := strings.HasSuffix(rawText, " ") || strings.HasSuffix(rawText, "\t") || strings.HasSuffix(rawText, "\n")

			campos := strings.Fields(rawText)
			textoUnido := strings.Join(campos, " ")
			if temEspacoInicio {
				textoUnido = " " + textoUnido
			}
			if temEspacoFim {
				textoUnido = textoUnido + " "
			}
			node.Filhos = append(node.Filhos, &TextoLiteral{Valor: "\"" + textoUnido + "\""})
		}
	}

	if !fechouTag {
		return nil, p.errof("tag '<%s>' não foi fechada correspondente", tag)
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
		return nil, p.errof("esperava atributo 'condicao' na tag <se>, mas recebi '%s'", p.token.Valor)
	}

	if err := p.consome(">"); err != nil {
		return nil, err
	}

	node := &NoSeJSX{Condicao: condicao}

	fechouTag := false
	for !p.fimDeArquivo() {
		if p.token.Tipo == lexer.TokenMenorQue && p.próximoToken != nil && p.próximoToken.Valor == "/" {
			p.avancar() // '<'
			p.avancar() // '/'
			if p.token.Valor != "se" {
				return nil, p.errof("tag de fechamento incorreta para <se>: recebido '</%s>'", p.token.Valor)
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
		return nil, p.errof("tag '<se>' não foi fechada correspondente")
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
		return nil, p.errof("esperava nome de variável após '<para', mas recebi '%s'", p.token.Valor)
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
		return nil, p.errof("esperava atributo 'lista' na tag <para>, mas recebi '%s'", p.token.Valor)
	}

	// ponytail: consome qualquer atributo adicional na tag <para> (ex: chave="id")
	for p.token.Tipo != lexer.TokenMaiorQue && !p.fimDeArquivo() {
		p.avancar()
	}

	if err := p.consome(">"); err != nil {
		return nil, err
	}

	node := &NoParaJSX{Item: item, Lista: lista}

	fechouTag := false
	for !p.fimDeArquivo() {
		if p.token.Tipo == lexer.TokenMenorQue && p.próximoToken != nil && p.próximoToken.Valor == "/" {
			p.avancar() // '<'
			p.avancar() // '/'
			if p.token.Valor != "para" {
				return nil, p.errof("tag de fechamento incorreta para <para>: recebido '</%s>'", p.token.Valor)
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
		return nil, p.errof("tag '<para>' não foi fechada correspondente")
	}

	return node, nil
}

// ErroSintatico representa um erro de parsing formatado com contexto visual.
//
// Todos os campos são exportados para permitir inspeção programática por
// ferramentas externas (CLI, TUI, IDE, linter). A renderização amigável em
// terminal é feita pelo método String, que aplica cores ANSI quando o
// terminal suporta.
//
// # Campos principais
//
//   - Mensagem: descrição legível do erro.
//   - Codigo: código de erro estilo "HRP-E101" para indexação em docs.
//   - Inicio/Fim: coordenadas do token ofensivo.
//   - LinhaFonte: linha do código onde o erro ocorreu (sem indentação).
//   - Indicador: espaços seguidos de "^^^^^^^" apontando para o token.
//   - Dica: sugestão de correção em português.
type ErroSintatico struct {
	Mensagem    string
	Inicio      lexer.PosicaoToken
	Fim         lexer.PosicaoToken
	Token       *lexer.Token
	Arquivo     string
	Codigo      string
	LinhaFonte  string
	Indicador   string
	Dica        string
	CodigoFonte string
}

// Erros retorna a lista completa de erros sintáticos acumulados.
//
// Permite que consumidores (CLI, TUI, linter) inspecionem todos os erros de
// uma única execução, em vez de abortar no primeiro.
func (p *Parser) Erros() []ErroSintatico {
	return p.erros
}

// sincronizar é o mecanismo clássico de recuperação de erros em Parsers.
//
// Diante de um erro, ele consome tokens até encontrar uma fronteira lógica
// segura (como ; ou \n) onde é plausível continuar o parsing sem corromper o
// resto do arquivo. É o que permite que múltiplos erros sejam reportados em
// uma única execução do Parser (em vez de parar no primeiro).
//
// # Fronteiras reconhecidas
//
//   - TokenNovaLinha (final de linha física).
//   - Ponto-e-vírgula (;) explícito.
//   - Palavras-chave de início de instrução (se, senao, enquanto, para, var,
//     const, funcao, retornar, escolha, tente).
//   - EOF.
func (p *Parser) sincronizar() {
	p.avancar()
	for !p.fimDeArquivo() {
		if p.token.Tipo == lexer.TokenPontoEVirgula || p.token.Tipo == lexer.TokenNovaLinha {
			p.avancar()
			return
		}
		switch p.token.Tipo {
		case lexer.TokenVar, lexer.TokenConst, lexer.TokenFunc, lexer.TokenSe,
			lexer.TokenEnquanto, lexer.TokenPara, lexer.TokenRetorne, lexer.TokenTente:
			return
		}
		p.avancar()
	}
}

// Error renderiza o erro sintático com informações ricas de posição: arquivo, linha, coluna
// e destaca o trecho de código culpado — similar ao renderizador do Erro em hrp/erros.go.
// Error implementa a interface error do Go.
//
// Renderiza a mensagem formatada com cores ANSI quando o terminal suporta:
//
//	HRP-E101  arquivo.hrp:linha 12, coluna 5
//	  const x 10
//	      ^^^^^^
//	  dica: Esperava-se '=' após o nome da constante em declaração 'const'.
//
// Implementação de error permite que *ErroSintatico seja retornado por
// qualquer função que retorne error, mantendo compatibilidade total com o
// ecossistema Go (errors.As, errors.Is, fmt.Errorf wrapping).
func (e *ErroSintatico) Error() string {
	corErro := "\x1b[1;31m"
	corCodigo := "\x1b[1;37m"
	corPonteiro := "\x1b[1;36m"
	corDica := "\x1b[1;33m"
	corReset := "\x1b[0m"

	arquivoStr := e.Arquivo
	if arquivoStr == "" {
		arquivoStr = "<desconhecido>"
	}

	codigoErro := e.Codigo
	if codigoErro == "" {
		codigoErro = "HRP-E101"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("\n%serro[%s]%s: %sSintaxeErro%s — %s\n",
		corErro, codigoErro, corReset, corCodigo, corReset, e.Mensagem))

	if e.Token != nil && e.Token.Inicio.Linha > 0 {
		linha := e.Token.Inicio.Linha
		coluna := e.Token.Inicio.Coluna

		sb.WriteString(fmt.Sprintf("  %s┌──>%s %s:%d:%d\n", corPonteiro, corReset, arquivoStr, linha, coluna))
		sb.WriteString(fmt.Sprintf("  %s│%s\n", corPonteiro, corReset))

		linhaTexto := e.LinhaFonte
		if linhaTexto == "" && e.CodigoFonte != "" {
			linhas := strings.Split(e.CodigoFonte, "\n")
			idxLinha := linha - 1
			if idxLinha >= 0 && idxLinha < len(linhas) {
				linhaTexto = linhas[idxLinha]
			}
		}

		if linhaTexto != "" {
			sb.WriteString(fmt.Sprintf("%s%d │%s %s\n", corPonteiro, linha, corReset, linhaTexto))

			sb.WriteString(fmt.Sprintf("  %s│%s ", corPonteiro, corReset))
			if e.Indicador != "" {
				sb.WriteString(fmt.Sprintf("%s%s%s\n", corErro, e.Indicador, corReset))
			} else {
				tamanho := 1
				if len(e.Token.Valor) > 0 && e.Token.Valor != "\n" {
					tamanho = len(e.Token.Valor)
				}
				sb.WriteString(fmt.Sprintf("%s%s%s\n", corErro, strings.Repeat("^", tamanho), corReset))
			}
		}
		sb.WriteString(fmt.Sprintf("  %s│%s\n", corPonteiro, corReset))
	} else {
		sb.WriteString(fmt.Sprintf("  %s┌──>%s %s\n", corPonteiro, corReset, arquivoStr))
		sb.WriteString(fmt.Sprintf("  %s│%s %s%s%s\n", corPonteiro, corReset, corErro, e.Mensagem, corReset))
	}

	if e.Dica != "" {
		sb.WriteString(fmt.Sprintf("  %sdica:%s %s\n", corDica, corReset, e.Dica))
	}

	return sb.String()
}

func ObterDicaSintatica(msg string) string {
	msgLower := strings.ToLower(msg)
	switch {
	case strings.Contains(msgLower, "esperado o token ';'"):
		return "Adicione ';' ou uma quebra de linha para encerrar a instrução."
	case strings.Contains(msgLower, "esperado o token"):
		return "Verifique a sintaxe e certifique-se de incluir os delimitadores ou pontuações esperadas."
	case strings.Contains(msgLower, "constante") && strings.Contains(msgLower, "inicializador"):
		return "Toda constante deve ser declarada com um valor inicial (ex: const X = 10)."
	case strings.Contains(msgLower, "capture") || strings.Contains(msgLower, "finalmente"):
		return "Um bloco 'tente' exige ao menos um bloco 'capture (erro)' ou 'finalmente'."
	default:
		return "Verifique a sintaxe ao redor do token indicado."
	}
}

// formatarErro constrói um *ErroSintatico com coordenadas precisas do token corrente.
//
// Inclui a linha de código fonte (LinhaFonte), um indicador visual com
// circunflexos (^^^^^^^^^^^) apontando para o token ofensivo (Indicador) e
// uma Dica contextual derivada da mensagem via ObterDicaSintatica.
//
// # Renderização
//
// O método String() do ErroSintatico aplica cores ANSI quando o terminal
// suporta, produzindo uma saída amigável:
//
//	HRP-E101  linha 12, coluna 5
//	  const x 10
//	      ^^^^^^
//	  dica: Esperava-se '=' após o nome da constante em declaração 'const'.
func (p *Parser) formatarErro(codigo string, msg string, dica string) *ErroSintatico {
	var ini, fim lexer.PosicaoToken
	if p.token != nil {
		ini = p.token.Inicio
		fim = p.token.Fim
	}

	linhaFonte := ""
	indicador := ""
	if p.token != nil && p.token.Inicio.Linha > 0 && p.codigo != "" {
		linhas := strings.Split(p.codigo, "\n")
		idxLinha := p.token.Inicio.Linha - 1
		if idxLinha >= 0 && idxLinha < len(linhas) {
			linhaFonte = linhas[idxLinha]
			col := p.token.Inicio.Coluna
			if col < 1 {
				col = 1
			}
			tamanho := len(p.token.Valor)
			if tamanho < 1 || p.token.Valor == "\n" {
				tamanho = 1
			}

			var sb strings.Builder
			for j := 0; j < len(linhaFonte) && j < col-1; j++ {
				if linhaFonte[j] == '\t' {
					sb.WriteRune('\t')
				} else {
					sb.WriteRune(' ')
				}
			}
			sb.WriteString(strings.Repeat("^", tamanho))
			indicador = sb.String()
		}
	}
	if indicador == "" {
		indicador = "^"
	}

	if codigo == "" {
		codigo = "HRP-E101"
	}

	if dica == "" {
		dica = ObterDicaSintatica(msg)
	}

	return &ErroSintatico{
		Mensagem:    msg,
		Inicio:      ini,
		Fim:         fim,
		Token:       p.token,
		Arquivo:     p.arquivo,
		Codigo:      codigo,
		LinhaFonte:  linhaFonte,
		Indicador:   indicador,
		Dica:        dica,
		CodigoFonte: p.codigo,
	}
}

// erro é o atalho padrão para criar um erro sintático com código HRP-E101 e
// dica derivada da mensagem.
//
// É o método de uso preferencial dentro do Parser para sinalizar erros de
// parsing — raramente é necessário construir um ErroSintatico manualmente.
func (p *Parser) erro(msg string) error {
	return p.formatarErro("HRP-E101", msg, ObterDicaSintatica(msg))
}

// errof é a variante formatada de erro, equivalente a fmt.Errorf.
//
// Útil quando o erro depende de valores runtime (ex: identificador recebido
// vs. esperado): p.errof("esperado '%s', encontrado '%s'", esperado, atual).
func (p *Parser) errof(format string, args ...interface{}) error {
	return p.erro(fmt.Sprintf(format, args...))
}
