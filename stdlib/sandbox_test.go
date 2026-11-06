package stdlib

import (
	"strings"
	"testing"

	"github.com/mat-dgruber/Harpia/hrp"
)

// funcaoDoModulo importa `nome` do módulo dentro do contexto e devolve a função.
func funcaoDoModulo(t *testing.T, ctx *hrp.Contexto, modulo, nome string) hrp.Objeto {
	t.Helper()
	mod, err := hrp.ExecutarString(ctx, "de \""+modulo+"\" importe "+nome)
	if err != nil {
		t.Fatalf("importar %s.%s: %v", modulo, nome, err)
	}
	fn, err := mod.Escopo.ObterValor(nome)
	if err != nil {
		t.Fatal(err)
	}
	return fn
}

func esperaNegado(t *testing.T, err error) {
	t.Helper()
	if err == nil || !strings.Contains(err.Error(), "Acesso Negado") {
		t.Fatalf("esperava 'Acesso Negado', veio %v", err)
	}
}

// Um contexto sem sandbox criado depois (que vira o "ativo" global) não pode liberar
// o acesso de um script que nasceu em sandbox.
func TestSandboxNaoDependeDoContextoAtivo(t *testing.T) {
	bloqueado := hrp.NewContexto(hrp.OpcsContexto{BloquearArquivos: true, BloquearRede: true})
	defer bloqueado.Terminar()
	executar := funcaoDoModulo(t, bloqueado, "sistema", "executar")
	requisitar := funcaoDoModulo(t, bloqueado, "http", "requisitar")

	livre := hrp.NewContexto(hrp.OpcsContexto{})
	defer livre.Terminar()

	_, err := hrp.Chamar(executar, hrp.Tupla{hrp.Texto("echo"), hrp.Tupla{hrp.Texto("x")}})
	esperaNegado(t, err)
	_, err = hrp.Chamar(requisitar, hrp.Tupla{hrp.Texto("GET"), hrp.Texto("http://127.0.0.1:1")})
	esperaNegado(t, err)
}

// Sem nenhum contexto para consultar, a operação é negada (fail-closed).
func TestExigirPermissaoSemContextoNega(t *testing.T) {
	ctx := hrp.NewContexto(hrp.OpcsContexto{})
	ctx.Terminar() // limpa o contexto ativo global
	if hrp.ObterContextoAtivo() != nil {
		t.Skip("outro contexto ativo no processo")
	}
	esperaNegado(t, hrp.ExigirPermissao(nil, hrp.RecursoArquivos))
	esperaNegado(t, hrp.ExigirPermissao(hrp.Nulo, hrp.RecursoRede))
}

// Todo módulo que toca disco, rede ou processos precisa respeitar o sandbox.
func TestSandboxCobreTodosOsModulosDeIO(t *testing.T) {
	casos := map[string]string{
		"sqlite":   "de \"bd\" importe conectarSqlite\nconectarSqlite(\":memory:\")",
		"postgres": "de \"bd\" importe conectarPostgres\nconectarPostgres(\"postgres://u:s@127.0.0.1:1/db\")",
		"redis":    "de \"bd\" importe conectarRedis\nconectarRedis(\"127.0.0.1:1\")",
		"qdrant":   "de \"bd\" importe conectarQdrant\nconectarQdrant(\"http://127.0.0.1:1\", \"c\")",
		"email":    "de \"email\" importe enviar\nenviar({\"servidor\": \"127.0.0.1\", \"porta\": 1, \"para\": \"a@b.c\"})",
		"ffi":      "de \"ffi\" importe abrir\nabrir(\"/tmp/inexistente.so\")",
		"soquete":  "de \"soquete\" importe Soquete, AF_INET, SOCK_STREAM\nSoquete(AF_INET, SOCK_STREAM)",
		"ia":       "de \"ia\" importe Agente\nvar a = Agente(\"ollama\", \"m\")\na.perguntar(\"oi\")",
		"ambiente": "de \"sistema\" importe Sistema\nimprimir(Sistema[\"variaveisAmbiente\"])",
	}
	for nome, codigo := range casos {
		t.Run(nome, func(t *testing.T) {
			ctx := hrp.NewContexto(hrp.OpcsContexto{BloquearArquivos: true, BloquearRede: true})
			defer ctx.Terminar()
			_, err := hrp.ExecutarString(ctx, codigo)
			esperaNegado(t, err)
		})
	}
}

// Controle positivo: sem sandbox, o atributo continua funcionando (P5: segue como atributo).
func TestVariaveisAmbienteSemSandbox(t *testing.T) {
	t.Setenv("HARPIA_TESTE_AMBIENTE", "ok")
	ctx := hrp.NewContexto(hrp.OpcsContexto{})
	defer ctx.Terminar()
	mod, err := hrp.ExecutarString(ctx, "de \"sistema\" importe Sistema\nvar v = Sistema[\"variaveisAmbiente\"][\"HARPIA_TESTE_AMBIENTE\"]\nvar so = Sistema[\"NOME\"]")
	if err != nil {
		t.Fatal(err)
	}
	if v, _ := mod.Escopo.ObterValor("v"); v != hrp.Texto("ok") {
		t.Fatalf("variaveisAmbiente deveria ler o ambiente, veio %v", v)
	}
	if so, _ := mod.Escopo.ObterValor("so"); so == nil || so == hrp.Nulo {
		t.Fatal("NOME deveria continuar disponível")
	}
}
