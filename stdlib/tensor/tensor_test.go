package tensor

import (
	"testing"

	"github.com/mat-dgruber/Harpia/hrp"
)

func TestTensorCriarDimensoesInvalidas(t *testing.T) {
	// criar([0]) deve retornar erro
	f0 := &hrp.Lista{Itens: hrp.Tupla{hrp.Decimal(0)}}
	_, err := met_tensor_criar(nil, hrp.Tupla{f0})
	if err == nil {
		t.Fatal("esperava erro ao criar tensor com dimensão 0")
	}

	// criar([-1]) deve retornar erro
	fNeg := &hrp.Lista{Itens: hrp.Tupla{hrp.Decimal(-1)}}
	_, err = met_tensor_criar(nil, hrp.Tupla{fNeg})
	if err == nil {
		t.Fatal("esperava erro ao criar tensor com dimensão negativa")
	}

	// criar([]) deve retornar erro
	fVazio := &hrp.Lista{Itens: hrp.Tupla{}}
	_, err = met_tensor_criar(nil, hrp.Tupla{fVazio})
	if err == nil {
		t.Fatal("esperava erro ao criar tensor com formato vazio")
	}
}

func TestTensorSomarFormatoIncompativel(t *testing.T) {
	// [2, 3] vs [3, 2] -> mesmo total de elementos (6), mas formatos diferentes
	f1 := &hrp.Lista{Itens: hrp.Tupla{hrp.Decimal(2), hrp.Decimal(3)}}
	f2 := &hrp.Lista{Itens: hrp.Tupla{hrp.Decimal(3), hrp.Decimal(2)}}

	t1, err := met_tensor_criar(nil, hrp.Tupla{f1, hrp.Decimal(1)})
	if err != nil {
		t.Fatalf("erro ao criar t1: %v", err)
	}
	t2, err := met_tensor_criar(nil, hrp.Tupla{f2, hrp.Decimal(1)})
	if err != nil {
		t.Fatalf("erro ao criar t2: %v", err)
	}

	_, err = met_tensor_somar(nil, hrp.Tupla{t1, t2})
	if err == nil {
		t.Fatal("esperava erro ao somar tensores de formatos [2,3] e [3,2]")
	}
}

func TestTensorRedimensionarClonaDadosEFormato(t *testing.T) {
	f1 := &hrp.Lista{Itens: hrp.Tupla{hrp.Decimal(2), hrp.Decimal(2)}}
	orig, err := met_tensor_criar(nil, hrp.Tupla{f1, hrp.Decimal(5)})
	if err != nil {
		t.Fatalf("erro ao criar: %v", err)
	}

	f2 := &hrp.Lista{Itens: hrp.Tupla{hrp.Decimal(4)}}
	redimObj, err := met_tensor_redimensionar(nil, hrp.Tupla{orig, f2})
	if err != nil {
		t.Fatalf("erro ao redimensionar: %v", err)
	}

	redim := redimObj.(*TensorObj)
	origTensor := orig.(*TensorObj)

	// Alterar dados do redimensionado não deve alterar o original
	redim.Dados[0] = 999
	if origTensor.Dados[0] == 999 {
		t.Fatal("redimensionar não clonou os dados, mutação vazou para o original")
	}

	// Alterar formato do redimensionado não deve alterar f2
	redim.Formato[0] = 123
	if origTensor.Formato[0] == 123 {
		t.Fatal("redimensionar compartilhou fatia de formato")
	}
}
