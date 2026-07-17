// Package sistema implementa o módulo nativo de informações e propriedades de ambiente
// da biblioteca padrão do Harpia.
//
// Este pacote permite que scripts em Harpia consultem detalhes dinâmicos sobre
// a arquitetura de processador e o sistema operacional hospedeiro no qual o interpretador está rodando.
package sistema

import (
	"runtime"

	"github.com/mat-dgruber/Harpia/ptst"
)

func init() {
	// constantes expõe metadados estáticos do runtime do Go convertidos para ptst.Texto.
	constantes := ptst.Mapa{
		// ARQUITETURA expõe a arquitetura de CPU onde o interpretador está compilado (ex: "amd64", "arm64").
		"ARQUITETURA": ptst.Texto(runtime.GOARCH),

		// NOME expõe o identificador padrão do sistema operacional do computador hospedeiro (ex: "darwin", "linux", "windows").
		"NOME": ptst.Texto(runtime.GOOS),
	}

	// metodos é inicializado vazio, reservado para futuras expansões e comandos do sistema operacional (ex: 'saida', 'executa_comando').
	metodos := []*ptst.Metodo{}

	// Registra o módulo 'sistema' na lista interna de módulos nativos do Harpia.
	ptst.RegistraModuloImpl(
		&ptst.ModuloImpl{
			Info: ptst.ModuloInfo{
				Nome:    "sistema",
				Arquivo: "stdlib/sistema",
			},
			Constantes: constantes,
			Metodos:    metodos,
		},
	)
}
