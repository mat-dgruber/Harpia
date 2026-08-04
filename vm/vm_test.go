package vm

import (
	"testing"

	"github.com/mat-dgruber/Harpia/hrp"
	"github.com/mat-dgruber/Harpia/parser"
	_ "github.com/mat-dgruber/Harpia/stdlib"
)

func TestCompiladorV2(t *testing.T) {
	codigo := `
		var x = 10;
		var y = 20;
		var z = x + y;
		z;
	`
	p := parser.NewParserFromString(codigo, "test.hrp")
	ast, err := p.Parse()
	if err != nil {
		t.Fatalf("Erro ao analisar AST: %v", err)
	}

	comp := NewCompilador()
	prog, err := comp.Compilar(ast)
	if err != nil {
		t.Fatalf("Erro ao compilar AST: %v", err)
	}

	ctx := hrp.NewContexto(hrp.OpcsContexto{})
	defer ctx.Terminar()

	virtualMachine := NewVM(ctx)
	frame := NewFrame(prog.Bytecode, prog.Constantes, hrp.NewEscopo(), nil)
	res, err := virtualMachine.Executar(frame)
	if err != nil {
		t.Fatalf("Erro ao executar VM: %v", err)
	}

	if res != hrp.Inteiro(30) {
		t.Errorf("Esperava 30, obteve: %v", res)
	}
}

func TestVMSwitchLoop(t *testing.T) {
	ctx := hrp.NewContexto(hrp.OpcsContexto{})
	defer ctx.Terminar()

	// Programa manual em bytecode 16-bit:
	// 1. PUSH_CONST (100) -> const_idx 0 (u16: 0x00, 0x00)
	// 2. ARMAZENAR_LOCAL -> slot_idx 0 (u16: 0x00, 0x00)
	// 3. CARREGAR_LOCAL -> slot_idx 0 (u16: 0x00, 0x00)
	// 4. RETORNE
	bytecode := []byte{
		OP_PUSH_CONST, 0x00, 0x00,
		OP_ARMAZENAR_LOCAL, 0x00, 0x00,
		OP_CARREGAR_LOCAL, 0x00, 0x00,
		OP_RETORNE,
	}

	consts := []hrp.Objeto{hrp.Inteiro(100)}
	frame := NewFrame(bytecode, consts, hrp.NewEscopo(), nil)
	frame.Locais = make([]hrp.Objeto, 1)

	virtualMachine := NewVM(ctx)
	res, err := virtualMachine.Executar(frame)

	if err != nil {
		t.Fatalf("Erro ao executar VM: %v", err)
	}

	if res != hrp.Inteiro(100) {
		t.Errorf("Esperava 100, obteve: %v", res)
	}
}

func TestOpcodesV2(t *testing.T) {
	opcodes := []Opcode{
		OP_PUSH_CONST, OP_POP, OP_DUP, OP_ADD, OP_SUB, OP_MUL, OP_DIV, OP_DIV_INT, OP_MOD,
		OP_EQ, OP_NEQ, OP_LT, OP_LTE, OP_GT, OP_GTE, OP_JMP, OP_JMP_FALSO,
		OP_CARREGAR_LOCAL, OP_ARMAZENAR_LOCAL, OP_CARREGAR_GLOBAL, OP_ARMAZENAR_GLOBAL,
		OP_CHAMAR, OP_RETORNE, OP_AWAIT, OP_CRIAR_FUNCAO, OP_RETORNE_CONST, OP_RETORNE_VAR,
		OP_NEG, OP_NOT, OP_CRIAR_LISTA, OP_CRIAR_MAPA, OP_OBTER_INDICE, OP_DEFINIR_INDICE,
		OP_OBTER_ATRIBUTO, OP_DEFINIR_ATRIBUTO, OP_COALESCE, OP_ACESSAR_OPCIONAL,
		OP_SETUP_TRY, OP_POP_TRY, OP_LANCAR,
	}

	for _, op := range opcodes {
		nome := NomeOpcode(op)
		if nome == "OP_DESCONHECIDO" {
			t.Errorf("Opcode 0x%X retornou NomeOpcode OP_DESCONHECIDO", op)
		}
	}
}

