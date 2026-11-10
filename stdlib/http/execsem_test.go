package http

import (
	"fmt"
	"net"
	"net/http"
	"testing"

	"github.com/mat-dgruber/Harpia/hrp"
)

// Um pânico dentro do handler não pode deixar o semáforo de execução preso.
func TestPanicoNoHandlerLiberaSemaforo(t *testing.T) {
	ctx := hrp.NewContexto(hrp.OpcsContexto{})
	defer ctx.Terminar()

	obj, err := TipoServidor.Nova(nil)
	if err != nil {
		t.Fatal(err)
	}
	s := obj.(*Servidor)
	s.registrarRota("GET", "/panico", hrp.NewMetodoOuPanic("panico", func(_ hrp.Objeto, _ hrp.Tupla) (hrp.Objeto, error) {
		panic("falha simulada")
	}, ""), nil)
	s.registrarRota("GET", "/ok", hrp.NewMetodoOuPanic("ok", func(_ hrp.Objeto, _ hrp.Tupla) (hrp.Objeto, error) {
		return hrp.Nulo, nil
	}, ""), nil)

	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	porta := l.Addr().(*net.TCPAddr).Port
	l.Close()
	escutar, _ := s.M__obtem_attributo__("escutar")
	if _, err := hrp.Chamar(escutar, hrp.Tupla{hrp.Texto(fmt.Sprint(porta))}); err != nil {
		t.Fatal(err)
	}
	defer s.server.Close()

	base := fmt.Sprintf("http://127.0.0.1:%d", porta)
	for i := 0; i < 2; i++ {
		if resp, err := http.Get(base + "/panico"); err == nil {
			resp.Body.Close()
		}
	}
	resp, err := http.Get(base + "/ok")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("depois de um pânico o servidor deveria seguir atendendo, veio %d", resp.StatusCode)
	}
}
