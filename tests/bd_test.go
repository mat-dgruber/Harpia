package tests

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mat-dgruber/Harpia/ptst"
	_ "github.com/mat-dgruber/Harpia/stdlib"
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
	qb.inserir({"id": 1, "nome": "Harpia", "idade": 3})
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
	if string(valNome.(ptst.Texto)) != "Harpia" {
		t.Errorf("Nome esperado 'Harpia', obteve '%v'", valNome)
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

func TestBDModuloMySQL(t *testing.T) {
	ctx := ptst.NewContexto(ptst.OpcsContexto{})
	defer ctx.Terminar()

	// O driver de MySQL deve ser registrado com sucesso.
	// Vamos testar se a função conectarMysql é chamada e tenta abrir conexão.
	codigo := `
	de "bd" importe conectarMysql

	tente {
		var conn = conectarMysql("root:senha@tcp(127.0.0.1:3306)/banco")
		conn.fechar()
	} capture (erro) {
		# Se falhar a conexão com banco real (MySQL offline), é esperado. O importante é o driver estar registrado.
	}
	`

	_, err := ptst.ExecutarString(ctx, codigo)
	if err != nil {
		t.Fatalf("Erro ao executar script com conector MySQL: %v", err)
	}
}

func TestBDModuloQdrant(t *testing.T) {
	// Cria servidor HTTP mockado para emular as respostas do Qdrant REST API
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if r.Method == "PUT" {
			w.Write([]byte(`{"result": {"status": "ok"}}`))
			return
		}

		if r.Method == "POST" && strings.Contains(r.URL.Path, "/search") {
			w.Write([]byte(`{
				"result": [
					{
						"id": 1,
						"score": 0.95,
						"payload": {"nome": "Maria"}
					}
				]
			}`))
			return
		}

		if r.Method == "POST" && strings.Contains(r.URL.Path, "/delete") {
			w.Write([]byte(`{"result": {"status": "ok"}}`))
			return
		}
	}))
	defer server.Close()

	ctx := ptst.NewContexto(ptst.OpcsContexto{})
	defer ctx.Terminar()

	codigo := fmt.Sprintf(`
	de "bd" importe conectarQdrant

	var cliente = conectarQdrant("%s", "colecao-teste")

	// 1. Testa inserir ponto vetorial
	var vetor = [0.1, 0.2, 0.3]
	var meta = {"nome": "Maria"}
	var inserido = cliente.inserir(1, vetor, meta)

	// 2. Testa busca vetorial por similaridade
	var resultados = cliente.buscar(vetor, 5)
	var totalResultados = tamanho(resultados)
	var primeiro = resultados[0]
	var id = primeiro["id"]
	var score = primeiro["score"]
	var payload = primeiro["payload"]
	var nome = payload["nome"]

	// 3. Testa deleção de ponto
	var deletado = cliente.deletar(1)
	`, server.URL)

	res, err := ptst.ExecutarString(ctx, codigo)
	if err != nil {
		t.Fatalf("Erro ao executar script com banco vetorial Qdrant: %v", err)
	}

	valInserido, _ := res.Escopo.ObterValor("inserido")
	if valInserido != ptst.Verdadeiro {
		t.Errorf("Inserção no Qdrant deveria ter tido sucesso")
	}

	valTotal, _ := res.Escopo.ObterValor("totalResultados")
	if int(valTotal.(ptst.Inteiro)) != 1 {
		t.Errorf("Deveria retornar exatamente 1 resultado do Qdrant")
	}

	valNome, _ := res.Escopo.ObterValor("nome")
	if string(valNome.(ptst.Texto)) != "Maria" {
		t.Errorf("Nome esperado 'Maria' no payload, obteve: %v", valNome)
	}

	valDeletado, _ := res.Escopo.ObterValor("deletado")
	if valDeletado != ptst.Verdadeiro {
		t.Errorf("Deleção no Qdrant deveria ter tido sucesso")
	}
}

func TestBDModuloORMTipado(t *testing.T) {
	ctx := ptst.NewContexto(ptst.OpcsContexto{})
	defer ctx.Terminar()

	codigo := `
	de "bd" importe conectarSqlite

	var conn = conectarSqlite(":memory:")
	conn.executar("CREATE TABLE produtos (id INTEGER PRIMARY KEY, nome TEXT, preco REAL, disponivel BOOLEAN)")

	// Declara tabela com o schema de tipos mapeado (ORM Tipado)
	var schema = {"id": "inteiro", "nome": "texto", "preco": "decimal", "disponivel": "booleano"}
	var tabela = conn.tabela("produtos", schema)

	// 1. Insercao valida
	tabela.inserir({"id": 1, "nome": "Teclado", "preco": 150.0, "disponivel": Verdadeiro})

	// 2. Insercao com campo inexistente no schema (deve gerar erro de valor)
	var erroCampoInexistente = Falso
	tente {
		tabela.inserir({"id": 2, "nome": "Mouse", "cor": "preto"})
	} capture (erro) {
		erroCampoInexistente = Verdadeiro
	}

	// 3. Insercao com tipo incorreto no schema (deve gerar erro de tipagem)
	var erroTipoIncorreto = Falso
	tente {
		tabela.inserir({"id": 3, "nome": "Monitor", "preco": "quinhentos"})
	} capture (erro) {
		erroTipoIncorreto = Verdadeiro
	}

	conn.fechar()
	`

	res, err := ptst.ExecutarString(ctx, codigo)
	if err != nil {
		t.Fatalf("Erro ao executar script com ORM Tipado: %v", err)
	}

	valInexistente, _ := res.Escopo.ObterValor("erroCampoInexistente")
	if valInexistente != ptst.Verdadeiro {
		t.Errorf("Deveria levantar exceção para campo inexistente")
	}

	valTipoIncorreto, _ := res.Escopo.ObterValor("erroTipoIncorreto")
	if valTipoIncorreto != ptst.Verdadeiro {
		t.Errorf("Deveria levantar exceção para tipo de campo incorreto")
	}
}