func TestVMExecucaoExpressoesSimples(t *testing.T) {
	testes := []struct {
		nome     string
		codigo   string
		esperado hrp.Objeto
	}{
		{
			nome:     "Soma simples",
			codigo:   "10 + 20;",
			esperado: hrp.Inteiro(30),
		},
		{
			nome:     "Subtração e Multiplicação",
			codigo:   "(10 - 2) * 3;",
			esperado: hrp.Inteiro(24),
		},
		{
			nome:     "Texto e concatenação",
			codigo:   `"Olá " + "Mundo";`,
			esperado: hrp.Texto("Olá Mundo"),
		},
		{
			nome:     "Comparação igualdade verdadeira",
			codigo:   "10 == 10;",
			esperado: hrp.Verdadeiro,
		},
		{
			nome:     "Comparação menor falsa",
			codigo:   "10 < 5;",
			esperado: hrp.Falso,
		},
		{
			nome: "Declaração e uso de variável",
			codigo: `
			var x = 42;
			x;
			`,
			esperado: hrp.Inteiro(42),
		},
		{
			nome: "Reatribuição de variável",
			codigo: `
			var x = 10;
			x = 20;
			x;
			`,
			esperado: hrp.Inteiro(20),
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
			esperado: hrp.Inteiro(100),
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
			esperado: hrp.Inteiro(200),
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
			esperado: hrp.Inteiro(50),
		},
		{
			nome: "Operador lógico E (curto-circuito verdadeiro)",
			codigo: `
			Verdadeiro e 42;
			`,
			esperado: hrp.Inteiro(42),
		},
		{
			nome: "Operador lógico E (curto-circuito falso)",
			codigo: `
			Falso e 42;
			`,
			esperado: hrp.Falso,
		},
		{
			nome: "Operador lógico OU (curto-circuito verdadeiro)",
			codigo: `
			Verdadeiro ou 42;
			`,
			esperado: hrp.Verdadeiro,
		},
		{
			nome: "Operador lógico OU (curto-circuito falso)",
			codigo: `
			Falso ou 42;
			`,
			esperado: hrp.Inteiro(42),
		},
	}

	for _, tc := range testes {
		t.Run(tc.nome, func(t *testing.T) {
			ctx := hrp.NewContexto(hrp.OpcsContexto{})
			defer ctx.Terminar()

			ast, err := ctx.StringParaAst(tc.codigo, "<teste>")
			if err != nil {
				t.Fatalf("Erro ao compilar AST: %v", err)
			}

			comp := NewCompilador(ctx)
			prog, err := comp.Compilar(ast)
			if err != nil {
				t.Fatalf("Erro na compilação do bytecode: %v", err)
			}

			mainModulo, errMod := ctx.ObterModulo("__main__")
			var mainEscopo *hrp.Escopo
			if errMod == nil && mainModulo != nil {
				mainEscopo = mainModulo.Escopo
			} else {
				mainEscopo = hrp.NewEscopo()
			}

			virtualMachine := NewVM(ctx)
			frame := NewFrame(prog.Bytecode, prog.Constantes, mainEscopo, nil)
			resultado, errExec := virtualMachine.Executar(frame)

			if errExec != nil {
				t.Fatalf("Erro de execução na VM: %v", errExec)
			}

			igual, errCmp := hrp.Igual(resultado, tc.esperado)
			if errCmp != nil {
				t.Fatalf("Erro ao comparar resultado: %v", errCmp)
			}

			if igual != hrp.Verdadeiro {
				t.Errorf("Esperava %v, obteve %v (tipo esperado: %s, tipo obtido: %s)", tc.esperado, resultado, tc.esperado.Tipo().Nome, resultado.Tipo().Nome)
			}
		})
	}
}

