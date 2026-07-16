package http

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/natanfeitosa/portuscript/ptst"
)

// Requisicao representa a requisição HTTP recebida pelo servidor.
type Requisicao struct {
	Metodo     ptst.Texto
	Caminho    ptst.Texto
	Cabecalho  ptst.Mapa
	Corpo      ptst.Texto
	Parametros ptst.Mapa
}

var TipoRequisicao = ptst.NewTipo("Requisicao", "Objeto que representa uma requisição HTTP")

func (r *Requisicao) Tipo() *ptst.Tipo {
	return TipoRequisicao
}

func (r *Requisicao) M__obtem_attributo__(nome string) (ptst.Objeto, error) {
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
	}
	return nil, ptst.NewErroF(ptst.AtributoErro, "Atributo '%s' não existe em Requisicao", nome)
}

// Resposta representa a resposta HTTP que o servidor devolverá.
type Resposta struct {
	Status    ptst.Inteiro
	Cabecalho ptst.Mapa
	Corpo     ptst.Texto
}

var TipoResposta = ptst.NewTipo("Resposta", "Objeto que representa uma resposta HTTP")

func (r *Resposta) Tipo() *ptst.Tipo {
	return TipoResposta
}

func (r *Resposta) M__define_atributo__(nome string, valor ptst.Objeto) error {
	switch nome {
	case "status":
		statusInt, err := ptst.NewInteiro(valor)
		if err != nil {
			return err
		}
		r.Status = statusInt.(ptst.Inteiro)
		return nil
	case "corpo":
		corpoTexto, err := ptst.NewTexto(valor)
		if err != nil {
			return err
		}
		r.Corpo = corpoTexto.(ptst.Texto)
		return nil
	}
	return ptst.NewErroF(ptst.AtributoErro, "Atributo '%s' não pode ser modificado em Resposta", nome)
}

func (r *Resposta) M__obtem_attributo__(nome string) (ptst.Objeto, error) {
	switch nome {
	case "status":
		return r.Status, nil
	case "corpo":
		return r.Corpo, nil
	case "cabecalho":
		return r.Cabecalho, nil
	case "definir_cabecalho":
		return ptst.NewMetodoOuPanic("definir_cabecalho", func(inst ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
			if err := ptst.VerificaNumeroArgumentos("definir_cabecalho", false, args, 2, 2); err != nil {
				return nil, err
			}
			k, err := ptst.NewTexto(args[0])
			if err != nil {
				return nil, err
			}
			v, err := ptst.NewTexto(args[1])
			if err != nil {
				return nil, err
			}
			r.Cabecalho.M__define_item__(k, v)
			return ptst.Nulo, nil
		}, ""), nil
	}
	return nil, ptst.NewErroF(ptst.AtributoErro, "Atributo '%s' não existe em Resposta", nome)
}

// Servidor representa o servidor HTTP nativo.
type Servidor struct {
	rotas       map[string]map[string]ptst.Objeto // metodo -> rota -> handler
	middlewares []ptst.Objeto
	server      *http.Server
	mu          sync.RWMutex
}

var TipoServidor = ptst.TipoObjeto.NewTipo("Servidor", "Servidor HTTP assíncrono")

func (s *Servidor) Tipo() *ptst.Tipo {
	return TipoServidor
}

func init() {
	TipoServidor.Nova = func(args ptst.Tupla) (ptst.Objeto, error) {
		return &Servidor{
			rotas:       make(map[string]map[string]ptst.Objeto),
			middlewares: make([]ptst.Objeto, 0),
		}, nil
	}
}

func (s *Servidor) registrarRota(metodo, rota string, handler ptst.Objeto) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.rotas[metodo] == nil {
		s.rotas[metodo] = make(map[string]ptst.Objeto)
	}
	s.rotas[metodo][rota] = handler
}

