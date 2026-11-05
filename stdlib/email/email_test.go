package email

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
	"testing"

	"github.com/mat-dgruber/Harpia/hrp"
)

func TestEmailEnviarSimulacao(t *testing.T) {
	ctx := hrp.NewContexto(hrp.OpcsContexto{}) // o sandbox nega I/O sem contexto
	defer ctx.Terminar()
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

// smtpFalso aceita uma sessão SMTP mínima e devolve os comandos e a mensagem recebidos.
func smtpFalso(t *testing.T) (porta string, recebido <-chan string) {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { l.Close() })
	ch := make(chan string, 1)
	go func() {
		c, err := l.Accept()
		if err != nil {
			return
		}
		defer c.Close()
		var log strings.Builder
		r := bufio.NewReader(c)
		fmt.Fprint(c, "220 falso\r\n")
		dados := false
		for {
			linha, err := r.ReadString('\n')
			if err != nil {
				break
			}
			log.WriteString(linha)
			if dados {
				if linha == ".\r\n" {
					dados = false
					fmt.Fprint(c, "250 ok\r\n")
				}
				continue
			}
			switch cmd := strings.ToUpper(strings.SplitN(strings.TrimSpace(linha), " ", 2)[0]); cmd {
			case "EHLO", "HELO":
				fmt.Fprint(c, "250 falso\r\n")
			case "DATA":
				dados = true
				fmt.Fprint(c, "354 envie\r\n")
			case "QUIT":
				fmt.Fprint(c, "221 tchau\r\n")
				ch <- log.String()
				return
			default:
				fmt.Fprint(c, "250 ok\r\n")
			}
		}
		ch <- log.String()
	}()
	return strconv.Itoa(l.Addr().(*net.TCPAddr).Port), ch
}

func TestEmailVariosDestinatariosEAssuntoUTF8(t *testing.T) {
	ctx := hrp.NewContexto(hrp.OpcsContexto{})
	defer ctx.Terminar()
	porta, recebido := smtpFalso(t)
	mapa := hrp.Mapa{
		"servidor":  hrp.Texto("127.0.0.1"),
		"porta":     hrp.Texto(porta),
		"de":        hrp.Texto("app@exemplo.com"),
		"para":      hrp.Texto("a@exemplo.com, b@exemplo.com"),
		"assunto":   hrp.Texto("Ação concluída"),
		"corpoHtml": hrp.Texto("<p>oi</p>"),
	}
	if _, err := met_email_enviar(nil, hrp.Tupla{mapa}); err != nil {
		t.Fatal(err)
	}
	sessao := <-recebido
	for _, esperado := range []string{"MAIL FROM:<app@exemplo.com>", "RCPT TO:<a@exemplo.com>", "RCPT TO:<b@exemplo.com>", "Subject: =?utf-8?q?"} {
		if !strings.Contains(sessao, esperado) {
			t.Errorf("sessão SMTP sem %q:\n%s", esperado, sessao)
		}
	}
	if strings.Contains(sessao, "<nil>") {
		t.Errorf("campo ausente virou '<nil>':\n%s", sessao)
	}
}
