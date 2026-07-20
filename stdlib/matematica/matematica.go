// Package matematica implementa o módulo nativo de funções matemáticas e constantes de alta precisão
// da biblioteca padrão do Harpia.
//
// Este pacote faz a ponte entre a biblioteca de matemática padrão do Go (math) e o interpretador,
// expondo constantes como PI e E, além de funções para cálculo de potência, raízes e arredondamentos.
package matematica

import (
	"math"

	"github.com/mat-dgruber/Harpia/hrp"
)

func init() {
	// constantes define as propriedades estáticas imutáveis expostas pelo módulo matemática.
	constantes := hrp.Mapa{
		"PI": hrp.Decimal(math.Pi), // Representação aproximada do número Pi (3.14159...)
		"E":  hrp.Decimal(math.E),  // Representação aproximada da constante de Euler (2.71828...)
	}

	// metodos é a relação de ponteiros de funções associadas que são registradas no escopo do módulo.
	metodos := []*hrp.Metodo{
		_mat_raiz,
		_mat_potencia,
		_mat_absoluto,
		_mat_piso,
		_mat_teto,
	}

	// Registra o módulo de matemática na tabela interna de módulos carregáveis da VM.
	// Qualquer script que execute 'importar matematica' receberá acesso a essas chaves.
	hrp.RegistraModuloImpl(
		&hrp.ModuloImpl{
			Info: hrp.ModuloInfo{
				Nome:    "matematica",
				Arquivo: "stdlib/matematica",
			},
			Constantes: constantes,
			Metodos:    metodos,
		},
	)
}
