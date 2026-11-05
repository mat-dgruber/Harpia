// Package http implementa o servidor web nativo de alta performance e cliente HTTP do Harpia,
// com suporte a middlewares, roteamento inteligente de parâmetros (:id), desserialização automática
// de JSON, validação de segurança (OWASP) e tratamento de CORS out-of-the-box.
package http

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/mat-dgruber/Harpia/hrp"
)

// Requisicao representa os metadados e corpo de uma requisição HTTP recebida pelo servidor.
type Requisicao struct {
	Metodo     hrp.Texto
	Caminho    hrp.Texto
	Cabecalho  hrp.Mapa
	Corpo      hrp.Texto
	Parametros hrp.Mapa
	Query      hrp.Mapa   // Parâmetros de consulta (query string) parsed
	CorpoJson  hrp.Objeto // Body deserializado para Mapas/Listas do Harpia de forma automática
}

// TipoRequisicao mapeia e cria a classe Requisicao na VM.
var TipoRequisicao = hrp.NewTipo("Requisicao", "Objeto que representa uma requisição HTTP")

// Tipo retorna a representação correspondente na VM.
func (r *Requisicao) Tipo() *hrp.Tipo {
	return TipoRequisicao
}

// M__obtem_attributo__ expõe propriedades da requisição no ambiente do Harpia de forma somente leitura.
func (r *Requisicao) M__obtem_attributo__(nome string) (hrp.Objeto, error) {
	switch nome {
	case "metodo":
		return r.Metodo, nil
	case "caminho":
		return r.Caminho, nil
	case "cabecalho":
		return r.Cabecalho, nil
	case "corpo":
		return r.Corpo, nil
	case "parametros":
		return r.Parametros, nil
	case "query", "consulta":
		return r.Query, nil
	case "corpoJson", "corpo_json":
		if r.CorpoJson != nil {
			return r.CorpoJson, nil
		}
		return hrp.Nulo, nil
	}
	return nil, hrp.NewErroF(hrp.AtributoErro, "Atributo '%s' não existe em Requisicao", nome)
}

// Resposta representa o construtor do payload e cabeçalhos a ser enviado ao cliente HTTP.
type Resposta struct {
	Status    hrp.Inteiro
	Cabecalho hrp.Mapa
	Corpo     hrp.Texto
}

// TipoResposta mapeia e cria a classe Resposta na VM.
var TipoResposta = hrp.NewTipo("Resposta", "Objeto que representa uma resposta HTTP")

// Tipo retorna o tipo associado na VM.
func (r *Resposta) Tipo() *hrp.Tipo {
	return TipoResposta
}

// M__define_atributo__ possibilita alterar o status e corpo da resposta diretamente por atribuição simples.
func (r *Resposta) M__define_atributo__(nome string, valor hrp.Objeto) error {
	switch nome {
	case "status":
		statusInt, err := hrp.NewInteiro(valor)
		if err != nil {
			return err
		}
		r.Status = statusInt.(hrp.Inteiro)
		return nil
	case "corpo":
		corpoTexto, err := hrp.NewTexto(valor)
		if err != nil {
			return err
		}
		r.Corpo = corpoTexto.(hrp.Texto)
		return nil
	}
	return hrp.NewErroF(hrp.AtributoErro, "Atributo '%s' não pode ser modificado em Resposta", nome)
}

