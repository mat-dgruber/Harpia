package seguranca

import (
	"html"
	"strings"

	"github.com/mat-dgruber/Harpia/hrp"
)

// met_sanitizarHtml implementa 'sanitizarHtml(texto)' -> Previne XSS
func met_sanitizarHtml(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if err := hrp.VerificaNumeroArgumentos("sanitizarHtml", false, args, 1, 1); err != nil {
		return nil, err
	}

	texto, err := hrp.NewTexto(args[0])
	if err != nil {
		return nil, err
	}

	escapado := html.EscapeString(string(texto.(hrp.Texto)))
	return hrp.Texto(escapado), nil
}

// met_sanitizarSqlArgumento implementa 'sanitizarSqlArgumento(texto)' -> Previne SQL Injection em queries brutas
func met_sanitizarSqlArgumento(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if err := hrp.VerificaNumeroArgumentos("sanitizarSqlArgumento", false, args, 1, 1); err != nil {
		return nil, err
	}

	texto, err := hrp.NewTexto(args[0])
	if err != nil {
		return nil, err
	}

	limpo := strings.ReplaceAll(string(texto.(hrp.Texto)), "'", "''")
	limpo = strings.ReplaceAll(limpo, ";", "")
	limpo = strings.ReplaceAll(limpo, "--", "")
	return hrp.Texto(limpo), nil
}

var _sanitizarHtml = hrp.NewMetodoOuPanic("sanitizarHtml", met_sanitizarHtml, "")
var _sanitizarSqlArgumento = hrp.NewMetodoOuPanic("sanitizarSqlArgumento", met_sanitizarSqlArgumento, "")

func init() {
	hrp.RegistraModuloImpl(&hrp.ModuloImpl{
		Info: hrp.ModuloInfo{
			Nome:    "seguranca",
			Arquivo: "stdlib/seguranca",
		},
		Metodos: []*hrp.Metodo{
			_sanitizarHtml,
			_sanitizarSqlArgumento,
		},
	})
}
