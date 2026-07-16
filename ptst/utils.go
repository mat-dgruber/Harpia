package ptst

import (
	"os"
	"path"
	"path/filepath"
	"strings"
)

// MesmoTipo compara as assinaturas de classe de dois objetos dinâmicos de forma extremamente rápida.
func MesmoTipo(a, b Objeto) bool {
	return a.Tipo() == b.Tipo()
}

// VerificaNumeroArgumentos é a rotina centralizada de verificação de aridade de parâmetros.
//
// Valida se a contagem de argumentos reais passados na tupla atende aos limites mínimo e máximo especificados,
// lançando um erro amigável de Tipagem (TipagemErro) se a restrição for violada.
func VerificaNumeroArgumentos(nome string, ehMetodo bool, args Objeto, min, max int) error {
	numArgs := len(args.(Tupla))

	if numArgs < min || numArgs > max {
		tipo := "a função"
		if ehMetodo {
			tipo = "o método"
		}
		return NewErroF(TipagemErro, "Número incorreto de argumentos para %s %s. Esperava entre %d e %d, mas recebeu %d", tipo, nome, min, max, numArgs)
	}

	return nil
}

// ResolveArquivoPtst localiza e calcula o caminho absoluto de arquivos físicos de scripts ou módulos no disco rígido.
//
// Algoritmo de Resolução de Caminhos:
//  - Varre sequencialmente as pastas bases de busca (incluindo o diretório de execução corrente 'curDir');
//  - Se a pasta mapeada contiver um arquivo de entrada mestre 'inicio.ptst' dentro dela, resolve importando-o;
//  - Se o caminho referenciado omitir a extensão física, tenta anexar e buscar arquivos compilados '.so'
//    e, em caso de insucesso, busca arquivos com a extensão '.ptst' (extensão antiga) ou '.pt' de forma transparente.
func ResolveArquivoPtst(caminhoArqOuMod string, bases []string, curDir string) (string, error) {
	caminhoArqOuMod = strings.TrimSuffix(caminhoArqOuMod, "/")

	if len(curDir) > 0 {
		bases = append([]string{curDir}, bases...)
	}

	stat, err := os.Stat(caminhoArqOuMod)

	if path.IsAbs(caminhoArqOuMod) && err == nil && !stat.IsDir() {
		return caminhoArqOuMod, nil
	}

	for _, base := range bases {
		caminho, _ := filepath.Abs(path.Join(base, caminhoArqOuMod))

		stat, err = os.Stat(caminho)
		if err == nil && stat.IsDir() {
			ca := path.Join(caminho, "inicio.hrp")
			_, err = os.Stat(ca)
			if err != nil {
				ca = path.Join(caminho, "inicio.ptst")
				_, err = os.Stat(ca)
			}

			if err == nil {
				caminho = ca
			}
		}

		if filepath.Ext(caminho) == "" && os.IsNotExist(err) {
			caminho += ".so"
			_, err = os.Stat(caminho)

			if err != nil {
				caminho = strings.Replace(caminho, filepath.Ext(caminho), ".hrp", 1)
				_, err = os.Stat(caminho)
				if err != nil {
					caminho = strings.Replace(caminho, filepath.Ext(caminho), ".ptst", 1)
					_, err = os.Stat(caminho)
				}
			}
		}

		if err != nil {
			if os.IsNotExist(err) {
				continue
			}

			return "", NewErroF(ErroDeSistema, "Erro ao acessar '%s': %s", caminho, err)
		}

		return caminho, nil
	}

	if err != nil && os.IsNotExist(err) {
		return "", NewErroF(ArquivoNaoEncontradoErro, "Não foi possível encontrar o arquivo '%s'", caminhoArqOuMod)
	}

	return "", nil
}

// TalvezLanceErroDivisaoPorZero intercepta tentativas de realizar divisões reais, inteiras ou restos por zero.
// Centraliza a prevenção de crash matemático lançando a exceção controlada DivisaoPorZeroErro.
func TalvezLanceErroDivisaoPorZero(obj Objeto) error {
	switch t := obj.(type) {
	case Inteiro:
		if t == 0 {
			return NewErroF(DivisaoPorZeroErro, "Não é possível dividir por zero")
		}
		return nil
	case Decimal:
		if t == 0.0 {
			return NewErroF(DivisaoPorZeroErro, "Não é possível dividir pelo decimal 0.0")
		}

		return nil
	default:
		return nil
	}
}

// FuncaoComErro é a assinatura genérica para closures de coerção/casting em Go.
type FuncaoComErro[T any] func(T) (Objeto, error)

// RetornaOuPanic executa o método com tratamento de pânico integrado em Go.
func RetornaOuPanic[T any](f FuncaoComErro[T], arg T) Objeto {
	result, err := f(arg)
	if err != nil {
		panic(err)
	}
	return result
}
