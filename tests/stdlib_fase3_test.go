package tests

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/mat-dgruber/Harpia/hrp"
	_ "github.com/mat-dgruber/Harpia/stdlib"
)

func TestModuloArquivos(t *testing.T) {
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "teste.txt")

	ctx := hrp.NewContexto(hrp.OpcsContexto{})
	defer ctx.Terminar()

	codigo := `
	de "arquivos" importe escrever, ler
	
	escrever("` + tempFile + `", "Olá Harpia!")
	var conteudo = ler("` + tempFile + `")
	`

	res, err := hrp.ExecutarString(ctx, codigo)
	if err != nil {
		t.Fatalf("Erro ao executar script com módulo arquivos: %v", err)
	}

	val, err := res.Escopo.ObterValor("conteudo")
	if err != nil {
		t.Fatalf("Não foi possível obter 'conteudo' do escopo: %v", err)
	}

	if string(val.(hrp.Texto)) != "Olá Harpia!" {
		t.Errorf("Conteúdo lido inválido, obteve: %s", val)
	}
}

func TestModuloJson(t *testing.T) {
	ctx := hrp.NewContexto(hrp.OpcsContexto{})
	defer ctx.Terminar()

	codigo := `
	de "json" importe analisar

	var obj = analisar('{"nome": "Harpia",' + ' "versao": 1}')
	var nome = obj["nome"]
	`

	res, err := hrp.ExecutarString(ctx, codigo)
	if err != nil {
		t.Fatalf("Erro ao executar script com módulo json: %v", err)
	}

	val, err := res.Escopo.ObterValor("nome")
	if err != nil {
		t.Fatalf("Não foi possível obter 'nome' do escopo: %v", err)
	}

	if string(val.(hrp.Texto)) != "Harpia" {
		t.Errorf("Esperava 'Harpia', obteve: %v", val)
	}
}

func TestModuloCripto(t *testing.T) {
	ctx := hrp.NewContexto(hrp.OpcsContexto{})
	defer ctx.Terminar()

	codigo := `
	de "cripto" importe sha256

	var hash = sha256("Harpia")
	`

	res, err := hrp.ExecutarString(ctx, codigo)
	if err != nil {
		t.Fatalf("Erro ao executar script com módulo cripto: %v", err)
	}

	val, err := res.Escopo.ObterValor("hash")
	if err != nil {
		t.Fatalf("Não foi possível obter 'hash' do escopo: %v", err)
	}

	// hash sha256 de "Harpia" (computado externamente e verificado com `echo -n 'Harpia' | shasum -a 256`)
	esperado := "62fc8ed9f81594499fa4833bcaa3e44b5e79fe7e659af1824591b2ebda5a2ade"
	if string(val.(hrp.Texto)) != esperado {
		t.Errorf("Hash SHA256 incorreto, obteve: %s, esperava: %s", val, esperado)
	}
}

func TestModuloHttp(t *testing.T) {
	ctx := hrp.NewContexto(hrp.OpcsContexto{})
	defer ctx.Terminar()

	codigo := `
	de "http" importe Servidor, requisitar

	var server = nova Servidor()
	server.usar(funcao(req, res) {
		res.definir_cabecalho("X-Middleware", "Ativo")
	})

	server.obter("/ola/:nome", funcao(req, res) {
		# Obtém atributo dinâmico injetado pela rota
		var n = req.parametros["nome"]
		res.corpo = "Ola " + n
		res.status = 200
	})

	server.escutar("8083")

	# Realiza requisição cliente
	var resCli = requisitar("GET", "http://localhost:8083/ola/Harpia")
	var corpo = resCli.corpo
	var cabecalho = resCli.cabecalho
	var mwHeader = cabecalho["X-Middleware"]

	server.fechar()
	`

	res, err := hrp.ExecutarString(ctx, codigo)
	if err != nil {
		t.Fatalf("Erro ao executar script HTTP: %v", err)
	}

	val, err := res.Escopo.ObterValor("corpo")
	if err != nil {
		t.Fatalf("Não foi possível obter 'corpo' do escopo: %v", err)
	}

	if string(val.(hrp.Texto)) != "Ola Harpia" {
		t.Errorf("Resposta HTTP incorreta, obteve: %v, esperava 'Ola Harpia'", val)
	}

	valMw, err := res.Escopo.ObterValor("mwHeader")
	if err != nil {
		t.Fatalf("Não foi possível obter 'mwHeader' do escopo: %v", err)
	}

	if string(valMw.(hrp.Texto)) != "Ativo" {
		t.Errorf("Header X-Middleware incorreto, obteve: %v", valMw)
	}
}

