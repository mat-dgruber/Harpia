package parser_test

import (
	"testing"

	"github.com/mat-dgruber/Harpia/parser"
)

func TestParseEstilo(t *testing.T) {
	codigo := `
	estilo MeuComponente {
		corDeFundo: "azul";
		botao:hover {
			opacidade: 0.8;
		}
	}
	`
	p := parser.NewParserFromString(codigo, "<teste>")
	prog, err := p.Parse()
	if err != nil {
		t.Fatalf("Erro inesperado ao analisar bloco 'estilo': %v", err)
	}

	if len(prog.Declaracoes) != 1 {
		t.Fatalf("Esperava 1 declaração, mas recebi %d", len(prog.Declaracoes))
	}

	decl, ok := prog.Declaracoes[0].(*parser.DeclEstilo)
	if !ok {
		t.Fatalf("Esperava nó *parser.DeclEstilo, mas recebi %T", prog.Declaracoes[0])
	}

	if decl.Nome != "MeuComponente" {
		t.Errorf("Nome esperado: 'MeuComponente', recebido: '%s'", decl.Nome)
	}
}

func TestParseJSX(t *testing.T) {
	codigo := `
	var elemento = <div classe="p-4" ativo>
		<h1> Olá Mundo </h1>
		<se condicao={exibir()}>
			<span> Ativo </span>
		</se>
		<para item em lista={dados}>
			<p> Item: {item} </p>
		</para>
		<img url="foto.png" />
	</div>
	`
	p := parser.NewParserFromString(codigo, "<teste>")
	prog, err := p.Parse()
	if err != nil {
		t.Fatalf("Erro inesperado ao analisar JSX-like: %v", err)
	}

	if len(prog.Declaracoes) != 1 {
		t.Fatalf("Esperava 1 declaração (var), mas recebi %d", len(prog.Declaracoes))
	}

	declVar, ok := prog.Declaracoes[0].(*parser.DeclVar)
	if !ok {
		t.Fatalf("Esperava nó *parser.DeclVar, mas recebi %T", prog.Declaracoes[0])
	}

	jsx, ok := declVar.Inicializador.(*parser.NoJSX)
	if !ok {
		t.Fatalf("Esperava inicializador *parser.NoJSX, mas recebi %T", declVar.Inicializador)
	}

	if jsx.Tag != "div" {
		t.Errorf("Tag esperada: 'div', recebida: '%s'", jsx.Tag)
	}

	if len(jsx.Atributos) != 2 {
		t.Errorf("Esperava 2 atributos, recebido: %d", len(jsx.Atributos))
	}

	if jsx.Atributos[0].Nome != "classe" {
		t.Errorf("Atributo 1 esperado: 'classe', recebido: '%s'", jsx.Atributos[0].Nome)
	}

	if jsx.Atributos[1].Nome != "ativo" {
		t.Errorf("Atributo 2 esperado: 'ativo', recebido: '%s'", jsx.Atributos[1].Nome)
	}
}

func TestParseDestructuring(t *testing.T) {
	codigo := `
	var [a, b] = obterSinal()
	`
	p := parser.NewParserFromString(codigo, "<teste>")
	prog, err := p.Parse()
	if err != nil {
		t.Fatalf("Erro inesperado ao analisar desestruturação: %v", err)
	}

	if len(prog.Declaracoes) != 1 {
		t.Fatalf("Esperava 1 declaração, mas recebi %d", len(prog.Declaracoes))
	}

	decl, ok := prog.Declaracoes[0].(*parser.DeclVarDestructuring)
	if !ok {
		t.Fatalf("Esperava nó *parser.DeclVarDestructuring, mas recebi %T", prog.Declaracoes[0])
	}

	if len(decl.Nomes) != 2 || decl.Nomes[0] != "a" || decl.Nomes[1] != "b" {
		t.Errorf("Nomes de desestruturação inesperados: %v", decl.Nomes)
	}
}

