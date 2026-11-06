package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/spf13/cobra"
)

// urlTagsAtualizar é a API de tags consultada por 'harpia atualizar' (variável para testes).
var urlTagsAtualizar = "https://api.github.com/repos/mat-dgruber/Harpia/tags"

// obterUltimaVersaoURL devolve a maior versão semver entre as tags (a API não ordena por versão:
// tags[0] podia ser v1.9.0 com v1.10.0 publicada). Tags que não são semver são ignoradas.
func obterUltimaVersaoURL(urlAPI string) (string, error) {
	cliente := &http.Client{Timeout: 15 * time.Second}
	resp, err := cliente.Get(urlAPI)
	if err != nil {
		return "", fmt.Errorf("erro de conexao com API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("resposta HTTP invalida: %s", resp.Status)
	}

	var tags []struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 5<<20)).Decode(&tags); err != nil {
		return "", fmt.Errorf("erro ao decodificar JSON de versoes: %w", err)
	}

	var maior *semver.Version
	for _, tag := range tags {
		v, err := semver.NewVersion(tag.Name)
		if err != nil {
			continue
		}
		if maior == nil || v.GreaterThan(maior) {
			maior = v
		}
	}
	if maior == nil {
		return "", fmt.Errorf("nenhuma tag/versao encontrada")
	}
	return maior.String(), nil
}

func comandoAtualizar() *cobra.Command {
	var apenasChecar bool

	cmd := &cobra.Command{
		Use:           "atualizar",
		Aliases:       []string{"atualize"},
		Short:         "Atualiza a CLI do Harpia para a versão mais recente",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			errOut := cmd.ErrOrStderr()

			remota, err := obterUltimaVersaoURL(urlTagsAtualizar)
			if err != nil {
				fmt.Fprintf(errOut, "Erro ao verificar atualizações: %v\n", err)
				return err
			}

			vAtual, _ := semver.NewVersion(Version)
			vRemota, _ := semver.NewVersion(remota)

			if vAtual != nil && vRemota != nil && !vAtual.LessThan(vRemota) {
				fmt.Fprintf(out, "Sua CLI do Harpia já está atualizada (%s).\n", Version)
				return nil
			}

			fmt.Fprintf(out, "Nova versão disponível: %s (versão atual: %s)\n", remota, Version)

			if apenasChecar {
				return nil
			}

			// ponytail: sem auto-update (exigiria download + verificação de checksum/assinatura).
			// Em vez de anunciar uma atualização que não acontece, aponta a release.
			fmt.Fprintf(out, "Baixe a nova versão em https://github.com/mat-dgruber/Harpia/releases/tag/v%s\n", remota)
			return nil
		},
	}

	cmd.Flags().BoolVar(&apenasChecar, "checar", false, "Verifica se há nova versão sem realizar o download")
	return cmd
}