func TestVMRecursosAvancados(t *testing.T) {
	testes := []struct {
		nome     string
		codigo   string
		esperado hrp.Objeto
	}{
		{
			nome:     "Lista literal e índice",
			codigo:   "var lista = [10, 20, 30]; lista[1];",
			esperado: hrp.Inteiro(20),
		},
		{
			nome:     "Mapa literal e índice",
			codigo:   `var mapa = {"chave": 100}; mapa["chave"];`,
			esperado: hrp.Inteiro(100),
		},
		{
			nome:     "Coalescência nula com Nulo",
			codigo:   "var a = Nulo; a ?? 50;",
			esperado: hrp.Inteiro(50),
		},
		{
			nome:     "Coalescência nula com valor",
			codigo:   "var a = 10; a ?? 50;",
			esperado: hrp.Inteiro(10),
		},
		{
			nome:     "Acesso opcional em Nulo",
			codigo:   "var obj = Nulo; obj?.nome;",
			esperado: hrp.Nulo,
		},
		{
			nome:     "Reatribuição composta",
			codigo:   "var x = 10; x += 5; x *= 2; x;",
			esperado: hrp.Inteiro(30),
		},
		{
			nome:     "Reatribuição em índice de lista",
			codigo:   "var lista = [10, 20, 30]; lista[1] = 99; lista[1];",
			esperado: hrp.Inteiro(99),
		},
		{
			nome:     "Laço para-em e soma",
			codigo:   "var soma = 0; para x em [10, 20, 30] { soma += x; } soma;",
			esperado: hrp.Inteiro(60),
		},
		{
			nome:     "Instrução pare em laço enquanto",
			codigo:   "var x = 0; enquanto x < 10 { x += 1; se x == 5 { pare; } } x;",
			esperado: hrp.Inteiro(5),
		},
	}

	for _, tc := range testes {
		t.Run(tc.nome, func(t *testing.T) {
			ctx := hrp.NewContexto(hrp.OpcsContexto{})
			defer ctx.Terminar()

			ast, err := ctx.StringParaAst(tc.codigo, "<teste_avancado>")
			if err != nil {
				t.Fatalf("Erro ao compilar AST: %v", err)
			}

			comp := NewCompilador(ctx)
			prog, err := comp.Compilar(ast)
			if err != nil {
				t.Fatalf("Erro na compilação do bytecode: %v", err)
			}

			virtualMachine := NewVM(ctx)
			frame := NewFrame(prog.Bytecode, prog.Constantes, hrp.NewEscopo(), nil)
			resultado, errExec := virtualMachine.Executar(frame)

			if errExec != nil {
				t.Fatalf("Erro de execução na VM: %v", errExec)
			}

			igual, errCmp := hrp.Igual(resultado, tc.esperado)
			if errCmp != nil {
				t.Fatalf("Erro ao comparar resultado: %v", errCmp)
			}

			if igual != hrp.Verdadeiro {
				t.Errorf("Esperava %v, obteve %v", tc.esperado, resultado)
			}
		})
	}
}

func TestVMExcecoes(t *testing.T) {
	codigo := `
	var resultado = 0;
	tente {
		resultado = 10;
		10 / 0;
		resultado = 20;
	} capture (erro) {
		resultado = 30;
	}
	resultado;
	`

	ctx := hrp.NewContexto(hrp.OpcsContexto{})
	defer ctx.Terminar()

	ast, err := ctx.StringParaAst(codigo, "<teste_excecao>")
	if err != nil {
		t.Fatalf("Erro ao compilar AST: %v", err)
	}

	comp := NewCompilador(ctx)
	prog, err := comp.Compilar(ast)
	if err != nil {
		t.Fatalf("Erro na compilação do bytecode: %v", err)
	}

	virtualMachine := NewVM(ctx)
	frame := NewFrame(prog.Bytecode, prog.Constantes, hrp.NewEscopo(), nil)
	resultado, errExec := virtualMachine.Executar(frame)

	if errExec != nil {
		t.Fatalf("Erro de execução na VM: %v", errExec)
	}

	if resultado != hrp.Inteiro(30) {
		t.Errorf("Esperava resultado 30 após captura de exceção, obtive: %v", resultado)
	}
}

