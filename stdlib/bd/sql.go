package bd

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/mat-dgruber/Harpia/hrp"
)

type ConexaoSQL struct {
	db     *sql.DB
	driver string
}

var TipoConexaoSQL = hrp.NewTipo("ConexaoSQL", "Conexão de Banco de Dados SQL")

func (c *ConexaoSQL) Tipo() *hrp.Tipo {
	return TipoConexaoSQL
}

func (c *ConexaoSQL) M__obtem_attributo__(nome string) (hrp.Objeto, error) {
	switch nome {
	case "executar":
		return hrp.NewMetodoOuPanic("executar", func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
			if len(args) < 1 {
				return nil, hrp.NewErroF(hrp.TipagemErro, "executar esperava no mínimo 1 argumento (query)")
			}
			query, err := hrp.NewTexto(args[0])
			if err != nil {
				return nil, err
			}
			var goArgs []interface{}
			for _, arg := range args[1:] {
				goArgs = append(goArgs, toGoType(arg))
			}
			_, errExec := c.db.Exec(string(query.(hrp.Texto)), goArgs...)
			if errExec != nil {
				return nil, hrp.NewErroF(hrp.ErroDeSistema, "Erro ao executar query SQL: %v", errExec)
			}
			return hrp.Nulo, nil
		}, ""), nil

	case "consultar":
		return hrp.NewMetodoOuPanic("consultar", func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
			if len(args) < 1 {
				return nil, hrp.NewErroF(hrp.TipagemErro, "consultar esperava no mínimo 1 argumento (query)")
			}
			query, err := hrp.NewTexto(args[0])
			if err != nil {
				return nil, err
			}
			var goArgs []interface{}
			for _, arg := range args[1:] {
				goArgs = append(goArgs, toGoType(arg))
			}
			rows, errQuery := c.db.Query(string(query.(hrp.Texto)), goArgs...)
			if errQuery != nil {
				return nil, hrp.NewErroF(hrp.ErroDeSistema, "Erro ao executar consulta SQL: %v", errQuery)
			}
			defer rows.Close()

			cols, errCols := rows.Columns()
			if errCols != nil {
				return nil, hrp.NewErroF(hrp.ErroDeSistema, "Erro ao obter colunas: %v", errCols)
			}

			lista := &hrp.Lista{}
			for rows.Next() {
				columns := make([]interface{}, len(cols))
				columnPointers := make([]interface{}, len(cols))
				for i := range columns {
					columnPointers[i] = &columns[i]
				}

				if errScan := rows.Scan(columnPointers...); errScan != nil {
					return nil, hrp.NewErroF(hrp.ErroDeSistema, "Erro ao escanear linha: %v", errScan)
				}

				mapaLinha := hrp.NewMapaVazio()
				for i, colName := range cols {
					val := columns[i]
					mapaLinha.M__define_item__(hrp.Texto(colName), toPtObject(val))
				}
				lista.Adiciona(mapaLinha)
			}

			return lista, nil
		}, ""), nil

	case "tabela":
		return hrp.NewMetodoOuPanic("tabela", func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
			if err := hrp.VerificaNumeroArgumentos("tabela", false, args, 1, 2); err != nil {
				return nil, err
			}
			tabelaTexto, err := hrp.NewTexto(args[0])
			if err != nil {
				return nil, err
			}
			var schema hrp.Mapa
			if len(args) == 2 {
				s, ok := args[1].(hrp.Mapa)
				if !ok {
					return nil, hrp.NewErroF(hrp.TipagemErro, "tabela esperava um Mapa como segundo argumento (schema)")
				}
				schema = s
			}
			return &QueryBuilder{
				conexao: c,
				tabela:  string(tabelaTexto.(hrp.Texto)),
				schema:  schema,
			}, nil
		}, ""), nil

	case "fechar":
		return hrp.NewMetodoOuPanic("fechar", func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
			c.db.Close()
			return hrp.Nulo, nil
		}, ""), nil
	}

	return nil, hrp.NewErroF(hrp.AtributoErro, "Atributo '%s' não existe em ConexaoSQL", nome)
}

type QueryBuilder struct {
	conexao      *ConexaoSQL
	tabela       string
	selecionados []string
	condicoes    []string
	args         []hrp.Objeto
	limiteVal    int
	schema       hrp.Mapa
}

var TipoQueryBuilder = hrp.NewTipo("QueryBuilder", "Query Builder dinâmico")

func (q *QueryBuilder) Tipo() *hrp.Tipo {
	return TipoQueryBuilder
}

