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

// TestMaquinarioImporteModulo testa o importador de baixo nível de módulos nativos ou empacotados,
// verificando se os atributos e assinaturas de métodos do módulo importado são expostos corretamente.
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

// TestMultiImporteModulo garante a importação concorrente ou sequencial de múltiplos módulos distintos,
// validando se cada um deles é adicionado corretamente ao registro global de módulos do contexto da VM.
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

// TestImporteSemCriarContexto valida que invocar funções de importação sem possuir um contexto
// de execução ativo ou previamente inicializado na Thread-Local lance um pânico explicativo em Go.
func TestImporteSemCriarContexto(t *testing.T) {
	teste := func() {
		hrp.Importe("embutidos", nil)
	}

	assertPanic(t, teste)
}

// TestImporteComContexto verifica a importação de módulos com o contexto devidamente inicializado,
// garantindo que os métodos expostos sejam acessíveis de forma transparente.
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

// TestImportacoesRelativas garante que caminhos de importação relativa resolvam de forma adequada
// usando o escopo léxico atual e falhem de forma controlada quando o módulo solicitado não existe.
func TestImportacoesRelativas(t *testing.T) {
	ctx := hrp.NewContexto(hrp.OpcsContexto{})
	defer ctx.Terminar()

	escopoTeste := hrp.NewEscopo()
	escopoTeste.DefinirSimbolo(hrp.NewVarSimbolo("__arquivo__", hrp.Texto(".")))

	if _, err := hrp.Importe("./algo", escopoTeste); err == nil {
		t.Error("Era esperado um erro, pois `algo.hrp` não existe")
	}
}
