// Package http implementa o servidor web nativo de alta performance e cliente HTTP do Harpia.
package http

import (
	"encoding/json"
	"fmt"
	"html"
	"os"
	"path/filepath"
	"strings"

	"github.com/mat-dgruber/Harpia/hrp"
)

// ServirAppHandler implementa a lógica de servir e reidratar um SPA Harpia com SSR (Server-Side Rendering),
// gerando tags AEO (Answer Engine Optimization), GEO (Geolocalização) e tags de redes sociais dinamicamente.
func ServirAppHandler(diretorioDist string, componenteRaiz hrp.Objeto, metadados hrp.Objeto) hrp.Objeto {
	return hrp.NewMetodoOuPanic("handler_spa", func(_ hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("argumentos inválidos no handler_spa")
		}
		req := args[0].(*Requisicao)
		res := args[1].(*Resposta)

		caminhoFisico := filepath.Clean(filepath.Join(diretorioDist, string(req.Caminho)))
		distLimpo := filepath.Clean(diretorioDist)

		// Valida se o caminho físico não sofreu Path Traversal escapando do diretório dist
		if caminhoFisico != distLimpo && !strings.HasPrefix(caminhoFisico, distLimpo+string(filepath.Separator)) {
			res.Status = 403
			res.Corpo = hrp.Texto("Acesso negado")
			return hrp.Nulo, nil
		}

		info, err := os.Stat(caminhoFisico)

		// Se o arquivo estático físico existir (e não for arquivo HTML), serve-o diretamente com cabeçalhos MIME corretos.
		servirFisico := false
		if err == nil && !info.IsDir() {
			ext := filepath.Ext(caminhoFisico)
			if ext != ".html" {
				servirFisico = true
			}
		}

		if servirFisico {
			bytes, err := os.ReadFile(caminhoFisico)
			if err != nil {
				res.Status = 500
				res.Corpo = hrp.Texto("Erro ao ler ativo estático")
				return hrp.Nulo, nil
			}

			// Define content-type básico conforme a extensão do arquivo
			ext := filepath.Ext(caminhoFisico)
			contentType := "text/plain"
			switch ext {
			case ".js":
				contentType = "application/javascript"
			case ".css":
				contentType = "text/css"
			case ".html":
				contentType = "text/html"
			case ".png":
				contentType = "image/png"
			case ".jpg", ".jpeg":
				contentType = "image/jpeg"
			case ".svg":
				contentType = "image/svg+xml"
			}
			res.Cabecalho["Content-Type"] = hrp.Texto(contentType)
			res.Corpo = hrp.Texto(bytes)
			return hrp.Nulo, nil
		}

		// Caso contrário (rota virtual de SPA ou arquivo inexistente), renderiza a página index.html usando SSR + Hidratação.
		indexPath := filepath.Join(diretorioDist, "index.html")
		htmlBytes, err := os.ReadFile(indexPath)
		if err != nil {
			res.Status = 404
			res.Corpo = hrp.Texto("SPA index.html não encontrado no diretório de build.")
			return hrp.Nulo, nil
		}

		htmlStr := string(htmlBytes)

		// 1. Executa e renderiza o componente raiz no backend (SSR)
		ssrHtml := ""
		if componenteRaiz != nil && componenteRaiz != hrp.Nulo {
			vdomObjeto, err := hrp.Chamar(componenteRaiz, hrp.Tupla{req.Caminho})
			if err != nil {
				var err2 error
				vdomObjeto, err2 = hrp.Chamar(componenteRaiz, hrp.Tupla{})
				if err2 != nil {
					res.Status = 500
					res.Corpo = hrp.Texto(fmt.Sprintf("Erro no SSR ao executar componente raiz: %v", err2))
					return hrp.Nulo, nil
				}
			}

			if el, ok := vdomObjeto.(*hrp.ElementoJSX); ok {
				ssrHtml = el.RenderizarHTML()
			} else {
				if s, err := hrp.NewTexto(vdomObjeto); err == nil {
					ssrHtml = string(s.(hrp.Texto))
				}
			}
		}

		// 2. Extrai e gera metadados AEO/GEO e OpenGraph de forma declarativa e dinâmica
		headMetaTags := strings.Builder{}
		if metadados != nil && metadados != hrp.Nulo {
			if mapa, ok := metadados.(hrp.Mapa); ok {
				// Título da Página
				if t, ok := mapa["titulo"]; ok {
					if txt, ok := t.(hrp.Texto); ok {
						escaped := html.EscapeString(string(txt))
						htmlStr = strings.Replace(htmlStr, "<title>Harpia App</title>", fmt.Sprintf("<title>%s</title>", escaped), 1)
						headMetaTags.WriteString(fmt.Sprintf("\n    <meta property=\"og:title\" content=\"%s\">", escaped))
					}
				}
				// Descrição SEO
				if d, ok := mapa["descricao"]; ok {
					if txt, ok := d.(hrp.Texto); ok {
						escaped := html.EscapeString(string(txt))
						headMetaTags.WriteString(fmt.Sprintf("\n    <meta name=\"description\" content=\"%s\">", escaped))
						headMetaTags.WriteString(fmt.Sprintf("\n    <meta property=\"og:description\" content=\"%s\">", escaped))
					}
				}
				// Autor
				if a, ok := mapa["autor"]; ok {
					if txt, ok := a.(hrp.Texto); ok {
						headMetaTags.WriteString(fmt.Sprintf("\n    <meta name=\"author\" content=\"%s\">", html.EscapeString(string(txt))))
					}
				}
				// OpenGraph Titulo específico
				if ogT, ok := mapa["ogTitulo"]; ok {
					if txt, ok := ogT.(hrp.Texto); ok {
						headMetaTags.WriteString(fmt.Sprintf("\n    <meta property=\"og:title\" content=\"%s\">", html.EscapeString(string(txt))))
					}
				}
				// OpenGraph Descrição específica
				if ogD, ok := mapa["ogDescricao"]; ok {
					if txt, ok := ogD.(hrp.Texto); ok {
						headMetaTags.WriteString(fmt.Sprintf("\n    <meta property=\"og:description\" content=\"%s\">", html.EscapeString(string(txt))))
					}
				}
				// Imagem de Compartilhamento OpenGraph
				if img, ok := mapa["imagem"]; ok {
					if txt, ok := img.(hrp.Texto); ok {
						headMetaTags.WriteString(fmt.Sprintf("\n    <meta property=\"og:image\" content=\"%s\">", html.EscapeString(string(txt))))
					}
				}
				if ogImg, ok := mapa["ogImagem"]; ok {
					if txt, ok := ogImg.(hrp.Texto); ok {
						headMetaTags.WriteString(fmt.Sprintf("\n    <meta property=\"og:image\" content=\"%s\">", html.EscapeString(string(txt))))
					}
				}
				// Canonical
				if can, ok := mapa["canonical"]; ok {
					if txt, ok := can.(hrp.Texto); ok {
						headMetaTags.WriteString(fmt.Sprintf("\n    <link rel=\"canonical\" href=\"%s\">", html.EscapeString(string(txt))))
					}
				}

				// Dados Estruturados Schema.org (AEO) para indexação inteligente por IA
				if esquema, ok := mapa["esquema"]; ok {
					if m, ok := esquema.(hrp.Mapa); ok {
						rawMap := make(map[string]interface{})
						for k, v := range m {
							if txt, ok := v.(hrp.Texto); ok {
								rawMap[k] = string(txt)
							} else {
								rawMap[k] = fmt.Sprintf("%v", v)
							}
						}
						if jsonBytes, err := json.Marshal(rawMap); err == nil {
							headMetaTags.WriteString(fmt.Sprintf("\n    <script type=\"application/ld+json\">\n    %s\n    </script>", string(jsonBytes)))
						}
					}
				}

				// Geolocalização Física (GEO Metatags)
				if geo, ok := mapa["geo"]; ok {
					if m, ok := geo.(hrp.Mapa); ok {
						lat := m["latitude"]
						lon := m["longitude"]
						if lat != nil && lon != nil {
							headMetaTags.WriteString(fmt.Sprintf("\n    <meta name=\"geo.position\" content=\"%s; %s\">", html.EscapeString(fmt.Sprintf("%v", lat)), html.EscapeString(fmt.Sprintf("%v", lon))))
							headMetaTags.WriteString(fmt.Sprintf("\n    <meta name=\"ICBM\" content=\"%s, %s\">", html.EscapeString(fmt.Sprintf("%v", lat)), html.EscapeString(fmt.Sprintf("%v", lon))))
						}
					}
				}
			}
		}

		// Injeta as tags head adicionais antes do fechamento do elemento </head>
		if headMetaTags.Len() > 0 {
			htmlStr = strings.Replace(htmlStr, "</head>", headMetaTags.String()+"\n</head>", 1)
		}

		// 3. Substitui e insere o HTML do componente reativo para Hidratação no Cliente
		htmlStr = strings.Replace(htmlStr, "<div id=\"app\"></div>", fmt.Sprintf("<div id=\"app\">%s</div>", ssrHtml), 1)

		res.Status = 200
		res.Cabecalho["Content-Type"] = hrp.Texto("text/html; charset=utf-8")
		res.Corpo = hrp.Texto(htmlStr)
		return hrp.Nulo, nil
	}, "")
}
