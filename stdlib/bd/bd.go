package bd

import (
	"database/sql"

	_ "github.com/glebarez/go-sqlite"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/mat-dgruber/Harpia/hrp"
)

// met_bd_conectarSqlite abre conexao SQLite pura em Go
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

// met_bd_conectarPostgres abre conexao PostgreSQL
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

// met_bd_conectarMysql abre conexao MySQL
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

// met_bd_conectarMongo abre conexao MongoDB
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

// met_bd_conectarRedis abre conexao Redis
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
			hrp.NewMetodoOuPanic("conectarSqlite", met_bd_conectarSqlite, ""),
			hrp.NewMetodoOuPanic("conectarPostgres", met_bd_conectarPostgres, ""),
			hrp.NewMetodoOuPanic("conectarMysql", met_bd_conectarMysql, ""),
			hrp.NewMetodoOuPanic("conectarMongo", met_bd_conectarMongo, ""),
			hrp.NewMetodoOuPanic("conectarRedis", met_bd_conectarRedis, ""),
			hrp.NewMetodoOuPanic("conectarQdrant", met_conectar_qdrant, "Conecta ao banco vetorial Qdrant(url, colecao)"),
		},
	})
}
