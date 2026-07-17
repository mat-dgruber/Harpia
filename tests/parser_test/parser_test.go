package parser_test

import (
	"testing"

	"github.com/mat-dgruber/Harpia/parser"
)

func TestParseEstilo(t *testing.T) {
	codigo := `
	estilo MeuComponente {
		corDeFundo: "azul";
		botao:hover {
			opacidade: 0.8;
		}
	}
	`
	p := parser.NewParserFromString(codigo, "<teste>")
	prog, err := p.Parse()
	if err != nil {
		t.Fatalf("Erro inesperado ao analisar bloco 'estilo': %v", err)
	}

	if len(prog.Declaracoes) != 1 {
		t.Fatalf("Esperava 1 declaração, mas recebi %d", len(prog.Declaracoes))
	}

	decl, ok := prog.Declaracoes[0].(*parser.DeclEstilo)
	if !ok {
		t.Fatalf("Esperava nó *parser.DeclEstilo, mas recebi %T", prog.Declaracoes[0])
	}

	if decl.Nome != "MeuComponente" {
		t.Errorf("Nome esperado: 'MeuComponente', recebido: '%s'", decl.Nome)
	}
}

func TestParseJSX(t *testing.T) {
	codigo := `
	var elemento = <div classe="p-4" ativo>
		<h1> Olá Mundo </h1>
		<se condicao={exibir()}>
			<span> Ativo </span>
		</se>
		<para item em lista={dados}>
			<p> Item: {item} </p>
		</para>
		<img url="foto.png" />
	</div>
	`
	p := parser.NewParserFromString(codigo, "<teste>")
	prog, err := p.Parse()
	if err != nil {
		t.Fatalf("Erro inesperado ao analisar JSX-like: %v", err)
	}

	if len(prog.Declaracoes) != 1 {
		t.Fatalf("Esperava 1 declaração (var), mas recebi %d", len(prog.Declaracoes))
	}

	declVar, ok := prog.Declaracoes[0].(*parser.DeclVar)
	if !ok {
		t.Fatalf("Esperava nó *parser.DeclVar, mas recebi %T", prog.Declaracoes[0])
	}

	jsx, ok := declVar.Inicializador.(*parser.NoJSX)
	if !ok {
		t.Fatalf("Esperava inicializador *parser.NoJSX, mas recebi %T", declVar.Inicializador)
	}

	if jsx.Tag != "div" {
		t.Errorf("Tag esperada: 'div', recebida: '%s'", jsx.Tag)
	}

	if len(jsx.Atributos) != 2 {
		t.Errorf("Esperava 2 atributos, recebido: %d", len(jsx.Atributos))
	}

	if jsx.Atributos[0].Nome != "classe" {
		t.Errorf("Atributo 1 esperado: 'classe', recebido: '%s'", jsx.Atributos[0].Nome)
	}

	if jsx.Atributos[1].Nome != "ativo" {
		t.Errorf("Atributo 2 esperado: 'ativo', recebido: '%s'", jsx.Atributos[1].Nome)
	}
}
