package cmd

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	_ "github.com/glebarez/go-sqlite"
	"github.com/spf13/cobra"
)

// caminhoMigracoes resolve o diretório canônico de migrations no projeto.
//
// Ordem de resolução: prioriza `infra/migracoes` (alinhado com a Clean Architecture)
// e cai para `migracoes` na raiz caso o primeiro não exista. Quando nenhum dos
// dois existe, retorna o caminho preferido (`infra/migracoes`) para que os
// subcomandos `migrar criar` criem a estrutura automaticamente.
func caminhoMigracoes() string {
	for _, c := range []string{filepath.Join("infra", "migracoes"), "migracoes"} {
		if info, err := os.Stat(c); err == nil && info.IsDir() {
			return c
		}
	}
	return filepath.Join("infra", "migracoes")
}

// abrirBanco estabelece a conexão SQLite com a biblioteca `glebarez/go-sqlite` (sem CGO).
// Aplica duas pragmas automaticamente:
//   - `journal_mode(WAL)`: habilita Write-Ahead Logging para reduzir contenção
//     em cenários com múltiplas leituras concorrentes;
//   - `foreign_keys(1)`: liga enforcement de chaves estrangeiras por padrão
//     (desligado em SQLite por motivos históricos).
//
// Cria o diretório-pai do arquivo do banco se necessário para evitar erros de
// `os.IsNotExist` em workspaces recém-inicializados.
func abrirBanco(dbPath string) (*sql.DB, error) {
	if dir := filepath.Dir(dbPath); dir != "" && dir != "." {
		_ = os.MkdirAll(dir, 0755)
	}
	return sql.Open("sqlite", dbPath+"?_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)")
}

