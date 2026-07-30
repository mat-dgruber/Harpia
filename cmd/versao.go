package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

var (
	Versao    = "dev"
	GitCommit = "desconhecido"
	DataBuild = "desconhecida"
)

func comandoVersao() *cobra.Command {
	var formato string
	var curto bool

	cmd := &cobra.Command{
		Use:           "versao",
		Aliases:       []string{"versão"},
		Short:         "Exibe a versão e metadados de compilação da CLI",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()

			if formato == "json" {
				meta := map[string]string{
					"versao":     Versao,
					"git_commit": GitCommit,
					"data_build": DataBuild,
				}
				enc := json.NewEncoder(out)
				enc.SetIndent("", "  ")
				return enc.Encode(meta)
			}

			if curto {
				fmt.Fprintln(out, Versao)
				return nil
			}

			fmt.Fprintf(out, "Harpia CLI %s (commit %s, compilado em %s)\n", Versao, GitCommit, DataBuild)
			return nil
		},
	}

	cmd.Flags().StringVar(&formato, "formato", "", "Exibe os metadados em formato específico (ex: json)")
	cmd.Flags().BoolVarP(&curto, "curto", "c", false, "Exibe apenas o número da versão")
	return cmd
}