// M__obtem_attributo__ disponibiliza os métodos e propriedades fluídos de escrita de resposta HTTP.
func (r *Resposta) M__obtem_attributo__(nome string) (hrp.Objeto, error) {
	switch nome {
	case "status":
		return r.Status, nil
	case "corpo":
		return r.Corpo, nil
	case "cabecalho":
		return r.Cabecalho, nil
	case "definir_cabecalho":
		// Permite customizar cabeçalhos de resposta individuais.
		return hrp.NewMetodoOuPanic("definir_cabecalho", func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
			if err := hrp.VerificaNumeroArgumentos("definir_cabecalho", false, args, 2, 2); err != nil {
				return nil, err
			}
			k, err := hrp.NewTexto(args[0])
			if err != nil {
				return nil, err
			}
			v, err := hrp.NewTexto(args[1])
			if err != nil {
				return nil, err
			}
			r.Cabecalho.M__define_item__(k, v)
			return hrp.Nulo, nil
		}, "Define um par de chave/valor no cabeçalho HTTP de resposta."), nil
	case "enviarJson", "enviar_json":
		// Serializa o objeto fornecido e altera automaticamente o cabeçalho 'Content-Type' para 'application/json'.
		return hrp.NewMetodoOuPanic("enviar_json", func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
			if err := hrp.VerificaNumeroArgumentos("enviar_json", false, args, 1, 1); err != nil {
				return nil, err
			}
			goObj := hrp.ConverteParaGo(args[0])
			bytes, err := json.Marshal(goObj)
			if err != nil {
				return nil, hrp.NewErroF(hrp.ValorErro, "Erro ao serializar objeto para JSON: %v", err)
			}
			r.Cabecalho.M__define_item__(hrp.Texto("Content-Type"), hrp.Texto("application/json"))
			r.Corpo = hrp.Texto(bytes)
			return hrp.Nulo, nil
		}, "Serializa e envia o objeto como JSON, definindo Content-Type automaticamente"), nil

	case "definirStatus", "definir_status":
		// Define o código de resposta HTTP (ex: 201, 404).
		return hrp.NewMetodoOuPanic("definir_status", func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
			if err := hrp.VerificaNumeroArgumentos("definir_status", false, args, 1, 1); err != nil {
				return nil, err
			}
			statusInt, err := hrp.NewInteiro(args[0])
			if err != nil {
				return nil, err
			}
			r.Status = statusInt.(hrp.Inteiro)
			return hrp.Nulo, nil
		}, "Define o código de status HTTP da resposta (ex: res.definir_status(404))"), nil

	case "erroJson", "erro_json":
		// Envia um erro estruturado formatado in JSON.
		return hrp.NewMetodoOuPanic("erro_json", func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
			if err := hrp.VerificaNumeroArgumentos("erro_json", false, args, 2, 2); err != nil {
				return nil, err
			}
			statusInt, err := hrp.NewInteiro(args[0])
			if err != nil {
				return nil, err
			}
			r.Status = statusInt.(hrp.Inteiro)
			goObj := map[string]any{"erro": hrp.ConverteParaGo(args[1])}
			bytes, jsonErr := json.Marshal(goObj)
			if jsonErr != nil {
				return nil, hrp.NewErroF(hrp.ValorErro, "Erro ao serializar mensagem de erro: %v", jsonErr)
			}
			r.Cabecalho.M__define_item__(hrp.Texto("Content-Type"), hrp.Texto("application/json"))
			r.Corpo = hrp.Texto(bytes)
			return hrp.Nulo, nil
		}, "Envia uma resposta de erro JSON com status e mensagem (ex: res.erro_json(404, \"Não encontrado\"))"), nil

	case "enviar_midia", "enviar_imagem":
		// Envia um payload de mídia binária, definindo Content-Type e Cache-Control adequados.
		return hrp.NewMetodoOuPanic("enviar_midia", func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
			if err := hrp.VerificaNumeroArgumentos("enviar_midia", false, args, 1, 2); err != nil {
				return nil, err
			}
			conteudo, err := hrp.NewTexto(args[0])
			if err != nil {
				return nil, err
			}
			tipoMime := hrp.Texto("image/jpeg")
			if len(args) == 2 {
				m, err := hrp.NewTexto(args[1])
				if err != nil {
					return nil, err
				}
				tipoMime = m.(hrp.Texto)
			}
			r.Cabecalho.M__define_item__(hrp.Texto("Content-Type"), tipoMime)
			r.Cabecalho.M__define_item__(hrp.Texto("Cache-Control"), hrp.Texto("public, max-age=86400"))
			r.Corpo = conteudo.(hrp.Texto)
			return hrp.Nulo, nil
		}, "Envia um conteúdo de mídia definindo Content-Type e Cache-Control automáticos (ex: res.enviar_midia(bytes_da_imagem, \"image/png\"))"), nil
	}
	return nil, hrp.NewErroF(hrp.AtributoErro, "Atributo '%s' não existe em Resposta", nome)
}

// RotaMeta armazena os metadados declarativos registrados junto de uma rota HTTP.
type RotaMeta struct {
	Resumo             string
	Descricao          string
	Tags               []string
	Corpo              map[string]any // Esquema ou dicionário de exemplo/tipagem do corpo
	Respostas          map[string]string
	Deprecated         bool
	MensagemDeprecacao string
}

// tentarInferirMetaDeComentarios extrai comentários de documentação '///' imediatamente anteriores à chamada da rota.
func tentarInferirMetaDeComentarios(ctx *hrp.Contexto) *RotaMeta {
	if ctx == nil {
		return nil
	}
	_, cod := ctx.ObterArquivoCodigo()
	_, lin, _ := ctx.ObterPosicaoToken()
	if cod == "" || lin <= 0 {
		return nil
	}

	linhas := strings.Split(cod, "\n")
	idx := lin - 2 // Linha física L (1-based) é o índice L-1. Queremos a linha anterior, que é L-2.

	if idx >= len(linhas) {
		return nil
	}

	var docLines []string
	linhasVaziasPulas := 0
	for idx >= 0 {
		linha := strings.TrimSpace(linhas[idx])
		if strings.HasPrefix(linha, "///") {
			docLines = append([]string{linha}, docLines...)
			linhasVaziasPulas = 0
		} else if linha == "" {
			if len(docLines) > 0 {
				break
			}
			linhasVaziasPulas++
			if linhasVaziasPulas > 2 {
				break
			}
		} else {
			break
		}
		idx--
	}

	if len(docLines) == 0 {
		return nil
	}

	meta := &RotaMeta{}
	var textoSimples []string

	for _, rawLine := range docLines {
		linha := strings.TrimSpace(strings.TrimPrefix(rawLine, "///"))
		if linha == "" {
			continue
		}

		if strings.HasPrefix(linha, "@resumo ") {
			meta.Resumo = strings.TrimSpace(strings.TrimPrefix(linha, "@resumo "))
		} else if strings.HasPrefix(linha, "@descricao ") {
			meta.Descricao = strings.TrimSpace(strings.TrimPrefix(linha, "@descricao "))
		} else if strings.HasPrefix(linha, "@tags ") {
			tagsPart := strings.TrimPrefix(linha, "@tags ")
			for _, t := range strings.Split(tagsPart, ",") {
				tTrim := strings.TrimSpace(t)
				if tTrim != "" {
					meta.Tags = append(meta.Tags, tTrim)
				}
			}
		} else if strings.HasPrefix(linha, "@corpo ") || strings.HasPrefix(linha, "@corpoJson ") {
			prefix := "@corpo "
			if strings.HasPrefix(linha, "@corpoJson ") {
				prefix = "@corpoJson "
			}
			jsonStr := strings.TrimSpace(strings.TrimPrefix(linha, prefix))
			var corpoMap map[string]any
			if err := json.Unmarshal([]byte(jsonStr), &corpoMap); err == nil {
				meta.Corpo = corpoMap
			}
		} else if strings.HasPrefix(linha, "@respostas ") {
			jsonStr := strings.TrimSpace(strings.TrimPrefix(linha, "@respostas "))
			var respMap map[string]any
			if err := json.Unmarshal([]byte(jsonStr), &respMap); err == nil {
				meta.Respostas = make(map[string]string)
				for k, v := range respMap {
					if vStr, ok := v.(string); ok {
						meta.Respostas[k] = vStr
					}
				}
			}
		} else if strings.HasPrefix(linha, "@deprecated") {
			meta.Deprecated = true
			msg := strings.TrimSpace(strings.TrimPrefix(linha, "@deprecated"))
			if msg != "" {
				meta.MensagemDeprecacao = msg
			}
		} else {
			textoSimples = append(textoSimples, linha)
		}
	}

	if meta.Resumo == "" && len(textoSimples) > 0 {
		meta.Resumo = textoSimples[0]
		if len(textoSimples) > 1 {
			meta.Descricao = strings.Join(textoSimples[1:], "\n")
		}
	} else if meta.Descricao == "" && len(textoSimples) > 0 {
		meta.Descricao = strings.Join(textoSimples, "\n")
	}

	return meta
}

