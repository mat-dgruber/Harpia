package hrp

// ObjetoGC define os comportamentos esperados de um objeto que participa da gerência de contagem de referências.
type ObjetoGC interface {
	Objeto
	Reter()
	Liberar()
	ObterRefs() int
	ObterFilhos() []Objeto
}

// GCMixin é uma estrutura leve e embutível para dar suporte nativo a contagem de referências e listagem de dependências de forma fail-safe.
type GCMixin struct {
	RefsCount int
}

// Reter incrementa o contador de referências se o objeto não for imutável/singleton (-1).
func (g *GCMixin) Reter() {
	if g.RefsCount == -1 {
		return
	}
	g.RefsCount++
}

// Liberar decrementa o contador de referências.
func (g *GCMixin) Liberar() {
	if g.RefsCount == -1 {
		return
	}
	g.RefsCount--
}

// ObterRefs retorna a quantidade ativa de referências do objeto.
func (g *GCMixin) ObterRefs() int {
	return g.RefsCount
}

// ObterFilhos padrão para o mixin (retorna nenhum filho por padrão, deve ser sobrescrito).
func (g *GCMixin) ObterFilhos() []Objeto {
	return nil
}

// ReterObjeto tenta realizar o incremento de referências de forma segura caso o objeto satisfaça a interface ObjetoGC.
func ReterObjeto(obj Objeto) {
	if gcObj, ok := obj.(ObjetoGC); ok && gcObj != nil {
		gcObj.Reter()
	}
}

// LiberarObjeto tenta realizar o decremento de referências de forma segura caso o objeto satisfaça a interface ObjetoGC.
func LiberarObjeto(obj Objeto) {
	if gcObj, ok := obj.(ObjetoGC); ok && gcObj != nil {
		gcObj.Liberar()
	}
}

// ColetarCiclos realiza a varredura "Trial Deletion" a partir do escopo fornecido,
// identificando referências circulares isoladas e quebrando-as para liberar memória.
func ColetarCiclos(escopo *Escopo) {
	if escopo == nil {
		return
	}

	// 1. Coleta todos os objetos GC alcançáveis a partir das variáveis do escopo
	alcancaveis := make(map[ObjetoGC]bool)
	var varrer func(obj Objeto)
	varrer = func(obj Objeto) {
		if gcObj, ok := obj.(ObjetoGC); ok && gcObj != nil {
			if alcancaveis[gcObj] {
				return
			}
			alcancaveis[gcObj] = true
			for _, filho := range gcObj.ObterFilhos() {
				varrer(filho)
			}
		}
	}

	for _, simb := range escopo.ObterSimbolosSeguro() {
		varrer(simb.ObterValor())
	}

	if len(alcancaveis) == 0 {
		return
	}

	// 2. Copia as referências reais para o mapa temporário
	refsTemp := make(map[ObjetoGC]int)
	for obj := range alcancaveis {
		refsTemp[obj] = obj.ObterRefs()
	}

	// 3. Simula decremento: reduz referências provenientes das conexões internas do grafo
	for obj := range alcancaveis {
		for _, filho := range obj.ObterFilhos() {
			if gcFilho, ok := filho.(ObjetoGC); ok && gcFilho != nil {
				if _, existe := refsTemp[gcFilho]; existe {
					refsTemp[gcFilho]--
				}
			}
		}
	}

	// 4. Se a contagem temporária chegou a 0, significa que o objeto é referenciado apenas por conexões internas do ciclo.
	// Reunimos todos os alvos elegíveis para quebra.
	var alvosParaQuebrar []ObjetoGC
	for obj, tempRefs := range refsTemp {
		if tempRefs == 0 && obj.ObterRefs() > 0 {
			alvosParaQuebrar = append(alvosParaQuebrar, obj)
		}
	}

	// 5. Quebramos todos os ciclos circulares detectados de forma simétrica e fail-safe
	for _, obj := range alvosParaQuebrar {
		switch v := obj.(type) {
		case *Lista:
			for _, item := range v.Itens {
				LiberarObjeto(item)
			}
			v.Itens = nil
		case Mapa:
			for _, item := range v {
				LiberarObjeto(item)
			}
			for k := range v {
				delete(v, k)
			}
		}
	}
}
