package cmd

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/spf13/cobra"
)

func obterUltimaVersaoURL(urlAPI string) (string, error) {
	resp, err := http.Get(urlAPI)
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
	if err := json.NewDecoder(resp.Body).Decode(&tags); err != nil {
		return "", fmt.Errorf("erro ao decodificar JSON de versoes: %w", err)
	}

	if len(tags) == 0 {
		return "", fmt.Errorf("nenhuma tag/versao encontrada")
	}

	return strings.TrimPrefix(tags[0].Name, "v"), nil
}

func extrairZipNativo(zipPath, destDir string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	cleanDest := filepath.Clean(destDir)

	for _, f := range r.File {
		fpath := filepath.Join(destDir, f.Name)
		if !strings.HasPrefix(fpath, cleanDest+string(os.PathSeparator)) && fpath != cleanDest {
			return fmt.Errorf("caminho invalido no zip: %s", f.Name)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func extrairTarGzNativo(tarGzPath, destDir string) error {
	f, err := os.Open(tarGzPath)
	if err != nil {
		return err
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	cleanDest := filepath.Clean(destDir)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		fpath := filepath.Join(destDir, header.Name)
		if !strings.HasPrefix(fpath, cleanDest+string(os.PathSeparator)) && fpath != cleanDest {
			return fmt.Errorf("caminho invalido no tar.gz: %s", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(fpath, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(fpath), 0755); err != nil {
				return err
			}
			outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, header.FileInfo().Mode())
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
		}
	}
	return nil
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

			urlAPI := "https://api.github.com/repos/mat-dgruber/Harpia/tags"
			remota, err := obterUltimaVersaoURL(urlAPI)
			if err != nil {
				fmt.Fprintf(errOut, "Erro ao verificar atualizações: %v\n", err)
				return err
			}

			vAtual, _ := semver.NewVersion(Versao)
			vRemota, _ := semver.NewVersion(remota)

			if vAtual != nil && vRemota != nil && !vAtual.LessThan(vRemota) {
				fmt.Fprintf(out, "Sua CLI do Harpia já está atualizada (%s).\n", Versao)
				return nil
			}

			fmt.Fprintf(out, "Nova versão disponível: %s (versão atual: %s)\n", remota, Versao)

			if apenasChecar {
				return nil
			}

			fmt.Fprintln(out, "Iniciando atualização automatizada...")
			return nil
		},
	}

	cmd.Flags().BoolVar(&apenasChecar, "checar", false, "Verifica se há nova versão sem realizar o download")
	return cmd
}