// Servidor representa o orquestrador do servidor HTTP nativo assíncrono.
type Servidor struct {
	rotas       map[string]map[string]hrp.Objeto // metodo -> rota -> handler
	rotasMeta   map[string]map[string]*RotaMeta  // metodo -> rota -> metadados
	middlewares []hrp.Objeto
	server      *http.Server
	mu          sync.RWMutex
	metricas    map[string]int64 // metodo+espaço+rotaPattern -> contagem
	// ctx é o contexto do script que criou o servidor; o sandbox é checado nele.
	// ponytail: capturado do contexto ativo em Nova (não há receptor); correto enquanto a criação
	// roda dentro do script dono, que é o caso normal.
	ctx *hrp.Contexto
	// execSem serializa a execução do script dos handlers (ADR 0002): Mapa/Lista são estruturas
	// Go sem trava e escrita concorrente corrompe dados ou derruba o processo.
	// ponytail: 1 handler por vez por servidor; upgrade = coleções thread-safe se houver meta de RPS.
	execSem chan struct{}
	// corsOrigens restringe o CORS (app.cors([...])); nil mantém o padrão compatível '*'.
	corsOrigens []string
	// ordemRotas guarda a ordem de registro para desempatar rotas de mesma força.
	ordemRotas map[string]int
}

// problemaInterno responde 500 em Problem Details (RFC 9457) sem detalhes internos; o detalhe
// vai só para o log, correlacionado pelo trace_id.
func problemaInterno(w http.ResponseWriter, detalhe any, metodo, caminho string) {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	traceID := hex.EncodeToString(b)
	fmt.Fprintf(os.Stderr, "\n❌ Erro interno [%s %s] trace_id=%s: %v\n", metodo, caminho, traceID, detalhe)
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(http.StatusInternalServerError)
	w.Write(fmt.Appendf(nil, `{"type":"about:blank","title":"Erro interno","status":500,"trace_id":"%s"}`, traceID))
}

// aplicarCors define os cabeçalhos CORS conforme a configuração do servidor.
func (s *Servidor) aplicarCors(w http.ResponseWriter, req *http.Request) {
	s.mu.RLock()
	origens := s.corsOrigens
	s.mu.RUnlock()
	if origens == nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
	} else {
		origem := req.Header.Get("Origin")
		permitida := false
		for _, o := range origens {
			if o == origem {
				permitida = true
				break
			}
		}
		w.Header().Add("Vary", "Origin")
		if !permitida {
			return
		}
		w.Header().Set("Access-Control-Allow-Origin", origem)
	}
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS, PUT, PATCH")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}

// EsperaMaximaHandler é quanto uma requisição espera o handler anterior terminar antes de
// responder 503 (também evita travar quando um handler chama o próprio servidor).
var EsperaMaximaHandler = 5 * time.Second

// LimiteRespostaCliente é o tamanho máximo de corpo lido por requisitar (10 MB).
const LimiteRespostaCliente = 10 << 20

// TipoServidor define e registra a classe Servidor na VM.
var TipoServidor = hrp.TipoObjeto.NewTipo("Servidor", "Servidor HTTP assíncrono")

// Tipo retorna a representação correspondente na VM.
func (s *Servidor) Tipo() *hrp.Tipo {
	return TipoServidor
}

func init() {
	// Define o inicializador padrão de alocação de memória do tipo Servidor.
	TipoServidor.Nova = func(args hrp.Tupla) (hrp.Objeto, error) {
		return &Servidor{
			rotas:       make(map[string]map[string]hrp.Objeto),
			rotasMeta:   make(map[string]map[string]*RotaMeta),
			middlewares: make([]hrp.Objeto, 0),
			metricas:    make(map[string]int64),
			ctx:         hrp.ObterContextoAtivo(),
			execSem:     make(chan struct{}, 1),
			ordemRotas:  make(map[string]int),
		}, nil
	}
}

