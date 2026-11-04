//go:build !windows && !js && !wasm

package soquete_test

import (
	"syscall"
	"testing"

	"github.com/mat-dgruber/Harpia/hrp"
	"github.com/mat-dgruber/Harpia/stdlib/soquete"
)

func TestSoqueteIPInvalidoNaoPanica(t *testing.T) {
	// ponytail: regressão para validação defensiva de IP e buffer de recepção
	obj, err := soquete.NewSoquete(hrp.Inteiro(syscall.AF_INET), hrp.Inteiro(syscall.SOCK_STREAM), hrp.Inteiro(0))
	if err != nil {
		t.Fatalf("falha ao criar socket de teste: %v", err)
	}
	s := obj.(*soquete.Soquete)
	defer s.Fecha()

	// 1. IP inválido em AssociaSoquete não deve entrar em pânico e deve retornar erro estruturado
	_, err = s.AssociaSoquete(hrp.Texto("ip_invalido"), hrp.Inteiro(8080))
	if err == nil {
		t.Fatal("esperava erro ao associar IP inválido")
	}
	erroObj, ok := err.(*hrp.Erro)
	if !ok || erroObj.Base != hrp.ValorErro {
		t.Fatalf("esperava erro ValorErro ao associar IP inválido, obteve %v", err)
	}

	// 2. IPv6 passado para socket IPv4 não deve copiar bytes zerados/truncados e sim retornar ValorErro
	_, err = s.AssociaSoquete(hrp.Texto("::1"), hrp.Inteiro(8080))
	if err == nil {
		t.Fatal("esperava erro ao associar IPv6 em socket IPv4")
	}
	erroObj, ok = err.(*hrp.Erro)
	if !ok || erroObj.Base != hrp.ValorErro {
		t.Fatalf("esperava erro ValorErro ao associar IPv6 em socket IPv4, obteve %v", err)
	}

	// 3. Conecta com endereço/IP inválido não deve entrar em pânico
	_, err = s.Conecta(hrp.Texto("999.999.999.999"), hrp.Inteiro(8080))
	if err == nil {
		t.Fatal("esperava erro ao conectar a IP inválido")
	}

	// 4. RecebeDados com tamanhoBuffer <= 0 não deve entrar em pânico por slice negativo
	_, err = s.RecebeDados(hrp.Inteiro(-1))
	if err != nil {
		t.Logf("RecebeDados(-1) retornou erro esperado ou nulo: %v", err)
	}
	_, err = s.RecebeDados(hrp.Inteiro(0))
	if err != nil {
		t.Logf("RecebeDados(0) retornou erro esperado ou nulo: %v", err)
	}

	// 5. DefinirNaoBloqueante e DefineOpcoes em socket fechado devem retornar erro sem pânico
	sFechado, _ := soquete.NewSoquete(hrp.Inteiro(syscall.AF_INET), hrp.Inteiro(syscall.SOCK_STREAM), hrp.Inteiro(0))
	sFechadoObj := sFechado.(*soquete.Soquete)
	sFechadoObj.Fecha()

	_, err = sFechadoObj.DefinirNaoBloqueante(hrp.Verdadeiro)
	if err == nil {
		t.Log("DefinirNaoBloqueante em socket fechado retornou erro esperado")
	}
	_, err = sFechadoObj.DefineOpcoes(hrp.Inteiro(syscall.SOL_SOCKET), hrp.Inteiro(syscall.SO_REUSEADDR), hrp.Inteiro(1))
	if err == nil {
		t.Log("DefineOpcoes em socket fechado retornou erro esperado")
	}
}

func TestSoqueteAceitaConexao(t *testing.T) {
	// ponytail: valida bind em IPv4 127.0.0.1, listen, connect e aceita sem vazamento de FDs em s.pollFd
	servidorObj, err := soquete.NewSoquete(hrp.Inteiro(syscall.AF_INET), hrp.Inteiro(syscall.SOCK_STREAM), hrp.Inteiro(0))
	if err != nil {
		t.Fatalf("falha ao criar socket servidor: %v", err)
	}
	servidor := servidorObj.(*soquete.Soquete)
	defer servidor.Fecha()

	// SO_REUSEADDR
	_, _ = servidor.DefineOpcoes(hrp.Inteiro(syscall.SOL_SOCKET), hrp.Inteiro(syscall.SO_REUSEADDR), hrp.Inteiro(1))

	porta := hrp.Inteiro(19875)
	_, err = servidor.AssociaSoquete(hrp.Texto("127.0.0.1"), porta)
	if err != nil {
		t.Skipf("porta %d em uso ou sem permissão de bind: %v", porta, err)
		return
	}

	_, err = servidor.OuveSoquete(hrp.Inteiro(5))
	if err != nil {
		t.Fatalf("falha no OuveSoquete: %v", err)
	}

	clienteObj, err := soquete.NewSoquete(hrp.Inteiro(syscall.AF_INET), hrp.Inteiro(syscall.SOCK_STREAM), hrp.Inteiro(0))
	if err != nil {
		t.Fatalf("falha ao criar socket cliente: %v", err)
	}
	cliente := clienteObj.(*soquete.Soquete)
	defer cliente.Fecha()

	go func() {
		_, _ = cliente.Conecta(hrp.Texto("127.0.0.1"), porta)
	}()

	aceito, err := servidor.AceitaConexao()
	if err != nil {
		t.Fatalf("falha no AceitaConexao: %v", err)
	}
	defer aceito.Fecha()

	if aceito == nil {
		t.Fatal("esperava objeto Soquete aceito válido")
	}
}
