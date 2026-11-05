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
	"time"

	"github.com/mat-dgruber/Harpia/hrp"
)

// TempoLimiteExecutar é o tempo máximo padrão de sistema.executar (ajustável por {tempoLimite}).
var TempoLimiteExecutar = 60 * time.Second

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

// TipoObjetoSistema é o tipo do objeto `Sistema` exportado pelo módulo.
var TipoObjetoSistema = hrp.NewTipo("Sistema", "Informações do sistema operacional e variáveis de ambiente")

// objetoSistema substitui o antigo Mapa estático: `variaveisAmbiente` é montado a cada leitura,
// exige permissão de processos (o ambiente costuma carregar segredos) e devolve uma cópia nova.
// Isso também elimina a corrida entre o recarregamento do .env e os scripts que liam o mapa.
type objetoSistema struct{}

func (objetoSistema) Tipo() *hrp.Tipo { return TipoObjetoSistema }

func (o objetoSistema) M__obtem_item__(chave hrp.Objeto) (hrp.Objeto, error) {
	nome, ok := chave.(hrp.Texto)
	if !ok {
		return nil, hrp.NewErroF(hrp.TipagemErro, "Sistema só aceita chaves do tipo Texto")
	}
	return o.M__obtem_attributo__(string(nome))
}

func (objetoSistema) M__obtem_attributo__(nome string) (hrp.Objeto, error) {
	switch nome {
	case "ARQUITETURA":
		return hrp.Texto(runtime.GOARCH), nil
	case "NOME":
		return hrp.Texto(runtime.GOOS), nil
	case "variaveisAmbiente":
		// ponytail: objeto global sem receptor; o contexto ativo é o do script em execução
		if err := hrp.ExigirPermissaoNoContexto(hrp.ObterContextoAtivo(), hrp.RecursoProcessos); err != nil {
			return nil, err
		}
		env := hrp.Mapa{}
		for _, par := range os.Environ() {
			if k, v, ok := strings.Cut(par, "="); ok {
				env[k] = hrp.Texto(v)
			}
		}
		return env, nil
	}
	return nil, hrp.NewErroF(hrp.ChaveErro, "Sistema não possui '%s'", nome)
}

func init() {
	sistemaObj := objetoSistema{}

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

		if err := hrp.VerificaNumeroArgumentos("executar", false, args, 1, 3); err != nil {
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

		// opções: {tempoLimite: segundos}; padrão TempoLimiteExecutar
		limite := TempoLimiteExecutar
		if len(args) == 3 {
			opcoes, ok := args[2].(hrp.Mapa)
			if !ok {
				return nil, hrp.NewErroF(hrp.TipagemErro, "O terceiro argumento de 'executar' deve ser um Mapa de opções, ex.: {tempoLimite: 30}")
			}
			if v, ok := opcoes["tempoLimite"]; ok {
				seg, ok := v.(hrp.Inteiro)
				if !ok || seg <= 0 {
					return nil, hrp.NewErroF(hrp.ValorErro, "'tempoLimite' deve ser um Inteiro positivo (segundos)")
				}
				limite = time.Duration(seg) * time.Second
			}
		}

		ctxExec, cancel := context.WithTimeout(context.Background(), limite)
		defer cancel()
		cmd := exec.CommandContext(ctxExec, string(cmdStr), cmdArgs...)
		stdout := &bufferLimitado{limite: limiteSaidaIsolada}
		stderr := &bufferLimitado{limite: limiteSaidaIsolada}
		cmd.Stdout = stdout
		cmd.Stderr = stderr
		errRun := cmd.Run()

		// codigo: código de saída do processo; -1 quando nem chegou a rodar ou foi interrompido
		codigo := 0
		erroTexto := stderr.String()
		if errRun != nil {
			codigo = -1
			if saida, ok := errRun.(*exec.ExitError); ok && saida.ExitCode() >= 0 {
				codigo = saida.ExitCode()
			} else if erroTexto == "" {
				erroTexto = errRun.Error()
			}
		}
		if ctxExec.Err() == context.DeadlineExceeded {
			codigo = -1
			erroTexto += fmt.Sprintf("\nTempo limite de %v excedido: o processo foi interrompido.", limite)
		}

		return hrp.Mapa{
			"saida":  hrp.Texto(stdout.String()),
			"erro":   hrp.Texto(erroTexto),
			"codigo": hrp.Inteiro(codigo),
		}, nil
	}, "executa um comando (com {tempoLimite} opcional) e retorna {saida, erro, codigo}")

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
