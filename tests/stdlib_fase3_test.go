package tests

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/natanfeitosa/portuscript/ptst"
	_ "github.com/natanfeitosa/portuscript/stdlib"
)

func TestModuloArquivos(t *testing.T) {
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "teste.txt")

	ctx := ptst.NewContexto(ptst.OpcsContexto{})
	defer ctx.Terminar()

	codigo := `
	de "arquivos" importe escrever, ler
	
	escrever("` + tempFile + `", "Olá Portuscript!")
	var conteudo = ler("` + tempFile + `")
	`

	res, err := ptst.ExecutarString(ctx, codigo)
	if err != nil {
		t.Fatalf("Erro ao executar script com módulo arquivos: %v", err)
	}

	val, err := res.Escopo.ObterValor("conteudo")
	if err != nil {
		t.Fatalf("Não foi possível obter 'conteudo' do escopo: %v", err)
	}

	if string(val.(ptst.Texto)) != "Olá Portuscript!" {
		t.Errorf("Conteúdo lido inválido, obteve: %s", val)
	}
}

func TestModuloJson(t *testing.T) {
	ctx := ptst.NewContexto(ptst.OpcsContexto{})
	defer ctx.Terminar()

	codigo := `
	de "json" importe analisar

	var obj = analisar('{"nome": "Portuscript",' + ' "versao": 1}')
	var nome = obj["nome"]
	`

	res, err := ptst.ExecutarString(ctx, codigo)
	if err != nil {
		t.Fatalf("Erro ao executar script com módulo json: %v", err)
	}

	val, err := res.Escopo.ObterValor("nome")
	if err != nil {
		t.Fatalf("Não foi possível obter 'nome' do escopo: %v", err)
	}

	if string(val.(ptst.Texto)) != "Portuscript" {
		t.Errorf("Esperava 'Portuscript', obteve: %v", val)
	}
}

func TestModuloCripto(t *testing.T) {
	ctx := ptst.NewContexto(ptst.OpcsContexto{})
	defer ctx.Terminar()

	codigo := `
	de "cripto" importe sha256

	var hash = sha256("portuscript")
	`

	res, err := ptst.ExecutarString(ctx, codigo)
	if err != nil {
		t.Fatalf("Erro ao executar script com módulo cripto: %v", err)
	}

	val, err := res.Escopo.ObterValor("hash")
	if err != nil {
		t.Fatalf("Não foi possível obter 'hash' do escopo: %v", err)
	}

	// hash sha256 de "portuscript"
	esperado := "b3ca93021241e9f943a325d1534a3063f352899f36f88affc12c9aca9f2951e7"
	if string(val.(ptst.Texto)) != esperado {
		t.Errorf("Hash SHA256 incorreto, obteve: %s, esperava: %s", val, esperado)
	}
}

