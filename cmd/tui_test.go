package cmd

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// TestInicializarTuiModel assevera o comportamento síncrono e correto de inicialização da TUI Bubbletea
func TestInicializarTuiModel(t *testing.T) {
	m := inicializarTuiModel()

	// 1. Deve possuir contexto e escopo alocados
	if m.ctx == nil {
		t.Errorf("Esperava contexto alocado")
	}
	if m.escopo == nil {
		t.Errorf("Esperava escopo alocado")
	}

	// 2. O editor textarea deve estar focado
	if !m.editor.Focused() {
		t.Errorf("Esperava editor de texto focado para digitação instantânea")
	}

	// 3. Os painéis devem possuir suas mensagens didáticas padrões
	if !strings.Contains(m.saida, "Console ativo") {
		t.Errorf("Saída do console inicial esperada não encontrada")
	}
	if len(m.variaveis) == 0 {
		t.Errorf("Painel de variáveis esperado não deve estar vazio")
	}
}

// TestTUIFocoInterativo assevera a alternância lógica de foco usando a tecla Tab
func TestTUIFocoInterativo(t *testing.T) {
	m := inicializarTuiModel()

	// 1. Inicia com o editor focado
	if m.foco != 0 {
		t.Errorf("Esperava foco inicial no editor (0), obtive %d", m.foco)
	}
	if !m.editor.Focused() {
		t.Errorf("Esperava que o editor iniciasse focado")
	}

	// 2. Simula o pressionamento da tecla 'tab' para alternar foco
	novaModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab, Runes: []rune{}, Alt: false})
	m = novaModel.(tuiModel)

	if m.foco != 1 {
		t.Errorf("Esperava que o foco mudasse para o inspetor (1) após Tab, obtive %d", m.foco)
	}
	if m.editor.Focused() {
		t.Errorf("Esperava que o editor perdesse o foco após Tab")
	}

	// 3. Simula novamente 'tab' para voltar ao editor
	novaModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab, Runes: []rune{}, Alt: false})
	m = novaModel.(tuiModel)

	if m.foco != 0 {
		t.Errorf("Esperava que o foco retornasse ao editor (0), obtive %d", m.foco)
	}
	if !m.editor.Focused() {
		t.Errorf("Esperava que o editor re-ganhasse o foco")
	}
}

// TestTUIDebuggerSimulation assevera que o fluxo de depuração síncrono e teclas de controle funcionam na TUI
func TestTUIDebuggerSimulation(t *testing.T) {
	model := inicializarTuiModel()

	// 1. Simula escrita de código de teste
	model.editor.SetValue("var a = 10;\nvar b = 20;\nimprimir(a + b);")

	// 2. Simula tecla F8 para iniciar a depuração
	msgF8 := tea.KeyMsg{Type: tea.KeyF8, Runes: []rune{}, Alt: false}
	updatedModel, _ := model.Update(msgF8)
	m := updatedModel.(tuiModel)

	if !m.depurando {
		t.Fatalf("Esperava que o modo de depuração estivesse ativo após pressionar F8")
	}

	if len(m.linhas) != 3 {
		t.Errorf("Esperava 3 linhas carregadas para depurar, obtive %d", len(m.linhas))
	}

	// 3. Simula tecla F7 para avançar o primeiro passo (executa var a = 10;)
	msgF7 := tea.KeyMsg{Type: tea.KeyF7, Runes: []rune{}, Alt: false}
	updatedModel, _ = m.Update(msgF7)
	m = updatedModel.(tuiModel)

	if m.linhaAtiva != 1 {
		t.Errorf("Esperava avançar para o passo/linha 1, obtive %d", m.linhaAtiva)
	}

	// 4. Executa todos os passos até o fim da depuração
	updatedModel, _ = m.Update(msgF7)
	m = updatedModel.(tuiModel)
	updatedModel, _ = m.Update(msgF7)
	m = updatedModel.(tuiModel)

	// Avança o passo extra para terminar
	updatedModel, _ = m.Update(msgF7)
	m = updatedModel.(tuiModel)

	if m.depurando {
		t.Errorf("Esperava que a depuração estivesse concluída após esgotar as linhas de código")
	}

	if !strings.Contains(m.saida, "FIM DA DEPURAÇÃO") {
		t.Errorf("Esperava log de fim de depuração na saída do console")
	}
}
