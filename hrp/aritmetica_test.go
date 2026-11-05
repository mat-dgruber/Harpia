package hrp

import (
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestAritmeticaDecimal(t *testing.T) {
	t.Run("Subtrai", func(t *testing.T) {
		res, err := Subtrai(Decimal(10), Decimal(3))
		if err != nil {
			t.Fatalf("Erro inesperado em Subtrai: %v", err)
		}
		if dec, ok := res.(Decimal); !ok || dec != Decimal(7) {
			t.Errorf("Subtrai(10, 3) = %v; esperado Decimal(7)", res)
		}
	})

	t.Run("Divide", func(t *testing.T) {
		res, err := Divide(Decimal(10), Decimal(4))
		if err != nil {
			t.Fatalf("Erro inesperado em Divide: %v", err)
		}
		if dec, ok := res.(Decimal); !ok || dec != Decimal(2.5) {
			t.Errorf("Divide(10, 4) = %v; esperado Decimal(2.5)", res)
		}
	})

	t.Run("Mod", func(t *testing.T) {
		res, err := Decimal(5.5).M__mod__(Decimal(2))
		if err != nil {
			t.Fatalf("Erro inesperado em Mod: %v", err)
		}
		if dec, ok := res.(Decimal); !ok || dec != Decimal(1.5) {
			t.Errorf("Decimal(5.5).M__mod__(Decimal(2)) = %v; esperado Decimal(1.5)", res)
		}

		_, errZero := Decimal(5.5).M__mod__(Decimal(0))
		if errZero == nil {
			t.Errorf("Decimal(5.5).M__mod__(Decimal(0)) esperava erro de divisão por zero, mas retornou nil")
		}
	})
}

func TestRPC(t *testing.T) {
	t.Run("RejeitaEsquemaInvalido", func(t *testing.T) {
		tempDir := t.TempDir()
		curWd, _ := os.Getwd()
		_ = os.Chdir(tempDir)
		defer os.Chdir(curWd)

		for _, esquemaInvalido := range []string{"ftp://localhost:8080", "javascript:alert(1)"} {
			depConfig := map[string]string{
				"urlBackend": esquemaInvalido,
			}
			depBytes, _ := json.Marshal(depConfig)
			_ = os.WriteFile("dependencias.json", depBytes, 0644)

			ctx := NewContexto(OpcsContexto{})
			_, err := CarregarModuloRPC(ctx, "@backend/usuarios")
			if err == nil {
				t.Errorf("Esperava erro para URL inválida '%s', mas obteve nil", esquemaInvalido)
			}
		}
	})

	t.Run("ErroStatusHttp400Mais", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "Erro interno de teste", http.StatusInternalServerError)
		}))
		defer srv.Close()

		tempDir := t.TempDir()
		curWd, _ := os.Getwd()
		_ = os.Chdir(tempDir)
		defer os.Chdir(curWd)

		depConfig := map[string]string{
			"urlBackend": srv.URL,
		}
		depBytes, _ := json.Marshal(depConfig)
		_ = os.WriteFile("dependencias.json", depBytes, 0644)

		ctx := NewContexto(OpcsContexto{})
		modObj, err := CarregarModuloRPC(ctx, "@backend/usuarios")
		if err != nil {
			t.Fatalf("Erro ao carregar módulo RPC: %v", err)
		}

		mod := modObj.(*Modulo)
		metodoObj, errAttr := mod.M__obtem_attributo__("obterUsuario")
		if errAttr != nil {
			t.Fatalf("Método obterUsuario não encontrado no módulo RPC: %v", errAttr)
		}

		metodo := metodoObj.(*Metodo)
		_, errCall := metodo.Chamar(nil, Tupla{Texto("1")})
		if errCall == nil {
			t.Fatalf("Esperava erro ao chamar método com status >= 400, mas obteve nil")
		}
	})

	t.Run("ErroJsonInvalido", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("não-é-um-json-válido"))
		}))
		defer srv.Close()

		tempDir := t.TempDir()
		curWd, _ := os.Getwd()
		_ = os.Chdir(tempDir)
		defer os.Chdir(curWd)

		depConfig := map[string]string{
			"urlBackend": srv.URL,
		}
		depBytes, _ := json.Marshal(depConfig)
		_ = os.WriteFile("dependencias.json", depBytes, 0644)

		ctx := NewContexto(OpcsContexto{})
		modObj, err := CarregarModuloRPC(ctx, "@backend/usuarios")
		if err != nil {
			t.Fatalf("Erro ao carregar módulo RPC: %v", err)
		}

		mod := modObj.(*Modulo)
		metodoObj, errAttr := mod.M__obtem_attributo__("obterUsuario")
		if errAttr != nil {
			t.Fatalf("Método obterUsuario não encontrado no módulo RPC: %v", errAttr)
		}

		metodo := metodoObj.(*Metodo)
		_, errCall := metodo.Chamar(nil, Tupla{Texto("1")})
		if errCall == nil {
			t.Fatalf("Esperava erro ao chamar método com JSON inválido, mas obteve nil")
		}
	})
}

