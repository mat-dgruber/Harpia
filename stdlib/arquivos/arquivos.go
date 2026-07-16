package arquivos

import (
	"io/fs"
	"os"
	"path/filepath"

	"github.com/natanfeitosa/portuscript/ptst"
)

func verificarPermissao(inst ptst.Objeto) error {
	if modulo, ok := inst.(*ptst.Modulo); ok && modulo != nil {
		return modulo.Contexto.VerificarPermissaoArquivos()
	}
	return nil
}

// met_arq_ler implementa 'ler(caminho)' -> retorna Texto com o conteúdo do arquivo
func met_arq_ler(inst ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
	if err := verificarPermissao(inst); err != nil {
		return nil, err
	}

	if err := ptst.VerificaNumeroArgumentos("ler", false, args, 1, 1); err != nil {
		return nil, err
	}

	caminho, err := ptst.NewTexto(args[0])
	if err != nil {
		return nil, err
	}

	bytes, err := os.ReadFile(string(caminho.(ptst.Texto)))
	if err != nil {
		return nil, ptst.NewErroF(ptst.ErroDeSistema, "Erro ao ler arquivo '%s': %v", caminho, err)
	}

	return ptst.Texto(bytes), nil
}

// met_arq_escrever implementa 'escrever(caminho, conteudo)' -> escreve texto ou bytes no arquivo
func met_arq_escrever(inst ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
	if err := verificarPermissao(inst); err != nil {
		return nil, err
	}

	if err := ptst.VerificaNumeroArgumentos("escrever", false, args, 2, 2); err != nil {
		return nil, err
	}

	caminho, err := ptst.NewTexto(args[0])
	if err != nil {
		return nil, err
	}

	conteudo, err := ptst.NewTexto(args[1])
	if err != nil {
		return nil, err
	}

	err = os.WriteFile(string(caminho.(ptst.Texto)), []byte(conteudo.(ptst.Texto)), 0644)
	if err != nil {
		return nil, ptst.NewErroF(ptst.ErroDeSistema, "Erro ao escrever no arquivo '%s': %v", caminho, err)
	}

	return ptst.Nulo, nil
}

// met_arq_acrescentar implementa 'acrescentar(caminho, conteudo)' -> anexa texto ao final do arquivo
func met_arq_acrescentar(inst ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
	if err := verificarPermissao(inst); err != nil {
		return nil, err
	}

	if err := ptst.VerificaNumeroArgumentos("acrescentar", false, args, 2, 2); err != nil {
		return nil, err
	}

	caminho, err := ptst.NewTexto(args[0])
	if err != nil {
		return nil, err
	}

	conteudo, err := ptst.NewTexto(args[1])
	if err != nil {
		return nil, err
	}

	f, err := os.OpenFile(string(caminho.(ptst.Texto)), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, ptst.NewErroF(ptst.ErroDeSistema, "Erro ao abrir arquivo '%s' para acrescentar: %v", caminho, err)
	}
	defer f.Close()

	if _, err := f.Write([]byte(conteudo.(ptst.Texto))); err != nil {
		return nil, ptst.NewErroF(ptst.ErroDeSistema, "Erro ao acrescentar dados no arquivo '%s': %v", caminho, err)
	}

	return ptst.Nulo, nil
}

// met_arq_remover implementa 'remover(caminho)' -> exclui um arquivo ou diretório
func met_arq_remover(inst ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
	if err := verificarPermissao(inst); err != nil {
		return nil, err
	}

	if err := ptst.VerificaNumeroArgumentos("remover", false, args, 1, 1); err != nil {
		return nil, err
	}

	caminho, err := ptst.NewTexto(args[0])
	if err != nil {
		return nil, err
	}

	err = os.RemoveAll(string(caminho.(ptst.Texto)))
	if err != nil {
		return nil, ptst.NewErroF(ptst.ErroDeSistema, "Erro ao remover '%s': %v", caminho, err)
	}

	return ptst.Nulo, nil
}

