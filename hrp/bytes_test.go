package hrp_test

import (
	"reflect"
	"testing"

	"github.com/mat-dgruber/Harpia/hrp"
)

func TestStringParaBytes(t *testing.T) {
	frase := "tipo bytes"
	bytes, err := hrp.NewBytes(frase)
	if err != nil {
		t.Errorf("Não foi possível fazer a conversão, o seguinte erro foi retornado: %s", err)
	}

	if bytes == nil {
		t.Error("Era esperado uma instancia do tipo Bytes, mas foi retornado um 'nil'")
	}

	if !reflect.DeepEqual(bytes, &hrp.Bytes{[]byte(frase), false}) {
		t.Error("Aparentemente o construtor não está lidando direito com a conversão")
	}
}

func TestConversaoParaBytesPorMetodoImplementado(t *testing.T) {
	hrp.TipoTexto.Mapa["__bytes__"] = hrp.NewMetodoOuPanic(
		"__bytes__",
		func(inst hrp.Objeto) (hrp.Objeto, error) {
			return hrp.NewBytes(string(inst.(hrp.Texto)))
		},
		"",
	)

	texto := hrp.Texto("tipo bytes")
	bytes, err := hrp.NewBytes(texto)
	if err != nil {
		t.Errorf("Não foi possível fazer a conversão, o seguinte erro foi retornado: %s", err)
	}

	if bytes == nil {
		t.Error("Era esperado uma instancia do tipo Bytes, mas foi retornado um 'nil'")
	}

	if !reflect.DeepEqual(bytes, &hrp.Bytes{[]byte(texto), false}) {
		t.Error("Aparentemente o construtor não está lidando direito com a conversão")
	}
}

func TestCriacaoDeInstanciaVazia(t *testing.T) {
	if _, err := hrp.NewBytes(nil); err != nil {
		t.Error(err)
	}
}

func TestConversaoPelaChamadaDoConstrutor(t *testing.T) {
	if _, err := hrp.NovaInstancia(hrp.TipoBytes, hrp.Tupla{&hrp.Bytes{}}); err != nil {
		t.Error(err)
	}
}

func TestComparacaoRica(t *testing.T) {
	var a, b = &hrp.Bytes{}, &hrp.Bytes{}

	t.Run("`a == b`", func(t *testing.T) {
		res, err := hrp.Igual(a, b)
		if err != nil {
			t.Error(err)
		}

		if !res.(hrp.Booleano) {
			t.Error("deveria ser Verdadeiro, mas deu Falso")
		}
	})

	t.Run("`a != b`", func(t *testing.T) {
		res, err := hrp.Diferente(a, b)
		if err != nil {
			t.Error(err)
		}

		if res.(hrp.Booleano) {
			t.Error("deveria ser Falso, mas deu Verdadeiro")
		}
	})

	t.Run("`a >= b`", func(t *testing.T) {
		res, err := hrp.MaiorOuIgual(a, b)
		if err != nil {
			t.Error(err)
		}

		if !res.(hrp.Booleano) {
			t.Error("deveria ser Verdadeiro, mas deu Falso")
		}
	})

	t.Run("`a > b`", func(t *testing.T) {
		res, err := hrp.MaiorQue(a, b)
		if err != nil {
			t.Error(err)
		}

		if res.(hrp.Booleano) {
			t.Error("deveria ser Falso, mas deu Verdadeiro")
		}
	})

	t.Run("`a <= b`", func(t *testing.T) {
		res, err := hrp.MenorOuIgual(a, b)
		if err != nil {
			t.Error(err)
		}

		if !res.(hrp.Booleano) {
			t.Error("deveria ser Verdadeiro, mas deu Falso")
		}
	})

	t.Run("`a < b`", func(t *testing.T) {
		res, err := hrp.MenorQue(a, b)
		if err != nil {
			t.Error(err)
		}

		if res.(hrp.Booleano) {
			t.Error("deveria ser Falso, mas deu Verdadeiro")
		}
	})
}

func TestConversaoDeTipos(t *testing.T) {
	a := &hrp.Bytes{[]byte("a"), false}

	t.Run("=> booleano", func(t *testing.T) {
		res, err := hrp.NewBooleano(a)
		if err != nil {
			t.Error(err)
		}

		if !res.(hrp.Booleano) {
			t.Error("deveria ser Verdadeiro, mas deu Falso")
		}
	})

	t.Run("=> texto", func(t *testing.T) {
		text, err := hrp.NewTexto(a)
		if err != nil {
			t.Error(err)
		}

		if res, _ := hrp.Igual(hrp.Texto("a"), text); !res.(hrp.Booleano) {
			t.Error("a comparação com o tipo convertido não deu verdadeiro")
		}
	})
}
