// Package bd implementa os conectores, gerenciadores e adaptadores de acesso
// a bancos de dados relacionais (SQL), não-relacionais (NoSQL) e vetoriais do Harpia.
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

// ConexaoMongo gerencia e encapsula o cliente de conexão físico do driver oficial MongoDB.
type ConexaoMongo struct {
	client *mongo.Client
	dbName string
}

// TipoConexaoMongo define o tipo estrutural representativo de ConexaoMongo na VM do Harpia.
var TipoConexaoMongo = hrp.NewTipo("ConexaoMongo", "Conexão de Banco de Dados MongoDB")

// Tipo retorna a definição de classe na VM do Harpia.
func (m *ConexaoMongo) Tipo() *hrp.Tipo {
	return TipoConexaoMongo
}

// M__obtem_attributo__ expõe dinamicamente os métodos colecao() e fechar() no escopo dos objetos MongoDB do Harpia.
func (m *ConexaoMongo) M__obtem_attributo__(nome string) (hrp.Objeto, error) {
	switch nome {
	case "colecao":
		// Retorna um método vinculável que permite interagir com uma coleção específica (Collection) no MongoDB.
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
		}, "Abre ou cria uma coleção lógica no banco de dados MongoDB."), nil

	case "fechar":
		// Fecha síncronamente e de forma limpa o pool de conexões ativas com o cluster MongoDB.
		return hrp.NewMetodoOuPanic("fechar", func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			m.client.Disconnect(ctx)
			return hrp.Nulo, nil
		}, "Encerra a conexão e libera os recursos do MongoDB."), nil
	}

	return nil, hrp.NewErroF(hrp.AtributoErro, "Atributo '%s' não existe em ConexaoMongo", nome)
}

// ColecaoMongo encapsula a referência direta do driver oficial para interação de operações CRUD no MongoDB.
type ColecaoMongo struct {
	col *mongo.Collection
}

// TipoColecaoMongo define o tipo da classe na VM para ColecaoMongo.
var TipoColecaoMongo = hrp.NewTipo("ColecaoMongo", "Coleção do MongoDB")

// Tipo retorna a representação na VM.
func (c *ColecaoMongo) Tipo() *hrp.Tipo {
	return TipoColecaoMongo
}

// M__obtem_attributo__ expõe os métodos de inserção e busca estruturada de documentos no MongoDB.
func (c *ColecaoMongo) M__obtem_attributo__(nome string) (hrp.Objeto, error) {
	switch nome {
	case "inserirUm":
		// Insere um único documento representado por um Mapa do Harpia na coleção física do MongoDB.
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
		}, "Insere um registro estruturado (Mapa) no MongoDB."), nil

	case "buscarUm":
		// Executa uma consulta baseada em um filtro estruturado (Mapa) e retorna o primeiro documento correspondente.
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
		}, "Filtra e encontra um registro específico no MongoDB utilizando mapeamento de campos."), nil
	}

	return nil, hrp.NewErroF(hrp.AtributoErro, "Atributo '%s' não existe em ColecaoMongo", nome)
}

// ConexaoRedis modela o adaptador oficial em Go para sessões de conexão ativa no Redis.
type ConexaoRedis struct {
	client *redis.Client
}

// TipoConexaoRedis mapeia e registra a classe ConexaoRedis na VM do Harpia.
var TipoConexaoRedis = hrp.NewTipo("ConexaoRedis", "Conexão de Banco de Dados Redis")

// Tipo retorna a representação para a máquina virtual.
func (r *ConexaoRedis) Tipo() *hrp.Tipo {
	return TipoConexaoRedis
}

// M__obtem_attributo__ disponibiliza comandos do Redis para manipulação de cache chave-valor estruturado.
func (r *ConexaoRedis) M__obtem_attributo__(nome string) (hrp.Objeto, error) {
	switch nome {
	case "definir":
		// Define o valor de uma chave textual, opcionalmente associando um TTL de expiração em segundos.
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
		}, "Associa um valor textual a uma chave única com TTL opcional."), nil

	case "obter":
		// Resgata o valor associado a uma chave cadastrada no Redis. Devolve nulo caso a chave não exista.
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
		}, "Resgata os dados gravados sob a chave especificada."), nil

	case "remover":
		// Remove fisicamente a chave informada do armazenamento em memória do Redis.
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
		}, "Deleta permanentemente a chave mapeada e limpa a memória."), nil

	case "fechar":
		// Fecha de forma limpa a conexão e libera os recursos do pool Redis.
		return hrp.NewMetodoOuPanic("fechar", func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
			r.client.Close()
			return hrp.Nulo, nil
		}, "Fecha e encerra as conexões ativas com o servidor Redis."), nil
	}

	return nil, hrp.NewErroF(hrp.AtributoErro, "Atributo '%s' não existe em ConexaoRedis", nome)
}

// conectarMongoImpl cria uma instância configurada do cliente MongoDB.
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

// conectarRedisImpl analisa e configura a conexão contra instâncias Redis a partir de uma URI padrão.
func conectarRedisImpl(url string) (*ConexaoRedis, error) {
	opt, err := redis.ParseURL(url)
	if err != nil {
		return nil, err
	}
	client := redis.NewClient(opt)
	return &ConexaoRedis{client: client}, nil
}
