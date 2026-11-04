package bd

import (
	"testing"

	"github.com/mat-dgruber/Harpia/hrp"
)

func TestConverterPlaceholdersPostgres(t *testing.T) {
	conexaoPostgres := &ConexaoSQL{
		driver: "postgres",
	}

	qb := &QueryBuilder{
		conexao: conexaoPostgres,
	}

	queryOriginal := "SELECT * FROM usuarios WHERE nome = ? AND idade > ? AND ativo = ?"
	queryEsperada := "SELECT * FROM usuarios WHERE nome = $1 AND idade > $2 AND ativo = $3"

	resultado := qb.converterPlaceholders(queryOriginal)
	if resultado != queryEsperada {
		t.Errorf("Erro ao converter placeholders para Postgres.\nEsperado: %s\nObtido:   %s", queryEsperada, resultado)
	}
}

func TestConverterPlaceholdersSqlite(t *testing.T) {
	conexaoSqlite := &ConexaoSQL{
		driver: "sqlite",
	}

	qb := &QueryBuilder{
		conexao: conexaoSqlite,
	}

	queryOriginal := "SELECT * FROM usuarios WHERE nome = ? AND idade > ? AND ativo = ?"
	queryEsperada := "SELECT * FROM usuarios WHERE nome = ? AND idade > ? AND ativo = ?"

	resultado := qb.converterPlaceholders(queryOriginal)
	if resultado != queryEsperada {
		t.Errorf("Erro ao converter placeholders para Sqlite (deveria manter '?').\nEsperado: %s\nObtido:   %s", queryEsperada, resultado)
	}
}

func TestConverterPlaceholdersMysql(t *testing.T) {
	conexaoMysql := &ConexaoSQL{
		driver: "mysql",
	}

	qb := &QueryBuilder{
		conexao: conexaoMysql,
	}

	queryOriginal := "SELECT * FROM usuarios WHERE nome = ? AND idade > ? AND ativo = ?"
	queryEsperada := "SELECT * FROM usuarios WHERE nome = ? AND idade > ? AND ativo = ?"

	resultado := qb.converterPlaceholders(queryOriginal)
	if resultado != queryEsperada {
		t.Errorf("Erro ao converter placeholders para Mysql (deveria manter '?').\nEsperado: %s\nObtido:   %s", queryEsperada, resultado)
	}
}

func TestQueryBuilderBloqueiaSQLInjectionEmColunas(t *testing.T) {
	qb := &QueryBuilder{tabela: "usuarios"}
	dadosMaliciosos := hrp.Mapa{
		"nome); DROP TABLE usuarios; --": hrp.Texto("admin"),
	}
	if _, err := qb.inserir(hrp.Tupla{dadosMaliciosos}); err == nil {
		t.Fatal("esperava erro de SQL injection na chave do mapa em inserir, mas passou")
	}
	if _, err := qb.atualizar(hrp.Tupla{dadosMaliciosos}); err == nil {
		t.Fatal("esperava erro de SQL injection na chave do mapa em atualizar, mas passou")
	}

	qbTabelaInvalida := &QueryBuilder{tabela: "usuarios; DROP TABLE usuarios; --"}
	dadosOk := hrp.Mapa{"nome": hrp.Texto("admin")}
	if _, err := qbTabelaInvalida.inserir(hrp.Tupla{dadosOk}); err == nil {
		t.Fatal("esperava erro ao inserir com tabela inválida")
	}
	if _, err := qbTabelaInvalida.atualizar(hrp.Tupla{dadosOk}); err == nil {
		t.Fatal("esperava erro ao atualizar com tabela inválida")
	}
	if _, err := qbTabelaInvalida.deletar(hrp.Tupla{}); err == nil {
		t.Fatal("esperava erro ao deletar com tabela inválida")
	}

	qbSchemaInvalido := &QueryBuilder{
		tabela: "usuarios",
		schema: hrp.Mapa{"coluna; DROP TABLE": hrp.Texto("texto")},
	}
	if _, err := qbSchemaInvalido.inserir(hrp.Tupla{dadosOk}); err == nil {
		t.Fatal("esperava erro ao inserir com schema inválido")
	}
	if _, err := qbSchemaInvalido.atualizar(hrp.Tupla{dadosOk}); err == nil {
		t.Fatal("esperava erro ao atualizar com schema inválido")
	}
	if _, err := qbSchemaInvalido.deletar(hrp.Tupla{}); err == nil {
		t.Fatal("esperava erro ao deletar com schema inválido")
	}
}

