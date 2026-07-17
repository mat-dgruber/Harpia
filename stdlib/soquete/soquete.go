// Package soquete implementa suporte a conexões de rede de baixo nível (sockets) TCP/IP
// da biblioteca padrão do Harpia, utilizando chamadas de sistema Unix/POSIX.
//
// O pacote expõe chaves e constantes nativas de rede que permitem a criação
// de canais de transmissão de fluxo de dados (como TCP) ou pacotes datagramas (como UDP).
package soquete

import (
	"syscall"

	"github.com/mat-dgruber/Harpia/ptst"
)

// familia define a relação das constantes nativas de rede mapeadas a partir das syscalls do sistema operacional.
var familia = ptst.Mapa{
	// AF_INET define o protocolo de transporte IPv4 (Internet Protocol versão 4).
	"AF_INET":     ptst.Inteiro(syscall.AF_INET),

	// AF_INET6 define o protocolo de transporte IPv6 (Internet Protocol versão 6).
	"AF_INET6":    ptst.Inteiro(syscall.AF_INET6),

	// SOCK_STREAM representa um fluxo de dados contínuo, confiável e orientado a conexão (geralmente TCP).
	"SOCK_STREAM": ptst.Inteiro(syscall.SOCK_STREAM),

	// SOCK_DGRAM representa mensagens discretas (datagramas) de conexão não confiável e sem conexão (geralmente UDP).
	"SOCK_DGRAM":  ptst.Inteiro(syscall.SOCK_DGRAM),
}

func init() {
	// constantes define o dicionário de chaves exportadas do módulo soquete.
	constantes := ptst.Mapa{
		TipoSoquete.Nome: TipoSoquete, // Registra a classe do Objeto 'Soquete' no módulo.
	}
	constantes.Atualizar(familia, false)

	metodos := []*ptst.Metodo{}

	// Registra o módulo 'soquete' no interpretador para importação.
	ptst.RegistraModuloImpl(
		&ptst.ModuloImpl{
			Info: ptst.ModuloInfo{
				Nome:    "soquete",
				Arquivo: "stdlib/soquete",
			},
			Constantes: constantes,
			Metodos:    metodos,
		},
	)
}
