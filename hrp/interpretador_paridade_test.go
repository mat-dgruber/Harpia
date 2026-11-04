package hrp

import (
	"testing"

	"github.com/mat-dgruber/Harpia/parser"
)

func executarAST(t *testing.T, codigo string) *Escopo {
	ctx := NewContexto(OpcsContexto{})
	defer ctx.Terminar()

	ast, err := ctx.StringParaAst(codigo, "<teste>")
	if err != nil {
		t.Fatalf("Erro ao parsear código: %v", err)
	}

	escopo := NewEscopo()
	interpret := &Interpretador{Ast: ast, Contexto: ctx, Escopo: escopo}
	if prog, ok := ast.(*parser.Programa); ok {
		interpret.Arquivo = prog.Arquivo
		interpret.Codigo = prog.Codigo
		interpret.Posicoes = prog.Posicoes
	}
	_, err = interpret.Inicializa()
	if err != nil {
		t.Fatalf("Erro ao interpretar AST: %v", err)
	}
	return escopo
}

func TestInterpretadorParidade(t *testing.T) {
	t.Run("CORE-02 retorne dentro de enquanto e para interrompe laco", func(t *testing.T) {
		codigo := `
		funcao testarEnquanto() {
			var i = 0
			enquanto i < 10 {
				se i == 3 {
					retorne i
				}
				i = i + 1
			}
			retorne -1
		}

		funcao testarPara() {
			para x em [10, 20, 30, 40] {
				se x == 30 {
					retorne x
				}
			}
			retorne -1
		}

		var resEnquanto = testarEnquanto()
		var resPara = testarPara()
		`
		escopo := executarAST(t, codigo)
		vEnquanto, err := escopo.ObterValor("resEnquanto")
		if err != nil {
			t.Fatalf("resEnquanto não encontrado: %v", err)
		}
		if vEnquanto != Inteiro(3) {
			t.Errorf("Esperava resEnquanto=3, obteve %v", vEnquanto)
		}

		vPara, err := escopo.ObterValor("resPara")
		if err != nil {
			t.Fatalf("resPara não encontrado: %v", err)
		}
		if vPara != Inteiro(30) {
			t.Errorf("Esperava resPara=30, obteve %v", vPara)
		}
	})

	t.Run("CORE-15 escolha caso padrao", func(t *testing.T) {
		codigo := `
		funcao avaliarOpcao(x) {
			escolha x {
				caso 1 {
					retorne "um"
				}
				caso 2 {
					retorne "dois"
				}
				caso padrao {
					retorne "outro"
				}
			}
		}

		var r1 = avaliarOpcao(1)
		var r2 = avaliarOpcao(2)
		var r3 = avaliarOpcao(99)
		`
		escopo := executarAST(t, codigo)
		v1, _ := escopo.ObterValor("r1")
		v2, _ := escopo.ObterValor("r2")
		v3, _ := escopo.ObterValor("r3")

		if v1 != Texto("um") {
			t.Errorf("Esperava r1='um', obteve %v", v1)
		}
		if v2 != Texto("dois") {
			t.Errorf("Esperava r2='dois', obteve %v", v2)
		}
		if v3 != Texto("outro") {
			t.Errorf("Esperava r3='outro', obteve %v", v3)
		}
	})

	t.Run("CORE-16 encadeamento opcional sobre Nulo", func(t *testing.T) {
		codigo := `
		var obj = Nulo
		var resProp = obj?.nome
		var resIndex = obj?.[0]
		var resChamada = obj?.()

		var lista = [100, 200]
		var vLista = lista?.[1]

		funcao dobro(n) { retorne n * 2 }
		var vFn = dobro?.(21)
		`
		escopo := executarAST(t, codigo)
		rp, _ := escopo.ObterValor("resProp")
		ri, _ := escopo.ObterValor("resIndex")
		rc, _ := escopo.ObterValor("resChamada")
		vl, _ := escopo.ObterValor("vLista")
		vf, _ := escopo.ObterValor("vFn")

		if rp != Nulo {
			t.Errorf("Esperava resProp=Nulo, obteve %v", rp)
		}
		if ri != Nulo {
			t.Errorf("Esperava resIndex=Nulo, obteve %v", ri)
		}
		if rc != Nulo {
			t.Errorf("Esperava resChamada=Nulo, obteve %v", rc)
		}
		if vl != Inteiro(200) {
			t.Errorf("Esperava vLista=200, obteve %v", vl)
		}
		if vf != Inteiro(42) {
			t.Errorf("Esperava vFn=42, obteve %v", vf)
		}
	})

	t.Run("R3-VM-02 avaliacao unica em atribuicao composta", func(t *testing.T) {
		codigo := `
		var chamadas = 0
		funcao obterIdx() {
			chamadas = chamadas + 1
			retorne 0
		}

		var lista = [10]
		lista[obterIdx()] += 5
		var resFinal = lista[0]
		`
		escopo := executarAST(t, codigo)
		qtd, _ := escopo.ObterValor("chamadas")
		res, _ := escopo.ObterValor("resFinal")

		if qtd != Inteiro(1) {
			t.Errorf("Esperava chamadas=1, obteve %v", qtd)
		}
		if res != Inteiro(15) {
			t.Errorf("Esperava resFinal=15, obteve %v", res)
		}
	})
}
