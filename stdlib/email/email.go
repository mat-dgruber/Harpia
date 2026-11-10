// Package email fornece facilidades para envio de e-mails transacionais (texto e HTML)
// utilizando autenticação simples via protocolo SMTP de forma síncrona.
package email

import (
	"fmt"
	"mime"
	"net"
	"net/smtp"
	"strings"

	"github.com/mat-dgruber/Harpia/hrp"
)

// met_email_enviar implementa 'enviar(configMapa)' em nível de script Harpia.
// Recebe um Mapa do Harpia contendo os parâmetros SMTP e o payload da mensagem (para, assunto, corpoHtml).
// Se o servidor SMTP ou a porta não forem fornecidos, ele automaticamente ativa o modo de simulação no console.
func met_email_enviar(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if err := hrp.ExigirPermissao(inst, hrp.RecursoRede); err != nil {
		return nil, err
	}
	if err := hrp.VerificaNumeroArgumentos("enviar", false, args, 1, 1); err != nil {
		return nil, err
	}

	mapa, ok := args[0].(hrp.Mapa)
	if !ok {
		return nil, hrp.NewErroF(hrp.TipoErro, "O argumento de 'email.enviar' deve ser um Mapa com as configurações do e-mail")
	}

	// campo ausente ou Nulo vira "" (antes virava a string "<nil>", inclusive no MAIL FROM)
	campo := func(chave string) string {
		v, ok := mapa[chave]
		if !ok || v == nil || v == hrp.Nulo {
			return ""
		}
		return strings.TrimSpace(fmt.Sprintf("%v", v))
	}
	servidor, porta := campo("servidor"), campo("porta")
	usuario, senha := campo("usuario"), campo("senha")
	para, assunto, corpoHtml := campo("para"), campo("assunto"), campo("corpoHtml")
	de := campo("de")
	if de == "" {
		de = usuario
	}

	// Sanitiza cabeçalhos contra CRLF Injection (SMTP Header Injection)
	replacer := strings.NewReplacer("\r", "", "\n", "")
	deClean := replacer.Replace(de)
	assuntoClean := replacer.Replace(assunto)
	var destinatarios []string // "a@x, b@y" = dois RCPT TO
	for _, d := range strings.Split(replacer.Replace(para), ",") {
		if d = strings.TrimSpace(d); d != "" {
			destinatarios = append(destinatarios, d)
		}
	}
	paraClean := strings.Join(destinatarios, ", ")

	if servidor == "" || porta == "" {
		// Simulação de envio local para desenvolvimento/teste para evitar bloqueios ou exceptions
		fmt.Printf("📧 [SIMULAÇÃO EMAIL] De: '%s' | Para: '%s' | Assunto: '%s'\n", deClean, paraClean, assuntoClean)
		return hrp.Booleano(true), nil
	}

	var auth smtp.Auth
	if usuario != "" {
		auth = smtp.PlainAuth("", usuario, senha, servidor)
	}

	if len(destinatarios) == 0 {
		return hrp.Booleano(false), hrp.NewErroF(hrp.ValorErro, "email.enviar: informe ao menos um destinatário em 'para'")
	}

	// Subject em UTF-8 precisa de codificação RFC 2047 ("Ação" chegava corrompido)
	msg := fmt.Appendf(nil, "From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s", deClean, paraClean, mime.QEncoding.Encode("utf-8", assuntoClean), corpoHtml)

	addr := net.JoinHostPort(servidor, porta)
	remetente := usuario // servidores com auth (ex.: Gmail) exigem a conta autenticada no envelope
	if remetente == "" {
		remetente = deClean
	}
	err := smtp.SendMail(addr, auth, remetente, destinatarios, msg)
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
