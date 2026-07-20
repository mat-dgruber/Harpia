package bd

import (
	"testing"
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
