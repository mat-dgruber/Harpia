// Package logs implementa um gerador de logs unificado e estruturado (JSON ou texto colorizado),
// otimizado para ambientes produtivos nativos, cloud-native (Kubernetes) ou desenvolvimento local.
package logs

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/mat-dgruber/Harpia/hrp"
)

var (
	formatoLog = "texto" // "texto" ou "json"
	usarCores  = true
)

// formatarData retorna um timestamp simplificado e inteligível para logs locais em formato de texto.
func formatarData() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

// logar processa internamente a emissão das mensagens de log estruturado ou texto de forma otimizada.
// Suporta anotação opcional de um Mapa de metadados como segundo argumento para enriquecer os logs.
func logar(nivel string, cor string, args hrp.Tupla) (hrp.Objeto, error) {
	if len(args) == 0 {
		return nil, hrp.NewErroF(hrp.TipagemErro, "esperava no mínimo 1 argumento (mensagem)")
	}

	msgStr := fmt.Sprintf("%v", args[0])

	var meta map[string]interface{}
	if len(args) > 1 {
		if mapa, ok := args[1].(hrp.Mapa); ok {
			meta = make(map[string]interface{})
			for k, v := range mapa {
				meta[k] = fmt.Sprintf("%v", v)
			}
		}
	}

	if formatoLog == "json" {
		// Formato estruturado para agregadores de logs modernos (ElasticSearch, Grafana Loki, Datadog)
		logObj := map[string]interface{}{
			"data":  time.Now().Format(time.RFC3339),
			"nivel": strings.ToLower(nivel),
			"msg":   msgStr,
		}
		if meta != nil {
			logObj["metadados"] = meta
		}
		bytes, _ := json.Marshal(logObj)
		fmt.Println(string(bytes))
	} else {
		// Modo texto amigável e legível para console local com suporte a escape ANSI
		corInicio := ""
		corReset := ""
		if usarCores && os.Getenv("NO_COLOR") == "" {
			corInicio = cor
			corReset = "\x1b[0m"
		}

		metaStr := ""
		if meta != nil {
			bytes, _ := json.Marshal(meta)
			metaStr = " " + string(bytes)
		}

		fmt.Printf("[%s] %s%s%s: %s%s\n", formatarData(), corInicio, nivel, corReset, msgStr, metaStr)
	}

	return hrp.Nulo, nil
}

// met_logs_info implementa 'info(mensagem, meta?)' em nível de script Harpia.
func met_logs_info(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	return logar("INFO", "\x1b[1;34m", args) // Azul
}

// met_logs_alerta implementa 'alerta(mensagem, meta?)' em nível de script Harpia.
func met_logs_alerta(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	return logar("ALERTA", "\x1b[1;33m", args) // Amarelo
}

// met_logs_erro implementa 'erro(mensagem, meta?)' em nível de script Harpia.
func met_logs_erro(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	return logar("ERRO", "\x1b[1;31m", args) // Vermelho
}

// met_logs_depurar implementa 'depurar(mensagem, meta?)' em nível de script Harpia.
func met_logs_depurar(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	return logar("DEPURAR", "\x1b[1;36m", args) // Ciano
}

// met_logs_configurar implementa 'configurar(formato, cores?)' em nível de script Harpia.
// Modifica as propriedades globais de renderização (formato: "texto" ou "json").
func met_logs_configurar(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if len(args) >= 1 {
		if fmtStr, ok := args[0].(hrp.Texto); ok {
			formatoLog = strings.ToLower(string(fmtStr))
		}
	}
	if len(args) >= 2 {
		if coresBool, ok := args[1].(hrp.Booleano); ok {
			usarCores = bool(coresBool)
		}
	}
	return hrp.Nulo, nil
}

func init() {
	// Registra o módulo 'logs' no sistema central da biblioteca padrão do Harpia.
	hrp.RegistraModuloImpl(&hrp.ModuloImpl{
		Info: hrp.ModuloInfo{
			Nome:    "logs",
			Arquivo: "stdlib/logs",
		},
		Metodos: []*hrp.Metodo{
			hrp.NewMetodoOuPanic("info", met_logs_info, "Loga uma mensagem informativa."),
			hrp.NewMetodoOuPanic("alerta", met_logs_alerta, "Loga um alerta de sistema."),
			hrp.NewMetodoOuPanic("erro", met_logs_erro, "Loga uma mensagem de erro grave."),
			hrp.NewMetodoOuPanic("depurar", met_logs_depurar, "Loga informações detalhadas para depuração de fluxo."),
			hrp.NewMetodoOuPanic("configurar", met_logs_configurar, "Configura o formato ('texto' ou 'json') e uso de cores para o terminal."),
		},
	})
}
