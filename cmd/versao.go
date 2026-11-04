package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

// Os metadados vêm de cmd.Version/Commit/Datetime (cmd.go), preenchidos por main.go a partir
// das ldflags da release (.goreleaser.yaml). As antigas variáveis Versao/GitCommit/DataBuild
// nunca eram preenchidas, e `harpia versao` sempre mostrava "dev".

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
					"versao":     Version,
					"git_commit": Commit,
					"data_build": Datetime,
				}
				enc := json.NewEncoder(out)
				enc.SetIndent("", "  ")
				return enc.Encode(meta)
			}

			if curto {
				fmt.Fprintln(out, Version)
				return nil
			}

			fmt.Fprintf(out, "Harpia CLI %s (commit %s, compilado em %s)\n", Version, Commit, Datetime)
			return nil
		},
	}

	cmd.Flags().StringVar(&formato, "formato", "", "Exibe os metadados em formato específico (ex: json)")
	cmd.Flags().BoolVarP(&curto, "curto", "c", false, "Exibe apenas o número da versão")
	return cmd
}
