package json

import (
	"encoding/json"

	"github.com/natanfeitosa/portuscript/ptst"
)

// converteParaGo converte recursivamente objetos Portuscript para tipos Go primitivos analisáveis pelo encoding/json.
func converteParaGo(obj ptst.Objeto) any {
	if obj == nil || obj == ptst.Nulo {
		return nil
	}
	switch v := obj.(type) {
	case ptst.Texto:
		return string(v)
	case ptst.Inteiro:
		return int64(v)
	case ptst.Decimal:
		return float64(v)
	case ptst.Booleano:
		return bool(v)
	case *ptst.Lista:
		res := make([]any, len(v.Itens))
		for i, item := range v.Itens {
			res[i] = converteParaGo(item)
		}
		return res
	case ptst.Mapa:
		res := make(map[string]any)
		for chave, val := range v {
			res[chave] = converteParaGo(val)
		}
		return res
	}
	return nil
}

// converteParaPtst converte recursivamente dados tipados de volta em tipos Portuscript nativos.
func converteParaPtst(dados any) ptst.Objeto {
	if dados == nil {
		return ptst.Nulo
	}
	switch v := dados.(type) {
	case string:
		return ptst.Texto(v)
	case float64:
		// JSON interpreta números como float64. Se não contiver ponto, convertemos para inteiro por comodidade
		if float64(int64(v)) == v {
			return ptst.Inteiro(int64(v))
		}
		return ptst.Decimal(v)
	case bool:
		return ptst.Booleano(v)
	case []any:
		lista := &ptst.Lista{Itens: make([]ptst.Objeto, len(v))}
		for i, item := range v {
			lista.Itens[i] = converteParaPtst(item)
		}
		return lista
	case map[string]any:
		mapa := ptst.NewMapaVazio()
		for k, val := range v {
			mapa.M__define_item__(ptst.Texto(k), converteParaPtst(val))
		}
		return mapa
	}
	return ptst.Nulo
}

// met_json_analisar implementa 'analisar(texto)' -> decodifica string JSON em objeto Portuscript
func met_json_analisar(inst ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
	if err := ptst.VerificaNumeroArgumentos("analisar", false, args, 1, 1); err != nil {
		return nil, err
	}

	texto, err := ptst.NewTexto(args[0])
	if err != nil {
		return nil, err
	}

	var dados any
	err = json.Unmarshal([]byte(texto.(ptst.Texto)), &dados)
	if err != nil {
		return nil, ptst.NewErroF(ptst.ValorErro, "Erro ao analisar string JSON: %v", err)
	}

	return converteParaPtst(dados), nil
}

// met_json_serializar implementa 'serializar(objeto)' -> codifica objeto Portuscript em string JSON
func met_json_serializar(inst ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
	if err := ptst.VerificaNumeroArgumentos("serializar", false, args, 1, 1); err != nil {
		return nil, err
	}

	goObj := converteParaGo(args[0])
	bytes, err := json.Marshal(goObj)
	if err != nil {
		return nil, ptst.NewErroF(ptst.ValorErro, "Erro ao serializar objeto para JSON: %v", err)
	}

	return ptst.Texto(bytes), nil
}

var _analisar = ptst.NewMetodoOuPanic("analisar", met_json_analisar, "")
var _serializar = ptst.NewMetodoOuPanic("serializar", met_json_serializar, "")

func init() {
	ptst.RegistraModuloImpl(&ptst.ModuloImpl{
		Info: ptst.ModuloInfo{
			Nome:    "json",
			Arquivo: "stdlib/json",
		},
		Metodos: []*ptst.Metodo{
			_analisar,
			_serializar,
		},
	})
}
