package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// ExplicacaoErro armazena metadados didáticos sobre os erros do compilador/VM do Harpia.
type ExplicacaoErro struct {
	Nome      string
	Descricao string
	Exemplo   string
}

// explicacoes centraliza o catálogo estático de códigos de erros do Harpia (PSC-xxxx)
// mapeando-os para explicações em português e exemplos de correção para os desenvolvedores.
var explicacoes = map[string]ExplicacaoErro{
	"PSC-0001": {
		Nome:      "SintaxeErro",
		Descricao: "O código fonte possui um erro gramatical ou de escrita que o Parser não conseguiu entender.",
		Exemplo:   "Código incorreto: se 1 == 2 { imprima(\"Erro\") # falta fechar aspas ou parênteses\nCódigo correto: se 1 == 2 { imprima(\"Erro\") }",
	},
	"PSC-0002": {
		Nome:      "ReatribuicaoErro",
		Descricao: "Você tentou redeclarar ou alterar o valor de uma constante imutável criada com a palavra-chave 'const'.",
		Exemplo:   "const X = 10\nX = 20 # Gera ReatribuicaoErro",
	},
	"PSC-0003": {
		Nome:      "AtributoErro",
		Descricao: "Ocorreu uma tentativa de acessar um atributo, método ou propriedade inexistente em um objeto ou módulo.",
		Exemplo:   "var texto = \"Olá\"\ntexto.atributo_inexistente # Gera AtributoErro",
	},
	"PSC-0004": {
		Nome:      "TipagemErro",
		Descricao: "Foi tentada uma operação com tipos incompatíveis (ex: somar texto com número sem conversão) ou número inválido de argumentos em funções.",
		Exemplo:   "var soma = \"texto\" + 10 # Gera TipagemErro",
	},
	"PSC-0005": {
		Nome:      "NomeErro",
		Descricao: "Uma variável, constante ou função foi usada no código antes de ser declarada ou definida.",
		Exemplo:   "imprima(x) # Gera NomeErro se 'x' não foi declarado com 'var x' ou 'const x'",
	},
	"PSC-0006": {
		Nome:      "ImportacaoErro",
		Descricao: "Ocorreu uma falha ao tentar importar um módulo físico ou símbolo de outro módulo, ou há uma importação cíclica/loop de dependências.",
		Exemplo:   "de \"./modulo_inexistente.hrp\" importe soma # Gera ImportacaoErro",
	},
	"PSC-0007": {
		Nome:      "ValorErro",
		Descricao: "Um valor fornecido a uma operação ou função possui o tipo correto, mas o valor em si não é apropriado para a operação.",
		Exemplo:   "Inteiro(\"texto_nao_numerico\") # Gera ValorErro ao tentar converter para número",
	},
	"PSC-0008": {
		Nome:      "ErroDeLimite",
		Descricao: "O valor fornecido está fora do intervalo ou limite permitido para a operação na máquina virtual.",
		Exemplo:   "Tentar acessar posições além dos limites permitidos de memória.",
	},
	"PSC-0009": {
		Nome:      "IndiceErro",
		Descricao: "Ocorreu uma tentativa de acesso a um índice que está fora dos limites de tamanho de uma sequência (como Lista, Tupla ou Texto).",
		Exemplo:   "var lista = [1, 2]\nlista[5] # Gera IndiceErro",
	},
	"PSC-0010": {
		Nome:      "RuntimeErro",
		Descricao: "Ocorreu uma falha genérica ou erro inesperado no ambiente de execução da máquina virtual.",
		Exemplo:   "Erros internos no interpretador ou comportamento indefinido.",
	},
	"PSC-0011": {
		Nome:      "ErroDeAsseguracao",
		Descricao: "Uma asserção lógica feita com a instrução 'assegura' falhou, pois resultou em Falso.",
		Exemplo:   "assegura 1 == 2, \"Um nunca é igual a dois\" # Gera ErroDeAsseguracao",
	},
	"PSC-0012": {
		Nome:      "DivisaoPorZeroErro",
		Descricao: "Ocorreu uma tentativa matemática proibida de realizar divisão de um número real, inteiro ou resto (módulo) por zero.",
		Exemplo:   "var x = 10 / 0 # Gera DivisaoPorZeroErro",
	},
	"PSC-0013": {
		Nome:      "ErroDeSistema",
		Descricao: "Ocorreu uma falha associada a chamadas do sistema operacional, como leitura de arquivos ou entrada e saída de dados.",
		Exemplo:   "Tentar ler arquivos sem as permissões adequadas no sistema.",
	},
	"PSC-0014": {
		Nome:      "ArquivoNaoEncontradoErro",
		Descricao: "O interpretador não conseguiu encontrar o arquivo especificado no caminho fornecido.",
		Exemplo:   "de \"./arquivo_que_nao_existe.hrp\" importe modulo",
	},
}

