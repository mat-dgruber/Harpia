package bd

import (
	"database/sql"

	_ "github.com/glebarez/go-sqlite"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/mat-dgruber/Harpia/ptst"
)

// met_bd_conectarSqlite abre conexao SQLite pura em Go
func met_bd_conectarSqlite(inst ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
	if err := ptst.VerificaNumeroArgumentos("conectarSqlite", false, args, 1, 1); err != nil {
		return nil, err
	}
	caminho, err := ptst.NewTexto(args[0])
	if err != nil {
		return nil, err
	}
	db, errOpen := sql.Open("sqlite", string(caminho.(ptst.Texto)))
	if errOpen != nil {
		return nil, ptst.NewErroF(ptst.ErroDeSistema, "Erro ao abrir banco SQLite: %v", errOpen)
	}
	return &ConexaoSQL{db: db, driver: "sqlite"}, nil
}

// met_bd_conectarPostgres abre conexao PostgreSQL
func met_bd_conectarPostgres(inst ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
	if err := ptst.VerificaNumeroArgumentos("conectarPostgres", false, args, 1, 1); err != nil {
		return nil, err
	}
	urlConn, err := ptst.NewTexto(args[0])
	if err != nil {
		return nil, err
	}
	db, errOpen := sql.Open("postgres", string(urlConn.(ptst.Texto)))
	if errOpen != nil {
		return nil, ptst.NewErroF(ptst.ErroDeSistema, "Erro ao abrir banco PostgreSQL: %v", errOpen)
	}
	return &ConexaoSQL{db: db, driver: "postgres"}, nil
}

// met_bd_conectarMysql abre conexao MySQL
func met_bd_conectarMysql(inst ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
	if err := ptst.VerificaNumeroArgumentos("conectarMysql", false, args, 1, 1); err != nil {
		return nil, err
	}
	urlConn, err := ptst.NewTexto(args[0])
	if err != nil {
		return nil, err
	}
	db, errOpen := sql.Open("mysql", string(urlConn.(ptst.Texto)))
	if errOpen != nil {
		return nil, ptst.NewErroF(ptst.ErroDeSistema, "Erro ao abrir banco MySQL: %v", errOpen)
	}
	return &ConexaoSQL{db: db, driver: "mysql"}, nil
}

// met_bd_conectarMongo abre conexao MongoDB
func met_bd_conectarMongo(inst ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
	if err := ptst.VerificaNumeroArgumentos("conectarMongo", false, args, 1, 1); err != nil {
		return nil, err
	}
	urlConn, err := ptst.NewTexto(args[0])
	if err != nil {
		return nil, err
	}
	conn, errConn := conectarMongoImpl(string(urlConn.(ptst.Texto)))
	if errConn != nil {
		return nil, ptst.NewErroF(ptst.ErroDeSistema, "Erro ao conectar ao MongoDB: %v", errConn)
	}
	return conn, nil
}

// met_bd_conectarRedis abre conexao Redis
func met_bd_conectarRedis(inst ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
	if err := ptst.VerificaNumeroArgumentos("conectarRedis", false, args, 1, 1); err != nil {
		return nil, err
	}
	urlConn, err := ptst.NewTexto(args[0])
	if err != nil {
		return nil, err
	}
	conn, errConn := conectarRedisImpl(string(urlConn.(ptst.Texto)))
	if errConn != nil {
		return nil, ptst.NewErroF(ptst.ErroDeSistema, "Erro ao conectar ao Redis: %v", errConn)
	}
	return conn, nil
}

func init() {
	ptst.RegistraModuloImpl(&ptst.ModuloImpl{
		Info: ptst.ModuloInfo{
			Nome:    "bd",
			Arquivo: "stdlib/bd",
		},
		Constantes: ptst.Mapa{
			"ConexaoSQL":      TipoConexaoSQL,
			"QueryBuilder":    TipoQueryBuilder,
			"ConexaoMongo":    TipoConexaoMongo,
			"ConexaoRedis":    TipoConexaoRedis,
			"ClienteVetorial": TipoClienteVetorial,
		},
		Metodos: []*ptst.Metodo{
			ptst.NewMetodoOuPanic("conectarSqlite", met_bd_conectarSqlite, ""),
			ptst.NewMetodoOuPanic("conectarPostgres", met_bd_conectarPostgres, ""),
			ptst.NewMetodoOuPanic("conectarMysql", met_bd_conectarMysql, ""),
			ptst.NewMetodoOuPanic("conectarMongo", met_bd_conectarMongo, ""),
			ptst.NewMetodoOuPanic("conectarRedis", met_bd_conectarRedis, ""),
			ptst.NewMetodoOuPanic("conectarQdrant", met_conectar_qdrant, "Conecta ao banco vetorial Qdrant(url, colecao)"),
		},
	})
}