func (s *Servidor) M__obtem_attributo__(nome string) (ptst.Objeto, error) {
	switch nome {
	case "obter":
		return ptst.NewMetodoOuPanic("obter", func(inst ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
			if err := ptst.VerificaNumeroArgumentos("obter", false, args, 2, 2); err != nil {
				return nil, err
			}
			rota, err := ptst.NewTexto(args[0])
			if err != nil {
				return nil, err
			}
			s.registrarRota("GET", string(rota.(ptst.Texto)), args[1])
			return ptst.Nulo, nil
		}, ""), nil

	case "postar":
		return ptst.NewMetodoOuPanic("postar", func(inst ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
			if err := ptst.VerificaNumeroArgumentos("postar", false, args, 2, 2); err != nil {
				return nil, err
			}
			rota, err := ptst.NewTexto(args[0])
			if err != nil {
				return nil, err
			}
			s.registrarRota("POST", string(rota.(ptst.Texto)), args[1])
			return ptst.Nulo, nil
		}, ""), nil

	case "deletar":
		return ptst.NewMetodoOuPanic("deletar", func(inst ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
			if err := ptst.VerificaNumeroArgumentos("deletar", false, args, 2, 2); err != nil {
				return nil, err
			}
			rota, err := ptst.NewTexto(args[0])
			if err != nil {
				return nil, err
			}
			s.registrarRota("DELETE", string(rota.(ptst.Texto)), args[1])
			return ptst.Nulo, nil
		}, ""), nil

	case "usar":
		return ptst.NewMetodoOuPanic("usar", func(inst ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
			if err := ptst.VerificaNumeroArgumentos("usar", false, args, 1, 1); err != nil {
				return nil, err
			}
			s.mu.Lock()
			s.middlewares = append(s.middlewares, args[0])
			s.mu.Unlock()
			return ptst.Nulo, nil
		}, ""), nil

	case "servir_app":
		return ptst.NewMetodoOuPanic("servir_app", func(inst ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
			if err := ptst.VerificaNumeroArgumentos("servir_app", false, args, 2, 3); err != nil {
				return nil, err
			}
			diretorioDist, err := ptst.NewTexto(args[0])
			if err != nil {
				return nil, err
			}
			componenteRaiz := args[1]
			var metadados ptst.Objeto = ptst.Nulo
			if len(args) == 3 {
				metadados = args[2]
			}

			handler := ServirAppHandler(string(diretorioDist.(ptst.Texto)), componenteRaiz, metadados)
			s.registrarRota("GET", "/*", handler)
			return ptst.Nulo, nil
		}, ""), nil

	case "escutar":
		return ptst.NewMetodoOuPanic("escutar", func(inst ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
			if ptst.ContextoAtivo != nil {
				if err := ptst.ContextoAtivo.VerificarPermissaoRede(); err != nil {
					return nil, err
				}
			}

			if err := ptst.VerificaNumeroArgumentos("escutar", false, args, 1, 1); err != nil {
				return nil, err
			}
			porta, err := ptst.NewTexto(args[0])
			if err != nil {
				return nil, err
			}

			mux := http.NewServeMux()
			mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
				defer func() {
					if r := recover(); r != nil {
						w.WriteHeader(http.StatusInternalServerError)
						w.Write(fmt.Appendf(nil, "Erro interno do servidor (Pânico): %v", r))
					}
				}()

				s.mu.RLock()
				rotasMetodo := s.rotas[req.Method]
				s.mu.RUnlock()

				var handler ptst.Objeto
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
				reqHeaders := ptst.NewMapaVazio()
				for k, vals := range req.Header {
					reqHeaders.M__define_item__(ptst.Texto(k), ptst.Texto(strings.Join(vals, ", ")))
				}

				bodyBytes, _ := io.ReadAll(req.Body)
				// Se houver parâmetros de rota (:id), mescla no objeto de requisição adicionando dinamicamente
				reqParams := ptst.NewMapaVazio()
				for k, v := range params {
					reqParams.M__define_item__(ptst.Texto(k), ptst.Texto(v))
				}

				reqObj := &Requisicao{
					Metodo:     ptst.Texto(req.Method),
					Caminho:    ptst.Texto(req.URL.Path),
					Cabecalho:  reqHeaders,
					Corpo:      ptst.Texto(bodyBytes),
					Parametros: reqParams,
				}

				resObj := &Resposta{
					Status:    ptst.Inteiro(http.StatusOK),
					Cabecalho: ptst.NewMapaVazio(),
					Corpo:     ptst.Texto(""),
				}

				// Roda middlewares
				s.mu.RLock()
				mws := make([]ptst.Objeto, len(s.middlewares))
				copy(mws, s.middlewares)
				s.mu.RUnlock()

				for _, mw := range mws {
					_, errMw := ptst.Chamar(mw, ptst.Tupla{reqObj, resObj})
					if errMw != nil {
						w.WriteHeader(http.StatusInternalServerError)
						w.Write([]byte(errMw.Error()))
						return
					}
				}

				// Roda o handler principal
				_, errHandler := ptst.Chamar(handler, ptst.Tupla{reqObj, resObj})
				if errHandler != nil {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(errHandler.Error()))
					return
				}

				// Escreve cabecalhos de resposta
				for k, v := range resObj.Cabecalho {
					w.Header().Set(k, string(v.(ptst.Texto)))
				}

				w.WriteHeader(int(resObj.Status))
				w.Write([]byte(resObj.Corpo))
			})

			s.server = &http.Server{
				Addr:         ":" + string(porta.(ptst.Texto)),
				Handler:      mux,
				ReadTimeout:  5 * time.Second,
				WriteTimeout: 10 * time.Second,
				IdleTimeout:  120 * time.Second,
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
				return nil, ptst.NewErroF(ptst.ErroDeSistema, "Erro ao escutar porta: %v", errListen)
			}

			return ptst.Nulo, nil
		}, ""), nil

	case "fechar":
		return ptst.NewMetodoOuPanic("fechar", func(inst ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
			if s.server != nil {
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()
				s.server.Shutdown(ctx)
			}
			return ptst.Nulo, nil
		}, ""), nil
	}

	return nil, ptst.NewErroF(ptst.AtributoErro, "Atributo '%s' não existe no Servidor", nome)
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
func met_http_requisitar(inst ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
	if ptst.ContextoAtivo != nil {
		if err := ptst.ContextoAtivo.VerificarPermissaoRede(); err != nil {
			return nil, err
		}
	}

	if err := ptst.VerificaNumeroArgumentos("requisitar", false, args, 2, 4); err != nil {
		return nil, err
	}

	metodo, err := ptst.NewTexto(args[0])
	if err != nil {
		return nil, err
	}

	url, err := ptst.NewTexto(args[1])
	if err != nil {
		return nil, err
	}

	var reqBody io.Reader
	if len(args) >= 3 && args[2] != ptst.Nulo {
		corpoText, err := ptst.NewTexto(args[2])
		if err != nil {
			return nil, err
		}
		reqBody = strings.NewReader(string(corpoText.(ptst.Texto)))
	}

	req, err := http.NewRequest(string(metodo.(ptst.Texto)), string(url.(ptst.Texto)), reqBody)
	if err != nil {
		return nil, ptst.NewErroF(ptst.ErroDeSistema, "Erro ao criar requisição cliente: %v", err)
	}

	if len(args) >= 4 && args[3] != ptst.Nulo {
		if cabeçalhos, ok := args[3].(ptst.Mapa); ok {
			for k, v := range cabeçalhos {
				vTexto, _ := ptst.NewTexto(v)
				req.Header.Set(k, string(vTexto.(ptst.Texto)))
			}
		}
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, ptst.NewErroF(ptst.ErroDeSistema, "Erro ao executar requisição HTTP: %v", err)
	}
	defer res.Body.Close()

	bodyBytes, _ := io.ReadAll(res.Body)
	resHeaders := ptst.NewMapaVazio()
	for k, vals := range res.Header {
		resHeaders.M__define_item__(ptst.Texto(k), ptst.Texto(strings.Join(vals, ", ")))
	}

	resObj := &Resposta{
		Status:    ptst.Inteiro(res.StatusCode),
		Cabecalho: resHeaders,
		Corpo:     ptst.Texto(bodyBytes),
	}
	return resObj, nil
}

func init() {
	ptst.RegistraModuloImpl(&ptst.ModuloImpl{
		Info: ptst.ModuloInfo{
			Nome:    "http",
			Arquivo: "stdlib/http",
		},
		Constantes: ptst.Mapa{
			"Servidor": TipoServidor,
		},
		Metodos: []*ptst.Metodo{
			ptst.NewMetodoOuPanic("requisitar", met_http_requisitar, ""),
		},
	})
}
