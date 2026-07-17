package resiliencia

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestDisjuntorAbreAposLimiteFalhas(t *testing.T) {
	d := NovoDisjuntor(3, 1.0)
	if d.Estado() != "fechado" {
		t.Fatal("deveria começar fechado")
	}
	for i := 0; i < 3; i++ {
		d.RegistrarFalha()
	}
	if d.Estado() != "aberto" {
		t.Fatalf("deveria estar aberto, got %s", d.Estado())
	}
	if d.Permitir() {
		t.Fatal("disjuntor aberto não deveria permitir")
	}
}

func TestDisjuntorRecuperaAposTimeout(t *testing.T) {
	d := NovoDisjuntor(2, 0.1) // 100ms timeout
	d.RegistrarFalha()
	d.RegistrarFalha()
	if d.Estado() != "aberto" {
		t.Fatal("deveria estar aberto")
	}
	time.Sleep(150 * time.Millisecond)
	if !d.Permitir() {
		t.Fatal("deveria permitir após timeout (meio-aberto)")
	}
	if d.Estado() != "meio_aberto" {
		t.Fatalf("deveria estar meio_aberto, got %s", d.Estado())
	}
}

func TestDisjuntorFechaAposSucesso(t *testing.T) {
	d := NovoDisjuntor(2, 0.1)
	d.RegistrarFalha()
	d.RegistrarSucesso()
	if d.Estado() != "fechado" {
		t.Fatalf("deveria estar fechado, got %s", d.Estado())
	}
}

func TestLimiteDeTaxa(t *testing.T) {
	l := NovoLimiteDeTaxa(5, 100) // 5 tokens, 100/s
	for i := 0; i < 5; i++ {
		if !l.Permitir() {
			t.Fatalf("deveria permitir na %dª chamada", i+1)
		}
	}
	if l.Permitir() {
		t.Fatal("deveria bloquear após esgotar tokens")
	}
}

func TestRetentativaSucessoNaPrimeira(t *testing.T) {
	r := NovaRetentativa()
	var chamadas int32
	err := r.Executar(func() error {
		atomic.AddInt32(&chamadas, 1)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if atomic.LoadInt32(&chamadas) != 1 {
		t.Fatal("deveria chamar apenas 1 vez")
	}
}

func TestRetentativaSucessoNaTerceira(t *testing.T) {
	r := &Retentativa{Tentativas: 3, BaseEsperaMs: 1, FatorBackoff: 1.0}
	var chamadas int32
	err := r.Executar(func() error {
		n := atomic.AddInt32(&chamadas, 1)
		if n < 3 {
			return &dummyErr{}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if atomic.LoadInt32(&chamadas) != 3 {
		t.Fatal("deveria chamar 3 vezes")
	}
}

func TestRetentativaEsgotaTentativas(t *testing.T) {
	r := &Retentativa{Tentativas: 2, BaseEsperaMs: 1, FatorBackoff: 1.0}
	var chamadas int32
	err := r.Executar(func() error {
		atomic.AddInt32(&chamadas, 1)
		return &dummyErr{}
	})
	if err == nil {
		t.Fatal("deveria retornar erro")
	}
	if atomic.LoadInt32(&chamadas) != 2 {
		t.Fatal("deveria chamar 2 vezes")
	}
}

type dummyErr struct{}

func (e *dummyErr) Error() string { return "dummy" }
