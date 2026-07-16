package vm

import (
	"testing"

	"github.com/natanfeitosa/portuscript/ptst"
	_ "github.com/natanfeitosa/portuscript/stdlib"
)

func TestVMExecucaoExpressoesSimples(t *testing.T) {
	testes := []struct {
		nome     string
		codigo   string
		esperado ptst.Objeto
	}{
		{
			nome:     "Soma simples",
			codigo:   "10 + 20;",
			esperado: ptst.Inteiro(30),
		},
		{
			nome:     "Subtração e Multiplicação",
			codigo:   "(10 - 2) * 3;",
			esperado: ptst.Inteiro(24),
		},
		{
			nome:     "Texto e concatenação",
			codigo:   `"Olá " + "Mundo";`,
			esperado: ptst.Texto("Olá Mundo"),
		},
		{
			nome:     "Comparação igualdade verdadeira",
			codigo:   "10 == 10;",
			esperado: ptst.Verdadeiro,
		},
		{
			nome:     "Comparação menor falsa",
			codigo:   "10 < 5;",
			esperado: ptst.Falso,
		},
		{
			nome: "Declaração e uso de variável",
			codigo: `
			var x = 42;
			x;
			`,
			esperado: ptst.Inteiro(42),
		},
		{
			nome: "Reatribuição de variável",
			codigo: `
			var x = 10;
			x = 20;
			x;
			`,
			esperado: ptst.Inteiro(20),
		},
		{
			nome: "Condicional se (Verdadeiro)",
			codigo: `
			var resultado = 0;
			se (Verdadeiro) {
				resultado = 100;
			} senao {
				resultado = 200;
			}
			resultado;
			`,
			esperado: ptst.Inteiro(100),
		},
		{
			nome: "Condicional se (Falso)",
			codigo: `
			var resultado = 0;
			se (Falso) {
				resultado = 100;
			} senao {
				resultado = 200;
			}
			resultado;
			`,
			esperado: ptst.Inteiro(200),
		},
		{
			nome: "Laço enquanto (loop)",
			codigo: `
			var i = 0;
			var acumulado = 0;
			enquanto (i < 5) {
				acumulado = acumulado + 10;
				i = i + 1;
			}
			acumulado;
			`,
			esperado: ptst.Inteiro(50),
		},
	}

	for _, tc := range testes {
		t.Run(tc.nome, func(t *testing.T) {
			ctx := ptst.NewContexto(ptst.OpcsContexto{})
			defer ctx.Terminar()

			ast, err := ctx.StringParaAst(tc.codigo, "<teste>")
			if err != nil {
				t.Fatalf("Erro ao compilar AST: %v", err)
			}

			comp := NewCompilador()
			prog, err := comp.Compilar(ast)
			if err != nil {
				t.Fatalf("Erro na compilação do bytecode: %v", err)
			}

			mainModulo, errMod := ctx.ObterModulo("__main__")
			var mainEscopo *ptst.Escopo
			if errMod == nil && mainModulo != nil {
				mainEscopo = mainModulo.Escopo
			} else {
				mainEscopo = ptst.NewEscopo()
			}

			virtualMachine := NewVM(ctx)
			frame := NewFrame(prog.Bytecode, prog.Constantes, mainEscopo, nil)
			resultado, errExec := virtualMachine.Executar(frame)

			if errExec != nil {
				t.Fatalf("Erro de execução na VM: %v", errExec)
			}

			igual, errCmp := ptst.Igual(resultado, tc.esperado)
			if errCmp != nil {
				t.Fatalf("Erro ao comparar resultado: %v", errCmp)
			}

			if igual != ptst.Verdadeiro {
				t.Errorf("Esperava %v, obteve %v (tipo esperado: %s, tipo obtido: %s)", tc.esperado, resultado, tc.esperado.Tipo().Nome, resultado.Tipo().Nome)
			}
		})
	}
}

func BenchmarkInterpretadorLoop(b *testing.B) {
	codigo := `
	var i = 0;
	enquanto (i < 1000) {
		i = i + 1;
	}
	`

	ctx := ptst.NewContexto(ptst.OpcsContexto{})
	defer ctx.Terminar()

	ast, err := ctx.StringParaAst(codigo, "<benchmark>")
	if err != nil {
		b.Fatalf("Erro ao compilar AST: %v", err)
	}

	b.ResetTimer()
	for b.Loop() {
		escopo := ptst.NewEscopo()
		_, err = (&ptst.Interpretador{Ast: ast, Contexto: ctx, Escopo: escopo}).Inicializa()
		if err != nil {
			b.Fatalf("Erro de execução: %v", err)
		}
	}
}

