package vetores

import (
	"math"
	"sort"

	"github.com/mat-dgruber/Harpia/hrp"
)

type ItemVetorial struct {
	ID        string
	Vetor     []float64
	Metadados hrp.Mapa
}

type BancoVetorialObj struct {
	hrp.GCMixin
	Itens []*ItemVetorial
}

var TipoBancoVetorial = hrp.TipoObjeto.NewTipo(
	"BancoVetorial",
	"BancoVetorial() -> Instância de banco de dados de vetores (Vector Store)",
)

func (b *BancoVetorialObj) Tipo() *hrp.Tipo {
	return TipoBancoVetorial
}

func (b *BancoVetorialObj) ObterFilhos() []hrp.Objeto {
	return nil
}

func init() {
	metodos := []*hrp.Metodo{
		_vetores_criarBanco,
	}

	hrp.RegistraModuloImpl(
		&hrp.ModuloImpl{
			Info: hrp.ModuloInfo{
				Nome:    "vetores",
				Arquivo: "stdlib/bd/vetores",
			},
			Constantes: hrp.Mapa{},
			Metodos:    metodos,
		},
	)

	TipoBancoVetorial.Mapa["inserir"] = hrp.NewMetodoOuPanic("inserir", func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
		if err := hrp.VerificaNumeroArgumentos("inserir", true, args, 2, 3); err != nil {
			return nil, err
		}
		banco := inst.(*BancoVetorialObj)
		id := string(args[0].(hrp.Texto))

		listaVetor, ok := args[1].(*hrp.Lista)
		if !ok {
			return nil, hrp.NewErroF(hrp.TipagemErro, "vetor inserido deve ser uma Lista de numeros")
		}

		vFloats := make([]float64, len(listaVetor.Itens))
		for i, elem := range listaVetor.Itens {
			dec, err := hrp.NewDecimal(elem)
			if err != nil {
				return nil, err
			}
			vFloats[i] = float64(dec.(hrp.Decimal))
		}

		metadados := hrp.Mapa{}
		if len(args) > 2 {
			if m, ok := args[2].(hrp.Mapa); ok {
				metadados = m
			}
		}

		banco.Itens = append(banco.Itens, &ItemVetorial{
			ID:        id,
			Vetor:     vFloats,
			Metadados: metadados,
		})
		return hrp.Verdadeiro, nil
	}, "inserir(id, vetor, [metadados]) -> Insere um item vetorial no banco")

	TipoBancoVetorial.Mapa["buscarMaisProximos"] = hrp.NewMetodoOuPanic("buscarMaisProximos", func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
		if err := hrp.VerificaNumeroArgumentos("buscarMaisProximos", true, args, 1, 2); err != nil {
			return nil, err
		}
		banco := inst.(*BancoVetorialObj)
		listaVetor, ok := args[0].(*hrp.Lista)
		if !ok {
			return nil, hrp.NewErroF(hrp.TipagemErro, "vetor de consulta deve ser uma Lista de numeros")
		}

		vConsulta := make([]float64, len(listaVetor.Itens))
		for i, elem := range listaVetor.Itens {
			dec, err := hrp.NewDecimal(elem)
			if err != nil {
				return nil, err
			}
			vConsulta[i] = float64(dec.(hrp.Decimal))
		}

		k := 5
		if len(args) > 1 {
			kDec, err := hrp.NewDecimal(args[1])
			if err == nil {
				k = int(kDec.(hrp.Decimal))
			}
		}

		type ResultadoBusca struct {
			Item  *ItemVetorial
			Score float64
		}

		resultados := make([]ResultadoBusca, 0, len(banco.Itens))
		for _, item := range banco.Itens {
			score := calcularSimilaridadeCosseno(vConsulta, item.Vetor)
			resultados = append(resultados, ResultadoBusca{Item: item, Score: score})
		}

		sort.Slice(resultados, func(i, j int) bool {
			return resultados[i].Score > resultados[j].Score
		})

		if len(resultados) > k {
			resultados = resultados[:k]
		}

		resLista := make([]hrp.Objeto, len(resultados))
		for i, r := range resultados {
			resLista[i] = hrp.Mapa{
				"id":        hrp.Texto(r.Item.ID),
				"score":     hrp.Decimal(r.Score),
				"metadados": r.Item.Metadados,
			}
		}
		return &hrp.Lista{Itens: hrp.Tupla(resLista)}, nil
	}, "buscarMaisProximos(vetorConsulta, [k]) -> Retorna os k vetores mais parecidos")
}

func calcularSimilaridadeCosseno(v1, v2 []float64) float64 {
	if len(v1) != len(v2) || len(v1) == 0 {
		return 0.0
	}
	dot := 0.0
	n1 := 0.0
	n2 := 0.0
	for i := 0; i < len(v1); i++ {
		dot += v1[i] * v2[i]
		n1 += v1[i] * v1[i]
		n2 += v2[i] * v2[i]
	}
	if n1 == 0.0 || n2 == 0.0 {
		return 0.0
	}
	return dot / (math.Sqrt(n1) * math.Sqrt(n2))
}

func met_vetores_criarBanco(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	return &BancoVetorialObj{
		Itens: []*ItemVetorial{},
	}, nil
}

var _vetores_criarBanco = hrp.NewMetodoOuPanic("criarBanco", met_vetores_criarBanco, "criarBanco() -> Instancia um novo BancoVetorial")
