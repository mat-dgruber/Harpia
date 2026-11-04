package hrp

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
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
