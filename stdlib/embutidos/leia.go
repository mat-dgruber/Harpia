package embutidos

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/mat-dgruber/Harpia/hrp"
)

// met_emb_leia implementa a lógica nativa para a função global 'leia()'.
//
// Esta função opcionalmente exibe uma mensagem de instrução (prompt) no terminal,
// pausa a execução do interpretador aguardando a digitação de dados pelo desenvolvedor/usuário
// e retorna a cadeia textual digitada (como hrp.Texto) assim que a tecla Enter é acionada.
//
// Mecânica de Entrada:
//   - Se um argumento de prompt for passado, converte-o para texto e imprime via fmt.Printf;
//   - Instancia um leitor bufio.NewReader sobre a entrada padrão (os.Stdin);
//   - Lê os bytes de caracteres até encontrar o delimitador de quebra de linha ('\n');
//   - Remove o caractere final de quebra de linha com strings.TrimRight e devolve como hrp.Texto.
func met_emb_leia(mod hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if err := hrp.VerificaNumeroArgumentos("leia", false, args, 0, 1); err != nil {
		return nil, err
	}

	if len(args) == 1 {
		texto, err := hrp.NewTexto(args[0])

		if err != nil {
			return nil, err
		}

		fmt.Printf("%s", texto)
	}

	reader := bufio.NewReader(os.Stdin)
	leitura, err := reader.ReadString('\n')
	if err != nil {
		return nil, hrp.NewErroF(hrp.ErroDeSistema, "Erro ao ler a entrada: %v", err)
	}
	return hrp.Texto(strings.TrimRight(leitura, "\n")), nil
}

// _emb_leia cria e define a assinatura do método 'leia' exposto globalmente.
var _emb_leia = hrp.NewMetodoOuPanic(
	"leia",
	met_emb_leia,
	"leia(frase_para_imprimir) -> imprime um texto se especificado e lê uma entrada do usuário, retornando-a",
)
