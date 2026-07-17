package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mat-dgruber/Harpia/parser"
	"github.com/mat-dgruber/Harpia/ptst"
)

// TranspilerWeb converte uma AST do Harpia para código JavaScript ES6 correspondente.
type TranspilerWeb struct {
	Styles        []string // Acumula blocos de estilo declarados para salvar no CSS
	Estiro        bool     // ponytail: deprecating typo anterior
	Estrito       bool     // ponytail: ativa tipagem estrita de JSDoc para DX
	DiretorioBase string   // ponytail: diretório base do arquivo que está sendo compilado
}

func (t *TranspilerWeb) Transpile(node parser.BaseNode) string {
	if node == nil {
		return ""
	}

	switch n := node.(type) {
	case *parser.Programa:
		var sb strings.Builder
		for _, decl := range n.Declaracoes {
			sb.WriteString(t.Transpile(decl))
			sb.WriteString("\n")
		}
		return sb.String()

	case *parser.DeclVar:
		keyword := "let"
		if n.Constante {
			keyword = "const"
		}
		initVal := "undefined"
		if n.Inicializador != nil {
			initVal = t.Transpile(n.Inicializador)
		}
		return fmt.Sprintf("%s %s = %s;", keyword, n.Nome, initVal)

	case *parser.Reatribuicao:
		dest := t.Transpile(n.Objeto)
		val := t.Transpile(n.Expressao)
		return fmt.Sprintf("%s %s %s;", dest, n.Operador, val)

	case *parser.Identificador:
		return n.Nome

	case *parser.TextoLiteral:
		if len(n.Valor) >= 2 {
			valorLimpo := n.Valor[1 : len(n.Valor)-1]
			return fmt.Sprintf(`"%s"`, strings.ReplaceAll(valorLimpo, `"`, `\"`))
		}
		return `""`

	case *parser.InteiroLiteral:
		return n.Valor

	case *parser.DecimalLiteral:
		return n.Valor

	case *parser.ConstanteLiteral:
		switch n.Valor {
		case "Verdadeiro":
			return "true"
		case "Falso":
			return "false"
		case "Nulo":
			return "null"
		}
		return "undefined"

	case *parser.OpBinaria:
		esq := t.Transpile(n.Esq)
		dir := t.Transpile(n.Dir)
		op := n.Operador
		// Mapeamentos de operadores do Harpia para JS
		if op == "e" {
			op = "&&"
		} else if op == "ou" {
			op = "||"
		} else if op == "//" {
			return fmt.Sprintf("Math.floor(%s / %s)", esq, dir)
		}
		return fmt.Sprintf("(%s %s %s)", esq, op, dir)

	case *parser.OpUnaria:
		exp := t.Transpile(n.Expressao)
		op := n.Operador
		if op == "nao" {
			op = "!"
		}
		return fmt.Sprintf("(%s%s)", op, exp)

	case *parser.DeclFuncao:
		var params []string
		var jsdoc strings.Builder

		if t.Estrito {
			jsdoc.WriteString("/**\n")
			for _, p := range n.Parametros {
				tipoJS := "any"
				if p.Tipo != "" {
					switch p.Tipo {
					case "Inteiro", "Decimal":
						tipoJS = "number"
					case "Texto":
						tipoJS = "string"
					case "Booleano":
						tipoJS = "boolean"
					default:
						tipoJS = p.Tipo
					}
				}
				jsdoc.WriteString(fmt.Sprintf(" * @param {%s} %s\n", tipoJS, p.Nome))
			}
			if n.TipoRetorno != "" {
				tipoRet := "any"
				switch n.TipoRetorno {
				case "Inteiro", "Decimal":
					tipoRet = "number"
				case "Texto":
					tipoRet = "string"
				case "Booleano":
					tipoRet = "boolean"
				default:
					tipoRet = n.TipoRetorno
				}
				jsdoc.WriteString(fmt.Sprintf(" * @returns {%s}\n", tipoRet))
			} else {
				jsdoc.WriteString(" * @returns {any}\n")
			}
			jsdoc.WriteString(" */\n")
		}

		for _, p := range n.Parametros {
			if p.Padrao != nil {
				params = append(params, fmt.Sprintf("%s = %s", p.Nome, t.Transpile(p.Padrao)))
			} else {
				params = append(params, p.Nome)
			}
		}
		asyncPrefix := ""
		if n.Assincrono {
			asyncPrefix = "async "
		}
		body := t.Transpile(n.Corpo)
		return fmt.Sprintf("%s%sfunction %s(%s) %s", jsdoc.String(), asyncPrefix, n.Nome, strings.Join(params, ", "), body)

	case *parser.Bloco:
		var sb strings.Builder
		sb.WriteString("{\n")
		for _, decl := range n.Declaracoes {
			sb.WriteString(t.Transpile(decl))
			sb.WriteString("\n")
		}
		sb.WriteString("}")
		return sb.String()

	case *parser.RetorneNode:
		if n.Expressao == nil {
			return "return;"
		}
		return fmt.Sprintf("return %s;", t.Transpile(n.Expressao))

	case *parser.ChamadaFuncao:
		fn := t.Transpile(n.Identificador)
		// ponytail: inline dinâmico de layouts HTML físicos externos
		if fn == "importarHtml" && len(n.Argumentos) == 1 {
			if txt, ok := n.Argumentos[0].(*parser.TextoLiteral); ok {
				caminho := txt.Valor
				if len(caminho) >= 2 {
					caminho = caminho[1 : len(caminho)-1]
				}
				caminhoCompleto := filepath.Join(t.DiretorioBase, caminho)
				conteudo, err := os.ReadFile(caminhoCompleto)
				if err != nil {
					return fmt.Sprintf("/* Erro ao carregar html de %s: %v */", caminho, err)
				}
				ctx := ptst.NewContexto(ptst.OpcsContexto{})
				defer ctx.Terminar()
				subAst, err := ctx.StringParaAst(string(conteudo), caminho)
				if err != nil {
					return fmt.Sprintf("/* Erro de sintaxe no template HTML de %s: %v */", caminho, err)
				}
				return t.Transpile(subAst)
			}
		}
		var args []string
		for _, arg := range n.Argumentos {
			args = append(args, t.Transpile(arg))
		}
		return fmt.Sprintf("%s(%s)", fn, strings.Join(args, ", "))

	case *parser.AcessoMembro:
		dono := t.Transpile(n.Dono)
		membro := t.Transpile(n.Membro)
		return fmt.Sprintf("%s.%s", dono, membro)

	case *parser.Indexacao:
		obj := t.Transpile(n.Objeto)
		arg := t.Transpile(n.Argumento)
		return fmt.Sprintf("%s[%s]", obj, arg)

	case *parser.ExpressaoSe:
		cond := t.Transpile(n.Condicao)
		corpo := t.Transpile(n.Corpo)
		alt := ""
		if n.Alternativa != nil {
			alt = " else " + t.Transpile(n.Alternativa)
		}
		return fmt.Sprintf("if (%s) %s%s", cond, corpo, alt)

	case *parser.Enquanto:
		cond := t.Transpile(n.Condicao)
		corpo := t.Transpile(n.Corpo)
		return fmt.Sprintf("while (%s) %s", cond, corpo)

	case *parser.BlocoPara:
		iter := t.Transpile(n.Iterador)
		corpo := t.Transpile(n.Corpo)
		return fmt.Sprintf("for (const %s of %s) %s", n.Identificador, iter, corpo)

	case *parser.AguardeNode:
		return fmt.Sprintf("await %s", t.Transpile(n.Expressao))

	case *parser.DeclClasse:
		extends := ""
		if n.Heranca != "" {
			extends = " extends " + n.Heranca
		}
		var methods []string
		for _, m := range n.Metodos {
			// Métodos mágicos de inicialização no Harpia convertem para constructor()
			name := m.Nome
			if name == "__init__" || name == "inicializar" {
				name = "constructor"
			}
			var params []string
			for _, p := range m.Parametros {
				if p.Padrao != nil {
					params = append(params, fmt.Sprintf("%s = %s", p.Nome, t.Transpile(p.Padrao)))
				} else {
					params = append(params, p.Nome)
				}
			}
			body := t.Transpile(m.Corpo)
			methods = append(methods, fmt.Sprintf("  %s(%s) %s", name, strings.Join(params, ", "), body))
		}
		return fmt.Sprintf("class %s%s {\n%s\n}", n.Nome, extends, strings.Join(methods, "\n"))

	case *parser.NovaNode:
		return fmt.Sprintf("new %s", t.Transpile(n.Objeto))

	case *parser.OpPipe:
		esq := t.Transpile(n.Esq)
		// Operador pipe (x |> dobrar) vira dobrar(x) ou dobrar(x, args) se for chamada
		switch d := n.Dir.(type) {
		case *parser.ChamadaFuncao:
			fn := t.Transpile(d.Identificador)
			var args []string
			args = append(args, esq)
			for _, arg := range d.Argumentos {
				args = append(args, t.Transpile(arg))
			}
			return fmt.Sprintf("%s(%s)", fn, strings.Join(args, ", "))
		default:
			dir := t.Transpile(n.Dir)
			return fmt.Sprintf("%s(%s)", dir, esq)
		}

	case *parser.DeclExportar:
		return fmt.Sprintf("export %s", t.Transpile(n.Expressao))

	case *parser.ListaLiteral:
		var elems []string
		for _, el := range n.Elementos {
			elems = append(elems, t.Transpile(el))
		}
		return fmt.Sprintf("[%s]", strings.Join(elems, ", "))

	case *parser.MapaLiteral:
		var entries []string
		for _, entry := range n.Entradas {
			key := t.Transpile(entry.Chave)
			val := t.Transpile(entry.Valor)
			entries = append(entries, fmt.Sprintf("%s: %s", key, val))
		}
		return fmt.Sprintf("{ %s }", strings.Join(entries, ", "))

	case *parser.TemplateLiteral:
		var parts []string
		for _, part := range n.Partes {
			switch p := part.(type) {
			case *parser.TextoLiteral:
				parts = append(parts, p.Valor)
			case *parser.TemplateExpr:
				parts = append(parts, fmt.Sprintf("${%s}", t.Transpile(p.Expressao)))
			}
		}
		return fmt.Sprintf("`%s`", strings.Join(parts, ""))

	// ============================================================================
	// CASOS SINTÁTICOS DO FRONTEND (JSX & ESTILO)
	// ============================================================================
	case *parser.NoJSX:
		if n.Tag == "Link" || n.Tag == "link" {
			var para string = `"#"`
			for _, attr := range n.Atributos {
				if attr.Nome == "para" {
					para = t.Transpile(attr.Valor)
				}
			}
			var children []string
			for _, filho := range n.Filhos {
				if txt, ok := filho.(*parser.TextoLiteral); ok {
					trimmed := strings.TrimSpace(txt.Valor)
					if trimmed == "" {
						continue
					}
				}
				children = append(children, t.Transpile(filho))
			}
			childrenStr := "null"
			if len(children) > 0 {
				childrenStr = strings.Join(children, ", ")
			}
			return fmt.Sprintf("h('a', { href: %s, aoClicar: (e) => { e.preventDefault(); navegar(%s); } }, %s)", para, para, childrenStr)
		}

		var attrs []string
		for _, attr := range n.Atributos {
			val := "true"
			if attr.Valor != nil {
				val = t.Transpile(attr.Valor)
			}
			nomeAttr := attr.Nome

			// ponytail: mapeamento simples de binding bidirecional nativo
			if nomeAttr == "ligar" {
				nomeAttr = "_ligar"
			}

			// ponytail: açúcar sintático de modificadores de eventos (ex: aoEnviar_prevenir)
			if strings.HasPrefix(nomeAttr, "ao") && strings.Contains(nomeAttr, "_") {
				partes := strings.Split(nomeAttr, "_")
				eventoPrincipal := partes[0]
				modificadores := partes[1:]

				embrulho := val
				for _, mod := range modificadores {
					switch mod {
					case "prevenir":
						embrulho = fmt.Sprintf("(e) => { e.preventDefault(); (%s)(e); }", embrulho)
					case "parar":
						embrulho = fmt.Sprintf("(e) => { e.stopPropagation(); (%s)(e); }", embrulho)
					}
				}
				nomeAttr = eventoPrincipal
				val = embrulho
			}

			if nomeAttr == "classe" {
				nomeAttr = "classe" // O runtime cuida de mapear pra class
			}
			attrs = append(attrs, entries(nomeAttr, val))
		}
		var children []string
		for _, filho := range n.Filhos {
			// Se o filho for apenas texto literal não formatado, limpa-se os espaços ou trata-se adequadamente
			if txt, ok := filho.(*parser.TextoLiteral); ok {
				trimmed := strings.TrimSpace(txt.Valor)
				if trimmed == "" {
					continue
				}
			}
			children = append(children, t.Transpile(filho))
		}
		attrsObj := "{}"
		if len(attrs) > 0 {
			attrsObj = fmt.Sprintf("{ %s }", strings.Join(attrs, ", "))
		}
		// Transpila para chamada h() do runtime VDOM
		tagArg := fmt.Sprintf("'%s'", n.Tag)
		if len(n.Tag) > 0 && n.Tag[0] >= 'A' && n.Tag[0] <= 'Z' {
			tagArg = n.Tag
		}
		return fmt.Sprintf("h(%s, %s, %s)", tagArg, attrsObj, strings.Join(children, ", "))

	case *parser.NoSeJSX:
		cond := t.Transpile(n.Condicao)
		var children []string
		for _, filho := range n.Filhos {
			children = append(children, t.Transpile(filho))
		}
		childVal := "null"
		if len(children) > 0 {
			childVal = strings.Join(children, ", ")
		}
		return fmt.Sprintf("(%s ? %s : null)", cond, childVal)

	case *parser.NoParaJSX:
		list := t.Transpile(n.Lista)
		var children []string
		for _, filho := range n.Filhos {
			children = append(children, t.Transpile(filho))
		}
		childVal := "null"
		if len(children) > 0 {
			childVal = strings.Join(children, ", ")
		}
		return fmt.Sprintf("(%s).map(%s => %s)", list, n.Item, childVal)

	case *parser.ImporteDe:
		caminho := n.Caminho.Valor
		if len(caminho) >= 2 {
			caminho = caminho[1 : len(caminho)-1]
		}
		// ponytail: trata imports de estilos .estilo.ptst e os resolve para constantes locais
		if strings.HasSuffix(caminho, ".estilo.ptst") {
			caminhoCompleto := filepath.Join(t.DiretorioBase, caminho)
			conteudo, err := os.ReadFile(caminhoCompleto)
			if err == nil {
				ctx := ptst.NewContexto(ptst.OpcsContexto{})
				defer ctx.Terminar()
				styleAst, err := ctx.StringParaAst(string(conteudo), caminho)
				if err == nil {
					t.Transpile(styleAst)
				}
			}
			var sb strings.Builder
			for _, nome := range n.Nomes {
				sb.WriteString(fmt.Sprintf("const %s = \"%s\";\n", nome, nome))
			}
			return sb.String()
		}

		jsPath := caminho
		if strings.HasSuffix(jsPath, ".ptst") {
			jsPath = jsPath[:len(jsPath)-5] + ".js"
		}
		if !strings.HasPrefix(jsPath, ".") && !strings.HasPrefix(jsPath, "/") && !strings.Contains(jsPath, "://") {
			if jsPath == "web" {
				jsPath = "./runtime-web.js"
			}
		}
		return fmt.Sprintf("import { %s } from \"%s\";", strings.Join(n.Nomes, ", "), jsPath)

	case *parser.DeclEstilo:
		// Enviar pelo pipeline CSS enxuto: mapa canônico PT→CSS,
		// strip de aspas e suporte a nesting.
		cssBlock := processaBlocoEstilo(n.Nome, n.Regras)
		t.Styles = append(t.Styles, cssBlock)
		return "" // Sai do .js; alimenta somente estilos.css
	}

	return fmt.Sprintf("/* Erro ao transpilar tipo %T */", node)
}

func entries(key, val string) string {
	return fmt.Sprintf("%s: %s", key, val)
}
