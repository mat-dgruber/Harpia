package tensor

import (
	"fmt"
	"slices"

	"github.com/mat-dgruber/Harpia/hrp"
)

type TensorObj struct {
	hrp.GCMixin
	Formato    []int
	Dados      []float64
	Gradiente  []float64
	RequerGrad bool
}

var TipoTensor = hrp.TipoObjeto.NewTipo(
	"Tensor",
	"Tensor(formato, dados) -> Instância de Tensor N-dimensional",
)

func (t *TensorObj) Tipo() *hrp.Tipo {
	return TipoTensor
}

func (t *TensorObj) ObterFilhos() []hrp.Objeto {
	return nil
}

func (t *TensorObj) M__texto__() (hrp.Objeto, error) {
	return hrp.Texto(fmt.Sprintf("Tensor(formato=%v, elementos=%d)", t.Formato, len(t.Dados))), nil
}

func init() {
	metodos := []*hrp.Metodo{
		_tensor_criar,
		_tensor_redimensionar,
		_tensor_somar,
		_tensor_multiplicar,
	}

	hrp.RegistraModuloImpl(
		&hrp.ModuloImpl{
			Info: hrp.ModuloInfo{
				Nome:    "tensor",
				Arquivo: "stdlib/tensor",
			},
			Constantes: hrp.Mapa{},
			Metodos:    metodos,
		},
	)

	TipoTensor.Mapa["texto"] = hrp.NewMetodoOuPanic("texto", func(inst hrp.Objeto) (hrp.Objeto, error) {
		t, ok := inst.(*TensorObj)
		if !ok {
			return nil, hrp.NewErroF(hrp.TipagemErro, "esperava instância de TensorObj")
		}
		return t.M__texto__()
	}, "representacao em texto do tensor")
}

func extrairFormato(obj hrp.Objeto) ([]int, error) {
	lista, ok := obj.(*hrp.Lista)
	if !ok {
		return nil, hrp.NewErroF(hrp.TipagemErro, "esperava lista de dimensoes para formato do Tensor")
	}
	if len(lista.Itens) == 0 {
		return nil, hrp.NewErroF(hrp.ValorErro, "formato do tensor não pode ser vazio")
	}
	formato := make([]int, len(lista.Itens))
	total := 1
	const maxElementos = 1 << 24
	for i, elem := range lista.Itens {
		dec, err := hrp.NewDecimal(elem)
		if err != nil {
			return nil, err
		}
		d := int(dec.(hrp.Decimal))
		if d <= 0 {
			return nil, hrp.NewErroF(hrp.ValorErro, "dimensão do tensor deve ser positiva, recebeu %d", d)
		}
		if d > maxElementos/total {
			return nil, hrp.NewErroF(hrp.ValorErro, "tamanho total do tensor excede o limite máximo permitido (%d)", maxElementos)
		}
		total *= d
		formato[i] = d
	}
	return formato, nil
}

func met_tensor_criar(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if err := hrp.VerificaNumeroArgumentos("criar", false, args, 1, 2); err != nil {
		return nil, err
	}
	formato, err := extrairFormato(args[0])
	if err != nil {
		return nil, err
	}
	total := 1
	for _, d := range formato {
		total *= d
	}

	dados := make([]float64, total)
	if len(args) > 1 {
		valObj, err := hrp.NewDecimal(args[1])
		if err != nil {
			return nil, err
		}
		v := float64(valObj.(hrp.Decimal))
		for i := 0; i < total; i++ {
			dados[i] = v
		}
	}

	formatoClonado := make([]int, len(formato))
	copy(formatoClonado, formato)

	return &TensorObj{
		Formato: formatoClonado,
		Dados:   dados,
	}, nil
}

