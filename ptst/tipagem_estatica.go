package ptst

import (
	"strings"
)

// ValidarTipo realiza a verificação estrita de tipagem dinâmica.
// Retorna verdadeiro se o objeto for compatível com o tipo esperado indicado na assinatura/declaração.
func ValidarTipo(esperado string, obtido Objeto) bool {
	if esperado == "" || esperado == "qualquer" || esperado == "Objeto" {
		return true
	}

	if obtido == nil {
		return esperado == "Nulo"
	}

	nomeObtido := obtido.Tipo().Nome

	// Se o esperado contiver genéricos, como Lista<T> ou Mapa<K, V>
	if strings.Contains(esperado, "<") && strings.HasSuffix(esperado, ">") {
		idxOpen := strings.Index(esperado, "<")
		tipoBase := esperado[:idxOpen]
		tipoInternoStr := esperado[idxOpen+1 : len(esperado)-1]

		if tipoBase == "Lista" {
			if nomeObtido != "Lista" {
				return false
			}

			lista, ok := obtido.(*Lista)
			if !ok {
				return false
			}

			// Se a lista estiver vazia, assumimos como válida (lazy)
			if len(lista.Itens) == 0 {
				return true
			}

			// Valida cada item recursivamente
			for _, item := range lista.Itens {
				if !ValidarTipo(tipoInternoStr, item) {
					return false
				}
			}
			return true
		}

		if tipoBase == "Mapa" {
			if nomeObtido != "Mapa" {
				return false
			}

			mapa, ok := obtido.(Mapa)
			if !ok {
				return false
			}

			if len(mapa) == 0 {
				return true
			}

			partes := strings.SplitN(tipoInternoStr, ",", 2)
			if len(partes) != 2 {
				return true // Fallback fail-safe
			}
			tipoChave := strings.TrimSpace(partes[0])
			tipoValor := strings.TrimSpace(partes[1])

			for chave, valor := range mapa {
				txtChave, _ := NewTexto(chave)
				// Na VM do Harpia as chaves de mapas são strings (Texto)
				if !ValidarTipo(tipoChave, txtChave) {
					return false
				}
				if !ValidarTipo(tipoValor, valor) {
					return false
				}
			}
			return true
		}
	}

	// Mapeamentos diretos
	switch esperado {
	case "Inteiro":
		return nomeObtido == "Inteiro"
	case "Decimal":
		return nomeObtido == "Decimal"
	case "Texto":
		return nomeObtido == "Texto"
	case "Booleano":
		return nomeObtido == "Booleano" || nomeObtido == "Logico"
	case "Nulo":
		return nomeObtido == "Nulo"
	case "Tupla":
		return nomeObtido == "Tupla"
	case "funcao", "Funcao":
		return nomeObtido == "Funcao" || nomeObtido == "Metodo" || nomeObtido == "MetodoProxy"
	}

	// Caso especial: herança ou compatibilidade de nome direta
	return nomeObtido == esperado
}
