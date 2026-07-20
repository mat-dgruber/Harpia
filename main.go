//go:build !js || !wasm

package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/mat-dgruber/Harpia/cmd"
	"github.com/mat-dgruber/Harpia/hrp"
	"github.com/spf13/cobra"
)

var (
	Commit   string = "-"
	Datetime string = "0000-00-00T00:00:00"
	Version  string = "dev"
)

const LongDescription = `
	Uma linguagem reativa orientada a objetos e eventos completamente em português que visa
facilitar os estudos por parte de novos aventureiros no mundo da programação
com foco em Clean Architecture e DDD, sem ficar apenas criando códigos sem uso prático.

	A documentação completa pode ser encontrada em https://github.com/mat-dgruber/Harpia
`

func init() {
	cmd.Commit = Commit
	cmd.Datetime = Datetime
	cmd.Version = Version
}

var embeddedSource string

func main() {
	if embeddedSource != "" {
		// Importa a biblioteca padrão implicitamente
		_ = "github.com/mat-dgruber/Harpia/stdlib"

		ctx := hrp.NewContexto(hrp.OpcsContexto{})
		defer ctx.Terminar()

		_, err := hrp.ExecutarString(ctx, embeddedSource)
		if err != nil {
			hrp.LancarErro(err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	rootCmd := &cobra.Command{
		Use:     "harpia [arquivo]",
		Short:   "Harpia é uma linguagem de programação reativa de alto desempenho em Português",
		Long:    strings.ReplaceAll(strings.Trim(LongDescription, "\n "), "\t", "    "),
		Version: Version,
	}
	cmd.InstalarComandos(rootCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
