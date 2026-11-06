// Package ffi fornece recursos de FFI (Foreign Function Interface) portátil para o Harpia,
// permitindo carregar dinamicamente bibliotecas C compiled (.so, .dll, .dylib) e chamar assinaturas de funções síncronas.
package ffi

import (
	"os"

	"github.com/mat-dgruber/Harpia/hrp"
)

// Biblioteca representa uma referência aberta a uma biblioteca compartilhada dinâmica (.so, .dll, .dylib).
type Biblioteca struct {
	Caminho string
}

// TipoBiblioteca mapeia a classe Biblioteca na VM do Harpia.
var TipoBiblioteca = hrp.NewTipo("Biblioteca", "Referência para uma biblioteca dinâmica carregada (FFI)")

// Tipo retorna a representação na VM.
func (b *Biblioteca) Tipo() *hrp.Tipo {
	return TipoBiblioteca
}

// M__obtem_attributo__ mapeia o método obterFuncao() para buscar assinaturas exportadas da biblioteca dinâmica.
func (b *Biblioteca) M__obtem_attributo__(nome string) (hrp.Objeto, error) {
	switch nome {
	case "obterFuncao":
		// Retorna um objeto FuncaoFFI mapeado a um símbolo exportado na biblioteca.
		return hrp.NewMetodoOuPanic("obterFuncao", func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
			if err := hrp.VerificaNumeroArgumentos("obterFuncao", false, args, 1, 1); err != nil {
				return nil, err
			}
			funcNome, err := hrp.NewTexto(args[0])
			if err != nil {
				return nil, err
			}
			funcNomeStr := string(funcNome.(hrp.Texto))

			return &FuncaoFFI{
				Nome:       funcNomeStr,
				Biblioteca: b,
			}, nil
		}, "Obtém a referência de um símbolo (função) exportado pela biblioteca dinâmica."), nil
	}
	return nil, hrp.NewErroF(hrp.AtributoErro, "Atributo '%s' não existe na Biblioteca", nome)
}

// FuncaoFFI representa o símbolo carregado de uma função externa pronto para ser invocado.
type FuncaoFFI struct {
	Nome       string
	Biblioteca *Biblioteca
}

// TipoFuncaoFFI define a classe da Função FFI na VM.
var TipoFuncaoFFI = hrp.NewTipo("FuncaoFFI", "Função nativa importada via FFI")

// Tipo retorna o tipo correspondente na VM.
func (f *FuncaoFFI) Tipo() *hrp.Tipo {
	return TipoFuncaoFFI
}

// M__chame__ torna a função FFI chamável pelo interpretador.
// A chamada nativa real (ABI System V / Windows x64 via cgo ou dlopen) ainda não existe: em vez de
// fingir sucesso devolvendo Nulo, a chamada falha com erro explícito.
// ponytail: implementar com purego/dlopen quando houver demanda; até lá, erro honesto.
func (f *FuncaoFFI) M__chame__(args hrp.Tupla) (hrp.Objeto, error) {
	info, err := os.Stat(f.Biblioteca.Caminho)
	if err != nil || info.IsDir() {
		return nil, hrp.NewErroF(hrp.ErroDeSistema, "biblioteca dinâmica '%s' não pôde ser carregada ou não existe", f.Biblioteca.Caminho)
	}
	return nil, hrp.NewErroF(hrp.ErroDeSistema, "chamada FFI nativa para '%s' ainda não implementada nesta versão do Harpia", f.Nome)
}

// met_ffi_abrir implementa 'abrir(caminhoBiblioteca)' para iniciar o ciclo FFI.
func met_ffi_abrir(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if err := hrp.ExigirPermissao(inst, hrp.RecursoArquivos); err != nil {
		return nil, err
	}
	if err := hrp.VerificaNumeroArgumentos("abrir", false, args, 1, 1); err != nil {
		return nil, err
	}
	caminho, err := hrp.NewTexto(args[0])
	if err != nil {
		return nil, err
	}
	caminhoStr := string(caminho.(hrp.Texto))

	return &Biblioteca{Caminho: caminhoStr}, nil
}

func init() {
	// Registra o módulo 'ffi' na VM do Harpia.
	hrp.RegistraModuloImpl(&hrp.ModuloImpl{
		Info: hrp.ModuloInfo{
			Nome:    "ffi",
			Arquivo: "stdlib/ffi",
		},
		Metodos: []*hrp.Metodo{
			hrp.NewMetodoOuPanic("abrir", met_ffi_abrir, "Carrega uma biblioteca dinâmica (.so, .dll, .dylib) via FFI."),
		},
	})
}
