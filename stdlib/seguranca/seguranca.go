// Package seguranca implementa utilitários focados em segurança corporativa e programação defensiva,
// contendo filtros sanitizadores para mitigar vetores de ataques OWASP como XSS e SQL Injection.
package seguranca

import (
	"html"
	"strings"

	"github.com/mat-dgruber/Harpia/hrp"
)

// met_sanitizarHtml implementa 'sanitizarHtml(texto)' em nível de script Harpia.
// Previne ataques de injeção de scripts maliciosos (XSS) codificando tags HTML como entidades seguras (&lt;, &gt;, etc.).
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

// met_sanitizarSqlArgumento implementa 'sanitizarSqlArgumento(texto)' em nível de script Harpia.
// Mitiga vulnerabilidades do tipo SQL Injection em strings e queries SQL concatenadas manualmente.
// NOTA: Recomenda-se utilizar parametrização com marcadores de binding em vez de concatenação crua.
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

var _sanitizarHtml = hrp.NewMetodoOuPanic("sanitizarHtml", met_sanitizarHtml, "Escapa tags HTML especiais de uma string para mitigar riscos de ataque de injeção XSS.")
var _sanitizarSqlArgumento = hrp.NewMetodoOuPanic("sanitizarSqlArgumento", met_sanitizarSqlArgumento, "Sanitiza caracteres e marcadores especiais para prevenir injeção de comandos SQL (SQL Injection).")

func init() {
	// Registra o módulo 'seguranca' na biblioteca padrão do Harpia.
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
