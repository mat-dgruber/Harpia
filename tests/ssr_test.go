package tests

import (
	"strings"
	"testing"

	"github.com/natanfeitosa/portuscript/ptst"
)

func TestSSRBasico(t *testing.T) {
	codigo := `
	var res = sinal(10)
	var contador = res[0]
	var elemento = <div classe="p-4" id="principal">
		<h1>Contador: {contador()}</h1>
	</div>
	var html = elemento
	`
	ctx := ptst.NewContexto(ptst.OpcsContexto{})
	defer ctx.Terminar()

	ast, err := ctx.StringParaAst(strings.ReplaceAll(codigo, "\r", ""), "<teste>")
	if err != nil {
		t.Fatalf("Erro ao compilar para AST: %v", err)
	}

	interpretador := &ptst.Interpretador{
		Ast:      ast,
		Contexto: ctx,
		Escopo:   ptst.NewEscopo(),
	}
	_, err = interpretador.Inicializa()
	if err != nil {
		t.Fatalf("Erro ao executar AST: %v", err)
	}

	simbolo, err := interpretador.Escopo.ObterValor("html")
	if err != nil {
		t.Fatalf("Erro ao obter simbolo 'html': %v", err)
	}

	elementoJSX, ok := simbolo.(*ptst.ElementoJSX)
	if !ok {
		t.Fatalf("Esperava *ptst.ElementoJSX, obtido %T", simbolo)
	}

	if elementoJSX.Tag != "div" {
		t.Errorf("Tag esperada: 'div', obtida: '%s'", elementoJSX.Tag)
	}

	htmlStr := elementoJSX.RenderizarHTML()
	if !strings.Contains(htmlStr, `class="p-4"`) || !strings.Contains(htmlStr, `id="principal"`) {
		t.Errorf("HTML esperado conter class e id, obtido: %s", htmlStr)
	}

	if !strings.Contains(htmlStr, "Contador") || !strings.Contains(htmlStr, "10") {
		t.Errorf("HTML esperado conter o h1 com o valor do sinal 10, obtido: %s", htmlStr)
	}
}
