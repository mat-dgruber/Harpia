package sistema

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/mat-dgruber/Harpia/hrp"
)

var (
	binOnce sync.Once
	binPath string
	binErr  error
)

// harpiaCompilado compila o CLI uma vez para testar executarCodigoHarpia de ponta a ponta.
func harpiaCompilado(t *testing.T) string {
	t.Helper()
	if testing.Short() {
		t.Skip("compila o CLI; pulado em -short")
	}
	binOnce.Do(func() {
		dir, err := os.MkdirTemp("", "harpia_bin_*")
		if err != nil {
			binErr = err
			return
		}
		binPath = filepath.Join(dir, "harpia")
		out, err := exec.Command("go", "build", "-o", binPath, "github.com/mat-dgruber/Harpia").CombinedOutput()
		if err != nil {
			binErr = err
			t.Logf("%s", out)
		}
	})
	if binErr != nil {
		t.Fatalf("falha ao compilar o CLI: %v", binErr)
	}
	return binPath
}

func executarIsolado(t *testing.T, codigo string) (string, string) {
	t.Helper()
	bin := harpiaCompilado(t)
	antigo := caminhoExecutavel
	caminhoExecutavel = func() (string, error) { return bin, nil }
	t.Cleanup(func() { caminhoExecutavel = antigo })

	mod, err := hrp.NewContexto(hrp.OpcsContexto{}).ObterModulo("sistema")
	if err != nil {
		ctx := hrp.NewContexto(hrp.OpcsContexto{})
		hrp.MultiImporteModulo(ctx, "sistema")
		mod, err = ctx.ObterModulo("sistema")
		if err != nil {
			t.Fatalf("módulo sistema indisponível: %v", err)
		}
	}
	fn, err := mod.M__obtem_attributo__("executarCodigoHarpia")
	if err != nil {
		t.Fatal(err)
	}
	res, err := hrp.Chamar(fn, hrp.Tupla{hrp.Texto(codigo)})
	if err != nil {
		t.Fatal(err)
	}
	m := res.(hrp.Mapa)
	return string(m["saida"].(hrp.Texto)), string(m["erro"].(hrp.Texto))
}

func TestExecutarCodigoHarpiaRodaCodigoSimples(t *testing.T) {
	saida, erro := executarIsolado(t, "imprimir(1 + 1)")
	if strings.TrimSpace(saida) != "2" {
		t.Fatalf("esperava '2', saida=%q erro=%q", saida, erro)
	}
}

func TestExecutarCodigoHarpiaEstaEmSandbox(t *testing.T) {
	saida, erro := executarIsolado(t, "de \"sistema\" importe executar\nimprimir(executar(\"echo\", [\"pwned\"]))")
	if strings.Contains(saida, "pwned") || !strings.Contains(erro, "Acesso Negado") {
		t.Fatalf("subprocesso não está em sandbox: saida=%q erro=%q", saida, erro)
	}
}

func TestExecutarCodigoHarpiaNaoHerdaSegredos(t *testing.T) {
	t.Setenv("SEGREDO_DO_HOST", "nao-vaze")
	saida, erro := executarIsolado(t, "de \"sistema\" importe Sistema\nimprimir(Sistema[\"variaveisAmbiente\"])")
	if strings.Contains(saida+erro, "nao-vaze") {
		t.Fatalf("segredo do processo pai vazou para o subprocesso: %q", saida)
	}
}

func TestExecutarCodigoHarpiaRespeitaTempoLimite(t *testing.T) {
	antigo := TempoLimiteExecucaoIsolada
	TempoLimiteExecucaoIsolada = 2 * time.Second
	t.Cleanup(func() { TempoLimiteExecucaoIsolada = antigo })

	inicio := time.Now()
	_, erro := executarIsolado(t, "enquanto Verdadeiro { }")
	if d := time.Since(inicio); d > 6*time.Second {
		t.Fatalf("não respeitou o tempo limite: %v", d)
	}
	if !strings.Contains(erro, "Tempo limite") {
		t.Fatalf("esperava aviso de tempo limite, veio %q", erro)
	}
}

func TestBufferLimitadoTrunca(t *testing.T) {
	b := &bufferLimitado{limite: 4}
	n, err := b.Write([]byte("abcdef"))
	if n != 6 || err != nil {
		t.Fatalf("Write deve aceitar tudo: n=%d err=%v", n, err)
	}
	if !strings.HasPrefix(b.String(), "abcd") || !strings.Contains(b.String(), "truncada") {
		t.Fatalf("esperava truncar em 4 bytes: %q", b.String())
	}
}

func chamarExecutar(t *testing.T, args ...hrp.Objeto) hrp.Mapa {
	t.Helper()
	ctx := hrp.NewContexto(hrp.OpcsContexto{})
	defer ctx.Terminar()
	mod, err := hrp.ExecutarString(ctx, "de \"sistema\" importe executar")
	if err != nil {
		t.Fatal(err)
	}
	fn, _ := mod.Escopo.ObterValor("executar")
	res, err := hrp.Chamar(fn, hrp.Tupla(args))
	if err != nil {
		t.Fatal(err)
	}
	return res.(hrp.Mapa)
}

// M14: executar devolve o código de saída e respeita tempo limite.
func TestExecutarDevolveCodigoETempoLimite(t *testing.T) {
	ok := chamarExecutar(t, hrp.Texto("sh"), hrp.Tupla{hrp.Texto("-c"), hrp.Texto("exit 3")})
	if ok["codigo"] != hrp.Inteiro(3) {
		t.Fatalf("esperava codigo 3, veio %v", ok["codigo"])
	}
	inicio := time.Now()
	lento := chamarExecutar(t, hrp.Texto("sleep"), hrp.Tupla{hrp.Texto("30")}, hrp.Mapa{"tempoLimite": hrp.Inteiro(1)})
	if d := time.Since(inicio); d > 5*time.Second {
		t.Fatalf("tempoLimite ignorado: %v", d)
	}
	if !strings.Contains(string(lento["erro"].(hrp.Texto)), "Tempo limite") {
		t.Fatalf("esperava aviso de tempo limite, veio %v", lento["erro"])
	}
	inexistente := chamarExecutar(t, hrp.Texto("comando-que-nao-existe-xyz"))
	if inexistente["codigo"] != hrp.Inteiro(-1) || inexistente["erro"] == hrp.Texto("") {
		t.Fatalf("comando inexistente deveria dar codigo -1 e erro, veio %v", inexistente)
	}
}
