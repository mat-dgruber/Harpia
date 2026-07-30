package lexer_test

import (
	// "reflect"

	"testing"

	"github.com/mat-dgruber/Harpia/lexer"
)

func TestFuncaoSimples(t *testing.T) {
	code := `nenhumaOperacao(2023, "harpia")`

	lex := lexer.NewLexer(code)
	tokens := []lexer.Token{
		{Tipo: lexer.TokenIdentificador, Valor: "nenhumaOperacao"},
		{Tipo: lexer.TokenAbreParenteses, Valor: "("},
		{Tipo: lexer.TokenInteiro, Valor: "2023"},
		{Tipo: lexer.TokenVirgula, Valor: ","},
		{Tipo: lexer.TokenTexto, Valor: "\"harpia\""},
		{Tipo: lexer.TokenFechaParenteses, Valor: ")"},
	}

	for _, esperado := range tokens {
		token := lex.PróximoToken()
		if token.Tipo != esperado.Tipo || token.Valor != esperado.Valor {
			t.Errorf("Esperado: (%v, %s), Obtido: (%v, %s)", esperado.Tipo, esperado.Valor, token.Tipo, token.Valor)
		}
	}

	// Verifica se o último token é EOF.
	ultimoToken := lex.PróximoToken()
	if ultimoToken.Tipo != lexer.TokenFimDeArquivo {
		t.Errorf("Esperado: TokenFimDeArquivo, Obtido: %v", ultimoToken.Tipo)
	}
}