func (q *QueryBuilder) M__obtem_attributo__(nome string) (hrp.Objeto, error) {
	switch nome {
	case "selecionar":
		return hrp.NewMetodoOuPanic("selecionar", func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
			for _, arg := range args {
				col, err := hrp.NewTexto(arg)
				if err != nil {
					return nil, err
				}
				q.selecionados = append(q.selecionados, string(col.(hrp.Texto)))
			}
			return q, nil
		}, ""), nil

	case "onde":
		return hrp.NewMetodoOuPanic("onde", func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
			if err := hrp.VerificaNumeroArgumentos("onde", false, args, 3, 3); err != nil {
				return nil, err
			}
			coluna, err := hrp.NewTexto(args[0])
			if err != nil {
				return nil, err
			}
			operador, err := hrp.NewTexto(args[1])
			if err != nil {
				return nil, err
			}
			q.condicoes = append(q.condicoes, fmt.Sprintf("%s %s ?", string(coluna.(hrp.Texto)), string(operador.(hrp.Texto))))
			q.args = append(q.args, args[2])
			return q, nil
		}, ""), nil

	case "limite":
		return hrp.NewMetodoOuPanic("limite", func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
			if err := hrp.VerificaNumeroArgumentos("limite", false, args, 1, 1); err != nil {
				return nil, err
			}
			n, err := hrp.NewInteiro(args[0])
			if err != nil {
				return nil, err
			}
			nVal := int64(n.(hrp.Inteiro))
			if nVal < 0 || nVal > 1000000 {
				return nil, fmt.Errorf("limite inválido")
			}
			q.limiteVal = int(nVal)
			return q, nil
		}, ""), nil

	case "obterMuitos":
		return hrp.NewMetodoOuPanic("obterMuitos", func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
			cols := "*"
			if len(q.selecionados) > 0 {
				cols = strings.Join(q.selecionados, ", ")
			}
			query := fmt.Sprintf("SELECT %s FROM %s", cols, q.tabela)
			if len(q.condicoes) > 0 {
				query += " WHERE " + strings.Join(q.condicoes, " AND ")
			}
			if q.limiteVal > 0 {
				query += fmt.Sprintf(" LIMIT %d", q.limiteVal)
			}
			query = q.converterPlaceholders(query)

			var callArgs hrp.Tupla
			callArgs = append(callArgs, hrp.Texto(query))
			for _, arg := range q.args {
				callArgs = append(callArgs, arg)
			}

			consultarMetodo, err := q.conexao.M__obtem_attributo__("consultar")
			if err != nil {
				return nil, err
			}
			return hrp.Chamar(consultarMetodo, callArgs)
		}, ""), nil

	case "obterUm":
		return hrp.NewMetodoOuPanic("obterUm", func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
			q.limiteVal = 1
			obterMuitosMetodo, err := q.M__obtem_attributo__("obterMuitos")
			if err != nil {
				return nil, err
			}
			res, err := hrp.Chamar(obterMuitosMetodo, hrp.Tupla{})
			if err != nil {
				return nil, err
			}
			lista := res.(*hrp.Lista)
			if len(lista.Itens) == 0 {
				return hrp.Nulo, nil
			}
			return lista.Itens[0], nil
		}, ""), nil

	case "inserir":
		return hrp.NewMetodoOuPanic("inserir", func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
			if err := hrp.VerificaNumeroArgumentos("inserir", false, args, 1, 1); err != nil {
				return nil, err
			}
			mapa, ok := args[0].(hrp.Mapa)
			if !ok {
				return nil, hrp.NewErroF(hrp.TipagemErro, "inserir esperava um Mapa como argumento")
			}

			if q.schema != nil {
				for k, v := range mapa {
					tipoEsperado, existe := q.schema[k]
					if !existe {
						return nil, hrp.NewErroF(hrp.ValorErro, "campo '%s' nao existe no schema da tabela '%s'", k, q.tabela)
					}
					tipoTexto, ok := tipoEsperado.(hrp.Texto)
					if ok {
						if errVal := validarTipoCampo(string(tipoTexto), v); errVal != nil {
							return nil, hrp.NewErroF(hrp.TipagemErro, "campo '%s' da tabela '%s': %v", k, q.tabela, errVal)
						}
					}
				}
			}

			var colunas []string
			var placeholders []string
			var valueArgs hrp.Tupla

			for k, v := range mapa {
				colunas = append(colunas, k)
				placeholders = append(placeholders, "?")
				valueArgs = append(valueArgs, v)
			}

			query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", q.tabela, strings.Join(colunas, ", "), strings.Join(placeholders, ", "))
			query = q.converterPlaceholders(query)

			var callArgs hrp.Tupla
			callArgs = append(callArgs, hrp.Texto(query))
			for _, arg := range valueArgs {
				callArgs = append(callArgs, arg)
			}

			executarMetodo, err := q.conexao.M__obtem_attributo__("executar")
			if err != nil {
				return nil, err
			}
			return hrp.Chamar(executarMetodo, callArgs)
		}, ""), nil

	case "atualizar":
		return hrp.NewMetodoOuPanic("atualizar", func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
			if err := hrp.VerificaNumeroArgumentos("atualizar", false, args, 1, 1); err != nil {
				return nil, err
			}
			mapa, ok := args[0].(hrp.Mapa)
			if !ok {
				return nil, hrp.NewErroF(hrp.TipagemErro, "atualizar esperava um Mapa como argumento")
			}

			var sets []string
			var valueArgs hrp.Tupla

			for k, v := range mapa {
				sets = append(sets, fmt.Sprintf("%s = ?", k))
				valueArgs = append(valueArgs, v)
			}

			query := fmt.Sprintf("UPDATE %s SET %s", q.tabela, strings.Join(sets, ", "))
			if len(q.condicoes) > 0 {
				query += " WHERE " + strings.Join(q.condicoes, " AND ")
			}
			query = q.converterPlaceholders(query)

			var callArgs hrp.Tupla
			callArgs = append(callArgs, hrp.Texto(query))
			for _, arg := range valueArgs {
				callArgs = append(callArgs, arg)
			}
			for _, arg := range q.args {
				callArgs = append(callArgs, arg)
			}

			executarMetodo, err := q.conexao.M__obtem_attributo__("executar")
			if err != nil {
				return nil, err
			}
			return hrp.Chamar(executarMetodo, callArgs)
		}, ""), nil

	case "deletar":
		return hrp.NewMetodoOuPanic("deletar", func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
			query := fmt.Sprintf("DELETE FROM %s", q.tabela)
			if len(q.condicoes) > 0 {
				query += " WHERE " + strings.Join(q.condicoes, " AND ")
			}
			query = q.converterPlaceholders(query)

			var callArgs hrp.Tupla
			callArgs = append(callArgs, hrp.Texto(query))
			for _, arg := range q.args {
				callArgs = append(callArgs, arg)
			}

			executarMetodo, err := q.conexao.M__obtem_attributo__("executar")
			if err != nil {
				return nil, err
			}
			return hrp.Chamar(executarMetodo, callArgs)
		}, ""), nil
	}

	return nil, hrp.NewErroF(hrp.AtributoErro, "Atributo '%s' não existe no QueryBuilder", nome)
}

