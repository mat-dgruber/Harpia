package ia

import (
	"fmt"
	"math"

	"github.com/mat-dgruber/Harpia/hrp"
)

func extrairSliceFloatIA(obj hrp.Objeto) ([]float64, error) {
	lista, ok := obj.(*hrp.Lista)
	if !ok {
		return nil, hrp.NewErroF(hrp.TipagemErro, "esperava uma Lista, mas recebeu %s", obj.Tipo().Nome)
	}
	res := make([]float64, len(lista.Itens))
	for i, elem := range lista.Itens {
		dec, err := hrp.NewDecimal(elem)
		if err != nil {
			return nil, fmt.Errorf("elemento no indice %d nao e um numero: %v", i, err)
		}
		res[i] = float64(dec.(hrp.Decimal))
	}
	return res, nil
}

func floatSliceParaListaIA(slice []float64) *hrp.Lista {
	elementos := make([]hrp.Objeto, len(slice))
	for i, v := range slice {
		elementos[i] = hrp.Decimal(v)
	}
	return &hrp.Lista{Itens: hrp.Tupla(elementos)}
}

func met_ia_sigmoide(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if err := hrp.VerificaNumeroArgumentos("sigmoide", false, args, 1, 1); err != nil {
		return nil, err
	}
	dec, err := hrp.NewDecimal(args[0])
	if err != nil {
		return nil, err
	}
	x := float64(dec.(hrp.Decimal))
	res := 1.0 / (1.0 + math.Exp(-x))
	return hrp.Decimal(res), nil
}

func met_ia_derivadaSigmoide(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if err := hrp.VerificaNumeroArgumentos("derivadaSigmoide", false, args, 1, 1); err != nil {
		return nil, err
	}
	dec, err := hrp.NewDecimal(args[0])
	if err != nil {
		return nil, err
	}
	s := float64(dec.(hrp.Decimal))
	return hrp.Decimal(s * (1.0 - s)), nil
}

func met_ia_relu(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if err := hrp.VerificaNumeroArgumentos("relu", false, args, 1, 1); err != nil {
		return nil, err
	}
	dec, err := hrp.NewDecimal(args[0])
	if err != nil {
		return nil, err
	}
	x := float64(dec.(hrp.Decimal))
	return hrp.Decimal(math.Max(0.0, x)), nil
}

func met_ia_softmax(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if err := hrp.VerificaNumeroArgumentos("softmax", false, args, 1, 1); err != nil {
		return nil, err
	}
	v, err := extrairSliceFloatIA(args[0])
	if err != nil {
		return nil, err
	}
	if len(v) == 0 {
		return floatSliceParaListaIA(v), nil
	}

	for _, x := range v {
		if math.IsNaN(x) || math.IsInf(x, 0) {
			return nil, hrp.NewErroF(hrp.ValorErro, "softmax nao aceita valores NaN ou Infinito")
		}
	}

	maxVal := v[0]
	for _, x := range v {
		if x > maxVal {
			maxVal = x
		}
	}

	expSoma := 0.0
	exps := make([]float64, len(v))
	for i, x := range v {
		exps[i] = math.Exp(x - maxVal)
		expSoma += exps[i]
	}

	if expSoma == 0.0 || math.IsNaN(expSoma) || math.IsInf(expSoma, 0) {
		return nil, hrp.NewErroF(hrp.ValorErro, "soma de exponenciais invalida no softmax")
	}

	res := make([]float64, len(v))
	for i := 0; i < len(v); i++ {
		res[i] = exps[i] / expSoma
	}
	return floatSliceParaListaIA(res), nil
}

