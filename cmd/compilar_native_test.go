package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mat-dgruber/Harpia/ptst"
)

func TestTranspileNative(t *testing.T) {
	codigo := `
	var x = 10
	var y = 20
	var soma = x + y
	`

	ctx := ptst.NewContexto(ptst.OpcsContexto{})
	defer ctx.Terminar()

	ast, err := ctx.StringParaAst(codigo, "<teste>")
	if err != nil {
		t.Fatalf("Erro ao gerar AST: %v", err)
	}

	transpiler := &TranspilerNative{}
	goCode := transpiler.GenerateFullCode(ast)

	if !strings.Contains(goCode, "ptst.Inteiro(10)") {
		t.Errorf("Código Go esperado contendo 'ptst.Inteiro(10)' não encontrado. Recebido: %s", goCode)
	}

	if !strings.Contains(goCode, "ptst.Adiciona(") {
		t.Errorf("Operação de soma esperada 'ptst.Adiciona(' não encontrada. Recebido: %s", goCode)
	}
}

func TestTranspileSe(t *testing.T) {
	codigo := `var x = 10
se (x == 10) {
	var a = 1
} senao {
	var b = 2
}`
	ctx := ptst.NewContexto(ptst.OpcsContexto{})
	defer ctx.Terminar()
	ast, err := ctx.StringParaAst(codigo, "<teste>")
	if err != nil {
		t.Fatalf("Erro ao gerar AST: %v", err)
	}
	goCode := (&TranspilerNative{}).GenerateFullCode(ast)
	if !strings.Contains(goCode, "ptst.Verdadeiro") {
		t.Errorf("Expressão Se não gerou verificação de Verdadeiro. Recebido: %s", goCode)
	}
	if !strings.Contains(goCode, "} else {") {
		t.Errorf("Expressão Se não gerou bloco else. Recebido: %s", goCode)
	}
}

func TestTranspileEnquanto(t *testing.T) {
	codigo := `var i = 0
enquanto (i < 10) {
	i = i + 1
}`
	ctx := ptst.NewContexto(ptst.OpcsContexto{})
	defer ctx.Terminar()
	ast, err := ctx.StringParaAst(codigo, "<teste>")
	if err != nil {
		t.Fatalf("Erro ao gerar AST: %v", err)
	}
	goCode := (&TranspilerNative{}).GenerateFullCode(ast)
	if !strings.Contains(goCode, "for {") {
		t.Errorf("Enquanto não gerou loop for. Recebido: %s", goCode)
	}
	if !strings.Contains(goCode, "break") {
		t.Errorf("Enquanto não gerou break. Recebido: %s", goCode)
	}
}

func TestTranspileFuncao(t *testing.T) {
	codigo := `func somar(a, b) {
	retorne a + b
}
somar(1, 2)
`
	ctx := ptst.NewContexto(ptst.OpcsContexto{})
	defer ctx.Terminar()
	ast, err := ctx.StringParaAst(codigo, "<teste>")
	if err != nil {
		t.Fatalf("Erro ao gerar AST: %v", err)
	}
	goCode := (&TranspilerNative{}).GenerateFullCode(ast)
	if !strings.Contains(goCode, "ptst.NewFuncaoNativa") {
		t.Errorf("DeclFuncao não gerou NewFuncaoNativa. Recebido: %s", goCode)
	}
	if !strings.Contains(goCode, "return") {
		t.Errorf("Retorne não gerou return. Recebido: %s", goCode)
	}
}

func TestTranspileLista(t *testing.T) {
	codigo := `var lista = [1, 2, 3]`
	ctx := ptst.NewContexto(ptst.OpcsContexto{})
	defer ctx.Terminar()
	ast, err := ctx.StringParaAst(codigo, "<teste>")
	if err != nil {
		t.Fatalf("Erro ao gerar AST: %v", err)
	}
	goCode := (&TranspilerNative{}).GenerateFullCode(ast)
	if !strings.Contains(goCode, "ptst.ListaVazia()") {
		t.Errorf("ListaLiteral não gerou ListaVazia(). Recebido: %s", goCode)
	}
	if !strings.Contains(goCode, ".Adicionar(") {
		t.Errorf("ListaLiteral não gerou .Adicionar(). Recebido: %s", goCode)
	}
}

