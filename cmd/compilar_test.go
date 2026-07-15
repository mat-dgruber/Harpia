package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/natanfeitosa/portuscript/ptst"
)

func TestTranspileWeb(t *testing.T) {
	codigo := `
	var contadorSinal = sinal(5);
	var contador = contadorSinal[0];

	estilo Botao {
		corDeFundo: "red";
	}

	funcao MeuApp() {
		retorne <div classe="App">
			<h1>Olá</h1>
			<se condicao={contador() > 2}>
				<span>Alto</span>
			</se>
		</div>;
	}
	`
	ctx := ptst.NewContexto(ptst.OpcsContexto{})
	defer ctx.Terminar()

	ast, err := ctx.StringParaAst(codigo, "<teste>")
	if err != nil {
		t.Fatalf("Erro ao gerar AST: %v", err)
	}

	transpiler := &TranspilerWeb{}
	jsOutput := transpiler.Transpile(ast)

	// Validações básicas de transpilação
	if !strings.Contains(jsOutput, "let contadorSinal = sinal(5);") {
		t.Errorf("Código JS esperado 'let contadorSinal = sinal(5);' não encontrado")
	}

	if !strings.Contains(jsOutput, "h('div', { classe: \"App\" }") {
		t.Errorf("JSX esperado 'h('div', { classe: \"App\" }' não encontrado. Recebido: %s", jsOutput)
	}

	if !strings.Contains(jsOutput, "((contador() > 2) ? h('span', {}, \"Alto\") : null)") {
		t.Errorf("Expressão condicional se-JSX esperada não encontrada. Recebido: %s", jsOutput)
	}

	// Validações de estilo
	if len(transpiler.Styles) != 1 {
		t.Errorf("Esperava 1 bloco de estilo, mas recebi %d", len(transpiler.Styles))
	}

	if !strings.Contains(transpiler.Styles[0], ".Botao") {
		t.Errorf("Classe de estilo '.Botao' esperada não encontrada")
	}
}

func TestComandoCompilarWeb(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "portuscript_build_*")
	if err != nil {
		t.Fatalf("Erro ao criar diretório temporário: %v", err)
	}
	defer os.RemoveAll(tempDir)

	ptstFile := filepath.Join(tempDir, "main.ptst")
	err = os.WriteFile(ptstFile, []byte(`
	funcao MeuApp() {
		retorne <h1>Olá</h1>;
	}
	`), 0644)
	if err != nil {
		t.Fatalf("Erro ao criar main.ptst temporário: %v", err)
	}

	// Como criar os testes simulando a Cobra CLI seria complexo, podemos testar chamando a função de comando
	cmd := comandoCompilar()
	saidaDir := filepath.Join(tempDir, "dist")
	cmd.SetArgs([]string{"--entrada", ptstFile, "--saida", saidaDir})

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("Erro ao executar comando de compilação: %v", err)
	}

	// Verifica se os arquivos finais de build foram gerados
	indexFile := filepath.Join(saidaDir, "index.html")
	appFile := filepath.Join(saidaDir, "app.js")
	runtimeFile := filepath.Join(saidaDir, "runtime-web.js")

	if _, err := os.Stat(indexFile); os.IsNotExist(err) {
		t.Errorf("Arquivo de index.html não foi gerado")
	}

	if _, err := os.Stat(appFile); os.IsNotExist(err) {
		t.Errorf("Arquivo de app.js não foi gerado")
	}

	if _, err := os.Stat(runtimeFile); os.IsNotExist(err) {
		t.Errorf("Arquivo de runtime-web.js não foi gerado")
	}
}

func TestRouterAndLinkTranspilation(t *testing.T) {
	codigo := `
	funcao Navegacao() {
		retorne <Link para="/contato">Fale conosco</Link>;
	}
	`
	ctx := ptst.NewContexto(ptst.OpcsContexto{})
	defer ctx.Terminar()

	ast, err := ctx.StringParaAst(codigo, "<teste>")
	if err != nil {
		t.Fatalf("Erro ao gerar AST: %v", err)
	}

	transpiler := &TranspilerWeb{}
	jsOutput := transpiler.Transpile(ast)

	esperado := "h('a', { href: \"/contato\", aoClicar: (e) => { e.preventDefault(); navegar(\"/contato\"); } }, \"Fale conosco\")"
	if !strings.Contains(jsOutput, esperado) {
		t.Errorf("Esperava a transpilação do Link contendo:\n%s\nRecebido:\n%s", esperado, jsOutput)
	}
}

func TestComandoCompilarComRotas(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "portuscript_build_rotas_*")
	if err != nil {
		t.Fatalf("Erro ao criar diretório temporário: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Cria estrutura temporária: web/rotas/
	rotasDir := filepath.Join(tempDir, "web", "rotas")
	err = os.MkdirAll(rotasDir, 0755)
	if err != nil {
		t.Fatalf("Erro ao criar pasta de rotas temporárias: %v", err)
	}

	// Cria o arquivo principal
	ptstFile := filepath.Join(tempDir, "main.ptst")
	err = os.WriteFile(ptstFile, []byte(`
	var principal = "app";
	`), 0644)
	if err != nil {
		t.Fatalf("Erro ao criar main.ptst: %v", err)
	}

	// Cria rotas: index.ptst e sobre.ptst
	err = os.WriteFile(filepath.Join(rotasDir, "index.ptst"), []byte(`
	retorne <h1>Início</h1>;
	`), 0644)
	if err != nil {
		t.Fatalf("Erro ao criar rota index: %v", err)
	}

	err = os.WriteFile(filepath.Join(rotasDir, "sobre.ptst"), []byte(`
	retorne <h2>Sobre nós</h2>;
	`), 0644)
	if err != nil {
		t.Fatalf("Erro ao criar rota sobre: %v", err)
	}

	cmd := comandoCompilar()
	saidaDir := filepath.Join(tempDir, "dist")
	cmd.SetArgs([]string{"--entrada", ptstFile, "--saida", saidaDir})

	// Executa a compilação de dentro do tempDir para que o compilador encontre as pastas
	oldWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(oldWd)

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("Erro ao executar compilação com rotas: %v", err)
	}

	// Verifica se gerou o arquivo e lê o app.js
	appFile := filepath.Join(saidaDir, "app.js")
	content, err := os.ReadFile(appFile)
	if err != nil {
		t.Fatalf("Erro ao ler app.js gerado: %v", err)
	}

	jsStr := string(content)

	// Valida se as funções de rota foram geradas
	if !strings.Contains(jsStr, "function Rota_Index()") {
		t.Errorf("Esperava a geração da função 'Rota_Index'")
	}
	if !strings.Contains(jsStr, "function Rota_Sobre()") {
		t.Errorf("Esperava a geração da função 'Rota_Sobre'")
	}

	// Valida se a configuração automática do MeuApp roteador foi gerada
	if !strings.Contains(jsStr, "'/': Rota_Index") {
		t.Errorf("Mapeamento de rota '/' esperado não encontrado")
	}
	if !strings.Contains(jsStr, "'/sobre': Rota_Sobre") {
		t.Errorf("Mapeamento de rota '/sobre' esperado não encontrado")
	}
	if !strings.Contains(jsStr, "export function MeuApp()") {
		t.Errorf("Export do componente root 'MeuApp' reativo esperado não encontrado")
	}
}
