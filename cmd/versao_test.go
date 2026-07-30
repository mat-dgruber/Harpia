package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestComandoVersaoJSON(t *testing.T) {
	cmd := comandoVersao()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--formato=json"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("erro ao executar comando: %v", err)
	}

	var meta map[string]string
	if err := json.Unmarshal(buf.Bytes(), &meta); err != nil {
		t.Fatalf("erro ao decodificar JSON: %v. Saida: %s", err, buf.String())
	}

	if meta["versao"] == "" {
		t.Errorf("campo versao ausente no JSON")
	}
}

func TestComandoVersaoTexto(t *testing.T) {
	cmd := comandoVersao()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("erro ao executar comando: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Harpia CLI") {
		t.Errorf("saida inesperada: %s", out)
	}
}

func TestComandoVersaoCurto(t *testing.T) {
	cmd := comandoVersao()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--curto"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("erro ao executar comando: %v", err)
	}

	out := strings.TrimSpace(buf.String())
	if out != Versao {
		t.Errorf("esperado %s, obtido %s", Versao, out)
	}
}
