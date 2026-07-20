package hrp

import (
	"fmt"
	"strings"
)

// ElementoJSX representa um nó de Virtual DOM avaliado no backend para fins de SSR.
type ElementoJSX struct {
	Tag       string
	Atributos map[string]Objeto
	Filhos    []Objeto
}

var TipoElementoJSX = NewTipo("ElementoJSX", "Nó do Virtual DOM no backend para SSR")

func (e *ElementoJSX) Tipo() *Tipo {
	return TipoElementoJSX
}

func (e *ElementoJSX) M__texto__() (Objeto, error) {
	return Texto(e.RenderizarHTML()), nil
}

// RenderizarHTML converte a árvore de nós JSX em uma string HTML estática limpa.
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
