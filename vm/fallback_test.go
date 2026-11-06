package vm

import (
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

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
	_, _ = hrp.ExecutarString(ctx, "imprimir(\"passo\")\nvar n = Inteiro(\"x não suportada na VM\")")
	ctx.Terminar()
	w.Close()
	os.Stdout = antigo
	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	if c := strings.Count(string(buf[:n]), "passo"); c != 1 {
		t.Fatalf("programa executado %d vezes (saída %q)", c, buf[:n])
	}
}

// TestRecursaoInfinitaViraErroDePilha roda em subprocesso: antes da correção o processo
// morria com "fatal error: stack overflow" (não recuperável).
func TestRecursaoInfinitaViraErroDePilha(t *testing.T) {
	if os.Getenv("HARPIA_TESTE_FILHO_RECURSAO") == "1" {
		ctx := hrp.NewContexto(hrp.OpcsContexto{})
		defer ctx.Terminar()
		_, err := hrp.ExecutarString(ctx, "funcao f(n) { retorne f(n + 1) }\nf(0)")
		if err == nil || !strings.Contains(err.Error(), "profundidade") {
			t.Fatalf("esperava ErroDePilha de profundidade, veio %v", err)
		}
		// recursão legítima e profunda continua funcionando
		mod, err := hrp.ExecutarString(ctx, "funcao soma(n) { se n == 0 { retorne 0 } retorne n + soma(n - 1) }\nvar r = soma(500)")
		if err != nil {
			t.Fatalf("recursão de 500 níveis deveria funcionar: %v", err)
		}
		if v, _ := mod.Escopo.ObterValor("r"); v != hrp.Inteiro(125250) {
			t.Fatalf("soma(500) = %v", v)
		}
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=^TestRecursaoInfinitaViraErroDePilha$")
	cmd.Env = append(os.Environ(), "HARPIA_TESTE_FILHO_RECURSAO=1")
	done := make(chan struct{})
	var out []byte
	var err error
	go func() { out, err = cmd.CombinedOutput(); close(done) }()
	select {
	case <-done:
	case <-time.After(60 * time.Second):
		_ = cmd.Process.Kill()
		t.Fatal("recursão infinita não terminou em 60s")
	}
	if err != nil {
		t.Fatalf("processo filho falhou: %v\n%.2000s", err, out)
	}
}

// M30: sequências de escape em strings eram mantidas literalmente ("\n" imprimia barra+n),
// tanto na VM quanto no interpretador.
func TestEscapesEmTextoLiteral(t *testing.T) {
	casos := map[string]string{
		`var r = "a\nb"`:          "a\nb",
		`var r = "a\tb"`:          "a\tb",
		`var r = "aspas \"x\""`:   `aspas "x"`,
		`var r = "barra \\ fim"`:  `barra \ fim`,
		`var r = 'simples \'x\''`: `simples 'x'`,
		`var r = "chave \{x\}"`:   "chave {x}",
		`var r = "regex \d+"`:     `regex \d+`,
		"var r = `bruta \\n`":     `bruta \n`,
	}
	for codigo, esperado := range casos {
		v, err := valorGlobal(t, codigo, "r")
		if err != nil {
			t.Errorf("%s: %v", codigo, err)
			continue
		}
		if v != hrp.Texto(esperado) {
			t.Errorf("%s: esperava %q, veio %q", codigo, esperado, v)
		}
	}
}

// O laço para-em da VM percorre por índice; Mapa, sequencia() e Texto precisam funcionar igual ao AST.
func TestParaEmIteraveisNaoIndexados(t *testing.T) {
	casos := map[string]hrp.Objeto{
		"var r = 0\npara i em sequencia(4) { r = r + i }":                hrp.Inteiro(6),
		"var r = \"\"\npara c em \"abc\" { r = c + r }":                  hrp.Texto("cba"),
		"var r = 0\npara par em {\"a\": 1, \"b\": 2} { r = r + par[1] }": hrp.Inteiro(3),
		"var r = 0\npara x em [1, 2, 3] { r = r + x }":                   hrp.Inteiro(6),
	}
	for codigo, esperado := range casos {
		v, err := valorGlobal(t, codigo, "r")
		if err != nil {
			t.Errorf("%s: %v", codigo, err)
			continue
		}
		if v != esperado {
			t.Errorf("%s: esperava %v, veio %v", codigo, esperado, v)
		}
	}
}

// Comentário no fim da linha não pode engolir o '\n' que termina a instrução.
func TestComentarioNoFimDaLinha(t *testing.T) {
	for _, codigo := range []string{
		"var r = 1 # nota\nr = r + 1",
		"var r = 1 // nota\nr = r + 1",
	} {
		v, err := valorGlobal(t, codigo, "r")
		if err != nil || v != hrp.Inteiro(2) {
			t.Errorf("%q: esperava 2, veio %v (%v)", codigo, v, err)
		}
	}
}

// assegura(cond, "msg") com parênteses precisa levar a mensagem para o erro.
func TestAsseguraComParentesesMantemMensagem(t *testing.T) {
	v, err := valorGlobal(t, "var r = \"\"\ntente { assegura(1 == 2, \"idade inválida\") } capture (falha) { r = falha.mensagem }", "r")
	if err != nil || v != hrp.Texto("idade inválida") {
		t.Fatalf("esperava a mensagem, veio %v (%v)", v, err)
	}
}

// Pipe em várias linhas: a linha que começa com |> continua a expressão anterior.
func TestPipeEmVariasLinhas(t *testing.T) {
	v, err := valorGlobal(t, "funcao dobro(x) { retorne x * 2 }\nvar r = 3\n    |> dobro\n\n    |> dobro", "r")
	if err != nil || v != hrp.Inteiro(12) {
		t.Fatalf("esperava 12, veio %v (%v)", v, err)
	}
}

// No backend, o definidor do sinal precisa guardar o valor (antes era um stub vazio).
func TestSinalNoBackendGuardaValor(t *testing.T) {
	v, err := valorGlobal(t, "var [n, definirN] = sinal(1)\ndefinirN(5)\ndefinirN(funcao(x) { retorne x + 1 })\nvar r = n()", "r")
	if err != nil || v != hrp.Inteiro(6) {
		t.Fatalf("esperava 6, veio %v (%v)", v, err)
	}
}