func TestModuloHttp(t *testing.T) {
	ctx := ptst.NewContexto(ptst.OpcsContexto{})
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
	var resCli = requisitar("GET", "http://localhost:8083/ola/portuscript")
	var corpo = resCli.corpo
	var cabecalho = resCli.cabecalho
	var mwHeader = cabecalho["X-Middleware"]

	server.fechar()
	`

	res, err := ptst.ExecutarString(ctx, codigo)
	if err != nil {
		t.Fatalf("Erro ao executar script HTTP: %v", err)
	}

	val, err := res.Escopo.ObterValor("corpo")
	if err != nil {
		t.Fatalf("Não foi possível obter 'corpo' do escopo: %v", err)
	}

	if string(val.(ptst.Texto)) != "Ola portuscript" {
		t.Errorf("Resposta HTTP incorreta, obteve: %v, esperava 'Ola portuscript'", val)
	}

	valMw, err := res.Escopo.ObterValor("mwHeader")
	if err != nil {
		t.Fatalf("Não foi possível obter 'mwHeader' do escopo: %v", err)
	}

	if string(valMw.(ptst.Texto)) != "Ativo" {
		t.Errorf("Header X-Middleware incorreto, obteve: %v", valMw)
	}
}

func TestModuloYaml(t *testing.T) {
	ctx := ptst.NewContexto(ptst.OpcsContexto{})
	defer ctx.Terminar()

	codigo := `
	de "yaml" importe analisar, serializar

	var obj = analisar("nome: Portuscript
versao: 1")
	var nome = obj["nome"]
	var textoYaml = serializar({"nome": "Portuscript"})
	`

	res, err := ptst.ExecutarString(ctx, codigo)
	if err != nil {
		t.Fatalf("Erro ao executar script com módulo yaml: %v", err)
	}

	valNome, err := res.Escopo.ObterValor("nome")
	if err != nil {
		t.Fatalf("Não foi possível obter 'nome': %v", err)
	}
	if string(valNome.(ptst.Texto)) != "Portuscript" {
		t.Errorf("Nome esperado 'Portuscript', obteve: %s", valNome)
	}

	valYaml, err := res.Escopo.ObterValor("textoYaml")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(valYaml.(ptst.Texto)), "nome: Portuscript") {
		t.Errorf("Serialização YAML incorreta: %s", valYaml)
	}
}

func TestModuloXml(t *testing.T) {
	ctx := ptst.NewContexto(ptst.OpcsContexto{})
	defer ctx.Terminar()

	codigo := `
	de "xml" importe analisar, serializar

	var obj = analisar("<usuario><nome>Portuscript</nome></usuario>")
	var nome = obj["usuario"]["nome"]
	var textoXml = serializar({"nome": "Portuscript"}, "usuario")
	`

	res, err := ptst.ExecutarString(ctx, codigo)
	if err != nil {
		t.Fatalf("Erro ao executar script com módulo xml: %v", err)
	}

	valNome, err := res.Escopo.ObterValor("nome")
	if err != nil {
		t.Fatalf("Não foi possível obter 'nome': %v", err)
	}
	if string(valNome.(ptst.Texto)) != "Portuscript" {
		t.Errorf("Nome esperado 'Portuscript', obteve: %s", valNome)
	}

	valXml, err := res.Escopo.ObterValor("textoXml")
	if err != nil {
		t.Fatal(err)
	}
	if string(valXml.(ptst.Texto)) != "<usuario><nome>Portuscript</nome></usuario>" {
		t.Errorf("Serialização XML incorreta, obteve: %s", valXml)
	}
}

func TestSandboxBloqueioArquivos(t *testing.T) {
	ctx := ptst.NewContexto(ptst.OpcsContexto{
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

	res, err := ptst.ExecutarString(ctx, codigo)
	if err != nil {
		t.Fatalf("Erro ao executar script: %v", err)
	}

	val, err := res.Escopo.ObterValor("erroOcorrido")
	if err != nil {
		t.Fatal(err)
	}

	if val != ptst.Verdadeiro {
		t.Error("Esperava que o acesso ao arquivo fosse bloqueado pelo Sandbox, mas não foi!")
	}
}

func TestSandboxBloqueioRede(t *testing.T) {
	ctx := ptst.NewContexto(ptst.OpcsContexto{
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

	res, err := ptst.ExecutarString(ctx, codigo)
	if err != nil {
		t.Fatalf("Erro ao executar script: %v", err)
	}

	val, err := res.Escopo.ObterValor("erroOcorrido")
	if err != nil {
		t.Fatal(err)
	}

	if val != ptst.Verdadeiro {
		t.Error("Esperava que as operações de rede fossem bloqueadas pelo Sandbox, mas não foram!")
	}
}

func TestConcorrenciaPorCanaisCsp(t *testing.T) {
	ctx := ptst.NewContexto(ptst.OpcsContexto{})

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

	res, err := ptst.ExecutarString(ctx, codigo)
	if err != nil {
		t.Fatalf("Erro ao executar script CSP: %v", err)
	}

	// Aguarda estabilização das goroutines concorrentes na VM de forma atômica
	ctx.Terminar()

	val, err := res.Escopo.ObterValor("mensagemRecebida")
	if err != nil {
		t.Fatal(err)
	}

	if string(val.(ptst.Texto)) != "Ola via Canal!" {
		t.Errorf("Recebimento incorreto via Canal, obteve: %v, esperava 'Ola via Canal!'", val)
	}
}
