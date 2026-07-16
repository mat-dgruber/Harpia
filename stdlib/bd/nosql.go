package bd

import (
	"context"
	"time"

	"github.com/natanfeitosa/portuscript/ptst"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ConexaoMongo struct {
	client *mongo.Client
	dbName string
}

var TipoConexaoMongo = ptst.NewTipo("ConexaoMongo", "Conexão de Banco de Dados MongoDB")

func (m *ConexaoMongo) Tipo() *ptst.Tipo {
	return TipoConexaoMongo
}

func (m *ConexaoMongo) M__obtem_attributo__(nome string) (ptst.Objeto, error) {
	switch nome {
	case "colecao":
		return ptst.NewMetodoOuPanic("colecao", func(inst ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
			if err := ptst.VerificaNumeroArgumentos("colecao", false, args, 1, 1); err != nil {
				return nil, err
			}
			colName, err := ptst.NewTexto(args[0])
			if err != nil {
				return nil, err
			}
			return &ColecaoMongo{
				col: m.client.Database(m.dbName).Collection(string(colName.(ptst.Texto))),
			}, nil
		}, ""), nil

	case "fechar":
		return ptst.NewMetodoOuPanic("fechar", func(inst ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			m.client.Disconnect(ctx)
			return ptst.Nulo, nil
		}, ""), nil
	}

	return nil, ptst.NewErroF(ptst.AtributoErro, "Atributo '%s' não existe em ConexaoMongo", nome)
}

type ColecaoMongo struct {
	col *mongo.Collection
}

var TipoColecaoMongo = ptst.NewTipo("ColecaoMongo", "Coleção do MongoDB")

func (c *ColecaoMongo) Tipo() *ptst.Tipo {
	return TipoColecaoMongo
}

func (c *ColecaoMongo) M__obtem_attributo__(nome string) (ptst.Objeto, error) {
	switch nome {
	case "inserirUm":
		return ptst.NewMetodoOuPanic("inserirUm", func(inst ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
			if err := ptst.VerificaNumeroArgumentos("inserirUm", false, args, 1, 1); err != nil {
				return nil, err
			}
			mapa, ok := args[0].(ptst.Mapa)
			if !ok {
				return nil, ptst.NewErroF(ptst.TipagemErro, "inserirUm esperava um Mapa")
			}
			doc := make(map[string]interface{})
			for k, v := range mapa {
				doc[k] = toGoType(v)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_, err := c.col.InsertOne(ctx, doc)
			if err != nil {
				return nil, ptst.NewErroF(ptst.ErroDeSistema, "Erro ao inserir documento no MongoDB: %v", err)
			}
			return ptst.Nulo, nil
		}, ""), nil

	case "buscarUm":
		return ptst.NewMetodoOuPanic("buscarUm", func(inst ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
			if err := ptst.VerificaNumeroArgumentos("buscarUm", false, args, 1, 1); err != nil {
				return nil, err
			}
			filtroMapa, ok := args[0].(ptst.Mapa)
			if !ok {
				return nil, ptst.NewErroF(ptst.TipagemErro, "buscarUm esperava um Mapa como filtro")
			}
			filtro := make(map[string]interface{})
			for k, v := range filtroMapa {
				filtro[k] = toGoType(v)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			var doc bson.M
			err := c.col.FindOne(ctx, filtro).Decode(&doc)
			if err == mongo.ErrNoDocuments {
				return ptst.Nulo, nil
			}
			if err != nil {
				return nil, ptst.NewErroF(ptst.ErroDeSistema, "Erro ao buscar documento no MongoDB: %v", err)
			}
			res := ptst.NewMapaVazio()
			for k, v := range doc {
				res.M__define_item__(ptst.Texto(k), toPtObject(v))
			}
			return res, nil
		}, ""), nil
	}

	return nil, ptst.NewErroF(ptst.AtributoErro, "Atributo '%s' não existe em ColecaoMongo", nome)
}

type ConexaoRedis struct {
	client *redis.Client
}

var TipoConexaoRedis = ptst.NewTipo("ConexaoRedis", "Conexão de Banco de Dados Redis")

func (r *ConexaoRedis) Tipo() *ptst.Tipo {
	return TipoConexaoRedis
}

func (r *ConexaoRedis) M__obtem_attributo__(nome string) (ptst.Objeto, error) {
	switch nome {
	case "definir":
		return ptst.NewMetodoOuPanic("definir", func(inst ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
			if err := ptst.VerificaNumeroArgumentos("definir", false, args, 2, 3); err != nil {
				return nil, err
			}
			chave, err := ptst.NewTexto(args[0])
			if err != nil {
				return nil, err
			}
			valor, err := ptst.NewTexto(args[1])
			if err != nil {
				return nil, err
			}
			var expiracao time.Duration
			if len(args) == 3 {
				expSegundos, err := ptst.NewInteiro(args[2])
				if err != nil {
					return nil, err
				}
				expiracao = time.Duration(expSegundos.(ptst.Inteiro)) * time.Second
			}
			ctx := context.Background()
			errCmd := r.client.Set(ctx, string(chave.(ptst.Texto)), string(valor.(ptst.Texto)), expiracao).Err()
			if errCmd != nil {
				return nil, ptst.NewErroF(ptst.ErroDeSistema, "Erro no Redis: %v", errCmd)
			}
			return ptst.Nulo, nil
		}, ""), nil

	case "obter":
		return ptst.NewMetodoOuPanic("obter", func(inst ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
			if err := ptst.VerificaNumeroArgumentos("obter", false, args, 1, 1); err != nil {
				return nil, err
			}
			chave, err := ptst.NewTexto(args[0])
			if err != nil {
				return nil, err
			}
			ctx := context.Background()
			val, errCmd := r.client.Get(ctx, string(chave.(ptst.Texto))).Result()
			if errCmd == redis.Nil {
				return ptst.Nulo, nil
			}
			if errCmd != nil {
				return nil, ptst.NewErroF(ptst.ErroDeSistema, "Erro no Redis: %v", errCmd)
			}
			return ptst.Texto(val), nil
		}, ""), nil

	case "remover":
		return ptst.NewMetodoOuPanic("remover", func(inst ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
			if err := ptst.VerificaNumeroArgumentos("remover", false, args, 1, 1); err != nil {
				return nil, err
			}
			chave, err := ptst.NewTexto(args[0])
			if err != nil {
				return nil, err
			}
			ctx := context.Background()
			errCmd := r.client.Del(ctx, string(chave.(ptst.Texto))).Err()
			if errCmd != nil {
				return nil, ptst.NewErroF(ptst.ErroDeSistema, "Erro no Redis: %v", errCmd)
			}
			return ptst.Nulo, nil
		}, ""), nil

	case "fechar":
		return ptst.NewMetodoOuPanic("fechar", func(inst ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
			r.client.Close()
			return ptst.Nulo, nil
		}, ""), nil
	}

	return nil, ptst.NewErroF(ptst.AtributoErro, "Atributo '%s' não existe em ConexaoRedis", nome)
}

func conectarMongoImpl(url string) (*ConexaoMongo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(url))
	if err != nil {
		return nil, err
	}
	dbName := "portuscript"
	return &ConexaoMongo{client: client, dbName: dbName}, nil
}

func conectarRedisImpl(url string) (*ConexaoRedis, error) {
	opt, err := redis.ParseURL(url)
	if err != nil {
		return nil, err
	}
	client := redis.NewClient(opt)
	return &ConexaoRedis{client: client}, nil
}