func TestPotencia(t *testing.T) {
	casos := []struct {
		a, b Objeto
		want Objeto
	}{
		{Inteiro(2), Inteiro(3), Inteiro(8)},
		{Inteiro(2), Inteiro(0), Inteiro(1)},
		{Inteiro(-3), Inteiro(3), Inteiro(-27)},
		{Inteiro(2), Inteiro(62), Inteiro(1 << 62)},
		{Inteiro(2), Inteiro(-1), Decimal(0.5)},
		{Decimal(2.25), Decimal(0.5), Decimal(1.5)},
	}
	for _, c := range casos {
		got, err := Potencia(c.a, c.b)
		if err != nil || got != c.want {
			t.Errorf("%v ** %v = %v (%v), esperava %v", c.a, c.b, got, err, c.want)
		}
	}
	for _, c := range [][2]Objeto{{Inteiro(2), Inteiro(63)}, {Inteiro(10), Inteiro(19)}} {
		if _, err := Potencia(c[0], c[1]); err == nil {
			t.Errorf("%v ** %v deveria lançar ErroDeLimite", c[0], c[1])
		}
	}
	if _, err := Potencia(Decimal(-8), Decimal(0.5)); err == nil {
		t.Error("-8 ** 0.5 deveria lançar ValorErro")
	}
	if _, err := Potencia(Texto("a"), Inteiro(2)); err == nil {
		t.Error("'a' ** 2 deveria lançar TipagemErro")
	}
}

func TestOverflowInteiroLancaErroDeLimite(t *testing.T) {
	const max, min = Inteiro(math.MaxInt64), Inteiro(math.MinInt64)
	estouros := map[string]func() (Objeto, error){
		"max+1":   func() (Objeto, error) { return Adiciona(max, Inteiro(1)) },
		"min-1":   func() (Objeto, error) { return Subtrai(min, Inteiro(1)) },
		"max*2":   func() (Objeto, error) { return Multiplica(max, Inteiro(2)) },
		"min*-1":  func() (Objeto, error) { return Multiplica(min, Inteiro(-1)) },
		"-min":    func() (Objeto, error) { return min.M__neg__() },
		"min//-1": func() (Objeto, error) { return min.M__divide_inteiro__(Inteiro(-1)) },
	}
	for nome, op := range estouros {
		if _, err := op(); err == nil || !strings.Contains(err.Error(), "limite") {
			t.Errorf("%s deveria lançar ErroDeLimite, veio %v", nome, err)
		}
	}
	normais := map[string]struct {
		op   func() (Objeto, error)
		want Objeto
	}{
		"max-1+1": {func() (Objeto, error) { return Adiciona(max-1, Inteiro(1)) }, max},
		"min+1-1": {func() (Objeto, error) { return Subtrai(min+1, Inteiro(1)) }, min},
		"-3*4":    {func() (Objeto, error) { return Multiplica(Inteiro(-3), Inteiro(4)) }, Inteiro(-12)},
		"0*min":   {func() (Objeto, error) { return Multiplica(Inteiro(0), min) }, Inteiro(0)},
		"-max":    {func() (Objeto, error) { return max.M__neg__() }, -max},
		"7//-2":   {func() (Objeto, error) { return Inteiro(7).M__divide_inteiro__(Inteiro(-2)) }, Inteiro(-3)},
	}
	for nome, c := range normais {
		got, err := c.op()
		if err != nil || got != c.want {
			t.Errorf("%s = %v (%v), esperava %v", nome, got, err, c.want)
		}
	}
}

func BenchmarkAdicionaInteiro(b *testing.B) {
	for b.Loop() {
		_, _ = Adiciona(Inteiro(12345), Inteiro(678))
	}
}

// M31: Decimal não tinha <, <=, >, >= e a igualdade entre Inteiro e Decimal era assimétrica
// (1 == 1.0 dava Falso, 1.0 == 1 dava Verdadeiro).
func TestComparacoesNumericasMistas(t *testing.T) {
	type op func(a, b Objeto) (Objeto, error)
	casos := []struct {
		nome string
		f    op
		a, b Objeto
		want Booleano
	}{
		{"0.5<1.0", MenorQue, Decimal(0.5), Decimal(1.0), true},
		{"1<2.5", MenorQue, Inteiro(1), Decimal(2.5), true},
		{"2.5>1", MaiorQue, Decimal(2.5), Inteiro(1), true},
		{"2.0>=2.0", MaiorOuIgual, Decimal(2.0), Decimal(2.0), true},
		{"2<=1.5", MenorOuIgual, Inteiro(2), Decimal(1.5), false},
		{"1==1.0", Igual, Inteiro(1), Decimal(1.0), true},
		{"1.0==1", Igual, Decimal(1.0), Inteiro(1), true},
		{"1!=1.0", Diferente, Inteiro(1), Decimal(1.0), false},
		{"3<4", MenorQue, Inteiro(3), Inteiro(4), true},
		{"maxint", MenorQue, Inteiro(math.MaxInt64 - 1), Inteiro(math.MaxInt64), true},
	}
	for _, c := range casos {
		got, err := c.f(c.a, c.b)
		if err != nil || got != c.want {
			t.Errorf("%s = %v (%v), esperava %v", c.nome, got, err, c.want)
		}
	}
	if _, err := MenorQue(Texto("a"), Inteiro(1)); err == nil {
		t.Error("'a' < 1 deveria continuar sendo erro de tipo")
	}
}
