package http

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/mat-dgruber/Harpia/hrp"
)

func nativo(fn func(req *Requisicao, res *Resposta) error) hrp.Objeto {
	return hrp.NewMetodoOuPanic("h", func(_ hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
		return hrp.Nulo, fn(args[0].(*Requisicao), args[1].(*Resposta))
	}, "")
}

// servidorDeTeste monta o mux real (escutar não bloqueante em porta efêmera) e devolve o handler.
func servidorDeTeste(t *testing.T, configurar func(s *Servidor)) (*Servidor, http.Handler) {
	t.Helper()
	ctx := hrp.NewContexto(hrp.OpcsContexto{})
	t.Cleanup(ctx.Terminar)
	obj, err := TipoServidor.Nova(nil)
	if err != nil {
		t.Fatal(err)
	}
	s := obj.(*Servidor)
	configurar(s)
	escutar, _ := s.M__obtem_attributo__("escutar")
	if _, err := hrp.Chamar(escutar, hrp.Tupla{hrp.Texto("0")}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.server.Close() })
	return s, s.server.Handler
}

func fazer(h http.Handler, metodo, caminho string, cab map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(metodo, caminho, nil)
	for k, v := range cab {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

// H10: erro interno e pânico não vazam detalhes ao cliente.
func TestErroInternoNaoVazaDetalhes(t *testing.T) {
	_, h := servidorDeTeste(t, func(s *Servidor) {
		s.registrarRota("GET", "/erro", nativo(func(*Requisicao, *Resposta) error {
			return errors.New("SELECT * FROM segredo WHERE senha='x'")
		}), nil)
		s.registrarRota("GET", "/panico", nativo(func(*Requisicao, *Resposta) error { panic("/srv/app/segredo.go") }), nil)
	})
	for _, rota := range []string{"/erro", "/panico"} {
		rec := fazer(h, "GET", rota, nil)
		corpo := rec.Body.String()
		if rec.Code != http.StatusInternalServerError {
			t.Errorf("%s: status %d", rota, rec.Code)
		}
		if strings.Contains(corpo, "segredo") {
			t.Errorf("%s vazou detalhe interno: %q", rota, corpo)
		}
		if !strings.Contains(rec.Header().Get("Content-Type"), "application/problem+json") || !strings.Contains(corpo, "trace_id") {
			t.Errorf("%s: esperava Problem Details com trace_id, veio %q (%s)", rota, corpo, rec.Header().Get("Content-Type"))
		}
	}
}

// M1: um 500 não tira o servidor de rotação para sempre.
func TestReadyzContinuaProntoDepoisDeErro(t *testing.T) {
	_, h := servidorDeTeste(t, func(s *Servidor) {
		s.registrarRota("GET", "/erro", nativo(func(*Requisicao, *Resposta) error { return errors.New("x") }), nil)
	})
	fazer(h, "GET", "/erro", nil)
	if rec := fazer(h, "GET", "/readyz", nil); rec.Code != http.StatusOK {
		t.Fatalf("/readyz depois de um 500: %d %s", rec.Code, rec.Body.String())
	}
}

// M2: cabeçalho com valor não-Texto não derruba a requisição.
func TestCabecalhoNaoTexto(t *testing.T) {
	_, h := servidorDeTeste(t, func(s *Servidor) {
		s.registrarRota("GET", "/n", nativo(func(_ *Requisicao, res *Resposta) error {
			res.Cabecalho["X-Total"] = hrp.Inteiro(5)
			return nil
		}), nil)
	})
	rec := fazer(h, "GET", "/n", nil)
	if rec.Code != http.StatusOK || rec.Header().Get("X-Total") != "5" {
		t.Fatalf("status %d, X-Total=%q", rec.Code, rec.Header().Get("X-Total"))
	}
}

// M3: rotas com a mesma força são desempatadas pela ordem de registro (determinístico).
func TestRoteamentoDeterministico(t *testing.T) {
	_, h := servidorDeTeste(t, func(s *Servidor) {
		s.registrarRota("GET", "/u/:id", nativo(func(_ *Requisicao, res *Resposta) error { res.Corpo = "id"; return nil }), nil)
		s.registrarRota("GET", "/u/:nome", nativo(func(_ *Requisicao, res *Resposta) error { res.Corpo = "nome"; return nil }), nil)
	})
	for i := 0; i < 100; i++ {
		if corpo := fazer(h, "GET", "/u/7", nil).Body.String(); corpo != "id" {
			t.Fatalf("iteração %d escolheu %q", i, corpo)
		}
	}
}

// M4: middleware que lança erro com a frase mágica não é executado duas vezes.
func TestMiddlewareNaoReexecutaPorTextoDeErro(t *testing.T) {
	var chamadas atomic.Int32
	_, h := servidorDeTeste(t, func(s *Servidor) {
		s.middlewares = append(s.middlewares, hrp.NewMetodoOuPanic("mw", func(_ hrp.Objeto, _ hrp.Tupla) (hrp.Objeto, error) {
			chamadas.Add(1)
			return nil, errors.New("validação: esperava no máximo 2 argumentos na regra X")
		}, ""))
		s.registrarRota("GET", "/x", nativo(func(*Requisicao, *Resposta) error { return nil }), nil)
	})
	fazer(h, "GET", "/x", nil)
	if c := chamadas.Load(); c != 1 {
		t.Fatalf("middleware executado %d vezes", c)
	}
}

// M5: CORS restringível por app.cors([...]) e /docs desligável com HARPIA_DOCS=0.
func TestCorsRestritoEDocsDesligavel(t *testing.T) {
	t.Setenv("HARPIA_DOCS", "0")
	_, h := servidorDeTeste(t, func(s *Servidor) {
		cors, _ := s.M__obtem_attributo__("cors")
		if _, err := hrp.Chamar(cors, hrp.Tupla{&hrp.Lista{Itens: []hrp.Objeto{hrp.Texto("https://app.exemplo")}}}); err != nil {
			t.Fatal(err)
		}
		s.registrarRota("GET", "/x", nativo(func(*Requisicao, *Resposta) error { return nil }), nil)
	})
	if got := fazer(h, "GET", "/x", map[string]string{"Origin": "https://evil.example"}).Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("origem fora da lista recebeu CORS %q", got)
	}
	if got := fazer(h, "GET", "/x", map[string]string{"Origin": "https://app.exemplo"}).Header().Get("Access-Control-Allow-Origin"); got != "https://app.exemplo" {
		t.Errorf("origem permitida recebeu %q", got)
	}
	if rec := fazer(h, "GET", "/docs.json", nil); rec.Code != http.StatusNotFound {
		t.Errorf("/docs.json com HARPIA_DOCS=0 deveria ser 404, veio %d", rec.Code)
	}
}

// M5 (compatível): sem app.cors, o padrão continua '*'.
func TestCorsPadraoContinuaAberto(t *testing.T) {
	_, h := servidorDeTeste(t, func(s *Servidor) {
		s.registrarRota("GET", "/x", nativo(func(*Requisicao, *Resposta) error { return nil }), nil)
	})
	if got := fazer(h, "GET", "/x", map[string]string{"Origin": "https://qualquer.exemplo"}).Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Errorf("padrão deveria continuar '*', veio %q", got)
	}
}

// M6: requisitar não lê respostas gigantes sem limite.
func TestRequisitarLimitaTamanhoDaResposta(t *testing.T) {
	ctx := hrp.NewContexto(hrp.OpcsContexto{})
	defer ctx.Terminar()
	grande := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		bloco := make([]byte, 1<<20)
		for i := 0; i < 20; i++ {
			w.Write(bloco)
		}
	}))
	defer grande.Close()
	if _, err := met_http_requisitar(nil, hrp.Tupla{hrp.Texto("GET"), hrp.Texto(grande.URL)}); err == nil || !strings.Contains(err.Error(), "limite") {
		t.Fatalf("esperava erro de limite para resposta de 20 MB, veio %v", err)
	}
}
