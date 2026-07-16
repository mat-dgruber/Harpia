package metricas

import (
	"fmt"
	"strings"
	"sync"

	"github.com/natanfeitosa/portuscript/ptst"
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

var TipoContador = ptst.NewTipo("Contador", "Contador Prometheus para incrementar valores")

func (c *Contador) Tipo() *ptst.Tipo {
	return TipoContador
}

func (c *Contador) M__obtem_attributo__(nome string) (ptst.Objeto, error) {
	switch nome {
	case "incrementar":
		return ptst.NewMetodoOuPanic("incrementar", func(inst ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
			c.info.mu.Lock()
			c.info.Valor += 1
			c.info.mu.Unlock()
			return ptst.Nulo, nil
		}, ""), nil
	}
	return nil, ptst.NewErroF(ptst.AtributoErro, "Atributo '%s' não existe no Contador", nome)
}

type Medidor struct {
	info *MetricaInfo
}

var TipoMedidor = ptst.NewTipo("Medidor", "Medidor Prometheus (Gauge) para alterar valores")

func (m *Medidor) Tipo() *ptst.Tipo {
	return TipoMedidor
}

func (m *Medidor) M__obtem_attributo__(nome string) (ptst.Objeto, error) {
	switch nome {
	case "definir":
		return ptst.NewMetodoOuPanic("definir", func(inst ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
			if err := ptst.VerificaNumeroArgumentos("definir", false, args, 1, 1); err != nil {
				return nil, err
			}
			val, err := ptst.NewDecimal(args[0])
			if err != nil {
				return nil, err
			}
			m.info.mu.Lock()
			m.info.Valor = float64(val.(ptst.Decimal))
			m.info.mu.Unlock()
			return ptst.Nulo, nil
		}, ""), nil
	}
	return nil, ptst.NewErroF(ptst.AtributoErro, "Atributo '%s' não existe no Medidor", nome)
}

func met_metricas_criarContador(inst ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
	if err := ptst.VerificaNumeroArgumentos("criarContador", false, args, 2, 2); err != nil {
		return nil, err
	}
	nome, _ := ptst.NewTexto(args[0])
	desc, _ := ptst.NewTexto(args[1])

	nomeStr := string(nome.(ptst.Texto))
	descStr := string(desc.(ptst.Texto))

	registroMu.Lock()
	defer registroMu.Unlock()

	info, ok := registro[nomeStr]
	if !ok {
		info = &MetricaInfo{Nome: nomeStr, Descricao: descStr, Tipo: "counter"}
		registro[nomeStr] = info
	}

	return &Contador{info: info}, nil
}

func met_metricas_criarMedidor(inst ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
	if err := ptst.VerificaNumeroArgumentos("criarMedidor", false, args, 2, 2); err != nil {
		return nil, err
	}
	nome, _ := ptst.NewTexto(args[0])
	desc, _ := ptst.NewTexto(args[1])

	nomeStr := string(nome.(ptst.Texto))
	descStr := string(desc.(ptst.Texto))

	registroMu.Lock()
	defer registroMu.Unlock()

	info, ok := registro[nomeStr]
	if !ok {
		info = &MetricaInfo{Nome: nomeStr, Descricao: descStr, Tipo: "gauge"}
		registro[nomeStr] = info
	}

	return &Medidor{info: info}, nil
}

func met_metricas_expor(inst ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
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
	return ptst.Texto(sb.String()), nil
}

func init() {
	ptst.RegistraModuloImpl(&ptst.ModuloImpl{
		Info: ptst.ModuloInfo{
			Nome:    "metricas",
			Arquivo: "stdlib/metricas",
		},
		Constantes: ptst.Mapa{
			"Contador": TipoContador,
			"Medidor":  TipoMedidor,
		},
		Metodos: []*ptst.Metodo{
			ptst.NewMetodoOuPanic("criarContador", met_metricas_criarContador, "Cria um contador Prometheus."),
			ptst.NewMetodoOuPanic("criarMedidor", met_metricas_criarMedidor, "Cria um medidor (Gauge) Prometheus."),
			ptst.NewMetodoOuPanic("expor", met_metricas_expor, "Exprime todas as métricas registradas em formato Prometheus."),
		},
	})
}