func met_tensor_redimensionar(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if err := hrp.VerificaNumeroArgumentos("redimensionar", false, args, 2, 2); err != nil {
		return nil, err
	}
	t, ok := args[0].(*TensorObj)
	if !ok {
		return nil, hrp.NewErroF(hrp.TipagemErro, "primeiro argumento deve ser um Tensor")
	}
	novoFormato, err := extrairFormato(args[1])
	if err != nil {
		return nil, err
	}

	novoTotal := 1
	for _, d := range novoFormato {
		novoTotal *= d
	}
	if novoTotal != len(t.Dados) {
		return nil, hrp.NewErroF(hrp.SintaxeErro, "o novo formato (%v -> total %d) e incompativel com a quantidade de dados (%d)", novoFormato, novoTotal, len(t.Dados))
	}

	novosDados := make([]float64, len(t.Dados))
	copy(novosDados, t.Dados)
	formatoClonado := make([]int, len(novoFormato))
	copy(formatoClonado, novoFormato)

	return &TensorObj{
		Formato: formatoClonado,
		Dados:   novosDados,
	}, nil
}

func met_tensor_somar(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if err := hrp.VerificaNumeroArgumentos("somar", false, args, 2, 2); err != nil {
		return nil, err
	}
	t1, ok1 := args[0].(*TensorObj)
	t2, ok2 := args[1].(*TensorObj)
	if !ok1 || !ok2 {
		return nil, hrp.NewErroF(hrp.TipagemErro, "ambos os argumentos devem ser do tipo Tensor")
	}
	if !slices.Equal(t1.Formato, t2.Formato) {
		return nil, hrp.NewErroF(hrp.SintaxeErro, "tensores devem ter o mesmo formato para a soma (%v != %v)", t1.Formato, t2.Formato)
	}

	resDados := make([]float64, len(t1.Dados))
	for i := 0; i < len(t1.Dados); i++ {
		resDados[i] = t1.Dados[i] + t2.Dados[i]
	}
	formatoRes := make([]int, len(t1.Formato))
	copy(formatoRes, t1.Formato)

	return &TensorObj{
		Formato: formatoRes,
		Dados:   resDados,
	}, nil
}

func met_tensor_multiplicar(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if err := hrp.VerificaNumeroArgumentos("multiplicar", false, args, 2, 2); err != nil {
		return nil, err
	}
	t1, ok1 := args[0].(*TensorObj)
	t2, ok2 := args[1].(*TensorObj)
	if !ok1 || !ok2 {
		return nil, hrp.NewErroF(hrp.TipagemErro, "ambos os argumentos devem ser do tipo Tensor")
	}
	if !slices.Equal(t1.Formato, t2.Formato) {
		return nil, hrp.NewErroF(hrp.SintaxeErro, "tensores devem ter o mesmo formato para a multiplicacao ponto a ponto (%v != %v)", t1.Formato, t2.Formato)
	}

	resDados := make([]float64, len(t1.Dados))
	for i := 0; i < len(t1.Dados); i++ {
		resDados[i] = t1.Dados[i] * t2.Dados[i]
	}
	formatoRes := make([]int, len(t1.Formato))
	copy(formatoRes, t1.Formato)

	return &TensorObj{
		Formato: formatoRes,
		Dados:   resDados,
	}, nil
}

var _tensor_criar = hrp.NewMetodoOuPanic("criar", met_tensor_criar, "criar(formato, [valorPadrao]) -> Cria um novo Tensor N-Dimensional")
var _tensor_redimensionar = hrp.NewMetodoOuPanic("redimensionar", met_tensor_redimensionar, "redimensionar(tensor, novoFormato) -> Altera a forma do tensor mantendo os dados")
var _tensor_somar = hrp.NewMetodoOuPanic("somar", met_tensor_somar, "somar(t1, t2) -> Soma dois tensores ponto a ponto")
var _tensor_multiplicar = hrp.NewMetodoOuPanic("multiplicar", met_tensor_multiplicar, "multiplicar(t1, t2) -> Multiplica dois tensores ponto a ponto")
