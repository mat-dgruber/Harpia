package hrp

import (
	"testing"
)

func TestParametrosAvancados(t *testing.T) {
	codigo := `
	funcao cumprimentar(nome, saudacao = "Olá") {
		retorne saudacao + ", " + nome
	}

	funcao calcular(a, b = 2, c = 3) {
		retorne a * b + c
	}

	var res1 = cumprimentar("Carlos")
	var res2 = cumprimentar("Carlos", "Bem-vindo")
	var res3 = cumprimentar(saudacao = "Bom dia", nome = "Ana")

	var res4 = calcular(5)
	var res5 = calcular(5, c = 10)
	`

	ctx := NewContexto(OpcsContexto{})
	defer ctx.Terminar()

	_, err := ExecutarString(ctx, codigo)
	if err != nil {
		t.Fatalf("Erro inesperado ao executar: %v", err)
	}

	modulo, err := ctx.ObterModulo("__entrada__")
	if err != nil {
		t.Fatal(err)
	}

	res1, _ := modulo.Escopo.ObterValor("res1")
	if string(res1.(Texto)) != "Olá, Carlos" {
		t.Errorf("Esperava 'Olá, Carlos', obteve: %v", res1)
	}

	res2, _ := modulo.Escopo.ObterValor("res2")
	if string(res2.(Texto)) != "Bem-vindo, Carlos" {
		t.Errorf("Esperava 'Bem-vindo, Carlos', obteve: %v", res2)
	}

	res3, _ := modulo.Escopo.ObterValor("res3")
	if string(res3.(Texto)) != "Bom dia, Ana" {
		t.Errorf("Esperava 'Bom dia, Ana', obteve: %v", res3)
	}

	res4, _ := modulo.Escopo.ObterValor("res4")
	if int(res4.(Inteiro)) != 13 { // 5 * 2 + 3 = 13
		t.Errorf("Esperava 13, obteve: %v", res4)
	}

	res5, _ := modulo.Escopo.ObterValor("res5")
	if int(res5.(Inteiro)) != 20 { // 5 * 2 + 10 = 20
		t.Errorf("Esperava 20, obteve: %v", res5)
	}
}
