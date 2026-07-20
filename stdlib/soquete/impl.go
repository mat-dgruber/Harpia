package soquete

import (
	"fmt"
	"net"
	"syscall"
	"time"

	"github.com/mat-dgruber/Harpia/hrp"
	"golang.org/x/sys/unix"
)

// ultimoElemento é um utilitário genérico que retorna de forma segura a referência ao último item de um slice.
func ultimoElemento[T any](slice []T) T {
	if len(slice) == 0 {
		var none T
		return none
	}
	return slice[len(slice)-1]
}

// Soquete representa o objeto encapsulador de um descritor de arquivo de socket de rede na VM do Harpia.
type Soquete struct {
	// descritorDoSoquete armazena o manipulador de baixo nível (File Descriptor) gerenciado pelo kernel do SO.
	descritorDoSoquete int

	// Metadados sobre a família de IPs, o tipo de transporte e o protocolo de controle da conexão.
	familia, tipo, protocolo hrp.Inteiro

	// fechado é uma flag que impede tentativas duplicadas de fechar descritores de sockets já encerrados.
	fechado hrp.Booleano

	// pollFd guarda as especificações de monitoramento de eventos de E/S assíncronas do socket.
	pollFd []unix.PollFd

	// p é um ponteiro de soquete pai (usado para encadeamento e rastreio de sockets filhos resultantes de aceita()).
	p *Soquete
}

// TipoSoquete define as propriedades, o manual de inicialização e as assinaturas de classe para a classe Soquete.
var TipoSoquete = hrp.TipoObjeto.NewTipo(
	"Soquete",
	`Soquete(familia, tipo) -> Soquete
Cria um novo soquete usando a família de endereços, o tipo de soquete e o número de protocolo fornecidos.`,
)

var _ hrp.Objeto = (*Soquete)(nil)

// Tipo retorna a assinatura de classe da estrutura Soquete para a VM.
func (s *Soquete) Tipo() *hrp.Tipo {
	return TipoSoquete
}

// NewSoquete aloca uma chamada de sistema (unix.Socket) e retorna a instância de objeto Soquete correspondente.
//
// Retorna um erro detalhado se a família de IP não for suportada ou se a chamada de sistema do kernel falhar.
func NewSoquete(familia, tipo, protocolo hrp.Inteiro) (hrp.Objeto, error) {
	fd, err := unix.Socket(int(familia), int(tipo), int(protocolo))
	if err != nil {
		if err == unix.EAFNOSUPPORT {
			return nil, hrp.NewErroF(hrp.ValorErro, "Família de endereço não suportada")
		}

		return nil, hrp.NewErroF(hrp.ErroDeSistema, "Erro externo: %s", err)
	}

	s := &Soquete{descritorDoSoquete: fd, familia: familia, tipo: tipo, protocolo: protocolo, fechado: hrp.Falso}
	s.pollFd = []unix.PollFd{{Fd: int32(fd), Events: unix.POLLIN}} // Inicializa o pollFd para eventos de leitura (leitura pendente).

	return s, nil
}

// DefinirNaoBloqueante altera as propriedades de espera de E/S do socket.
//
// Se definido como verdadeiro, as operações de leitura e escrita retornam imediatamente em vez de bloquear o processo.
func (s *Soquete) DefinirNaoBloqueante(naobloqueante hrp.Booleano) (hrp.Objeto, error) {
	if err := unix.SetNonblock(s.descritorDoSoquete, bool(naobloqueante)); err != nil {
		panic(err)
	}

	return hrp.Nulo, nil
}

// DefineOpcoes configura opções do nível de soquete (via unix.SetsockoptInt, ex: reuso de portas com SO_REUSEADDR).
func (s *Soquete) DefineOpcoes(nivel, opcao, valor hrp.Inteiro) (hrp.Objeto, error) {
	if err := unix.SetsockoptInt(s.descritorDoSoquete, int(nivel), int(opcao), int(valor)); err != nil {
		panic(fmt.Sprintf("Erro ao definir opções do socket: %s", err))
	}

	return hrp.Nulo, nil
}

