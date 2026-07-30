package parser_test

import (
	"reflect"

	"github.com/mat-dgruber/Harpia/lexer"
	"github.com/mat-dgruber/Harpia/parser"
)

func createParser(code string) *parser.Parser {
	lexer := lexer.NewLexer(code)
	return parser.NewParser(lexer)
}

func limparPosicoesVal(v reflect.Value) {
	if !v.IsValid() {
		return
	}
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return
		}
		if v.Type().Elem().Kind() == reflect.Struct {
			elem := v.Elem()
			if field := elem.FieldByName("NoBase"); field.IsValid() && field.CanSet() {
				field.Set(reflect.ValueOf(parser.NoBase{}))
			}
			for i := 0; i < elem.NumField(); i++ {
				limparPosicoesVal(elem.Field(i))
			}
		}
		return
	}
	if v.Kind() == reflect.Slice || v.Kind() == reflect.Array {
		for i := 0; i < v.Len(); i++ {
			limparPosicoesVal(v.Index(i))
		}
		return
	}
	if v.Kind() == reflect.Interface {
		if !v.IsNil() {
			limparPosicoesVal(v.Elem())
		}
		return
	}
	if v.Kind() == reflect.Struct {
		if field := v.FieldByName("NoBase"); field.IsValid() && field.CanSet() {
			field.Set(reflect.ValueOf(parser.NoBase{}))
		}
		for i := 0; i < v.NumField(); i++ {
			limparPosicoesVal(v.Field(i))
		}
	}
}

func isAstEquals(recebida *parser.Programa, esperada *parser.Programa) bool {
	if recebida != nil {
		limparPosicoesVal(reflect.ValueOf(recebida))
	}
	if esperada != nil {
		limparPosicoesVal(reflect.ValueOf(esperada))
	}
	return reflect.DeepEqual(recebida.Declaracoes, esperada.Declaracoes)
}

func createParserAndCompare(code string, expected *parser.Programa) (error, bool) {
	received, err := createParser(code).Parse()

	if err != nil {
		return err, false
	}

	return nil, isAstEquals(received, expected)
}
