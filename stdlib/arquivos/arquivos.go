// Package arquivos implementa as rotinas da biblioteca padrão (stdlib) do Harpia
// destinadas a operações de entrada e saída (I/O) de arquivos e manipulação de diretórios,
// contendo regras de restrição de segurança e sandbox para confinamento e validação.
package arquivos

import (
	"io/fs"
	"os"
	"path/filepath"

	"github.com/mat-dgruber/Harpia/hrp"
)

// verificarPermissao é um utilitário interno que valida se a instância do módulo ativo
// tem privilégios adequados no sandbox de segurança do Harpia para tocar em recursos físicos de arquivo.
// Ele impede que o código execute ataques do tipo Path Traversal se o sandbox estiver travado.
func verificarPermissao(inst hrp.Objeto) error {
	if modulo, ok := inst.(*hrp.Modulo); ok && modulo != nil {
		return modulo.Contexto.VerificarPermissaoArquivos()
	}
	return nil
}

// met_arq_ler implementa a função 'ler(caminho)' em nível de script Harpia.
// Ele realiza a validação de segurança antes de abrir o arquivo local e retorna seu conteúdo como um objeto Texto.
func met_arq_ler(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if err := verificarPermissao(inst); err != nil {
		return nil, err
	}

	if err := hrp.VerificaNumeroArgumentos("ler", false, args, 1, 1); err != nil {
		return nil, err
	}

	caminho, err := hrp.NewTexto(args[0])
	if err != nil {
		return nil, err
	}

	// Lê todo o conteúdo físico do arquivo utilizando a API nativa do SO.
	bytes, err := os.ReadFile(string(caminho.(hrp.Texto)))
	if err != nil {
		return nil, hrp.NewErroF(hrp.ErroDeSistema, "Erro ao ler arquivo '%s': %v", caminho, err)
	}

	return hrp.Texto(bytes), nil
}

// met_arq_escrever implementa a função 'escrever(caminho, conteudo)' em nível de script Harpia.
// Sobrescreve totalmente ou cria um novo arquivo com o conteúdo textual fornecido, respeitando
// a máscara de permissões padrão de escrita de arquivos do sistema de arquivos (0644).
func met_arq_escrever(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if err := verificarPermissao(inst); err != nil {
		return nil, err
	}

	if err := hrp.VerificaNumeroArgumentos("escrever", false, args, 2, 2); err != nil {
		return nil, err
	}

	caminho, err := hrp.NewTexto(args[0])
	if err != nil {
		return nil, err
	}

	conteudo, err := hrp.NewTexto(args[1])
	if err != nil {
		return nil, err
	}

	err = os.WriteFile(string(caminho.(hrp.Texto)), []byte(conteudo.(hrp.Texto)), 0644)
	if err != nil {
		return nil, hrp.NewErroF(hrp.ErroDeSistema, "Erro ao escrever no arquivo '%s': %v", caminho, err)
	}

	return hrp.Nulo, nil
}

// met_arq_acrescentar implementa a função 'acrescentar(caminho, conteudo)' em nível de script Harpia.
// Abre um arquivo no modo de concatenação (append), criando-o se não existir, e injeta o conteúdo
// ao final das linhas existentes de forma thread-safe baseada em semântica nativa de arquivo do SO.
func met_arq_acrescentar(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if err := verificarPermissao(inst); err != nil {
		return nil, err
	}

	if err := hrp.VerificaNumeroArgumentos("acrescentar", false, args, 2, 2); err != nil {
		return nil, err
	}

	caminho, err := hrp.NewTexto(args[0])
	if err != nil {
		return nil, err
	}

	conteudo, err := hrp.NewTexto(args[1])
	if err != nil {
		return nil, err
	}

	f, err := os.OpenFile(string(caminho.(hrp.Texto)), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, hrp.NewErroF(hrp.ErroDeSistema, "Erro ao abrir arquivo '%s' para acrescentar: %v", caminho, err)
	}
	defer f.Close()

	if _, err := f.Write([]byte(conteudo.(hrp.Texto))); err != nil {
		return nil, hrp.NewErroF(hrp.ErroDeSistema, "Erro ao acrescentar dados no arquivo '%s': %v", caminho, err)
	}

	return hrp.Nulo, nil
}

// met_arq_remover implementa a função 'remover(caminho)' em nível de script Harpia.
// Exclui de forma recursiva arquivos ou pastas inteiras, simulando um comando de exclusão direta.
func met_arq_remover(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if err := verificarPermissao(inst); err != nil {
		return nil, err
	}

	if err := hrp.VerificaNumeroArgumentos("remover", false, args, 1, 1); err != nil {
		return nil, err
	}

	caminho, err := hrp.NewTexto(args[0])
	if err != nil {
		return nil, err
	}

	err = os.RemoveAll(string(caminho.(hrp.Texto)))
	if err != nil {
		return nil, hrp.NewErroF(hrp.ErroDeSistema, "Erro ao remover '%s': %v", caminho, err)
	}

	return hrp.Nulo, nil
}

