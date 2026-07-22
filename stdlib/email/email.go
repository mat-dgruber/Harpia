// Package email fornece facilidades para envio de e-mails transacionais (texto e HTML)
// utilizando autenticação simples via protocolo SMTP de forma síncrona.
package email

import (
	"fmt"
	"net/smtp"

	"github.com/mat-dgruber/Harpia/hrp"
)

// met_email_enviar implementa 'enviar(configMapa)' em nível de script Harpia.
// Recebe um Mapa do Harpia contendo os parâmetros SMTP e o payload da mensagem (para, assunto, corpoHtml).
// Se o servidor SMTP ou a porta não forem fornecidos, ele automaticamente ativa o modo de simulação no console.
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
		// Simulação de envio local para desenvolvimento/teste para evitar bloqueios ou exceptions
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

var _enviar = hrp.NewMetodoOuPanic("enviar", met_email_enviar, "Envia um e-mail com conteúdo em formato HTML através do protocolo SMTP.")

func init() {
	// Registra o módulo 'email' globalmente no ecossistema de módulos do Harpia.
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