func TestLexer(t *testing.T) {
	input := `
		var x = 10;
        const PI = 3.1415;
        funcao soma(a, b) {
            retorne a + b;
        }

        se (x == 10) {
            texto = "Hello, World!";
        } senao {
            texto = "Não é 10";
        }

        enquanto (x > 0) {
            x = x - 1;
        }

        para (i = 0; i < 5; i = i + 1) {
            imprima(i);
        }
    `

	lex := lexer.NewLexer(input)
	tokens := []lexer.Token{
		{Tipo: lexer.TokenNovaLinha, Valor: "\n"},
		{Tipo: lexer.TokenVar, Valor: "var"},
		{Tipo: lexer.TokenIdentificador, Valor: "x"},
		{Tipo: lexer.TokenIgual, Valor: "="},
		{Tipo: lexer.TokenInteiro, Valor: "10"},
		{Tipo: lexer.TokenPontoEVirgula, Valor: ";"},
		{Tipo: lexer.TokenNovaLinha, Valor: "\n"},

		{Tipo: lexer.TokenConst, Valor: "const"},
		{Tipo: lexer.TokenIdentificador, Valor: "PI"},
		{Tipo: lexer.TokenIgual, Valor: "="},
		{Tipo: lexer.TokenDecimal, Valor: "3.1415"},
		{Tipo: lexer.TokenPontoEVirgula, Valor: ";"},
		{Tipo: lexer.TokenNovaLinha, Valor: "\n"},

		{Tipo: lexer.TokenFunc, Valor: "funcao"},
		{Tipo: lexer.TokenIdentificador, Valor: "soma"},
		{Tipo: lexer.TokenAbreParenteses, Valor: "("},
		{Tipo: lexer.TokenIdentificador, Valor: "a"},
		{Tipo: lexer.TokenVirgula, Valor: ","},
		{Tipo: lexer.TokenIdentificador, Valor: "b"},
		{Tipo: lexer.TokenFechaParenteses, Valor: ")"},
		{Tipo: lexer.TokenAbreChaves, Valor: "{"},
		{Tipo: lexer.TokenNovaLinha, Valor: "\n"},
		{Tipo: lexer.TokenRetorne, Valor: "retorne"},
		{Tipo: lexer.TokenIdentificador, Valor: "a"},
		{Tipo: lexer.TokenMais, Valor: "+"},
		{Tipo: lexer.TokenIdentificador, Valor: "b"},
		{Tipo: lexer.TokenPontoEVirgula, Valor: ";"},
		{Tipo: lexer.TokenNovaLinha, Valor: "\n"},
		{Tipo: lexer.TokenFechaChaves, Valor: "}"},
		{Tipo: lexer.TokenNovaLinha, Valor: "\n"},

		{Tipo: lexer.TokenNovaLinha, Valor: "\n"},

		{Tipo: lexer.TokenSe, Valor: "se"},
		{Tipo: lexer.TokenAbreParenteses, Valor: "("},
		{Tipo: lexer.TokenIdentificador, Valor: "x"},
		{Tipo: lexer.TokenIgualIgual, Valor: "=="},
		{Tipo: lexer.TokenInteiro, Valor: "10"},
		{Tipo: lexer.TokenFechaParenteses, Valor: ")"},
		{Tipo: lexer.TokenAbreChaves, Valor: "{"},
		{Tipo: lexer.TokenNovaLinha, Valor: "\n"},
		{Tipo: lexer.TokenIdentificador, Valor: "texto"},
		{Tipo: lexer.TokenIgual, Valor: "="},
		{Tipo: lexer.TokenTexto, Valor: "\"Hello, World!\""},
		{Tipo: lexer.TokenPontoEVirgula, Valor: ";"},
		{Tipo: lexer.TokenNovaLinha, Valor: "\n"},
		{Tipo: lexer.TokenFechaChaves, Valor: "}"},
		{Tipo: lexer.TokenSenao, Valor: "senao"},
		{Tipo: lexer.TokenAbreChaves, Valor: "{"},
		{Tipo: lexer.TokenNovaLinha, Valor: "\n"},
		{Tipo: lexer.TokenIdentificador, Valor: "texto"},
		{Tipo: lexer.TokenIgual, Valor: "="},
		{Tipo: lexer.TokenTexto, Valor: "\"Não é 10\""},
		{Tipo: lexer.TokenPontoEVirgula, Valor: ";"},
		{Tipo: lexer.TokenNovaLinha, Valor: "\n"},
		{Tipo: lexer.TokenFechaChaves, Valor: "}"},
		{Tipo: lexer.TokenNovaLinha, Valor: "\n"},

		{Tipo: lexer.TokenNovaLinha, Valor: "\n"},

		{Tipo: lexer.TokenEnquanto, Valor: "enquanto"},
		{Tipo: lexer.TokenAbreParenteses, Valor: "("},
		{Tipo: lexer.TokenIdentificador, Valor: "x"},
		{Tipo: lexer.TokenMaiorQue, Valor: ">"},
		{Tipo: lexer.TokenInteiro, Valor: "0"},
		{Tipo: lexer.TokenFechaParenteses, Valor: ")"},
		{Tipo: lexer.TokenAbreChaves, Valor: "{"},
		{Tipo: lexer.TokenNovaLinha, Valor: "\n"},
		{Tipo: lexer.TokenIdentificador, Valor: "x"},
		{Tipo: lexer.TokenIgual, Valor: "="},
		{Tipo: lexer.TokenIdentificador, Valor: "x"},
		{Tipo: lexer.TokenMenos, Valor: "-"},
		{Tipo: lexer.TokenInteiro, Valor: "1"},
		{Tipo: lexer.TokenPontoEVirgula, Valor: ";"},
		{Tipo: lexer.TokenNovaLinha, Valor: "\n"},
		{Tipo: lexer.TokenFechaChaves, Valor: "}"},
		{Tipo: lexer.TokenNovaLinha, Valor: "\n"},

		{Tipo: lexer.TokenNovaLinha, Valor: "\n"},

		{Tipo: lexer.TokenPara, Valor: "para"},
		{Tipo: lexer.TokenAbreParenteses, Valor: "("},
		{Tipo: lexer.TokenIdentificador, Valor: "i"},
		{Tipo: lexer.TokenIgual, Valor: "="},
		{Tipo: lexer.TokenInteiro, Valor: "0"},
		{Tipo: lexer.TokenPontoEVirgula, Valor: ";"},
		{Tipo: lexer.TokenIdentificador, Valor: "i"},
		{Tipo: lexer.TokenMenorQue, Valor: "<"},
		{Tipo: lexer.TokenInteiro, Valor: "5"},
		{Tipo: lexer.TokenPontoEVirgula, Valor: ";"},
		{Tipo: lexer.TokenIdentificador, Valor: "i"},
		{Tipo: lexer.TokenIgual, Valor: "="},
		{Tipo: lexer.TokenIdentificador, Valor: "i"},
		{Tipo: lexer.TokenMais, Valor: "+"},
		{Tipo: lexer.TokenInteiro, Valor: "1"},
		{Tipo: lexer.TokenFechaParenteses, Valor: ")"},
		{Tipo: lexer.TokenAbreChaves, Valor: "{"},
		{Tipo: lexer.TokenNovaLinha, Valor: "\n"},
		{Tipo: lexer.TokenIdentificador, Valor: "imprima"},
		{Tipo: lexer.TokenAbreParenteses, Valor: "("},
		{Tipo: lexer.TokenIdentificador, Valor: "i"},
		{Tipo: lexer.TokenFechaParenteses, Valor: ")"},
		{Tipo: lexer.TokenPontoEVirgula, Valor: ";"},
		{Tipo: lexer.TokenNovaLinha, Valor: "\n"},
		{Tipo: lexer.TokenFechaChaves, Valor: "}"},
		{Tipo: lexer.TokenNovaLinha, Valor: "\n"},
	}

	for _, esperado := range tokens {
		token := lex.PróximoToken()
		if token.Tipo != esperado.Tipo || token.Valor != esperado.Valor {
			t.Errorf("Esperado: (%v, %s), Obtido: (%v, %s)", esperado.Tipo, esperado.Valor, token.Tipo, token.Valor)
		}
	}

	// Verifica se o último token é EOF.
	ultimoToken := lex.PróximoToken()
	if ultimoToken.Tipo != lexer.TokenFimDeArquivo {
		t.Errorf("Esperado: TokenFimDeArquivo, Obtido: %v", ultimoToken.Tipo)
	}
}