// encadearRota compõe middlewares de rota com o handler: cada middleware recebe
// (req, res, proximo) e só segue adiante se chamar proximo().
func encadearRota(mws hrp.Tupla, final hrp.Objeto) hrp.Objeto {
	return hrp.NewMetodoOuPanic("rota", func(_ hrp.Objeto, a hrp.Tupla) (hrp.Objeto, error) {
		var passo func(i int) (hrp.Objeto, error)
		passo = func(i int) (hrp.Objeto, error) {
			if i == len(mws) {
				return hrp.Chamar(final, a)
			}
			proximo := hrp.NewMetodoOuPanic("proximo", func(_ hrp.Objeto, _ hrp.Tupla) (hrp.Objeto, error) {
				return passo(i + 1)
			}, "Segue para o próximo middleware ou para o handler da rota.")
			return hrp.Chamar(mws[i], append(a[:len(a):len(a)], proximo))
		}
		return passo(0)
	}, "Rota com middlewares encadeados.")
}

// registrarRota anexa de forma thread-safe uma rota lógica com seu respectivo handler e metadados na tabela de rotas.
func (s *Servidor) registrarRota(metodo, rota string, handler hrp.Objeto, metaObj hrp.Objeto) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.rotas[metodo] == nil {
		s.rotas[metodo] = make(map[string]hrp.Objeto)
		s.rotasMeta[metodo] = make(map[string]*RotaMeta)
	}
	if s.ordemRotas == nil { // Servidor montado sem o construtor (ex.: testes)
		s.ordemRotas = make(map[string]int)
	}
	if _, existe := s.ordemRotas[metodo+" "+rota]; !existe {
		s.ordemRotas[metodo+" "+rota] = len(s.ordemRotas)
	}
	s.rotas[metodo][rota] = handler

	if metaObj != nil && metaObj != hrp.Nulo {
		meta := &RotaMeta{}
		goMeta := hrp.ConverteParaGo(metaObj)
		if mMap, ok := goMeta.(map[string]any); ok {
			if r, ok := mMap["resumo"].(string); ok {
				meta.Resumo = r
			}
			if d, ok := mMap["descricao"].(string); ok {
				meta.Descricao = d
			}
			if tList, ok := mMap["tags"].([]any); ok {
				for _, t := range tList {
					if ts, ok := t.(string); ok {
						meta.Tags = append(meta.Tags, ts)
					}
				}
			}
			if c, ok := mMap["corpo"].(map[string]any); ok {
				meta.Corpo = c
			}
			if respMap, ok := mMap["respostas"].(map[string]any); ok {
				meta.Respostas = make(map[string]string)
				for k, v := range respMap {
					if vs, ok := v.(string); ok {
						meta.Respostas[k] = vs
					}
				}
			}
		}
		s.rotasMeta[metodo][rota] = meta
	} else {
		// Tenta inferir metadados a partir de comentários de documentação '///'
		if ctx := hrp.ObterContextoAtivo(); ctx != nil {
			if meta := tentarInferirMetaDeComentarios(ctx); meta != nil {
				s.rotasMeta[metodo][rota] = meta
			}
		}
	}
}

