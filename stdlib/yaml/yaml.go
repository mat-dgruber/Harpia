package yaml

import (
	"gopkg.in/yaml.v3"

	"github.com/mat-dgruber/Harpia/hrp"
)

func converteParaGo(obj hrp.Objeto) any {
	if obj == nil || obj == hrp.Nulo {
		return nil
	}
	switch v := obj.(type) {
	case hrp.Texto:
		return string(v)
	case hrp.Inteiro:
		return int64(v)
	case hrp.Decimal:
		return float64(v)
	case hrp.Booleano:
		return bool(v)
	case *hrp.Lista:
		res := make([]any, len(v.Itens))
		for i, item := range v.Itens {
			res[i] = converteParaGo(item)
		}
		return res
	case hrp.Mapa:
		res := make(map[string]any)
		for chave, val := range v {
			res[chave] = converteParaGo(val)
		}
		return res
	}
	return nil
}

func converteParahrp(dados any) hrp.Objeto {
	if dados == nil {
		return hrp.Nulo
	}
	switch v := dados.(type) {
	case string:
		return hrp.Texto(v)
	case int:
		return hrp.Inteiro(int64(v))
	case int64:
		return hrp.Inteiro(v)
	case float64:
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

func met_yaml_analisar(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if err := hrp.VerificaNumeroArgumentos("analisar", false, args, 1, 1); err != nil {
		return nil, err
	}

	texto, err := hrp.NewTexto(args[0])
	if err != nil {
		return nil, err
	}

	var dados any
	err = yaml.Unmarshal([]byte(texto.(hrp.Texto)), &dados)
	if err != nil {
		return nil, hrp.NewErroF(hrp.ValorErro, "Erro ao analisar string YAML: %v", err)
	}

	return converteParahrp(dados), nil
}

func met_yaml_serializar(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if err := hrp.VerificaNumeroArgumentos("serializar", false, args, 1, 1); err != nil {
		return nil, err
	}

	goObj := converteParaGo(args[0])
	bytes, err := yaml.Marshal(goObj)
	if err != nil {
		return nil, hrp.NewErroF(hrp.ValorErro, "Erro ao serializar objeto para YAML: %v", err)
	}

	return hrp.Texto(bytes), nil
}

func init() {
	hrp.RegistraModuloImpl(&hrp.ModuloImpl{
		Info: hrp.ModuloInfo{
			Nome:    "yaml",
			Arquivo: "stdlib/yaml",
		},
		Metodos: []*hrp.Metodo{
			hrp.NewMetodoOuPanic("analisar", met_yaml_analisar, ""),
			hrp.NewMetodoOuPanic("serializar", met_yaml_serializar, ""),
		},
	})
}
