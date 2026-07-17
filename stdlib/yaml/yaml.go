package yaml

import (
	"gopkg.in/yaml.v3"

	"github.com/mat-dgruber/Harpia/ptst"
)

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

func converteParaPtst(dados any) ptst.Objeto {
	if dados == nil {
		return ptst.Nulo
	}
	switch v := dados.(type) {
	case string:
		return ptst.Texto(v)
	case int:
		return ptst.Inteiro(int64(v))
	case int64:
		return ptst.Inteiro(v)
	case float64:
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

func met_yaml_analisar(inst ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
	if err := ptst.VerificaNumeroArgumentos("analisar", false, args, 1, 1); err != nil {
		return nil, err
	}

	texto, err := ptst.NewTexto(args[0])
	if err != nil {
		return nil, err
	}

	var dados any
	err = yaml.Unmarshal([]byte(texto.(ptst.Texto)), &dados)
	if err != nil {
		return nil, ptst.NewErroF(ptst.ValorErro, "Erro ao analisar string YAML: %v", err)
	}

	return converteParaPtst(dados), nil
}

func met_yaml_serializar(inst ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
	if err := ptst.VerificaNumeroArgumentos("serializar", false, args, 1, 1); err != nil {
		return nil, err
	}

	goObj := converteParaGo(args[0])
	bytes, err := yaml.Marshal(goObj)
	if err != nil {
		return nil, ptst.NewErroF(ptst.ValorErro, "Erro ao serializar objeto para YAML: %v", err)
	}

	return ptst.Texto(bytes), nil
}

func init() {
	ptst.RegistraModuloImpl(&ptst.ModuloImpl{
		Info: ptst.ModuloInfo{
			Nome:    "yaml",
			Arquivo: "stdlib/yaml",
		},
		Metodos: []*ptst.Metodo{
			ptst.NewMetodoOuPanic("analisar", met_yaml_analisar, ""),
			ptst.NewMetodoOuPanic("serializar", met_yaml_serializar, ""),
		},
	})
}
