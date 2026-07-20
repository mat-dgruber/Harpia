package cmd

import (
	"fmt"
	"strings"

	"github.com/mat-dgruber/Harpia/parser"
)

type TranspilerNative struct {
	varCount int
}

func (t *TranspilerNative) newVar() string {
	t.varCount++
	return fmt.Sprintf("v_%d", t.varCount)
}

func (t *TranspilerNative) Transpile(node parser.BaseNode) (string, string) {
	if node == nil {
		return "", "hrp.Nulo"
	}

	switch n := node.(type) {
	case *parser.Programa:
		var sb strings.Builder
		for _, decl := range n.Declaracoes {
			code, _ := t.Transpile(decl)
			sb.WriteString(code)
			sb.WriteString("\n")
		}
		return sb.String(), ""

	case *parser.DeclVar:
		var sb strings.Builder
		initCode, initVar := t.Transpile(n.Inicializador)
		sb.WriteString(initCode)

		// Define o valor no escopo léxico nativo e como variável Go local
		varName := fmt.Sprintf("var_%s", n.Nome)
		sb.WriteString(fmt.Sprintf("\t%s := %s\n", varName, initVar))
		sb.WriteString(fmt.Sprintf("\tescopo.DefinirSimbolo(hrp.NewVarSimbolo(\"%s\", %s))\n", n.Nome, varName))
		return sb.String(), varName

	case *parser.Reatribuicao:
		var sb strings.Builder
		valCode, valVar := t.Transpile(n.Expressao)
		sb.WriteString(valCode)

		if id, ok := n.Objeto.(*parser.Identificador); ok {
			varName := fmt.Sprintf("var_%s", id.Nome)
			sb.WriteString(fmt.Sprintf("\t%s = %s\n", varName, valVar))
			sb.WriteString(fmt.Sprintf("\tescopo.DefinirSimbolo(hrp.NewVarSimbolo(\"%s\", %s))\n", id.Nome, varName))
			return sb.String(), varName
		}
		return sb.String(), "hrp.Nulo"

	case *parser.Identificador:
		varName := t.newVar()
		code := fmt.Sprintf(`	var %s hrp.Objeto
	if val, errVal := escopo.ObterValor("%s"); errVal == nil {
		%s = val
	} else if val, errB := ctx.Modulos.Embutidos.M__obtem_attributo__("%s"); errB == nil {
		%s = val
	} else {
		panic(fmt.Sprintf("identificador '%s' não encontrado no escopo", "%s"))
	}
`, varName, n.Nome, varName, n.Nome, varName, n.Nome, n.Nome)
		return code, varName

	case *parser.TextoLiteral:
		varName := t.newVar()
		valorLimpo := n.Valor
		if len(n.Valor) >= 2 {
			valorLimpo = n.Valor[1 : len(n.Valor)-1]
		}
		code := fmt.Sprintf("\t%s := hrp.Texto(\"%s\")\n", varName, strings.ReplaceAll(valorLimpo, `"`, `\"`))
		return code, varName

	case *parser.InteiroLiteral:
		varName := t.newVar()
		code := fmt.Sprintf("\t%s := hrp.Inteiro(%s)\n", varName, n.Valor)
		return code, varName

	case *parser.DecimalLiteral:
		varName := t.newVar()
		code := fmt.Sprintf("\t%s := hrp.Decimal(%s)\n", varName, n.Valor)
		return code, varName

	case *parser.ConstanteLiteral:
		varName := t.newVar()
		var val string
		switch n.Valor {
		case "Verdadeiro":
			val = "hrp.Verdadeiro"
		case "Falso":
			val = "hrp.Falso"
		default:
			val = "hrp.Nulo"
		}
		code := fmt.Sprintf("\t%s := %s\n", varName, val)
		return code, varName

	case *parser.OpBinaria:
		var sb strings.Builder
		esqCode, esqVar := t.Transpile(n.Esq)
		dirCode, dirVar := t.Transpile(n.Dir)
		sb.WriteString(esqCode)
		sb.WriteString(dirCode)

		varName := t.newVar()
		var opFunc string
		switch n.Operador {
		case "+":
			opFunc = "hrp.Adiciona"
		case "-":
			opFunc = "hrp.Subtrai"
		case "*":
			opFunc = "hrp.Multiplica"
		case "/":
			opFunc = "hrp.Divide"
		case "==":
			opFunc = "hrp.Igual"
		default:
			opFunc = "hrp.Adiciona" // fallback simples
		}

		sb.WriteString(fmt.Sprintf("\t%s, _ := %s(%s, %s)\n", varName, opFunc, esqVar, dirVar))
		return sb.String(), varName

	case *parser.ChamadaFuncao:
		var sb strings.Builder
		targetCode, targetVar := t.Transpile(n.Identificador)
		sb.WriteString(targetCode)

		var argVars []string
		for _, arg := range n.Argumentos {
			argCode, argVar := t.Transpile(arg)
			sb.WriteString(argCode)
			argVars = append(argVars, argVar)
		}

		varName := t.newVar()
		argsSlice := "hrp.Tupla{" + strings.Join(argVars, ", ") + "}"
		sb.WriteString(fmt.Sprintf("\t%s, _ := hrp.Chamar(%s, %s)\n", varName, targetVar, argsSlice))
		return sb.String(), varName

	case *parser.ImporteDe:
		var sb strings.Builder
		varName := t.newVar()
		valorLimpo := n.Caminho.Valor
		if len(valorLimpo) >= 2 && (strings.HasPrefix(valorLimpo, `"`) || strings.HasPrefix(valorLimpo, `'`)) {
			valorLimpo = valorLimpo[1 : len(valorLimpo)-1]
		}
		sb.WriteString(fmt.Sprintf("\t%s, err_%s := hrp.MaquinarioImporteModulo(ctx, \"%s\", escopo)\n", varName, varName, valorLimpo))
		sb.WriteString(fmt.Sprintf("\tif err_%s != nil { panic(err_%s) }\n", varName, varName))
		for _, nome := range n.Nomes {
			nomeVar := t.newVar()
			sb.WriteString(fmt.Sprintf("\t%s, _ := hrp.ObtemAtributoS(%s, \"%s\")\n", nomeVar, varName, nome))
			sb.WriteString(fmt.Sprintf("\tescopo.DefinirSimbolo(hrp.NewVarSimbolo(\"%s\", %s))\n", nome, nomeVar))
		}
		return sb.String(), "hrp.Nulo"

	case *parser.Bloco:
		var sb strings.Builder
		for _, decl := range n.Declaracoes {
			code, _ := t.Transpile(decl)
			sb.WriteString(code)
		}
		return sb.String(), "hrp.Nulo"

	case *parser.DeclFuncao:
		var sb strings.Builder
		var params []string
		for _, p := range n.Parametros {
			params = append(params, p.Nome)
		}
		body, _ := t.Transpile(n.Corpo)
		sb.WriteString(fmt.Sprintf("\t{\n"))
		sb.WriteString(fmt.Sprintf("\t\tfn_%s := func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {\n", n.Nome))
		for i, p := range n.Parametros {
			sb.WriteString(fmt.Sprintf("\t\t\tescopo.DefinirSimbolo(hrp.NewVarSimbolo(\"%s\", args[%d]))\n", p.Nome, i))
		}
		sb.WriteString(body)
		sb.WriteString(fmt.Sprintf("\t\t\treturn hrp.Nulo, nil\n"))
		sb.WriteString(fmt.Sprintf("\t\t}\n"))
		sb.WriteString(fmt.Sprintf("\t\tescopo.DefinirSimbolo(hrp.NewVarSimbolo(\"%s\", hrp.NewFuncaoNativa(\"%s\", fn_%s, %d)))\n", n.Nome, n.Nome, n.Nome, len(params)))
		sb.WriteString(fmt.Sprintf("\t}\n"))
		return sb.String(), "hrp.Nulo"

	case *parser.ExpressaoSe:
		var sb strings.Builder
		condCode, condVar := t.Transpile(n.Condicao)
		sb.WriteString(condCode)
		ifCode, _ := t.Transpile(n.Corpo)
		var elseCode string
		if n.Alternativa != nil {
			elseCode, _ = t.Transpile(n.Alternativa)
		}
		sb.WriteString(fmt.Sprintf("\tif %s == hrp.Verdadeiro {\n", condVar))
		sb.WriteString(ifCode)
		if elseCode != "" {
			sb.WriteString(fmt.Sprintf("\t} else {\n"))
			sb.WriteString(elseCode)
		}
		sb.WriteString("\t}\n")
		return sb.String(), "hrp.Nulo"

	case *parser.Enquanto:
		var sb strings.Builder
		condCode, condVar := t.Transpile(n.Condicao)
		body, _ := t.Transpile(n.Corpo)
		sb.WriteString(fmt.Sprintf("\tfor {\n"))
		sb.WriteString(condCode)
		sb.WriteString(fmt.Sprintf("\t\tif %s != hrp.Verdadeiro { break }\n", condVar))
		sb.WriteString(body)
		sb.WriteString("\t}\n")
		return sb.String(), "hrp.Nulo"

	case *parser.BlocoPara:
		var sb strings.Builder
		iterCode, iterVar := t.Transpile(n.Iterador)
		sb.WriteString(iterCode)
		body, _ := t.Transpile(n.Corpo)
		loopVar := t.newVar()
		sb.WriteString(fmt.Sprintf("\tif _lista_%s, ok_%s := %s.(hrp.Iteravel); ok_%s {\n", loopVar, loopVar, iterVar, loopVar))
		sb.WriteString(fmt.Sprintf("\t\tfor _i_%s := 0; _i_%s < _lista_%s.Contagem(); _i_%s++ {\n", loopVar, loopVar, loopVar, loopVar))
		sb.WriteString(fmt.Sprintf("\t\t\tescopo.DefinirSimbolo(hrp.NewVarSimbolo(\"%s\", _lista_%s.ObtemItem(hrp.Inteiro(_i_%s))))\n", n.Identificador, loopVar, loopVar))
		sb.WriteString(body)
		sb.WriteString("\t\t}\n\t}\n")
		return sb.String(), "hrp.Nulo"

	case *parser.RetorneNode:
		var sb strings.Builder
		if n.Expressao != nil {
			retCode, retVar := t.Transpile(n.Expressao)
			sb.WriteString(retCode)
			sb.WriteString(fmt.Sprintf("\t\treturn %s, nil\n", retVar))
		} else {
			sb.WriteString("\t\treturn hrp.Nulo, nil\n")
		}
		return sb.String(), "hrp.Nulo"

	case *parser.AcessoMembro:
		var sb strings.Builder
		donoCode, donoVar := t.Transpile(n.Dono)
		membroCode, membroVar := t.Transpile(n.Membro)
		sb.WriteString(donoCode)
		sb.WriteString(membroCode)
		varName := t.newVar()
		sb.WriteString(fmt.Sprintf("\t%s, _ := hrp.ObtemAtributoS(%s, %s.(hrp.Texto))\n", varName, donoVar, membroVar))
		return sb.String(), varName

	case *parser.OpUnaria:
		var sb strings.Builder
		exprCode, exprVar := t.Transpile(n.Expressao)
		sb.WriteString(exprCode)
		varName := t.newVar()
		switch n.Operador {
		case "-":
			sb.WriteString(fmt.Sprintf("\t%s, _ := hrp.Multiplica(hrp.Inteiro(-1), %s)\n", varName, exprVar))
		case "nao":
			sb.WriteString(fmt.Sprintf("\t%s := hrp.Falso\n", varName))
			sb.WriteString(fmt.Sprintf("\tif %s == hrp.Falso { %s = hrp.Verdadeiro }\n", exprVar, varName))
		default:
			sb.WriteString(fmt.Sprintf("\t%s := %s\n", varName, exprVar))
		}
		return sb.String(), varName

	case *parser.OpPipe:
		var sb strings.Builder
		esqCode, esqVar := t.Transpile(n.Esq)
		sb.WriteString(esqCode)
		if call, ok := n.Dir.(*parser.ChamadaFuncao); ok {
			callCode, callVar := t.Transpile(&parser.ChamadaFuncao{
				Identificador: call.Identificador,
				Argumentos:    append([]parser.BaseNode{nil}, call.Argumentos...),
			})
			sb.WriteString(callCode)
			varName := t.newVar()
			sb.WriteString(fmt.Sprintf("\t%s, _ = %s(%s)\n", varName, callVar, esqVar))
			return sb.String(), varName
		}
		dirCode, dirVar := t.Transpile(n.Dir)
		sb.WriteString(dirCode)
		varName := t.newVar()
		sb.WriteString(fmt.Sprintf("\t%s, _ := hrp.Chamar(%s, hrp.Tupla{%s})\n", varName, dirVar, esqVar))
		return sb.String(), varName

	case *parser.ListaLiteral:
		var sb strings.Builder
		varName := t.newVar()
		sb.WriteString(fmt.Sprintf("\t%s := hrp.ListaVazia()\n", varName))
		for _, elem := range n.Elementos {
			eCode, eVar := t.Transpile(elem)
			sb.WriteString(eCode)
			sb.WriteString(fmt.Sprintf("\t%s = %s.Adicionar(%s)\n", varName, varName, eVar))
		}
		return sb.String(), varName

	case *parser.TuplaLiteral:
		var sb strings.Builder
		varName := t.newVar()
		var elems []string
		for _, elem := range n.Elementos {
			eCode, eVar := t.Transpile(elem)
			sb.WriteString(eCode)
			elems = append(elems, eVar)
		}
		sb.WriteString(fmt.Sprintf("\t%s := hrp.Tupla{%s}\n", varName, strings.Join(elems, ", ")))
		return sb.String(), varName

	case *parser.Indexacao:
		var sb strings.Builder
		objCode, objVar := t.Transpile(n.Objeto)
		idxCode, idxVar := t.Transpile(n.Argumento)
		sb.WriteString(objCode)
		sb.WriteString(idxCode)
		varName := t.newVar()
		sb.WriteString(fmt.Sprintf("\t%s, _ := hrp.Indice(%s, %s)\n", varName, objVar, idxVar))
		return sb.String(), varName

	case *parser.PareNode:
		return "\tbreak\n", "hrp.Nulo"

	case *parser.ContinueNode:
		return "\tcontinue\n", "hrp.Nulo"

	case *parser.NovaNode:
		var sb strings.Builder
		ctorCode, ctorVar := t.Transpile(n.Objeto)
		sb.WriteString(ctorCode)
		varName := t.newVar()
		sb.WriteString(fmt.Sprintf("\t%s, _ := hrp.Chamar(%s, hrp.Tupla{})\n", varName, ctorVar))
		return sb.String(), varName

	case *parser.TenteCaptureFinalmente:
		var sb strings.Builder
		tenteCode, _ := t.Transpile(n.TenteBlock)
		captureCode, _ := t.Transpile(n.CaptureBlock)
		finalmenteCode, _ := t.Transpile(n.FinalmenteBlock)
		sb.WriteString(fmt.Sprintf("\tfunc() {\n"))
		sb.WriteString(fmt.Sprintf("\t\tdefer func() {\n"))
		sb.WriteString(fmt.Sprintf("\t\t\tif r := recover(); r != nil {\n"))
		sb.WriteString(captureCode)
		sb.WriteString(fmt.Sprintf("\t\t\t}\n"))
		sb.WriteString(fmt.Sprintf("\t\t}()\n"))
		sb.WriteString(tenteCode)
		sb.WriteString(fmt.Sprintf("\t}()\n"))
		if n.FinalmenteBlock != nil {
			sb.WriteString(finalmenteCode)
		}
		return sb.String(), "hrp.Nulo"

	case *parser.AguardeNode:
		var sb strings.Builder
		exprCode, exprVar := t.Transpile(n.Expressao)
		sb.WriteString(exprCode)
		varName := t.newVar()
		sb.WriteString(fmt.Sprintf("\t%s := %s\n", varName, exprVar))
		return sb.String(), varName

	case *parser.Anotacao:
		return "", "hrp.Nulo"

	case *parser.DeclClasse:
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("\t{\n"))
		sb.WriteString(fmt.Sprintf("\t\ttipo_%s := &hrp.TipoClasse{Nome: \"%s\"}\n", n.Nome, n.Nome))
		for _, metodo := range n.Metodos {
			metCode, _ := t.Transpile(metodo)
			sb.WriteString(metCode)
		}
		sb.WriteString(fmt.Sprintf("\t\tescopo.DefinirSimbolo(hrp.NewVarSimbolo(\"%s\", tipo_%s))\n", n.Nome, n.Nome))
		sb.WriteString(fmt.Sprintf("\t}\n"))
		return sb.String(), "hrp.Nulo"

	case *parser.MapaLiteral:
		var sb strings.Builder
		varName := t.newVar()
		sb.WriteString(fmt.Sprintf("\t%s := hrp.Mapa{}\n", varName))
		for _, par := range n.Entradas {
			chaveCode, chaveVar := t.Transpile(par.Chave)
			valorCode, valorVar := t.Transpile(par.Valor)
			sb.WriteString(chaveCode)
			sb.WriteString(valorCode)
			sb.WriteString(fmt.Sprintf("\t%s.Definir(%s, %s)\n", varName, chaveVar, valorVar))
		}
		return sb.String(), varName

	case *parser.AsseguraNode:
		var sb strings.Builder
		condCode, condVar := t.Transpile(n.Condicao)
		sb.WriteString(condCode)
		if n.Mensagem != nil {
			msgCode, msgVar := t.Transpile(n.Mensagem)
			sb.WriteString(msgCode)
			sb.WriteString(fmt.Sprintf("\tif %s != hrp.Verdadeiro { panic(fmt.Sprintf(\"asserção falhou: %%v\", %s)) }\n", condVar, msgVar))
		} else {
			sb.WriteString(fmt.Sprintf("\tif %s != hrp.Verdadeiro { panic(\"asserção falhou\") }\n", condVar))
		}
		return sb.String(), "hrp.Nulo"

	case *parser.DeclExportar:
		return "", "hrp.Nulo"

	case *parser.TemplateLiteral:
		var sb strings.Builder
		varName := t.newVar()
		sb.WriteString(fmt.Sprintf("\t%s := \"\"\n", varName))
		for _, parte := range n.Partes {
			pCode, pVar := t.Transpile(parte)
			sb.WriteString(pCode)
			sb.WriteString(fmt.Sprintf("\t%s += fmt.Sprintf(\"%%v\", %s)\n", varName, pVar))
		}
		return sb.String(), varName

	case *parser.TemplateExpr:
		return t.Transpile(n.Expressao)
	}

	return "", "hrp.Nulo"
}

func (t *TranspilerNative) GenerateFullCode(ast parser.BaseNode) string {
	body, _ := t.Transpile(ast)

	return fmt.Sprintf(`package main

import (
	"fmt"

	"github.com/mat-dgruber/Harpia/hrp"
	_ "github.com/mat-dgruber/Harpia/stdlib"
)

func main() {
	var _ = fmt.Printf
	ctx := hrp.NewContexto(hrp.OpcsContexto{})
	defer ctx.Terminar()

	escopo := hrp.NewEscopo()

%s
}
`, body)
}
