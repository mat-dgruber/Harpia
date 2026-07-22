package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

// ponytail: heurística simples — pega string literal como primeiro arg de qualquer
// chamada tipo `t("...")`/`i18n.texto("...")`/`tr(..."...")`. Refinar quando AST
// pública do parser permitir visitar ChamadaExpr com segurança sem quebrar caches.

var (
	reChamadaI18n     = regexp.MustCompile(`(?i)(?:^|[^\w])(?:t|tr|i18n\.texto|texto|traduzir)\s*\(\s*"((?:[^"\\]|\\.)*)"`)
	reComentarioLinha = regexp.MustCompile(`(?m)^\s*--.*$`)
)

type entradaPot struct {
	Arquivo string
	Linha   int
	Msgid   string
	Refs    []string
}

// stripComentarios remove linhas de comentário iniciadas com `--` do código fonte
// antes da aplicação do regex de extração.
//
// Importante: isso evita falsos positivos quando, por exemplo, um trecho de
// documentação menciona `t("exemplo")` como ilustração dentro de um comentário.
func stripComentarios(src string) string {
	return reComentarioLinha.ReplaceAllString(src, "")
}

func nomeArquivoSemExt(p string) string {
	base := filepath.Base(p)
	if i := strings.LastIndex(base, "."); i >= 0 {
		return base[:i]
	}
	return base
}

func extrairEntradas(src string) []entradaPot {
	clean := stripComentarios(src)
	matches := reChamadaI18n.FindAllStringSubmatchIndex(clean, -1)
	var out []entradaPot
	for _, m := range matches {
		if len(m) < 4 {
			continue
		}
		// m[2]:m[3] é o grupo capturado da msgid
		msgid := clean[m[2]:m[3]]
		if msgid == "" {
			continue
		}
		// decoding básico de escapes JSON-like
		msgid = strings.ReplaceAll(msgid, `\"`, `"`)
		msgid = strings.ReplaceAll(msgid, `\\`, `\`)
		msgid = strings.ReplaceAll(msgid, `\n`, "\n")
		msgid = strings.ReplaceAll(msgid, `\t`, "\t")

		linha := 1 + strings.Count(clean[:m[0]], "\n")
		out = append(out, entradaPot{
			Msgid: msgid,
			Linha: linha,
		})
	}
	return out
}

func escaparPo(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	s = strings.ReplaceAll(s, "\n", `\n`)
	s = strings.ReplaceAll(s, "\t", `\t`)
	return s
}

// escreverPot materializa o catálogo `.pot` (template PO) a partir das entradas extraídas.
//
// O arquivo final segue as convenções do formato gettext:
//  1. Cabeçalho com metadados de Content-Type, Transfer-Encoding e Language;
//  2. Para cada `msgid` distinta (deduplicada via mapa), emite a linha de referência
//     agregada (`#: arquivo:linha arquivo:linha ...`) e o par `msgid`/`msgstr`.
//
// A deduplicação preserva a primeira ocorrência de cada chave e soma as referências
// subsequentes para facilitar o trabalho do tradutor.
func escreverPot(path, dominio string, entradas []entradaPot) error {
	var b strings.Builder
	fmt.Fprintf(&b, `# Catálogo POT (template) gerado por `+"`harpia i18n extrair`"+`
# Domínio: %s
msgid ""
msgstr ""
"Content-Type: text/plain; charset=UTF-8\n"
"Content-Transfer-Encoding: 8bit\n"
"Language: pt\n"

`, dominio)

	// dedup preservando refs
	porMsg := map[string]*entradaPot{}
	refs := map[string][]string{}
	for i := range entradas {
		e := &entradas[i]
		key := e.Msgid
		if ex, ok := porMsg[key]; ok {
			refs[key] = append(refs[key], fmt.Sprintf("%s:%d", e.Arquivo, e.Linha))
			_ = ex
			continue
		}
		c := *e
		porMsg[key] = &c
		refs[key] = []string{fmt.Sprintf("%s:%d", e.Arquivo, e.Linha)}
	}

	keys := make([]string, 0, len(porMsg))
	for k := range porMsg {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		fmt.Fprintf(&b, "#: %s\n", strings.Join(refs[k], " "))
		fmt.Fprintf(&b, "msgid \"%s\"\n", escaparPo(k))
		fmt.Fprintf(&b, "msgstr \"\"\n\n")
	}

	return os.WriteFile(path, []byte(b.String()), 0644)
}

func caminhoTraducoes(dir string) string {
	if dir == "" {
		dir = "traducoes"
	}
	return dir
}

func caminharPtst(raiz string) ([]string, error) {
	var out []string
	err := filepath.WalkDir(raiz, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if ext := filepath.Ext(path); ext == ".ptst" || ext == ".pt" {
			out = append(out, path)
		}
		return nil
	})
	return out, err
}

func comandoI18n() *cobra.Command {
	var dir, dominio string
	root := &cobra.Command{
		Use:   "i18n",
		Short: "Extrai e gerencia catálogos de tradução (.pot/.po)",
	}
	root.PersistentFlags().StringVar(&dir, "dir", "traducoes", "Diretório de catálogos")
	root.PersistentFlags().StringVar(&dominio, "dominio", "harpia", "Domínio gettext")

	extrair := &cobra.Command{
		Use:   "extrair <arquivo|dir>",
		Short: "Extrai strings traduzíveis e gera <dir>/<dominio>.pot",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			target := args[0]
			info, err := os.Stat(target)
			if err != nil {
				return err
			}
			var arquivos []string
			if info.IsDir() {
				arquivos, err = caminharPtst(target)
				if err != nil {
					return err
				}
			} else {
				arquivos = []string{target}
			}

			var entradas []entradaPot
			for _, a := range arquivos {
				b, err := os.ReadFile(a)
				if err != nil {
					return err
				}
				src := string(b)
				for _, e := range extrairEntradas(src) {
					e.Arquivo = a
					entradas = append(entradas, e)
				}
			}

			outDir := caminhoTraducoes(dir)
			if err := os.MkdirAll(outDir, 0755); err != nil {
				return err
			}
			out := filepath.Join(outDir, dominio+".pot")
			if err := escreverPot(out, dominio, entradas); err != nil {
				return err
			}
			fmt.Printf("extraídas %d strings em %d arquivos → %s\n",
				len(entradas), len(arquivos), out)
			return nil
		},
	}

	novo := &cobra.Command{
		Use:   "novo <idioma>",
		Short: "Cria catálogo .po vazio para o idioma (cabeçalho copiado do .pot)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			idioma := args[0]
			outDir := caminhoTraducoes(dir)
			if err := os.MkdirAll(outDir, 0755); err != nil {
				return err
			}
			pot := filepath.Join(outDir, dominio+".pot")
			po := filepath.Join(outDir, idioma+".po")

			cabecalho := fmt.Sprintf(`# Catálogo PO para %s (gerado por harpia i18n novo)
msgid ""
msgstr ""
"Content-Type: text/plain; charset=UTF-8\n"
"Content-Transfer-Encoding: 8bit\n"
"Language: %s\n"

`, idioma, idioma)

			if b, err := os.ReadFile(pot); err == nil {
				cabecalho = string(b)
			} else {
				fmt.Printf("aviso: %s não existe, usando header genérico\n", pot)
			}

			if err := os.WriteFile(po, []byte(cabecalho), 0644); err != nil {
				return err
			}
			fmt.Printf("criado: %s\n", po)
			return nil
		},
	}

	root.AddCommand(extrair, novo)
	return root
}