func TestQueryBuilderImutabilidade(t *testing.T) {
	base := &QueryBuilder{tabela: "usuarios"}

	res1, err := base.onde(hrp.Tupla{hrp.Texto("ativo"), hrp.Texto("="), hrp.Verdadeiro})
	if err != nil {
		t.Fatalf("erro ao executar onde: %v", err)
	}
	q1, ok := res1.(*QueryBuilder)
	if !ok {
		t.Fatalf("esperava *QueryBuilder em res1")
	}

	res2, err := base.onde(hrp.Tupla{hrp.Texto("idade"), hrp.Texto(">"), hrp.Inteiro(18)})
	if err != nil {
		t.Fatalf("erro ao executar onde: %v", err)
	}
	q2, ok := res2.(*QueryBuilder)
	if !ok {
		t.Fatalf("esperava *QueryBuilder em res2")
	}

	if len(base.condicoes) != 0 {
		t.Fatalf("base foi mutado! condicoes esperadas 0, obteve %d", len(base.condicoes))
	}
	if len(base.args) != 0 {
		t.Fatalf("base foi mutado! args esperados 0, obteve %d", len(base.args))
	}

	if len(q1.condicoes) != 1 || q1.condicoes[0] != "ativo = ?" {
		t.Fatalf("condicoes inesperadas em q1: %v", q1.condicoes)
	}
	if len(q2.condicoes) != 1 || q2.condicoes[0] != "idade > ?" {
		t.Fatalf("condicoes inesperadas em q2: %v", q2.condicoes)
	}

	resS1, err := base.selecionar(hrp.Tupla{hrp.Texto("id"), hrp.Texto("nome")})
	if err != nil {
		t.Fatalf("erro ao selecionar: %v", err)
	}
	s1 := resS1.(*QueryBuilder)

	resS2, err := base.selecionar(hrp.Tupla{hrp.Texto("email")})
	if err != nil {
		t.Fatalf("erro ao selecionar: %v", err)
	}
	s2 := resS2.(*QueryBuilder)

	if len(base.selecionados) != 0 {
		t.Fatalf("base foi mutado em selecionar! selecionados esperados 0, obteve %d", len(base.selecionados))
	}
	if len(s1.selecionados) != 2 {
		t.Fatalf("esperava 2 selecionados em s1, obteve %d", len(s1.selecionados))
	}
	if len(s2.selecionados) != 1 {
		t.Fatalf("esperava 1 selecionado em s2, obteve %d", len(s2.selecionados))
	}

	resL1, err := base.limite(hrp.Tupla{hrp.Inteiro(10)})
	if err != nil {
		t.Fatalf("erro ao definir limite: %v", err)
	}
	l1 := resL1.(*QueryBuilder)
	if base.limiteVal != 0 {
		t.Fatalf("base foi mutado em limite! esperado 0, obteve %d", base.limiteVal)
	}
	if l1.limiteVal != 10 {
		t.Fatalf("esperava 10 em l1.limiteVal, obteve %d", l1.limiteVal)
	}

	resOrd, err := base.ordenarPor(hrp.Tupla{hrp.Texto("nome"), hrp.Texto("ASC")})
	if err != nil {
		t.Fatalf("erro ao ordenarPor: %v", err)
	}
	o1 := resOrd.(*QueryBuilder)
	if len(base.ordenacoes) != 0 {
		t.Fatalf("base foi mutado em ordenarPor! esperado 0, obteve %d", len(base.ordenacoes))
	}
	if len(o1.ordenacoes) != 1 || o1.ordenacoes[0] != "nome ASC" {
		t.Fatalf("esperava 'nome ASC' em o1.ordenacoes, obteve %v", o1.ordenacoes)
	}
}
