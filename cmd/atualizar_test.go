package cmd

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestChecarVersaoMock(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[{"name": "v1.5.0"}]`))
	}))
	defer ts.Close()

	versao, err := obterUltimaVersaoURL(ts.URL)
	if err != nil {
		t.Fatalf("erro ao obter versao: %v", err)
	}
	if versao != "1.5.0" {
		t.Errorf("esperava 1.5.0, obteve %s", versao)
	}
}

func TestExtrairZipNativo(t *testing.T) {
	zipPath := createTestZip(t, "testfile.txt", "hello zip")
	defer os.Remove(zipPath)

	destDir, err := os.MkdirTemp("", "testzip_dest")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(destDir)

	if err := extrairZipNativo(zipPath, destDir); err != nil {
		t.Fatalf("falha ao extrair zip: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(destDir, "testfile.txt"))
	if err != nil {
		t.Fatalf("falha ao ler arquivo extraido: %v", err)
	}

	if string(content) != "hello zip" {
		t.Errorf("conteudo esperado 'hello zip', obteve '%s'", string(content))
	}
}

func TestExtrairTarGzNativo(t *testing.T) {
	tarGzPath := createTestTarGz(t, "testfile.txt", "hello tar")
	defer os.Remove(tarGzPath)

	destDir, err := os.MkdirTemp("", "testtargz_dest")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(destDir)

	if err := extrairTarGzNativo(tarGzPath, destDir); err != nil {
		t.Fatalf("falha ao extrair tar.gz: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(destDir, "testfile.txt"))
	if err != nil {
		t.Fatalf("falha ao ler arquivo extraido: %v", err)
	}

	if string(content) != "hello tar" {
		t.Errorf("conteudo esperado 'hello tar', obteve '%s'", string(content))
	}
}

func TestComandoAtualizarComandaOuAlias(t *testing.T) {
	cmd := comandoAtualizar()
	if cmd.Use != "atualizar" {
		t.Errorf("esperava Use 'atualizar', obteve '%s'", cmd.Use)
	}

	temAlias := false
	for _, alias := range cmd.Aliases {
		if alias == "atualize" {
			temAlias = true
			break
		}
	}
	if !temAlias {
		t.Errorf("esperava alias 'atualize' no comando atualizar")
	}
}

func createTestZip(t *testing.T, filename, content string) string {
	tmpFile, err := os.CreateTemp("", "test*.zip")
	if err != nil {
		t.Fatal(err)
	}
	defer tmpFile.Close()

	zw := zip.NewWriter(tmpFile)
	fw, err := zw.Create(filename)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fw.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return tmpFile.Name()
}

func createTestTarGz(t *testing.T, filename, content string) string {
	tmpFile, err := os.CreateTemp("", "test*.tar.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer tmpFile.Close()

	gw := gzip.NewWriter(tmpFile)
	tw := tar.NewWriter(gw)

	hdr := &tar.Header{
		Name: filename,
		Mode: 0644,
		Size: int64(len(content)),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gw.Close(); err != nil {
		t.Fatal(err)
	}
	return tmpFile.Name()
}

func TestComandoAtualizarChecarFlag(t *testing.T) {
	cmd := comandoAtualizar()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	flag := cmd.Flags().Lookup("checar")
	if flag == nil {
		t.Fatalf("flag --checar nao encontrada")
	}

	if err := cmd.ParseFlags([]string{"--checar"}); err != nil {
		t.Fatalf("erro ao processar flag --checar: %v", err)
	}

	val, err := cmd.Flags().GetBool("checar")
	if err != nil {
		t.Fatalf("erro ao obter valor da flag checar: %v", err)
	}
	if !val {
		t.Errorf("esperava flag checar como true apos parse de '--checar', obteve false")
	}
}
