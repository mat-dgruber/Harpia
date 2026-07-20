package hrp_test

import (
	"testing"

	"github.com/mat-dgruber/Harpia/hrp"
	_ "github.com/mat-dgruber/Harpia/stdlib"
)

func assertPanic(t *testing.T, f func()) {
	t.Helper()
	defer func() { _ = recover() }()
	f()
	t.Errorf("Era esperado que houvesse um `panic`")
}

func TestMaquinarioImporteModulo(t *testing.T) {
	var mod, obj hrp.Objeto
	var err error

	ctx := hrp.NewContexto(hrp.OpcsContexto{})
	defer ctx.Terminar()

	if mod, err = hrp.MaquinarioImporteModulo(ctx, "colorize", nil); err != nil {
		t.Error(err)
	}

	if obj, err = hrp.ObtemAtributoS(mod, "converteRGB"); err != nil {
		t.Error(err)
	}

	if obj.(*hrp.Metodo).Nome != "converteRGB" {
		t.Error("erro no nome do método")
	}
}

func TestMultiImporteModulo(t *testing.T) {
	ctx := hrp.NewContexto(hrp.OpcsContexto{})
	defer ctx.Terminar()

	if err := hrp.MultiImporteModulo(ctx, "colorize", "embutidos"); err != nil {
		t.Error(err)
	}

	if _, err := ctx.Modulos.ObterModulo("embutidos"); err != nil {
		t.Error(err)
	}
}

func TestImporteSemCriarContexto(t *testing.T) {
	teste := func() {
		hrp.Importe("embutidos", nil)
	}

	assertPanic(t, teste)
}

func TestImporteComContexto(t *testing.T) {
	var mod, obj hrp.Objeto
	var err error

	ctx := hrp.NewContexto(hrp.OpcsContexto{})
	defer ctx.Terminar()

	if mod, err = hrp.Importe("colorize", nil); err != nil {
		t.Error(err)
	}

	if obj, err = hrp.ObtemAtributoS(mod, "converteRGB"); err != nil {
		t.Error(err)
	}

	if obj.(*hrp.Metodo).Nome != "converteRGB" {
		t.Error("erro no nome do método")
	}
}

func TestImportacoesRelativas(t *testing.T) {
	ctx := hrp.NewContexto(hrp.OpcsContexto{})
	defer ctx.Terminar()

	escopoTeste := hrp.NewEscopo()
	escopoTeste.DefinirSimbolo(hrp.NewVarSimbolo("__arquivo__", hrp.Texto(".")))

	if _, err := hrp.Importe("./algo", escopoTeste); err == nil {
		t.Error("Era esperado um erro, pois `algo.hrp` não existe")
	}
}
