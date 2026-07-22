// Package metricas fornece suporte nativo a métricas instrumentadas em formato compatível com Prometheus (Exposition Format),
// expondo tipos estruturais como Contadores (Counter) e Medidores (Gauge) thread-safe para observabilidade.
package metricas

import (
	"fmt"
	"strings"
	"sync"

	"github.com/mat-dgruber/Harpia/hrp"
)

// MetricaInfo guarda os metadados descritivos estruturais e o valor instantâneo de uma métrica física.
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

// Contador representa uma métrica cumulativa monotônica cujo valor só pode aumentar.
type Contador struct {
	info *MetricaInfo
}

// TipoContador define e expõe a classe Contador na VM do Harpia.
var TipoContador = hrp.NewTipo("Contador", "Contador Prometheus para incrementar valores")

// Tipo retorna o tipo da classe na VM.
func (c *Contador) Tipo() *hrp.Tipo {
	return TipoContador
}

// M__obtem_attributo__ mapeia o método incrementar() no Contador de forma thread-safe.
func (c *Contador) M__obtem_attributo__(nome string) (hrp.Objeto, error) {
	switch nome {
	case "incrementar":
		// Incrementa em +1 o valor do Contador de forma atômica sob lock.
		return hrp.NewMetodoOuPanic("incrementar", func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
			c.info.mu.Lock()
			c.info.Valor += 1
			c.info.mu.Unlock()
			return hrp.Nulo, nil
		}, "Soma mais uma unidade ao contador."), nil
	}
	return nil, hrp.NewErroF(hrp.AtributoErro, "Atributo '%s' não existe no Contador", nome)
}

// Medidor (Gauge) representa uma métrica instantânea que pode tanto subir quanto descer (ex: uso de RAM, CPU, conexões ativas).
type Medidor struct {
	info *MetricaInfo
}

// TipoMedidor define e expõe a classe Medidor na VM do Harpia.
var TipoMedidor = hrp.NewTipo("Medidor", "Medidor Prometheus (Gauge) para alterar valores")

// Tipo retorna a representação na VM.
func (m *Medidor) Tipo() *hrp.Tipo {
	return TipoMedidor
}

// M__obtem_attributo__ mapeia o método definir() no Medidor de forma thread-safe.
func (m *Medidor) M__obtem_attributo__(nome string) (hrp.Objeto, error) {
	switch nome {
	case "definir":
		// Altera de forma instantânea e arbitrária o valor do Medidor sob lock.
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
		}, "Define um valor decimal instantâneo para o medidor."), nil
	}
	return nil, hrp.NewErroF(hrp.AtributoErro, "Atributo '%s' não existe no Medidor", nome)
}

// met_metricas_criarContador implementa 'criarContador(nome, descricao)' em nível de script Harpia.
// Fabrica e cadastra de forma única e thread-safe um Contador no registro de observabilidade.
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

// met_metricas_criarMedidor implementa 'criarMedidor(nome, descricao)' em nível de script Harpia.
// Fabrica e cadastra de forma única e thread-safe um Medidor no registro de observabilidade.
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

// met_metricas_expor implementa 'expor()' em nível de script Harpia.
// Compila e serializa todas as métricas cadastradas no formato de texto oficial aceito pelo Prometheus (/metrics).
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
	// Registra o módulo 'metricas' no sistema de módulos da biblioteca padrão do Harpia.
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
			hrp.NewMetodoOuPanic("criarContador", met_metricas_criarContador, "Cria ou retorna uma referência registrada para um Contador Prometheus."),
			hrp.NewMetodoOuPanic("criarMedidor", met_metricas_criarMedidor, "Cria ou retorna uma referência registrada para um Medidor (Gauge) Prometheus."),
			hrp.NewMetodoOuPanic("expor", met_metricas_expor, "Exprime todas as métricas registradas em formato Prometheus estruturado para raspagem (Scraping)."),
		},
	})
}
