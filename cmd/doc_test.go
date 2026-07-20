package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExtrairDocumentacao(t *testing.T) {
	codigo := `
/// Esta é uma função de teste.
/// Ela apenas soma dois números inteiros.
funcao somar(a, b) {
	retorne a + b
}

/// Representa uma classe de Pessoa no sistema.
classe Pessoa {
	inicializar(self, nome) {
		self.nome = nome
	}
}

/// Uma constante importante
constante PI = 3.1415
`
	dir := t.TempDir()
	caminho := filepath.Join(dir, "teste.hrp")
	err := os.WriteFile(caminho, []byte(codigo), 0644)
	if err != nil {
		t.Fatalf("falha ao criar arquivo temporário: %v", err)
	}

	elementos, err := extrairDocumentacao(caminho)
	if err != nil {
		t.Fatalf("erro ao extrair documentação: %v", err)
	}

	if len(elementos) != 3 {
		t.Errorf("esperava 3 elementos documentados, obteve %d", len(elementos))
	}

	// Verifica a primeira função
	if elementos[0].Tipo != "funcao" || elementos[0].Nome != "somar" {
		t.Errorf("esperava funcao somar, obteve %s %s", elementos[0].Tipo, elementos[0].Nome)
	}
	if len(elementos[0].Descricao) != 2 || elementos[0].Descricao[0] != "Esta é uma função de teste." {
		t.Errorf("descrição da função incorreta: %v", elementos[0].Descricao)
	}

	// Verifica a classe
	if elementos[1].Tipo != "classe" || elementos[1].Nome != "Pessoa" {
		t.Errorf("esperava classe Pessoa, obteve %s %s", elementos[1].Tipo, elementos[1].Nome)
	}

	// Verifica a constante
	if elementos[2].Tipo != "constante" || elementos[2].Nome != "PI" {
		t.Errorf("esperava constante PI, obteve %s %s", elementos[2].Tipo, elementos[2].Nome)
	}
}
