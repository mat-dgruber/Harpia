package bd

import (
	"database/sql"
	"testing"

	"github.com/mat-dgruber/Harpia/hrp"
)

func conexaoMemoria(t *testing.T) *ConexaoSQL {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1) // :memory: é por conexão
	t.Cleanup(func() { db.Close() })
	if _, err := db.Exec("CREATE TABLE u (id INTEGER PRIMARY KEY, nome TEXT)"); err != nil {
		t.Fatal(err)
	}
	return &ConexaoSQL{db: db, driver: "sqlite"}
}

func metodo(t *testing.T, obj interface {
	M__obtem_attributo__(string) (hrp.Objeto, error)
}, nome string) hrp.Objeto {
	t.Helper()
	m, err := obj.M__obtem_attributo__(nome)
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func contarLinhas(t *testing.T, c *ConexaoSQL) int {
	t.Helper()
	var n int
	if err := c.db.QueryRow("SELECT COUNT(*) FROM u").Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

// M8: pânico no callback da transação desfaz o que foi feito.
func TestTransacaoFazRollbackEmPanico(t *testing.T) {
	c := conexaoMemoria(t)
	cb := hrp.NewMetodoOuPanic("cb", func(_ hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
		tx := args[0].(*TransacaoSQL)
		if _, err := hrp.Chamar(metodo(t, tx, "executar"), hrp.Tupla{hrp.Texto("INSERT INTO u (nome) VALUES ('a')")}); err != nil {
			t.Fatal(err)
		}
		panic("falha no meio da transação")
	}, "")
	func() {
		defer func() { _ = recover() }()
		_, _ = hrp.Chamar(metodo(t, c, "transacao"), hrp.Tupla{cb})
	}()
	if n := contarLinhas(t, c); n != 0 {
		t.Fatalf("pânico deveria reverter a transação; linhas = %d", n)
	}
}

// M9: executar só devolve id em INSERT; executarDetalhado expõe linhas afetadas.
func TestExecutarRetornos(t *testing.T) {
	c := conexaoMemoria(t)
	exec := metodo(t, c, "executar")
	id, err := hrp.Chamar(exec, hrp.Tupla{hrp.Texto("INSERT INTO u (nome) VALUES ('a')")})
	if err != nil || id != hrp.Inteiro(1) {
		t.Fatalf("INSERT deveria devolver id 1, veio %v (%v)", id, err)
	}
	res, err := hrp.Chamar(exec, hrp.Tupla{hrp.Texto("UPDATE u SET nome = 'b'")})
	if err != nil || res != hrp.Nulo {
		t.Fatalf("UPDATE não deveria devolver id antigo, veio %v (%v)", res, err)
	}
	det, err := hrp.Chamar(metodo(t, c, "executarDetalhado"), hrp.Tupla{hrp.Texto("UPDATE u SET nome = 'c'")})
	if err != nil {
		t.Fatal(err)
	}
	if m := det.(hrp.Mapa); m["linhasAfetadas"] != hrp.Inteiro(1) {
		t.Fatalf("linhasAfetadas = %v", m["linhasAfetadas"])
	}
}

// M21: "==" (sintaxe da linguagem, usada na documentação) funciona como "=".
func TestOndeAceitaIgualIgual(t *testing.T) {
	if err := validarOperadorSQL("=="); err != nil {
		t.Fatalf("'==' deveria ser aceito: %v", err)
	}
}
