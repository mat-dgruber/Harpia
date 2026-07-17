package bd

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/mat-dgruber/Harpia/ptst"
)

type ConexaoSQL struct {
	db     *sql.DB
	driver string
}

var TipoConexaoSQL = ptst.NewTipo("ConexaoSQL", "Conexão de Banco de Dados SQL")

func (c *ConexaoSQL) Tipo() *ptst.Tipo {
	return TipoConexaoSQL
}

func (c *ConexaoSQL) M__obtem_attributo__(nome string) (ptst.Objeto, error) {
	switch nome {
	case "executar":
		return ptst.NewMetodoOuPanic("executar", func(inst ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
			if len(args) < 1 {
				return nil, ptst.NewErroF(ptst.TipagemErro, "executar esperava no mínimo 1 argumento (query)")
			}
			query, err := ptst.NewTexto(args[0])
			if err != nil {
				return nil, err
			}
			var goArgs []interface{}
			for _, arg := range args[1:] {
				goArgs = append(goArgs, toGoType(arg))
			}
			_, errExec := c.db.Exec(string(query.(ptst.Texto)), goArgs...)
			if errExec != nil {
				return nil, ptst.NewErroF(ptst.ErroDeSistema, "Erro ao executar query SQL: %v", errExec)
			}
			return ptst.Nulo, nil
		}, ""), nil

	case "consultar":
		return ptst.NewMetodoOuPanic("consultar", func(inst ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
			if len(args) < 1 {
				return nil, ptst.NewErroF(ptst.TipagemErro, "consultar esperava no mínimo 1 argumento (query)")
			}
			query, err := ptst.NewTexto(args[0])
			if err != nil {
				return nil, err
			}
			var goArgs []interface{}
			for _, arg := range args[1:] {
				goArgs = append(goArgs, toGoType(arg))
			}
			rows, errQuery := c.db.Query(string(query.(ptst.Texto)), goArgs...)
			if errQuery != nil {
				return nil, ptst.NewErroF(ptst.ErroDeSistema, "Erro ao executar consulta SQL: %v", errQuery)
			}
			defer rows.Close()

			cols, errCols := rows.Columns()
			if errCols != nil {
				return nil, ptst.NewErroF(ptst.ErroDeSistema, "Erro ao obter colunas: %v", errCols)
			}

			lista := &ptst.Lista{}
			for rows.Next() {
				columns := make([]interface{}, len(cols))
				columnPointers := make([]interface{}, len(cols))
				for i := range columns {
					columnPointers[i] = &columns[i]
				}

				if errScan := rows.Scan(columnPointers...); errScan != nil {
					return nil, ptst.NewErroF(ptst.ErroDeSistema, "Erro ao escanear linha: %v", errScan)
				}

				mapaLinha := ptst.NewMapaVazio()
				for i, colName := range cols {
					val := columns[i]
					mapaLinha.M__define_item__(ptst.Texto(colName), toPtObject(val))
				}
				lista.Adiciona(mapaLinha)
			}

			return lista, nil
		}, ""), nil

	case "tabela":
		return ptst.NewMetodoOuPanic("tabela", func(inst ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
			if err := ptst.VerificaNumeroArgumentos("tabela", false, args, 1, 2); err != nil {
				return nil, err
			}
			tabelaTexto, err := ptst.NewTexto(args[0])
			if err != nil {
				return nil, err
			}
			var schema ptst.Mapa
			if len(args) == 2 {
				s, ok := args[1].(ptst.Mapa)
				if !ok {
					return nil, ptst.NewErroF(ptst.TipagemErro, "tabela esperava um Mapa como segundo argumento (schema)")
				}
				schema = s
			}
			return &QueryBuilder{
				conexao: c,
				tabela:  string(tabelaTexto.(ptst.Texto)),
				schema:  schema,
			}, nil
		}, ""), nil

	case "fechar":
		return ptst.NewMetodoOuPanic("fechar", func(inst ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
			c.db.Close()
			return ptst.Nulo, nil
		}, ""), nil
	}

	return nil, ptst.NewErroF(ptst.AtributoErro, "Atributo '%s' não existe em ConexaoSQL", nome)
}

type QueryBuilder struct {
	conexao      *ConexaoSQL
	tabela       string
	selecionados []string
	condicoes    []string
	args         []ptst.Objeto
	limiteVal    int
	schema       ptst.Mapa
}