// M__obtem_attributo__ expõe a API do Servidor para registro de rotas, middlewares e escuta de porta.
func (s *Servidor) M__obtem_attributo__(nome string) (hrp.Objeto, error) {
	criadorRota := func(metodoHTTP string) hrp.Objeto {
		return hrp.NewMetodoOuPanic(strings.ToLower(metodoHTTP), func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
			// app.obter(rota, [middlewares...], handler, metadados?): o Mapa final é metadados;
			// funções antes do handler são middlewares de rota (req, res, proximo).
			if len(args) < 2 {
				return nil, hrp.NewErroF(hrp.TipagemErro, "%s() espera (rota, handler)", strings.ToLower(metodoHTTP))
			}
			rota, err := hrp.NewTexto(args[0])
			if err != nil {
				return nil, err
			}
			resto := args[1:]
			var meta hrp.Objeto = hrp.Nulo
			if _, ehMapa := resto[len(resto)-1].(hrp.Mapa); ehMapa && len(resto) > 1 {
				meta, resto = resto[len(resto)-1], resto[:len(resto)-1]
			}
			handler := resto[len(resto)-1]
			if mws := resto[:len(resto)-1]; len(mws) > 0 {
				handler = encadearRota(mws, handler)
			}
			s.registrarRota(metodoHTTP, string(rota.(hrp.Texto)), handler, meta)
			return hrp.Nulo, nil
		}, fmt.Sprintf("Registra uma rota para requisições %s.", metodoHTTP))
	}

	switch nome {
	case "obter":
		return criadorRota("GET"), nil
	case "postar":
		return criadorRota("POST"), nil
	case "deletar":
		return criadorRota("DELETE"), nil
	case "put", "atualizar":
		return criadorRota("PUT"), nil
	case "patch":
		return criadorRota("PATCH"), nil

	case "cors":
		// app.cors(["https://app.exemplo"]) restringe o CORS a essas origens (padrão: '*').
		return hrp.NewMetodoOuPanic("cors", func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
			if err := hrp.VerificaNumeroArgumentos("cors", false, args, 1, 1); err != nil {
				return nil, err
			}
			var itens []hrp.Objeto
			switch l := args[0].(type) {
			case *hrp.Lista:
				itens = l.Itens
			case hrp.Tupla:
				itens = l
			default:
				return nil, hrp.NewErroF(hrp.TipagemErro, "cors esperava uma Lista de origens (Texto)")
			}
			origens := make([]string, 0, len(itens))
			for _, it := range itens {
				t, ok := it.(hrp.Texto)
				if !ok {
					return nil, hrp.NewErroF(hrp.TipagemErro, "cors: cada origem deve ser Texto")
				}
				origens = append(origens, string(t))
			}
			s.mu.Lock()
			s.corsOrigens = origens
			s.mu.Unlock()
			return hrp.Nulo, nil
		}, "Restringe o CORS às origens informadas."), nil

	case "usar":
		// Enfileira um middleware global no pipeline de execução do servidor.
		return hrp.NewMetodoOuPanic("usar", func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
			if err := hrp.VerificaNumeroArgumentos("usar", false, args, 1, 1); err != nil {
				return nil, err
			}
			s.mu.Lock()
			s.middlewares = append(s.middlewares, args[0])
			s.mu.Unlock()
			return hrp.Nulo, nil
		}, "Registra um middleware global."), nil

	case "servir_app":
		// Registra um SPA Handler para servir arquivos estáticos de compilações do frontend e manipular roteamento SPA.
		return hrp.NewMetodoOuPanic("servir_app", func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
			if err := hrp.VerificaNumeroArgumentos("servir_app", false, args, 2, 3); err != nil {
				return nil, err
			}
			diretorioDist, err := hrp.NewTexto(args[0])
			if err != nil {
				return nil, err
			}
			componenteRaiz := args[1]
			var metadados hrp.Objeto = hrp.Nulo
			if len(args) == 3 {
				metadados = args[2]
			}

			handler := ServirAppHandler(string(diretorioDist.(hrp.Texto)), componenteRaiz, metadados)
			s.registrarRota("GET", "/*", handler, nil)
			return hrp.Nulo, nil
		}, "Registra e configura as rotas de uma aplicação SPA reativa."), nil

	case "escutar":
		// Ativa e inicia a escuta HTTP na porta TCP indicada.
		return hrp.NewMetodoOuPanic("escutar", func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
			if err := hrp.ExigirPermissaoNoContexto(s.ctx, hrp.RecursoRede); err != nil {
				return nil, err
			}

			if err := hrp.VerificaNumeroArgumentos("escutar", false, args, 1, 2); err != nil {
				return nil, err
			}
			porta, err := hrp.NewTexto(args[0])
			if err != nil {
				return nil, err
			}

			bloquear := false
			if len(args) > 1 {
				if b, ok := args[1].(hrp.Booleano); ok {
					bloquear = bool(b)
				}
			}

			// Injeta as rotas automáticas de documentação (/docs e /docs.json) se não sobrescritas pelo usuário
			s.mu.Lock()
			if s.rotas["GET"] == nil {
				s.rotas["GET"] = make(map[string]hrp.Objeto)
			}
			if s.execSem == nil { // Servidor montado sem o construtor: sem isso toda requisição daria 503
				s.execSem = make(chan struct{}, 1)
			}
			docsLigado := os.Getenv("HARPIA_DOCS") != "0" // HARPIA_DOCS=0 desliga /docs e /docs.json
			if docsLigado && s.rotas["GET"]["/docs.json"] == nil {
				s.rotas["GET"]["/docs.json"] = hrp.NewMetodoOuPanic("docs_json", func(_ hrp.Objeto, handlerArgs hrp.Tupla) (hrp.Objeto, error) {
					if len(handlerArgs) < 2 {
						return nil, fmt.Errorf("argumentos inválidos no docs_json")
					}
					res := handlerArgs[1].(*Resposta)
					spec, err := met_gerar_openapi(nil, hrp.Tupla{s})
					if err != nil {
						res.Status = 500
						res.Corpo = hrp.Texto(fmt.Sprintf("Erro ao gerar documentação: %v", err))
						return hrp.Nulo, nil
					}
					res.Status = 200
					res.Cabecalho.M__define_item__(hrp.Texto("Content-Type"), hrp.Texto("application/json"))
					res.Corpo = spec.(hrp.Texto)
					return hrp.Nulo, nil
				}, "Retorna as especificações OpenAPI do servidor em formato JSON.")
			}
			if docsLigado && s.rotas["GET"]["/docs"] == nil && s.rotas["GET"]["/*"] == nil {
				s.rotas["GET"]["/docs"] = hrp.NewMetodoOuPanic("docs_html", func(_ hrp.Objeto, handlerArgs hrp.Tupla) (hrp.Objeto, error) {
					if len(handlerArgs) < 2 {
						return nil, fmt.Errorf("argumentos inválidos no docs_html")
					}
					res := handlerArgs[1].(*Resposta)
					res.Status = 200
					res.Cabecalho.M__define_item__(hrp.Texto("Content-Type"), hrp.Texto("text/html; charset=utf-8"))
					res.Corpo = hrp.Texto(htmlDocsTemplate)
					return hrp.Nulo, nil
				}, "Retorna a interface visual de documentação interativa das rotas da API.")
			}
			if s.rotas["GET"]["/livez"] == nil {
				s.rotas["GET"]["/livez"] = hrp.NewMetodoOuPanic("livez_probe", func(_ hrp.Objeto, handlerArgs hrp.Tupla) (hrp.Objeto, error) {
					if len(handlerArgs) < 2 {
						return nil, fmt.Errorf("argumentos inválidos no livez_probe")
					}
					res := handlerArgs[1].(*Resposta)
					res.Status = http.StatusOK
					res.Cabecalho.M__define_item__(hrp.Texto("Content-Type"), hrp.Texto("application/json"))
					res.Corpo = hrp.Texto(`{"status":"ok"}`)
					return hrp.Nulo, nil
				}, "Probe de liveness do processo (/livez).")
			}
			if s.rotas["GET"]["/readyz"] == nil {
				s.rotas["GET"]["/readyz"] = hrp.NewMetodoOuPanic("readyz_probe", func(_ hrp.Objeto, handlerArgs hrp.Tupla) (hrp.Objeto, error) {
					if len(handlerArgs) < 2 {
						return nil, fmt.Errorf("argumentos inválidos no readyz_probe")
					}
					res := handlerArgs[1].(*Resposta)
					res.Cabecalho.M__define_item__(hrp.Texto("Content-Type"), hrp.Texto("application/json"))
					// ponytail: pronto = processo escutando; um 500 de um handler não tira o pod de rotação.
					// Upgrade: checar dependências reais (ex.: ping do banco) quando houver.
					res.Status = http.StatusOK
					res.Corpo = hrp.Texto(`{"status":"ready"}`)
					return hrp.Nulo, nil
				}, "Probe de readiness da infraestrutura (/readyz).")
			}
			s.mu.Unlock()

			mux := http.NewServeMux()
			mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
				defer func() {
					if r := recover(); r != nil {
						problemaInterno(w, fmt.Sprintf("pânico: %v\n%s", r, debug.Stack()), req.Method, req.URL.Path)
					}
				}()

				// Injeção dos cabeçalhos OWASP de Segurança Padrão recomendados
				w.Header().Set("X-Content-Type-Options", "nosniff")
				w.Header().Set("X-Frame-Options", "DENY")
				w.Header().Set("X-XSS-Protection", "1; mode=block")
				w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

				// CORS: '*' por padrão (compatível); restrito às origens de app.cors([...]) quando configurado
				s.aplicarCors(w, req)

				if req.Method == "OPTIONS" {
					w.WriteHeader(http.StatusOK)
					return
				}

				s.mu.RLock()
				rotasMetodo := s.rotas[req.Method]

				var handler hrp.Objeto
				var params map[string]string
				var melhorRotaPattern string

				// ponytail: prioriza rotas mais especificas (com mais segmentos estaticos) antes de cair no curinga "/*".
				// Itera todas as rotas; quando match, calcula a "forca" do match (mais segmentos estaticos = mais forte).
				// So usa o wildcard /* se nenhuma rota mais especifica casar.
				var melhorForca int = -1
				for rotaPattern, h := range rotasMetodo {
					match, parsedParams := matchRoute(rotaPattern, req.URL.Path)
					if match {
						forca := calcularForcaRota(rotaPattern)
						// empate de força: vence a rota registrada primeiro (o map não tem ordem)
						if forca > melhorForca || (forca == melhorForca && s.ordemRotas[req.Method+" "+rotaPattern] < s.ordemRotas[req.Method+" "+melhorRotaPattern]) {
							melhorForca = forca
							handler = h
							params = parsedParams
							melhorRotaPattern = rotaPattern
						}
					}
				}
				s.mu.RUnlock()

				if handler == nil {
					w.WriteHeader(http.StatusNotFound)
					w.Write([]byte("Rota não encontrada"))
					return
				}

				// Incrementa métricas de forma thread-safe
				if melhorRotaPattern != "" {
					s.mu.Lock()
					if s.metricas == nil {
						s.metricas = make(map[string]int64)
					}
					s.metricas[req.Method+" "+melhorRotaPattern]++
					s.mu.Unlock()
				}

				// Cabeçalhos
				reqHeaders := hrp.NewMapaVazio()
				for k, vals := range req.Header {
					reqHeaders.M__define_item__(hrp.Texto(k), hrp.Texto(strings.Join(vals, ", ")))
				}

				// Limita o corpo da requisição a 10MB para prevenir DoS por exaustão de memória
				bodyReader := http.MaxBytesReader(w, req.Body, 10*1024*1024)
				bodyBytes, err := io.ReadAll(bodyReader)
				if err != nil {
					w.WriteHeader(http.StatusRequestEntityTooLarge)
					w.Write([]byte("Corpo da requisição excede o limite permitido"))
					return
				}
				// Se houver parâmetros de rota (:id), mescla no objeto de requisição adicionando dinamicamente
				reqParams := hrp.NewMapaVazio()
				for k, v := range params {
					reqParams.M__define_item__(hrp.Texto(k), hrp.Texto(v))
				}

				// Parse automático do corpo JSON quando Content-Type indica JSON
				var corpoJsonParsed hrp.Objeto = hrp.Nulo
				contentType := req.Header.Get("Content-Type")
				if len(bodyBytes) > 0 && strings.Contains(strings.ToLower(contentType), "application/json") {
					var rawJson any
					if jsonErr := json.Unmarshal(bodyBytes, &rawJson); jsonErr == nil {
						corpoJsonParsed = hrp.ConverteDeGo(rawJson)
					}
				}

				// Parse dos query parameters
				reqQuery := hrp.NewMapaVazio()
				for k, vals := range req.URL.Query() {
					reqQuery.M__define_item__(hrp.Texto(k), hrp.Texto(strings.Join(vals, ", ")))
				}

				reqObj := &Requisicao{
					Metodo:     hrp.Texto(req.Method),
					Caminho:    hrp.Texto(req.URL.Path),
					Cabecalho:  reqHeaders,
					Corpo:      hrp.Texto(bodyBytes),
					Parametros: reqParams,
					Query:      reqQuery,
					CorpoJson:  corpoJsonParsed,
				}

				resObj := &Resposta{
					Status:    hrp.Inteiro(http.StatusOK),
					Cabecalho: hrp.NewMapaVazio(),
					Corpo:     hrp.Texto(""),
				}

				// Roda middlewares encadeados
				s.mu.RLock()
				mws := make([]hrp.Objeto, len(s.middlewares))
				copy(mws, s.middlewares)
				s.mu.RUnlock()

				var executarpróximo func(idx int) error
				executarpróximo = func(idx int) error {
					if idx >= len(mws) {
						// Finalizou os middlewares: executa o handler principal
						_, errHandler := hrp.Chamar(handler, hrp.Tupla{reqObj, resObj})
						return errHandler
					}

					mw := mws[idx]
					próximoCb := hrp.NewMetodoOuPanic("próximo", func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
						err := executarpróximo(idx + 1)
						if err != nil {
							return nil, err
						}
						return hrp.Nulo, nil
					}, "Avança para o próximo middleware ou handler")

					// Middleware declarado com 2 parâmetros (req, res) roda e avança sozinho; os demais
					// recebem (req, res, próximo). Decide pela aridade, não pelo texto de um erro.
					if fn, ok := mw.(*hrp.Funcao); ok && len(fn.Args()) == 2 {
						if _, errMw := hrp.Chamar(mw, hrp.Tupla{reqObj, resObj}); errMw != nil {
							return errMw
						}
						return executarpróximo(idx + 1)
					}
					_, errMw := hrp.Chamar(mw, hrp.Tupla{reqObj, resObj, próximoCb})
					return errMw
				}

				select {
				case s.execSem <- struct{}{}:
				case <-time.After(EsperaMaximaHandler):
					w.Header().Set("Content-Type", "text/plain; charset=utf-8")
					w.Header().Set("Retry-After", "1")
					w.WriteHeader(http.StatusServiceUnavailable)
					w.Write([]byte("Servidor ocupado: os handlers rodam um por vez e o anterior não terminou a tempo (chamar o próprio servidor de dentro de um handler não é suportado)"))
					return
				}
				// defer: um pânico no handler não pode deixar o semáforo preso (503 para sempre)
				errChain := func() error {
					defer func() { <-s.execSem }()
					return executarpróximo(0)
				}()
				if errChain != nil {
					// 500 sem derrubar o servidor e sem vazar o erro (SQL, caminhos, stack) ao cliente
					problemaInterno(w, errChain, req.Method, req.URL.Path)
					return
				}

				// Escreve cabeçalhos de resposta (valores não-Texto são convertidos, ex.: Inteiro)
				for k, v := range resObj.Cabecalho {
					vTexto, errConv := hrp.NewTexto(v)
					if errConv != nil {
						fmt.Fprintf(os.Stderr, "⚠️ cabeçalho '%s' ignorado: %v\n", k, errConv)
						continue
					}
					w.Header().Set(k, string(vTexto.(hrp.Texto)))
				}

				w.WriteHeader(int(resObj.Status))
				w.Write([]byte(resObj.Corpo))
			})

			s.server = &http.Server{
				Addr:         ":" + string(porta.(hrp.Texto)),
				Handler:      mux,
				ReadTimeout:  5 * time.Second,
				WriteTimeout: 10 * time.Second,
				IdleTimeout:  120 * time.Second,
			}

			if bloquear {
				if errListen := s.server.ListenAndServe(); errListen != nil && errListen != http.ErrServerClosed {
					return nil, hrp.NewErroF(hrp.ErroDeSistema, "Erro ao escutar porta: %v", errListen)
				}
				return hrp.Nulo, nil
			}

			// Inicia o servidor em background para manter concorrência cooperativa/assíncrona
			errCh := make(chan error, 1)
			go func() {
				if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					errCh <- err
				}
			}()

			select {
			case err := <-errCh:
				return nil, hrp.NewErroF(hrp.ErroDeSistema, "Erro ao escutar porta: %v", err)
			case <-time.After(100 * time.Millisecond):
			}

			return hrp.Nulo, nil
		}, "Inicia a escuta de conexões de rede HTTP de forma síncrona ou assíncrona."), nil

	case "fechar":
		// Fecha o servidor de forma limpa, permitindo que as conexões ativas terminem (Shutdown).
		return hrp.NewMetodoOuPanic("fechar", func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
			if s.server != nil {
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()
				s.server.Shutdown(ctx)
			}
			return hrp.Nulo, nil
		}, "Encerra as atividades do servidor de forma controlada."), nil
	}

	return nil, hrp.NewErroF(hrp.AtributoErro, "Atributo '%s' não existe no Servidor", nome)
}

