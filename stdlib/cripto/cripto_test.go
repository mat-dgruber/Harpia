package cripto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"strings"
	"testing"

	"github.com/mat-dgruber/Harpia/hrp"
)

func cifrar(t *testing.T, texto, chave string) string {
	t.Helper()
	res, err := met_cripto_cifrar(nil, hrp.Tupla{hrp.Texto(texto), hrp.Texto(chave)})
	if err != nil {
		t.Fatal(err)
	}
	return string(res.(hrp.Texto))
}

func decifrar(texto, chave string) (string, error) {
	res, err := met_cripto_decifrar(nil, hrp.Tupla{hrp.Texto(texto), hrp.Texto(chave)})
	if err != nil {
		return "", err
	}
	return string(res.(hrp.Texto)), nil
}

// M11: cifrar usa KDF com sal (envelope v2) e sal novo a cada chamada.
func TestCifrarUsaEnvelopeV2ComSal(t *testing.T) {
	a := cifrar(t, "segredo", "senha")
	b := cifrar(t, "segredo", "senha")
	if !strings.HasPrefix(a, "v2:") {
		t.Fatalf("esperava envelope v2, veio %q", a)
	}
	salA, _ := base64.StdEncoding.DecodeString(strings.TrimPrefix(a, "v2:"))
	salB, _ := base64.StdEncoding.DecodeString(strings.TrimPrefix(b, "v2:"))
	if string(salA[:16]) == string(salB[:16]) {
		t.Fatal("o sal deveria mudar a cada chamada")
	}
	if got, err := decifrar(a, "senha"); err != nil || got != "segredo" {
		t.Fatalf("ida e volta falhou: %q %v", got, err)
	}
	if _, err := decifrar(a, "errada"); err == nil {
		t.Fatal("chave errada deveria falhar")
	}
}

// Compatibilidade: textos cifrados no formato antigo (sha256 da chave, sem sal) continuam legíveis.
func TestDecifrarFormatoLegado(t *testing.T) {
	chave := sha256.Sum256([]byte("senha"))
	block, _ := aes.NewCipher(chave[:])
	gcm, _ := cipher.NewGCM(block)
	nonce := make([]byte, gcm.NonceSize())
	_, _ = rand.Read(nonce)
	legado := base64.StdEncoding.EncodeToString(gcm.Seal(nonce, nonce, []byte("antigo"), nil))
	if got, err := decifrar(legado, "senha"); err != nil || got != "antigo" {
		t.Fatalf("formato legado deveria decifrar: %q %v", got, err)
	}
}
