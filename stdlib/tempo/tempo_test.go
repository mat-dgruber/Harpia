package tempo

import (
	"testing"
	"time"

	"github.com/mat-dgruber/Harpia/hrp"
)

func TestTempoAgoraEFormatar(t *testing.T) {
	res, err := met_tempo_agora(nil, hrp.Tupla{})
	if err != nil {
		t.Fatalf("erro em agora(): %v", err)
	}

	dh, ok := res.(*DataHora)
	if !ok {
		t.Fatalf("esperava *DataHora")
	}

	// Atributos de data
	ano, _ := dh.M__obtem_attributo__("ano")
	if int(ano.(hrp.Inteiro)) != time.Now().Year() {
		t.Errorf("ano incorreto")
	}

	// Formatar padrão
	metFormatar, _ := dh.M__obtem_attributo__("formatar")
	resFmtPadrao, err := hrp.Chamar(metFormatar, hrp.Tupla{})
	if err != nil {
		t.Fatalf("erro ao formatar padrao: %v", err)
	}
	if len(string(resFmtPadrao.(hrp.Texto))) == 0 {
		t.Errorf("formatação vazia")
	}

	// Formatar customizado YYYY-MM-DD
	resFmtCustom, err := hrp.Chamar(metFormatar, hrp.Tupla{hrp.Texto("YYYY-MM-DD")})
	if err != nil {
		t.Fatalf("erro ao formatar customizado: %v", err)
	}
	fmtStr := string(resFmtCustom.(hrp.Texto))
	if len(fmtStr) != 10 || fmtStr[4] != '-' || fmtStr[7] != '-' {
		t.Errorf("formato YYYY-MM-DD invalido: %s", fmtStr)
	}

	// Adicionar dias
	metAdd, _ := dh.M__obtem_attributo__("adicionarDias")
	resAdd, err := hrp.Chamar(metAdd, hrp.Tupla{hrp.Inteiro(5)})
	if err != nil {
		t.Fatalf("erro em adicionarDias: %v", err)
	}
	dhAdd := resAdd.(*DataHora)
	if dhAdd.t.Sub(dh.t).Hours() < 24*4 {
		t.Errorf("esperava acréscimo de 5 dias")
	}
}
