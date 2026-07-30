package email

import (
	"testing"

	"github.com/mat-dgruber/Harpia/hrp"
)

func TestEmailEnviarSimulacao(t *testing.T) {
	mapa := hrp.NewMapaVazio()
	mapa.M__define_item__(hrp.Texto("para"), hrp.Texto("teste@exemplo.com"))
	mapa.M__define_item__(hrp.Texto("assunto"), hrp.Texto("Teste"))
	mapa.M__define_item__(hrp.Texto("corpoHtml"), hrp.Texto("<p>Ola</p>"))

	res, err := met_email_enviar(nil, hrp.Tupla{mapa})
	if err != nil {
		t.Fatalf("Erro no envio simulado: %v", err)
	}

	if boolRes, ok := res.(hrp.Booleano); !ok || !bool(boolRes) {
		t.Errorf("Esperava retorno verdadeiro no envio simulado")
	}
}
