package cmd

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/mat-dgruber/Harpia/hrp"
	"github.com/spf13/cobra"
)

func comandoStressar() *cobra.Command {
	var arquivo string
	var concorrencia int
	var requisicoes int

	stressar := &cobra.Command{
		Use:   "stressar",
		Short: "Executa testes de estresse concorrentes em um script Harpia",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) > 0 {
				arquivo = args[0]
			}

			if arquivo == "" {
				fmt.Fprintln(os.Stderr, "erro: arquivo de entrada não especificado")
				os.Exit(1)
			}

			fmt.Printf("🔥 Iniciando teste de estresse em '%s'...\n", arquivo)
			fmt.Printf("Concorrência: %d | Total de Execuções: %d\n\n", concorrencia, requisicoes)

			// Leitura prévia para carregar o código na memória antes do benchmark
			conteudo, err := os.ReadFile(arquivo)
			if err != nil {
				fmt.Fprintf(os.Stderr, "erro ao ler arquivo: %v\n", err)
				os.Exit(1)
			}
			codigoStr := string(conteudo)

			var wg sync.WaitGroup
			sem := make(chan struct{}, concorrencia)
			tempos := make([]time.Duration, requisicoes)
			sucessos := 0
			var temposMu sync.Mutex

			inicioGeral := time.Now()

			for i := 0; i < requisicoes; i++ {
				wg.Add(1)
				sem <- struct{}{}

				go func(idx int) {
					defer func() {
						<-sem
						wg.Done()
					}()

					ctx := hrp.NewContexto(hrp.OpcsContexto{})
					defer ctx.Terminar()

					inicioExec := time.Now()
					_, err := hrp.ExecutarString(ctx, codigoStr)
					duracao := time.Since(inicioExec)

					temposMu.Lock()
					tempos[idx] = duracao
					if err == nil {
						sucessos++
					} else {
						fmt.Printf("[Falha #%d]: %v\n", idx, err)
					}
					temposMu.Unlock()
				}(i)
			}

			wg.Wait()
			duracaoGeral := time.Since(inicioGeral)

			// Estatísticas
			var totalTempo time.Duration
			var minTempo = time.Hour
			var maxTempo time.Duration

			for _, t := range tempos {
				totalTempo += t
				if t < minTempo {
					minTempo = t
				}
				if t > maxTempo {
					maxTempo = t
				}
			}

			mediaTempo := totalTempo / time.Duration(requisicoes)

			fmt.Println("\n📊 RELATÓRIO DO TESTE DE ESTRESSE")
			fmt.Println("=========================================================")
			fmt.Printf("Tempo Total Geral:   %v\n", duracaoGeral)
			fmt.Printf("Sucessos:            %d/%d (%.2f%%)\n", sucessos, requisicoes, float64(sucessos)/float64(requisicoes)*100)
			fmt.Printf("Tempo Mínimo:        %v\n", minTempo)
			fmt.Printf("Tempo Máximo:        %v\n", maxTempo)
			fmt.Printf("Tempo Médio:         %v\n", mediaTempo)
			fmt.Println("=========================================================")
		},
	}

	stressar.Flags().StringVarP(&arquivo, "arquivo", "a", "", "Caminho do arquivo Harpia.")
	stressar.Flags().IntVarP(&concorrencia, "concorrencia", "c", 10, "Quantidade de instâncias concorrentes.")
	stressar.Flags().IntVarP(&requisicoes, "requisicoes", "r", 100, "Número total de execuções de teste.")

	return stressar
}
