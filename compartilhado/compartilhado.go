package compartilhado

import (
	"unicode"
)

// ContemApenasAlfaNum analisa se a string informada e estritamente composta por letras e/ou digitos Unicode.
func ContemApenasAlfaNum(str string) bool {
	if str == "" {
		return false
	}
	for _, char := range str {
		if !unicode.IsLetter(char) && !unicode.IsDigit(char) {
			return false
		}
	}
	return true
}

// EIdentificadorValido verifica se a string e um identificador valido na linguagem Harpia
// (comeca com letra ou underline _, seguido por letras, digitos ou underline).
func EIdentificadorValido(str string) bool {
	if str == "" {
		return false
	}
	for i, char := range str {
		if i == 0 {
			if !unicode.IsLetter(char) && char != '_' {
				return false
			}
		} else {
			if !unicode.IsLetter(char) && !unicode.IsDigit(char) && char != '_' {
				return false
			}
		}
	}
	return true
}
