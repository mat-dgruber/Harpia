package tempo

import (
	"fmt"
	"strings"
	"time"

	"github.com/mat-dgruber/Harpia/hrp"
)

type DataHora struct {
	t time.Time
}

var TipoDataHora = hrp.NewTipo("DataHora", "Representa uma data e hora no Harpia")

func (d *DataHora) Tipo() *hrp.Tipo {
	return TipoDataHora
}

func (d *DataHora) M__obtem_attributo__(nome string) (hrp.Objeto, error) {
	switch nome {
	case "ano":
		return hrp.Inteiro(d.t.Year()), nil
	case "mes":
		return hrp.Inteiro(d.t.Month()), nil
	case "dia":
		return hrp.Inteiro(d.t.Day()), nil
	case "hora":
		return hrp.Inteiro(d.t.Hour()), nil
	case "minuto":
		return hrp.Inteiro(d.t.Minute()), nil
	case "segundo":
		return hrp.Inteiro(d.t.Second()), nil
	case "formatar":
		return hrp.NewMetodoOuPanic("formatar", func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
			formato := "02/01/2006 15:04:05"
			if len(args) > 0 {
				if f, err := hrp.NewTexto(args[0]); err == nil {
					formatoStr := string(f.(hrp.Texto))
					// Suporte amigável a tokens comuns (YYYY, MM, DD, HH, mm, ss)
					formatoStr = strings.ReplaceAll(formatoStr, "YYYY", "2006")
					formatoStr = strings.ReplaceAll(formatoStr, "MM", "01")
					formatoStr = strings.ReplaceAll(formatoStr, "DD", "02")
					formatoStr = strings.ReplaceAll(formatoStr, "HH", "15")
					formatoStr = strings.ReplaceAll(formatoStr, "mm", "04")
					formatoStr = strings.ReplaceAll(formatoStr, "ss", "05")
					formato = formatoStr
				}
			}
			return hrp.Texto(d.t.Format(formato)), nil
		}, "Formata a data/hora segundo o padrão especificado ou padrão 'DD/MM/YYYY HH:mm:ss'"), nil
	case "adicionarDias":
		return hrp.NewMetodoOuPanic("adicionarDias", func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
			if len(args) == 0 {
				return nil, fmt.Errorf("adicionarDias requer o número de dias")
			}
			diasInt, err := hrp.NewInteiro(args[0])
			if err != nil {
				return nil, err
			}
			dias := int(diasInt.(hrp.Inteiro))
			return &DataHora{t: d.t.AddDate(0, 0, dias)}, nil
		}, ""), nil
	}
	return nil, hrp.NewErroF(hrp.AtributoErro, "Atributo '%s' não existe em DataHora", nome)
}

func met_tempo_agora(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	return &DataHora{t: time.Now()}, nil
}

var _agora = hrp.NewMetodoOuPanic("agora", met_tempo_agora, "")

func init() {
	hrp.RegistraModuloImpl(&hrp.ModuloImpl{
		Info: hrp.ModuloInfo{
			Nome:    "tempo",
			Arquivo: "stdlib/tempo",
		},
		Metodos: []*hrp.Metodo{
			_agora,
		},
	})
}
