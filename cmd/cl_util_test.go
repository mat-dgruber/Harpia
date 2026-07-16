package cmd

import (
	"strings"
	"testing"
)

// ponytail: assevera o funcionamento de UTIL-1 de forma independente e limpa.
func TestExtraiEGerarCssUtilitarios(t *testing.T) {
	jsInput := `
		let el = h('div', { classe: "flex-linha itens-centro p-4 classe-desconhecida" }, "Ola");
		let botao = h('botao', { classe: "fundo-azul texto-branco px-2 py-1" }, "Clique");
	`

	css := extraiEGerarCssUtilitarios(jsInput)

	// Deve conter as utilitárias em português mapeadas
	classesEsperadas := []string{
		".flex-linha",
		".itens-centro",
		".p-4",
		".fundo-azul",
		".texto-branco",
		".px-2",
		".py-1",
	}

	for _, classe := range classesEsperadas {
		if !strings.Contains(css, classe) {
			t.Errorf("Esperava que o utilitário CSS gerasse regra para '%s', mas não foi encontrado no CSS gerado:\n%s", classe, css)
		}
	}

	// Não deve gerar regras para a classe desconhecida
	if strings.Contains(css, ".classe-desconhecida") {
		t.Errorf("Utilitário CSS incorretamente gerou uma classe desconhecida '.classe-desconhecida'")
	}
}