// Fecha encerra o soquete e libera o File Descriptor do sistema de forma segura.
func (s *Soquete) Fecha() (hrp.Objeto, error) {
	if !s.fechado {
		s.fechado = hrp.Verdadeiro

		if err := unix.Close(s.descritorDoSoquete); err != nil {
			return nil, hrp.NewErroF(hrp.ErroDeSistema, "Erro ao fechar soquete: %s", err)
		}
	}

	return hrp.Nulo, nil
}

// AssociaSoquete executa o bind (ligação) do soquete a uma interface de IP específica e porta no computador local.
func (s *Soquete) AssociaSoquete(ip hrp.Texto, porta hrp.Inteiro) (hrp.Objeto, error) {
	addr := &unix.SockaddrInet4{Port: int(porta)}
	copy(addr.Addr[:], net.ParseIP(string(ip)).To16())

	if err := unix.Bind(s.descritorDoSoquete, addr); err != nil {
		return nil, hrp.NewErroF(hrp.ErroDeSistema, "Erro ao associar soquete: %s", err)
	}

	return hrp.Nulo, nil
}

// OuveSoquete ativa a escuta (listen) do soquete por conexões entrantes com uma fila de backlog definida.
func (s *Soquete) OuveSoquete(backlog hrp.Inteiro) (hrp.Objeto, error) {
	if err := unix.Listen(s.descritorDoSoquete, int(backlog)); err != nil {
		return nil, hrp.NewErroF(hrp.ErroDeSistema, "Erro ao ouvir soquete: %s", err)
	}

	return hrp.Nulo, nil
}

