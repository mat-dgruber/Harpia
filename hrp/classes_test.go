package hrp

import (
	"testing"
)

func TestClassesEHeranca(t *testing.T) {
	codigo := `
	classe Animal {
		func inicializar(self, nome) {
			self.nome = nome
		}

		func falar(self) {
			retorne "Som de animal: " + self.nome
		}
	}

	classe Cachorro estende Animal {
		func inicializar(self, nome, raca) {
			self.nome = nome
			self.raca = raca
		}

		func falar(self) {
			retorne "Au! Meu nome é " + self.nome
		}
	}

	var cao = nova Cachorro("Rex", "Pastor")
	var animal = nova Animal("Generico")

	var caoNome = cao.nome
	var caoFalar = cao.falar()
	var animalFalar = animal.falar()

	var caoEhCachorro = cao instancia de Cachorro
	var caoEhAnimal = cao instancia de Animal
	var animalEhCachorro = animal instancia de Cachorro
	`

	ctx := NewContexto(OpcsContexto{})
	defer ctx.Terminar()

	_, err := ExecutarString(ctx, codigo)
	if err != nil {
		t.Fatalf("Erro inesperado ao executar código de classes: %v", err)
	}

	// Verifica se os valores foram criados corretamente no escopo global
	modulo, err := ctx.ObterModulo("__entrada__")
	if err != nil {
		t.Fatal(err)
	}

	caoNome, err := modulo.Escopo.ObterValor("caoNome")
	if err != nil || string(caoNome.(Texto)) != "Rex" {
		t.Errorf("Esperava cao.nome ser 'Rex', obteve: %v (erro: %v)", caoNome, err)
	}

	caoFalar, err := modulo.Escopo.ObterValor("caoFalar")
	if err != nil || string(caoFalar.(Texto)) != "Au! Meu nome é Rex" {
		t.Errorf("Esperava cao.falar() ser 'Au! Meu nome é Rex', obteve: %v", caoFalar)
	}

	animalFalar, err := modulo.Escopo.ObterValor("animalFalar")
	if err != nil || string(animalFalar.(Texto)) != "Som de animal: Generico" {
		t.Errorf("Esperava animal.falar() ser 'Som de animal: Generico', obteve: %v", animalFalar)
	}

	caoEhCachorro, err := modulo.Escopo.ObterValor("caoEhCachorro")
	if err != nil || caoEhCachorro.(Booleano) != Verdadeiro {
		t.Errorf("Esperava cao instancia de Cachorro ser Verdadeiro, obteve: %v", caoEhCachorro)
	}

	caoEhAnimal, err := modulo.Escopo.ObterValor("caoEhAnimal")
	if err != nil || caoEhAnimal.(Booleano) != Verdadeiro {
		t.Errorf("Esperava cao instancia de Animal ser Verdadeiro, obteve: %v", caoEhAnimal)
	}

	animalEhCachorro, err := modulo.Escopo.ObterValor("animalEhCachorro")
	if err != nil || animalEhCachorro.(Booleano) != Falso {
		t.Errorf("Esperava animal instancia de Cachorro ser Falso, obteve: %v", animalEhCachorro)
	}
}