func BenchmarkVMLoop(b *testing.B) {
	codigo := `
	var i = 0;
	enquanto (i < 1000) {
		i = i + 1;
	}
	`

	ctx := ptst.NewContexto(ptst.OpcsContexto{})
	defer ctx.Terminar()

	ast, err := ctx.StringParaAst(codigo, "<benchmark>")
	if err != nil {
		b.Fatalf("Erro ao compilar AST: %v", err)
	}

	comp := NewCompilador()
	prog, err := comp.Compilar(ast)
	if err != nil {
		b.Fatalf("Erro ao compilar para bytecode: %v", err)
	}

	b.ResetTimer()
	for b.Loop() {
		escopo := ptst.NewEscopo()
		virtualMachine := NewVM(ctx)
		frame := NewFrame(prog.Bytecode, prog.Constantes, escopo, nil)
		_, err = virtualMachine.Executar(frame)
		if err != nil {
			b.Fatalf("Erro de execução na VM: %v", err)
		}
	}
}

func TestVMGCQuebraDeCiclos(t *testing.T) {
	// Cria duas listas mutáveis
	a := &ptst.Lista{Itens: make([]ptst.Objeto, 0)}
	b := &ptst.Lista{Itens: make([]ptst.Objeto, 0)}

	// Estabelece referências cruzadas circulares: a aponta para b, b aponta para a
	a.Itens = append(a.Itens, b)
	b.Itens = append(b.Itens, a)

	// Incrementa referências de retenção de posse no grafo
	ptst.ReterObjeto(a)
	ptst.ReterObjeto(b)

	// Registra as raízes em um escopo
	escopo := ptst.NewEscopo()
	escopo.DefinirSimbolo(ptst.NewVarSimbolo("lista_a", a))
	escopo.DefinirSimbolo(ptst.NewVarSimbolo("lista_b", b))

	// Como um contém o outro de forma cíclica fechada,
	// se executarmos a coleta de ciclos, o coletor deve quebrar e desalocar os filhos
	ptst.ColetarCiclos(escopo)

	if a.Itens != nil {
		t.Errorf("Esperava que o ciclo circular de 'a' fosse quebrado e Itens fosse limpo (nil)")
	}

	if b.Itens != nil {
		t.Errorf("Esperava que o ciclo circular de 'b' fosse quebrado")
	}
}

func TestVMConcorrenciaEDeclFuncao(t *testing.T) {
	codigo := `
	assincrono funcao pegarValor() {
		retorne 123;
	}

	funcao principal() {
		var x = aguarde pegarValor();
		x;
	}

	principal();
	`

	ctx := ptst.NewContexto(ptst.OpcsContexto{})
	defer ctx.Terminar()

	ast, err := ctx.StringParaAst(codigo, "<teste_vm_async>")
	if err != nil {
		t.Fatalf("Erro ao compilar AST: %v", err)
	}

	comp := NewCompilador()
	prog, err := comp.Compilar(ast)
	if err != nil {
		t.Fatalf("Erro ao compilar bytecode: %v", err)
	}

	escopo := ptst.NewEscopo()
	virtualMachine := NewVM(ctx)
	frame := NewFrame(prog.Bytecode, prog.Constantes, escopo, nil)
	resultado, errExec := virtualMachine.Executar(frame)
	if errExec != nil {
		t.Fatalf("Erro de execução na VM: %v", errExec)
	}

	if resultado != ptst.Inteiro(123) {
		t.Errorf("Esperava retorno 123 da execução assíncrona da VM, obtive: %v", resultado)
	}
}

func TestVMRecursionGuard(t *testing.T) {
	ctx := ptst.NewContexto(ptst.OpcsContexto{})
	defer ctx.Terminar()

	virtualMachine := NewVM(ctx)
	
	// Cria uma pilha encadeada de frames artificiais com profundidade de 1001 para disparar o estouro
	var topo *Frame
	for i := 0; i < 1002; i++ {
		topo = NewFrame([]byte{OP_PUSH_CONST, 0, OP_RETORNE}, []ptst.Objeto{ptst.Inteiro(1)}, ptst.NewEscopo(), topo)
	}

	_, errExec := virtualMachine.Executar(topo)
	if errExec == nil {
		t.Fatal("Esperava estouro de pilha/erro de recursão, mas a execução terminou com sucesso.")
	}

	if erroObj, ok := errExec.(*ptst.Erro); ok {
		if erroObj.Base != ptst.ErroDePilha {
			t.Errorf("Esperava erro do tipo ErroDePilha, obteve: %v", erroObj.Base.Nome)
		}
	} else {
		t.Errorf("Esperava erro do tipo ptst.Erro estruturado, obteve: %T", errExec)
	}
}

