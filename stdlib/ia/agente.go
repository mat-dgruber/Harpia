// Package ia implementa as facilidades de integração com modelos de inteligência artificial generativa,
// orquestração de agentes de IA locais (Ollama), suporte a prompt engineering e validações estritas de esquemas JSON.
package ia

import (
	"fmt"

	"github.com/mat-dgruber/Harpia/hrp"
)

// Agente representa uma entidade autônoma reativa inteligente que encapsula instruções de sistema,
// histórico persistente de conversas e regras de comunicação direta com provedores de LLM.
type Agente struct {
	Nome       hrp.Texto
	Instrucoes hrp.Texto
	Provedor   hrp.Texto
	Modelo     hrp.Texto
	Historico  *hrp.Lista
}

// TipoAgente define e registra a classe Agente na VM.
var TipoAgente = hrp.NewTipo("Agente", "Tipo nativo para criação de agentes autônomos inteligentes")

// Tipo retorna o tipo da classe na VM.
func (a *Agente) Tipo() *hrp.Tipo {
	return TipoAgente
}

func init() {
	// Inicializador estático do tipo Agente, gerenciando parâmetros opcionais para provedor e modelo de LLM.
	TipoAgente.Nova = func(args hrp.Tupla) (hrp.Objeto, error) {
		if err := hrp.VerificaNumeroArgumentos("Agente", false, args, 2, 4); err != nil {
			return nil, err
		}

		nome, err := hrp.NewTexto(args[0])
		if err != nil {
			return nil, err
		}

		instrucoes, err := hrp.NewTexto(args[1])
		if err != nil {
			return nil, err
		}

		provedor := hrp.Texto("ollama")
		if len(args) >= 3 && args[2] != hrp.Nulo {
			prov, err := hrp.NewTexto(args[2])
			if err != nil {
				return nil, err
			}
			provedor = prov.(hrp.Texto)
		}

		modelo := hrp.Texto("llama3")
		if len(args) >= 4 && args[3] != hrp.Nulo {
			mod, err := hrp.NewTexto(args[3])
			if err != nil {
				return nil, err
			}
			modelo = mod.(hrp.Texto)
		}

		historico := &hrp.Lista{Itens: hrp.Tupla{}}

		return &Agente{
			Nome:       nome.(hrp.Texto),
			Instrucoes: instrucoes.(hrp.Texto),
			Provedor:   provedor,
			Modelo:     modelo,
			Historico:  historico,
		}, nil
	}
}

// Perguntar interage diretamente com o LLM provido através de um canal síncrono, gerenciando e
// persistindo o histórico de mensagens em memória sob o formato de mensagens de Chat (User/Assistant).
func (a *Agente) Perguntar(mensagem string) (string, error) {
	// Adiciona a mensagem do usuário ao histórico local
	userMsg := hrp.NewMapaVazio()
	userMsg.M__define_item__(hrp.Texto("role"), hrp.Texto("user"))
	userMsg.M__define_item__(hrp.Texto("content"), hrp.Texto(mensagem))
	a.Historico.Adiciona(userMsg)

	// Converte histórico local da VM para o formato go do LLM de destino
	var msgs []Mensagem
	for i := 0; i < len(a.Historico.Itens); i++ {
		item := a.Historico.Itens[i]
		if mapaItem, ok := item.(hrp.Mapa); ok {
			roleObj, _ := mapaItem.M__obtem_item__(hrp.Texto("role"))
			contentObj, _ := mapaItem.M__obtem_item__(hrp.Texto("content"))
			msgs = append(msgs, Mensagem{
				Role:    string(roleObj.(hrp.Texto)),
				Content: string(contentObj.(hrp.Texto)),
			})
		}
	}

	resposta, err := ChamarLLM(string(a.Provedor), string(a.Modelo), string(a.Instrucoes), msgs)
	if err != nil {
		return "", err
	}

	// Adiciona a resposta gerada pelo assistente ao histórico local para manutenção de contexto
	assistantMsg := hrp.NewMapaVazio()
	assistantMsg.M__define_item__(hrp.Texto("role"), hrp.Texto("assistant"))
	assistantMsg.M__define_item__(hrp.Texto("content"), hrp.Texto(resposta))
	a.Historico.Adiciona(assistantMsg)

	return resposta, nil
}

// M__obtem_attributo__ mapeia as propriedades e métodos do Agente (historico, perguntar, comunicar, limpar_memoria).
func (a *Agente) M__obtem_attributo__(nome string) (hrp.Objeto, error) {
	switch nome {
	case "nome":
		return a.Nome, nil
	case "instrucoes":
		return a.Instrucoes, nil
	case "provedor":
		return a.Provedor, nil
	case "modelo":
		return a.Modelo, nil
	case "historico":
		return a.Historico, nil
	case "perguntar":
		// Faz uma pergunta ao agente, que responde baseado em seu histórico e personalidade.
		return hrp.NewMetodoOuPanic("perguntar", func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
			if err := hrp.VerificaNumeroArgumentos("perguntar", false, args, 1, 1); err != nil {
				return nil, err
			}
			msgTexto, err := hrp.NewTexto(args[0])
			if err != nil {
				return nil, err
			}
			resposta, err := a.Perguntar(string(msgTexto.(hrp.Texto)))
			if err != nil {
				return nil, hrp.NewErroF(hrp.ErroDeSistema, "Erro ao interagir com IA: %v", err)
			}
			return hrp.Texto(resposta), nil
		}, "Envia um prompt para o agente, guardando o histórico de conversas em memória."), nil
	case "limpar_memoria":
		// Zera todo o histórico de conversas acumulado na memória do agente.
		return hrp.NewMetodoOuPanic("limpar_memoria", func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
			a.Historico.Itens = hrp.Tupla(nil)
			return hrp.Nulo, nil
		}, "Apaga o histórico de conversação do agente."), nil
	case "comunicar":
		// Permite a orquestração multi-agente, onde dois agentes trocam prompts e sintetizam respostas cooperativas.
		return hrp.NewMetodoOuPanic("comunicar", func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
			if err := hrp.VerificaNumeroArgumentos("comunicar", false, args, 2, 2); err != nil {
				return nil, err
			}
			outroAgente, ok := args[0].(*Agente)
			if !ok {
				return nil, hrp.NewErroF(hrp.TipagemErro, "esperado um objeto Agente para comunicação")
			}
			mensagem, err := hrp.NewTexto(args[1])
			if err != nil {
				return nil, err
			}

			// Envia prompt ao outro agente
			respOutro, err := outroAgente.Perguntar(string(mensagem.(hrp.Texto)))
			if err != nil {
				return nil, hrp.NewErroF(hrp.ErroDeSistema, "Erro ao comunicar com outro agente: %v", err)
			}

			// Registra a resposta na nossa própria memória como input e gera a síntese/conclusão final
			minhaMsg := fmt.Sprintf("Agente %s respondeu: %s", outroAgente.Nome, respOutro)
			minhaResp, err := a.Perguntar(minhaMsg)
			if err != nil {
				return nil, hrp.NewErroF(hrp.ErroDeSistema, "Erro ao processar resposta do outro agente: %v", err)
			}

			return hrp.Texto(minhaResp), nil
		}, "Estabelece comunicação bidirecional entre dois agentes locais."), nil
	}
	return nil, hrp.NewErroF(hrp.AtributoErro, "Atributo '%s' não existe em Agente", nome)
}

var _ hrp.I__obtem_attributo__ = (*Agente)(nil)
