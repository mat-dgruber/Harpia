package json

import (
	"encoding/json"

	"github.com/mat-dgruber/Harpia/hrp"
)

// converteParahrp converte recursivamente dados tipados de volta em tipos Harpia nativos.
func converteParahrp(dados any) hrp.Objeto {
	if dados == nil {
		return hrp.Nulo
	}
	switch v := dados.(type) {
	case string:
		return hrp.Texto(v)
	case float64:
		// JSON interpreta números como float64. Se não contiver ponto, convertemos para inteiro por comodidade
		if float64(int64(v)) == v {
			return hrp.Inteiro(int64(v))
		}
		return hrp.Decimal(v)
	case bool:
		return hrp.Booleano(v)
	case []any:
		lista := &hrp.Lista{Itens: make([]hrp.Objeto, len(v))}
		for i, item := range v {
			lista.Itens[i] = converteParahrp(item)
		}
		return lista
	case map[string]any:
		mapa := hrp.NewMapaVazio()
		for k, val := range v {
			mapa.M__define_item__(hrp.Texto(k), converteParahrp(val))
		}
		return mapa
	}
	return hrp.Nulo
}

// met_json_analisar implementa 'analisar(texto)' -> decodifica string JSON em objeto Harpia
func met_json_analisar(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if err := hrp.VerificaNumeroArgumentos("analisar", false, args, 1, 1); err != nil {
		return nil, err
	}

	texto, err := hrp.NewTexto(args[0])
	if err != nil {
		return nil, err
	}

	var dados any
	err = json.Unmarshal([]byte(texto.(hrp.Texto)), &dados)
	if err != nil {
		return nil, hrp.NewErroF(hrp.ValorErro, "Erro ao analisar string JSON: %v", err)
	}

	return converteParahrp(dados), nil
}

// met_json_serializar implementa 'serializar(objeto)' -> codifica objeto Harpia em string JSON
func met_json_serializar(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if err := hrp.VerificaNumeroArgumentos("serializar", false, args, 1, 1); err != nil {
		return nil, err
	}

	goObj := hrp.ConverteParaGo(args[0])
	bytes, err := json.Marshal(goObj)
	if err != nil {
		return nil, hrp.NewErroF(hrp.ValorErro, "Erro ao serializar objeto para JSON: %v", err)
	}

	return hrp.Texto(bytes), nil
}

func init() {
	hrp.RegistraModuloImpl(&hrp.ModuloImpl{
		Info: hrp.ModuloInfo{
			Nome:    "json",
			Arquivo: "stdlib/json",
		},
		Metodos: []*hrp.Metodo{
			hrp.NewMetodoOuPanic("analisar", met_json_analisar, "Decodifica uma string JSON em um objeto Harpia."),
			hrp.NewMetodoOuPanic("serializar", met_json_serializar, "Codifica um objeto Harpia em uma string JSON."),
		},
	})
}