func TestShebangEComentarioBloco(t *testing.T) {
	input := `#!/usr/bin/env harpia
/*
 * Comentário de bloco
 * em múltiplas linhas
 */
var x = 10; /* Comentário de bloco inline */
const y = 20;
`

	lex := lexer.NewLexer(input)
	tokens := []lexer.Token{
		{Tipo: lexer.TokenNovaLinha, Valor: "\n"},
		{Tipo: lexer.TokenVar, Valor: "var"},
		{Tipo: lexer.TokenIdentificador, Valor: "x"},
		{Tipo: lexer.TokenIgual, Valor: "="},
		{Tipo: lexer.TokenInteiro, Valor: "10"},
		{Tipo: lexer.TokenPontoEVirgula, Valor: ";"},
		{Tipo: lexer.TokenNovaLinha, Valor: "\n"},
		{Tipo: lexer.TokenConst, Valor: "const"},
		{Tipo: lexer.TokenIdentificador, Valor: "y"},
		{Tipo: lexer.TokenIgual, Valor: "="},
		{Tipo: lexer.TokenInteiro, Valor: "20"},
		{Tipo: lexer.TokenPontoEVirgula, Valor: ";"},
		{Tipo: lexer.TokenNovaLinha, Valor: "\n"},
	}

	for i, esperado := range tokens {
		token := lex.PróximoToken()
		if token.Tipo != esperado.Tipo || token.Valor != esperado.Valor {
			t.Errorf("[%d] Esperado: (%v, %q), Obtido: (%v, %q)", i, esperado.Tipo, esperado.Valor, token.Tipo, token.Valor)
		}
	}

	ultimoToken := lex.PróximoToken()
	if ultimoToken.Tipo != lexer.TokenFimDeArquivo {
		t.Errorf("Esperado: TokenFimDeArquivo, Obtido: %v", ultimoToken.Tipo)
	}
}