var TipoQueryBuilder = ptst.NewTipo("QueryBuilder", "Query Builder dinâmico")

func (q *QueryBuilder) Tipo() *ptst.Tipo {
	return TipoQueryBuilder
}

func (q *QueryBuilder) M__obtem_attributo__(nome string) (ptst.Objeto, error) {
	switch nome {
	case "selecionar":
		return ptst.NewMetodoOuPanic("selecionar", func(inst ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
			for _, arg := range args {
				col, err := ptst.NewTexto(arg)
				if err != nil {
					return nil, err
				}
				q.selecionados = append(q.selecionados, string(col.(ptst.Texto)))
			}
			return q, nil
		}, ""), nil

	case "onde":
		return ptst.NewMetodoOuPanic("onde", func(inst ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
			if err := ptst.VerificaNumeroArgumentos("onde", false, args, 3, 3); err != nil {
				return nil, err
			}
			coluna, err := ptst.NewTexto(args[0])
			if err != nil {
				return nil, err
			}
			operador, err := ptst.NewTexto(args[1])
			if err != nil {
				return nil, err
			}
			q.condicoes = append(q.condicoes, fmt.Sprintf("%s %s ?", string(coluna.(ptst.Texto)), string(operador.(ptst.Texto))))
			q.args = append(q.args, args[2])
			return q, nil
		}, ""), nil

	case "limite":
		return ptst.NewMetodoOuPanic("limite", func(inst ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
			if err := ptst.VerificaNumeroArgumentos("limite", false, args, 1, 1); err != nil {
				return nil, err
			}
			n, err := ptst.NewInteiro(args[0])
			if err != nil {
				return nil, err
			}
			q.limiteVal = int(n.(ptst.Inteiro))
			return q, nil
		}, ""), nil

	case "obterMuitos":
		return ptst.NewMetodoOuPanic("obterMuitos", func(inst ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
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

			var callArgs ptst.Tupla
			callArgs = append(callArgs, ptst.Texto(query))
			for _, arg := range q.args {
				callArgs = append(callArgs, arg)
			}

			consultarMetodo, err := q.conexao.M__obtem_attributo__("consultar")
			if err != nil {
				return nil, err
			}
			return ptst.Chamar(consultarMetodo, callArgs)
		}, ""), nil

	case "obterUm":
		return ptst.NewMetodoOuPanic("obterUm", func(inst ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
			q.limiteVal = 1
			obterMuitosMetodo, err := q.M__obtem_attributo__("obterMuitos")
			if err != nil {
				return nil, err
			}
			res, err := ptst.Chamar(obterMuitosMetodo, ptst.Tupla{})
			if err != nil {
				return nil, err
			}
			lista := res.(*ptst.Lista)
			if len(lista.Itens) == 0 {
				return ptst.Nulo, nil
			}
			return lista.Itens[0], nil
		}, ""), nil

	case "inserir":
		return ptst.NewMetodoOuPanic("inserir", func(inst ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
			if err := ptst.VerificaNumeroArgumentos("inserir", false, args, 1, 1); err != nil {
				return nil, err
			}
			mapa, ok := args[0].(ptst.Mapa)
			if !ok {
				return nil, ptst.NewErroF(ptst.TipagemErro, "inserir esperava um Mapa como argumento")
			}

			if q.schema != nil {
				for k, v := range mapa {
					tipoEsperado, existe := q.schema[k]
					if !existe {
						return nil, ptst.NewErroF(ptst.ValorErro, "campo '%s' nao existe no schema da tabela '%s'", k, q.tabela)
					}
					tipoTexto, ok := tipoEsperado.(ptst.Texto)
					if ok {
						if errVal := validarTipoCampo(string(tipoTexto), v); errVal != nil {
							return nil, ptst.NewErroF(ptst.TipagemErro, "campo '%s' da tabela '%s': %v", k, q.tabela, errVal)
						}
					}
				}
			}

			var colunas []string
			var placeholders []string
			var valueArgs ptst.Tupla

			for k, v := range mapa {
				colunas = append(colunas, k)
				placeholders = append(placeholders, "?")
				valueArgs = append(valueArgs, v)
			}

			query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", q.tabela, strings.Join(colunas, ", "), strings.Join(placeholders, ", "))
			query = q.converterPlaceholders(query)

			var callArgs ptst.Tupla
			callArgs = append(callArgs, ptst.Texto(query))
			for _, arg := range valueArgs {
				callArgs = append(callArgs, arg)
			}

			executarMetodo, err := q.conexao.M__obtem_attributo__("executar")
			if err != nil {
				return nil, err
			}
			return ptst.Chamar(executarMetodo, callArgs)
		}, ""), nil

	case "atualizar":
		return ptst.NewMetodoOuPanic("atualizar", func(inst ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
			if err := ptst.VerificaNumeroArgumentos("atualizar", false, args, 1, 1); err != nil {
				return nil, err
			}
			mapa, ok := args[0].(ptst.Mapa)
			if !ok {
				return nil, ptst.NewErroF(ptst.TipagemErro, "atualizar esperava um Mapa como argumento")
			}

			var sets []string
			var valueArgs ptst.Tupla

			for k, v := range mapa {
				sets = append(sets, fmt.Sprintf("%s = ?", k))
				valueArgs = append(valueArgs, v)
			}

			query := fmt.Sprintf("UPDATE %s SET %s", q.tabela, strings.Join(sets, ", "))
			if len(q.condicoes) > 0 {
				query += " WHERE " + strings.Join(q.condicoes, " AND ")
			}
			query = q.converterPlaceholders(query)

			var callArgs ptst.Tupla
			callArgs = append(callArgs, ptst.Texto(query))
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
			return ptst.Chamar(executarMetodo, callArgs)
		}, ""), nil

	case "deletar":
		return ptst.NewMetodoOuPanic("deletar", func(inst ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
			query := fmt.Sprintf("DELETE FROM %s", q.tabela)
			if len(q.condicoes) > 0 {
				query += " WHERE " + strings.Join(q.condicoes, " AND ")
			}
			query = q.converterPlaceholders(query)

			var callArgs ptst.Tupla
			callArgs = append(callArgs, ptst.Texto(query))
			for _, arg := range q.args {
				callArgs = append(callArgs, arg)
			}

			executarMetodo, err := q.conexao.M__obtem_attributo__("executar")
			if err != nil {
				return nil, err
			}
			return ptst.Chamar(executarMetodo, callArgs)
		}, ""), nil
	}

	return nil, ptst.NewErroF(ptst.AtributoErro, "Atributo '%s' não existe no QueryBuilder", nome)
}