// comandoErroCLI cria e retorna o comando Cobra 'erro' (`harpia erro`).
// Este comando disponibiliza explicações estáticas sobre os códigos de erros do Harpia e
// possui um subcomando 'explicar' que conecta-se a uma instância do Ollama local
// (utilizando o modelo 'gemma') para gerar uma explicação pedagógica dinâmica com IA.
func comandoErroCLI() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "erro [codigo]",
		Short: "Fornece explicações didáticas em português sobre os códigos de erros do Harpia",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				fmt.Println("Lista de Códigos de Erros do Harpia:")
				fmt.Println("================================================================================")
				// Print sorted error codes or simply loop over the map
				for code, exp := range explicacoes {
					fmt.Printf("  %-10s | %-20s | %s\n", code, exp.Nome, exp.Descricao)
				}
				fmt.Println("================================================================================")
				fmt.Println("Use 'harpia erro [codigo]' para ver detalhes e exemplos (ex: 'harpia erro PSC-0005')")
				return
			}
			codigo := strings.ToUpper(args[0])
			explicacao, encontrada := explicacoes[codigo]

			if !encontrada {
				fmt.Printf("Código de erro '%s' não encontrado. Use um código como 'PSC-0005'.\n", codigo)
				return
			}

			fmt.Printf("================================================================================\n")
			fmt.Printf("Código: %s (%s)\n", codigo, explicacao.Nome)
			fmt.Printf("================================================================================\n\n")
			fmt.Printf("Descrição:\n  %s\n\n", explicacao.Descricao)
			fmt.Printf("Exemplo:\n  %s\n", strings.ReplaceAll(explicacao.Exemplo, "\n", "\n  "))
			fmt.Printf("================================================================================\n")
		},
	}

	explicarCmd := &cobra.Command{
		Use:   "explicar [codigo]",
		Short: "Usa IA Local (Ollama) para gerar uma explicação pedagógica personalizada sobre o erro",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			codigo := strings.ToUpper(args[0])
			explicacao, encontrada := explicacoes[codigo]

			if !encontrada {
				fmt.Printf("Código de erro '%s' não reconhecido.\n", codigo)
				return
			}

			fmt.Printf("Solicitando explicação inteligente com IA para o erro %s (%s)...\n", codigo, explicacao.Nome)

			requestBody, _ := json.Marshal(map[string]interface{}{
				"model":  "gemma",
				"prompt": fmt.Sprintf("Explique o código de erro '%s' (%s) do compilador Harpia de forma extremamente pedagógica, em português brasileiro, fornecendo um exemplo curto de código com erro e a respectiva correção.", codigo, explicacao.Nome),
				"stream": false,
			})

			client := &http.Client{Timeout: 3 * time.Second}
			resp, err := client.Post("http://127.0.0.1:11434/api/generate", "application/json", bytes.NewBuffer(requestBody))

			if err != nil {
				fmt.Println("\n⚠️  Não foi possível conectar ao Ollama local (127.0.0.1:11434).")
				fmt.Println("Para usar IA local do Harpia, siga estes passos:")
				fmt.Println("  1. Baixe e instale o Ollama em: https://ollama.com")
				fmt.Println("  2. Instale o modelo 'gemma' via terminal:")
				fmt.Println("     ollama run gemma")
				fmt.Println("  3. Certifique-se de que o Ollama está rodando em segundo plano.")
				fmt.Println("\nExibindo explicação estática local como alternativa:")

				fmt.Printf("\n================================================================================\n")
				fmt.Printf("Código: %s (%s)\n", codigo, explicacao.Nome)
				fmt.Printf("================================================================================\n\n")
				fmt.Printf("Descrição:\n  %s\n\n", explicacao.Descricao)
				fmt.Printf("Exemplo:\n  %s\n", strings.ReplaceAll(explicacao.Exemplo, "\n", "\n  "))
				fmt.Printf("================================================================================\n")
				return
			}
			defer resp.Body.Close()

			var result struct {
				Response string `json:"response"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				fmt.Println("Erro ao ler resposta do Ollama local. Verifique se o modelo 'gemma' está pronto.")
				return
			}

			fmt.Println("\n================================================================================")
			fmt.Printf("Explicação IA para %s (%s):\n", codigo, explicacao.Nome)
			fmt.Println("================================================================================")
			fmt.Println(result.Response)
			fmt.Println("\n================================================================================")
		},
	}

	cmd.AddCommand(explicarCmd)
	return cmd
}