func BenchmarkInterpretadorLoop(b *testing.B) {
	codigo := `
	var i = 0;
	enquanto (i < 1000) {
		i = i + 1;
	}
	`

	ctx := hrp.NewContexto(hrp.OpcsContexto{})
	defer ctx.Terminar()

	ast, err := ctx.StringParaAst(codigo, "<benchmark>")
	if err != nil {
		b.Fatalf("Erro ao compilar AST: %v", err)
	}

	b.ResetTimer()
	for b.Loop() {
		escopo := hrp.NewEscopo()
		_, err = (&hrp.Interpretador{Ast: ast, Contexto: ctx, Escopo: escopo}).Inicializa()
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

	ctx := hrp.NewContexto(hrp.OpcsContexto{})
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
		escopo := hrp.NewEscopo()
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
	a := &hrp.Lista{Itens: make([]hrp.Objeto, 0)}
	b := &hrp.Lista{Itens: make([]hrp.Objeto, 0)}

	// Estabelece referências cruzadas circulares: a aponta para b, b aponta para a
	a.Itens = append(a.Itens, b)
	b.Itens = append(b.Itens, a)

	// Incrementa referências de retenção de posse no grafo
	hrp.ReterObjeto(a)
	hrp.ReterObjeto(b)

	// Registra as raízes em um escopo
	escopo := hrp.NewEscopo()
	escopo.DefinirSimbolo(hrp.NewVarSimbolo("lista_a", a))
	escopo.DefinirSimbolo(hrp.NewVarSimbolo("lista_b", b))

	// Como um contém o outro de forma cíclica fechada,
	// se executarmos a coleta de ciclos, o coletor deve quebrar e desalocar os filhos
	hrp.ColetarCiclos(escopo)

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

	ctx := hrp.NewContexto(hrp.OpcsContexto{})
	defer ctx.Terminar()

	ast, err := ctx.StringParaAst(codigo, "<teste_vm_async>")
	if err != nil {
		t.Fatalf("Erro ao compilar AST: %v", err)
	}

	comp := NewCompilador(ctx)
	prog, err := comp.Compilar(ast)
	if err != nil {
		t.Fatalf("Erro ao compilar bytecode: %v", err)
	}

	escopo := hrp.NewEscopo()
	virtualMachine := NewVM(ctx)
	frame := NewFrame(prog.Bytecode, prog.Constantes, escopo, nil)
	resultado, errExec := virtualMachine.Executar(frame)
	if errExec != nil {
		t.Fatalf("Erro de execução na VM: %v", errExec)
	}

	if resultado != hrp.Inteiro(123) {
		t.Errorf("Esperava retorno 123 da execução assíncrona da VM, obtive: %v", resultado)
	}
}

func TestVMRecursionGuard(t *testing.T) {
	ctx := hrp.NewContexto(hrp.OpcsContexto{})
	defer ctx.Terminar()

	virtualMachine := NewVM(ctx)

	// Cria uma pilha encadeada de frames artificiais com profundidade de 1001 para disparar o estouro
	var topo *Frame
	for i := 0; i < 1002; i++ {
		topo = NewFrame([]byte{OP_PUSH_CONST, 0, OP_RETORNE}, []hrp.Objeto{hrp.Inteiro(1)}, hrp.NewEscopo(), topo)
	}

	_, errExec := virtualMachine.Executar(topo)
	if errExec == nil {
		t.Fatal("Esperava estouro de pilha/erro de recursão, mas a execução terminou com sucesso.")
	}

	if erroObj, ok := errExec.(*hrp.Erro); ok {
		if erroObj.Base != hrp.ErroDePilha {
			t.Errorf("Esperava erro do tipo ErroDePilha, obteve: %v", erroObj.Base.Nome)
		}
	} else {
		t.Errorf("Esperava erro do tipo hrp.Erro estruturado, obteve: %T", errExec)
	}
}
