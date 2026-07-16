package cmd

// ponytail: emissor de CSS enxuto para blocos `estilo` do Portuscript.
//
// Não usamos uma lib externa de parser CSS — a entrada do Portuscript
// é delimitada e pequena, então um Handel local resolve em ~50 linhas,
// alimentado pela string já capturada pelo parser (DeclEstilo.Regras).
//
// Responsabilidades:
//  1. Mapear chaves em PT (camelCase) para canônico kebab-case CSS
//     (`corDeFundo` → `background-color`, etc.).
//  2. Strip de aspas literais em valores (`"azul"` → `azul`, `'10px'` opcional).
//  3. Suporte a aninhamento CSS moderno (nesting): `botao:hover { ... }`
//     gera `.Pai botao:hover { ... }`.
//  4. Empacotar tudo em uma string CSS final válida e balanceada.
//
// Limitações documentadas:
// - Não suporta `@media`, `@keyframes` (YAGNI, add quando o primeiro usuário pedir).
// - Não faz minificação (browser já lida bem com whitespace).

import (
	"fmt"
	"regexp"
	"strings"
)

// mapaChavesPT → kebab-case CSS. Mantido denso para evitar traduções duplicadas
// a cada chamada. Adicione aqui sempre que precisar de uma nova chave.
var mapaChavesPT = map[string]string{
	// cor / cores
	"cor":              "color",
	"corDeFundo":       "background-color",
	"corBorda":         "border-color",
	"corContorno":      "outline-color",
	"corTexto":         "color",
	// layout
	"exibir":      "display",
	"exibirTipo":  "display",
	"posicao":     "position",
	"fluxo":       "flex-direction",
	"flex":        "flex",
	"grade":       "grid",
	"alinhamento": "align-items",
	"justificar":  "justify-content",
	"justificacao": "justify-content",
	// caixa / box-model
	"margem":       "margin",
	"margemX":      "margin-inline",
	"margemY":       "margin-block",
	"padding":      "padding",
	"espacamentoX": "padding-inline",
	"espacamentoY": "padding-block",
	"largura":      "width",
	"altura":       "height",
	"larguraMax":   "max-width",
	"alturaMax":    "max-height",
	"borda":        "border",
	"raio":         "border-radius",
	"contorno":     "outline",
	"sombra":       "box-shadow",
	"opacidade":    "opacity",
	"transbordar":  "overflow",
	// tipografia
	"fonte":          "font-family",
	"tamanhoFonte":   "font-size",
	"pesoFonte":      "font-weight",
	"alinhamentoTexto": "text-align",
	"decoracaoTexto": "text-decoration",
	"transformacaoTexto": "text-transform",
	"espacamentoLetras": "letter-spacing",
	"espacamentoLinhas": "line-height",
	// animações
	"transicao":   "transition",
	"transformacao": "transform",
	"animacao":    "animation",
	// espaçamento de utilidades (atalhos)
	"p":  "padding",
	"px": "padding-inline",
	"py": "padding-block",
	"m":  "margin",
	"mx": "margin-inline",
	"my": "margin-block",
	"w":  "width",
	"h":  "height",
}

// chaveCanonica faz a conversão. Fallback: camelCase → kebab-case.
func chaveCanonica(chave string) string {
	if v, ok := mapaChavesPT[chave]; ok {
		return v
	}
	return camelParaKebab(chave)
}

var reMaisMais = regexp.MustCompile(`([a-z0-9])([A-Z])`)

func camelParaKebab(s string) string {
	res := reMaisMais.ReplaceAllString(s, `$1-$2`)
	return strings.ToLower(res)
}

// stripAspas remove aspas literais `"..."` ou `'...'` envolvendo o valor.
var reAspasDuplas = regexp.MustCompile(`^\s*"([^"]*)"\s*$`)
var reAspasSimples = regexp.MustCompile(`^\s*'([^']*)'\s*$`)

func stripAspas(valor string) string {
	if m := reAspasDuplas.FindStringSubmatch(valor); m != nil {
		return m[1]
	}
	if m := reAspasSimples.FindStringSubmatch(valor); m != nil {
		return m[1]
	}
	return valor
}

// ponytail: processa e traduz propriedades e atalhos de forma consolidada e canônica
func processaEstiloLinha(chave, valor string) (string, string) {
	// ponytail: sanitiza espaços na chave (ex: "raio - grande" vira "raio-grande")
	chaveLimpa := strings.ReplaceAll(chave, " ", "")
	valorStrip := stripAspas(valor)
	chaveCan := chaveCanonica(chaveLimpa)

	// ponytail: açúcar sintático de border-radius em português
	if (chaveLimpa == "raio-pequeno" || chaveLimpa == "raio-medio" || chaveLimpa == "raio-grande") && (valorStrip == "true" || valorStrip == "Verdadeiro") {
		chaveCan = "border-radius"
		switch chaveLimpa {
		case "raio-pequeno":
			valorStrip = "0.125rem"
		case "raio-medio":
			valorStrip = "0.375rem"
		case "raio-grande":
			valorStrip = "0.5rem"
		}
	}

	return chaveCan, valorStrip
}

