package tests

import (
	"strings"
	"testing"

	"github.com/mat-dgruber/Harpia/hrp"
	_ "github.com/mat-dgruber/Harpia/stdlib"
)

func TestSSRBasico(t *testing.T) {
	codigo := `
	funcao sinal(v) { retorne [funcao() { retorne v }] }
	var res = sinal(10)
	var contador = res[0]
	var elemento = <div classe="p-4" id="principal">
		<h1>Contador: {contador()}</h1>
	</div>
	var html = elemento
	`
	ctx := hrp.NewContexto(hrp.OpcsContexto{})
	defer ctx.Terminar()

	ast, err := ctx.StringParaAst(strings.ReplaceAll(codigo, "\r", ""), "<teste>")
	if err != nil {
		t.Fatalf("Erro ao compilar para AST: %v", err)
	}

	interpretador := &hrp.Interpretador{
		Ast:      ast,
		Contexto: ctx,
		Escopo:   hrp.NewEscopo(),
	}
	_, err = interpretador.Inicializa()
	if err != nil {
		t.Fatalf("Erro ao executar AST: %v", err)
	}

	simbolo, err := interpretador.Escopo.ObterValor("html")
	if err != nil {
		t.Fatalf("Erro ao obter simbolo 'html': %v", err)
	}

	elementoJSX, ok := simbolo.(*hrp.ElementoJSX)
	if !ok {
		t.Fatalf("Esperava *hrp.ElementoJSX, obtido %T", simbolo)
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

func TestSSRElementosCustomizados(t *testing.T) {
	codigo := `
	funcao Botao(props) {
		var texto = props["texto"] ? props["texto"] : "Clique"
		retorne <button classe="btn">{texto}</button>
	}

	funcao Layout(props) {
		retorne <div classe="layout">
			<header>Menu</header>
			<main>{props.children}</main>
		</div>
	}

	var elemento = <Layout>
		<Botao texto="Enviar" />
	</Layout>

	var html = elemento
	`
	ctx := hrp.NewContexto(hrp.OpcsContexto{})
	defer ctx.Terminar()

	ast, err := ctx.StringParaAst(strings.ReplaceAll(codigo, "\r", ""), "<teste>")
	if err != nil {
		t.Fatalf("Erro ao compilar para AST: %v", err)
	}

	interpretador := &hrp.Interpretador{
		Ast:      ast,
		Contexto: ctx,
		Escopo:   hrp.NewEscopo(),
	}
	_, err = interpretador.Inicializa()
	if err != nil {
		t.Fatalf("Erro ao executar AST: %v", err)
	}

	simbolo, err := interpretador.Escopo.ObterValor("html")
	if err != nil {
		t.Fatalf("Erro ao obter simbolo 'html': %v", err)
	}

	elementoJSX, ok := simbolo.(*hrp.ElementoJSX)
	if !ok {
		t.Fatalf("Esperava *hrp.ElementoJSX, obtido %T", simbolo)
	}

	htmlStr := elementoJSX.RenderizarHTML()

	if !strings.Contains(htmlStr, `class="layout"`) {
		t.Errorf("Esperava classe 'layout' no HTML, obtido: %s", htmlStr)
	}
	if !strings.Contains(htmlStr, `<header>Menu</header>`) {
		t.Errorf("Esperava header no HTML, obtido: %s", htmlStr)
	}
	if !strings.Contains(htmlStr, `class="btn"`) {
		t.Errorf("Esperava classe 'btn' no HTML, obtido: %s", htmlStr)
	}
	if !strings.Contains(htmlStr, `Enviar`) {
		t.Errorf("Esperava texto 'Enviar' no HTML, obtido: %s", htmlStr)
	}
}