func TestTranspileMapa(t *testing.T) {
	codigo := `var m = { "chave": "valor" }`
	ctx := ptst.NewContexto(ptst.OpcsContexto{})
	defer ctx.Terminar()
	ast, err := ctx.StringParaAst(codigo, "<teste>")
	if err != nil {
		t.Fatalf("Erro ao gerar AST: %v", err)
	}
	goCode := (&TranspilerNative{}).GenerateFullCode(ast)
	if !strings.Contains(goCode, "ptst.Mapa{}") {
		t.Errorf("MapaLiteral não gerou ptst.Mapa{}. Recebido: %s", goCode)
	}
	if !strings.Contains(goCode, ".Definir(") {
		t.Errorf("MapaLiteral não gerou .Definir(). Recebido: %s", goCode)
	}
}

func TestTranspileIndexacao(t *testing.T) {
	codigo := `var x = [1, 2, 3]
var y = x[0]`
	ctx := ptst.NewContexto(ptst.OpcsContexto{})
	defer ctx.Terminar()
	ast, err := ctx.StringParaAst(codigo, "<teste>")
	if err != nil {
		t.Fatalf("Erro ao gerar AST: %v", err)
	}
	goCode := (&TranspilerNative{}).GenerateFullCode(ast)
	if !strings.Contains(goCode, "ptst.Indice(") {
		t.Errorf("Indexacao não gerou ptst.Indice(). Recebido: %s", goCode)
	}
}

func TestTranspilePare(t *testing.T) {
	codigo := `para (x em [1, 2, 3]) {
	pare
}`
	ctx := ptst.NewContexto(ptst.OpcsContexto{})
	defer ctx.Terminar()
	ast, err := ctx.StringParaAst(codigo, "<teste>")
	if err != nil {
		t.Fatalf("Erro ao gerar AST: %v", err)
	}
	goCode := (&TranspilerNative{}).GenerateFullCode(ast)
	if !strings.Contains(goCode, "break") {
		t.Errorf("PareNode não gerou break. Recebido: %s", goCode)
	}
}

func TestTranspileClasse(t *testing.T) {
	codigo := `classe Pessoa {
	func init(nome) {
		this.nome = nome
	}
}`
	ctx := ptst.NewContexto(ptst.OpcsContexto{})
	defer ctx.Terminar()
	ast, err := ctx.StringParaAst(codigo, "<teste>")
	if err != nil {
		t.Fatalf("Erro ao gerar AST: %v", err)
	}
	goCode := (&TranspilerNative{}).GenerateFullCode(ast)
	if !strings.Contains(goCode, "TipoClasse") {
		t.Errorf("DeclClasse não gerou TipoClasse. Recebido: %s", goCode)
	}
}

func TestTranspileOpUnaria(t *testing.T) {
	codigo := `var x = 5
var y = -x`
	ctx := ptst.NewContexto(ptst.OpcsContexto{})
	defer ctx.Terminar()
	ast, err := ctx.StringParaAst(codigo, "<teste>")
	if err != nil {
		t.Fatalf("Erro ao gerar AST: %v", err)
	}
	goCode := (&TranspilerNative{}).GenerateFullCode(ast)
	if !strings.Contains(goCode, "ptst.Multiplica(ptst.Inteiro(-1),") {
		t.Errorf("OpUnaria negativa não gerou Multiplica(-1, x). Recebido: %s", goCode)
	}
}

func TestTranspileTente(t *testing.T) {
	codigo := `tente {
	var x = 1
} capture(erro) {
	var y = 2
} finalmente {
	var z = 3
}`
	ctx := ptst.NewContexto(ptst.OpcsContexto{})
	defer ctx.Terminar()
	ast, err := ctx.StringParaAst(codigo, "<teste>")
	if err != nil {
		t.Fatalf("Erro ao gerar AST: %v", err)
	}
	goCode := (&TranspilerNative{}).GenerateFullCode(ast)
	if !strings.Contains(goCode, "defer func()") {
		t.Errorf("TenteCaptureFinalmente não gerou defer. Recebido: %s", goCode)
	}
	if !strings.Contains(goCode, "recover()") {
		t.Errorf("TenteCaptureFinalmente não gerou recover(). Recebido: %s", goCode)
	}
}

