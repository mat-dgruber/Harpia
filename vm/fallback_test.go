package vm

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/mat-dgruber/Harpia/hrp"
)

// valorGlobal executa o código pelo caminho padrão (VM com fallback para o AST) e lê uma variável.
func valorGlobal(t *testing.T, codigo, nome string) (hrp.Objeto, error) {
	t.Helper()
	ctx := hrp.NewContexto(hrp.OpcsContexto{})
	defer ctx.Terminar()
	mod, err := hrp.ExecutarString(ctx, codigo)
	if err != nil {
		return nil, err
	}
	return mod.Escopo.ObterValor(nome)
}

func TestFallbackOperadorPotencia(t *testing.T) {
	v, err := valorGlobal(t, "var r = 2 ** 3", "r")
	if err != nil {
		t.Fatalf("'2 ** 3' deveria cair no interpretador: %v", err)
	}
	if v != hrp.Inteiro(8) && v != hrp.Decimal(8) {
		t.Fatalf("esperava 8, veio %v", v)
	}
}

func TestErroDeUsuarioNaoReexecutaPrograma(t *testing.T) {
	codigo := `var contador = 0
contador = contador + 1
tente {
    assegura Falso, "x não implementada na VM"
} capture (erro) { }
`
	v, err := valorGlobal(t, codigo, "contador")
	if err != nil {
		t.Fatal(err)
	}
	if v != hrp.Inteiro(1) {
		t.Fatalf("programa executado mais de uma vez: contador=%v", v)
	}

	// Mesmo sem capturar: o erro de runtime não pode disparar reexecução.
	_, err = valorGlobal(t, "assegura Falso, \"x não suportada na VM\"", "nada")
	if err == nil || !strings.Contains(err.Error(), "não suportada na VM") {
		t.Fatalf("esperava o erro do usuário propagado, veio %v", err)
	}
}

// TestAssincronaAposTerminarNaoDerrubaProcesso roda em subprocesso porque, antes da correção,
// o Done() sem Add() do WaitGroup causava pânico fatal que derrubaria o runner de testes.
func TestAssincronaAposTerminarNaoDerrubaProcesso(t *testing.T) {
	if os.Getenv("HARPIA_TESTE_FILHO") == "1" {
		ctx := hrp.NewContexto(hrp.OpcsContexto{})
		mod, err := hrp.ExecutarString(ctx, "assincrono funcao tarefa() { retorne 1 }")
		if err != nil {
			t.Fatal(err)
		}
		fn, err := mod.Escopo.ObterValor("tarefa")
		if err != nil {
			t.Fatal(err)
		}
		ctx.Terminar()
		prom, err := hrp.Chamar(fn, hrp.Tupla{})
		if err != nil {
			return // rejeitar na chamada também é aceitável
		}
		if p, ok := prom.(*hrp.Promessa); ok {
			fim := make(chan error, 1)
			p.Registre(func(_ hrp.Objeto, e error) { fim <- e })
			if e := <-fim; e == nil {
				t.Fatal("promessa deveria ser rejeitada com o contexto fechado")
			}
		}
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=^TestAssincronaAposTerminarNaoDerrubaProcesso$")
	cmd.Env = append(os.Environ(), "HARPIA_TESTE_FILHO=1")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("processo filho caiu: %v\n%s", err, out)
	}
}

// TestErroDeRuntimeNaoReexecuta captura o stdout: se o fallback reexecutar o programa,
// "passo" aparece duas vezes.
func TestErroDeRuntimeNaoReexecuta(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	antigo := os.Stdout
	os.Stdout = w
	ctx := hrp.NewContexto(hrp.OpcsContexto{})
	_, _ = hrp.ExecutarString(ctx, "imprimir(\"passo\")\nde \"modulo não suportada na VM\" importe x")
	ctx.Terminar()
	w.Close()
	os.Stdout = antigo
	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	if c := strings.Count(string(buf[:n]), "passo"); c != 1 {
		t.Fatalf("programa executado %d vezes (saída %q)", c, buf[:n])
	}
}
