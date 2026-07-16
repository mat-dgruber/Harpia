package ptst_test

import (
	"testing"

	"github.com/natanfeitosa/portuscript/ptst"
)

func TestTemplateStringsEInterpolacao(t *testing.T) {
	codigo := `
	var nome = "Portuscript"
	var versao = 1
	var msg = "Olá, {nome}! Versão {versao}."
	assegura msg == "Olá, Portuscript! Versão 1."
	`

	ctx := ptst.NewContexto(ptst.OpcsContexto{})
	defer ctx.Terminar()

	_, err := ptst.ExecutarString(ctx, codigo)
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

	ctx := ptst.NewContexto(ptst.OpcsContexto{})
	defer ctx.Terminar()

	_, err := ptst.ExecutarString(ctx, codigo)
	if err != nil {
		t.Fatalf("Erro inesperado durante execução com interpolação contendo operador Pipe: %v", err)
	}
}
