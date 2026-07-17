package logs

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/mat-dgruber/Harpia/ptst"
)

var (
	formatoLog = "texto" // "texto" ou "json"
	usarCores  = true
)

func formatarData() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func logar(nivel string, cor string, args ptst.Tupla) (ptst.Objeto, error) {
	if len(args) == 0 {
		return nil, ptst.NewErroF(ptst.TipagemErro, "esperava no mínimo 1 argumento (mensagem)")
	}

	msgStr := fmt.Sprintf("%v", args[0])
	
	var meta map[string]interface{}
	if len(args) > 1 {
		if mapa, ok := args[1].(ptst.Mapa); ok {
			meta = make(map[string]interface{})
			for k, v := range mapa {
				meta[k] = fmt.Sprintf("%v", v)
			}
		}
	}

	if formatoLog == "json" {
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
		// Modo Texto
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

	return ptst.Nulo, nil
}

func met_logs_info(inst ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
	return logar("INFO", "\x1b[1;34m", args) // Azul
}

func met_logs_alerta(inst ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
	return logar("ALERTA", "\x1b[1;33m", args) // Amarelo
}

func met_logs_erro(inst ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
	return logar("ERRO", "\x1b[1;31m", args) // Vermelho
}

func met_logs_depurar(inst ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
	return logar("DEPURAR", "\x1b[1;36m", args) // Ciano
}

func met_logs_configurar(inst ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
	if len(args) >= 1 {
		if fmtStr, ok := args[0].(ptst.Texto); ok {
			formatoLog = strings.ToLower(string(fmtStr))
		}
	}
	if len(args) >= 2 {
		if coresBool, ok := args[1].(ptst.Booleano); ok {
			usarCores = bool(coresBool)
		}
	}
	return ptst.Nulo, nil
}

func init() {
	ptst.RegistraModuloImpl(&ptst.ModuloImpl{
		Info: ptst.ModuloInfo{
			Nome:    "logs",
			Arquivo: "stdlib/logs",
		},
		Metodos: []*ptst.Metodo{
			ptst.NewMetodoOuPanic("info", met_logs_info, "Loga uma mensagem informativa."),
			ptst.NewMetodoOuPanic("alerta", met_logs_alerta, "Loga um alerta."),
			ptst.NewMetodoOuPanic("erro", met_logs_erro, "Loga uma mensagem de erro."),
			ptst.NewMetodoOuPanic("depurar", met_logs_depurar, "Loga informações de depuração."),
			ptst.NewMetodoOuPanic("configurar", met_logs_configurar, "Configura o formato ('texto' ou 'json') e uso de cores."),
		},
	})
}
