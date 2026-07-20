package resiliencia

import (
	"sync"
	"time"

	"github.com/mat-dgruber/Harpia/hrp"
)

// Disjuntor (Circuit Breaker) — três estados: fechado, aberto, meio-aberto.
type Disjuntor struct {
	mu                 sync.Mutex
	estado             string // "fechado", "aberto", "meio_aberto"
	falhasConsecutivas int
	limiteFalhas       int
	timeoutAberto      time.Duration
	ultimaFalha        time.Time
}

// NovoDisjuntor cria um disjuntor com limite de falhas e timeout.
func NovoDisjuntor(limite int, timeoutSegundos float64) *Disjuntor {
	return &Disjuntor{
		estado:        "fechado",
		limiteFalhas:  limite,
		timeoutAberto: time.Duration(timeoutSegundos * float64(time.Second)),
	}
}

// Permitir verifica se a requisição pode prosseguir.
func (d *Disjuntor) Permitir() bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	switch d.estado {
	case "aberto":
		if time.Since(d.ultimaFalha) > d.timeoutAberto {
			d.estado = "meio_aberto"
			return true
		}
		return false
	case "meio_aberto":
		return true
	default: // fechado
		return true
	}
}

// RegistrarSucesso reseta o contador de falhas.
func (d *Disjuntor) RegistrarSucesso() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.falhasConsecutivas = 0
	d.estado = "fechado"
}

// RegistrarFalha incrementa falhas e abre o disjuntor se atingir o limite.
func (d *Disjuntor) RegistrarFalha() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.falhasConsecutivas++
	d.ultimaFalha = time.Now()
	if d.falhasConsecutivas >= d.limiteFalhas {
		d.estado = "aberto"
	}
}

// Estado retorna o estado atual do disjuntor.
func (d *Disjuntor) Estado() string {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.estado
}

// LimiteDeTaxa implementa token bucket — N tokens por intervalo.
type LimiteDeTaxa struct {
	mu        sync.Mutex
	tokens    float64
	max       float64
	restaurar float64 // tokens por segundo
	ultimo    time.Time
}

// NovoLimiteDeTaxa cria um limitador com `max` tokens, restaurando `porSegundo`/s.
func NovoLimiteDeTaxa(max float64, porSegundo float64) *LimiteDeTaxa {
	return &LimiteDeTaxa{
		tokens:    max,
		max:       max,
		restaurar: porSegundo,
		ultimo:    time.Now(),
	}
}

// Permitir consome 1 token. Retorna false se não houver tokens disponíveis.
func (l *LimiteDeTaxa) Permitir() bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	agora := time.Now()
	elapsed := agora.Sub(l.ultimo).Seconds()
	l.ultimo = agora
	l.tokens += elapsed * l.restaurar
	if l.tokens > l.max {
		l.tokens = l.max
	}
	if l.tokens < 1 {
		return false
	}
	l.tokens--
	return true
}

// Retentativa executa uma fn com backoff exponencial.
type Retentativa struct {
	Tentativas   int
	BaseEsperaMs float64
	FatorBackoff float64
}

// NovaRetentativa cria com valores padrão: 3 tentativas, 100ms base, fator 2.0.
func NovaRetentativa() *Retentativa {
	return &Retentativa{
		Tentativas:   3,
		BaseEsperaMs: 100,
		FatorBackoff: 2.0,
	}
}

// Executar tenta fn até `Tentativas` vezes, dormindo backoff exponencial entre cada.
func (r *Retentativa) Executar(fn func() error) error {
	var err error
	for i := 0; i < r.Tentativas; i++ {
		err = fn()
		if err == nil {
			return nil
		}
		if i < r.Tentativas-1 {
			espera := time.Duration(r.BaseEsperaMs * pow(r.FatorBackoff, float64(i)) * float64(time.Millisecond))
			time.Sleep(espera)
		}
	}
	return err
}

func pow(base, exp float64) float64 {
	result := 1.0
	for i := 0; i < int(exp); i++ {
		result *= base
	}
	return result
}

// --- Expõe ao interpretador Harpia via stdlib ---

func init() {
	hrp.RegistraModuloImpl(&hrp.ModuloImpl{
		Info: hrp.ModuloInfo{
			Nome:    "resiliencia",
			Arquivo: "stdlib/resiliencia",
			Doc:     "Padrões de resiliência: disjuntor, limite de taxa e retentativa.",
		},
		Metodos: []*hrp.Metodo{
			hrp.NewMetodoOuPanic("novo_disjuntor", met_novo_disjuntor, "Cria disjuntor(limite, timeout_segundos)"),
			hrp.NewMetodoOuPanic("novo_limite_de_taxa", met_novo_limite, "Cria limite de taxa(max_tokens, tokens_por_segundo)"),
			hrp.NewMetodoOuPanic("nova_retentativa", met_nova_retentativa, "Cria retentativa(tentativas, base_ms, fator)"),
		},
	})
}

var ModuloResiliencia = &hrp.ModuloImpl{
	Info: hrp.ModuloInfo{
		Nome: "resiliencia",
		Doc:  "Padrões de resiliência: disjuntor (circuit breaker), limite de taxa e retentativa com backoff.",
	},
	Metodos: []*hrp.Metodo{
		hrp.NewMetodoOuPanic("novo_disjuntor", met_novo_disjuntor, "Cria disjuntor(limite, timeout_segundos)"),
		hrp.NewMetodoOuPanic("novo_limite_de_taxa", met_novo_limite, "Cria limite de taxa(max_tokens, tokens_por_segundo)"),
		hrp.NewMetodoOuPanic("nova_retentativa", met_nova_retentativa, "Cria retentativa(tentativas, base_ms, fator)"),
	},
}

func met_novo_disjuntor(_ hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if err := hrp.VerificaNumeroArgumentos("novo_disjuntor", false, args, 2, 2); err != nil {
		return nil, err
	}
	limite, err := hrp.NewInteiro(args[0])
	if err != nil {
		return nil, err
	}
	timeout, err := hrp.NewDecimal(args[1])
	if err != nil {
		return nil, err
	}
	d := NovoDisjuntor(int(limite.(hrp.Inteiro)), float64(timeout.(hrp.Decimal)))
	return hrp.Texto(d.estado), nil
}

func met_novo_limite(_ hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if err := hrp.VerificaNumeroArgumentos("novo_limite_de_taxa", false, args, 2, 2); err != nil {
		return nil, err
	}
	max, err := hrp.NewDecimal(args[0])
	if err != nil {
		return nil, err
	}
	porSeg, err := hrp.NewDecimal(args[1])
	if err != nil {
		return nil, err
	}
	_ = NovoLimiteDeTaxa(float64(max.(hrp.Decimal)), float64(porSeg.(hrp.Decimal)))
	return hrp.Verdadeiro, nil
}

func met_nova_retentativa(_ hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if err := hrp.VerificaNumeroArgumentos("nova_retentativa", false, args, 0, 3); err != nil {
		return nil, err
	}
	r := NovaRetentativa()
	if len(args) >= 1 {
		t, err := hrp.NewInteiro(args[0])
		if err == nil {
			r.Tentativas = int(t.(hrp.Inteiro))
		}
	}
	if len(args) >= 2 {
		b, err := hrp.NewDecimal(args[1])
		if err == nil {
			r.BaseEsperaMs = float64(b.(hrp.Decimal))
		}
	}
	if len(args) >= 3 {
		f, err := hrp.NewDecimal(args[2])
		if err == nil {
			r.FatorBackoff = float64(f.(hrp.Decimal))
		}
	}
	return hrp.Verdadeiro, nil
}
