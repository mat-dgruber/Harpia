//go:build js && wasm

// Package main implementa a ponte de execução WebAssembly do Harpia,
// permitindo rodar códigos-fonte e regras de negócio da linguagem diretamente em navegadores e runtimes JS.
package main

import (
	"syscall/js"

	"github.com/mat-dgruber/Harpia/hrp"
)

var (
	// Commit armazena a hash do git commit injetada dinamicamente via LDFLAGS.
	Commit   string = "-"
	// Datetime armazena o carimbo de data/hora de geração do binário.
	Datetime string = "0000-00-00T00:00:00"
	// Version armazena a versão semântica do compilador WASM.
	Version  string = "dev"
)

// main inicializa a ponte WASM exportando funções de avaliação de código Harpia
// para o objeto global do JavaScript e mantendo a execução viva por meio de canais.
func main() {
	c := make(chan struct{}, 0)

	// Expõe a função global rodarHarpia no escopo do JavaScript (window/globalThis).
	// Exemplo de uso em JS:
	//   let resultado = window.rodarHarpia("imprimir('Olá do Navegador!')");
	js.Global().Set("rodarHarpia", js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) < 1 {
			return "erro: código Harpia não fornecido"
		}
		codigo := args[0].String()

		// Instancia um contexto de execução leve
		ctx := hrp.NewContexto(hrp.OpcsContexto{})
		defer ctx.Terminar()

		// Executa a string de código Harpia no interpretador reativo
		_, err := hrp.ExecutarString(ctx, codigo)
		if err != nil {
			return "erro: " + err.Error()
		}
		return "sucesso"
	}))

	// Bloqueia indefinidamente para manter as funções exportadas ativas no Event Loop do JS
	<-c
}
