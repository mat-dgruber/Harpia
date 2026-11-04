// Package sistema implementa o módulo nativo de informações e propriedades de ambiente
// da biblioteca padrão do Harpia.
//
// Este pacote permite que scripts em Harpia consultem detalhes dinâmicos sobre
// a arquitetura de processador e o sistema operacional hospedeiro no qual o interpretador está rodando.
package sistema

import (
	"bytes"
	"os"
	"os/exec"
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

	// executar executa um comando do sistema operacional e retorna um Mapa com "saida" e "erro".
	fnExecutar := hrp.NewMetodoOuPanic("executar", func(_ hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
		if ctx := hrp.ObterContextoAtivo(); ctx != nil {
			if err := ctx.VerificarPermissaoArquivos(); err != nil {
				return nil, err
			}
		}

		if err := hrp.VerificaNumeroArgumentos("executar", false, args, 1, 2); err != nil {
			return nil, err
		}
		cmdStr, ok := args[0].(hrp.Texto)
		if !ok {
			return nil, hrp.NewErroF(hrp.TipagemErro, "O primeiro argumento de 'executar' deve ser do tipo 'Texto'")
		}

		var cmdArgs []string
		if len(args) >= 2 {
			switch a := args[1].(type) {
			case hrp.Tupla:
				for _, item := range a {
					t, ok := item.(hrp.Texto)
					if !ok {
						return nil, hrp.NewErroF(hrp.TipagemErro, "Todos os argumentos da lista devem ser do tipo 'Texto'")
					}
					cmdArgs = append(cmdArgs, string(t))
				}
			case *hrp.Lista:
				for _, item := range a.Itens {
					t, ok := item.(hrp.Texto)
					if !ok {
						return nil, hrp.NewErroF(hrp.TipagemErro, "Todos os argumentos da lista devem ser do tipo 'Texto'")
					}
					cmdArgs = append(cmdArgs, string(t))
				}
			}
		}

		cmd := exec.Command(string(cmdStr), cmdArgs...)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		_ = cmd.Run() // ignorar código de saída; retornar saída mesmo em erro

		return hrp.Mapa{
			"saida": hrp.Texto(stdout.String()),
			"erro":  hrp.Texto(stderr.String()),
		}, nil
	}, "executa um comando do sistema operacional e retorna {saida, erro}")

	// executarCodigoHarpia grava o código em um arquivo temporário e executa o interpretador Harpia nativo,
	// retornando {saida, erro} com o stdout/stderr reais da execução.
	fnExecutarCodigoHarpia := hrp.NewMetodoOuPanic("executarCodigoHarpia", func(_ hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
		if ctx := hrp.ObterContextoAtivo(); ctx != nil {
			if err := ctx.VerificarPermissaoArquivos(); err != nil {
				return nil, err
			}
		}

		if err := hrp.VerificaNumeroArgumentos("executarCodigoHarpia", false, args, 1, 1); err != nil {
			return nil, err
		}
		codigo, ok := args[0].(hrp.Texto)
		if !ok {
			return nil, hrp.NewErroF(hrp.TipagemErro, "O argumento de 'executarCodigoHarpia' deve ser do tipo 'Texto'")
		}

		// Escrever o código num arquivo temporário
		tmpFile, err := os.CreateTemp("", "hrp_playground_*.hrp")
		if err != nil {
			return hrp.Mapa{
				"saida": hrp.Texto(""),
				"erro":  hrp.Texto("Erro ao criar arquivo temporário: " + err.Error()),
			}, nil
		}
		defer os.Remove(tmpFile.Name())

		if _, err := tmpFile.WriteString(string(codigo)); err != nil {
			tmpFile.Close()
			return hrp.Mapa{
				"saida": hrp.Texto(""),
				"erro":  hrp.Texto("Erro ao escrever código: " + err.Error()),
			}, nil
		}
		tmpFile.Close()

		// Descobrir o executável atual do Harpia
		execPath, err := os.Executable()
		if err != nil {
			execPath = "Harpia" // fallback para PATH
		}

		cmd := exec.Command(execPath, "executar", "--silencioso", tmpFile.Name())
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		_ = cmd.Run()

		return hrp.Mapa{
			"saida": hrp.Texto(stdout.String()),
			"erro":  hrp.Texto(stderr.String()),
		}, nil
	}, "executa um código Harpia em subprocesso e retorna {saida, erro}")

	// metodos expõe funções executáveis do módulo sistema.
	metodos := []*hrp.Metodo{fnExecutar, fnExecutarCodigoHarpia}

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
