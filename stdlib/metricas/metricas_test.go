package metricas

import (
	"strings"
	"testing"

	"github.com/mat-dgruber/Harpia/hrp"
)

func TestContadorMetrica(t *testing.T) {
	args := hrp.Tupla{hrp.Texto("requisicoes_total"), hrp.Texto("Total de requisições HTTP recebidas")}
	obj, err := met_metricas_criarContador(nil, args)
	if err != nil {
		t.Fatalf("Erro ao criar contador: %v", err)
	}

	contador := obj.(*Contador)
	incMetodo, errInc := contador.M__obtem_attributo__("incrementar")
	if errInc != nil {
		t.Fatalf("Erro ao obter método de incremento: %v", errInc)
	}

	_, errCall := hrp.Chamar(incMetodo, hrp.Tupla{})
	if errCall != nil {
		t.Fatalf("Erro ao chamar incrementar: %v", errCall)
	}

	if contador.info.Valor != 1 {
		t.Errorf("Valor do contador incorreto. Esperava 1, obtive: %g", contador.info.Valor)
	}

	// Testa expor
	exporObj, errExpor := met_metricas_expor(nil, hrp.Tupla{})
	if errExpor != nil {
		t.Fatalf("Erro ao expor métricas: %v", errExpor)
	}

	exporStr := string(exporObj.(hrp.Texto))
	if !strings.Contains(exporStr, "requisicoes_total 1") {
		t.Errorf("Saída do expor incorreta, obtive:\n%s", exporStr)
	}
}

func TestMedidorMetrica(t *testing.T) {
	args := hrp.Tupla{hrp.Texto("uso_memoria"), hrp.Texto("Memória consumida")}
	obj, err := met_metricas_criarMedidor(nil, args)
	if err != nil {
		t.Fatalf("Erro ao criar medidor: %v", err)
	}

	medidor := obj.(*Medidor)
	defMetodo, errDef := medidor.M__obtem_attributo__("definir")
	if errDef != nil {
		t.Fatalf("Erro ao obter método definir: %v", errDef)
	}

	_, errCall := hrp.Chamar(defMetodo, hrp.Tupla{hrp.Decimal(512.5)})
	if errCall != nil {
		t.Fatalf("Erro ao chamar definir: %v", errCall)
	}

	if medidor.info.Valor != 512.5 {
		t.Errorf("Valor do medidor incorreto. Esperava 512.5, obtive: %g", medidor.info.Valor)
	}
}