func TestComandoCompilarNativo(t *testing.T) {
	tempDir := t.TempDir()
	cur, _ := os.Getwd()
	
	// Vamos criar um script simples
	scriptPath := filepath.Join(tempDir, "app.hrp")
	err := os.WriteFile(scriptPath, []byte("var a = 40\nvar b = 2\nvar c = a + b\n"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	saidaBin := filepath.Join(tempDir, "app_bin")

	// Prepara a execução do comando compilar nativo
	cmd := comandoCompilar()
	cmd.SetArgs([]string{
		scriptPath,
		"--alvo=nativo",
		"--saida=" + saidaBin,
	})

	// Muda para a pasta raiz temporariamente se necessário para usar go.mod correto
	// Mas como 'go build' precisa achar github.com/mat-dgruber/Harpia, o Cwd deve ser o diretório atual do repo.
	// Vamos rodar com o Cwd atual do projeto.
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tempDir) // vai para a pasta do tempDir

	projectRoot := filepath.Dir(cur)

	// Escreve um go.mod mínimo para a sandbox de teste para resolver a dependência local
	goModConteudo := fmt.Sprintf(`module test_aot

go 1.24.2

require (
	github.com/mat-dgruber/Harpia v0.0.0
)

replace github.com/mat-dgruber/Harpia => %s
`, projectRoot)

	os.WriteFile("go.mod", []byte(goModConteudo), 0644)

	// Copia go.sum para o compilador resolver somas criptográficas dos drivers
	goSumOrigem := filepath.Join(projectRoot, "go.sum")
	if dataSum, errSum := os.ReadFile(goSumOrigem); errSum == nil {
		os.WriteFile("go.sum", dataSum, 0644)
	}

	if errCmd := cmd.Execute(); errCmd != nil {
		t.Fatalf("Erro ao executar compilar nativo: %v", errCmd)
	}

	// Verifica se o binário foi gerado
	if _, errStat := os.Stat(saidaBin); os.IsNotExist(errStat) {
		t.Fatalf("O binário nativo não foi gerado no caminho esperado: %s", saidaBin)
	}

	// Executa o binário gerado e garante que roda sem pânicos!
	runCmd := exec.Command(saidaBin)
	if errRun := runCmd.Run(); errRun != nil {
		t.Errorf("Erro ao executar o binário AOT gerado: %v", errRun)
	}
}

func TestComandoCompilarWasm(t *testing.T) {
	tempDir := t.TempDir()
	cur, _ := os.Getwd()

	scriptPath := filepath.Join(tempDir, "app.hrp")
	err := os.WriteFile(scriptPath, []byte("var a = 40\nvar b = 2\nvar c = a + b\n"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	saidaWasm := filepath.Join(tempDir, "app.wasm")

	cmd := comandoCompilar()
	cmd.SetArgs([]string{
		scriptPath,
		"--alvo=wasm",
		"--saida=" + saidaWasm,
	})

	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tempDir)

	projectRoot := filepath.Dir(cur)

	goModConteudo := fmt.Sprintf(`module test_aot
go 1.24.2
require github.com/mat-dgruber/Harpia v0.0.0
replace github.com/mat-dgruber/Harpia => %s
`, projectRoot)

	os.WriteFile("go.mod", []byte(goModConteudo), 0644)

	goSumOrigem := filepath.Join(projectRoot, "go.sum")
	if dataSum, errSum := os.ReadFile(goSumOrigem); errSum == nil {
		os.WriteFile("go.sum", dataSum, 0644)
	}

	if errCmd := cmd.Execute(); errCmd != nil {
		t.Fatalf("Erro ao executar compilar wasm: %v", errCmd)
	}

	if _, errStat := os.Stat(saidaWasm); os.IsNotExist(errStat) {
		t.Fatalf("O arquivo .wasm não foi gerado no caminho esperado: %s", saidaWasm)
	}
}
