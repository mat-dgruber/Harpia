//go:build windows

package soquete

import (
	"github.com/mat-dgruber/Harpia/hrp"
)

type Soquete struct {
	descritorDoSoquete       int
	familia, tipo, protocolo hrp.Inteiro
	fechado                  hrp.Booleano
}

var TipoSoquete = hrp.TipoObjeto.NewTipo(
	"Soquete",
	`Soquete(familia, tipo) -> Soquete
Cria um novo soquete usando a família de endereços, o tipo de soquete e o número de protocolo fornecidos.`,
)

func (s *Soquete) Tipo() *hrp.Tipo {
	return TipoSoquete
}

func init() {
	TipoSoquete.Nova = func(args hrp.Tupla) (hrp.Objeto, error) {
		return nil, hrp.NewErroF(hrp.ErroDeSistema, "o módulo 'soquete' não é suportado nativamente no Windows")
	}
}
