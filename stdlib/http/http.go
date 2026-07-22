package http

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/mat-dgruber/Harpia/hrp"
)

// Requisicao representa a requisição HTTP recebida pelo servidor.
type Requisicao struct {
	Metodo     hrp.Texto
	Caminho    hrp.Texto
	Cabecalho  hrp.Mapa
	Corpo      hrp.Texto
	Parametros hrp.Mapa
	Query      hrp.Mapa   // Parâmetros de consulta (query string)
	CorpoJson  hrp.Objeto // Body deserializado de JSON automaticamente (Nulo se não for JSON)
}

var TipoRequisicao = hrp.NewTipo("Requisicao", "Objeto que representa uma requisição HTTP")

func (r *Requisicao) Tipo() *hrp.Tipo {
	return TipoRequisicao
}

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
	case "query":
		return r.Query, nil
	case "corpoJson":
		if r.CorpoJson != nil {
			return r.CorpoJson, nil
		}
		return hrp.Nulo, nil
	}
	return nil, hrp.NewErroF(hrp.AtributoErro, "Atributo '%s' não existe em Requisicao", nome)
}

// Resposta representa a resposta HTTP que o servidor devolverá.
type Resposta struct {
	Status    hrp.Inteiro
	Cabecalho hrp.Mapa
	Corpo     hrp.Texto
}

var TipoResposta = hrp.NewTipo("Resposta", "Objeto que representa uma resposta HTTP")

func (r *Resposta) Tipo() *hrp.Tipo {
	return TipoResposta
}

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

func (r *Resposta) M__obtem_attributo__(nome string) (hrp.Objeto, error) {
	switch nome {
	case "status":
		return r.Status, nil
	case "corpo":
		return r.Corpo, nil
	case "cabecalho":
		return r.Cabecalho, nil
	case "definir_cabecalho":
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
		}, ""), nil
	case "enviarJson":
		return hrp.NewMetodoOuPanic("enviarJson", func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
			if err := hrp.VerificaNumeroArgumentos("enviarJson", false, args, 1, 1); err != nil {
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

	case "definirStatus":
		return hrp.NewMetodoOuPanic("definirStatus", func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
			if err := hrp.VerificaNumeroArgumentos("definirStatus", false, args, 1, 1); err != nil {
				return nil, err
			}
			statusInt, err := hrp.NewInteiro(args[0])
			if err != nil {
				return nil, err
			}
			r.Status = statusInt.(hrp.Inteiro)
			return hrp.Nulo, nil
		}, "Define o código de status HTTP da resposta (ex: res.definirStatus(404))"), nil

	case "erroJson":
		return hrp.NewMetodoOuPanic("erroJson", func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
			if err := hrp.VerificaNumeroArgumentos("erroJson", false, args, 2, 2); err != nil {
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
		}, "Envia uma resposta de erro JSON com status e mensagem (ex: res.erroJson(404, \"Não encontrado\"))"), nil
	}
	return nil, hrp.NewErroF(hrp.AtributoErro, "Atributo '%s' não existe em Resposta", nome)
}

// Servidor representa o servidor HTTP nativo.
type Servidor struct {
	rotas       map[string]map[string]hrp.Objeto // metodo -> rota -> handler
	middlewares []hrp.Objeto
	server      *http.Server
	mu          sync.RWMutex
}

var TipoServidor = hrp.TipoObjeto.NewTipo("Servidor", "Servidor HTTP assíncrono")

func (s *Servidor) Tipo() *hrp.Tipo {
	return TipoServidor
}

func init() {
	TipoServidor.Nova = func(args hrp.Tupla) (hrp.Objeto, error) {
		return &Servidor{
			rotas:       make(map[string]map[string]hrp.Objeto),
			middlewares: make([]hrp.Objeto, 0),
		}, nil
	}
}

func (s *Servidor) registrarRota(metodo, rota string, handler hrp.Objeto) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.rotas[metodo] == nil {
		s.rotas[metodo] = make(map[string]hrp.Objeto)
	}
	s.rotas[metodo][rota] = handler
}