func met_ia_similaridadeCosseno(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if err := hrp.VerificaNumeroArgumentos("similaridadeCosseno", false, args, 2, 2); err != nil {
		return nil, err
	}
	v1, err := extrairSliceFloatIA(args[0])
	if err != nil {
		return nil, err
	}
	v2, err := extrairSliceFloatIA(args[1])
	if err != nil {
		return nil, err
	}
	if len(v1) != len(v2) {
		return nil, hrp.NewErroF(hrp.SintaxeErro, "vetores devem ter a mesma dimensao para similaridadeCosseno")
	}

	dot := 0.0
	norm1 := 0.0
	norm2 := 0.0
	for i := 0; i < len(v1); i++ {
		dot += v1[i] * v2[i]
		norm1 += v1[i] * v1[i]
		norm2 += v2[i] * v2[i]
	}
	if norm1 == 0.0 || norm2 == 0.0 {
		return hrp.Decimal(0.0), nil
	}
	return hrp.Decimal(dot / (math.Sqrt(norm1) * math.Sqrt(norm2))), nil
}

func met_ia_normalizarMinMax(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if err := hrp.VerificaNumeroArgumentos("normalizarMinMax", false, args, 1, 1); err != nil {
		return nil, err
	}
	v, err := extrairSliceFloatIA(args[0])
	if err != nil {
		return nil, err
	}
	if len(v) == 0 {
		return floatSliceParaListaIA(v), nil
	}

	minVal, maxVal := v[0], v[0]
	for _, x := range v {
		if x < minVal {
			minVal = x
		}
		if x > maxVal {
			maxVal = x
		}
	}

	if maxVal == minVal {
		res := make([]float64, len(v))
		return floatSliceParaListaIA(res), nil
	}

	res := make([]float64, len(v))
	for i, x := range v {
		res[i] = (x - minVal) / (maxVal - minVal)
	}
	return floatSliceParaListaIA(res), nil
}

func met_ia_erroQuadraticoMedio(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if err := hrp.VerificaNumeroArgumentos("erroQuadraticoMedio", false, args, 2, 2); err != nil {
		return nil, err
	}
	yReal, err := extrairSliceFloatIA(args[0])
	if err != nil {
		return nil, err
	}
	yPred, err := extrairSliceFloatIA(args[1])
	if err != nil {
		return nil, err
	}
	if len(yReal) != len(yPred) {
		return nil, hrp.NewErroF(hrp.SintaxeErro, "vetores de erroQuadraticoMedio devem ter o mesmo tamanho")
	}
	if len(yReal) == 0 {
		return hrp.Decimal(0.0), nil
	}

	soma := 0.0
	for i := 0; i < len(yReal); i++ {
		diff := yReal[i] - yPred[i]
		soma += diff * diff
	}
	return hrp.Decimal(soma / float64(len(yReal))), nil
}

var _ia_sigmoide = hrp.NewMetodoOuPanic("sigmoide", met_ia_sigmoide, "sigmoide(x) -> Retorna a funcao de ativacao sigmoide 1/(1+e^-x)")
var _ia_derivadaSigmoide = hrp.NewMetodoOuPanic("derivadaSigmoide", met_ia_derivadaSigmoide, "derivadaSigmoide(s) -> Retorna a derivada s*(1-s)")
var _ia_relu = hrp.NewMetodoOuPanic("relu", met_ia_relu, "relu(x) -> Retorna max(0, x)")
var _ia_softmax = hrp.NewMetodoOuPanic("softmax", met_ia_softmax, "softmax(vetor) -> Retorna distribuicao de probabilidade softmax")
var _ia_similaridadeCosseno = hrp.NewMetodoOuPanic("similaridadeCosseno", met_ia_similaridadeCosseno, "similaridadeCosseno(v1, v2) -> Retorna a similaridade de cosseno entre dois vetores")
var _ia_normalizarMinMax = hrp.NewMetodoOuPanic("normalizarMinMax", met_ia_normalizarMinMax, "normalizarMinMax(vetor) -> Normaliza os valores do vetor para o intervalo [0, 1]")
var _ia_erroQuadraticoMedio = hrp.NewMetodoOuPanic("erroQuadraticoMedio", met_ia_erroQuadraticoMedio, "erroQuadraticoMedio(yReal, yPredito) -> Retorna o MSE entre dois vetores")
