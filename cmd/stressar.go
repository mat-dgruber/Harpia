//go:build !js && !wasm

package cmd

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/mat-dgruber/Harpia/hrp"
	"github.com/spf13/cobra"
)

// comandoStressar executa testes de carga concorrentes em um arquivo Harpia.
//
// Fluxo principal:
//  1. Lê o código fonte uma única vez e o armazena em memória (evita I/O repetido);
//  2. Dispara N goroutines limitadas por um semáforo buffered (`concorrencia`);
//  3. Mede duração de cada execução e agrega em um slice indexado para cálculo
//     de mínimo/máximo/média ao final.
//  4. Saídas de erros são impressas em stderr, mas não interrompem o benchmark;
//     apenas a falha de uma execução conta para a métrica de "sucessos".
//
// Importante: o interpretador compartilha `hrp.NewContexto` por iteração
// (descartado via `defer ctx.Terminar()`), garantindo que não há vazamento de
// memória entre execuções mesmo sob alta concorrência.
func comandoStressar() *cobra.Command {
	var arquivo string
	var concorrencia int
	var requisicoes int

	stressar := &cobra.Command{
		Use:           "stressar",
		Short:         "Executa testes de estresse concorrentes em um script Harpia",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			errOut := cmd.ErrOrStderr()

			if len(args) > 0 {
				arquivo = args[0]
			}

			if arquivo == "" {
				fmt.Fprintln(errOut, "erro: arquivo de entrada não especificado")
				return fmt.Errorf("arquivo de entrada não especificado")
			}

			fmt.Fprintf(out, "🔥 Iniciando teste de estresse em '%s'...\n", arquivo)
			fmt.Fprintf(out, "Concorrência: %d | Total de Execuções: %d\n\n", concorrencia, requisicoes)

			conteudo, err := os.ReadFile(arquivo)
			if err != nil {
				fmt.Fprintf(errOut, "erro ao ler arquivo: %v\n", err)
				return err
			}
			codigoStr := string(conteudo)

			var wg sync.WaitGroup
			sem := make(chan struct{}, concorrencia)
			tempos := make([]time.Duration, requisicoes)
			sucessos := 0
			var temposMu sync.Mutex

			inicioGeral := time.Now()

			for i := 0; i < requisicoes; i++ {
				select {
				case <-cmd.Context().Done():
					fmt.Fprintln(errOut, "⚠️ Teste de estresse cancelado.")
					return cmd.Context().Err()
				default:
				}

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
						fmt.Fprintf(errOut, "[Falha #%d]: %v\n", idx, err)
					}
					temposMu.Unlock()
				}(i)
			}

			wg.Wait()
			duracaoGeral := time.Since(inicioGeral)

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

			fmt.Fprintln(out, "\n📊 RELATÓRIO DO TESTE DE ESTRESSE")
			fmt.Fprintln(out, "=========================================================")
			fmt.Fprintf(out, "Tempo Total Geral:   %v\n", duracaoGeral)
			fmt.Fprintf(out, "Sucessos:            %d/%d (%.2f%%)\n", sucessos, requisicoes, float64(sucessos)/float64(requisicoes)*100)
			fmt.Fprintf(out, "Tempo Mínimo:        %v\n", minTempo)
			fmt.Fprintf(out, "Tempo Máximo:        %v\n", maxTempo)
			fmt.Fprintf(out, "Tempo Médio:         %v\n", mediaTempo)
			fmt.Fprintln(out, "=========================================================")
			return nil
		},
	}

	stressar.Flags().StringVarP(&arquivo, "arquivo", "a", "", "Caminho do arquivo Harpia.")
	stressar.Flags().IntVarP(&concorrencia, "concorrencia", "c", 10, "Quantidade de instâncias concorrentes.")
	stressar.Flags().IntVarP(&requisicoes, "requisicoes", "r", 100, "Número total de execuções de teste.")

	return stressar
}
