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
		filepath.Join("meu_monolito", "main.hrp"),
		filepath.Join("meu_monolito", "dominio", "entidades", "usuario.hrp"),
		filepath.Join("meu_monolito", "infra", "banco", "conexao.hrp"),
		filepath.Join("meu_monolito", "web", "rotas", "rotas.hrp"),
		filepath.Join("meu_monolito", "web", "global.estilos.hrp"),
		filepath.Join("meu_monolito", "web", "pages", "Inicio.hrp"),
		filepath.Join("meu_monolito", "web", "pages", "Inicio.estilo.hrp"),
		filepath.Join("meu_monolito", "web", "pages", "Inicio.html"),
		filepath.Join("meu_monolito", "testes", "usuario_test.hrp"),
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
		filepath.Join("meu_back", "main.hrp"),
		filepath.Join("meu_back", "dominio", "entidades", "produto.hrp"),
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

	// Simula a existência da pasta web/rotas, web/componentes e web/pages para teste de detecção inteligente
	os.MkdirAll("web/rotas", 0755)
	os.MkdirAll("web/componentes", 0755)
	os.MkdirAll("web/pages", 0755)

	// 1. Testa crie rota
	cmdRota := comandoCrie()
	cmdRota.SetArgs([]string{"rota", "contato"})
	if err := cmdRota.Execute(); err != nil {
		t.Fatalf("Erro ao executar crie rota: %v", err)
	}

	caminhosEsperados := []string{
		filepath.Join("web", "pages", "Contato.hrp"),
		filepath.Join("web", "pages", "Contato.estilo.hrp"),
		filepath.Join("web", "pages", "Contato.html"),
	}
	for _, arq := range caminhosEsperados {
		if _, err := os.Stat(arq); os.IsNotExist(err) {
			t.Errorf("Arquivo de rota esperada '%s' não foi gerado", arq)
		}
	}

	// 2. Testa crie componente
	cmdComp := comandoCrie()
	cmdComp.SetArgs([]string{"componente", "alerta"})
	if err := cmdComp.Execute(); err != nil {
		t.Fatalf("Erro ao executar crie componente: %v", err)
	}

	caminhoComp := filepath.Join("web", "componentes", "Alerta.hrp")
	caminhoEstilo := filepath.Join("web", "componentes", "Alerta.estilo.hrp")

	if _, err := os.Stat(caminhoComp); os.IsNotExist(err) {
		t.Errorf("Arquivo de componente esperado '%s' não foi gerado", caminhoComp)
	}
	if _, err := os.Stat(caminhoEstilo); os.IsNotExist(err) {
		t.Errorf("Arquivo de estilo esperado '%s' não foi gerado", caminhoEstilo)
	}
}
