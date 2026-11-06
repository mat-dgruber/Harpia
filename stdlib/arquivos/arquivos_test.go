package arquivos

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mat-dgruber/Harpia/hrp"
)

func chamar(t *testing.T, fn func(hrp.Objeto, hrp.Tupla) (hrp.Objeto, error), args ...hrp.Objeto) (hrp.Objeto, error) {
	t.Helper()
	return fn(nil, hrp.Tupla(args))
}

func TestArquivosCicloDeVida(t *testing.T) {
	ctx := hrp.NewContexto(hrp.OpcsContexto{})
	defer ctx.Terminar()
	dir := t.TempDir()
	alvo := hrp.Texto(filepath.Join(dir, "a.txt"))

	if _, err := chamar(t, met_arq_escrever, alvo, hrp.Texto("olá")); err != nil {
		t.Fatal(err)
	}
	if _, err := chamar(t, met_arq_acrescentar, alvo, hrp.Texto(" mundo")); err != nil {
		t.Fatal(err)
	}
	conteudo, err := chamar(t, met_arq_ler, alvo)
	if err != nil || conteudo != hrp.Texto("olá mundo") {
		t.Fatalf("ler = %v (%v)", conteudo, err)
	}

	novo := hrp.Texto(filepath.Join(dir, "sub", "b.txt"))
	os.MkdirAll(filepath.Join(dir, "sub"), 0755)
	if _, err := chamar(t, met_arq_renomear, alvo, novo); err != nil {
		t.Fatal(err)
	}
	lista, err := chamar(t, met_arq_caminhar, hrp.Texto(dir))
	if err != nil || !strings.Contains(fmtLista(lista), "b.txt") {
		t.Fatalf("caminhar não listou b.txt: %v (%v)", lista, err)
	}
	if _, err := chamar(t, met_arq_remover, hrp.Texto(filepath.Join(dir, "sub"))); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(string(novo)); !os.IsNotExist(err) {
		t.Fatal("remover deveria apagar a pasta")
	}
	if _, err := chamar(t, met_arq_ler, hrp.Texto(filepath.Join(dir, "nao-existe"))); err == nil {
		t.Fatal("ler arquivo inexistente deveria falhar")
	}
}

func TestArquivosNegadosEmSandbox(t *testing.T) {
	ctx := hrp.NewContexto(hrp.OpcsContexto{BloquearArquivos: true})
	defer ctx.Terminar()
	alvo := hrp.Texto(filepath.Join(t.TempDir(), "x.txt"))
	casos := map[string]func() (hrp.Objeto, error){
		"ler":         func() (hrp.Objeto, error) { return chamar(t, met_arq_ler, alvo) },
		"escrever":    func() (hrp.Objeto, error) { return chamar(t, met_arq_escrever, alvo, hrp.Texto("x")) },
		"acrescentar": func() (hrp.Objeto, error) { return chamar(t, met_arq_acrescentar, alvo, hrp.Texto("x")) },
		"remover":     func() (hrp.Objeto, error) { return chamar(t, met_arq_remover, alvo) },
		"renomear":    func() (hrp.Objeto, error) { return chamar(t, met_arq_renomear, alvo, alvo) },
		"caminhar":    func() (hrp.Objeto, error) { return chamar(t, met_arq_caminhar, alvo) },
		"resolver":    func() (hrp.Objeto, error) { return chamar(t, met_arq_resolver, alvo) },
	}
	for nome, op := range casos {
		if _, err := op(); err == nil || !strings.Contains(err.Error(), "Acesso Negado") {
			t.Errorf("%s deveria ser negado em sandbox, veio %v", nome, err)
		}
	}
	if _, err := os.Stat(string(alvo)); !os.IsNotExist(err) {
		t.Fatal("nada deveria ter sido escrito em sandbox")
	}
}

func TestJuntarCaminhos(t *testing.T) {
	res, err := met_arq_juntar(nil, hrp.Tupla{hrp.Texto("a"), hrp.Texto("b"), hrp.Texto("c.hrp")})
	if err != nil || res != hrp.Texto(filepath.Join("a", "b", "c.hrp")) {
		t.Fatalf("juntar = %v (%v)", res, err)
	}
}

func fmtLista(o hrp.Objeto) string {
	l, ok := o.(*hrp.Lista)
	if !ok {
		return ""
	}
	var partes []string
	for _, it := range l.Itens {
		partes = append(partes, string(it.(hrp.Texto)))
	}
	return strings.Join(partes, "\n")
}