func TestNumerosAvancados(t *testing.T) {
	input := "0x1F 0b1010 0o755 1_000_000 0x1F_2A 0b1010_1100 0o755_644 3.14_15"
	lex := lexer.NewLexer(input)

	esperados := []struct {
		tipo  lexer.TokenType
		valor string
	}{
		{lexer.TokenInteiro, "0x1F"},
		{lexer.TokenInteiro, "0b1010"},
		{lexer.TokenInteiro, "0o755"},
		{lexer.TokenInteiro, "1_000_000"},
		{lexer.TokenInteiro, "0x1F_2A"},
		{lexer.TokenInteiro, "0b1010_1100"},
		{lexer.TokenInteiro, "0o755_644"},
		{lexer.TokenDecimal, "3.14_15"},
	}

	for i, esp := range esperados {
		tok := lex.PróximoToken()
		if tok.Tipo != esp.tipo || tok.Valor != esp.valor {
			t.Errorf("[%d] Esperado (%v, %q), obtido (%v, %q)", i, esp.tipo, esp.valor, tok.Tipo, tok.Valor)
		}
	}
}

func TestStringNaoTerminada(t *testing.T) {
	input := `"Texto sem fechar`
	lex := lexer.NewLexer(input)
	tok := lex.PróximoToken()

	if tok.Tipo != lexer.TokenErro {
		t.Errorf("Esperado TokenErro, obtido: %v", tok.Tipo)
	}
	if tok.Valor != "string não terminada" {
		t.Errorf("Esperado 'string não terminada', obtido: %q", tok.Valor)
	}
}

func TestBugFixesEInStreamCRLF(t *testing.T) {
	t.Run("Negação !", func(t *testing.T) {
		lex := lexer.NewLexer("!verdadeiro")
		tok1 := lex.PróximoToken()
		if tok1.Tipo != lexer.TokenBoolNao || tok1.Valor != "!" {
			t.Errorf("Esperado (TokenBoolNao, '!'), obtido (%v, %q)", tok1.Tipo, tok1.Valor)
		}
		tok2 := lex.PróximoToken()
		if tok2.Tipo != lexer.TokenIdentificador || tok2.Valor != "verdadeiro" {
			t.Errorf("Esperado (TokenIdentificador, 'verdadeiro'), obtido (%v, %q)", tok2.Tipo, tok2.Valor)
		}
	})

	t.Run("Base sem digitos 0x", func(t *testing.T) {
		lex := lexer.NewLexer("0x + 1")
		tok1 := lex.PróximoToken()
		if tok1.Tipo != lexer.TokenErro || tok1.Valor != "literal numérico malformado" {
			t.Errorf("Esperado (TokenErro, 'literal numérico malformado'), obtido (%v, %q)", tok1.Tipo, tok1.Valor)
		}
		tok2 := lex.PróximoToken()
		if tok2.Tipo != lexer.TokenMais || tok2.Valor != "+" {
			t.Errorf("Esperado (TokenMais, '+'), obtido (%v, %q)", tok2.Tipo, tok2.Valor)
		}
	})

	t.Run("CRLF in-stream", func(t *testing.T) {
		code := "var x = 1;\r\nconst y = 2;"
		lex := lexer.NewLexer(code)
		esperados := []struct {
			tipo  lexer.TokenType
			valor string
		}{
			{lexer.TokenVar, "var"},
			{lexer.TokenIdentificador, "x"},
			{lexer.TokenIgual, "="},
			{lexer.TokenInteiro, "1"},
			{lexer.TokenPontoEVirgula, ";"},
			{lexer.TokenNovaLinha, "\n"},
			{lexer.TokenConst, "const"},
			{lexer.TokenIdentificador, "y"},
			{lexer.TokenIgual, "="},
			{lexer.TokenInteiro, "2"},
			{lexer.TokenPontoEVirgula, ";"},
		}

		for i, esp := range esperados {
			tok := lex.PróximoToken()
			if tok.Tipo != esp.tipo || tok.Valor != esp.valor {
				t.Errorf("[%d] Esperado (%v, %q), obtido (%v, %q)", i, esp.tipo, esp.valor, tok.Tipo, tok.Valor)
			}
		}
	})
}


