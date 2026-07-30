// Package parser fornece o analisador sintático (parser) da linguagem Harpia.
//
// Ele consome a sequência de tokens produzida pelo pacote lexer e a transforma
// em uma Árvore de Sintaxe Abstrata (AST) fortemente tipada, pronta para ser
// avaliada pela máquina virtual (VM) ou serializada para diagnóstico.
//
// # Algoritmo
//
// O parser utiliza descida recursiva manual (_recursive descent_) organizada em
// camadas de precedência (ver Parser.parseEsqLst). Tokens são consumidos
// através de um cursor duplo (token corrente + lookahead) gerenciado pelos
// métodos avançar, consome e sincronizar do Parser.
//
// # Erros e diagnósticos
//
// Erros sintáticos são acumulados em ErroSintatico, que preserva coordenadas
// precisas de linha/coluna e renderização colorida em terminal (HRP-E101).
//
// # Nota de Segurança / Uso
//
// Este arquivo é parte do compilador Harpia. Não é código malicioso — é um
// módulo de software aberto cuja única finalidade é converter código-fonte
// Harpia em uma estrutura de dados intermediária (AST).
package parser

import (
	"encoding/json"
)

// palavrasChave define o conjunto fechado de palavras reservadas da linguagem Harpia.
//
// Manter essa lista sincronizada com o lexer e o linter (cmd/checar.go) é
// essencial para impedir que o programador utilize acidentalmente nomes
// sintáticos como identificadores comuns.
//
// # Por que não no lexer?
//
// Embora o lexer identifique _tokens_ do tipo palavra-chave, a verificação
// semântica (rejeitar o uso de uma palavra-chave como nome de variável) é
// responsabilidade deste pacote, pois o lexer não distingue contexto de uso.
var palavrasChave = map[string]bool{
	"se":       true, // Abertura de bloco condicional (ex: se x { ... }).
	"senao":    true, // Ramo alternativo de ExpressaoSe (ex: } senao { ... }).
	"enquanto": true, // Laço condicional pré-testado.
	"const":    true, // Declaração de variável imutável (não permite reatribuição).
	"var":      true, // Declaração de variável mutável.
	"func":     true, // Declaração de função ou método.
	"pare":     true, // Quebra imediata de laço estruturado (break).
	"continue": true, // Avanço para a próxima iteração de laço.
	"para":     true, // Laço iterativo sobre coleção (for-in / foreach).
	"nova":     true, // Instanciação explícita de classes (ex: nova Pessoa()).
	"assegura": true, // Asserção em tempo de execução (debug, contratos).
}

// IsKeyword verifica se a string informada é uma palavra-chave reservada da linguagem.
//
// Esta função é consultada durante a análise de átomos (Parser.parseAtomo) e
// durante a importação seletiva (Parser.parseImporteDe) para impedir que
// símbolos reservados sejam introduzidos no escopo do programa.
//
// # Retorno
//
//   - true:  o identificador é uma palavra-chave e não pode ser usado livremente.
//   - false: o identificador é um nome comum válido.
func IsKeyword(s string) bool {
	_, ok := palavrasChave[s]
	return ok
}

// Ast2string serializa uma sub-árvore da AST em JSON identado para fins de
// diagnóstico e depuração.
//
// O JSON resultante é identado com 4 espaços e preserva a hierarquia completa
// de campos públicos dos nós envolvidos, incluindo NoBase.PosicaoToken. É a
// ferramenta padrão para inspeção visual da árvore gerada pelo Parser.
//
// # Quando usar
//
//   - Em testes unitários para validar a estrutura sintática produzida.
//   - Em logs estruturados durante a compilação.
//   - Na CLI `Harpia compilar --dump-ast` (quando habilitada).
//
// # Erros
//
// Retorna o erro retornado por encoding/json.MarshalIndent caso algum nó
// contenha tipos não serializáveis (em prática, isso só ocorre se um nó
// customizado contiver campos não-exportados de tipos exóticos).
func Ast2string(ast BaseNode) ([]byte, error) {
	return json.MarshalIndent(ast, "", "    ")
}
