package cmd

import (
	"github.com/mat-dgruber/Harpia/parser"
)

type Otimizador struct {
	usados map[string]bool
}

func (o *Otimizador) ColetarUsos(node parser.BaseNode) {
	if node == nil {
		return
	}

	switch n := node.(type) {
	case *parser.Programa:
		for _, decl := range n.Declaracoes {
			o.ColetarUsos(decl)
		}

	case *parser.Identificador:
		o.usados[n.Nome] = true

	case *parser.DeclVar:
		o.ColetarUsos(n.Inicializador)

	case *parser.Reatribuicao:
		o.ColetarUsos(n.Objeto)
		o.ColetarUsos(n.Expressao)

	case *parser.OpBinaria:
		o.ColetarUsos(n.Esq)
		o.ColetarUsos(n.Dir)

	case *parser.OpUnaria:
		o.ColetarUsos(n.Expressao)

	case *parser.ChamadaFuncao:
		o.ColetarUsos(n.Identificador)
		for _, arg := range n.Argumentos {
			o.ColetarUsos(arg)
		}

	case *parser.Bloco:
		for _, decl := range n.Declaracoes {
			o.ColetarUsos(decl)
		}

	case *parser.DeclFuncao:
		o.ColetarUsos(n.Corpo)

	case *parser.ExpressaoSe:
		o.ColetarUsos(n.Condicao)
		o.ColetarUsos(n.Corpo)
		o.ColetarUsos(n.Alternativa)

	case *parser.Enquanto:
		o.ColetarUsos(n.Condicao)
		o.ColetarUsos(n.Corpo)

	case *parser.BlocoPara:
		o.ColetarUsos(n.Iterador)
		o.ColetarUsos(n.Corpo)

	case *parser.RetorneNode:
		o.ColetarUsos(n.Expressao)

	case *parser.ListaLiteral:
		for _, elem := range n.Elementos {
			o.ColetarUsos(elem)
		}

	case *parser.MapaLiteral:
		for _, par := range n.Entradas {
			o.ColetarUsos(par.Chave)
			o.ColetarUsos(par.Valor)
		}
	}
}

func (o *Otimizador) OtimizarNo(node parser.BaseNode) parser.BaseNode {
	if node == nil {
		return nil
	}

	switch n := node.(type) {
	case *parser.Programa:
		var novas []parser.BaseNode
		for _, decl := range n.Declaracoes {
			opt := o.OtimizarNo(decl)
			if opt != nil {
				if blk, ok := opt.(*parser.Bloco); ok {
					novas = append(novas, blk.Declaracoes...)
				} else {
					novas = append(novas, opt)
				}
			}
		}
		n.Declaracoes = novas
		return n

	case *parser.DeclVar:
		// Se a variável é 'usada' ou 'soma' (qualquer uma que esteja no mapa de usados), mantemos
		if !o.usados[n.Nome] {
			// ponytail: remove declarações de variáveis mortas / não referenciadas (DCE)
			return nil
		}
		n.Inicializador = o.OtimizarNo(n.Inicializador)
		return n

	case *parser.ExpressaoSe:
		n.Condicao = o.OtimizarNo(n.Condicao)
		if opt := o.OtimizarNo(n.Corpo); opt != nil {
			n.Corpo = opt.(*parser.Bloco)
		} else {
			n.Corpo = nil
		}
		n.Alternativa = o.OtimizarNo(n.Alternativa)

		// No Harpia, as constantes Verdadeiro e Falso são representadas pelo tipo Identificador
		if cond, ok := n.Condicao.(*parser.Identificador); ok {
			if cond.Nome == "Falso" {
				return n.Alternativa
			}
			if cond.Nome == "Verdadeiro" {
				return n.Corpo
			}
		}
		return n

	case *parser.Enquanto:
		n.Condicao = o.OtimizarNo(n.Condicao)
		if opt := o.OtimizarNo(n.Corpo); opt != nil {
			n.Corpo = opt.(*parser.Bloco)
		} else {
			n.Corpo = nil
		}

		if cond, ok := n.Condicao.(*parser.Identificador); ok {
			if cond.Nome == "Falso" {
				return nil
			}
		}
		return n

	case *parser.Bloco:
		var novas []parser.BaseNode
		for _, decl := range n.Declaracoes {
			opt := o.OtimizarNo(decl)
			if opt != nil {
				novas = append(novas, opt)
			}
		}
		n.Declaracoes = novas
		return n

	case *parser.DeclFuncao:
		if opt := o.OtimizarNo(n.Corpo); opt != nil {
			n.Corpo = opt.(*parser.Bloco)
		} else {
			n.Corpo = nil
		}
		return n
	}

	return node
}

func Otimizar(ast *parser.Programa) *parser.Programa {
	o := &Otimizador{usados: make(map[string]bool)}

	// Passo 1: Mapear todos os identificadores em uso
	o.ColetarUsos(ast)

	// Passo 2: Executar eliminação de código morto (DCE) e ramos constantes
	otimizado := o.OtimizarNo(ast)
	if otimizado == nil {
		return ast
	}
	return otimizado.(*parser.Programa)
}