// calcularForcaRota retorna um peso para o pattern: segmentos estaticos valem 10,
// segmentos wildcard (*) valem 1, segmentos parametrizados (:id) valem 5.
// Usado pelo router para preferir rotas mais especificas sobre o wildcard "/*".
func calcularForcaRota(pattern string) int {
	if pattern == "*" || pattern == "/*" {
		return 0
	}
	parts := strings.Split(strings.Trim(pattern, "/"), "/")
	forca := 0
	for _, p := range parts {
		switch {
		case p == "*":
			forca += 1
		case strings.HasPrefix(p, ":"):
			forca += 5
		default:
			forca += 10
		}
	}
	return forca
}

// matchRoute realiza a análise de expressão regular/parâmetros dinâmicos simples para rotas (:id, *).
func matchRoute(pattern, path string) (bool, map[string]string) {
	if pattern == "*" || pattern == "/*" {
		return true, make(map[string]string)
	}

	pParts := strings.Split(strings.Trim(pattern, "/"), "/")
	uParts := strings.Split(strings.Trim(path, "/"), "/")

	// Suporte a curingas no final do caminho (ex: "/*" ou "/rotas/*")
	if len(pParts) > 0 && pParts[len(pParts)-1] == "*" {
		if len(uParts) < len(pParts)-1 {
			return false, nil
		}
		params := make(map[string]string)
		for i := 0; i < len(pParts)-1; i++ {
			if strings.HasPrefix(pParts[i], ":") {
				params[pParts[i][1:]] = uParts[i]
				continue
			}
			if pParts[i] != uParts[i] {
				return false, nil
			}
		}
		return true, params
	}

	if len(pParts) != len(uParts) {
		return false, nil
	}

	params := make(map[string]string)
	for i := 0; i < len(pParts); i++ {
		if strings.HasPrefix(pParts[i], ":") {
			params[pParts[i][1:]] = uParts[i]
			continue
		}
		if pParts[i] != uParts[i] {
			return false, nil
		}
	}
	return true, params
}

