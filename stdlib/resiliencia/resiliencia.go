// Package resiliencia fornece padrões de programação defensiva e tolerância a falhas para microsserviços,
// incluindo Circuit Breaker (Disjuntor), Token Bucket Rate Limiting (Limite de Taxa) e Retry com Backoff Exponencial.
package resiliencia

import (
	"sync"
	"time"

	"github.com/mat-dgruber/Harpia/hrp"
)

const (
	// MaxLimiteDisjuntor estabelece teto máximo para falhas consecutivas aceitas no disjuntor.
	MaxLimiteDisjuntor = 1000000

	// MaxTentativasRetentativa limita o volume máximo de retries para evitar loops infinitos ou exaustão de threads.
	MaxTentativasRetentativa = 100
)

// Disjuntor implementa o padrão Circuit Breaker que isola integrações instáveis de terceiros ou falhas em APIs,
// protegendo a integridade geral do sistema através de uma máquina de estados finita: fechado, aberto, meio_aberto.
type Disjuntor struct {
	mu                 sync.Mutex
	estado             string // "fechado", "aberto", "meio_aberto"
	falhasConsecutivas int
	limiteFalhas       int
	timeoutAberto      time.Duration
	ultimaFalha        time.Time
}

// NovoDisjuntor constrói e configura uma nova instância da máquina de estados do Circuit Breaker.
func NovoDisjuntor(limite int, timeoutSegundos float64) *Disjuntor {
	return &Disjuntor{
		estado:        "fechado",
		limiteFalhas:  limite,
		timeoutAberto: time.Duration(timeoutSegundos * float64(time.Second)),
	}
}

// Permitir determina de forma atômica se uma chamada pode prosseguir ou deve falhar instantaneamente (Fast-Fail)
// dependendo do estado em que a máquina se encontra.
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

// RegistrarSucesso reseta o contador de falhas acumulado e fecha o disjuntor.
func (d *Disjuntor) RegistrarSucesso() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.falhasConsecutivas = 0
	d.estado = "fechado"
}

// RegistrarFalha incrementa o contador e abre o disjuntor caso o teto seja atingido, mudando para estado aberto.
func (d *Disjuntor) RegistrarFalha() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.falhasConsecutivas++
	d.ultimaFalha = time.Now()
	if d.falhasConsecutivas >= d.limiteFalhas {
		d.estado = "aberto"
	}
}

// Estado retorna com segurança o estado atual do disjuntor.
func (d *Disjuntor) Estado() string {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.estado
}

// LimiteDeTaxa implementa o algoritmo Token Bucket para controle de fluxo e mitigação de picos de tráfego.
type LimiteDeTaxa struct {
	mu        sync.Mutex
	tokens    float64
	max       float64
	restaurar float64 // tokens por segundo
	ultimo    time.Time
}

// NovoLimiteDeTaxa cria um limitador com capacidade máxima e velocidade de restauração de tokens/s.
func NovoLimiteDeTaxa(max float64, porSegundo float64) *LimiteDeTaxa {
	return &LimiteDeTaxa{
		tokens:    max,
		max:       max,
		restaurar: porSegundo,
		ultimo:    time.Now(),
	}
}

// Permitir tenta consumir de forma síncrona um token de execução, atualizando a recarga baseada no delta de tempo.
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

// Retentativa controla o fluxo de re-execução de rotinas com espaçamento inteligente e progressivo (Backoff).
type Retentativa struct {
	Tentativas   int
	BaseEsperaMs float64
	FatorBackoff float64
}

// NovaRetentativa cria um configurador de retries com valores de mercado recomendados (3 tentativas).
func NovaRetentativa() *Retentativa {
	return &Retentativa{
		Tentativas:   3,
		BaseEsperaMs: 100,
		FatorBackoff: 2.0,
	}
}

// Executar repete uma rotina que gera erro, aguardando um tempo exponencialmente maior entre falhas consecutivas.
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

func init() {
	// Registra o módulo 'resiliencia' no ecossistema central do Harpia.
	hrp.RegistraModuloImpl(&hrp.ModuloImpl{
		Info: hrp.ModuloInfo{
			Nome:    "resiliencia",
			Arquivo: "stdlib/resiliencia",
			Doc:     "Padrões de resiliência corporativa: disjuntor (Circuit Breaker), limite de taxa e retentativa com backoff.",
		},
		Metodos: []*hrp.Metodo{
			hrp.NewMetodoOuPanic("novo_disjuntor", met_novo_disjuntor, "Cria disjuntor(limite, timeout_segundos) para mitigar falhas."),
			hrp.NewMetodoOuPanic("novo_limite_de_taxa", met_novo_limite, "Cria limite de taxa(max_tokens, tokens_por_segundo) para mitigação de tráfego excessivo."),
			hrp.NewMetodoOuPanic("nova_retentativa", met_nova_retentativa, "Cria retentativa(tentativas, base_ms, fator) com backoff exponencial."),
		},
	})
}

// met_novo_disjuntor implementa 'novo_disjuntor(limite, timeout)' em nível de script Harpia.
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
	limVal := int64(limite.(hrp.Inteiro))
	if limVal < 0 || limVal > MaxLimiteDisjuntor {
		return nil, hrp.NewErroF(hrp.ValorErro, "limite de taxa inválido (deve ser entre 0 e %d)", MaxLimiteDisjuntor)
	}
	d := NovoDisjuntor(int(limVal), float64(timeout.(hrp.Decimal)))
	return hrp.Texto(d.estado), nil
}

// met_novo_limite implementa 'novo_limite_de_taxa(max, porSegundo)' em nível de script Harpia.
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

// met_nova_retentativa implementa 'nova_retentativa(tentativas?, base?, fator?)' em nível de script Harpia.
func met_nova_retentativa(_ hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if err := hrp.VerificaNumeroArgumentos("nova_retentativa", false, args, 0, 3); err != nil {
		return nil, err
	}
	r := NovaRetentativa()
	if len(args) >= 1 {
		t, err := hrp.NewInteiro(args[0])
		if err != nil {
			return nil, err
		}
		tVal := int64(t.(hrp.Inteiro))
		if tVal < 0 || tVal > MaxTentativasRetentativa {
			return nil, hrp.NewErroF(hrp.ValorErro, "número de tentativas inválido (deve ser entre 0 e %d)", MaxTentativasRetentativa)
		}
		r.Tentativas = int(tVal)
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
