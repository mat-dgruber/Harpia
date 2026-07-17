package parser_test

import (
	"testing"

	"github.com/mat-dgruber/Harpia/parser"
)

// 1 - 2 - 3 deve parsear como (1 - 2) - 3 = -4 (left-associative).
func TestAssociatividadeSubtracao(t *testing.T) {
	recebido, err := createParser("1 - 2 - 3").Parse()
	if err != nil {
		t.Fatal(err)
	}

	if len(recebido.Declaracoes) != 1 {
		t.Fatalf("esperava 1 declaracao, recebi %d", len(recebido.Declaracoes))
	}

	obtido, ok := recebido.Declaracoes[0].(*parser.OpBinaria)
	if !ok {
		t.Fatalf("esperava OpBinaria, recebi %T", recebido.Declaracoes[0])
	}

	esqExterno, ok := obtido.Esq.(*parser.OpBinaria)
	if !ok || esqExterno.Operador != "-" {
		t.Fatalf("esperava associatividade à esquerda na subtracao")
	}
}

// 8 / 4 / 2 deve parsear como (8 / 4) / 2 = 1.
func TestAssociatividadeDivisao(t *testing.T) {
	recebido, err := createParser("8 / 4 / 2").Parse()
	if err != nil {
		t.Fatal(err)
	}
	obtido := recebido.Declaracoes[0].(*parser.OpBinaria)
	esq, ok := obtido.Esq.(*parser.OpBinaria)
	if !ok || esq.Operador != "/" {
		t.Fatalf("esperava associatividade à esquerda na divisao")
	}
}

// 2 ** 3 ** 2 deve parsear como 2 ** (3 ** 2) = 512 (right-associative).
func TestAssociatividadePotencia(t *testing.T) {
	recebido, err := createParser("2 ** 3 ** 2").Parse()
	if err != nil {
		t.Fatal(err)
	}
	obtido := recebido.Declaracoes[0].(*parser.OpBinaria)

	// Esq deve ser literal simples "2", Dir deve ser OpBinaria (3 ** 2).
	if _, ok := obtido.Esq.(*parser.InteiroLiteral); !ok {
		t.Fatalf("esperava literal 2 à esquerda da potencia")
	}
	if _, ok := obtido.Dir.(*parser.OpBinaria); !ok {
		t.Fatalf("esperava associatividade à direita para **")
	}
}

// `{a: 1, b: 2}` deve marcar a entrada como não-implícita e usar o literal "a"/"b".
func TestMapaLiteralExplicito(t *testing.T) {
	recebido, err := createParser(`{a: 1, b: 2}`).Parse()
	if err != nil {
		t.Fatal(err)
	}
	mapa, ok := recebido.Declaracoes[0].(*parser.MapaLiteral)
	if !ok {
		t.Fatalf("esperava MapaLiteral")
	}
	if len(mapa.Entradas) != 2 {
		t.Fatalf("esperava 2 entradas")
	}
	for i, e := range mapa.Entradas {
		if e.EhImplicito {
			t.Fatalf("entrada %d deveria ser explicita", i)
		}
		if _, ok := e.Chave.(*parser.Identificador); !ok {
			t.Fatalf("entrada %d: esperava Identificador como chave", i)
		}
	}
}

// `{a, b}` deve marcar EhImplicito nas entradas.
func TestMapaLiteralImplicito(t *testing.T) {
	recebido, err := createParser(`{a, b}`).Parse()
	if err != nil {
		t.Fatal(err)
	}
	mapa := recebido.Declaracoes[0].(*parser.MapaLiteral)
	if len(mapa.Entradas) != 2 {
		t.Fatalf("esperava 2 entradas")
	}
	for i, e := range mapa.Entradas {
		if !e.EhImplicito {
			t.Fatalf("entrada %d deveria ser implicita", i)
		}
	}
}

// Import com identificadores válidos deve produzir nomes listados.
func TestImportValido(t *testing.T) {
	code := `de "./modulo" importe a, b;`
	recebido, err := createParser(code).Parse()
	if err != nil {
		t.Fatal(err)
	}
	imp, ok := recebido.Declaracoes[0].(*parser.ImporteDe)
	if !ok {
		t.Fatalf("esperava ImporteDe")
	}
	if len(imp.Nomes) != 2 || imp.Nomes[0] != "a" || imp.Nomes[1] != "b" {
		t.Fatalf("nomes errados: %#v", imp.Nomes)
	}
}

// Import sem identificadores após 'importe' deve falhar com erro de parse.
func TestImportSemNomes(t *testing.T) {
	_, err := createParser(`de "./modulo" importe 1`).Parse()
	if err == nil {
		t.Fatalf("esperava erro de parse para import sem identificadores")
	}
}

// Testa a análise sintática de classes com e sem herança
func TestParserClasses(t *testing.T) {
	code := `
	classe Animal {
		func falar() {}
	}
	classe Cachorro estende Animal {
		estatico func latir() {}
	}
	`
	recebido, err := createParser(code).Parse()
	if err != nil {
		t.Fatal(err)
	}

	if len(recebido.Declaracoes) != 2 {
		t.Fatalf("esperava 2 declarações de classe, obteve %d", len(recebido.Declaracoes))
	}

	classe1, ok := recebido.Declaracoes[0].(*parser.DeclClasse)
	if !ok || classe1.Nome != "Animal" || len(classe1.Metodos) != 1 || classe1.Heranca != "" {
		t.Fatalf("classe Animal incorreta: %#v", classe1)
	}

	classe2, ok := recebido.Declaracoes[1].(*parser.DeclClasse)
	if !ok || classe2.Nome != "Cachorro" || len(classe2.Metodos) != 1 || classe2.Heranca != "Animal" || !classe2.Metodos[0].Estatico {
		t.Fatalf("classe Cachorro incorreta: %#v", classe2)
	}
}

// Testa a análise sintática do operador binário 'instancia de'
func TestParserInstanciaDe(t *testing.T) {
	code := `x instancia de Animal`
	recebido, err := createParser(code).Parse()
	if err != nil {
		t.Fatal(err)
	}

	op, ok := recebido.Declaracoes[0].(*parser.OpBinaria)
	if !ok || op.Operador != "instancia" {
		t.Fatalf("esperava operador 'instancia', obteve: %#v", op)
	}
}
