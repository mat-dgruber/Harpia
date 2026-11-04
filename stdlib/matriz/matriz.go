package matriz

import (
	"fmt"
	"math"

	"github.com/mat-dgruber/Harpia/hrp"
)

func init() {
	metodos := []*hrp.Metodo{
		_matriz_produtoEscalar,
		_matriz_somarVetores,
		_matriz_subtrairVetores,
		_matriz_multiplicarEscalar,
		_matriz_norma,
		_matriz_multiplicar,
		_matriz_criar,
	}

	hrp.RegistraModuloImpl(
		&hrp.ModuloImpl{
			Info: hrp.ModuloInfo{
				Nome:    "matriz",
				Arquivo: "stdlib/matriz",
			},
			Constantes: hrp.Mapa{},
			Metodos:    metodos,
		},
	)
}

func extrairSliceFloat(obj hrp.Objeto) ([]float64, error) {
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

func floatSliceParaLista(slice []float64) *hrp.Lista {
	elementos := make([]hrp.Objeto, len(slice))
	for i, v := range slice {
		elementos[i] = hrp.Decimal(v)
	}
	return &hrp.Lista{Itens: hrp.Tupla(elementos)}
}

func met_matriz_produtoEscalar(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if err := hrp.VerificaNumeroArgumentos("produtoEscalar", false, args, 2, 2); err != nil {
		return nil, err
	}
	v1, err := extrairSliceFloat(args[0])
	if err != nil {
		return nil, err
	}
	v2, err := extrairSliceFloat(args[1])
	if err != nil {
		return nil, err
	}
	if len(v1) != len(v2) {
		return nil, hrp.NewErroF(hrp.SintaxeErro, "vetores devem ter o mesmo tamanho (v1: %d, v2: %d)", len(v1), len(v2))
	}
	soma := 0.0
	for i := 0; i < len(v1); i++ {
		soma += v1[i] * v2[i]
	}
	return hrp.Decimal(soma), nil
}

func met_matriz_somarVetores(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if err := hrp.VerificaNumeroArgumentos("somarVetores", false, args, 2, 2); err != nil {
		return nil, err
	}
	v1, err := extrairSliceFloat(args[0])
	if err != nil {
		return nil, err
	}
	v2, err := extrairSliceFloat(args[1])
	if err != nil {
		return nil, err
	}
	if len(v1) != len(v2) {
		return nil, hrp.NewErroF(hrp.SintaxeErro, "vetores devem ter o mesmo tamanho (v1: %d, v2: %d)", len(v1), len(v2))
	}
	res := make([]float64, len(v1))
	for i := 0; i < len(v1); i++ {
		res[i] = v1[i] + v2[i]
	}
	return floatSliceParaLista(res), nil
}

func met_matriz_subtrairVetores(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if err := hrp.VerificaNumeroArgumentos("subtrairVetores", false, args, 2, 2); err != nil {
		return nil, err
	}
	v1, err := extrairSliceFloat(args[0])
	if err != nil {
		return nil, err
	}
	v2, err := extrairSliceFloat(args[1])
	if err != nil {
		return nil, err
	}
	if len(v1) != len(v2) {
		return nil, hrp.NewErroF(hrp.SintaxeErro, "vetores devem ter o mesmo tamanho (v1: %d, v2: %d)", len(v1), len(v2))
	}
	res := make([]float64, len(v1))
	for i := 0; i < len(v1); i++ {
		res[i] = v1[i] - v2[i]
	}
	return floatSliceParaLista(res), nil
}

func met_matriz_multiplicarEscalar(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if err := hrp.VerificaNumeroArgumentos("multiplicarEscalar", false, args, 2, 2); err != nil {
		return nil, err
	}
	v, err := extrairSliceFloat(args[0])
	if err != nil {
		return nil, err
	}
	kObj, err := hrp.NewDecimal(args[1])
	if err != nil {
		return nil, err
	}
	k := float64(kObj.(hrp.Decimal))

	res := make([]float64, len(v))
	for i := 0; i < len(v); i++ {
		res[i] = v[i] * k
	}
	return floatSliceParaLista(res), nil
}

func met_matriz_norma(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if err := hrp.VerificaNumeroArgumentos("norma", false, args, 1, 1); err != nil {
		return nil, err
	}
	v, err := extrairSliceFloat(args[0])
	if err != nil {
		return nil, err
	}
	soma := 0.0
	for _, x := range v {
		soma += x * x
	}
	return hrp.Decimal(math.Sqrt(soma)), nil
}

func extrairMatriz(obj hrp.Objeto) ([][]float64, error) {
	lista, ok := obj.(*hrp.Lista)
	if !ok {
		return nil, hrp.NewErroF(hrp.TipagemErro, "esperava uma Lista de Listas, mas recebeu %s", obj.Tipo().Nome)
	}
	linhas := len(lista.Itens)
	if linhas == 0 {
		return [][]float64{}, nil
	}
	m := make([][]float64, linhas)
	var cols int
	for i, elem := range lista.Itens {
		v, err := extrairSliceFloat(elem)
		if err != nil {
			return nil, err
		}
		if i == 0 {
			cols = len(v)
			if cols == 0 {
				return nil, hrp.NewErroF(hrp.ValorErro, "linha da matriz não pode ser vazia")
			}
		} else if len(v) != cols {
			return nil, hrp.NewErroF(hrp.ValorErro, "matriz irregular: linha %d tem tamanho %d, esperado %d", i, len(v), cols)
		}
		m[i] = v
	}
	return m, nil
}

func met_matriz_criar(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if err := hrp.VerificaNumeroArgumentos("criar", false, args, 2, 3); err != nil {
		return nil, err
	}
	linhasObj, err := hrp.NewDecimal(args[0])
	if err != nil {
		return nil, err
	}
	colsObj, err := hrp.NewDecimal(args[1])
	if err != nil {
		return nil, err
	}
	linhas := int(linhasObj.(hrp.Decimal))
	cols := int(colsObj.(hrp.Decimal))
	if linhas <= 0 || cols <= 0 {
		return nil, hrp.NewErroF(hrp.ValorErro, "dimensões da matriz devem ser positivas (linhas: %d, colunas: %d)", linhas, cols)
	}
	const maxElementos = 1 << 24
	if cols > maxElementos/linhas {
		return nil, hrp.NewErroF(hrp.ValorErro, "tamanho total da matriz excede limite permitido (%d)", maxElementos)
	}

	valorPadrao := hrp.Decimal(0.0)
	if len(args) > 2 {
		v, err := hrp.NewDecimal(args[2])
		if err != nil {
			return nil, err
		}
		valorPadrao = v.(hrp.Decimal)
	}

	linhasLista := make([]hrp.Objeto, linhas)
	for i := 0; i < linhas; i++ {
		colLista := make([]hrp.Objeto, cols)
		for j := 0; j < cols; j++ {
			colLista[j] = valorPadrao
		}
		linhasLista[i] = &hrp.Lista{Itens: hrp.Tupla(colLista)}
	}
	return &hrp.Lista{Itens: hrp.Tupla(linhasLista)}, nil
}

func met_matriz_multiplicar(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if err := hrp.VerificaNumeroArgumentos("multiplicar", false, args, 2, 2); err != nil {
		return nil, err
	}
	m1, err := extrairMatriz(args[0])
	if err != nil {
		return nil, err
	}
	m2, err := extrairMatriz(args[1])
	if err != nil {
		return nil, err
	}

	linhas1 := len(m1)
	if linhas1 == 0 {
		return &hrp.Lista{Itens: hrp.Tupla([]hrp.Objeto{})}, nil
	}
	linhas2 := len(m2)
	if linhas2 == 0 {
		return nil, hrp.NewErroF(hrp.SintaxeErro, "matriz m2 nao pode ser vazia")
	}

	cols1 := len(m1[0])
	cols2 := len(m2[0])

	if cols1 != linhas2 {
		return nil, hrp.NewErroF(hrp.SintaxeErro, "dimensoes incompativeis para multiplicacao: (%dx%d) x (%dx%d)", linhas1, cols1, linhas2, cols2)
	}

	res := make([][]float64, linhas1)
	for i := 0; i < linhas1; i++ {
		res[i] = make([]float64, cols2)
		for j := 0; j < cols2; j++ {
			soma := 0.0
			for k := 0; k < cols1; k++ {
				soma += m1[i][k] * m2[k][j]
			}
			res[i][j] = soma
		}
	}

	resLista := make([]hrp.Objeto, linhas1)
	for i := 0; i < linhas1; i++ {
		resLista[i] = floatSliceParaLista(res[i])
	}
	return &hrp.Lista{Itens: hrp.Tupla(resLista)}, nil
}

var _matriz_produtoEscalar = hrp.NewMetodoOuPanic("produtoEscalar", met_matriz_produtoEscalar, "produtoEscalar(v1, v2) -> Retorna o produto escalar entre dois vetores")
var _matriz_somarVetores = hrp.NewMetodoOuPanic("somarVetores", met_matriz_somarVetores, "somarVetores(v1, v2) -> Retorna a soma elemento a elemento de dois vetores")
var _matriz_subtrairVetores = hrp.NewMetodoOuPanic("subtrairVetores", met_matriz_subtrairVetores, "subtrairVetores(v1, v2) -> Retorna a subtracao elemento a elemento de dois vetores")
var _matriz_multiplicarEscalar = hrp.NewMetodoOuPanic("multiplicarEscalar", met_matriz_multiplicarEscalar, "multiplicarEscalar(v, k) -> Multiplica cada elemento de um vetor pelo escalar k")
var _matriz_norma = hrp.NewMetodoOuPanic("norma", met_matriz_norma, "norma(v) -> Retorna a norma L2 (magnitude) do vetor")
var _matriz_criar = hrp.NewMetodoOuPanic("criar", met_matriz_criar, "criar(linhas, colunas, [valorPadrao]) -> Cria matriz inicializada")
var _matriz_multiplicar = hrp.NewMetodoOuPanic("multiplicar", met_matriz_multiplicar, "multiplicar(m1, m2) -> Multiplica duas matrizes 2D")