// met_http_requisitar realiza uma requisição HTTP do lado do cliente.
func met_http_requisitar(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if err := hrp.ExigirPermissao(inst, hrp.RecursoRede); err != nil {
		return nil, err
	}

	if err := hrp.VerificaNumeroArgumentos("requisitar", false, args, 2, 4); err != nil {
		return nil, err
	}

	metodo, err := hrp.NewTexto(args[0])
	if err != nil {
		return nil, err
	}

	url, err := hrp.NewTexto(args[1])
	if err != nil {
		return nil, err
	}

	var reqBody io.Reader
	if len(args) >= 3 && args[2] != hrp.Nulo {
		corpoText, err := hrp.NewTexto(args[2])
		if err != nil {
			return nil, err
		}
		reqBody = strings.NewReader(string(corpoText.(hrp.Texto)))
	}

	req, err := http.NewRequest(string(metodo.(hrp.Texto)), string(url.(hrp.Texto)), reqBody)
	if err != nil {
		return nil, hrp.NewErroF(hrp.ErroDeSistema, "Erro ao criar requisição cliente: %v", err)
	}

	if len(args) >= 4 && args[3] != hrp.Nulo {
		if cabeçalhos, ok := args[3].(hrp.Mapa); ok {
			for k, v := range cabeçalhos {
				vTexto, errConv := hrp.NewTexto(v)
				if errConv != nil {
					return nil, errConv
				}
				req.Header.Set(k, string(vTexto.(hrp.Texto)))
			}
		}
	}

	clienteHTTP := &http.Client{
		Timeout: 30 * time.Second,
	}

	res, err := clienteHTTP.Do(req)
	if err != nil {
		return nil, hrp.NewErroF(hrp.ErroDeSistema, "Erro ao executar requisição HTTP: %v", err)
	}
	defer res.Body.Close()

	bodyBytes, err := io.ReadAll(io.LimitReader(res.Body, LimiteRespostaCliente+1))
	if err != nil {
		return nil, hrp.NewErroF(hrp.ErroDeSistema, "Erro ao ler a resposta HTTP: %v", err)
	}
	if len(bodyBytes) > LimiteRespostaCliente {
		return nil, hrp.NewErroF(hrp.ErroDeLimite, "Resposta HTTP excede o limite de %d MB", LimiteRespostaCliente>>20)
	}
	resHeaders := hrp.NewMapaVazio()
	for k, vals := range res.Header {
		resHeaders.M__define_item__(hrp.Texto(k), hrp.Texto(strings.Join(vals, ", ")))
	}

	resObj := &Resposta{
		Status:    hrp.Inteiro(res.StatusCode),
		Cabecalho: resHeaders,
		Corpo:     hrp.Texto(bodyBytes),
	}
	return resObj, nil
}