func toGoType(obj hrp.Objeto) interface{} {
	if obj == hrp.Nulo || obj.Tipo() == hrp.TipoNulo {
		return nil
	}
	switch v := obj.(type) {
	case hrp.Texto:
		return string(v)
	case hrp.Inteiro:
		return int64(v)
	case hrp.Decimal:
		return float64(v)
	case hrp.Booleano:
		return bool(v)
	}
	return fmt.Sprintf("%v", obj)
}

func toPtObject(val interface{}) hrp.Objeto {
	if val == nil {
		return hrp.Nulo
	}
	switch v := val.(type) {
	case string:
		return hrp.Texto(v)
	case []byte:
		return hrp.Texto(v)
	case int64:
		return hrp.Inteiro(v)
	case int:
		return hrp.Inteiro(v)
	case float64:
		return hrp.Decimal(v)
	case bool:
		return hrp.Booleano(v)
	}
	return hrp.Texto(fmt.Sprintf("%v", val))
}

func (q *QueryBuilder) converterPlaceholders(query string) string {
	if q.conexao.driver != "postgres" {
		return query
	}
	partes := strings.Split(query, "?")
	if len(partes) <= 1 {
		return query
	}
	var sb strings.Builder
	sb.WriteString(partes[0])
	for i := 1; i < len(partes); i++ {
		sb.WriteString(fmt.Sprintf("$%d", i))
		sb.WriteString(partes[i])
	}
	return sb.String()
}

func validarTipoCampo(tipoEsperado string, valor hrp.Objeto) error {
	switch tipoEsperado {
	case "texto":
		if _, ok := valor.(hrp.Texto); !ok {
			return fmt.Errorf("esperava tipo Texto, obteve %s", valor.Tipo().Nome)
		}
	case "inteiro":
		if _, ok := valor.(hrp.Inteiro); !ok {
			return fmt.Errorf("esperava tipo Inteiro, obteve %s", valor.Tipo().Nome)
		}
	case "decimal":
		if _, ok := valor.(hrp.Decimal); !ok {
			return fmt.Errorf("esperava tipo Decimal, obteve %s", valor.Tipo().Nome)
		}
	case "booleano":
		if _, ok := valor.(hrp.Booleano); !ok {
			return fmt.Errorf("esperava tipo Booleano, obteve %s", valor.Tipo().Nome)
		}
	}
	return nil
}