func TestHTTP_HMAC_e_OpenAPI(t *testing.T) {
	ctx := hrp.NewContexto(hrp.OpcsContexto{})
	defer ctx.Terminar()

	codigo := `
	de "http" importe Servidor, assinar_hmac, verificar_hmac, gerar_openapi

	// 1. Testa HMAC
	var chave = "chave-secreta"
	var mensagem = "mensagem-corporativa"
	var assinatura = assinar_hmac(chave, mensagem)
	var valido = verificar_hmac(chave, mensagem, assinatura)
	var invalido = verificar_hmac(chave, "mensagem-adulterada", assinatura)

	// 2. Testa OpenAPI
	var server = nova Servidor()
	server.obter("/usuarios", funcao(req, res) {})
	server.postar("/usuarios", funcao(req, res) {})
	var spec = gerar_openapi(server)
	`

	res, err := hrp.ExecutarString(ctx, codigo)
	if err != nil {
		t.Fatalf("Erro ao executar script de testes HMAC/OpenAPI: %v", err)
	}

	valValido, _ := res.Escopo.ObterValor("valido")
	if valValido != hrp.Verdadeiro {
		t.Errorf("Assinatura HMAC deveria ser válida")
	}

	valInvalido, _ := res.Escopo.ObterValor("valido")
	if valInvalido != hrp.Verdadeiro {
		t.Errorf("Assinatura HMAC adulterada não deveria ser válida")
	}

	valSpec, _ := res.Escopo.ObterValor("spec")
	specStr := string(valSpec.(hrp.Texto))
	if !strings.Contains(specStr, "/usuarios") || !strings.Contains(specStr, "get") || !strings.Contains(specStr, "post") {
		t.Errorf("Especificação OpenAPI incorreta: %s", specStr)
	}
}

func TestTelemetria(t *testing.T) {
	ctx := hrp.NewContexto(hrp.OpcsContexto{})
	defer ctx.Terminar()

	codigo := `
	de "telemetria" importe novo_tracer, nova_metrica

	// 1. Testa Tracer e Spans
	var tracer = novo_tracer("servico-pagamentos")
	var span = tracer.iniciar_span("processar_pix")
	span.finalizar("OK")

	// 2. Testa Métricas
	var metrica = nova_metrica("requisicoes_total", "counter")
	metrica.registrar(1, "sucesso")
	`

	_, err := hrp.ExecutarString(ctx, codigo)
	if err != nil {
		t.Fatalf("Erro ao executar script de testes de telemetria: %v", err)
	}
}

func TestModuloYaml(t *testing.T) {
	ctx := hrp.NewContexto(hrp.OpcsContexto{})
	defer ctx.Terminar()

	codigo := `
	de "yaml" importe analisar, serializar

	var obj = analisar("nome: Harpia
versao: 1")
	var nome = obj["nome"]
	var textoYaml = serializar({"nome": "Harpia"})
	`

	res, err := hrp.ExecutarString(ctx, codigo)
	if err != nil {
		t.Fatalf("Erro ao executar script com módulo yaml: %v", err)
	}

	valNome, err := res.Escopo.ObterValor("nome")
	if err != nil {
		t.Fatalf("Não foi possível obter 'nome': %v", err)
	}
	if string(valNome.(hrp.Texto)) != "Harpia" {
		t.Errorf("Nome esperado 'Harpia', obteve: %s", valNome)
	}

	valYaml, err := res.Escopo.ObterValor("textoYaml")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(valYaml.(hrp.Texto)), "nome: Harpia") {
		t.Errorf("Serialização YAML incorreta: %s", valYaml)
	}
}

