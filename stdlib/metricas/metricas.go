package metricas

import (
	"fmt"
	"strings"
	"sync"

	"github.com/mat-dgruber/Harpia/hrp"
)

type MetricaInfo struct {
	Nome      string
	Descricao string
	Tipo      string // "counter" ou "gauge"
	Valor     float64
	mu        sync.RWMutex
}

var (
	registro   = make(map[string]*MetricaInfo)
	registroMu sync.RWMutex
)

type Contador struct {
	info *MetricaInfo
}

var TipoContador = hrp.NewTipo("Contador", "Contador Prometheus para incrementar valores")

func (c *Contador) Tipo() *hrp.Tipo {
	return TipoContador
}

func (c *Contador) M__obtem_attributo__(nome string) (hrp.Objeto, error) {
	switch nome {
	case "incrementar":
		return hrp.NewMetodoOuPanic("incrementar", func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
			c.info.mu.Lock()
			c.info.Valor += 1
			c.info.mu.Unlock()
			return hrp.Nulo, nil
		}, ""), nil
	}
	return nil, hrp.NewErroF(hrp.AtributoErro, "Atributo '%s' não existe no Contador", nome)
}

type Medidor struct {
	info *MetricaInfo
}

var TipoMedidor = hrp.NewTipo("Medidor", "Medidor Prometheus (Gauge) para alterar valores")

func (m *Medidor) Tipo() *hrp.Tipo {
	return TipoMedidor
}

func (m *Medidor) M__obtem_attributo__(nome string) (hrp.Objeto, error) {
	switch nome {
	case "definir":
		return hrp.NewMetodoOuPanic("definir", func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
			if err := hrp.VerificaNumeroArgumentos("definir", false, args, 1, 1); err != nil {
				return nil, err
			}
			val, err := hrp.NewDecimal(args[0])
			if err != nil {
				return nil, err
			}
			m.info.mu.Lock()
			m.info.Valor = float64(val.(hrp.Decimal))
			m.info.mu.Unlock()
			return hrp.Nulo, nil
		}, ""), nil
	}
	return nil, hrp.NewErroF(hrp.AtributoErro, "Atributo '%s' não existe no Medidor", nome)
}

func met_metricas_criarContador(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if err := hrp.VerificaNumeroArgumentos("criarContador", false, args, 2, 2); err != nil {
		return nil, err
	}
	nome, _ := hrp.NewTexto(args[0])
	desc, _ := hrp.NewTexto(args[1])

	nomeStr := string(nome.(hrp.Texto))
	descStr := string(desc.(hrp.Texto))

	registroMu.Lock()
	defer registroMu.Unlock()

	info, ok := registro[nomeStr]
	if !ok {
		info = &MetricaInfo{Nome: nomeStr, Descricao: descStr, Tipo: "counter"}
		registro[nomeStr] = info
	}

	return &Contador{info: info}, nil
}

func met_metricas_criarMedidor(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if err := hrp.VerificaNumeroArgumentos("criarMedidor", false, args, 2, 2); err != nil {
		return nil, err
	}
	nome, _ := hrp.NewTexto(args[0])
	desc, _ := hrp.NewTexto(args[1])

	nomeStr := string(nome.(hrp.Texto))
	descStr := string(desc.(hrp.Texto))

	registroMu.Lock()
	defer registroMu.Unlock()

	info, ok := registro[nomeStr]
	if !ok {
		info = &MetricaInfo{Nome: nomeStr, Descricao: descStr, Tipo: "gauge"}
		registro[nomeStr] = info
	}

	return &Medidor{info: info}, nil
}

func met_metricas_expor(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	registroMu.RLock()
	defer registroMu.RUnlock()

	var sb strings.Builder
	for _, info := range registro {
		info.mu.RLock()
		sb.WriteString(fmt.Sprintf("# HELP %s %s\n", info.Nome, info.Descricao))
		sb.WriteString(fmt.Sprintf("# TYPE %s %s\n", info.Nome, info.Tipo))
		sb.WriteString(fmt.Sprintf("%s %g\n", info.Nome, info.Valor))
		info.mu.RUnlock()
	}
	return hrp.Texto(sb.String()), nil
}

func init() {
	hrp.RegistraModuloImpl(&hrp.ModuloImpl{
		Info: hrp.ModuloInfo{
			Nome:    "metricas",
			Arquivo: "stdlib/metricas",
		},
		Constantes: hrp.Mapa{
			"Contador": TipoContador,
			"Medidor":  TipoMedidor,
		},
		Metodos: []*hrp.Metodo{
			hrp.NewMetodoOuPanic("criarContador", met_metricas_criarContador, "Cria um contador Prometheus."),
			hrp.NewMetodoOuPanic("criarMedidor", met_metricas_criarMedidor, "Cria um medidor (Gauge) Prometheus."),
			hrp.NewMetodoOuPanic("expor", met_metricas_expor, "Exprime todas as métricas registradas em formato Prometheus."),
		},
	})
}
