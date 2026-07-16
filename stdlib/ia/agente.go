package ia

import (
	"fmt"

	"github.com/natanfeitosa/portuscript/ptst"
)

type Agente struct {
	Nome       ptst.Texto
	Instrucoes ptst.Texto
	Provedor   ptst.Texto
	Modelo     ptst.Texto
	Historico  *ptst.Lista
}

var TipoAgente = ptst.NewTipo("Agente", "Tipo nativo para criação de agentes autônomos inteligentes")

func (a *Agente) Tipo() *ptst.Tipo {
	return TipoAgente
}

func init() {
	TipoAgente.Nova = func(args ptst.Tupla) (ptst.Objeto, error) {
		if err := ptst.VerificaNumeroArgumentos("Agente", false, args, 2, 4); err != nil {
			return nil, err
		}

		nome, err := ptst.NewTexto(args[0])
		if err != nil {
			return nil, err
		}

		instrucoes, err := ptst.NewTexto(args[1])
		if err != nil {
			return nil, err
		}

		provedor := ptst.Texto("ollama")
		if len(args) >= 3 && args[2] != ptst.Nulo {
			prov, err := ptst.NewTexto(args[2])
			if err != nil {
				return nil, err
			}
			provedor = prov.(ptst.Texto)
		}

		modelo := ptst.Texto("llama3")
		if len(args) >= 4 && args[3] != ptst.Nulo {
			mod, err := ptst.NewTexto(args[3])
			if err != nil {
				return nil, err
			}
			modelo = mod.(ptst.Texto)
		}

		historico := &ptst.Lista{Itens: ptst.Tupla{}}

		return &Agente{
			Nome:       nome.(ptst.Texto),
			Instrucoes: instrucoes.(ptst.Texto),
			Provedor:   provedor,
			Modelo:     modelo,
			Historico:  historico,
		}, nil
	}
}

func (a *Agente) Perguntar(mensagem string) (string, error) {
	// Adiciona a mensagem do usuário ao histórico
	userMsg := ptst.NewMapaVazio()
	userMsg.M__define_item__(ptst.Texto("role"), ptst.Texto("user"))
	userMsg.M__define_item__(ptst.Texto("content"), ptst.Texto(mensagem))
	a.Historico.Adiciona(userMsg)

	// Converte histórico local da VM para o formato go do LLM
	var msgs []Mensagem
	for i := 0; i < len(a.Historico.Itens); i++ {
		item := a.Historico.Itens[i]
		if mapaItem, ok := item.(ptst.Mapa); ok {
			roleObj, _ := mapaItem.M__obtem_item__(ptst.Texto("role"))
			contentObj, _ := mapaItem.M__obtem_item__(ptst.Texto("content"))
			msgs = append(msgs, Mensagem{
				Role:    string(roleObj.(ptst.Texto)),
				Content: string(contentObj.(ptst.Texto)),
			})
		}
	}

	resposta, err := ChamarLLM(string(a.Provedor), string(a.Modelo), string(a.Instrucoes), msgs)
	if err != nil {
		return "", err
	}

	// Adiciona a resposta do assistente ao histórico
	assistantMsg := ptst.NewMapaVazio()
	assistantMsg.M__define_item__(ptst.Texto("role"), ptst.Texto("assistant"))
	assistantMsg.M__define_item__(ptst.Texto("content"), ptst.Texto(resposta))
	a.Historico.Adiciona(assistantMsg)

	return resposta, nil
}

func (a *Agente) M__obtem_attributo__(nome string) (ptst.Objeto, error) {
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
		return ptst.NewMetodoOuPanic("perguntar", func(inst ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
			if err := ptst.VerificaNumeroArgumentos("perguntar", false, args, 1, 1); err != nil {
				return nil, err
			}
			msgTexto, err := ptst.NewTexto(args[0])
			if err != nil {
				return nil, err
			}
			resposta, err := a.Perguntar(string(msgTexto.(ptst.Texto)))
			if err != nil {
				return nil, ptst.NewErroF(ptst.ErroDeSistema, "Erro ao interagir com IA: %v", err)
			}
			return ptst.Texto(resposta), nil
		}, ""), nil
	case "limpar_memoria":
		return ptst.NewMetodoOuPanic("limpar_memoria", func(inst ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
			a.Historico.Itens = ptst.Tupla(nil)
			return ptst.Nulo, nil
		}, ""), nil
	case "comunicar":
		return ptst.NewMetodoOuPanic("comunicar", func(inst ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
			if err := ptst.VerificaNumeroArgumentos("comunicar", false, args, 2, 2); err != nil {
				return nil, err
			}
			outroAgente, ok := args[0].(*Agente)
			if !ok {
				return nil, ptst.NewErroF(ptst.TipagemErro, "esperado um objeto Agente para comunicação")
			}
			mensagem, err := ptst.NewTexto(args[1])
			if err != nil {
				return nil, err
			}

			// Pergunta ao outro agente
			respOutro, err := outroAgente.Perguntar(string(mensagem.(ptst.Texto)))
			if err != nil {
				return nil, ptst.NewErroF(ptst.ErroDeSistema, "Erro ao comunicar com outro agente: %v", err)
			}

			// Registra a resposta na nossa própria memória como input de usuário
			minhaMsg := fmt.Sprintf("Agente %s respondeu: %s", outroAgente.Nome, respOutro)
			minhaResp, err := a.Perguntar(minhaMsg)
			if err != nil {
				return nil, ptst.NewErroF(ptst.ErroDeSistema, "Erro ao processar resposta do outro agente: %v", err)
			}

			return ptst.Texto(minhaResp), nil
		}, ""), nil
	}
	return nil, ptst.NewErroF(ptst.AtributoErro, "Atributo '%s' não existe em Agente", nome)
}

var _ ptst.I__obtem_attributo__ = (*Agente)(nil)
