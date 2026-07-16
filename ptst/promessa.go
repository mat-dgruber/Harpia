package ptst

// Promessa representa um objeto de computação diferida/futura que armazena callbacks
// a serem executados assim que o valor final for resolvido ou rejeitado.
type Promessa struct {
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
// Dispara todos os ouvintes associados em background.
func (p *Promessa) Resolver(valor Objeto) {
	if p.resolvida || p.rejeitada {
		return
	}
	p.resolvida = true
	p.valor = valor

	for _, cb := range p.callbacks {
		cb(valor, nil)
	}
}

// Rejeitar define a promessa como falha contendo um erro associado.
// Dispara todos os ouvintes associados em background.
func (p *Promessa) Rejeitar(err error) {
	if p.resolvida || p.rejeitada {
		return
	}
	p.rejeitada = true
	p.erro = err

	for _, cb := range p.callbacks {
		cb(nil, err)
	}
}

// Registre registra um callback receptor de encerramento da computação assíncrona.
// Se a promessa já tiver concluída, executa imediatamente.
func (p *Promessa) Registre(cb func(Objeto, error)) {
	if p.resolvida {
		cb(p.valor, nil)
		return
	}
	if p.rejeitada {
		cb(nil, p.erro)
		return
	}
	p.callbacks = append(p.callbacks, cb)
}
