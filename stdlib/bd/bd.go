// Package bd implementa os conectores, gerenciadores e adaptadores de acesso
// a bancos de dados relacionais (SQL), não-relacionais (NoSQL) e vetoriais do Harpia.
package bd

import (
	"database/sql"

	_ "github.com/glebarez/go-sqlite"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/mat-dgruber/Harpia/hrp"
)

// met_bd_conectarSqlite abre conexão com o banco SQLite em modo arquivo local de forma embarcada.
// Excelente para ambientes de desenvolvimento local ou armazenamento embarcado de baixo overhead.
func met_bd_conectarSqlite(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if err := hrp.VerificaNumeroArgumentos("conectarSqlite", false, args, 1, 1); err != nil {
		return nil, err
	}
	caminho, err := hrp.NewTexto(args[0])
	if err != nil {
		return nil, err
	}
	db, errOpen := sql.Open("sqlite", string(caminho.(hrp.Texto)))
	if errOpen != nil {
		return nil, hrp.NewErroF(hrp.ErroDeSistema, "Erro ao abrir banco SQLite: %v", errOpen)
	}
	return &ConexaoSQL{db: db, driver: "sqlite"}, nil
}

// met_bd_conectarPostgres abre conexão ativa TCP/IP com um cluster PostgreSQL externo.
// Suporta pooling nativo e strings de conexão no formato padrão DSN ou URL.
func met_bd_conectarPostgres(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if err := hrp.VerificaNumeroArgumentos("conectarPostgres", false, args, 1, 1); err != nil {
		return nil, err
	}
	urlConn, err := hrp.NewTexto(args[0])
	if err != nil {
		return nil, err
	}
	db, errOpen := sql.Open("postgres", string(urlConn.(hrp.Texto)))
	if errOpen != nil {
		return nil, hrp.NewErroF(hrp.ErroDeSistema, "Erro ao abrir banco PostgreSQL: %v", errOpen)
	}
	return &ConexaoSQL{db: db, driver: "postgres"}, nil
}

// met_bd_conectarMysql inicia uma sessão com servidores remotos ou locais MySQL utilizando DSN típico.
func met_bd_conectarMysql(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if err := hrp.VerificaNumeroArgumentos("conectarMysql", false, args, 1, 1); err != nil {
		return nil, err
	}
	urlConn, err := hrp.NewTexto(args[0])
	if err != nil {
		return nil, err
	}
	db, errOpen := sql.Open("mysql", string(urlConn.(hrp.Texto)))
	if errOpen != nil {
		return nil, hrp.NewErroF(hrp.ErroDeSistema, "Erro ao abrir banco MySQL: %v", errOpen)
	}
	return &ConexaoSQL{db: db, driver: "mysql"}, nil
}

// met_bd_conectarMongo abre um pool de conexões com bancos de dados de documentos MongoDB.
func met_bd_conectarMongo(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if err := hrp.VerificaNumeroArgumentos("conectarMongo", false, args, 1, 1); err != nil {
		return nil, err
	}
	urlConn, err := hrp.NewTexto(args[0])
	if err != nil {
		return nil, err
	}
	conn, errConn := conectarMongoImpl(string(urlConn.(hrp.Texto)))
	if errConn != nil {
		return nil, hrp.NewErroF(hrp.ErroDeSistema, "Erro ao conectar ao MongoDB: %v", errConn)
	}
	return conn, nil
}

// met_bd_conectarRedis inicializa uma conexão com servidores Redis de armazenamento e cache chave-valor ultra-rápido.
func met_bd_conectarRedis(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if err := hrp.VerificaNumeroArgumentos("conectarRedis", false, args, 1, 1); err != nil {
		return nil, err
	}
	urlConn, err := hrp.NewTexto(args[0])
	if err != nil {
		return nil, err
	}
	conn, errConn := conectarRedisImpl(string(urlConn.(hrp.Texto)))
	if errConn != nil {
		return nil, hrp.NewErroF(hrp.ErroDeSistema, "Erro ao conectar ao Redis: %v", errConn)
	}
	return conn, nil
}

func init() {
	// Registra o módulo integrado 'bd' (Banco de Dados) e todos os seus subtipos internos.
	hrp.RegistraModuloImpl(&hrp.ModuloImpl{
		Info: hrp.ModuloInfo{
			Nome:    "bd",
			Arquivo: "stdlib/bd",
		},
		Constantes: hrp.Mapa{
			"ConexaoSQL":      TipoConexaoSQL,
			"QueryBuilder":    TipoQueryBuilder,
			"ConexaoMongo":    TipoConexaoMongo,
			"ConexaoRedis":    TipoConexaoRedis,
			"ClienteVetorial": TipoClienteVetorial,
		},
		Metodos: []*hrp.Metodo{
			hrp.NewMetodoOuPanic("conectarSqlite", met_bd_conectarSqlite, "Inicia conexão local com arquivo de banco de dados SQLite."),
			hrp.NewMetodoOuPanic("conectarPostgres", met_bd_conectarPostgres, "Abre sessão de conexão contra servidores PostgreSQL."),
			hrp.NewMetodoOuPanic("conectarMysql", met_bd_conectarMysql, "Abre sessão de conexão contra servidores MySQL."),
			hrp.NewMetodoOuPanic("conectarMongo", met_bd_conectarMongo, "Estabelece conexão com o banco de dados orientado a documentos MongoDB."),
			hrp.NewMetodoOuPanic("conectarRedis", met_bd_conectarRedis, "Conecta com o banco/cache chave-valor na memória Redis."),
			hrp.NewMetodoOuPanic("conectarQdrant", met_conectar_qdrant, "Conecta ao banco vetorial Qdrant(url, colecao) para busca semântica por proximidade."),
		},
	})
}
