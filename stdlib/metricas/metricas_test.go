package metricas

import (
	"strings"
	"testing"

	"github.com/natanfeitosa/portuscript/ptst"
)

func TestContadorMetrica(t *testing.T) {
	args := ptst.Tupla{ptst.Texto("requisicoes_total"), ptst.Texto("Total de requisições HTTP recebidas")}
	obj, err := met_metricas_criarContador(nil, args)
	if err != nil {
		t.Fatalf("Erro ao criar contador: %v", err)
	}

	contador := obj.(*Contador)
	incMetodo, errInc := contador.M__obtem_attributo__("incrementar")
	if errInc != nil {
		t.Fatalf("Erro ao obter método de incremento: %v", errInc)
	}

	_, errCall := ptst.Chamar(incMetodo, ptst.Tupla{})
	if errCall != nil {
		t.Fatalf("Erro ao chamar incrementar: %v", errCall)
	}

	if contador.info.Valor != 1 {
		t.Errorf("Valor do contador incorreto. Esperava 1, obtive: %g", contador.info.Valor)
	}

	// Testa expor
	exporObj, errExpor := met_metricas_expor(nil, ptst.Tupla{})
	if errExpor != nil {
		t.Fatalf("Erro ao expor métricas: %v", errExpor)
	}

	exporStr := string(exporObj.(ptst.Texto))
	if !strings.Contains(exporStr, "requisicoes_total 1") {
		t.Errorf("Saída do expor incorreta, obtive:\n%s", exporStr)
	}
}

func TestMedidorMetrica(t *testing.T) {
	args := ptst.Tupla{ptst.Texto("uso_memoria"), ptst.Texto("Memória consumida")}
	obj, err := met_metricas_criarMedidor(nil, args)
	if err != nil {
		t.Fatalf("Erro ao criar medidor: %v", err)
	}

	medidor := obj.(*Medidor)
	defMetodo, errDef := medidor.M__obtem_attributo__("definir")
	if errDef != nil {
		t.Fatalf("Erro ao obter método definir: %v", errDef)
	}

	_, errCall := ptst.Chamar(defMetodo, ptst.Tupla{ptst.Decimal(512.5)})
	if errCall != nil {
		t.Fatalf("Erro ao chamar definir: %v", errCall)
	}

	if medidor.info.Valor != 512.5 {
		t.Errorf("Valor do medidor incorreto. Esperava 512.5, obtive: %g", medidor.info.Valor)
	}
}