// met_arq_renomear implementa 'renomear(antigo, novo)' -> renomeia ou move arquivo/diretório
func met_arq_renomear(inst ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
	if err := verificarPermissao(inst); err != nil {
		return nil, err
	}

	if err := ptst.VerificaNumeroArgumentos("renomear", false, args, 2, 2); err != nil {
		return nil, err
	}

	antigo, err := ptst.NewTexto(args[0])
	if err != nil {
		return nil, err
	}

	novo, err := ptst.NewTexto(args[1])
	if err != nil {
		return nil, err
	}

	err = os.Rename(string(antigo.(ptst.Texto)), string(novo.(ptst.Texto)))
	if err != nil {
		return nil, ptst.NewErroF(ptst.ErroDeSistema, "Erro ao renomear '%s' para '%s': %v", antigo, novo, err)
	}

	return ptst.Nulo, nil
}

// met_arq_caminhar implementa 'caminhar(diretorio)' -> retorna Lista contendo caminhos encontrados recursivamente
func met_arq_caminhar(inst ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
	if err := verificarPermissao(inst); err != nil {
		return nil, err
	}

	if err := ptst.VerificaNumeroArgumentos("caminhar", false, args, 1, 1); err != nil {
		return nil, err
	}

	diretorio, err := ptst.NewTexto(args[0])
	if err != nil {
		return nil, err
	}

	lista := &ptst.Lista{Itens: make([]ptst.Objeto, 0)}

	err = filepath.WalkDir(string(diretorio.(ptst.Texto)), func(caminho string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		lista.Adiciona(ptst.Texto(caminho))
		return nil
	})

	if err != nil {
		return nil, ptst.NewErroF(ptst.ErroDeSistema, "Erro ao caminhar no diretório '%s': %v", diretorio, err)
	}

	return lista, nil
}

// met_arq_juntar implementa 'juntar(a, b, ...)' -> concatena caminhos lógicos usando os separadores corretos do SO
func met_arq_juntar(inst ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
	if len(args) == 0 {
		return ptst.Texto(""), nil
	}

	partes := make([]string, len(args))
	for i, arg := range args {
		txt, err := ptst.NewTexto(arg)
		if err != nil {
			return nil, err
		}
		partes[i] = string(txt.(ptst.Texto))
	}

	res := filepath.Join(partes...)
	return ptst.Texto(res), nil
}

// met_arq_resolver implementa 'resolver(caminho)' -> retorna o caminho físico absoluto unificado
func met_arq_resolver(inst ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
	if err := verificarPermissao(inst); err != nil {
		return nil, err
	}

	if err := ptst.VerificaNumeroArgumentos("resolver", false, args, 1, 1); err != nil {
		return nil, err
	}

	caminho, err := ptst.NewTexto(args[0])
	if err != nil {
		return nil, err
	}

	res, err := filepath.Abs(string(caminho.(ptst.Texto)))
	if err != nil {
		return nil, ptst.NewErroF(ptst.ErroDeSistema, "Erro ao resolver caminho absoluto de '%s': %v", caminho, err)
	}

	return ptst.Texto(res), nil
}

var _ler = ptst.NewMetodoOuPanic("ler", met_arq_ler, "")
var _escrever = ptst.NewMetodoOuPanic("escrever", met_arq_escrever, "")
var _acrescentar = ptst.NewMetodoOuPanic("acrescentar", met_arq_acrescentar, "")
var _remover = ptst.NewMetodoOuPanic("remover", met_arq_remover, "")
var _renomear = ptst.NewMetodoOuPanic("renomear", met_arq_renomear, "")
var _caminhar = ptst.NewMetodoOuPanic("caminhar", met_arq_caminhar, "")
var _juntar = ptst.NewMetodoOuPanic("juntar", met_arq_juntar, "")
var _resolver = ptst.NewMetodoOuPanic("resolver", met_arq_resolver, "")

func init() {
	ptst.RegistraModuloImpl(&ptst.ModuloImpl{
		Info: ptst.ModuloInfo{
			Nome:    "arquivos",
			Arquivo: "stdlib/arquivos",
		},
		Metodos: []*ptst.Metodo{
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
