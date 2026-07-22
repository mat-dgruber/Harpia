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

// RenderizarHTML converte a árvore de nós JSX em uma string HTML estática limpa.
// Trata o mapeamento de classes em português ("classe" -> "class"), filtra eventos e callbacks
// de runtime do frontend, trata aspas em atributos e concatena de forma recursiva todos os filhos.
func (e *ElementoJSX) RenderizarHTML() string {
	var sb strings.Builder
	sb.WriteString("<")
	sb.WriteString(e.Tag)

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
		sb.WriteString(fmt.Sprintf(` %s="%s"`, attrName, strings.ReplaceAll(valText, `"`, `\"`)))
	}

	sb.WriteString(">")

	// Filhos
	for _, filho := range e.Filhos {
		if filho == nil || filho == Nulo {
			continue
		}
		if el, ok := filho.(*ElementoJSX); ok {
			sb.WriteString(el.RenderizarHTML())
		} else if txt, ok := filho.(Texto); ok {
			sb.WriteString(string(txt))
		} else {
			if s, err := NewTexto(filho); err == nil {
				sb.WriteString(string(s.(Texto)))
			} else {
				sb.WriteString(fmt.Sprintf("%v", filho))
			}
		}
	}

	// Não fecha tags auto-fechadas comuns do HTML5 (opcional, mas seguro fechar todas para VDOM)
	sb.WriteString("</")
	sb.WriteString(e.Tag)
	sb.WriteString(">")
	return sb.String()
}
