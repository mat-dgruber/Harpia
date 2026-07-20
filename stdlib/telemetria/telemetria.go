package telemetria

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/mat-dgruber/Harpia/hrp"
)

// Representação interna de Span do OpenTelemetry em formato JSON leve
type Span struct {
	TraceID   string    `json:"trace_id"`
	SpanID    string    `json:"span_id"`
	Nome      string    `json:"name"`
	Servico   string    `json:"service"`
	Inicio    time.Time `json:"start_time"`
	Fim       time.Time `json:"end_time"`
	DuracaoMs int64     `json:"duration_ms"`
	Status    string    `json:"status"`
}

var TipoSpan = hrp.TipoObjeto.NewTipo("Span", "Span de rastreamento de OpenTelemetry")

func (s *Span) Tipo() *hrp.Tipo {
	return TipoSpan
}

func (s *Span) M__obtem_attributo__(nome string) (hrp.Objeto, error) {
	switch nome {
	case "finalizar":
		return hrp.NewMetodoOuPanic("finalizar", func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
			status := "OK"
			if len(args) >= 1 {
				st, _ := hrp.NewTexto(args[0])
				status = string(st.(hrp.Texto))
			}
			s.Fim = time.Now()
			s.DuracaoMs = s.Fim.Sub(s.Inicio).Milliseconds()
			s.Status = status

			// Exporta o Trace/Span em JSON estruturado para Stdout
			bytes, _ := json.Marshal(s)
			fmt.Fprintln(os.Stdout, string(bytes))
			return hrp.Nulo, nil
		}, ""), nil
	}
	return nil, hrp.NewErroF(hrp.AtributoErro, "Atributo '%s' não existe em Span", nome)
}

type Tracer struct {
	Servico string
}

var TipoTracer = hrp.TipoObjeto.NewTipo("Tracer", "Rastreador de traces de OpenTelemetry")

func (t *Tracer) Tipo() *hrp.Tipo {
	return TipoTracer
}

func (t *Tracer) M__obtem_attributo__(nome string) (hrp.Objeto, error) {
	switch nome {
	case "iniciar_span":
		return hrp.NewMetodoOuPanic("iniciar_span", func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
			if err := hrp.VerificaNumeroArgumentos("iniciar_span", false, args, 1, 1); err != nil {
				return nil, err
			}
			nomeSpan, err := hrp.NewTexto(args[0])
			if err != nil {
				return nil, err
			}

			now := time.Now().UnixNano()
			traceID := fmt.Sprintf("%x", now)
			spanID := fmt.Sprintf("%x", now&0xffff)

			return &Span{
				TraceID: traceID,
				SpanID:  spanID,
				Nome:    string(nomeSpan.(hrp.Texto)),
				Servico: t.Servico,
				Inicio:  time.Now(),
			}, nil
		}, ""), nil
	}
	return nil, hrp.NewErroF(hrp.AtributoErro, "Atributo '%s' não existe em Tracer", nome)
}

type Metrica struct {
	Nome    string
	Kind    string
	Valores map[string]float64
	mu      sync.Mutex
}

var TipoMetrica = hrp.TipoObjeto.NewTipo("Metrica", "Métrica quantitativa leve")

func (m *Metrica) Tipo() *hrp.Tipo {
	return TipoMetrica
}

func (m *Metrica) M__obtem_attributo__(nome string) (hrp.Objeto, error) {
	switch nome {
	case "registrar":
		return hrp.NewMetodoOuPanic("registrar", func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
			if err := hrp.VerificaNumeroArgumentos("registrar", false, args, 1, 2); err != nil {
				return nil, err
			}
			val, err := hrp.NewDecimal(args[0])
			if err != nil {
				return nil, err
			}
			tag := "default"
			if len(args) == 2 {
				t, _ := hrp.NewTexto(args[1])
				tag = string(t.(hrp.Texto))
			}

			m.mu.Lock()
			m.Valores[tag] += float64(val.(hrp.Decimal))
			m.mu.Unlock()

			// Emite métrica formatada
			fmt.Printf("{\"metric\": \"%s\", \"type\": \"%s\", \"tag\": \"%s\", \"value\": %f}\n", m.Nome, m.Kind, tag, val.(hrp.Decimal))
			return hrp.Nulo, nil
		}, ""), nil
	}
	return nil, hrp.NewErroF(hrp.AtributoErro, "Atributo '%s' não existe em Metrica", nome)
}

func met_novo_tracer(_ hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if err := hrp.VerificaNumeroArgumentos("novo_tracer", false, args, 1, 1); err != nil {
		return nil, err
	}
	servico, err := hrp.NewTexto(args[0])
	if err != nil {
		return nil, err
	}
	return &Tracer{Servico: string(servico.(hrp.Texto))}, nil
}

func met_nova_metrica(_ hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if err := hrp.VerificaNumeroArgumentos("nova_metrica", false, args, 2, 2); err != nil {
		return nil, err
	}
	nome, err := hrp.NewTexto(args[0])
	if err != nil {
		return nil, err
	}
	tipo, err := hrp.NewTexto(args[1])
	if err != nil {
		return nil, err
	}
	return &Metrica{
		Nome:    string(nome.(hrp.Texto)),
		Kind:    string(tipo.(hrp.Texto)),
		Valores: make(map[string]float64),
	}, nil
}

func init() {
	hrp.RegistraModuloImpl(&hrp.ModuloImpl{
		Info: hrp.ModuloInfo{
			Nome:    "telemetria",
			Arquivo: "stdlib/telemetria",
			Doc:     "Módulo leve de observabilidade compatível com OpenTelemetry (Traces e Métricas)",
		},
		Metodos: []*hrp.Metodo{
			hrp.NewMetodoOuPanic("novo_tracer", met_novo_tracer, "Cria um novo Tracer(servico)"),
			hrp.NewMetodoOuPanic("nova_metrica", met_nova_metrica, "Cria uma nova Métrica(nome, tipo)"),
		},
	})
}