func toGoType(obj ptst.Objeto) interface{} {
	if obj == ptst.Nulo || obj.Tipo() == ptst.TipoNulo {
		return nil
	}
	switch v := obj.(type) {
	case ptst.Texto:
		return string(v)
	case ptst.Inteiro:
		return int64(v)
	case ptst.Decimal:
		return float64(v)
	case ptst.Booleano:
		return bool(v)
	}
	return fmt.Sprintf("%v", obj)
}

func toPtObject(val interface{}) ptst.Objeto {
	if val == nil {
		return ptst.Nulo
	}
	switch v := val.(type) {
	case string:
		return ptst.Texto(v)
	case []byte:
		return ptst.Texto(v)
	case int64:
		return ptst.Inteiro(v)
	case int:
		return ptst.Inteiro(v)
	case float64:
		return ptst.Decimal(v)
	case bool:
		return ptst.Booleano(v)
	}
	return ptst.Texto(fmt.Sprintf("%v", val))
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

func validarTipoCampo(tipoEsperado string, valor ptst.Objeto) error {
	switch tipoEsperado {
	case "texto":
		if _, ok := valor.(ptst.Texto); !ok {
			return fmt.Errorf("esperava tipo Texto, obteve %s", valor.Tipo().Nome)
		}
	case "inteiro":
		if _, ok := valor.(ptst.Inteiro); !ok {
			return fmt.Errorf("esperava tipo Inteiro, obteve %s", valor.Tipo().Nome)
		}
	case "decimal":
		if _, ok := valor.(ptst.Decimal); !ok {
			return fmt.Errorf("esperava tipo Decimal, obteve %s", valor.Tipo().Nome)
		}
	case "booleano":
		if _, ok := valor.(ptst.Booleano); !ok {
			return fmt.Errorf("esperava tipo Booleano, obteve %s", valor.Tipo().Nome)
		}
	}
	return nil
}
