package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestAnalisarDependenciasViolacao assevera a detecção de violações da Clean Architecture
func TestAnalisarDependenciasViolacao(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "Harpia_diag_*")
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

	// 4. Testar colorização das conexões irregulares de linkStyle
	if !strings.Contains(mermaid, "linkStyle 0 stroke:#ff3333,stroke-width:3px;") {
		t.Errorf("Esperava estilo de linkStyle de violação em vermelho. Obtido:\n%s", mermaid)
	}
}

// TestDiagramarHTMLExport assevera que a exportação de HTML interativo e de alertas está correta
func TestDiagramarHTMLExport(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "Harpia_diag_html_*")
	if err != nil {
		t.Fatalf("Erro ao criar diretório temporário: %v", err)
	}
	defer os.RemoveAll(tempDir)

	saidaHtml := filepath.Join(tempDir, "diagrama_teste.html")

	// Prepara dados de teste
	rels := []ImportRel{
		{De: "dominio", Para: "infra", Arquivo: "dominio/usuario.ptst"},
	}
	violacoes := []string{
		"camada 'dominio' importando 'infra'",
	}

	codigoMermaid := gerarMermaid(rels)
	alertas := gerarAlertasHTML(violacoes)

	htmlFinal := strings.Replace(templateHTMLDiagrama, "{{MERMAID_CODE}}", codigoMermaid, 1)
	htmlFinal = strings.Replace(htmlFinal, "{{ALERTS_MARKUP}}", alertas, 1)

	err = os.WriteFile(saidaHtml, []byte(htmlFinal), 0644)
	if err != nil {
		t.Fatalf("Erro ao gravar arquivo HTML de teste: %v", err)
	}

	// Valida se o arquivo de fato existe e contém os componentes-chave
	conteudoBytes, err := os.ReadFile(saidaHtml)
	if err != nil {
		t.Fatalf("Erro ao ler arquivo gravado: %v", err)
	}
	conteudo := string(conteudoBytes)

	if !strings.Contains(conteudo, "Diagrama de Arquitetura do Harpia") {
		t.Errorf("Esperava cabeçalho do template no HTML. Obtido:\n%s", conteudo)
	}

	if !strings.Contains(conteudo, "dominio --> infra") {
		t.Errorf("Esperava o diagrama Mermaid embutido no HTML. Obtido:\n%s", conteudo)
	}

	if !strings.Contains(conteudo, "camada 'dominio' importando 'infra'") {
		t.Errorf("Esperava a listagem de alertas e violações injetada no HTML. Obtido:\n%s", conteudo)
	}

	if !strings.Contains(conteudo, "window.baixarSVG") {
		t.Errorf("Esperava a função de exportação para SVG interativo. Obtido:\n%s", conteudo)
	}
}

