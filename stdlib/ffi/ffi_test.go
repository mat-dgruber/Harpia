package ffi

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mat-dgruber/Harpia/hrp"
)

// chamarFuncaoFFI abre a biblioteca, resolve `nome` e chama com `args`.
func chamarFuncaoFFI(t *testing.T, caminho, nome string, args hrp.Tupla) error {
	t.Helper()
	ctx := hrp.NewContexto(hrp.OpcsContexto{})
	defer ctx.Terminar()

	libObj, err := met_ffi_abrir(nil, hrp.Tupla{hrp.Texto(caminho)})
	if err != nil {
		t.Fatalf("Erro ao abrir lib: %v", err)
	}
	obter, err := libObj.(*Biblioteca).M__obtem_attributo__("obterFuncao")
	if err != nil {
		t.Fatal(err)
	}
	fn, err := hrp.Chamar(obter, hrp.Tupla{hrp.Texto(nome)})
	if err != nil {
		t.Fatal(err)
	}
	_, err = hrp.Chamar(fn, args)
	return err
}

func TestFFIBibliotecaAusenteDaErro(t *testing.T) {
	err := chamarFuncaoFFI(t, "libinexistente.so", "soma", hrp.Tupla{hrp.Decimal(1), hrp.Decimal(2)})
	if err == nil || !strings.Contains(err.Error(), "não pôde ser carregada") {
		t.Fatalf("esperava erro de biblioteca ausente, veio %v", err)
	}
}

// Com a biblioteca presente, a chamada não pode fingir sucesso (antes devolvia Nulo).
func TestFFIChamadaNativaNaoFingeSucesso(t *testing.T) {
	lib := filepath.Join(t.TempDir(), "libfalsa.so")
	if err := os.WriteFile(lib, []byte("x"), 0600); err != nil {
		t.Fatal(err)
	}
	err := chamarFuncaoFFI(t, lib, "soma", hrp.Tupla{hrp.Decimal(1), hrp.Decimal(2)})
	if err == nil || !strings.Contains(err.Error(), "não implementada") {
		t.Fatalf("esperava erro explícito de FFI não implementada, veio %v", err)
	}
}
