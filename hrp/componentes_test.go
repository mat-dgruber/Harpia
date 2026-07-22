package hrp_test

import (
	"testing"

	"github.com/mat-dgruber/Harpia/hrp"
)

// TestTemplateStringsEInterpolacao valida o mecanismo nativo de interpolação em strings de template,
// garantindo que as chaves '{variavel}' sejam substituídas dinamicamente pelos valores resolvidos
// no escopo de execução em runtime.
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

// TestTemplateComPipes garante que o mecanismo de interpolação funcione corretamente quando
// acoplado com o operador Pipe (|>), permitindo o encadeamento de transformações fluentes
// na resolução de variáveis embutidas no template de string.
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
