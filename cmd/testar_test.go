package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

// TestComandoTestarComHTML assevera o funcionamento da geração do relatório estético de cobertura em HTML
func TestComandoTestarComHTML(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "portuscript_test_html_*")
	if err != nil {
		t.Fatalf("Erro ao criar diretório temporário: %v", err)
	}
	defer os.RemoveAll(tempDir)

	codigo := `
	var a = 1;
	testar "teste simples" {
		a = 2;
	}
	`

	testFile := filepath.Join(tempDir, "soma_test.ptst")
	err = os.WriteFile(testFile, []byte(codigo), 0644)
	if err != nil {
		t.Fatalf("Erro ao escrever arquivo de teste: %v", err)
	}

	oldWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(oldWd)

	cmd := comandoTestar()
	cmd.SetArgs([]string{"soma_test.ptst", "--html"})

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("Erro ao executar comando de testar com html: %v", err)
	}

	// Verifica se gerou o arquivo de cobertura cobertura.html
	if _, err := os.Stat("cobertura.html"); os.IsNotExist(err) {
		t.Errorf("Arquivo 'cobertura.html' esperado não foi gerado")
	}
}
