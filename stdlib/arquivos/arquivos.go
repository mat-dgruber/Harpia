package arquivos

import (
	"io/fs"
	"os"
	"path/filepath"

	"github.com/mat-dgruber/Harpia/hrp"
)

func verificarPermissao(inst hrp.Objeto) error {
	if modulo, ok := inst.(*hrp.Modulo); ok && modulo != nil {
		return modulo.Contexto.VerificarPermissaoArquivos()
	}
	return nil
}

// met_arq_ler implementa 'ler(caminho)' -> retorna Texto com o conteúdo do arquivo
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

	bytes, err := os.ReadFile(string(caminho.(hrp.Texto)))
	if err != nil {
		return nil, hrp.NewErroF(hrp.ErroDeSistema, "Erro ao ler arquivo '%s': %v", caminho, err)
	}

	return hrp.Texto(bytes), nil
}

// met_arq_escrever implementa 'escrever(caminho, conteudo)' -> escreve texto ou bytes no arquivo
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

// met_arq_acrescentar implementa 'acrescentar(caminho, conteudo)' -> anexa texto ao final do arquivo
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

// met_arq_remover implementa 'remover(caminho)' -> exclui um arquivo ou diretório
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

// met_arq_renomear implementa 'renomear(antigo, novo)' -> renomeia ou move arquivo/diretório
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

// met_arq_caminhar implementa 'caminhar(diretorio)' -> retorna Lista contendo caminhos encontrados recursivamente
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

// met_arq_juntar implementa 'juntar(a, b, ...)' -> concatena caminhos lógicos usando os separadores corretos do SO
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

// met_arq_resolver implementa 'resolver(caminho)' -> retorna o caminho físico absoluto unificado
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

var _ler = hrp.NewMetodoOuPanic("ler", met_arq_ler, "")
var _escrever = hrp.NewMetodoOuPanic("escrever", met_arq_escrever, "")
var _acrescentar = hrp.NewMetodoOuPanic("acrescentar", met_arq_acrescentar, "")
var _remover = hrp.NewMetodoOuPanic("remover", met_arq_remover, "")
var _renomear = hrp.NewMetodoOuPanic("renomear", met_arq_renomear, "")
var _caminhar = hrp.NewMetodoOuPanic("caminhar", met_arq_caminhar, "")
var _juntar = hrp.NewMetodoOuPanic("juntar", met_arq_juntar, "")
var _resolver = hrp.NewMetodoOuPanic("resolver", met_arq_resolver, "")

func init() {
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
