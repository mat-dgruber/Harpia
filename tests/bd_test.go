package tests

import (
	"testing"

	"github.com/natanfeitosa/portuscript/ptst"
	_ "github.com/natanfeitosa/portuscript/stdlib"
)

func TestBDModulo(t *testing.T) {
	ctx := ptst.NewContexto(ptst.OpcsContexto{})
	defer ctx.Terminar()

	codigo := `
	de "bd" importe conectarSqlite

	var conn = conectarSqlite(":memory:")
	conn.executar("CREATE TABLE usuarios (id INTEGER PRIMARY KEY, nome TEXT, idade INTEGER)")

	# Testando inserção via Query Builder
	var qb = conn.tabela("usuarios")
	qb.inserir({"id": 1, "nome": "Portuscript", "idade": 3})
	qb.inserir({"id": 2, "nome": "Go", "idade": 13})

	# Testando obterUm
	var qb1 = conn.tabela("usuarios")
	qb1.onde("id", "=", 1)
	var usuario = qb1.obterUm()
	var nome = usuario["nome"]
	var idade = usuario["idade"]

	# Testando obterMuitos
	var qb2 = conn.tabela("usuarios")
	var todos = qb2.obterMuitos()
	var total = tamanho(todos)

	# Testando atualizar e deletar
	var qb3 = conn.tabela("usuarios")
	qb3.onde("id", "=", 2)
	qb3.atualizar({"idade": 14})

	var qb4 = conn.tabela("usuarios")
	qb4.onde("id", "=", 2)
	var usuarioAtualizado = qb4.obterUm()
	var idadeAtualizada = usuarioAtualizado["idade"]

	var qb5 = conn.tabela("usuarios")
	qb5.onde("id", "=", 1)
	qb5.deletar()

	var qb6 = conn.tabela("usuarios")
	var todosDepois = qb6.obterMuitos()
	var totalDepois = tamanho(todosDepois)

	conn.fechar()
	`

	res, err := ptst.ExecutarString(ctx, codigo)
	if err != nil {
		t.Fatalf("Erro ao executar script de teste de BD: %v", err)
	}

	valNome, err := res.Escopo.ObterValor("nome")
	if err != nil {
		t.Fatal(err)
	}
	if string(valNome.(ptst.Texto)) != "Portuscript" {
		t.Errorf("Nome esperado 'Portuscript', obteve '%v'", valNome)
	}

	valIdade, err := res.Escopo.ObterValor("idade")
	if err != nil {
		t.Fatal(err)
	}
	if int(valIdade.(ptst.Inteiro)) != 3 {
		t.Errorf("Idade esperada 3, obteve '%v'", valIdade)
	}

	valTotal, err := res.Escopo.ObterValor("total")
	if err != nil {
		t.Fatal(err)
	}
	if int(valTotal.(ptst.Inteiro)) != 2 {
		t.Errorf("Total de usuários esperado 2, obteve '%v'", valTotal)
	}

	valIdadeAt, err := res.Escopo.ObterValor("idadeAtualizada")
	if err != nil {
		t.Fatal(err)
	}
	if int(valIdadeAt.(ptst.Inteiro)) != 14 {
		t.Errorf("Idade atualizada esperada 14, obteve '%v'", valIdadeAt)
	}

	valTotalDepois, err := res.Escopo.ObterValor("totalDepois")
	if err != nil {
		t.Fatal(err)
	}
	if int(valTotalDepois.(ptst.Inteiro)) != 1 {
		t.Errorf("Total depois de deletar esperado 1, obteve '%v'", valTotalDepois)
	}
}
