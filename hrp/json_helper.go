package hrp

// ConverteParaGo converte recursivamente objetos Harpia para tipos Go primitivos analisáveis pelo encoding/json.
func ConverteParaGo(obj Objeto) any {
	if obj == nil || obj == Nulo {
		return nil
	}
	switch v := obj.(type) {
	case Texto:
		return string(v)
	case Inteiro:
		return int64(v)
	case Decimal:
		return float64(v)
	case Booleano:
		return bool(v)
	case *Lista:
		res := make([]any, len(v.Itens))
		for i, item := range v.Itens {
			res[i] = ConverteParaGo(item)
		}
		return res
	case Mapa:
		res := make(map[string]any)
		for chave, val := range v {
			res[chave] = ConverteParaGo(val)
		}
		return res
	case *Instancia:
		res := make(map[string]any)
		for chave, val := range v.Atributos {
			res[chave] = ConverteParaGo(val)
		}
		return res
	}
	return nil
}

// ConverteDeGo converte recursivamente tipos Go primitivos (resultado de json.Unmarshal) para objetos Harpia.
func ConverteDeGo(val any) Objeto {
	if val == nil {
		return Nulo
	}
	switch v := val.(type) {
	case string:
		return Texto(v)
	case float64:
		// json.Unmarshal representa todos os números como float64
		if v == float64(int64(v)) {
			return Inteiro(int64(v))
		}
		return Decimal(v)
	case bool:
		return Booleano(v)
	case []any:
		lista := &Lista{}
		for _, item := range v {
			lista.Itens = append(lista.Itens, ConverteDeGo(item))
		}
		return lista
	case map[string]any:
		mapa := NewMapaVazio()
		for chave, item := range v {
			mapa.M__define_item__(Texto(chave), ConverteDeGo(item))
		}
		return mapa
	}
	return Nulo
}

