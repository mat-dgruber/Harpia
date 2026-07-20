package tests

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/mat-dgruber/Harpia/hrp"
	_ "github.com/mat-dgruber/Harpia/stdlib"
)

func TestRPCModulo(t *testing.T) {
	tempDir := t.TempDir()
	curWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(curWd)

	// Cria dependencias.json
	depConfig := map[string]string{
		"conectarBackend": "./meu-backend",
		"urlBackend":      "http://localhost:8085",
	}
	depBytes, _ := json.Marshal(depConfig)
	os.WriteFile("dependencias.json", depBytes, 0644)

	// Cria pasta meu-backend e usuarios.hrp
	os.Mkdir("meu-backend", 0755)
	os.WriteFile(filepath.Join("meu-backend", "usuarios.hrp"), []byte("exportar funcao obterUsuario(id) {}"), 0644)

	// Inicia um servidor HTTP em Go para simular o backend RPC
	mux := http.NewServeMux()
	mux.HandleFunc("/rpc/usuarios/obterUsuario", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var reqMap map[string]interface{}
		json.Unmarshal(body, &reqMap)

		resMap := map[string]interface{}{
			"retorno": "Ola ID " + reqMap["args"].([]interface{})[0].(string),
		}
		resBytes, _ := json.Marshal(resMap)
		w.Write(resBytes)
	})

	srv := &http.Server{
		Addr:    ":8085",
		Handler: mux,
	}

	go srv.ListenAndServe()
	defer srv.Shutdown(context.Background())

	ctx := hrp.NewContexto(hrp.OpcsContexto{})
	defer ctx.Terminar()

	codigo := `
	de "@backend/usuarios" importe obterUsuario

	var resposta = obterUsuario("42")
	`

	res, err := hrp.ExecutarString(ctx, codigo)
	if err != nil {
		t.Fatalf("Erro ao executar chamada RPC no Harpia: %v", err)
	}

	val, err := res.Escopo.ObterValor("resposta")
	if err != nil {
		t.Fatal(err)
	}

	if string(val.(hrp.Texto)) != "Ola ID 42" {
		t.Errorf("Retorno RPC incorreto, obteve '%v', esperava 'Ola ID 42'", val)
	}
}