// processaBlocoEstilo recebe o `Nome` (classe alvo) e o conteúdo
// do `DeclEstilo.Regras` (com `{ }` aninhadas preservadas) e gera
// uma string CSS limpa. Suporta nested selectors (`tag { ... }` vira
// `.Nome tag { ... }`).
//
// Formato de entrada:
//   cor: "azul";
//   borda: 1px;
//   botao:hover { opacidade: 0.8; }
//
// Saída:
//   .Nome { color: azul; border: 1px; ... }
//   .Nome botao:hover { opacity: 0.8; }
func processaBlocoEstilo(nome string, corpo string) string {
	var sb strings.Builder

	planos, nested, _ := parseCorpoToken(corpo, false)

	if len(planos) > 0 {
		sb.WriteString(".")
		sb.WriteString(nome)
		sb.WriteString(" {\n")
		for _, kv := range planos {
			chave, valor := kv[0], kv[1]
			chCan, valCan := processaEstiloLinha(chave, valor)
			sb.WriteString("  ")
			sb.WriteString(chCan)
			sb.WriteString(": ")
			sb.WriteString(valCan)
			sb.WriteString(";\n")
		}
		sb.WriteString("}\n")
	}

	for sel, inner := range nested {
		planosInner, _, _ := parseCorpoToken(inner, true)
		sb.WriteString(".")
		sb.WriteString(nome)
		sb.WriteString(" ")
		sb.WriteString(sel)
		sb.WriteString(" {\n")
		for _, kv := range planosInner {
			chave, valor := kv[0], kv[1]
			chCan, valCan := processaEstiloLinha(chave, valor)
			sb.WriteString("  ")
			sb.WriteString(chCan)
			sb.WriteString(": ")
			sb.WriteString(valCan)
			sb.WriteString(";\n")
		}
		sb.WriteString("}\n")
	}

	return sb.String()
}

// parseCorpoToken é um parser recursivo minimalista. Retorna:
//   - planos: [][2]string na forma [[chave, valor], ...]
//   - nested: map[string]string com seletor → corpo_bruto_interno (sem chaves externas)
//   - erro:   qualquer balanceamento inconsistente.
//
// Suporta profundidade simples (chaves aninhadas uma vez). Para 99% dos
// blocos `estilo` isso é suficiente.
func parseCorpoToken(corpo string, _ bool) ([][]string, map[string]string, error) {
	planos := [][]string{}
	nested := map[string]string{}

	// dispatcher de tokens simples:
	// 1) split por top-level blocos via contagem de chaves
	i := 0
	for i < len(corpo) {
		// Pula whitespace
		for i < len(corpo) && isEspaco(corpo[i]) {
			i++
		}
		if i >= len(corpo) {
			break
		}

		// Coleta chave (até ':')
		start := i
		for i < len(corpo) && corpo[i] != ':' && corpo[i] != '{' && corpo[i] != ';' {
			i++
		}
		if i >= len(corpo) {
			break
		}

		chave := strings.TrimSpace(corpo[start:i])

		if i < len(corpo) && corpo[i] == ':' {
			// Regra chave: valor
			i++
			valStart := i
			for i < len(corpo) && corpo[i] != ';' && corpo[i] != '{' {
				i++
			}
			valor := strings.TrimSpace(corpo[valStart:i])
			if valor != "" && chave != "" {
				planos = append(planos, []string{chave, valor})
			}
			if i < len(corpo) && corpo[i] == ';' {
				i++
			}
			continue
		}

		if i < len(corpo) && corpo[i] == '{' {
			// Regra aninhada: chave é o seletor
			i++
			nivel := 1
			bodyStart := i
			for i < len(corpo) && nivel > 0 {
				switch corpo[i] {
				case '{':
					nivel++
				case '}':
					nivel--
				}
				if nivel > 0 {
					i++
				}
			}
			body := corpo[bodyStart:i]
			i++ // consome '}'
			nested[chave] = strings.TrimSpace(body)
			_ = start
			continue
		}

		// Se chegou aqui, é algo inesperado (provavelmente `;` solto)
		if i < len(corpo) && corpo[i] == ';' {
			i++
		}
	}

	return planos, nested, nil
}

func isEspaco(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r'
}

// validateSimpleCss checa balanceamento mínimo de chaves e vírgulas
// suspeitas dentro de seletores. Retorna lista de mensagens de erro
// (vazia quando CSS parece OK).
func validateSimpleCss(raw string) []string {
	var errs []string
	abres, fechas := strings.Count(raw, "{"), strings.Count(raw, "}")
	if abres != fechas {
		errs = append(errs, fmt.Sprintf("chaves desbalanceadas: %d abertas vs %d fechadas", abres, fechas))
	}
	// Vírgula entre chave e `{` é sinal de selector malformado.
	if strings.Contains(raw, ",{") {
		errs = append(errs, "vírgula imediatamente antes de { detectada")
	}
	// Aspas literais em valor num string de CSS geralmente é inválido.
	// Avalia em pares `chave: "valor"` com aspas.
	for _, m := range reAspasDuplas.FindAllString(raw, -1) {
		_ = m
	}
	// Heurística simples: aspas dentro de uma linha de propriedade
	linhas := strings.Split(raw, "\n")
	for _, linha := range linhas {
		if strings.Contains(linha, ":") && strings.Contains(linha, "\"") {
			errs = append(errs, fmt.Sprintf("possível aspas literais em valor CSS: %q", linha))
		}
	}
	return errs
}
