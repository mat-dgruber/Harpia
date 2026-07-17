package ia

import (
	"github.com/mat-dgruber/Harpia/ptst"
)

func init() {
	// Registra o módulo 'ia' na tabela interna de módulos carregáveis da VM.
	ptst.RegistraModuloImpl(
		&ptst.ModuloImpl{
			Info: ptst.ModuloInfo{
				Nome:    "ia",
				Arquivo: "stdlib/ia",
			},
			Constantes: ptst.Mapa{
				"Agente": TipoAgente,
			},
			Metodos: []*ptst.Metodo{},
		},
	)
}
