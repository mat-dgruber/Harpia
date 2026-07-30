//go:build !js && !wasm

package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseLockfileJSON(t *testing.T) {
	jsonContent := []byte(`{
		"versao_lock": "1.0",
		"pacotes": {
			"http-utils": {
				"versao": "1.2.0",
				"url": "https://harpia.dev/pkgs/http-utils-1.2.0.zip",
				"sha256": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
			}
		}
	}`)

	lock, err := parseLockfile(jsonContent)
	if err != nil {
		t.Fatalf("parseLockfile falhou para JSON valido: %v", err)
	}

	if lock.VersaoLock != "1.0" {
		t.Errorf("VersaoLock esperada '1.0', obteve '%s'", lock.VersaoLock)
	}

	pkg, ok := lock.Pacotes["http-utils"]
	if !ok {
		t.Fatalf("Pacote 'http-utils' nao encontrado no lockfile")
	}

	if pkg.Versao != "1.2.0" {
		t.Errorf("Versao esperada '1.2.0', obteve '%s'", pkg.Versao)
	}
	if pkg.URL != "https://harpia.dev/pkgs/http-utils-1.2.0.zip" {
		t.Errorf("URL incorreta: %s", pkg.URL)
	}
	if pkg.SHA256 != "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855" {
		t.Errorf("SHA256 incorreto: %s", pkg.SHA256)
	}
}

func TestParseLockfileHRP(t *testing.T) {
	hrpContent := []byte(`# Lockfile de pacotes Harpia
versao_lock = "1.0"

[http-utils]
versao = "1.2.0"
url = "https://harpia.dev/pkgs/http-utils-1.2.0.zip"
sha256 = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"

[json-parser]
versao = "2.0.1"
url = "https://harpia.dev/pkgs/json-parser-2.0.1.zip"
sha256 = "abc12345"
`)

	lock, err := parseLockfile(hrpContent)
	if err != nil {
		t.Fatalf("parseLockfile falhou para formato HRP: %v", err)
	}

	if lock.VersaoLock != "1.0" {
		t.Errorf("VersaoLock esperada '1.0', obteve '%s'", lock.VersaoLock)
	}

	if len(lock.Pacotes) != 2 {
		t.Fatalf("Quantidade esperada de pacotes: 2, obteve: %d", len(lock.Pacotes))
	}

	pkg1, ok := lock.Pacotes["http-utils"]
	if !ok {
		t.Fatalf("Pacote 'http-utils' nao encontrado")
	}
	if pkg1.Versao != "1.2.0" || pkg1.SHA256 != "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855" {
		t.Errorf("Valores incorretos para 'http-utils': %+v", pkg1)
	}

	pkg2, ok := lock.Pacotes["json-parser"]
	if !ok {
		t.Fatalf("Pacote 'json-parser' nao encontrado")
	}
	if pkg2.Versao != "2.0.1" || pkg2.URL != "https://harpia.dev/pkgs/json-parser-2.0.1.zip" || pkg2.SHA256 != "abc12345" {
		t.Errorf("Valores incorretos para 'json-parser': %+v", pkg2)
	}
}

func TestSalvarLockfile(t *testing.T) {
	tempDir := t.TempDir()
	caminhoLock := filepath.Join(tempDir, "pacote.lock")

	lock := &Lockfile{
		VersaoLock: "1.0",
		Pacotes: map[string]PacoteLockInfo{
			"teste": {
				Versao: "0.1.0",
				URL:    "https://example.com/teste.zip",
				SHA256: "123456",
			},
		},
	}

	err := salvarLockfile(caminhoLock, lock)
	if err != nil {
		t.Fatalf("salvarLockfile falhou: %v", err)
	}

	conteudo, err := os.ReadFile(caminhoLock)
	if err != nil {
		t.Fatalf("falhou ao ler arquivo salvo: %v", err)
	}

	lockLido, err := parseLockfile(conteudo)
	if err != nil {
		t.Fatalf("parseLockfile falhou ao reler lockfile salvo: %v", err)
	}

	if lockLido.VersaoLock != "1.0" {
		t.Errorf("VersaoLock incorreta: %s", lockLido.VersaoLock)
	}

	pkg, ok := lockLido.Pacotes["teste"]
	if !ok {
		t.Fatalf("Pacote 'teste' nao encontrado no lockfile salvo")
	}
	if pkg.Versao != "0.1.0" || pkg.SHA256 != "123456" {
		t.Errorf("Dados incorretos do pacote salvo: %+v", pkg)
	}
}
