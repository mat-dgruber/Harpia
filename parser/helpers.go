package parser

import (
	"encoding/json"
)

// palavrasChave define a relação interna e exclusiva de palavras reservadas
// que possuem significado semântico fixo na sintaxe do Portuscript.
var palavrasChave = map[string]bool{
	"se":       true,
	"senao":    true,
	"enquanto": true,
	"const":    true,
	"var":      true,
	"func":     true,
	"pare":     true,
	"continue": true,
	"para":     true,
	"nova":     true,
	"assegura": true,
}

// IsKeyword verifica de forma reativa se a string informada é uma palavra reservada protegida do sistema.
//
// Retorna verdadeiro se for uma palavra-chave registrada, impedindo que ela seja utilizada livremente
// pelo programador como um nome identificador ordinário de variável ou função (gerando erro de sintaxe).
func IsKeyword(s string) bool {
	_, ok := palavrasChave[s]
	return ok
}

// Ast2string é um utilitário crucial de diagnóstico e depuração.
//
// Ela recebe um nó da Árvore de Sintaxe Abstrata (BaseNode) e realiza a serialização estruturada
// em formato JSON identado com 4 espaços. Isso permite exportar ou imprimir grafos de AST completos
// no console de forma humana e legível em tempo de projeto e testes.
func Ast2string(ast BaseNode) ([]byte, error) {
	return json.MarshalIndent(ast, "", "    ")
}
