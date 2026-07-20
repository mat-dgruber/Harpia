package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func lerDependenciasNome(dir string) (string, string) {
	nome, curto := "", ""
	for _, c := range []string{filepath.Join(dir, "..", "dependencias.json"), filepath.Join(dir, "dependencias.json"), "dependencias.json"} {
		b, err := os.ReadFile(c)
		if err != nil {
			continue
		}
		s := string(b)
		// crude extract, no JSON dependency needed
		if i := strings.Index(s, `"nome"`); i >= 0 {
			nome = extrairValor(s[i:])
		}
		if i := strings.Index(s, `"nomeCurto"`); i >= 0 {
			curto = extrairValor(s[i:])
		}
		if nome != "" || curto != "" {
			break
		}
	}
	if nome == "" {
		nome = "Aplicação Harpia"
	}
	if curto == "" {
		curto = "Harpia"
	}
	return nome, curto
}

func extrairValor(s string) string {
	// s começa logo após "chave"
	i := strings.Index(s, ":")
	if i < 0 {
		return ""
	}
	s = s[i+1:]
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, `"`) {
		if j := strings.Index(s[1:], `"`); j >= 0 {
			return s[1 : 1+j]
		}
	}
	return strings.TrimRight(strings.TrimSpace(s), ", \t\n")
}

type pwaIcon struct {
	Src   string `json:"src"`
	Sizes string `json:"sizes"`
	Type  string `json:"type"`
}

type pwaManifest struct {
	Name            string    `json:"name"`
	ShortName       string    `json:"short_name"`
	StartURL        string    `json:"start_url"`
	Display         string    `json:"display"`
	BackgroundColor string    `json:"background_color"`
	ThemeColor      string    `json:"theme_color"`
	Icons           []pwaIcon `json:"icons"`
}

func gerarManifest(dir, nome, curto, fundo, tema string) (string, error) {
	m := pwaManifest{
		Name:            nome,
		ShortName:       curto,
		StartURL:        "/",
		Display:         "standalone",
		BackgroundColor: fundo,
		ThemeColor:      tema,
		Icons: []pwaIcon{
			{Src: "/icon-192.png", Sizes: "192x192", Type: "image/png"},
			{Src: "/icon-512.png", Sizes: "512x512", Type: "image/png"},
		},
	}
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return "", err
	}
	out := filepath.Join(dir, "manifest.webmanifest")
	if err := os.WriteFile(out, b, 0644); err != nil {
		return "", err
	}
	return out, nil
}

func gerarSW(dir string) (string, error) {
	const sw = `// service worker básico cache-first
const PRECACHE = ["/", "/index.html", "/app.js", "/runtime-web.js", "/estilos.css", "/manifest.webmanifest"];

self.addEventListener("install", (e) => {
  e.waitUntil(caches.open("harpia-v1").then((c) => c.addAll(PRECACHE)).then(() => self.skipWaiting()));
});

self.addEventListener("fetch", (e) => {
  if (e.request.method !== "GET") return;
  e.respondWith(
    caches.match(e.request).then((cached) => {
      if (cached) return cached;
      return fetch(e.request).then((res) => {
        if (res && res.status === 200 && res.type === "basic") {
          const copy = res.clone();
          caches.open("harpia-v1").then((c) => c.put(e.request, copy));
        }
        return res;
      }).catch(() => caches.match("/index.html"));
    })
  );
});
`
	out := filepath.Join(dir, "sw.js")
	if err := os.WriteFile(out, []byte(sw), 0644); err != nil {
		return "", err
	}
	return out, nil
}

func patcharHTML(htmlPath, manifestHref, cor string, registrar bool) error {
	b, err := os.ReadFile(htmlPath)
	if err != nil {
		return err
	}
	s := string(b)

	if !strings.Contains(s, `rel="manifest"`) {
		tag := fmt.Sprintf(`<link rel="manifest" href="%s">`, manifestHref)
		if i := strings.Index(s, "</head>"); i >= 0 {
			s = s[:i] + tag + "\n" + s[i:]
		} else {
			s = tag + "\n" + s
		}
	}

	if !strings.Contains(s, `name="theme-color"`) {
		mt := fmt.Sprintf(`<meta name="theme-color" content="%s">`, cor)
		if i := strings.Index(s, "</head>"); i >= 0 {
			s = s[:i] + mt + "\n" + s[i:]
		}
	}

	if registrar && !strings.Contains(s, `"/sw.js"`) && !strings.Contains(s, "navigator.serviceWorker.register") {
		script := `<script>
if ("serviceWorker" in navigator) {
  window.addEventListener("load", () => {
    navigator.serviceWorker.register("/sw.js").catch(() => {});
  });
}
</script>`
		if i := strings.Index(s, "</body>"); i >= 0 {
			s = s[:i] + script + "\n" + s[i:]
		} else {
			s = s + "\n" + script
		}
	}

	return os.WriteFile(htmlPath, []byte(s), 0644)
}

func comandoPwa() *cobra.Command {
	var (
		dir, nome, curto, fundo, tema string
		registrar                     bool
	)
	c := &cobra.Command{
		Use:   "pwa",
		Short: "Gera manifest.webmanifest e sw.js (PWA) a partir de dist/",
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, err := os.Stat(dir); err != nil {
				return fmt.Errorf("diretório %s não existe (rode harpia compilar antes): %w", dir, err)
			}
			if nome == "" || curto == "" {
				n, ct := lerDependenciasNome(dir)
				if nome == "" {
					nome = n
				}
				if curto == "" {
					curto = ct
				}
				_ = n
				_ = ct
			}
			mf, err := gerarManifest(dir, nome, curto, fundo, tema)
			if err != nil {
				return err
			}
			fmt.Printf("gerado: %s\n", mf)

			sw, err := gerarSW(dir)
			if err != nil {
				return err
			}
			fmt.Printf("gerado: %s\n", sw)

			html := filepath.Join(dir, "index.html")
			if _, err := os.Stat(html); err == nil {
				if err := patcharHTML(html, "/manifest.webmanifest", tema, registrar); err != nil {
					return err
				}
				fmt.Printf("patchado: %s\n", html)
			} else {
				fmt.Printf("aviso: %s não existe, pulando patch\n", html)
			}
			return nil
		},
	}
	c.Flags().StringVar(&dir, "dir", "dist", "Diretório alvo")
	c.Flags().StringVar(&nome, "nome", "", "Nome do app (lido de dependencias.json se vazio)")
	c.Flags().StringVar(&curto, "curto", "", "Nome curto (lido de dependencias.json se vazio)")
	c.Flags().StringVar(&fundo, "cor-fundo", "#ffffff", "Cor de fundo")
	c.Flags().StringVar(&tema, "cor-tema", "#3b82f6", "Cor de tema")
	c.Flags().BoolVar(&registrar, "registrar", false, "Injeta registro do service worker em index.html")
	return c
}
