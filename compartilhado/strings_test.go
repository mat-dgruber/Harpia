package compartilhado

import (
	"testing"
)

func TestEhASCII(t *testing.T) {
	if !EhASCII("hello world 123") {
		t.Errorf("esperava true para string ASCII pura")
	}
	if EhASCII("Harpia é incrível") {
		t.Errorf("esperava false para string com caractere acentuado")
	}
}

func TestIndiceCaraterParaByte_BoundsEInfiniteLoop(t *testing.T) {
	str := "Harpia"
	// Indice muito maior que a contagem de runas nao deve causar loop infinito
	idx := IndiceCaraterParaByteSemCache(str, 100)
	if idx != len(str) {
		t.Errorf("esperava len(str)=%d para indice out-of-bounds, obteve %d", len(str), idx)
	}
}

func TestObtemCaraterPorIndice_Multibyte(t *testing.T) {
	str := "Harpia 🚀"
	caractere := ObtemCaraterPorIndice(str, 7, nil)
	if caractere != "🚀" {
		t.Errorf("esperava 🚀, obteve %s", caractere)
	}
}

func BenchmarkIndiceBytePorCarater_ASCII(b *testing.B) {
	str := "var nome = \"Harpia Fullstack Framework\";"
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = IndiceBytePorCarater(str)
	}
}
