package cmd

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mat-dgruber/Harpia/ptst"
	"github.com/spf13/cobra"
)

type tuiModel struct {
	editor     textarea.Model
	saida      string
	variaveis  []string
	largura    int
	altura     int
	ajuda      bool
	ctx        *ptst.Contexto
	escopo     *ptst.Escopo
	foco       int // ponytail: 0 = editor, 1 = inspetor
	depurando  bool // ponytail: se o modo de depuração passo-a-passo está ativo
	linhas     []string
	linhaAtiva int
}

// comandoTui registra o comando 'harpia tui' para inicializar a interface Bubbletea didática
func comandoTui() *cobra.Command {
	return &cobra.Command{
		Use:   "tui",
		Short: "Inicia a TUI Interativa e REPL didático do Harpia no terminal",
		Run: func(cmd *cobra.Command, args []string) {
			p := tea.NewProgram(inicializarTuiModel(), tea.WithAltScreen())
			if _, err := p.Run(); err != nil {
				fmt.Printf("Erro ao rodar TUI: %v\n", err)
				os.Exit(1)
			}
		},
	}
}

func inicializarTuiModel() tuiModel {
	ta := textarea.New()
	ta.Placeholder = "# Escreva seu código Harpia aqui...\n# Pressione F2 para executar ou F8 para depurar!"
	ta.Focus()
	ta.SetWidth(40)
	ta.SetHeight(20)

	ctx := ptst.NewContexto(ptst.OpcsContexto{})
	escopo := ptst.NewEscopo()

	return tuiModel{
		editor:    ta,
		ctx:       ctx,
		escopo:    escopo,
		saida:     "Console ativo. Digite o código na esquerda e pressione F2 para executar ou F8 para depurar!\n",
		variaveis: []string{"Nenhuma variável declarada ainda."},
	}
}

func (m tuiModel) Init() tea.Cmd {
	return textarea.Blink
}

func (m tuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.ctx.Terminar()
			return m, tea.Quit
		case "f1":
			m.ajuda = !m.ajuda
			return m, nil
		case "f2", "ctrl+e":
			m.depurando = false // cancela depuração se rodar tudo direto
			m.executarCodigoTui()
			return m, nil
		case "f8":
			// ponytail: liga/desliga o modo de depuração passo-a-passo
			if m.depurando {
				m.depurando = false
				m.saida = "Depuração cancelada pelo usuário.\n"
			} else {
				codigo := m.editor.Value()
				if strings.TrimSpace(codigo) != "" {
					m.depurando = true
					m.linhas = strings.Split(codigo, "\n")
					m.linhaAtiva = 0
					m.escopo = ptst.NewEscopo() // reinicia o escopo local
					m.saida = "=== MODO DEPURAÇÃO ATIVO ===\nPressione F7 para avançar linha por linha.\nPressione F8 para cancelar.\n"
					m.variaveis = []string{"Nenhuma variável ainda."}
				}
			}
			return m, nil
		case "f7":
			// ponytail: avança um passo no depurador síncrono
			if m.depurando {
				m.avancarPassoDepurador()
			}
			return m, nil
		case "tab":
			// ponytail: alterna foco interativo entre editor e inspetor/logs
			if m.foco == 0 {
				m.foco = 1
				m.editor.Blur()
			} else {
				m.foco = 0
				m.editor.Focus()
			}
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.largura = msg.Width
		m.altura = msg.Height
		m.editor.SetWidth(msg.Width/2 - 4)
		m.editor.SetHeight(msg.Height - 6)
	}

	if m.foco == 0 {
		m.editor, cmd = m.editor.Update(msg)
	}
	return m, cmd
}

func (m *tuiModel) avancarPassoDepurador() {
	if m.linhaAtiva >= len(m.linhas) {
		m.depurando = false
		m.saida += "\n=== FIM DA DEPURAÇÃO ===\n"
		return
	}

	// Ignora linhas puramente vazias
	for m.linhaAtiva < len(m.linhas) && strings.TrimSpace(m.linhas[m.linhaAtiva]) == "" {
		m.linhaAtiva++
	}

	if m.linhaAtiva >= len(m.linhas) {
		m.depurando = false
		m.saida += "\n=== FIM DA DEPURAÇÃO ===\n"
		return
	}

	linhaAtual := m.linhas[m.linhaAtiva]
	m.saida += fmt.Sprintf("\n[Linha %d] -> %s\n", m.linhaAtiva+1, strings.TrimSpace(linhaAtual))

	// ponytail: simulação síncrona re-avaliando as linhas acumuladas de forma limpa
	linhasExecutadas := m.linhas[0 : m.linhaAtiva+1]
	codigoAcumulado := strings.Join(linhasExecutadas, "\n")

	// Captura stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	m.escopo = ptst.NewEscopo() // Limpa escopo local para re-avaliar o acumulado
	ast, err := m.ctx.StringParaAst(codigoAcumulado, "<tui-debugger>")
	if err != nil {
		w.Close()
		os.Stdout = oldStdout
		m.saida += fmt.Sprintf("Erro de Sintaxe na linha %d: %v\n", m.linhaAtiva+1, err)
		m.depurando = false
		return
	}

	_, errExec := m.ctx.AvaliarAst(ast, m.escopo)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	saidaTerminal := buf.String()
	if saidaTerminal != "" {
		m.saida += saidaTerminal
	}

	if errExec != nil {
		m.saida += fmt.Sprintf("Erro de Execução: %v\n", errExec)
		m.depurando = false
		return
	}

	// Atualiza o inspetor de variáveis
	simbolos := m.escopo.ObterSimbolosSeguro()
	var vars []string
	for _, s := range simbolos {
		if s != nil {
			if strings.HasPrefix(s.Nome, "_") {
				continue
			}
			tipo := s.Tipo
			if tipo == "" {
				tipo = "Dinamico"
			}
			vars = append(vars, fmt.Sprintf("%s (%s) = %v", s.Nome, tipo, s.ObterValor()))
		}
	}
	if len(vars) == 0 {
		m.variaveis = []string{"Nenhuma variável local no escopo."}
	} else {
		m.variaveis = vars
	}

	m.linhaAtiva++
}

