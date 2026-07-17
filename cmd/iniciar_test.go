package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

// TestComandoNovoMonolito assevera que o scaffold do tipo monolito com Clean Architecture é gerado corretamente
func TestComandoNovoMonolito(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "Harpia_mono_*")
	if err != nil {
		t.Fatalf("Erro ao criar diretório temporário: %v", err)
	}
	defer os.RemoveAll(tempDir)

	oldWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(oldWd)

	cmd := comandoNovo()
	cmd.SetArgs([]string{"monolito", "meu_monolito"})

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("Erro ao executar scaffolding do monolito: %v", err)
	}

	diretoriosEsperados := []string{
		"meu_monolito",
		filepath.Join("meu_monolito", "dominio", "entidades"),
		filepath.Join("meu_monolito", "dominio", "repositorios"),
		filepath.Join("meu_monolito", "infra", "banco"),
		filepath.Join("meu_monolito", "web", "rotas"),
		filepath.Join("meu_monolito", "web", "componentes"),
		filepath.Join("meu_monolito", "testes"),
	}

	for _, dir := range diretoriosEsperados {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Errorf("Diretório esperado de Clean Arch '%s' não foi gerado", dir)
		}
	}

	arquivosEsperados := []string{
		filepath.Join("meu_monolito", "main.ptst"),
		filepath.Join("meu_monolito", "dominio", "entidades", "usuario.ptst"),
		filepath.Join("meu_monolito", "infra", "banco", "conexao.ptst"),
		filepath.Join("meu_monolito", "web", "rotas", "index.ptst"),
		filepath.Join("meu_monolito", "web", "componentes", "Layout.html"),
		filepath.Join("meu_monolito", "testes", "usuario_test.ptst"),
	}

	for _, arq := range arquivosEsperados {
		if _, err := os.Stat(arq); os.IsNotExist(err) {
			t.Errorf("Arquivo esperado de Clean Arch '%s' não foi gerado", arq)
		}
	}
}

// TestComandoNovoBackend assevera a geração física de projetos backend-only
func TestComandoNovoBackend(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "Harpia_back_*")
	if err != nil {
		t.Fatalf("Erro ao criar diretório temporário: %v", err)
	}
	defer os.RemoveAll(tempDir)

	oldWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(oldWd)

	cmd := comandoNovo()
	cmd.SetArgs([]string{"backend", "meu_back"})

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("Erro ao executar scaffolding de backend: %v", err)
	}

	arquivosEsperados := []string{
		filepath.Join("meu_back", "main.ptst"),
		filepath.Join("meu_back", "dominio", "entidades", "produto.ptst"),
	}

	for _, arq := range arquivosEsperados {
		if _, err := os.Stat(arq); os.IsNotExist(err) {
			t.Errorf("Arquivo de backend esperado '%s' não foi gerado", arq)
		}
	}
}

// TestComandoCrieAssistido assevera que geradores crie rota e crie componente inserem arquivos boilerplate
func TestComandoCrieAssistido(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "Harpia_crie_*")
	if err != nil {
		t.Fatalf("Erro ao criar diretório temporário: %v", err)
	}
	defer os.RemoveAll(tempDir)

	oldWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(oldWd)

	// Simula a existência da pasta web/rotas e web/componentes para teste de detecção inteligente
	os.MkdirAll("web/rotas", 0755)
	os.MkdirAll("web/componentes", 0755)

	// 1. Testa crie rota
	cmdRota := comandoCrie()
	cmdRota.SetArgs([]string{"rota", "contato"})
	if err := cmdRota.Execute(); err != nil {
		t.Fatalf("Erro ao executar crie rota: %v", err)
	}

	caminhoRota := filepath.Join("web", "rotas", "contato.ptst")
	if _, err := os.Stat(caminhoRota); os.IsNotExist(err) {
		t.Errorf("Arquivo de rota esperada '%s' não foi gerado", caminhoRota)
	}

	// 2. Testa crie componente
	cmdComp := comandoCrie()
	cmdComp.SetArgs([]string{"componente", "alerta"})
	if err := cmdComp.Execute(); err != nil {
		t.Fatalf("Erro ao executar crie componente: %v", err)
	}

	caminhoComp := filepath.Join("web", "componentes", "Alerta.ptst")
	caminhoEstilo := filepath.Join("web", "componentes", "Alerta.estilo.ptst")

	if _, err := os.Stat(caminhoComp); os.IsNotExist(err) {
		t.Errorf("Arquivo de componente esperado '%s' não foi gerado", caminhoComp)
	}
	if _, err := os.Stat(caminhoEstilo); os.IsNotExist(err) {
		t.Errorf("Arquivo de estilo esperado '%s' não foi gerado", caminhoEstilo)
	}
}
