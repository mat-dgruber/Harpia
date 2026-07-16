package http

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/natanfeitosa/portuscript/ptst"
)

// ServirAppHandler implementa a lógica de servir e reidratar um SPA Portuscript com SSR, AEO/GEO.
func ServirAppHandler(diretorioDist string, componenteRaiz ptst.Objeto, metadados ptst.Objeto) ptst.Objeto {
	return ptst.NewMetodoOuPanic("handler_spa", func(_ ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("argumentos inválidos no handler_spa")
		}
		req := args[0].(*Requisicao)
		res := args[1].(*Resposta)

		caminhoFisico := filepath.Join(diretorioDist, string(req.Caminho))
		info, err := os.Stat(caminhoFisico)

		// Se o arquivo estático existir (e não for um diretório), serve-o diretamente
		if err == nil && !info.IsDir() {
			bytes, err := os.ReadFile(caminhoFisico)
			if err != nil {
				res.Status = 500
				res.Corpo = ptst.Texto("Erro ao ler ativo estático")
				return ptst.Nulo, nil
			}

			// Define content-type básico
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
			res.Cabecalho["Content-Type"] = ptst.Texto(contentType)
			res.Corpo = ptst.Texto(bytes)
			return ptst.Nulo, nil
		}

		// Caso contrário, renderiza a página usando SSR + index.html
		indexPath := filepath.Join(diretorioDist, "index.html")
		htmlBytes, err := os.ReadFile(indexPath)
		if err != nil {
			res.Status = 404
			res.Corpo = ptst.Texto("SPA index.html não encontrado no diretório de build.")
			return ptst.Nulo, nil
		}

		htmlStr := string(htmlBytes)

		// 1. Executa o componente raiz no backend
		vdomObjeto, err := ptst.Chamar(componenteRaiz, ptst.Tupla{})
		if err != nil {
			res.Status = 500
			res.Corpo = ptst.Texto(fmt.Sprintf("Erro no SSR ao executar componente raiz: %v", err))
			return ptst.Nulo, nil
		}

		ssrHtml := ""
		if el, ok := vdomObjeto.(*ptst.ElementoJSX); ok {
			ssrHtml = el.RenderizarHTML()
		} else {
			if s, err := ptst.NewTexto(vdomObjeto); err == nil {
				ssrHtml = string(s.(ptst.Texto))
			}
		}

		// 2. Extrai e gera metadados AEO/GEO e OpenGraph
		headMetaTags := strings.Builder{}
		if metadados != nil && metadados != ptst.Nulo {
			if mapa, ok := metadados.(ptst.Mapa); ok {
				// Titulo
				if t, ok := mapa["titulo"]; ok {
					if txt, ok := t.(ptst.Texto); ok {
						htmlStr = strings.Replace(htmlStr, "<title>Portuscript App</title>", fmt.Sprintf("<title>%s</title>", txt), 1)
						headMetaTags.WriteString(fmt.Sprintf("\n    <meta property=\"og:title\" content=\"%s\">", txt))
					}
				}
				// Descricao
				if d, ok := mapa["descricao"]; ok {
					if txt, ok := d.(ptst.Texto); ok {
						headMetaTags.WriteString(fmt.Sprintf("\n    <meta name=\"description\" content=\"%s\">", txt))
						headMetaTags.WriteString(fmt.Sprintf("\n    <meta property=\"og:description\" content=\"%s\">", txt))
					}
				}
				// Imagem
				if img, ok := mapa["imagem"]; ok {
					if txt, ok := img.(ptst.Texto); ok {
						headMetaTags.WriteString(fmt.Sprintf("\n    <meta property=\"og:image\" content=\"%s\">", txt))
					}
				}

				// Dados Estruturados Schema.org (AEO)
				if esquema, ok := mapa["esquema"]; ok {
					if m, ok := esquema.(ptst.Mapa); ok {
						rawMap := make(map[string]interface{})
						for k, v := range m {
							if txt, ok := v.(ptst.Texto); ok {
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

				// Geolocalização (GEO)
				if geo, ok := mapa["geo"]; ok {
					if m, ok := geo.(ptst.Mapa); ok {
						lat := m["latitude"]
						lon := m["longitude"]
						if lat != nil && lon != nil {
							headMetaTags.WriteString(fmt.Sprintf("\n    <meta name=\"geo.position\" content=\"%v; %v\">", lat, lon))
							headMetaTags.WriteString(fmt.Sprintf("\n    <meta name=\"ICBM\" content=\"%v, %v\">", lat, lon))
						}
					}
				}
			}
		}

		// Injeta as tags head adicionais antes do fechamento de </head>
		if headMetaTags.Len() > 0 {
			htmlStr = strings.Replace(htmlStr, "</head>", headMetaTags.String()+"\n</head>", 1)
		}

		// 3. Substitui para realizar SSR e Hidratação
		htmlStr = strings.Replace(htmlStr, "<div id=\"app\"></div>", fmt.Sprintf("<div id=\"app\">%s</div>", ssrHtml), 1)

		res.Status = 200
		res.Cabecalho["Content-Type"] = ptst.Texto("text/html; charset=utf-8")
		res.Corpo = ptst.Texto(htmlStr)
		return ptst.Nulo, nil
	}, "")
}