func (s *Servidor) M__obtem_attributo__(nome string) (hrp.Objeto, error) {
	switch nome {
	case "obter":
		return hrp.NewMetodoOuPanic("obter", func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
			if err := hrp.VerificaNumeroArgumentos("obter", false, args, 2, 2); err != nil {
				return nil, err
			}
			rota, err := hrp.NewTexto(args[0])
			if err != nil {
				return nil, err
			}
			s.registrarRota("GET", string(rota.(hrp.Texto)), args[1])
			return hrp.Nulo, nil
		}, ""), nil

	case "postar":
		return hrp.NewMetodoOuPanic("postar", func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
			if err := hrp.VerificaNumeroArgumentos("postar", false, args, 2, 2); err != nil {
				return nil, err
			}
			rota, err := hrp.NewTexto(args[0])
			if err != nil {
				return nil, err
			}
			s.registrarRota("POST", string(rota.(hrp.Texto)), args[1])
			return hrp.Nulo, nil
		}, ""), nil

	case "deletar":
		return hrp.NewMetodoOuPanic("deletar", func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
			if err := hrp.VerificaNumeroArgumentos("deletar", false, args, 2, 2); err != nil {
				return nil, err
			}
			rota, err := hrp.NewTexto(args[0])
			if err != nil {
				return nil, err
			}
			s.registrarRota("DELETE", string(rota.(hrp.Texto)), args[1])
			return hrp.Nulo, nil
		}, ""), nil

	case "usar":
		return hrp.NewMetodoOuPanic("usar", func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
			if err := hrp.VerificaNumeroArgumentos("usar", false, args, 1, 1); err != nil {
				return nil, err
			}
			s.mu.Lock()
			s.middlewares = append(s.middlewares, args[0])
			s.mu.Unlock()
			return hrp.Nulo, nil
		}, ""), nil

	case "servir_app":
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
			s.registrarRota("GET", "/*", handler)
			return hrp.Nulo, nil
		}, ""), nil

	case "escutar":
		return hrp.NewMetodoOuPanic("escutar", func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
			ctx := hrp.ObterContextoAtivo()
			if ctx != nil {
				if err := ctx.VerificarPermissaoRede(); err != nil {
					return nil, err
				}
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

			mux := http.NewServeMux()
			mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
				defer func() {
					if r := recover(); r != nil {
						fmt.Fprintf(os.Stderr, "\n🔥 Erro interno no Servidor HTTP (Pânico):\n  ➔ %v\n\n", r)
						w.WriteHeader(http.StatusInternalServerError)
						w.Write(fmt.Appendf(nil, "Erro interno do servidor (Pânico): %v", r))
					}
				}()

				// Injeção dos cabeçalhos OWASP de Segurança Padrão
				w.Header().Set("X-Content-Type-Options", "nosniff")
				w.Header().Set("X-Frame-Options", "DENY")
				w.Header().Set("X-XSS-Protection", "1; mode=block")
				w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

				// ponytail: suporte transparente a CORS Preflight OPTIONS global para facilitar desenvolvimento de SPA
				if req.Method == "OPTIONS" {
					w.Header().Set("Access-Control-Allow-Origin", "*")
					w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS, PUT, PATCH")
					if reqHeaders := req.Header.Get("Access-Control-Request-Headers"); reqHeaders != "" {
						w.Header().Set("Access-Control-Allow-Headers", reqHeaders)
					} else {
						w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
					}
					w.WriteHeader(http.StatusOK)
					return
				}


				s.mu.RLock()
				rotasMetodo := s.rotas[req.Method]
				s.mu.RUnlock()

				var handler hrp.Objeto
				var params map[string]string

				// Roteador dinâmico simples (:id)
				for rotaPattern, h := range rotasMetodo {
					match, parsedParams := matchRoute(rotaPattern, req.URL.Path)
					if match {
						handler = h
						params = parsedParams
						break
					}
				}

				if handler == nil {
					w.WriteHeader(http.StatusNotFound)
					w.Write([]byte("Rota não encontrada"))
					return
				}

				// Cabecalhos mapa
				reqHeaders := hrp.NewMapaVazio()
				for k, vals := range req.Header {
					reqHeaders.M__define_item__(hrp.Texto(k), hrp.Texto(strings.Join(vals, ", ")))
				}

				bodyBytes, _ := io.ReadAll(req.Body)
				// Se houver parâmetros de rota (:id), mescla no objeto de requisição adicionando dinamicamente
				reqParams := hrp.NewMapaVazio()
				for k, v := range params {
					reqParams.M__define_item__(hrp.Texto(k), hrp.Texto(v))
				}

				// Parse automático do corpo JSON quando Content-Type indica JSON
				var corpoJsonParsed hrp.Objeto = hrp.Nulo
				contentType := req.Header.Get("Content-Type")
				if len(bodyBytes) > 0 && strings.Contains(contentType, "application/json") {
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

				var executarProximo func(idx int) error
				executarProximo = func(idx int) error {
					if idx >= len(mws) {
						// Finalizou os middlewares: executa o handler principal
						_, errHandler := hrp.Chamar(handler, hrp.Tupla{reqObj, resObj})
						return errHandler
					}

					mw := mws[idx]
					proximoCb := hrp.NewMetodoOuPanic("proximo", func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
						err := executarProximo(idx + 1)
						if err != nil {
							return nil, err
						}
						return hrp.Nulo, nil
					}, "Avança para o próximo middleware ou handler")

					// Tenta chamar o middleware com (req, res, proximo)
					_, errMw := hrp.Chamar(mw, hrp.Tupla{reqObj, resObj, proximoCb})
					if errMw != nil {
						// Se o erro foi por falta de argumento (middleware declarou só 2 params req, res), chama com 2 params e avança
						if strings.Contains(errMw.Error(), "esperava no máximo 2 argumentos") {
							_, errMw2 := hrp.Chamar(mw, hrp.Tupla{reqObj, resObj})
							if errMw2 != nil {
								return errMw2
							}
							return executarProximo(idx + 1)
						}
						return errMw
					}
					return nil
				}

				if errChain := executarProximo(0); errChain != nil {
					fmt.Fprintf(os.Stderr, "\n❌ Erro durante a requisição [%s %s]:\n", req.Method, req.URL.Path)
					hrp.LancarErro(errChain)
					fmt.Fprintln(os.Stderr)
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(errChain.Error()))
					return
				}

				// ponytail: adiciona cabeçalhos CORS por padrão nas respostas normais
				w.Header().Set("Access-Control-Allow-Origin", "*")
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS, PUT, PATCH")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

				// Escreve cabecalhos de resposta
				for k, v := range resObj.Cabecalho {
					w.Header().Set(k, string(v.(hrp.Texto)))
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
			var errListen error
			go func() {
				if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					errListen = err
				}
			}()

			// Aguarda breve estabilização
			time.Sleep(100 * time.Millisecond)
			if errListen != nil {
				return nil, hrp.NewErroF(hrp.ErroDeSistema, "Erro ao escutar porta: %v", errListen)
			}

			return hrp.Nulo, nil
		}, ""), nil

	case "fechar":
		return hrp.NewMetodoOuPanic("fechar", func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
			if s.server != nil {
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()
				s.server.Shutdown(ctx)
			}
			return hrp.Nulo, nil
		}, ""), nil
	}

	return nil, hrp.NewErroF(hrp.AtributoErro, "Atributo '%s' não existe no Servidor", nome)
}

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

// met_http_requisitar realiza uma requisição HTTP Cliente.
func met_http_requisitar(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	ctx := hrp.ObterContextoAtivo()
	if ctx != nil {
		if err := ctx.VerificarPermissaoRede(); err != nil {
			return nil, err
		}
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
				vTexto, _ := hrp.NewTexto(v)
				req.Header.Set(k, string(vTexto.(hrp.Texto)))
			}
		}
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, hrp.NewErroF(hrp.ErroDeSistema, "Erro ao executar requisição HTTP: %v", err)
	}
	defer res.Body.Close()

	bodyBytes, _ := io.ReadAll(res.Body)
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
			hrp.NewMetodoOuPanic("requisitar", met_http_requisitar, ""),
			hrp.NewMetodoOuPanic("assinar_hmac", met_assinar_hmac, "Gera assinatura HMAC SHA-256 (chave, mensagem)"),
			hrp.NewMetodoOuPanic("verificar_hmac", met_verificar_hmac, "Valida assinatura HMAC SHA-256 (chave, mensagem, assinatura)"),
			hrp.NewMetodoOuPanic("gerar_openapi", met_gerar_openapi, "Gera spec OpenAPI 3.0 para o Servidor"),
		},
	})
}
