// Package hrp implementa as estruturas do runtime da linguagem Harpia.
package hrp

import (
	"fmt"
	"strings"
)

// ElementoJSX representa um nó de Virtual DOM avaliado no backend para fins de SSR.
// Carrega as propriedades de Tag HTML, dicionário reativo de Atributos e a lista de Filhos aninhados.
type ElementoJSX struct {
	Tag       string            // Nome da tag HTML correspondente (ex: "div", "button").
	Atributos map[string]Objeto // Mapeamento de atributos ou diretivas do componente.
	Filhos    []Objeto          // Lista de sub-elementos físicos ou strings filhas de SSR.
}

// TipoElementoJSX especifica o metadado de classe do ElementoJSX no runtime.
var TipoElementoJSX = NewTipo("ElementoJSX", "Nó do Virtual DOM no backend para SSR")

// Tipo retorna a representação de classe (Tipo) da struct ElementoJSX.
func (e *ElementoJSX) Tipo() *Tipo {
	return TipoElementoJSX
}

// M__texto__ satisfaz o protocolo de coerção textual da VM (I__texto__),
// convertendo a árvore JSX diretamente em string HTML nativa.
func (e *ElementoJSX) M__texto__() (Objeto, error) {
	return Texto(e.RenderizarHTML()), nil
}

// tagsVoidHTML5 define as tags HTML5 auto-fechadas que não possuem conteúdo interno nem tag de fechamento.
var tagsVoidHTML5 = map[string]bool{
	"area": true, "base": true, "br": true, "col": true, "embed": true,
	"hr": true, "img": true, "input": true, "link": true, "meta": true,
	"param": true, "source": true, "track": true, "wbr": true,
}

func escaparAtributoHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}

func escaparConteudoHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}

// RenderizarHTML converte a árvore de nós JSX em uma string HTML estática limpa e segura contra XSS.
// Trata o mapeamento de classes em português ("classe" -> "class"), filtra eventos e callbacks
// de runtime do frontend, sanitiza atributos/conteúdo e respeita a especificação de tags Void do HTML5.
func (e *ElementoJSX) RenderizarHTML() string {
	tagLower := strings.ToLower(strings.TrimSpace(e.Tag))
	if tagLower == "fragmento" || tagLower == "fragment" {
		var sb strings.Builder
		for _, filho := range e.Filhos {
			if filho == nil || filho == Nulo {
				continue
			}
			if el, ok := filho.(*ElementoJSX); ok {
				sb.WriteString(el.RenderizarHTML())
			} else if txt, ok := filho.(Texto); ok {
				sb.WriteString(escaparConteudoHTML(string(txt)))
			} else {
				if s, err := NewTexto(filho); err == nil {
					sb.WriteString(escaparConteudoHTML(string(s.(Texto))))
				} else {
					sb.WriteString(escaparConteudoHTML(fmt.Sprintf("%v", filho)))
				}
			}
		}
		return sb.String()
	}

	var sb strings.Builder
	if tagLower == "" {
		tagLower = "div"
	}

	sb.WriteString("<")
	sb.WriteString(tagLower)

	// Trata atributos
	for k, v := range e.Atributos {
		attrName := k
		// Mapeamentos em português
		if attrName == "classe" {
			attrName = "class"
		} else if attrName == "aoClicar" || strings.HasPrefix(attrName, "ao") {
			// Ignora callbacks de eventos dinâmicos no HTML de SSR
			continue
		}

		valText := ""
		if txt, ok := v.(Texto); ok {
			valText = string(txt)
		} else {
			if s, err := NewTexto(v); err == nil {
				valText = string(s.(Texto))
			} else {
				valText = fmt.Sprintf("%v", v)
			}
		}
		sb.WriteString(fmt.Sprintf(` %s="%s"`, attrName, escaparAtributoHTML(valText)))
	}

	sb.WriteString(">")

	// Se for tag Void (auto-fechada), encerra sem filhos e sem tag de fechamento
	if tagsVoidHTML5[tagLower] {
		return sb.String()
	}

	// Filhos
	for _, filho := range e.Filhos {
		if filho == nil || filho == Nulo {
			continue
		}
		if el, ok := filho.(*ElementoJSX); ok {
			sb.WriteString(el.RenderizarHTML())
		} else if txt, ok := filho.(Texto); ok {
			sb.WriteString(escaparConteudoHTML(string(txt)))
		} else {
			if s, err := NewTexto(filho); err == nil {
				sb.WriteString(escaparConteudoHTML(string(s.(Texto))))
			} else {
				sb.WriteString(escaparConteudoHTML(fmt.Sprintf("%v", filho)))
			}
		}
	}

	sb.WriteString("</")
	sb.WriteString(tagLower)
	sb.WriteString(">")
	return sb.String()
}
