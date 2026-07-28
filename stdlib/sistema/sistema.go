// Package sistema implementa o módulo nativo de informações e propriedades de ambiente
// da biblioteca padrão do Harpia.
//
// Este pacote permite que scripts em Harpia consultem detalhes dinâmicos sobre
// a arquitetura de processador e o sistema operacional hospedeiro no qual o interpretador está rodando.
package sistema

import (
	"os"
	"runtime"
	"strings"
	"sync"

	"github.com/mat-dgruber/Harpia/hrp"
)

func init() {
	var mu sync.Mutex
	envMap := hrp.Mapa{}
	atualizarEnv := func() {
		mu.Lock()
		defer mu.Unlock()
		for k := range envMap {
			delete(envMap, k)
		}
		for _, env := range os.Environ() {
			pair := strings.SplitN(env, "=", 2)
			if len(pair) == 2 {
				envMap[pair[0]] = hrp.Texto(pair[1])
			}
		}
	}

	atualizarEnv()
	hrp.OnEnvLoaded = append(hrp.OnEnvLoaded, atualizarEnv)

	sistemaObj := hrp.Mapa{
		"ARQUITETURA":       hrp.Texto(runtime.GOARCH),
		"NOME":              hrp.Texto(runtime.GOOS),
		"variaveisAmbiente": envMap,
	}

	// constantes expõe metadados estáticos do runtime do Go convertidos para hrp.Texto.
	constantes := hrp.Mapa{
		// ARQUITETURA expõe a arquitetura de CPU onde o interpretador está compilado (ex: "amd64", "arm64").
		"ARQUITETURA": hrp.Texto(runtime.GOARCH),

		// NOME expõe o identificador padrão do sistema operacional do computador hospedeiro (ex: "darwin", "linux", "windows").
		"NOME": hrp.Texto(runtime.GOOS),

		// Sistema expõe o objeto do sistema com suporte a variáveis de ambiente
		"Sistema": sistemaObj,
	}

	// metodos é inicializado vazio, reservado para futuras expansões e comandos do sistema operacional (ex: 'saida', 'executa_comando').
	metodos := []*hrp.Metodo{}

	// Registra o módulo 'sistema' na lista interna de módulos nativos do Harpia.
	hrp.RegistraModuloImpl(
		&hrp.ModuloImpl{
			Info: hrp.ModuloInfo{
				Nome:    "sistema",
				Arquivo: "stdlib/sistema",
			},
			Constantes: constantes,
			Metodos:    metodos,
		},
	)
}