// met_arq_renomear implementa a função 'renomear(antigo, novo)' em nível de script Harpia.
// Realiza a movimentação física ou alteração do identificador do arquivo/pasta no disco rígido.
func met_arq_renomear(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if err := verificarPermissao(inst); err != nil {
		return nil, err
	}

	if err := hrp.VerificaNumeroArgumentos("renomear", false, args, 2, 2); err != nil {
		return nil, err
	}

	antigo, err := hrp.NewTexto(args[0])
	if err != nil {
		return nil, err
	}

	novo, err := hrp.NewTexto(args[1])
	if err != nil {
		return nil, err
	}

	err = os.Rename(string(antigo.(hrp.Texto)), string(novo.(hrp.Texto)))
	if err != nil {
		return nil, hrp.NewErroF(hrp.ErroDeSistema, "Erro ao renomear '%s' para '%s': %v", antigo, novo, err)
	}

	return hrp.Nulo, nil
}

// met_arq_caminhar implementa a função 'caminhar(diretorio)' em nível de script Harpia.
// Varre recursivamente todas as subdiretórios a partir de um nó base, extraindo e listando
// os caminhos relativos unificados, devolvendo-os ao interpretador na forma de uma Lista.
func met_arq_caminhar(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if err := verificarPermissao(inst); err != nil {
		return nil, err
	}

	if err := hrp.VerificaNumeroArgumentos("caminhar", false, args, 1, 1); err != nil {
		return nil, err
	}

	diretorio, err := hrp.NewTexto(args[0])
	if err != nil {
		return nil, err
	}

	lista := &hrp.Lista{Itens: make([]hrp.Objeto, 0)}

	err = filepath.WalkDir(string(diretorio.(hrp.Texto)), func(caminho string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		lista.Adiciona(hrp.Texto(caminho))
		return nil
	})

	if err != nil {
		return nil, hrp.NewErroF(hrp.ErroDeSistema, "Erro ao caminhar no diretório '%s': %v", diretorio, err)
	}

	return lista, nil
}

// met_arq_juntar implementa a função 'juntar(a, b, ...)' em nível de script Harpia.
// Concatena de forma segura e inteligível partes de um caminho de diretório utilizando
// a barramenclatura própria do sistema operacional em execução (ex: '\' no Windows e '/' no Unix/MacOS).
func met_arq_juntar(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if len(args) == 0 {
		return hrp.Texto(""), nil
	}

	partes := make([]string, len(args))
	for i, arg := range args {
		txt, err := hrp.NewTexto(arg)
		if err != nil {
			return nil, err
		}
		partes[i] = string(txt.(hrp.Texto))
	}

	res := filepath.Join(partes...)
	return hrp.Texto(res), nil
}

// met_arq_resolver implementa a função 'resolver(caminho)' em nível de script Harpia.
// Trata o caminho fornecido e retorna o endereço físico absoluto inequívoco no sistema.
func met_arq_resolver(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if err := verificarPermissao(inst); err != nil {
		return nil, err
	}

	if err := hrp.VerificaNumeroArgumentos("resolver", false, args, 1, 1); err != nil {
		return nil, err
	}

	caminho, err := hrp.NewTexto(args[0])
	if err != nil {
		return nil, err
	}

	res, err := filepath.Abs(string(caminho.(hrp.Texto)))
	if err != nil {
		return nil, hrp.NewErroF(hrp.ErroDeSistema, "Erro ao resolver caminho absoluto de '%s': %v", caminho, err)
	}

	return hrp.Texto(res), nil
}

var _ler = hrp.NewMetodoOuPanic("ler", met_arq_ler, "Lê todo o conteúdo de um arquivo em disco.")
var _escrever = hrp.NewMetodoOuPanic("escrever", met_arq_escrever, "Escreve ou sobrescreve conteúdo de texto em um arquivo.")
var _acrescentar = hrp.NewMetodoOuPanic("acrescentar", met_arq_acrescentar, "Anexa conteúdo textual ao final de um arquivo.")
var _remover = hrp.NewMetodoOuPanic("remover", met_arq_remover, "Remove um arquivo ou diretório físico do disco.")
var _renomear = hrp.NewMetodoOuPanic("renomear", met_arq_renomear, "Altera o nome ou move um arquivo/diretório de local.")
var _caminhar = hrp.NewMetodoOuPanic("caminhar", met_arq_caminhar, "Lista todos os arquivos e pastas de forma recursiva a partir do diretório fornecido.")
var _juntar = hrp.NewMetodoOuPanic("juntar", met_arq_juntar, "Une múltiplos segmentos de caminhos de diretório respeitando o separador do sistema operacional.")
var _resolver = hrp.NewMetodoOuPanic("resolver", met_arq_resolver, "Devolve o caminho absoluto de um endereço relativo.")

func init() {
	// Registra o módulo 'arquivos' globalmente no ecossistema do interpretador Harpia.
	hrp.RegistraModuloImpl(&hrp.ModuloImpl{
		Info: hrp.ModuloInfo{
			Nome:    "arquivos",
			Arquivo: "stdlib/arquivos",
		},
		Metodos: []*hrp.Metodo{
			_ler,
			_escrever,
			_acrescentar,
			_remover,
			_renomear,
			_caminhar,
			_juntar,
			_resolver,
		},
	})
}