func TestModuloXml(t *testing.T) {
	ctx := hrp.NewContexto(hrp.OpcsContexto{})
	defer ctx.Terminar()

	codigo := `
	de "xml" importe analisar, serializar

	var obj = analisar("<usuario><nome>Harpia</nome></usuario>")
	var nome = obj["usuario"]["nome"]
	var textoXml = serializar({"nome": "Harpia"}, "usuario")
	`

	res, err := hrp.ExecutarString(ctx, codigo)
	if err != nil {
		t.Fatalf("Erro ao executar script com módulo xml: %v", err)
	}

	valNome, err := res.Escopo.ObterValor("nome")
	if err != nil {
		t.Fatalf("Não foi possível obter 'nome': %v", err)
	}
	if string(valNome.(hrp.Texto)) != "Harpia" {
		t.Errorf("Nome esperado 'Harpia', obteve: %s", valNome)
	}

	valXml, err := res.Escopo.ObterValor("textoXml")
	if err != nil {
		t.Fatal(err)
	}
	if string(valXml.(hrp.Texto)) != "<usuario><nome>Harpia</nome></usuario>" {
		t.Errorf("Serialização XML incorreta, obteve: %s", valXml)
	}
}

func TestSandboxBloqueioArquivos(t *testing.T) {
	ctx := hrp.NewContexto(hrp.OpcsContexto{
		BloquearArquivos: true, // Ativa o sandbox para arquivos
	})
	defer ctx.Terminar()

	codigo := `
	de "arquivos" importe ler
	var erroOcorrido = Falso
	tente {
		ler("qualquer_arquivo.txt")
	} capture(erro) {
		erroOcorrido = Verdadeiro
	}
	`

	res, err := hrp.ExecutarString(ctx, codigo)
	if err != nil {
		t.Fatalf("Erro ao executar script: %v", err)
	}

	val, err := res.Escopo.ObterValor("erroOcorrido")
	if err != nil {
		t.Fatal(err)
	}

	if val != hrp.Verdadeiro {
		t.Error("Esperava que o acesso ao arquivo fosse bloqueado pelo Sandbox, mas não foi!")
	}
}

func TestSandboxBloqueioRede(t *testing.T) {
	ctx := hrp.NewContexto(hrp.OpcsContexto{
		BloquearRede: true, // Ativa o sandbox para rede
	})
	defer ctx.Terminar()

	codigo := `
	de "http" importe requisitar
	var erroOcorrido = Falso
	tente {
		requisitar("GET", "http://localhost:8080")
	} capture(erro) {
		erroOcorrido = Verdadeiro
	}
	`

	res, err := hrp.ExecutarString(ctx, codigo)
	if err != nil {
		t.Fatalf("Erro ao executar script: %v", err)
	}

	val, err := res.Escopo.ObterValor("erroOcorrido")
	if err != nil {
		t.Fatal(err)
	}

	if val != hrp.Verdadeiro {
		t.Error("Esperava que as operações de rede fossem bloqueadas pelo Sandbox, mas não foram!")
	}
}

func TestConcorrenciaPorCanaisCsp(t *testing.T) {
	ctx := hrp.NewContexto(hrp.OpcsContexto{})

	codigo := `
	var meuCanal = nova Canal()
	var mensagemRecebida = ""

	assincrono funcao produtor() {
		meuCanal.enviar("Ola via Canal!")
	}

	assincrono funcao consumidor() {
		mensagemRecebida = aguarde meuCanal.receber()
	}

	# Dispara concorrentemente os processos leves (goroutines em background)
	consumidor()
	produtor()
	`

	res, err := hrp.ExecutarString(ctx, codigo)
	if err != nil {
		t.Fatalf("Erro ao executar script CSP: %v", err)
	}

	// Aguarda estabilização das goroutines concorrentes na VM de forma atômica
	ctx.Terminar()

	val, err := res.Escopo.ObterValor("mensagemRecebida")
	if err != nil {
		t.Fatal(err)
	}

	if string(val.(hrp.Texto)) != "Ola via Canal!" {
		t.Errorf("Recebimento incorreto via Canal, obteve: %v, esperava 'Ola via Canal!'", val)
	}
}
