package cmd

import (
	"regexp"
	"strings"
)

// ponytail: tabela de mapeamento estático de classes utilitárias tipo "Tailwind em PT-BR".
// Isso permite usar classes rápidas em português que compilam em CSS real minúsculo sob demanda.
var tabelaUtilitariosPT = map[string]string{
	// display & flexbox
	"flex-linha":        "display: flex; flex-direction: row;",
	"flex-coluna":       "display: flex; flex-direction: column;",
	"itens-centro":      "align-items: center;",
	"itens-inicio":      "align-items: flex-start;",
	"itens-fim":         "align-items: flex-end;",
	"conteudo-centro":   "justify-content: center;",
	"conteudo-inicio":   "justify-content: flex-start;",
	"conteudo-fim":      "justify-content: flex-end;",
	"conteudo-espacado": "justify-content: space-between;",
	"flex-embrulhar":    "flex-wrap: wrap;",
	"flex-1":            "flex: 1 1 0%;",

	// espaçamento padding
	"p-0": "padding: 0px;",
	"p-1": "padding: 0.25rem;",
	"p-2": "padding: 0.5rem;",
	"p-3": "padding: 0.75rem;",
	"p-4": "padding: 1rem;",
	"p-6": "padding: 1.5rem;",
	"p-8": "padding: 2rem;",
	"px-1": "padding-inline: 0.25rem;",
	"px-2": "padding-inline: 0.5rem;",
	"px-4": "padding-inline: 1rem;",
	"py-1": "padding-block: 0.25rem;",
	"py-2": "padding-block: 0.5rem;",
	"py-4": "padding-block: 1rem;",

	// espaçamento margin
	"m-0": "margin: 0px;",
	"m-1": "margin: 0.25rem;",
	"m-2": "margin: 0.5rem;",
	"m-3": "margin: 0.75rem;",
	"m-4": "margin: 1rem;",
	"m-6": "margin: 1.5rem;",
	"m-8": "margin: 2rem;",
	"mx-1": "margin-inline: 0.25rem;",
	"mx-2": "margin-inline: 0.5rem;",
	"mx-4": "margin-inline: 1rem;",
	"my-1": "margin-block: 0.25rem;",
	"my-2": "margin-block: 0.5rem;",
	"my-4": "margin-block: 1rem;",

	// dimensões
	"largura-cheia": "width: 100%;",
	"altura-cheia":  "height: 100%;",
	"w-cheia":       "width: 100%;",
	"h-cheia":       "height: 100%;",
	"largura-tela":  "width: 100vw;",
	"altura-tela":   "height: 100vh;",

	// cores de texto
	"texto-branco": "color: #ffffff;",
	"texto-preto":  "color: #000000;",
	"texto-cinza":  "color: #6b7280;",
	"texto-azul":   "color: #3b82f6;",
	"texto-verde":  "color: #10b981;",
	"texto-vermelho": "color: #ef4444;",

	// cores de fundo
	"fundo-branco": "background-color: #ffffff;",
	"fundo-preto":  "background-color: #000000;",
	"fundo-cinza":  "background-color: #f3f4f6;",
	"fundo-azul":   "background-color: #3b82f6;",
	"fundo-verde":  "background-color: #10b981;",
	"fundo-vermelho": "background-color: #ef4444;",

	// bordas & cantos
	"borda":        "border: 1px solid #e5e7eb;",
	"raio-pequeno": "border-radius: 0.125rem;",
	"raio-medio":   "border-radius: 0.375rem;",
	"raio-grande":  "border-radius: 0.5rem;",
	"raio-cheio":   "border-radius: 9999px;",
}

// reClasse captura classes usadas no JS transpilado, ex: `classe: "p-4 itens-centro"`
var reClasse = regexp.MustCompile(`(?i)classe:\s*"([^"]+)"`)

// extraiEGerarCssUtilitarios varre o código transpilado JS, identifica as classes PT utilitárias
// usadas e gera as correspondentes regras CSS estáticas prontas para estilos.css.
func extraiEGerarCssUtilitarios(js string) string {
	classesDetectadas := make(map[string]bool)

	// Acha todos os matches de classes
	matches := reClasse.FindAllStringSubmatch(js, -1)
	for _, m := range matches {
		if len(m) > 1 {
			// Divide classes por espaço (pode haver múltiplas como "flex-linha p-4")
			partes := strings.Fields(m[1])
			for _, classe := range partes {
				classesDetectadas[classe] = true
			}
		}
	}

	var sb strings.Builder
	sb.WriteString("/* --- Classes Utilitárias PT-BR --- */\n")
	gerouAlguma := false

	for classe := range classesDetectadas {
		if regra, ok := tabelaUtilitariosPT[classe]; ok {
			sb.WriteString(".")
			sb.WriteString(classe)
			sb.WriteString(" {\n  ")
			sb.WriteString(regra)
			sb.WriteString("\n}\n\n")
			gerouAlguma = true
		}
	}

	if !gerouAlguma {
		return ""
	}

	return sb.String()
}