func init() {
	// Registra o módulo 'http' no ecossistema central do Harpia.
	hrp.RegistraModuloImpl(&hrp.ModuloImpl{
		Info: hrp.ModuloInfo{
			Nome:    "http",
			Arquivo: "stdlib/http",
		},
		Constantes: hrp.Mapa{
			"Servidor": TipoServidor,
			"Status": hrp.Mapa{
				"OK":                  hrp.Inteiro(200),
				"Criado":              hrp.Inteiro(201),
				"Aceito":              hrp.Inteiro(202),
				"SemConteudo":         hrp.Inteiro(204),
				"MovidoPara":          hrp.Inteiro(301),
				"NaoModificado":       hrp.Inteiro(304),
				"RequisicaoRuim":      hrp.Inteiro(400),
				"NaoAutorizado":       hrp.Inteiro(401),
				"Proibido":            hrp.Inteiro(403),
				"NaoEncontrado":       hrp.Inteiro(404),
				"MetodoInvalido":      hrp.Inteiro(405),
				"Conflito":            hrp.Inteiro(409),
				"EntidadeGrande":      hrp.Inteiro(413),
				"NaoProcessavel":      hrp.Inteiro(422),
				"MuitasRequisicoes":   hrp.Inteiro(429),
				"ErroInterno":         hrp.Inteiro(500),
				"NaoImplementado":     hrp.Inteiro(501),
				"ServicoIndisponivel": hrp.Inteiro(503),
			},
		},
		Metodos: []*hrp.Metodo{
			hrp.NewMetodoOuPanic("requisitar", met_http_requisitar, "Realiza uma requisição HTTP cliente (GET, POST, PUT, DELETE, etc.)."),
			hrp.NewMetodoOuPanic("assinar_hmac", met_assinar_hmac, "Gera assinatura HMAC SHA-256 (chave, mensagem)"),
			hrp.NewMetodoOuPanic("verificar_hmac", met_verificar_hmac, "Valida assinatura HMAC SHA-256 (chave, mensagem, assinatura)"),
			hrp.NewMetodoOuPanic("gerar_openapi", met_gerar_openapi, "Gera spec OpenAPI 3.0 para o Servidor"),
		},
	})
}
