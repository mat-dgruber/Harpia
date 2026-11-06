package hrp

import "strings"

// DecodificarTextoLiteral converte o lexema de um literal de texto (com as aspas) no valor final.
//
// Aspas duplas e simples processam as sequências de escape \n, \t, \r, \0, \\, \", \', \{ e \};
// um escape desconhecido (ex.: \d de uma regex) é mantido como está, barra incluída.
// Strings entre crases (`...`) são brutas: só as delimitações são removidas.
// Usado pelo interpretador e pela VM, que antes só cortavam as aspas ("\n" ficava literal,
// divergindo dos transpiladores web/nativo, onde JS e Go já decodificam).
func DecodificarTextoLiteral(lexema string) string {
	r := []rune(lexema)
	if len(r) < 2 {
		return ""
	}
	corpo := r[1 : len(r)-1]
	if r[0] == '`' || !strings.ContainsRune(string(corpo), '\\') {
		return string(corpo)
	}
	var b strings.Builder
	for i := 0; i < len(corpo); i++ {
		if corpo[i] != '\\' || i+1 >= len(corpo) {
			b.WriteRune(corpo[i])
			continue
		}
		i++
		switch corpo[i] {
		case 'n':
			b.WriteRune('\n')
		case 't':
			b.WriteRune('\t')
		case 'r':
			b.WriteRune('\r')
		case '0':
			b.WriteRune(0)
		case '\\', '"', '\'', '{', '}':
			b.WriteRune(corpo[i])
		default:
			b.WriteRune('\\')
			b.WriteRune(corpo[i])
		}
	}
	return b.String()
}
