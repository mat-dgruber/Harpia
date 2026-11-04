// Package sistema implementa o módulo nativo de informações e propriedades de ambiente
// da biblioteca padrão do Harpia.
//
// Este pacote permite que scripts em Harpia consultem detalhes dinâmicos sobre
// a arquitetura de processador e o sistema operacional hospedeiro no qual o interpretador está rodando.
package sistema

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/mat-dgruber/Harpia/hrp"
)

// TempoLimiteExecucaoIsolada é o tempo máximo de executarCodigoHarpia.
var TempoLimiteExecucaoIsolada = 10 * time.Second

// limiteSaidaIsolada limita stdout/stderr capturados de executarCodigoHarpia (1 MB cada).
const limiteSaidaIsolada = 1 << 20

// caminhoExecutavel localiza o binário do Harpia; substituível em testes.
var caminhoExecutavel = os.Executable

// bufferLimitado guarda no máximo `limite` bytes e descarta o resto, sem falhar a escrita
// (falhar faria o processo filho receber SIGPIPE e esconder a saída já produzida).
type bufferLimitado struct {
	buf     bytes.Buffer
	limite  int
	cortado bool
}

func (b *bufferLimitado) Write(p []byte) (int, error) {
	if resta := b.limite - b.buf.Len(); resta > 0 {
		if len(p) > resta {
			b.buf.Write(p[:resta])
			b.cortado = true
		} else {
			b.buf.Write(p)
		}
	} else if len(p) > 0 {
		b.cortado = true
	}
	return len(p), nil
}

func (b *bufferLimitado) String() string {
	if b.cortado {
		return b.buf.String() + "\n[saída truncada em 1 MB]"
	}
	return b.buf.String()
}

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
	fnExecutar := hrp.NewMetodoOuPanic("executar", func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
		if err := hrp.ExigirPermissao(inst, hrp.RecursoProcessos); err != nil {
			return nil, err
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
	fnExecutarCodigoHarpia := hrp.NewMetodoOuPanic("executarCodigoHarpia", func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
		if err := hrp.ExigirPermissao(inst, hrp.RecursoProcessos); err != nil {
			return nil, err
		}

		if err := hrp.VerificaNumeroArgumentos("executarCodigoHarpia", false, args, 1, 1); err != nil {
			return nil, err
		}
		codigo, ok := args[0].(hrp.Texto)
		if !ok {
			return nil, hrp.NewErroF(hrp.TipagemErro, "O argumento de 'executarCodigoHarpia' deve ser do tipo 'Texto'")
		}

		// O código costuma vir de fora (ex.: playground do portal), então o subprocesso roda isolado:
		// diretório temporário próprio (não carrega o .env do projeto), ambiente mínimo (não herda
		// segredos), flag --sandbox (sem arquivos/processos/rede), tempo limite e saída limitada.
		dir, err := os.MkdirTemp("", "hrp_exec_*")
		if err != nil {
			return hrp.Mapa{
				"saida": hrp.Texto(""),
				"erro":  hrp.Texto("Erro ao criar diretório temporário: " + err.Error()),
			}, nil
		}
		defer os.RemoveAll(dir)

		arquivo := filepath.Join(dir, "codigo.hrp")
		if err := os.WriteFile(arquivo, []byte(codigo), 0600); err != nil {
			return hrp.Mapa{
				"saida": hrp.Texto(""),
				"erro":  hrp.Texto("Erro ao escrever código: " + err.Error()),
			}, nil
		}

		execPath, err := caminhoExecutavel()
		if err != nil {
			execPath = "harpia" // fallback para PATH
		}

		ctxExec, cancel := context.WithTimeout(context.Background(), TempoLimiteExecucaoIsolada)
		defer cancel()
		cmd := exec.CommandContext(ctxExec, execPath, "executar", "--silencioso", "--sandbox", arquivo)
		cmd.Dir = dir
		cmd.Env = []string{"PATH=" + os.Getenv("PATH"), "HOME=" + dir, "LANG=C.UTF-8"}
		stdout := &bufferLimitado{limite: limiteSaidaIsolada}
		stderr := &bufferLimitado{limite: limiteSaidaIsolada}
		cmd.Stdout = stdout
		cmd.Stderr = stderr
		_ = cmd.Run()

		erroTexto := stderr.String()
		if ctxExec.Err() == context.DeadlineExceeded {
			erroTexto += fmt.Sprintf("\nTempo limite de %v excedido: a execução foi interrompida.", TempoLimiteExecucaoIsolada)
		}

		return hrp.Mapa{
			"saida": hrp.Texto(stdout.String()),
			"erro":  hrp.Texto(erroTexto),
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
