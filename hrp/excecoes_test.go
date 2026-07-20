package hrp

import (
	"testing"
)

func TestExcecoesTryCatch(t *testing.T) {
	codigo := `
	var capturou = Falso
	var rodouFinalmente = Falso

	tente {
		assegura 1 == 2, "Erro simulado"
	} capture (erro) {
		capturou = Verdadeiro
		var msg = Texto(erro.mensagem)
		assegura msg == "Erro simulado", "A mensagem do erro deve ser correta"
	} finalmente {
		rodouFinalmente = Verdadeiro
	}
	`

	ctx := NewContexto(OpcsContexto{})
	defer ctx.Terminar()

	_, err := ExecutarString(ctx, codigo)
	if err != nil {
		t.Fatalf("Erro inesperado durante a execução: %v", err)
	}

	modulo, err := ctx.ObterModulo("__entrada__")
	if err != nil {
		t.Fatal(err)
	}

	capturou, _ := modulo.Escopo.ObterValor("capturou")
	if capturou != Verdadeiro {
		t.Errorf("Esperava que 'capturou' fosse Verdadeiro, obteve: %v", capturou)
	}

	rodouFinalmente, _ := modulo.Escopo.ObterValor("rodouFinalmente")
	if rodouFinalmente != Verdadeiro {
		t.Errorf("Esperava que 'rodouFinalmente' fosse Verdadeiro, obteve: %v", rodouFinalmente)
	}
}

func TestExcecoesPropagacaoSemCapture(t *testing.T) {
	codigoValido := `
	var rodouFinalmente = Falso
	tente {
		assegura(1 == 2, "Erro propagado")
	} finalmente {
		rodouFinalmente = Verdadeiro
	}
	`

	ctx := NewContexto(OpcsContexto{})
	defer ctx.Terminar()

	_, err := ExecutarString(ctx, codigoValido)
	if err == nil {
		t.Fatal("Esperava erro não capturado se propagando")
	}

	if objErr, ok := err.(*Erro); ok {
		if objErr.Tipo() != ErroDeAsseguracao {
			t.Errorf("Esperava ErroDeAsseguracao, obteve: %v", objErr.Tipo())
		}
	} else {
		t.Fatalf("Erro inesperado: %v", err)
	}

	modulo, _ := ctx.ObterModulo("__entrada__")
	rodouFinalmente, _ := modulo.Escopo.ObterValor("rodouFinalmente")
	if rodouFinalmente != Verdadeiro {
		t.Errorf("Esperava que o bloco 'finalmente' rodasse mesmo com o erro propagado")
	}
}

// Garante que 'tente' funciona sem o bloco 'finalmente' (sintaxe opcional).
func TestExcecoesSemFinalmente(t *testing.T) {
	codigo := `
	var capturou = Falso
	tente {
		assegura 1 == 2, "falhou"
	} capture (erro) {
		capturou = Verdadeiro
	}
	`

	ctx := NewContexto(OpcsContexto{})
	defer ctx.Terminar()

	if _, err := ExecutarString(ctx, codigo); err != nil {
		t.Fatalf("Erro inesperado: %v", err)
	}

	modulo, _ := ctx.ObterModulo("__entrada__")
	capturou, _ := modulo.Escopo.ObterValor("capturou")
	if capturou != Verdadeiro {
		t.Errorf("Esperava 'capturou' Verdadeiro, obteve: %v", capturou)
	}
}

// 'tente' aninhado: o capture interno intercepta antes do externo.
func TestExcecoesAninhadas(t *testing.T) {
	codigo := `
	var nivel = 0
	tente {
		tente {
			assegura 1 == 2, "interno"
		} capture (erro) {
			nivel = 1
		}
		var valor = 1 + nivel
		assegura valor == 2, "depois do capture aninhado"
	} capture (erro) {
		nivel = 99
	}
	`

	ctx := NewContexto(OpcsContexto{})
	defer ctx.Terminar()

	if _, err := ExecutarString(ctx, codigo); err != nil {
		t.Fatalf("Erro inesperado: %v", err)
	}

	modulo, _ := ctx.ObterModulo("__entrada__")
	nivel, _ := modulo.Escopo.ObterValor("nivel")
	if nivel != Inteiro(1) {
		t.Errorf("Esperava nivel=1, obteve: %v", nivel)
	}
}

// finally sempre roda — mesmo quando o tente passa limpo sem erros.
func TestExcecoesFinallyEmSucesso(t *testing.T) {
	codigo := `
	var rodou = Falso
	tente {
		var total = 1 + 1
	} capture (erro) {
		rodou = Verdadeiro
	} finalmente {
		rodou = Verdadeiro
	}
	`

	ctx := NewContexto(OpcsContexto{})
	defer ctx.Terminar()

	if _, err := ExecutarString(ctx, codigo); err != nil {
		t.Fatalf("Erro inesperado: %v", err)
	}

	modulo, _ := ctx.ObterModulo("__entrada__")
	rodou, _ := modulo.Escopo.ObterValor("rodou")
	if rodou != Verdadeiro {
		t.Errorf("Esperava 'rodou' Verdadeiro do finally, obteve: %v", rodou)
	}
}

// Captura pode inspecionar metadados geográficos do erro (linha/coluna/arquivo).
func TestExcecoesInspecionarMetadados(t *testing.T) {
	codigo := `
	var capturouLinha = 0
	var capturouArquivo = ""
	tente {
		assegura 1 == 2, "falha aqui"
	} capture (erro) {
		capturouLinha = erro.linha
		capturouArquivo = erro.arquivo
	} finalmente {}
	`

	ctx := NewContexto(OpcsContexto{})
	defer ctx.Terminar()

	if _, err := ExecutarString(ctx, codigo); err != nil {
		t.Fatalf("Erro inesperado: %v", err)
	}

	modulo, _ := ctx.ObterModulo("__entrada__")
	linha, _ := modulo.Escopo.ObterValor("capturouLinha")
	arquivo, _ := modulo.Escopo.ObterValor("capturouArquivo")

	if linha == Inteiro(-1) {
		t.Errorf("Esperava linha preenchida, obteve valor default -1")
	}
	if arquivo == Texto("") {
		t.Errorf("Esperava arquivo preenchido, obteve string vazia")
	}
}
