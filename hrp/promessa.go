package hrp

import "sync"

// Promessa representa um objeto de computação diferida/futura que armazena callbacks
// a serem executados assim que o valor final for resolvido ou rejeitado.
//
// ponytail: Utiliza sync.Mutex para garantir thread-safety completo contra
// modificações e notificações concorrentes em ambientes multithread.
type Promessa struct {
	mu        sync.Mutex
	resolvida bool
	rejeitada bool
	valor     Objeto
	erro      error
	callbacks []func(Objeto, error)
}

var TipoPromessa = NewTipo("Promessa", "Representa um valor assíncrono que será resolvido no futuro")

func (p *Promessa) Tipo() *Tipo {
	return TipoPromessa
}

func NewPromessa() *Promessa {
	return &Promessa{
		callbacks: make([]func(Objeto, error), 0),
	}
}

// Resolver define a promessa como resolvida com sucesso contendo o objeto de retorno.
// Dispara todos os ouvintes associados.
func (p *Promessa) Resolver(valor Objeto) {
	p.mu.Lock()
	if p.resolvida || p.rejeitada {
		p.mu.Unlock()
		return
	}
	p.resolvida = true
	p.valor = valor
	cbs := append([]func(Objeto, error){}, p.callbacks...)
	p.mu.Unlock()

	for _, cb := range cbs {
		cb(valor, nil)
	}
}

// Rejeitar define a promessa como falha contendo um erro associado.
// Dispara todos os ouvintes associados.
func (p *Promessa) Rejeitar(err error) {
	p.mu.Lock()
	if p.resolvida || p.rejeitada {
		p.mu.Unlock()
		return
	}
	p.rejeitada = true
	p.erro = err
	cbs := append([]func(Objeto, error){}, p.callbacks...)
	p.mu.Unlock()

	for _, cb := range cbs {
		cb(nil, err)
	}
}

// Registre registra um callback receptor de encerramento da computação assíncrona.
// Se a promessa já tiver concluída, executa imediatamente.
func (p *Promessa) Registre(cb func(Objeto, error)) {
	p.mu.Lock()
	if p.resolvida {
		val := p.valor
		p.mu.Unlock()
		cb(val, nil)
		return
	}
	if p.rejeitada {
		e := p.erro
		p.mu.Unlock()
		cb(nil, e)
		return
	}
	p.callbacks = append(p.callbacks, cb)
	p.mu.Unlock()
}
