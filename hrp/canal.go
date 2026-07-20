package hrp

import (
	"sync"
)

// Canal representa uma primitiva de comunicação reativa e concorrente baseada no modelo CSP (estilo Go).
//
// Permite que múltiplas corotinas (goroutines da VM) troquem mensagens de forma segura e sincronizada,
// sem a necessidade de acoplamento direto ou gerenciamento manual de travas de exclusão mútua.
type Canal struct {
	mu          sync.Mutex
	buffer      []Objeto              // Buffer de armazenamento temporário de mensagens (FIFO)
	recebedores []func(Objeto, error) // Fila de callbacks ativas esperando por mensagens (FIFO)
}

// TipoCanal especifica a assinatura e os metadados de classe do tipo Canal na VM.
var TipoCanal = NewTipo("Canal", "Primitiva de concorrência baseada em canais de comunicação (modelo CSP)")

func (c *Canal) Tipo() *Tipo {
	return TipoCanal
}

func init() {
	// Nova define o construtor da classe Canal para scripts Harpia (ex: novo Canal())
	TipoCanal.Nova = func(args Tupla) (Objeto, error) {
		return &Canal{
			buffer:      make([]Objeto, 0),
			recebedores: make([]func(Objeto, error), 0),
		}, nil
	}
}

// Enviar deposita uma mensagem no canal.
// Se houver algum receptor ativamente aguardando dados na fila de recebimento, entrega o dado diretamente (FIFO).
// Caso contrário, enfileira a mensagem no buffer interno.
func (c *Canal) Enviar(dado Objeto) {
	c.mu.Lock()
	if len(c.recebedores) > 0 {
		recebedor := c.recebedores[0]
		c.recebedores = c.recebedores[1:]
		c.mu.Unlock()

		// Dispara a resolução da promessa correspondente
		recebedor(dado, nil)
		return
	}

	c.buffer = append(c.buffer, dado)
	c.mu.Unlock()
}

// Receber retorna uma Promessa que é resolvida imediatamente com o dado mais antigo do buffer (FIFO),
// ou resolvida futuramente de forma cooperativa e assíncrona assim que um dado for enviado ao canal.
func (c *Canal) Receber() *Promessa {
	prom := NewPromessa()

	c.mu.Lock()
	if len(c.buffer) > 0 {
		dado := c.buffer[0]
		c.buffer = c.buffer[1:]
		c.mu.Unlock()

		prom.Resolver(dado)
		return prom
	}

	c.recebedores = append(c.recebedores, func(res Objeto, err error) {
		if err != nil {
			prom.Rejeitar(err)
		} else {
			prom.Resolver(res)
		}
	})
	c.mu.Unlock()

	return prom
}

func (c *Canal) M__obtem_attributo__(nome string) (Objeto, error) {
	switch nome {
	case "enviar":
		return NewMetodoOuPanic("enviar", func(inst Objeto, args Tupla) (Objeto, error) {
			if err := VerificaNumeroArgumentos("enviar", false, args, 1, 1); err != nil {
				return nil, err
			}
			c.Enviar(args[0])
			return Nulo, nil
		}, ""), nil
	case "receber":
		return NewMetodoOuPanic("receber", func(inst Objeto, args Tupla) (Objeto, error) {
			if err := VerificaNumeroArgumentos("receber", false, args, 0, 0); err != nil {
				return nil, err
			}
			return c.Receber(), nil
		}, ""), nil
	}
	return nil, NewErroF(AtributoErro, "Atributo '%s' não existe em Canal", nome)
}

// Garantia de conformidade com as interfaces de objeto estruturadas da VM
var _ I__obtem_attributo__ = (*Canal)(nil)
