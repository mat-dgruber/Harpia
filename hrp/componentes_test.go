package hrp_test

import (
	"testing"

	"github.com/mat-dgruber/Harpia/hrp"
)

func TestTemplateStringsEInterpolacao(t *testing.T) {
	codigo := `
	var nome = "Harpia"
	var versao = 1
	var msg = "Olá, {nome}! Versão {versao}."
	assegura msg == "Olá, Harpia! Versão 1."
	`

	ctx := hrp.NewContexto(hrp.OpcsContexto{})
	defer ctx.Terminar()

	_, err := hrp.ExecutarString(ctx, codigo)
	if err != nil {
		t.Fatalf("Erro inesperado durante execução com interpolação de strings: %v", err)
	}
}

func TestTemplateComPipes(t *testing.T) {
	codigo := `
	var nome = "  cafe  "
	funcao maiusculas(t) {
		# Método auxiliar simulado
		retorne "CAFE"
	}
	var msg = "Café: {nome |> maiusculas}"
	assegura msg == "Café: CAFE"
	`

	ctx := hrp.NewContexto(hrp.OpcsContexto{})
	defer ctx.Terminar()

	_, err := hrp.ExecutarString(ctx, codigo)
	if err != nil {
		t.Fatalf("Erro inesperado durante execução com interpolação contendo operador Pipe: %v", err)
	}
}
