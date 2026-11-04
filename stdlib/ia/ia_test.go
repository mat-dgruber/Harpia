package ia

import (
	"testing"

	"github.com/mat-dgruber/Harpia/hrp"
)

func TestValidarResposta(t *testing.T) {
	esquema := hrp.Mapa{
		"nome":  hrp.Texto("texto"),
		"idade": hrp.Texto("inteiro"),
	}

	jsonValido := `{"nome": "Harpia", "idade": 2}`
	valido, err := ValidarResposta(esquema, jsonValido)
	if err != nil || !valido {
		t.Fatalf("Esperava JSON valido, obteve erro: %v", err)
	}

	jsonInvalido := `{"nome": "Harpia", "idade": "dois"}`
	valido, err = ValidarResposta(esquema, jsonInvalido)
	if err == nil || valido {
		t.Fatalf("Esperava falha na validacao por tipo incorreto")
	}

	// Regra de esquema que não seja hrp.Texto deve retornar erro
	esquemaInvalido := hrp.Mapa{
		"nome": hrp.Inteiro(123),
	}
	_, err = ValidarResposta(esquemaInvalido, `{"nome": "teste"}`)
	if err == nil {
		t.Fatalf("Esperava erro para regra de esquema que nao seja Texto")
	}
}

func TestAgenteCriacao(t *testing.T) {
	args := hrp.Tupla{
		hrp.Texto("Assistente"),
		hrp.Texto("Voce e um assistente util"),
		hrp.Texto("ollama"),
		hrp.Texto("llama3"),
	}
	obj, err := TipoAgente.Nova(args)
	if err != nil {
		t.Fatalf("Erro ao criar Agente: %v", err)
	}

	agente := obj.(*Agente)
	if string(agente.Nome) != "Assistente" || string(agente.Modelo) != "llama3" {
		t.Errorf("Atributos do Agente incorretos: %+v", agente)
	}
}
