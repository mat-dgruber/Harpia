package cmd

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCmdStressar(t *testing.T) {
	codigo := `
	var x = 10;
	`
	dir := t.TempDir()
	caminho := filepath.Join(dir, "teste_stress.pt")
	err := os.WriteFile(caminho, []byte(codigo), 0644)
	if err != nil {
		t.Fatalf("Erro ao criar arquivo: %v", err)
	}

	// Captura stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := comandoStressar()
	cmd.SetArgs([]string{caminho, "-c", "2", "-r", "5"})

	err = cmd.Execute()
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("Erro ao rodar comando stressar: %v", err)
	}

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "RELATÓRIO DO TESTE DE ESTRESSE") {
		t.Errorf("Saída inesperada do stressar:\n%s", output)
	}
}

func TestCmdDepurarDapHandshake(t *testing.T) {
	// Obtém porta livre dinamicamente
	l, errL := net.Listen("tcp", "127.0.0.1:0")
	if errL != nil {
		t.Fatalf("Erro ao obter porta livre: %v", errL)
	}
	addr := l.Addr().String()
	partes := strings.Split(addr, ":")
	portaLivre := partes[len(partes)-1]
	l.Close()

	cmd := comandoDepurar()
	cmd.SetArgs([]string{"--porta", portaLivre})

	go func() {
		_ = cmd.Execute()
	}()

	// Aguarda o servidor subir
	time.Sleep(100 * time.Millisecond)

	conn, err := net.Dial("tcp", "127.0.0.1:"+portaLivre)
	if err != nil {
		t.Fatalf("Erro ao conectar no servidor DAP: %v", err)
	}
	defer conn.Close()

	// Envia cabeçalho e corpo DAP com nova linha no final para o bufio.Scanner
	payload := `{"seq":15,"type":"request","command":"initialize"}`
	msg := fmt.Sprintf("Content-Length: %d\r\n\r\n%s\n", len(payload), payload)
	_, err = conn.Write([]byte(msg))
	if err != nil {
		t.Fatalf("Erro ao enviar dados DAP: %v", err)
	}

	// Lê resposta
	buf := make([]byte, 1024)
	n, errRead := conn.Read(buf)
	if errRead != nil {
		t.Fatalf("Erro ao ler dados do DAP: %v", errRead)
	}

	res := string(buf[:n])
	if !strings.Contains(res, `"type":"response"`) || !strings.Contains(res, "supportsConfigurationDoneRequest") {
		t.Errorf("Resposta DAP inválida ou incompleta: %s", res)
	}
}