func garantirTabela(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS _migracoes (versao TEXT PRIMARY KEY, aplicada_em DATETIME NOT NULL)`)
	return err
}

func extrairBloco(c, marcador string) string {
	lines := strings.Split(c, "\n")
	var buf strings.Builder
	dentro := false
	for _, l := range lines {
		t := strings.TrimSpace(l)
		if strings.HasPrefix(t, "-- +migrar ") {
			if dentro {
				break
			}
			if t == "-- +migrar "+marcador {
				dentro = true
			}
			continue
		}
		if dentro {
			buf.WriteString(l + "\n")
		}
	}
	return strings.TrimSpace(buf.String())
}

func listarMigracoes(dir string) ([]string, error) {
	es, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var a []string
	for _, e := range es {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			a = append(a, filepath.Join(dir, e.Name()))
		}
	}
	sort.Strings(a)
	return a, nil
}

func versaoAplicada(db *sql.DB) (map[string]bool, error) {
	rows, err := db.Query(`SELECT versao FROM _migracoes`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]bool{}
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		out[v] = true
	}
	return out, rows.Err()
}

func comandoMigrar() *cobra.Command {
	var banco string
	root := &cobra.Command{
		Use:   "migrar",
		Short: "Gerencia migrations SQL com SQLite",
	}
	root.PersistentFlags().StringVar(&banco, "banco", "dados.db", "Banco SQLite alvo")

	criar := &cobra.Command{
		Use:   "criar <nome>",
		Short: "Cria arquivo de migration timestamped",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := caminhoMigracoes()
			if err := os.MkdirAll(dir, 0755); err != nil {
				return err
			}
			ts := time.Now().Format("2006-01-02-150405")
			nome := strings.ReplaceAll(args[0], " ", "-")
			arq := filepath.Join(dir, fmt.Sprintf("%s-%s.sql", ts, nome))
			tpl := "-- +migrar ParaCima\n-- escreva SQL de subida aqui\n\n-- +migrar ParaBaixo\n-- escreva SQL de descida aqui\n"
			if err := os.WriteFile(arq, []byte(tpl), 0644); err != nil {
				return err
			}
			fmt.Printf("criado: %s\n", arq)
			return nil
		},
	}

	aplicar := &cobra.Command{
		Use:   "aplicar",
		Short: "Aplica todas as migrations pendentes em ordem",
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := caminhoMigracoes()
			arqs, err := listarMigracoes(dir)
			if err != nil {
				return err
			}
			db, err := abrirBanco(banco)
			if err != nil {
				return err
			}
			defer db.Close()
			if err := garantirTabela(db); err != nil {
				return err
			}
			aplicadas, err := versaoAplicada(db)
			if err != nil {
				return err
			}
			for _, a := range arqs {
				v := filepath.Base(a)
				if aplicadas[v] {
					continue
				}
				c, err := os.ReadFile(a)
				if err != nil {
					return err
				}
				sql := extrairBloco(string(c), "ParaCima")
				if sql == "" {
					fmt.Printf("pulando %s (sem bloco ParaCima)\n", v)
					continue
				}
				tx, err := db.Begin()
				if err != nil {
					return err
				}
				if _, err := tx.Exec(sql); err != nil {
					_ = tx.Rollback()
					return fmt.Errorf("falha em %s: %w", v, err)
				}
				if _, err := tx.Exec(`INSERT INTO _migracoes (versao, aplicada_em) VALUES (?, ?)`, v, time.Now()); err != nil {
					_ = tx.Rollback()
					return err
				}
				if err := tx.Commit(); err != nil {
					return err
				}
				fmt.Printf("aplicada: %s\n", v)
			}
			return nil
		},
	}

	status := &cobra.Command{
		Use:   "status",
		Short: "Lista migrations aplicadas e pendentes",
		RunE: func(cmd *cobra.Command, args []string) error {
			arqs, err := listarMigracoes(caminhoMigracoes())
			if err != nil {
				return err
			}
			db, err := abrirBanco(banco)
			if err != nil {
				return err
			}
			defer db.Close()
			if err := garantirTabela(db); err != nil {
				return err
			}
			aplicadas, err := versaoAplicada(db)
			if err != nil {
				return err
			}
			fmt.Printf("%-50s %-22s %s\n", "VERSAO", "ESTADO", "ARQUIVO")
			for _, a := range arqs {
				v := filepath.Base(a)
				est := "pendente"
				if aplicadas[v] {
					est = "aplicada"
				}
				fmt.Printf("%-50s %-22s %s\n", v, est, a)
			}
			return nil
		},
	}

	reverter := &cobra.Command{
		Use:   "reverter [n]",
		Short: "Reverte as últimas N migrations aplicadas (default 1)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			n := 1
			if len(args) > 0 {
				if _, err := fmt.Sscanf(args[0], "%d", &n); err != nil {
					return err
				}
			}
			db, err := abrirBanco(banco)
			if err != nil {
				return err
			}
			defer db.Close()
			if err := garantirTabela(db); err != nil {
				return err
			}
			rows, err := db.Query(`SELECT versao FROM _migracoes ORDER BY aplicada_em DESC LIMIT ?`, n)
			if err != nil {
				return err
			}
			defer rows.Close()
			var versoes []string
			for rows.Next() {
				var v string
				if err := rows.Scan(&v); err != nil {
					return err
				}
				versoes = append(versoes, v)
			}
			if err := rows.Err(); err != nil {
				return err
			}
			dir := caminhoMigracoes()
			for _, v := range versoes {
				a := filepath.Join(dir, v)
				c, err := os.ReadFile(a)
				if err != nil {
					fmt.Printf("aviso: %s não encontrado, removendo do registro\n", v)
					if _, err := db.Exec(`DELETE FROM _migracoes WHERE versao = ?`, v); err != nil {
						return err
					}
					continue
				}
				sql := extrairBloco(string(c), "ParaBaixo")
				if sql == "" {
					fmt.Printf("pulando reversão de %s (sem bloco ParaBaixo)\n", v)
					continue
				}
				if _, err := db.Exec(sql); err != nil {
					return fmt.Errorf("falha revertendo %s: %w", v, err)
				}
				if _, err := db.Exec(`DELETE FROM _migracoes WHERE versao = ?`, v); err != nil {
					return err
				}
				fmt.Printf("revertida: %s\n", v)
			}
			return nil
		},
	}

	root.AddCommand(criar, aplicar, status, reverter)
	return root
}
