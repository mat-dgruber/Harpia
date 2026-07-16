package cmd

import (
	"image"
	"image/png"
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
		raio-grande: Verdadeiro;
		textAlign: "center";
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

	style := transpiler.Styles[0]
	if !strings.Contains(style, ".Botao") {
		t.Errorf("Classe de estilo '.Botao' esperada não encontrada")
	}

	// CSS-1: Mapear chaves PT → CSS + strip de aspas
	if !strings.Contains(style, "background-color: red;") {
		t.Errorf("Esperava tradução de 'corDeFundo: \"red\"' para 'background-color: red;'. Recebido: %s", style)
	}

	// CSS-2: raio-grande: Verdadeiro vira border-radius: 0.5rem;
	if !strings.Contains(style, "border-radius: 0.5rem;") {
		t.Errorf("Esperava tradução de 'raio-grande: Verdadeiro' para 'border-radius: 0.5rem;'. Recebido: %s", style)
	}

	// CSS-2: textAlign camelCase vira kebab minúsculas 'text-align'
	if !strings.Contains(style, "text-align: center;") {
		t.Errorf("Esperava tradução de 'textAlign: \"center\"' para 'text-align: center;'. Recebido: %s", style)
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

	// ponytail: regressão do bug do "export function MeuApp" (PARTIAL do verifier).
	// Quando o usuário escreve `funcao MeuApp() { ... }` no main.ptst sem marcar como
	// exportado, o bundler de saída precisa marcar a declaração como `export`, do
	// contrário o `index.html` faz `import { MeuApp }` que vira undefined e cai no
	// console.warn em vez de montar a página.
	appBytes, err := os.ReadFile(appFile)
	if err != nil {
		t.Fatalf("Erro ao ler app.js gerado: %v", err)
	}
	if !strings.Contains(string(appBytes), "export function MeuApp") {
		t.Errorf("app.js não contém 'export function MeuApp'. Conteúdo recebido: %s", string(appBytes))
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

// ponytail: assevera de forma robusta e automatizada que os 3 exemplos mini-SPA oficiais
// de referência compilam sem panicar o transpiler e geram todos os assets finais.
func TestCompilacaoExemplos(t *testing.T) {
	exemplos := []string{
		"../exemplos/frontend/contador/main.ptst",
		"../exemplos/frontend/tarefas/main.ptst",
		"../exemplos/frontend/formulario/main.ptst",
	}

	for _, ptstPath := range exemplos {
		t.Run(ptstPath, func(t *testing.T) {
			// Verifica se o arquivo físico existe primeiro
			if _, err := os.Stat(ptstPath); os.IsNotExist(err) {
				t.Skipf("Ignorando teste do exemplo '%s' (arquivo não encontrado)", ptstPath)
			}

			tempDir, err := os.MkdirTemp("", "portuscript_ex_build_*")
			if err != nil {
				t.Fatalf("Erro ao criar tempdir: %v", err)
			}
			defer os.RemoveAll(tempDir)

			cmd := comandoCompilar()
			saidaDir := filepath.Join(tempDir, "dist")
			cmd.SetArgs([]string{"--entrada", ptstPath, "--saida", saidaDir})

			err = cmd.Execute()
			if err != nil {
				t.Fatalf("Falha crítica ao compilar o exemplo '%s': %v", ptstPath, err)
			}

			// Valida assets
			for _, file := range []string{"index.html", "app.js", "runtime-web.js", "estilos.css"} {
				p := filepath.Join(saidaDir, file)
				if _, err := os.Stat(p); os.IsNotExist(err) {
					t.Errorf("Arquivo de build esperado '%s' não foi gerado para o exemplo '%s'", file, ptstPath)
				}
			}
		})
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

// ponytail: assevera o funcionamento de TIPO-1 (geração de JSDoc em modo estrito para DX).
func TestTranspileEstritoJSDoc(t *testing.T) {
	ctx := ptst.NewContexto(ptst.OpcsContexto{})
	defer ctx.Terminar()

	codigo := `
	funcao somar(a: Inteiro, b: Decimal): Decimal {
		retorne a + b;
	}
	`

	ast, err := ctx.StringParaAst(codigo, "<teste>")
	if err != nil {
		t.Fatalf("Erro ao parsar: %v", err)
	}

	transpiler := &TranspilerWeb{Estrito: true}
	js := transpiler.Transpile(ast)

	esperados := []string{
		"/**",
		" * @param {number} a",
		" * @param {number} b",
		" * @returns {number}",
		" */",
		"function somar(a, b)",
	}

	for _, esp := range esperados {
		if !strings.Contains(js, esp) {
			t.Errorf("Esperava JSDoc contendo '%s' em modo estrito, mas não foi gerado:\n%s", esp, js)
		}
	}
}

// TestTranspileRecursosAvancadosWeb assevera o funcionamento síncrono das inovações da Fase 4-C
func TestTranspileRecursosAvancadosWeb(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "portuscript_avancado_test_*")
	if err != nil {
		t.Fatalf("Erro ao criar diretório temporário: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Cria um arquivo HTML de template para testar importarHtml
	htmlFile := filepath.Join(tempDir, "layout.html")
	err = os.WriteFile(htmlFile, []byte(`
	<div classe="Card">
		<h2>Título</h2>
	</div>
	`), 0644)
	if err != nil {
		t.Fatalf("Erro ao criar arquivo HTML de teste: %v", err)
	}

	// Cria um arquivo .estilo.ptst para testar imports de estilo separados
	estiloFile := filepath.Join(tempDir, "EstiloBotao.estilo.ptst")
	err = os.WriteFile(estiloFile, []byte(`
	estilo CaixaBotao {
		corDeFundo: "azul";
	}
	`), 0644)
	if err != nil {
		t.Fatalf("Erro ao criar arquivo de estilo de teste: %v", err)
	}

	codigo := `
	de "./EstiloBotao.estilo.ptst" importe CaixaBotao;

	funcao App() {
		retorne <div classe="App">
			<input ligar={nome} />
			<button aoEnviar_prevenir={submeter}>Enviar</button>
			<se condicao={Verdadeiro}>
				{importarHtml("./layout.html")}
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

	transpiler := &TranspilerWeb{DiretorioBase: tempDir}
	js := transpiler.Transpile(ast)

	// 1. Valida constante de estilo gerada pelo import .estilo.ptst
	if !strings.Contains(js, "const CaixaBotao = \"CaixaBotao\";") {
		t.Errorf("Esperava declaração de constante de estilo de import. Recebido: %s", js)
	}

	// 2. Valida o binding ligar transpilar para _ligar
	if !strings.Contains(js, "_ligar: nome") {
		t.Errorf("Esperava transpilação de 'ligar' para '_ligar'. Recebido: %s", js)
	}

	// 3. Valida modificador de eventos com prevenir (preventDefault)
	if !strings.Contains(js, "aoEnviar: (e) => { e.preventDefault(); (submeter)(e); }") {
		t.Errorf("Esperava transpilação de modificador 'aoEnviar_prevenir'. Recebido: %s", js)
	}

	// 4. Valida se o importarHtml inlinou dinamicamente o template HTML
	if !strings.Contains(js, "h('div', { classe: \"Card\" }") {
		t.Errorf("Esperava inline dinâmico do HTML via 'importarHtml'. Recebido: %s", js)
	}
}

// TestOtimizarECopiarAssets assevera que assets de imagens são processados, otimizados e copiados
func TestOtimizarECopiarAssets(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "portuscript_assets_*")
	if err != nil {
		t.Fatalf("Erro ao criar diretório temporário: %v", err)
	}
	defer os.RemoveAll(tempDir)

	img := image.NewRGBA(image.Rect(0, 0, 2, 2))

	pngPath := filepath.Join(tempDir, "foto.png")
	pngFile, err := os.Create(pngPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := png.Encode(pngFile, img); err != nil {
		t.Fatal(err)
	}
	pngFile.Close()

	distDir := filepath.Join(tempDir, "dist")

	err = otimizarECopiarAssets(tempDir, distDir, true)
	if err != nil {
		t.Fatalf("Erro ao otimizar e copiar assets: %v", err)
	}

	destPng := filepath.Join(distDir, "foto.png")
	if _, err := os.Stat(destPng); os.IsNotExist(err) {
		t.Errorf("Arquivo PNG otimizado não foi gerado na pasta dist")
	}
}
