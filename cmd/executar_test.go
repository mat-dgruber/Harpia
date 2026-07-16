package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCmdExecutarComPerfil(t *testing.T) {
	codigo := `
	var i = 0;
	enquanto (i < 10) {
		i = i + 1;
	}
	`
	dir := t.TempDir()
	caminho := filepath.Join(dir, "teste_perfil.ptst")
	err := os.WriteFile(caminho, []byte(codigo), 0644)
	if err != nil {
		t.Fatalf("Erro ao criar arquivo temporário: %v", err)
	}

	// Captura stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := comandoExecutar()
	cmd.SetArgs([]string{caminho, "--vm", "--perfil"})
	err = cmd.Execute()

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("Erro ao rodar comando executar: %v", err)
	}

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "RELATÓRIO DE PERFILAMENTO DA VM") {
		t.Errorf("Esperava relatório de perfilamento na saída, obtive:\n%s", output)
	}
	if !strings.Contains(output, "OP_CARREGAR_VAR") {
		t.Errorf("Esperava encontrar opcodes catalogados no relatório, obtive:\n%s", output)
	}
}
