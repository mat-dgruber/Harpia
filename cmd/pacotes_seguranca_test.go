//go:build !js && !wasm

package cmd

import (
	"archive/zip"
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// emPastaTemporaria roda o teste dentro de um diretório limpo (pt_modulos é relativo ao cwd).
func emPastaTemporaria(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	antigo, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(antigo) })
	return dir
}

func zipCom(t *testing.T, nome string, conteudo []byte, modo os.FileMode) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	h := &zip.FileHeader{Name: nome, Method: zip.Deflate}
	if modo != 0 {
		h.SetMode(modo)
	}
	w, err := zw.CreateHeader(h)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write(conteudo); err != nil {
		t.Fatal(err)
	}
	zw.Close()
	return buf.Bytes()
}

func servirBytes(t *testing.T, b []byte) string {
	t.Helper()
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.Write(b) }))
	t.Cleanup(s.Close)
	return s.URL
}

// H6: o nome da dependência não pode tirar a instalação de pt_modulos.
func TestPacoteNomeComTraversalERecusado(t *testing.T) {
	dir := emPastaTemporaria(t)
	url := servirBytes(t, zipCom(t, "x.hrp", []byte("1"), 0))
	for _, nome := range []string{"../../fora", "../fora", "a/b", "..", "."} {
		if _, err := baixarEExtrairPacote(nome, url, ""); err == nil {
			t.Errorf("nome %q deveria ser recusado", nome)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, "..", "fora")); err == nil {
		t.Fatal("arquivo escrito fora de pt_modulos")
	}
}

// H7: zip bomb (muito maior descomprimido) é recusada.
func TestPacoteZipBombERecusado(t *testing.T) {
	emPastaTemporaria(t)
	antigo := limiteDescompressaoPacote
	limiteDescompressaoPacote = 1 << 20 // 1 MB para o teste ser rápido
	t.Cleanup(func() { limiteDescompressaoPacote = antigo })
	url := servirBytes(t, zipCom(t, "bomba.bin", make([]byte, 8<<20), 0))
	if _, err := baixarEExtrairPacote("bomba", url, ""); err == nil || !strings.Contains(err.Error(), "limite") {
		t.Fatalf("esperava erro de limite de descompressão, veio %v", err)
	}
}

// H7: servidor que nunca responde não trava a instalação.
func TestPacoteDownloadTemTimeout(t *testing.T) {
	emPastaTemporaria(t)
	parado := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-time.After(10 * time.Second):
		}
	}))
	t.Cleanup(parado.Close)
	antigo := clientePacotes
	clientePacotes = &http.Client{Timeout: 300 * time.Millisecond}
	t.Cleanup(func() { clientePacotes = antigo })
	inicio := time.Now()
	if _, err := baixarEExtrairPacote("lento", parado.URL, ""); err == nil {
		t.Fatal("esperava erro de timeout")
	}
	if d := time.Since(inicio); d > 3*time.Second {
		t.Fatalf("timeout não respeitado: %v", d)
	}
}

// H7: http:// fora de loopback é recusado (MITM).
func TestPacoteExigeHTTPS(t *testing.T) {
	emPastaTemporaria(t)
	if _, err := baixarEExtrairPacote("x", "http://exemplo.invalid/pacote.zip", ""); err == nil || !strings.Contains(err.Error(), "https") {
		t.Fatalf("esperava recusa de http://, veio %v", err)
	}
}

// H7: bits setuid/setgid do zip não chegam ao disco.
func TestPacoteNaoPreservaSetuid(t *testing.T) {
	emPastaTemporaria(t)
	url := servirBytes(t, zipCom(t, "bin/ferramenta", []byte("#!/bin/sh"), os.ModeSetuid|0o4755))
	if _, err := baixarEExtrairPacote("perm", url, ""); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(filepath.Join("pt_modulos", "perm", "bin", "ferramenta"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode()&(os.ModeSetuid|os.ModeSetgid) != 0 || info.Mode().Perm() != 0o755 {
		t.Fatalf("modo gravado = %v", info.Mode())
	}
}

// P1: o índice precisa trazer sha256; a versão "última" é a maior semver (não a ordem do map).
func TestRegistroExigeSha256EVersaoDeterministica(t *testing.T) {
	indice := `{"pacotes": {
		"semhash": {"versoes": {"1.0.0": {"url": "https://github.com/mat-dgruber/x/releases/download/v1.0.0/x.zip"}}},
		"ok": {"versoes": {
			"1.9.0":  {"url": "https://github.com/mat-dgruber/ok/releases/download/v1.9.0/ok.zip", "sha256": "aa"},
			"1.10.0": {"url": "https://github.com/mat-dgruber/ok/releases/download/v1.10.0/ok.zip", "sha256": "bb"},
			"1.2.0":  {"url": "https://github.com/mat-dgruber/ok/releases/download/v1.2.0/ok.zip", "sha256": "cc"}
		}}
	}}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { io.WriteString(w, indice) }))
	t.Cleanup(srv.Close)
	t.Setenv("HARPIA_REGISTRO", srv.URL)

	if _, _, _, err := obterPacoteDoRegistro("semhash", ""); err == nil || !strings.Contains(err.Error(), "sha256") {
		t.Fatalf("entrada sem sha256 deveria ser recusada, veio %v", err)
	}
	for i := 0; i < 20; i++ {
		versao, _, sha, err := obterPacoteDoRegistro("ok", "")
		if err != nil || versao != "1.10.0" || sha != "bb" {
			t.Fatalf("esperava 1.10.0/bb, veio %s/%s (%v)", versao, sha, err)
		}
	}
	if versao, _, _, err := obterPacoteDoRegistro("ok", "1"); err != nil || versao != "1.10.0" {
		t.Fatalf("prefixo '1' deveria resolver para a maior 1.x, veio %s (%v)", versao, err)
	}
}

// P1: no índice padrão, só URLs de release do próprio projeto são aceitas.
func TestRegistroPadraoRestringeOrigem(t *testing.T) {
	if urlPermitidaNoRegistro("https://evil.example/x.zip", registroPadrao) {
		t.Error("índice padrão não pode apontar para fora de github.com/mat-dgruber")
	}
	if !urlPermitidaNoRegistro("https://github.com/mat-dgruber/pkg/releases/download/v1/x.zip", registroPadrao) {
		t.Error("release do próprio projeto deveria ser aceita")
	}
	if !urlPermitidaNoRegistro("https://espelho.interno/x.zip", "https://espelho.interno/pacotes.json") {
		t.Error("índice customizado (HARPIA_REGISTRO) decide as próprias origens")
	}
}
