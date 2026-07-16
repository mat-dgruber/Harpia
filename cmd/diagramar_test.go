package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestAnalisarDependenciasViolacao assevera a detecção de violações da Clean Architecture
func TestAnalisarDependenciasViolacao(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "portuscript_diag_*")
	if err != nil {
		t.Fatalf("Erro ao criar diretório temporário: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Cria estrutura: meu_app/dominio/ e meu_app/infra/
	domDir := filepath.Join(tempDir, "dominio")
	infDir := filepath.Join(tempDir, "infra")
	os.MkdirAll(domDir, 0755)
	os.MkdirAll(infDir, 0755)

	// 1. Cria infra/banco.ptst
	err = os.WriteFile(filepath.Join(infDir, "banco.ptst"), []byte(`
	var banco = "sqlite"
	`), 0644)
	if err != nil {
		t.Fatalf("Erro ao criar banco.ptst: %v", err)
	}

	// 2. Cria dominio/usuario.ptst importando incorretamente de infra/ (VIOLAÇÃO!)
	err = os.WriteFile(filepath.Join(domDir, "usuario.ptst"), []byte(`
	de "../infra/banco.ptst" importe banco;
	var usuario = "Natan"
	`), 0644)
	if err != nil {
		t.Fatalf("Erro ao criar usuario.ptst: %v", err)
	}

	// Executa análise estática de dependências
	rels, violacoes := analisarDependencias(tempDir)

	// 1. Deve encontrar 1 relação dominio -> infra
	if len(rels) != 1 {
		t.Errorf("Esperava 1 relação mapeada, obtive %d", len(rels))
	} else {
		if rels[0].De != "dominio" || rels[0].Para != "infra" {
			t.Errorf("Esperava relação de 'dominio' para 'infra', obtive de '%s' para '%s'", rels[0].De, rels[0].Para)
		}
	}

	// 2. Deve detectar a violação arquitetural
	if len(violacoes) != 1 {
		t.Errorf("Esperava 1 violação de dependência detectada, obtive %d", len(violacoes))
	} else {
		if !strings.Contains(violacoes[0], "camada 'dominio' em") || !strings.Contains(violacoes[0], "importando de 'infra'") {
			t.Errorf("Mensagem de violação inesperada: %s", violacoes[0])
		}
	}

	// 3. Testa gerador de Mermaid
	mermaid := gerarMermaid(rels)
	if !strings.Contains(mermaid, "dominio --> infra") {
		t.Errorf("Código Mermaid gerado esperado não encontrado. Obtido:\n%s", mermaid)
	}
}
