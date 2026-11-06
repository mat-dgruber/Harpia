//go:build !js && !wasm

package cmd

import (
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

// O modo --assistir deve reiniciar a execução a cada alteração e nunca deixar duas rodando juntas.
func TestAssistirArquivoReiniciaSemSobreposicao(t *testing.T) {
	alvo := filepath.Join(t.TempDir(), "app.hrp")
	if err := os.WriteFile(alvo, []byte("imprimir(1)"), 0600); err != nil {
		t.Fatal(err)
	}

	var ativas, maxAtivas, inicios atomic.Int32
	iniciar := func() func() {
		inicios.Add(1)
		if n := ativas.Add(1); n > maxAtivas.Load() {
			maxAtivas.Store(n)
		}
		return func() { ativas.Add(-1) }
	}

	fim := make(chan struct{})
	terminou := make(chan error, 1)
	go func() { terminou <- assistirArquivo(alvo, 20*time.Millisecond, iniciar, fim) }()

	time.Sleep(60 * time.Millisecond)
	for i := 1; i <= 2; i++ {
		futuro := time.Now().Add(time.Duration(i) * time.Second)
		if err := os.Chtimes(alvo, futuro, futuro); err != nil {
			t.Fatal(err)
		}
		time.Sleep(100 * time.Millisecond)
	}
	close(fim)
	if err := <-terminou; err != nil {
		t.Fatal(err)
	}

	if got := inicios.Load(); got != 3 {
		t.Fatalf("esperava 3 inícios (1 + 2 alterações), veio %d", got)
	}
	if got := maxAtivas.Load(); got != 1 {
		t.Fatalf("execuções sobrepostas: máximo simultâneo = %d", got)
	}
	if got := ativas.Load(); got != 0 {
		t.Fatalf("execução ficou ativa depois do fim: %d", got)
	}
}

func TestArgsFilhoAssistirRemoveFlag(t *testing.T) {
	got := argsFilhoAssistir([]string{"executar", "-w", "app.hrp", "--assistir", "--silencioso", "--assistir=true"})
	want := []string{"executar", "app.hrp", "--silencioso"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}
