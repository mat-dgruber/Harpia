package telemetria

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/mat-dgruber/Harpia/ptst"
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

var TipoSpan = ptst.TipoObjeto.NewTipo("Span", "Span de rastreamento de OpenTelemetry")

func (s *Span) Tipo() *ptst.Tipo {
	return TipoSpan
}

func (s *Span) M__obtem_attributo__(nome string) (ptst.Objeto, error) {
	switch nome {
	case "finalizar":
		return ptst.NewMetodoOuPanic("finalizar", func(inst ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
			status := "OK"
			if len(args) >= 1 {
				st, _ := ptst.NewTexto(args[0])
				status = string(st.(ptst.Texto))
			}
			s.Fim = time.Now()
			s.DuracaoMs = s.Fim.Sub(s.Inicio).Milliseconds()
			s.Status = status

			// Exporta o Trace/Span em JSON estruturado para Stdout
			bytes, _ := json.Marshal(s)
			fmt.Fprintln(os.Stdout, string(bytes))
			return ptst.Nulo, nil
		}, ""), nil
	}
	return nil, ptst.NewErroF(ptst.AtributoErro, "Atributo '%s' não existe em Span", nome)
}

type Tracer struct {
	Servico string
}

var TipoTracer = ptst.TipoObjeto.NewTipo("Tracer", "Rastreador de traces de OpenTelemetry")

func (t *Tracer) Tipo() *ptst.Tipo {
	return TipoTracer
}

func (t *Tracer) M__obtem_attributo__(nome string) (ptst.Objeto, error) {
	switch nome {
	case "iniciar_span":
		return ptst.NewMetodoOuPanic("iniciar_span", func(inst ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
			if err := ptst.VerificaNumeroArgumentos("iniciar_span", false, args, 1, 1); err != nil {
				return nil, err
			}
			nomeSpan, err := ptst.NewTexto(args[0])
			if err != nil {
				return nil, err
			}

			now := time.Now().UnixNano()
			traceID := fmt.Sprintf("%x", now)
			spanID := fmt.Sprintf("%x", now&0xffff)

			return &Span{
				TraceID: traceID,
				SpanID:  spanID,
				Nome:    string(nomeSpan.(ptst.Texto)),
				Servico: t.Servico,
				Inicio:  time.Now(),
			}, nil
		}, ""), nil
	}
	return nil, ptst.NewErroF(ptst.AtributoErro, "Atributo '%s' não existe em Tracer", nome)
}

type Metrica struct {
	Nome    string
	Kind    string
	Valores map[string]float64
	mu      sync.Mutex
}

var TipoMetrica = ptst.TipoObjeto.NewTipo("Metrica", "Métrica quantitativa leve")

func (m *Metrica) Tipo() *ptst.Tipo {
	return TipoMetrica
}

func (m *Metrica) M__obtem_attributo__(nome string) (ptst.Objeto, error) {
	switch nome {
	case "registrar":
		return ptst.NewMetodoOuPanic("registrar", func(inst ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
			if err := ptst.VerificaNumeroArgumentos("registrar", false, args, 1, 2); err != nil {
				return nil, err
			}
			val, err := ptst.NewDecimal(args[0])
			if err != nil {
				return nil, err
			}
			tag := "default"
			if len(args) == 2 {
				t, _ := ptst.NewTexto(args[1])
				tag = string(t.(ptst.Texto))
			}

			m.mu.Lock()
			m.Valores[tag] += float64(val.(ptst.Decimal))
			m.mu.Unlock()

			// Emite métrica formatada
			fmt.Printf("{\"metric\": \"%s\", \"type\": \"%s\", \"tag\": \"%s\", \"value\": %f}\n", m.Nome, m.Kind, tag, val.(ptst.Decimal))
			return ptst.Nulo, nil
		}, ""), nil
	}
	return nil, ptst.NewErroF(ptst.AtributoErro, "Atributo '%s' não existe em Metrica", nome)
}

func met_novo_tracer(_ ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
	if err := ptst.VerificaNumeroArgumentos("novo_tracer", false, args, 1, 1); err != nil {
		return nil, err
	}
	servico, err := ptst.NewTexto(args[0])
	if err != nil {
		return nil, err
	}
	return &Tracer{Servico: string(servico.(ptst.Texto))}, nil
}

func met_nova_metrica(_ ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
	if err := ptst.VerificaNumeroArgumentos("nova_metrica", false, args, 2, 2); err != nil {
		return nil, err
	}
	nome, err := ptst.NewTexto(args[0])
	if err != nil {
		return nil, err
	}
	tipo, err := ptst.NewTexto(args[1])
	if err != nil {
		return nil, err
	}
	return &Metrica{
		Nome:    string(nome.(ptst.Texto)),
		Kind:    string(tipo.(ptst.Texto)),
		Valores: make(map[string]float64),
	}, nil
}

func init() {
	ptst.RegistraModuloImpl(&ptst.ModuloImpl{
		Info: ptst.ModuloInfo{
			Nome:    "telemetria",
			Arquivo: "stdlib/telemetria",
			Doc:     "Módulo leve de observabilidade compatível com OpenTelemetry (Traces e Métricas)",
		},
		Metodos: []*ptst.Metodo{
			ptst.NewMetodoOuPanic("novo_tracer", met_novo_tracer, "Cria um novo Tracer(servico)"),
			ptst.NewMetodoOuPanic("nova_metrica", met_nova_metrica, "Cria uma nova Métrica(nome, tipo)"),
		},
	})
}
