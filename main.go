//go:build !js || !wasm

// Package main é o ponto de entrada da CLI nativa/desktop do Harpia.
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
	// Commit armazena a hash do git commit injetada dinamicamente via LDFLAGS durante a compilação.
	Commit   string = "-"
	// Datetime armazena o carimbo de data/hora de geração do binário injetado via LDFLAGS.
	Datetime string = "0000-00-00T00:00:00"
	// Version armazena a tag de versão semântica do compilador injetada via LDFLAGS.
	Version  string = "dev"
)

// LongDescription descreve o manifesto filosófico da linguagem Harpia exibido na ajuda da CLI.
const LongDescription = `
	Uma linguagem reativa orientada a objetos e eventos completamente em português que visa
facilitar os estudos por parte de novos aventureiros no mundo da programação
com foco em Clean Architecture e DDD, sem ficar apenas criando códigos sem uso prático.

	A documentação completa pode ser encontrada em https://github.com/mat-dgruber/Harpia
`

// init inicializa as variáveis compartilhadas de build e release no pacote de comandos da CLI.
func init() {
	cmd.Commit = Commit
	cmd.Datetime = Datetime
	cmd.Version = Version
}

// embeddedSource armazena o código-fonte Harpia embutido estaticamente no binário durante empacotamentos AOT de distribuição única.
var embeddedSource string

// main é a função principal que gerencia o ciclo de vida do interpretador Harpia,
// executando scripts embutidos diretamente ou inicializando os comandos da interface de linha de comando (CLI) baseada no Cobra.
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
