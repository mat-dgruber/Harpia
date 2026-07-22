package email

import (
	"fmt"
	"net/smtp"

	"github.com/mat-dgruber/Harpia/hrp"
)

func met_email_enviar(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if err := hrp.VerificaNumeroArgumentos("enviar", false, args, 1, 1); err != nil {
		return nil, err
	}

	mapa, ok := args[0].(hrp.Mapa)
	if !ok {
		return nil, hrp.NewErroF(hrp.TipoErro, "O argumento de 'email.enviar' deve ser um Mapa com as configurações do e-mail")
	}

	servidor := fmt.Sprintf("%v", mapa["servidor"])
	porta := fmt.Sprintf("%v", mapa["porta"])
	usuario := fmt.Sprintf("%v", mapa["usuario"])
	senha := fmt.Sprintf("%v", mapa["senha"])
	para := fmt.Sprintf("%v", mapa["para"])
	assunto := fmt.Sprintf("%v", mapa["assunto"])
	corpoHtml := fmt.Sprintf("%v", mapa["corpoHtml"])

	if servidor == "<nil>" || porta == "<nil>" {
		// Simulação de envio local para desenvolvimento/teste
		fmt.Printf("📧 [SIMULAÇÃO EMAIL] Enviado para '%s' | Assunto: '%s'\n", para, assunto)
		return hrp.Booleano(true), nil
	}

	auth := smtp.PlainAuth("", usuario, senha, servidor)
	msg := []byte(fmt.Sprintf("To: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s", para, assunto, corpoHtml))

	addr := fmt.Sprintf("%s:%s", servidor, porta)
	err := smtp.SendMail(addr, auth, usuario, []string{para}, msg)
	if err != nil {
		return hrp.Booleano(false), hrp.NewErroF(hrp.ValorErro, "Erro ao enviar e-mail via SMTP: %v", err)
	}

	return hrp.Booleano(true), nil
}

var _enviar = hrp.NewMetodoOuPanic("enviar", met_email_enviar, "")

func init() {
	hrp.RegistraModuloImpl(&hrp.ModuloImpl{
		Info: hrp.ModuloInfo{
			Nome:    "email",
			Arquivo: "stdlib/email",
		},
		Metodos: []*hrp.Metodo{
			_enviar,
		},
	})
}
