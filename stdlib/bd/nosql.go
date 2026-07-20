package bd

import (
	"context"
	"time"

	"github.com/mat-dgruber/Harpia/hrp"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ConexaoMongo struct {
	client *mongo.Client
	dbName string
}

var TipoConexaoMongo = hrp.NewTipo("ConexaoMongo", "Conexão de Banco de Dados MongoDB")

func (m *ConexaoMongo) Tipo() *hrp.Tipo {
	return TipoConexaoMongo
}

func (m *ConexaoMongo) M__obtem_attributo__(nome string) (hrp.Objeto, error) {
	switch nome {
	case "colecao":
		return hrp.NewMetodoOuPanic("colecao", func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
			if err := hrp.VerificaNumeroArgumentos("colecao", false, args, 1, 1); err != nil {
				return nil, err
			}
			colName, err := hrp.NewTexto(args[0])
			if err != nil {
				return nil, err
			}
			return &ColecaoMongo{
				col: m.client.Database(m.dbName).Collection(string(colName.(hrp.Texto))),
			}, nil
		}, ""), nil

	case "fechar":
		return hrp.NewMetodoOuPanic("fechar", func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			m.client.Disconnect(ctx)
			return hrp.Nulo, nil
		}, ""), nil
	}

	return nil, hrp.NewErroF(hrp.AtributoErro, "Atributo '%s' não existe em ConexaoMongo", nome)
}

type ColecaoMongo struct {
	col *mongo.Collection
}

var TipoColecaoMongo = hrp.NewTipo("ColecaoMongo", "Coleção do MongoDB")

func (c *ColecaoMongo) Tipo() *hrp.Tipo {
	return TipoColecaoMongo
}

func (c *ColecaoMongo) M__obtem_attributo__(nome string) (hrp.Objeto, error) {
	switch nome {
	case "inserirUm":
		return hrp.NewMetodoOuPanic("inserirUm", func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
			if err := hrp.VerificaNumeroArgumentos("inserirUm", false, args, 1, 1); err != nil {
				return nil, err
			}
			mapa, ok := args[0].(hrp.Mapa)
			if !ok {
				return nil, hrp.NewErroF(hrp.TipagemErro, "inserirUm esperava um Mapa")
			}
			doc := make(map[string]interface{})
			for k, v := range mapa {
				doc[k] = toGoType(v)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_, err := c.col.InsertOne(ctx, doc)
			if err != nil {
				return nil, hrp.NewErroF(hrp.ErroDeSistema, "Erro ao inserir documento no MongoDB: %v", err)
			}
			return hrp.Nulo, nil
		}, ""), nil

	case "buscarUm":
		return hrp.NewMetodoOuPanic("buscarUm", func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
			if err := hrp.VerificaNumeroArgumentos("buscarUm", false, args, 1, 1); err != nil {
				return nil, err
			}
			filtroMapa, ok := args[0].(hrp.Mapa)
			if !ok {
				return nil, hrp.NewErroF(hrp.TipagemErro, "buscarUm esperava um Mapa como filtro")
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
				return hrp.Nulo, nil
			}
			if err != nil {
				return nil, hrp.NewErroF(hrp.ErroDeSistema, "Erro ao buscar documento no MongoDB: %v", err)
			}
			res := hrp.NewMapaVazio()
			for k, v := range doc {
				res.M__define_item__(hrp.Texto(k), toPtObject(v))
			}
			return res, nil
		}, ""), nil
	}

	return nil, hrp.NewErroF(hrp.AtributoErro, "Atributo '%s' não existe em ColecaoMongo", nome)
}

type ConexaoRedis struct {
	client *redis.Client
}

var TipoConexaoRedis = hrp.NewTipo("ConexaoRedis", "Conexão de Banco de Dados Redis")

func (r *ConexaoRedis) Tipo() *hrp.Tipo {
	return TipoConexaoRedis
}

func (r *ConexaoRedis) M__obtem_attributo__(nome string) (hrp.Objeto, error) {
	switch nome {
	case "definir":
		return hrp.NewMetodoOuPanic("definir", func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
			if err := hrp.VerificaNumeroArgumentos("definir", false, args, 2, 3); err != nil {
				return nil, err
			}
			chave, err := hrp.NewTexto(args[0])
			if err != nil {
				return nil, err
			}
			valor, err := hrp.NewTexto(args[1])
			if err != nil {
				return nil, err
			}
			var expiracao time.Duration
			if len(args) == 3 {
				expSegundos, err := hrp.NewInteiro(args[2])
				if err != nil {
					return nil, err
				}
				expiracao = time.Duration(expSegundos.(hrp.Inteiro)) * time.Second
			}
			ctx := context.Background()
			errCmd := r.client.Set(ctx, string(chave.(hrp.Texto)), string(valor.(hrp.Texto)), expiracao).Err()
			if errCmd != nil {
				return nil, hrp.NewErroF(hrp.ErroDeSistema, "Erro no Redis: %v", errCmd)
			}
			return hrp.Nulo, nil
		}, ""), nil

	case "obter":
		return hrp.NewMetodoOuPanic("obter", func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
			if err := hrp.VerificaNumeroArgumentos("obter", false, args, 1, 1); err != nil {
				return nil, err
			}
			chave, err := hrp.NewTexto(args[0])
			if err != nil {
				return nil, err
			}
			ctx := context.Background()
			val, errCmd := r.client.Get(ctx, string(chave.(hrp.Texto))).Result()
			if errCmd == redis.Nil {
				return hrp.Nulo, nil
			}
			if errCmd != nil {
				return nil, hrp.NewErroF(hrp.ErroDeSistema, "Erro no Redis: %v", errCmd)
			}
			return hrp.Texto(val), nil
		}, ""), nil

	case "remover":
		return hrp.NewMetodoOuPanic("remover", func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
			if err := hrp.VerificaNumeroArgumentos("remover", false, args, 1, 1); err != nil {
				return nil, err
			}
			chave, err := hrp.NewTexto(args[0])
			if err != nil {
				return nil, err
			}
			ctx := context.Background()
			errCmd := r.client.Del(ctx, string(chave.(hrp.Texto))).Err()
			if errCmd != nil {
				return nil, hrp.NewErroF(hrp.ErroDeSistema, "Erro no Redis: %v", errCmd)
			}
			return hrp.Nulo, nil
		}, ""), nil

	case "fechar":
		return hrp.NewMetodoOuPanic("fechar", func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
			r.client.Close()
			return hrp.Nulo, nil
		}, ""), nil
	}

	return nil, hrp.NewErroF(hrp.AtributoErro, "Atributo '%s' não existe em ConexaoRedis", nome)
}

func conectarMongoImpl(url string) (*ConexaoMongo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(url))
	if err != nil {
		return nil, err
	}
	dbName := "Harpia"
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