// AceitaConexao aguarda (usando unix.Poll para evitar bloqueio estéril) e aceita uma nova conexão entrante no socket de escuta.
//
// Retorna um novo objeto de classe Soquete dedicado à troca de dados com o cliente conectado.
func (s *Soquete) AceitaConexao() (*Soquete, error) {
	for {
		_, err := unix.Poll(s.pollFd, 1000)
		if err != nil {
			return nil, hrp.NewErroF(hrp.ErroDeSistema, "Erro no poll: %s", err)
		}

		if s.pollFd[0].Revents&unix.POLLIN != 0 {
			fd, _, err := unix.Accept(s.descritorDoSoquete)
			if err != nil {
				return nil, hrp.NewErroF(hrp.ErroDeSistema, "Erro ao aceitar conexão: %s", err)
			}

			s.pollFd = append(s.pollFd, unix.PollFd{Fd: int32(fd), Events: unix.POLLIN})

			soq := &Soquete{
				descritorDoSoquete: fd,
				familia:            s.familia,
				tipo:               s.tipo,
				protocolo:          s.protocolo,
				fechado:            hrp.Falso,
				pollFd:             []unix.PollFd{{Fd: int32(fd), Events: unix.POLLIN}},
				p:                  s,
			}
			return soq, nil
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// RecebeDados lê bytes de dados do socket ativo em um buffer do tamanho máximo especificado.
func (s *Soquete) RecebeDados(tamanhoBuffer hrp.Inteiro) (*hrp.Bytes, error) {
	buffer := make([]byte, int(tamanhoBuffer))

	for {
		n, err := unix.Poll(s.pollFd, 1)
		if err != nil {
			if err == unix.EINTR {
				break
			}
			return nil, hrp.NewErroF(hrp.ErroDeSistema, "Erro no poll: %s", err)
		}

		if n > 0 && ultimoElemento(s.pollFd).Revents&unix.POLLIN != 0 {
			n, _, err := unix.Recvfrom(s.descritorDoSoquete, buffer, 0)
			if err != nil {
				return nil, hrp.NewErroF(hrp.ErroDeSistema, "Erro ao receber dados: %v", err)
			}

			if n <= 0 {
				return &hrp.Bytes{}, nil
			}
			val := &hrp.Bytes{Itens: buffer[:n]}
			return val, nil
		}
	}

	return &hrp.Bytes{}, nil
}

// EnviaDados escreve um array de bytes no socket de conexão de rede de forma a enviá-lo ao destinatário.
func (s *Soquete) EnviaDados(dados *hrp.Bytes) (hrp.Objeto, error) {
	_, err := unix.Write(s.descritorDoSoquete, dados.Itens)
	if err != nil {
		return nil, hrp.NewErroF(hrp.ErroDeSistema, "erro ao enviar dados: %v", err)
	}

	return hrp.Nulo, nil
}

// Conecta executa a conexão TCP a um host e porta remota fornecidos.
func (s *Soquete) Conecta(endereco hrp.Texto, porta hrp.Inteiro) (hrp.Objeto, error) {
	addr, err := s.resolveEndereco(string(endereco), int(porta))
	if err != nil {
		return nil, err
	}

	if err := unix.Connect(s.descritorDoSoquete, addr); err != nil {
		return nil, hrp.NewErroF(hrp.ErroDeSistema, "erro ao conectar ao servidor: %v", err)
	}

	return hrp.Nulo, nil
}

// resolveEndereco traduz de forma inteligente endereços de hosts (ex: "google.com") para endereços IP.
func (s *Soquete) resolveEndereco(endereco string, porta int) (unix.Sockaddr, error) {
	ips, err := net.LookupIP(string(endereco))
	if err != nil {
		return nil, hrp.NewErroF(hrp.ErroDeSistema, "Erro ao resolver o endereço: %v", err)
	}

	if len(ips) == 0 {
		return nil, hrp.NewErroF(hrp.ErroDeSistema, "Nenhum endereço IP encontrado para: %s", endereco)
	}

	for _, ip := range ips {
		switch s.familia {
		case syscall.AF_INET6:
			if ip6 := ip.To16(); ip6 != nil && ip.To4() == nil {
				addr := &unix.SockaddrInet6{Port: porta}
				copy(addr.Addr[:], ip6)
				return addr, nil
			}
		case syscall.AF_INET:
			if ip4 := ip.To4(); ip4 != nil {
				addr := &unix.SockaddrInet4{Port: porta}
				copy(addr.Addr[:], ip)
				return addr, nil
			}
		}
	}

	return nil, nil
}

func init() {
	// Nova é a função construtora para instanciar Sockets a partir de scripts Harpia.
	TipoSoquete.Nova = func(args hrp.Tupla) (hrp.Objeto, error) {
		if argsLen := len(args); argsLen != 3 {
			if argsLen < 2 {
				return nil, hrp.NewErroF(hrp.TipagemErro, "Soquete() esperava receber no mínimo 2 argumentos, mas recebeu %d", argsLen)
			}

			if argsLen > 3 {
				return nil, hrp.NewErroF(hrp.TipagemErro, "Soquete() esperava receber no máximo 3 argumentos, mas recebeu %d", argsLen)
			}
		}

		var familia, tipo, protocolo hrp.Inteiro = args[0].(hrp.Inteiro), args[1].(hrp.Inteiro), hrp.Inteiro(0)

		if len(args) == 3 {
			protocolo = args[2].(hrp.Inteiro)
		}

		return NewSoquete(familia, tipo, protocolo)
	}

	// Registro de todos os métodos de instância do Soquete no mapa de tipos da classe.

	TipoSoquete.Mapa["associa"] = hrp.NewMetodoOuPanic("associa", func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
		if err := hrp.VerificaNumeroArgumentos("associa", true, args, 2, 2); err != nil {
			return nil, err
		}

		return inst.(*Soquete).AssociaSoquete(args[0].(hrp.Texto), args[1].(hrp.Inteiro))
	}, "soquete.associa(ip, porta) -> Nulo\n\nAssocia um soquete a um endereço IP e porta.")

	TipoSoquete.Mapa["ouve"] = hrp.NewMetodoOuPanic("ouve", func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
		if err := hrp.VerificaNumeroArgumentos("ouve", true, args, 0, 1); err != nil {
			return nil, err
		}

		backlog := hrp.Inteiro(0)
		if len(args) == 1 {
			backlog = args[0].(hrp.Inteiro)
		}

		return inst.(*Soquete).OuveSoquete(backlog)
	}, "soquete.ouve(backlog?) -> Nulo\n\nInicia a escuta por conexões em um soquete.\nSe não for passado o backlog, que é o número máximo de conexões pendentes na fila, por padrão será 1.")

	TipoSoquete.Mapa["aceita"] = hrp.NewMetodoOuPanic("aceita", func(inst hrp.Objeto) (hrp.Objeto, error) {
		return inst.(*Soquete).AceitaConexao()
	}, "soquete.aceita() -> Soquete\n\nAceita uma nova conexão em um soquete que está escutando e retorna o soquete referente ao cliente.")

	TipoSoquete.Mapa["fecha"] = hrp.NewMetodoOuPanic("fecha", func(inst hrp.Objeto) (hrp.Objeto, error) {
		return inst.(*Soquete).Fecha()
	}, "soquete.fecha() -> Nulo\n\nFecha o soquete.")

	TipoSoquete.Mapa["recebe"] = hrp.NewMetodoOuPanic("recebe", func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
		if err := hrp.VerificaNumeroArgumentos("recebe", true, args, 0, 1); err != nil {
			return nil, err
		}

		tamanhoBuffer := hrp.Inteiro(0)
		if len(args) == 1 {
			tamanhoBuffer = args[0].(hrp.Inteiro)
		}

		return inst.(*Soquete).RecebeDados(tamanhoBuffer)
	}, "soquete.recebe(tamanhoBuffer?) -> Bytes\n\nRecebe os dados de uma conexão e retorna no tipo Bytes\nSe não for definido um tamanho de buffer, o padrão será 0")

	TipoSoquete.Mapa["envia"] = hrp.NewMetodoOuPanic("envia", func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
		if err := hrp.VerificaNumeroArgumentos("envia", true, args, 1, 1); err != nil {
			return nil, err
		}

		return inst.(*Soquete).EnviaDados(args[0].(*hrp.Bytes))
	}, "soquete.envia(dados) -> Nulo\n\nEnvia um objeto do tipo Bytes para o outro lado da conexão")

	TipoSoquete.Mapa["conecta"] = hrp.NewMetodoOuPanic("conecta", func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
		if err := hrp.VerificaNumeroArgumentos("conecta", true, args, 2, 2); err != nil {
			return nil, err
		}

		return inst.(*Soquete).Conecta(args[0].(hrp.Texto), args[1].(hrp.Inteiro))
	}, "soquete.conecta(endereco, porta) -> Nulo\n\nSe conecta a um servidor pela porta e endereço informado.\nO endereço pode ser um IP ou nome de domínio como: exemplo.com")

	TipoSoquete.Mapa["def_nao_bloqueante"] = hrp.NewMetodoOuPanic("def_nao_bloqueante", func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
		if err := hrp.VerificaNumeroArgumentos("def_nao_bloqueante", true, args, 1, 1); err != nil {
			return nil, err
		}

		return inst.(*Soquete).DefinirNaoBloqueante(args[0].(hrp.Booleano))
	}, "soquete.def_nao_bloqueante(naoBloqueante) -> Nulo\n\nDefine se o soquete deve operar em modo não bloqueante")

	TipoSoquete.Mapa["define_opcoes"] = hrp.NewMetodoOuPanic("define_opcoes", func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
		if err := hrp.VerificaNumeroArgumentos("define_opcoes", true, args, 2, 3); err != nil {
			return nil, err
		}

		valor := hrp.Inteiro(1)

		if len(args) == 3 {
			valor = args[2].(hrp.Inteiro)
		}

		return inst.(*Soquete).DefineOpcoes(args[0].(hrp.Inteiro), args[1].(hrp.Inteiro), valor)
	}, "soquete.define_opcoes(nivel, opcao, valor) -> Nulo\n\nDefine opções para o soquete.")
}