func (m *tuiModel) executarCodigoTui() {
	codigo := m.editor.Value()
	if strings.TrimSpace(codigo) == "" {
		return
	}

	// Captura stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	ast, err := m.ctx.StringParaAst(codigo, "<tui>")
	if err != nil {
		w.Close()
		os.Stdout = oldStdout
		m.saida = fmt.Sprintf("Erro de Sintaxe:\n%v\n", err)
		m.variaveis = []string{"-"}
		return
	}

	_, errExec := m.ctx.AvaliarAst(ast, m.escopo)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	m.saida = buf.String()

	if errExec != nil {
		m.saida += fmt.Sprintf("\nErro de Execução:\n%v\n", errExec)
	}

	// Coleta variáveis locais no escopo do REPL de forma síncrona
	simbolos := m.escopo.ObterSimbolosSeguro()
	var vars []string
	for _, s := range simbolos {
		if s != nil {
			if strings.HasPrefix(s.Nome, "_") {
				continue
			}
			tipo := s.Tipo
			if tipo == "" {
				tipo = "Dinamico"
			}
			vars = append(vars, fmt.Sprintf("%s (%s) = %v", s.Nome, tipo, s.ObterValor()))
		}
	}
	if len(vars) == 0 {
		m.variaveis = []string{"Nenhuma variável local no escopo."}
	} else {
		m.variaveis = vars
	}
}

func (m tuiModel) View() string {
	if m.ajuda {
		return lipgloss.NewStyle().
			Width(m.largura).
			Height(m.altura).
			Align(lipgloss.Center, lipgloss.Center).
			Render("=== AJUDA DO PLAYGROUND INTERATIVO (TUI) ===\n\n" +
				"F2 / Ctrl+E ➔ Executar o código de uma vez\n" +
				"F8          ➔ Iniciar / Cancelar modo de depuração passo-a-passo\n" +
				"F7          ➔ Avançar uma linha no modo de depuração\n" +
				"F1          ➔ Alternar esta tela de ajuda\n" +
				"Tab         ➔ Alternar foco entre o editor e o inspetor\n" +
				"Ctrl+C      ➔ Sair do REPL interativo de terminal\n\n" +
				"Pressione F1 para voltar ao console.")
	}

	// ponytail: destaca visualmente o painel em foco usando cores do Lipgloss
	corBordaEditor := "240" // cinza atenuado
	corBordaInspetor := "240"
	if m.foco == 0 {
		corBordaEditor = "62" // roxo destacado
	} else {
		corBordaInspetor = "33" // azul destacado
	}

	estiloEditor := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(corBordaEditor)).
		Padding(1).
		Width(m.largura/2 - 2).
		Height(m.altura - 4)

	estiloLado := lipgloss.NewStyle().
		Width(m.largura/2 - 2).
		Height(m.altura - 4)

	estiloModulo := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(corBordaInspetor)).
		Padding(1).
		Height(m.altura/2 - 2)

	estiloConsole := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(1).
		Height(m.altura/2 - 2)

	conteudoLadoDireito := lipgloss.JoinVertical(lipgloss.Left,
		estiloModulo.Render("=== INSPETOR DE ESCOPO / VARIÁVEIS ===\n\n"+strings.Join(m.variaveis, "\n")),
		estiloConsole.Render("=== SAÍDA DO CONSOLE / LOGS ===\n\n"+m.saida),
	)

	corpo := lipgloss.JoinHorizontal(lipgloss.Top,
		estiloEditor.Render("=== EDITOR DE CÓDIGO (HARPIA) ===\n\n"+m.editor.View()),
		estiloLado.Render(conteudoLadoDireito),
	)

	rodape := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render(" F1: Ajuda | F2: Executar | F7: Passo | F8: Depurar | Tab: Foco | Ctrl+C: Sair ")

	return lipgloss.JoinVertical(lipgloss.Left, corpo, rodape)
}