func TestParseCoalescenciaNula(t *testing.T) {
	codigo := `var nome = usuario?.nome ?? "Desconhecido"`
	p := parser.NewParserFromString(codigo, "<teste>")
	prog, err := p.Parse()
	if err != nil {
		t.Fatalf("Erro inesperado ao analisar coalescência nula: %v", err)
	}

	if len(prog.Declaracoes) != 1 {
		t.Fatalf("Esperava 1 declaração, recebi %d", len(prog.Declaracoes))
	}

	declVar, ok := prog.Declaracoes[0].(*parser.DeclVar)
	if !ok {
		t.Fatalf("Esperava *parser.DeclVar, recebi %T", prog.Declaracoes[0])
	}

	opCoalesce, ok := declVar.Inicializador.(*parser.OpCoalescenciaNula)
	if !ok {
		t.Fatalf("Esperava *parser.OpCoalescenciaNula, recebi %T", declVar.Inicializador)
	}

	acessoOpcional, ok := opCoalesce.Esq.(*parser.AcessoMembroOpcional)
	if !ok {
		t.Fatalf("Esperava *parser.AcessoMembroOpcional no membro esquerdo, recebi %T", opCoalesce.Esq)
	}

	if _, ok := opCoalesce.Dir.(*parser.TextoLiteral); !ok {
		t.Fatalf("Esperava *parser.TextoLiteral no membro direito, recebi %T", opCoalesce.Dir)
	}

	_ = acessoOpcional
}

func TestParseEnum(t *testing.T) {
	codigo := `
	enum StatusTarefa {
		Pendente,
		EmProgresso,
		Concluido
	}
	`
	p := parser.NewParserFromString(codigo, "<teste>")
	prog, err := p.Parse()
	if err != nil {
		t.Fatalf("Erro inesperado ao analisar enum: %v", err)
	}

	if len(prog.Declaracoes) != 1 {
		t.Fatalf("Esperava 1 declaração, recebi %d", len(prog.Declaracoes))
	}

	declEnum, ok := prog.Declaracoes[0].(*parser.DeclEnum)
	if !ok {
		t.Fatalf("Esperava *parser.DeclEnum, recebi %T", prog.Declaracoes[0])
	}

	if declEnum.Nome != "StatusTarefa" {
		t.Errorf("Nome esperado 'StatusTarefa', recebido: '%s'", declEnum.Nome)
	}

	if len(declEnum.Valores) != 3 || declEnum.Valores[0] != "Pendente" || declEnum.Valores[2] != "Concluido" {
		t.Errorf("Valores de enum inesperados: %v", declEnum.Valores)
	}
}

func TestParseInterface(t *testing.T) {
	codigo := `
	interface Repositorio {
		func salvar(item)
		func obterPorId(id): Objeto
	}
	`
	p := parser.NewParserFromString(codigo, "<teste>")
	prog, err := p.Parse()
	if err != nil {
		t.Fatalf("Erro inesperado ao analisar interface: %v", err)
	}

	if len(prog.Declaracoes) != 1 {
		t.Fatalf("Esperava 1 declaração, recebi %d", len(prog.Declaracoes))
	}

	declInterface, ok := prog.Declaracoes[0].(*parser.DeclInterface)
	if !ok {
		t.Fatalf("Esperava *parser.DeclInterface, recebi %T", prog.Declaracoes[0])
	}

	if declInterface.Nome != "Repositorio" {
		t.Errorf("Nome esperado 'Repositorio', recebido: '%s'", declInterface.Nome)
	}

	if len(declInterface.Metodos) != 2 || declInterface.Metodos[0].Nome != "salvar" || declInterface.Metodos[1].Nome != "obterPorId" {
		t.Errorf("Métodos de interface inesperados: %v", declInterface.Metodos)
	}
}

func TestNoPosicaoCoordenadas(t *testing.T) {
	p := parser.NewParserFromString("var x = 42", "teste.hrp")
	ast, err := p.Parse()
	if err != nil {
		t.Fatalf("Erro inesperado: %v", err)
	}
	if len(ast.Declaracoes) == 0 {
		t.Fatalf("AST sem instruções")
	}
	decl := ast.Declaracoes[0]
	inicio, fim := decl.Pos()
	if inicio.Linha != 1 || inicio.Coluna != 1 {
		t.Errorf("Posição inicial incorreta: %+v", inicio)
	}
	if fim.Indice <= inicio.Indice {
		t.Errorf("Posição final deve ser maior que inicial: %d vs %d", fim.Indice, inicio.Indice)
	}
}

func TestRecuperacaoErrosSintaticos(t *testing.T) {
	codigo := `
		var a = 1 + ;
		var b = 2;
		se { ;
		var c = 3;
	`
	p := parser.NewParserFromString(codigo, "teste_erros.hrp")
	_, err := p.Parse()
	if err == nil {
		t.Fatalf("Esperava erros de compilação")
	}
	erros := p.Erros()
	if len(erros) < 2 {
		t.Errorf("Esperava pelo menos 2 erros acumulados, obteve %d", len(erros))
	}
}


