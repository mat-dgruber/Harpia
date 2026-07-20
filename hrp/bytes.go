package hrp

import (
	"bytes"
)

// Bytes representa o tipo de dado de torrente e array de bytes raw mutável (ou congelado) do Harpia.
type Bytes struct {
	Itens     []byte // O slice standard de bytes do Go que guarda os dados físicos.
	Congelado bool   // Flag opcional que sinaliza se o array de bytes está bloqueado contra escritas.
}

// TipoBytes especifica as assinaturas e metadados de classe do tipo Bytes na VM.
var TipoBytes = TipoObjeto.NewTipo(
	"Bytes",
	"Bytes(obj) -> Bytes",
)

// NewBytes tenta realizar o casting coerção de qualquer tipo primitivo compatível em um objeto Bytes.
//
// Suporta:
//   - nil ➔ Retorna array de bytes vazio.
//   - string ➔ Retorna array populado com os bytes correspondentes da string.
//   - Objetos Customizados ➔ Verifica e chama o método nativo '__bytes__' (M__bytes__()).
func NewBytes(arg any) (Objeto, error) {
	switch obj := arg.(type) {
	case nil:
		return &Bytes{make([]byte, 0), false}, nil
	case string:
		return &Bytes{[]byte(obj), false}, nil
	case *Bytes:
		return obj, nil
	}

	if met, _ := ObtemAtributoS(arg.(Objeto), "__bytes__"); met != nil {
		return Chamar(met, Tupla{})
	}

	if O, ok := arg.(I__bytes__); ok {
		return O.M__bytes__()
	}

	return nil, nil
}

func init() {
	// Nova define o construtor do Bytes na VM para chamadas explícitas de scripts.
	TipoBytes.Nova = func(args Tupla) (Objeto, error) {
		return NewBytes(args[0])
	}
}

// Tipo retorna a representação de classe (Tipo de Bytes).
func (b *Bytes) Tipo() *Tipo {
	return TipoBytes
}

// M__diferente__ compara a desigualdade lógica de arrays de bytes.
func (b *Bytes) M__diferente__(outro Objeto) (Objeto, error) {
	res, err := b.M__igual__(outro)
	if err != nil {
		return nil, err
	}

	return Booleano(!res.(Booleano)), nil
}

// M__igual__ compara a igualdade lógica de dois arrays de bytes através de bytes.Equal do Go.
func (b *Bytes) M__igual__(outro Objeto) (Objeto, error) {
	if !MesmoTipo(b, outro) {
		return Falso, nil
	}

	return NewBooleano(bytes.Equal(b.Itens, outro.(*Bytes).Itens))
}

// M__maior_ou_igual__ compara os comprimentos físicos dos bytes.
func (b *Bytes) M__maior_ou_igual__(outro Objeto) (Objeto, error) {
	outroT, err := Tamanho(outro)
	if err != nil {
		return nil, err
	}

	return NewBooleano(int64(len(b.Itens)) >= int64(outroT.(Inteiro)))
}

// M__maior_que__ compara os comprimentos físicos dos bytes.
func (b *Bytes) M__maior_que__(outro Objeto) (Objeto, error) {
	outroT, err := Tamanho(outro)
	if err != nil {
		return nil, err
	}

	return NewBooleano(int64(len(b.Itens)) > int64(outroT.(Inteiro)))
}

// M__menor_ou_igual__ compara os comprimentos físicos dos bytes.
func (b *Bytes) M__menor_ou_igual__(outro Objeto) (Objeto, error) {
	outroT, err := Tamanho(outro)
	if err != nil {
		return nil, err
	}

	return NewBooleano(int64(len(b.Itens)) <= int64(outroT.(Inteiro)))
}

// M__menor_que__ compara os comprimentos físicos dos bytes.
func (b *Bytes) M__menor_que__(outro Objeto) (Objeto, error) {
	outroT, err := Tamanho(outro)
	if err != nil {
		return nil, err
	}

	return NewBooleano(int64(len(b.Itens)) < int64(outroT.(Inteiro)))
}

// M__tamanho__ retorna o comprimento físico total de bytes armazenados no array local.
func (b *Bytes) M__tamanho__() (Objeto, error) {
	if b.Itens == nil {
		return Inteiro(0), nil
	}

	return NewInteiro(len(b.Itens))
}

// M__booleano__ avalia a verdade lógica do array (Falso se vazio, Verdadeiro do contrário).
func (b *Bytes) M__booleano__() (Objeto, error) {
	return NewBooleano(len(b.Itens) > 0)
}

// M__decimal__ lança pânico (não implementado coerção de float a partir de raw bytes diretamente).
func (b *Bytes) M__decimal__() (Objeto, error) {
	panic("unimplemented")
}

// M__inteiro__ lança pânico (não implementado coerção de int a partir de raw bytes diretamente).
func (b *Bytes) M__inteiro__() (Objeto, error) {
	panic("unimplemented")
}

// M__texto__ converte os bytes raw em sua representação string correspondente e retorna um objeto Texto.
func (b *Bytes) M__texto__() (Objeto, error) {
	return Texto(b.Itens), nil
}

// M__adiciona__ realiza a concatenação física de dois arrays de bytes (via append nativo do Go).
func (b *Bytes) M__adiciona__(outro Objeto) (Objeto, error) {
	if !MesmoTipo(b, outro) {
		return nil, NewErroF(TipagemErro, "Não é possível concatenar o tipo '%s' com '%s'", b.Tipo().Nome, outro.Tipo().Nome)
	}

	bytesObj, err := NewBytes(outro)

	if err != nil {
		return nil, err
	}

	return &Bytes{Itens: append(b.Itens, bytesObj.(*Bytes).Itens...)}, nil
}

// Garantias de conformidade com as interfaces nativas Go.
var _ I_comparacaoRica = (*Bytes)(nil)
var _ I__tamanho__ = (*Bytes)(nil)
var _ I_conversaoEntreTipos = (*Bytes)(nil)
