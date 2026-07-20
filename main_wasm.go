//go:build js && wasm

package main

import (
	"syscall/js"

	"github.com/mat-dgruber/Harpia/hrp"
)

var (
	Commit   string = "-"
	Datetime string = "0000-00-00T00:00:00"
	Version  string = "dev"
)

func main() {
	c := make(chan struct{}, 0)

	// Expõe a função global rodarHarpia no escopo do JavaScript
	js.Global().Set("rodarHarpia", js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) < 1 {
			return "erro: código Harpia não fornecido"
		}
		codigo := args[0].String()

		ctx := hrp.NewContexto(hrp.OpcsContexto{})
		defer ctx.Terminar()

		_, err := hrp.ExecutarString(ctx, codigo)
		if err != nil {
			return "erro: " + err.Error()
		}
		return "sucesso"
	}))

	<-c
}
