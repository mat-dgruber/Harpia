package ia

import (
	"github.com/mat-dgruber/Harpia/hrp"
)

func init() {
	// Registra o módulo 'ia' na tabela interna de módulos carregáveis da VM.
	hrp.RegistraModuloImpl(
		&hrp.ModuloImpl{
			Info: hrp.ModuloInfo{
				Nome:    "ia",
				Arquivo: "stdlib/ia",
			},
			Constantes: hrp.Mapa{
				"Agente": TipoAgente,
			},
			Metodos: []*hrp.Metodo{
				hrp.NewMetodoOuPanic("validar_resposta", met_validar_resposta, "Valida resposta JSON contra esquema"),
				_ia_sigmoide,
				_ia_derivadaSigmoide,
				_ia_relu,
				_ia_softmax,
				_ia_similaridadeCosseno,
				_ia_normalizarMinMax,
				_ia_erroQuadraticoMedio,
			},
		},
	)
}
