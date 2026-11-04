package seguranca

import (
	"testing"

	"github.com/mat-dgruber/Harpia/hrp"
)

func TestSanitizarSqlArgumento(t *testing.T) {
	entrada := "admin' OR 1=1; -- /* comment */ \\ \x00"
	res, err := met_sanitizarSqlArgumento(nil, hrp.Tupla{hrp.Texto(entrada)})
	if err != nil {
		t.Fatalf("Erro ao sanitizar SQL: %v", err)
	}

	resStr := string(res.(hrp.Texto))
	if resStr != "admin'' OR 1=1   comment  \\\\ " {
		t.Errorf("Resultado inesperado da sanitização: %q", resStr)
	}
}

func TestSanitizarHtml(t *testing.T) {
	entrada := "<script>alert('xss')</script>"
	res, err := met_sanitizarHtml(nil, hrp.Tupla{hrp.Texto(entrada)})
	if err != nil {
		t.Fatalf("Erro ao sanitizar HTML: %v", err)
	}

	resStr := string(res.(hrp.Texto))
	if resStr != "&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;" {
		t.Errorf("Resultado inesperado da sanitização HTML: %q", resStr)
	}
}
