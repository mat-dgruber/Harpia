package hrp

import (
	"testing"
)

// TestOperadorPipe valida o comportamento do operador de canal/Pipe (|>), garantindo
// que valores à esquerda sejam passados de forma fluente como primeiro argumento para
// a função ou chamada de método posicionada à direita no encadeamento.
func TestOperadorPipe(t *testing.T) {
	codigo := `
	de "embutidos" importe imprimir

	func duplicar(x) {
		retorne x * 2
	}

	func somar(x, y) {
		retorne x + y
	}

	# Pipe passando identificador simples
	var res1 = 10 |> duplicar
	
	# Pipe passando chamada com argumento extra (10 será injetado antes de 5, resultando em somar(10, 5))
	var res2 = 10 |> somar(5)

	# Encadeamento de pipes
	var res3 = 5 |> duplicar |> somar(10)
	`

	ctx := NewContexto(OpcsContexto{})
	defer ctx.Terminar()

	_, err := ExecutarString(ctx, codigo)
	if err != nil {
		t.Fatalf("Erro ao executar código com operador pipe: %v", err)
	}

	modulo, err := ctx.ObterModulo("__entrada__")
	if err != nil {
		t.Fatal(err)
	}

	res1, _ := modulo.Escopo.ObterValor("res1")
	if int(res1.(Inteiro)) != 20 {
		t.Errorf("Esperava res1 ser 20, obteve: %v", res1)
	}

	res2, _ := modulo.Escopo.ObterValor("res2")
	if int(res2.(Inteiro)) != 15 {
		t.Errorf("Esperava res2 ser 15, obteve: %v", res2)
	}

	res3, _ := modulo.Escopo.ObterValor("res3")
	if int(res3.(Inteiro)) != 20 {
		t.Errorf("Esperava res3 ser 20, obteve: %v", res3)
	}
}

// TestOperadorPipeSemEfeitoColateralDuplo garante que a expressão do lado esquerdo de um operador Pipe
// seja avaliada apenas uma única vez, prevenindo a reexecução indesejada de efeitos colaterais.
func TestOperadorPipeSemEfeitoColateralDuplo(t *testing.T) {
	codigo := `
	var contador = 0

	func incrementar() {
		contador = contador + 1
		retorne 10
	}

	func somar(x, y) {
		retorne x + y
	}

	var res = incrementar() |> somar(5)
	`

	ctx := NewContexto(OpcsContexto{})
	defer ctx.Terminar()

	_, err := ExecutarString(ctx, codigo)
	if err != nil {
		t.Fatalf("Erro ao executar: %v", err)
	}

	modulo, err := ctx.ObterModulo("__entrada__")
	if err != nil {
		t.Fatal(err)
	}

	res, _ := modulo.Escopo.ObterValor("res")
	if int(res.(Inteiro)) != 15 {
		t.Errorf("Esperava res ser 15, obteve: %v", res)
	}

	contador, _ := modulo.Escopo.ObterValor("contador")
	if int(contador.(Inteiro)) != 1 {
		t.Errorf("Esperava contador ser 1, indicando apenas uma execução, obteve: %v", contador)
	}
}
