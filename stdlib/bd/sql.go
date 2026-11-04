// Package bd implementa os conectores, gerenciadores e adaptadores de acesso
// a bancos de dados relacionais (SQL), não-relacionais (NoSQL) e vetoriais do Harpia.
package bd

import (
	"database/sql"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/mat-dgruber/Harpia/hrp"
)

var rxIdentificadorSQL = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*(\.[a-zA-Z_][a-zA-Z0-9_]*)?$`)

func validarIdentificadorSQL(id string) error {
	id = strings.TrimSpace(id)
	if id == "*" {
		return nil
	}
	if !rxIdentificadorSQL.MatchString(id) {
		return hrp.NewErroF(hrp.ValorErro, "Identificador SQL inválido: '%s'", id)
	}
	return nil
}

var operadoresSegurosSQL = map[string]bool{
	"=": true, "==": true, "!=": true, "<>": true, "<": true, ">": true, "<=": true, ">=": true,
	"LIKE": true, "like": true, "NOT LIKE": true, "not like": true, "IS": true, "is": true, "IS NOT": true, "is not": true,
}

// ehInsert informa se a query é um INSERT (para decidir se LastInsertId é significativo).
func ehInsert(query string) bool {
	return strings.HasPrefix(strings.ToUpper(strings.TrimSpace(query)), "INSERT")
}

func validarOperadorSQL(op string) error {
	op = strings.TrimSpace(op)
	if !operadoresSegurosSQL[op] {
		return hrp.NewErroF(hrp.ValorErro, "Operador SQL inválido ou inseguro: '%s'", op)
	}
	return nil
}

func validarDirecaoSQL(dir string) (string, error) {
	d := strings.ToUpper(strings.TrimSpace(dir))
	if d == "" || d == "ASC" {
		return "ASC", nil
	}
	if d == "DESC" {
		return "DESC", nil
	}
	return "", hrp.NewErroF(hrp.ValorErro, "Direção SQL inválida: '%s' (use ASC ou DESC)", dir)
}

// MaxLimiteRegistrosSQL impõe um teto para consultas do QueryBuilder para mitigar exaustão de recursos.
const MaxLimiteRegistrosSQL = 1000000

// ConexaoSQL gerencia a pool de conexões físicas para bancos relacionais como SQLite, Postgres e MySQL.
type ConexaoSQL struct {
	db     *sql.DB
	driver string
}

// TipoConexaoSQL define a assinatura e o tipo da classe ConexaoSQL na VM.
var TipoConexaoSQL = hrp.NewTipo("ConexaoSQL", "Conexão de Banco de Dados SQL")

// Tipo retorna a representação na VM.
func (c *ConexaoSQL) Tipo() *hrp.Tipo {
	return TipoConexaoSQL
}

// M__obtem_attributo__ mapeia e despacha as chamadas para operações e transações no banco SQL.
func (c *ConexaoSQL) M__obtem_attributo__(nome string) (hrp.Objeto, error) {
	switch nome {
	case "executar":
		// Executa instruções SQL que não retornam linhas (ex: INSERT, UPDATE, DELETE).
		return hrp.NewMetodoOuPanic("executar", func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
			if len(args) < 1 {
				return nil, hrp.NewErroF(hrp.TipagemErro, "executar esperava no mínimo 1 argumento (query)")
			}
			query, err := hrp.NewTexto(args[0])
			if err != nil {
				return nil, err
			}
			goArgs := converterArgumentosParaGo(args)
			resExec, errExec := c.db.Exec(string(query.(hrp.Texto)), goArgs...)
			if errExec != nil {
				return nil, hrp.NewErroF(hrp.ErroDeSistema, "Erro ao executar query SQL: %v", errExec)
			}
			// LastInsertId só faz sentido em INSERT: no SQLite, um UPDATE devolveria o id do último insert
			if ehInsert(string(query.(hrp.Texto))) {
				if lastID, errID := resExec.LastInsertId(); errID == nil && lastID > 0 {
					return hrp.Inteiro(lastID), nil
				}
			}
			return hrp.Nulo, nil
		}, "Executa uma query de alteração de dados no banco (INSERT/UPDATE/DELETE)."), nil

	case "executarDetalhado":
		// Como executar, mas devolve {idInserido, linhasAfetadas} (Nulo quando o driver não informa).
		return hrp.NewMetodoOuPanic("executarDetalhado", func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
			if len(args) < 1 {
				return nil, hrp.NewErroF(hrp.TipagemErro, "executarDetalhado esperava no mínimo 1 argumento (query)")
			}
			query, err := hrp.NewTexto(args[0])
			if err != nil {
				return nil, err
			}
			resExec, errExec := c.db.Exec(string(query.(hrp.Texto)), converterArgumentosParaGo(args)...)
			if errExec != nil {
				return nil, hrp.NewErroF(hrp.ErroDeSistema, "Erro ao executar query SQL: %v", errExec)
			}
			res := hrp.Mapa{"idInserido": hrp.Nulo, "linhasAfetadas": hrp.Nulo}
			if ehInsert(string(query.(hrp.Texto))) {
				if id, errID := resExec.LastInsertId(); errID == nil {
					res["idInserido"] = hrp.Inteiro(id)
				}
			}
			if n, errN := resExec.RowsAffected(); errN == nil {
				res["linhasAfetadas"] = hrp.Inteiro(n)
			}
			return res, nil
		}, "Executa INSERT/UPDATE/DELETE e devolve {idInserido, linhasAfetadas}."), nil

	case "consultar":
		// Executa instruções SQL de busca estruturada que retornam linhas (SELECT).
		return hrp.NewMetodoOuPanic("consultar", func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
			if len(args) < 1 {
				return nil, hrp.NewErroF(hrp.TipagemErro, "consultar esperava no mínimo 1 argumento (query)")
			}
			query, err := hrp.NewTexto(args[0])
			if err != nil {
				return nil, err
			}
			goArgs := converterArgumentosParaGo(args)
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

			if errIter := rows.Err(); errIter != nil {
				return nil, hrp.NewErroF(hrp.ErroDeSistema, "Erro ao iterar resultados SQL: %v", errIter)
			}

			return lista, nil
		}, "Executa uma query SQL de seleção (SELECT) e devolve uma Lista de Mapas."), nil

	case "tabela":
		// Retorna um QueryBuilder fluído e opcionalmente tipado por schema para interações SQL abstratas.
		return hrp.NewMetodoOuPanic("tabela", func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
			if err := hrp.VerificaNumeroArgumentos("tabela", false, args, 1, 2); err != nil {
				return nil, err
			}
			tabelaTexto, err := hrp.NewTexto(args[0])
			if err != nil {
				return nil, err
			}
			nomeTab := string(tabelaTexto.(hrp.Texto))
			if errVal := validarIdentificadorSQL(nomeTab); errVal != nil {
				return nil, errVal
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
				tabela:  nomeTab,
				schema:  schema,
			}, nil
		}, "Prepara uma tabela para uso com QueryBuilder e validações de esquema opcional."), nil

	case "transacao":
		// Abre e gerencia uma transação atômica (Tx) que garante consistência (ACID) em falhas.
		return hrp.NewMetodoOuPanic("transacao", func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
			if err := hrp.VerificaNumeroArgumentos("transacao", false, args, 1, 1); err != nil {
				return nil, err
			}
			fnCallback := args[0]

			tx, errBegin := c.db.Begin()
			if errBegin != nil {
				return nil, hrp.NewErroF(hrp.ErroDeSistema, "Erro ao iniciar transação: %v", errBegin)
			}

			txObj := &TransacaoSQL{tx: tx, driver: c.driver}
			// pânico Go no callback: desfaz antes de propagar, senão a conexão fica presa na transação
			defer func() {
				if r := recover(); r != nil {
					_ = tx.Rollback()
					panic(r)
				}
			}()
			_, errExec := hrp.Chamar(fnCallback, hrp.Tupla{txObj})
			if errExec != nil {
				tx.Rollback()
				return nil, errExec
			}

			if errCommit := tx.Commit(); errCommit != nil {
				return nil, hrp.NewErroF(hrp.ErroDeSistema, "Erro ao efetivar (commit) transação: %v", errCommit)
			}

			return hrp.Nulo, nil
		}, "Executa uma transação atômica em banco de dados; executa rollback automático se houver falha."), nil

	case "fechar":
		// Encerra a pool de conexões do banco de dados relacional.
		return hrp.NewMetodoOuPanic("fechar", func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
			c.db.Close()
			return hrp.Nulo, nil
		}, "Fecha o pool de conexões ativas com o banco de dados SQL."), nil
	}

	return nil, hrp.NewErroF(hrp.AtributoErro, "Atributo '%s' não existe em ConexaoSQL", nome)
}

// TransacaoSQL representa uma transação ativa em um banco de dados relacional.
type TransacaoSQL struct {
	tx     *sql.Tx
	driver string
}

// TipoTransacaoSQL define a classe da Transação.
var TipoTransacaoSQL = hrp.NewTipo("TransacaoSQL", "Transação de Banco de Dados SQL")

// Tipo retorna a representação interna.
func (t *TransacaoSQL) Tipo() *hrp.Tipo {
	return TipoTransacaoSQL
}

// M__obtem_attributo__ mapeia os comandos transacionais (executar/consultar) sob escopo controlado da transação.
func (t *TransacaoSQL) M__obtem_attributo__(nome string) (hrp.Objeto, error) {
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
			goArgs := converterArgumentosParaGo(args)
			_, errExec := t.tx.Exec(string(query.(hrp.Texto)), goArgs...)
			if errExec != nil {
				return nil, hrp.NewErroF(hrp.ErroDeSistema, "Erro ao executar query SQL na transação: %v", errExec)
			}
			return hrp.Nulo, nil
		}, "Executa query transacional de modificação de dados."), nil

	case "consultar":
		return hrp.NewMetodoOuPanic("consultar", func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
			if len(args) < 1 {
				return nil, hrp.NewErroF(hrp.TipagemErro, "consultar esperava no mínimo 1 argumento (query)")
			}
			query, err := hrp.NewTexto(args[0])
			if err != nil {
				return nil, err
			}
			goArgs := converterArgumentosParaGo(args)
			rows, errQuery := t.tx.Query(string(query.(hrp.Texto)), goArgs...)
			if errQuery != nil {
				return nil, hrp.NewErroF(hrp.ErroDeSistema, "Erro ao executar consulta SQL na transação: %v", errQuery)
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
					return nil, hrp.NewErroF(hrp.ErroDeSistema, "Erro ao ler linha: %v", errScan)
				}
				m := hrp.NewMapaVazio()
				for i, colName := range cols {
					val := columns[i]
					m.M__define_item__(hrp.Texto(colName), toPtObject(val))
				}
				lista.Itens = append(lista.Itens, m)
			}

			if errIter := rows.Err(); errIter != nil {
				return nil, hrp.NewErroF(hrp.ErroDeSistema, "Erro ao iterar resultados SQL na transação: %v", errIter)
			}

			return lista, nil
		}, "Executa consulta transacional de dados."), nil
	}

	return nil, hrp.NewErroF(hrp.AtributoErro, "Atributo '%s' não existe em TransacaoSQL", nome)
}

// QueryBuilder provê montagem dinâmica e flexível de queries SQL fluídas contra injeção SQL.
type QueryBuilder struct {
	conexao      *ConexaoSQL
	tabela       string
	selecionados []string
	condicoes    []string
	args         []hrp.Objeto
	ordenacoes   []string
	limiteVal    int
	schema       hrp.Mapa
	ultimoFilho  *QueryBuilder
	numFilhos    int
}

// TipoQueryBuilder mapeia a definição na VM.
var TipoQueryBuilder = hrp.NewTipo("QueryBuilder", "Query Builder dinâmico")

// Tipo retorna a representação correspondente na VM.
func (q *QueryBuilder) Tipo() *hrp.Tipo {
	return TipoQueryBuilder
}

func (q *QueryBuilder) clone() *QueryBuilder {
	novo := &QueryBuilder{
		conexao:   q.conexao,
		tabela:    q.tabela,
		limiteVal: q.limiteVal,
		schema:    q.schema,
	}
	if len(q.selecionados) > 0 {
		novo.selecionados = append([]string(nil), q.selecionados...)
	}
	if len(q.condicoes) > 0 {
		novo.condicoes = append([]string(nil), q.condicoes...)
	}
	if len(q.args) > 0 {
		novo.args = append([]hrp.Objeto(nil), q.args...)
	}
	if len(q.ordenacoes) > 0 {
		novo.ordenacoes = append([]string(nil), q.ordenacoes...)
	}
	return novo
}

func (q *QueryBuilder) registrarFilho(filho *QueryBuilder) {
	if q.numFilhos == 0 {
		q.ultimoFilho = filho
	} else {
		q.ultimoFilho = nil
	}
	q.numFilhos++
}

func (q *QueryBuilder) alvoExecucao() *QueryBuilder {
	if q.ultimoFilho != nil {
		return q.ultimoFilho.alvoExecucao()
	}
	return q
}

func (q *QueryBuilder) validarSchemaEIdentificadores() error {
	if err := validarIdentificadorSQL(q.tabela); err != nil {
		return err
	}
	if q.schema != nil {
		for colSchema := range q.schema {
			if err := validarIdentificadorSQL(colSchema); err != nil {
				return err
			}
		}
	}
	return nil
}

func (q *QueryBuilder) selecionar(args hrp.Tupla) (hrp.Objeto, error) {
	novo := q.clone()
	var extrairCols func(hrp.Objeto) error
	extrairCols = func(arg hrp.Objeto) error {
		if lista, ok := arg.(*hrp.Lista); ok {
			for _, item := range lista.Itens {
				if err := extrairCols(item); err != nil {
					return err
				}
			}
			return nil
		}
		if tupla, ok := arg.(hrp.Tupla); ok {
			for _, item := range tupla {
				if err := extrairCols(item); err != nil {
					return err
				}
			}
			return nil
		}
		col, err := hrp.NewTexto(arg)
		if err != nil {
			return err
		}
		colStr := string(col.(hrp.Texto))
		if errVal := validarIdentificadorSQL(colStr); errVal != nil {
			return errVal
		}
		novo.selecionados = append(novo.selecionados, colStr)
		return nil
	}
	for _, arg := range args {
		if err := extrairCols(arg); err != nil {
			return nil, err
		}
	}
	q.registrarFilho(novo)
	return novo, nil
}

func (q *QueryBuilder) onde(args hrp.Tupla) (hrp.Objeto, error) {
	if err := hrp.VerificaNumeroArgumentos("onde", false, args, 3, 3); err != nil {
		return nil, err
	}
	coluna, err := hrp.NewTexto(args[0])
	if err != nil {
		return nil, err
	}
	colStr := string(coluna.(hrp.Texto))
	if errVal := validarIdentificadorSQL(colStr); errVal != nil {
		return nil, errVal
	}
	operador, err := hrp.NewTexto(args[1])
	if err != nil {
		return nil, err
	}
	opStr := string(operador.(hrp.Texto))
	if errVal := validarOperadorSQL(opStr); errVal != nil {
		return nil, errVal
	}
	novo := q.clone()
	if opStr == "==" { // sintaxe da linguagem (usada na documentação) mapeada para SQL
		opStr = "="
	}
	novo.condicoes = append(novo.condicoes, fmt.Sprintf("%s %s ?", colStr, opStr))
	novo.args = append(novo.args, args[2])
	q.registrarFilho(novo)
	return novo, nil
}

func (q *QueryBuilder) ordenarPor(args hrp.Tupla) (hrp.Objeto, error) {
	if err := hrp.VerificaNumeroArgumentos("ordenarPor", false, args, 1, 2); err != nil {
		return nil, err
	}
	col, err := hrp.NewTexto(args[0])
	if err != nil {
		return nil, err
	}
	colStr := string(col.(hrp.Texto))
	if errVal := validarIdentificadorSQL(colStr); errVal != nil {
		return nil, errVal
	}
	dirStr := "ASC"
	if len(args) == 2 {
		dir, errDir := hrp.NewTexto(args[1])
		if errDir != nil {
			return nil, errDir
		}
		d, errValDir := validarDirecaoSQL(string(dir.(hrp.Texto)))
		if errValDir != nil {
			return nil, errValDir
		}
		dirStr = d
	}
	novo := q.clone()
	novo.ordenacoes = append(novo.ordenacoes, fmt.Sprintf("%s %s", colStr, dirStr))
	q.registrarFilho(novo)
	return novo, nil
}

func (q *QueryBuilder) limite(args hrp.Tupla) (hrp.Objeto, error) {
	if err := hrp.VerificaNumeroArgumentos("limite", false, args, 1, 1); err != nil {
		return nil, err
	}
	n, err := hrp.NewInteiro(args[0])
	if err != nil {
		return nil, err
	}
	nVal := int64(n.(hrp.Inteiro))
	if nVal < 0 || nVal > MaxLimiteRegistrosSQL {
		return nil, hrp.NewErroF(hrp.ValorErro, "limite de registros SQL inválido (deve ser entre 0 e %d)", MaxLimiteRegistrosSQL)
	}
	novo := q.clone()
	novo.limiteVal = int(nVal)
	q.registrarFilho(novo)
	return novo, nil
}

func (q *QueryBuilder) obterMuitos(args hrp.Tupla) (hrp.Objeto, error) {
	if alvo := q.alvoExecucao(); alvo != q {
		return alvo.obterMuitos(args)
	}
	if err := q.validarSchemaEIdentificadores(); err != nil {
		return nil, err
	}
	if q.conexao == nil {
		return nil, hrp.NewErroF(hrp.ErroDeSistema, "Conexão de banco de dados não inicializada no QueryBuilder")
	}
	cols := "*"
	if len(q.selecionados) > 0 {
		cols = strings.Join(q.selecionados, ", ")
	}
	query := fmt.Sprintf("SELECT %s FROM %s", cols, q.tabela)
	if len(q.condicoes) > 0 {
		query += " WHERE " + strings.Join(q.condicoes, " AND ")
	}
	if len(q.ordenacoes) > 0 {
		query += " ORDER BY " + strings.Join(q.ordenacoes, ", ")
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
}

func (q *QueryBuilder) obterUm(args hrp.Tupla) (hrp.Objeto, error) {
	if alvo := q.alvoExecucao(); alvo != q {
		return alvo.obterUm(args)
	}
	qComLimite := q.clone()
	qComLimite.limiteVal = 1
	res, err := qComLimite.obterMuitos(hrp.Tupla{})
	if err != nil {
		return nil, err
	}
	lista, ok := res.(*hrp.Lista)
	if !ok || len(lista.Itens) == 0 {
		return hrp.Nulo, nil
	}
	return lista.Itens[0], nil
}

func (q *QueryBuilder) inserir(args hrp.Tupla) (hrp.Objeto, error) {
	if alvo := q.alvoExecucao(); alvo != q {
		return alvo.inserir(args)
	}
	if err := hrp.VerificaNumeroArgumentos("inserir", false, args, 1, 1); err != nil {
		return nil, err
	}
	if err := q.validarSchemaEIdentificadores(); err != nil {
		return nil, err
	}
	mapa, ok := args[0].(hrp.Mapa)
	if !ok {
		return nil, hrp.NewErroF(hrp.TipagemErro, "inserir esperava um Mapa como argumento")
	}

	colunas := make([]string, 0, len(mapa))
	for k := range mapa {
		if err := validarIdentificadorSQL(k); err != nil {
			return nil, err
		}
		colunas = append(colunas, k)
	}
	sort.Strings(colunas)

	if q.schema != nil {
		for _, k := range colunas {
			v := mapa[k]
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

	if q.conexao == nil {
		return nil, hrp.NewErroF(hrp.ErroDeSistema, "Conexão de banco de dados não inicializada no QueryBuilder")
	}

	placeholders := make([]string, len(colunas))
	valueArgs := make(hrp.Tupla, len(colunas))
	for i, col := range colunas {
		placeholders[i] = "?"
		valueArgs[i] = mapa[col]
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
}

func (q *QueryBuilder) atualizar(args hrp.Tupla) (hrp.Objeto, error) {
	if alvo := q.alvoExecucao(); alvo != q {
		return alvo.atualizar(args)
	}
	if err := hrp.VerificaNumeroArgumentos("atualizar", false, args, 1, 1); err != nil {
		return nil, err
	}
	if err := q.validarSchemaEIdentificadores(); err != nil {
		return nil, err
	}
	mapa, ok := args[0].(hrp.Mapa)
	if !ok {
		return nil, hrp.NewErroF(hrp.TipagemErro, "atualizar esperava um Mapa como argumento")
	}

	colunas := make([]string, 0, len(mapa))
	for k := range mapa {
		if err := validarIdentificadorSQL(k); err != nil {
			return nil, err
		}
		colunas = append(colunas, k)
	}
	sort.Strings(colunas)

	if q.conexao == nil {
		return nil, hrp.NewErroF(hrp.ErroDeSistema, "Conexão de banco de dados não inicializada no QueryBuilder")
	}

	sets := make([]string, len(colunas))
	valueArgs := make(hrp.Tupla, len(colunas))
	for i, col := range colunas {
		sets[i] = fmt.Sprintf("%s = ?", col)
		valueArgs[i] = mapa[col]
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
}

func (q *QueryBuilder) deletar(args hrp.Tupla) (hrp.Objeto, error) {
	if alvo := q.alvoExecucao(); alvo != q {
		return alvo.deletar(args)
	}
	if err := q.validarSchemaEIdentificadores(); err != nil {
		return nil, err
	}
	if q.conexao == nil {
		return nil, hrp.NewErroF(hrp.ErroDeSistema, "Conexão de banco de dados não inicializada no QueryBuilder")
	}
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
}

// M__obtem_attributo__ expõe os métodos fluídos da API do QueryBuilder.
func (q *QueryBuilder) M__obtem_attributo__(nome string) (hrp.Objeto, error) {
	switch nome {
	case "selecionar":
		// Define quais colunas serão selecionadas na query SELECT final.
		return hrp.NewMetodoOuPanic("selecionar", func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
			return q.selecionar(args)
		}, "Seleciona um subconjunto de colunas físicas para a query."), nil

	case "onde":
		// Adiciona cláusulas de filtro WHERE utilizando parâmetros de substituição (bindings) seguros.
		return hrp.NewMetodoOuPanic("onde", func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
			return q.onde(args)
		}, "Adiciona uma restrição condicional de seleção no filtro (WHERE)."), nil

	case "ordenarPor", "ordenar":
		// Define a ordenação dos resultados (ORDER BY).
		return hrp.NewMetodoOuPanic(nome, func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
			return q.ordenarPor(args)
		}, "Define a ordenação dos resultados (ORDER BY)."), nil

	case "limite":
		// Delimita o número máximo de registros a serem trazidos na query, respeitando MaxLimiteRegistrosSQL.
		return hrp.NewMetodoOuPanic("limite", func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
			return q.limite(args)
		}, "Limita o volume máximo de linhas da resposta."), nil

	case "obterMuitos":
		// Conclui e roda a query SELECT coletando uma lista com todas as correspondências do banco.
		return hrp.NewMetodoOuPanic("obterMuitos", func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
			return q.obterMuitos(args)
		}, "Executa a busca retornando todas as linhas correspondentes."), nil

	case "obterUm":
		// Atalho inteligente que impõe limite=1 e devolve o primeiro registro encontrado ou Nulo.
		return hrp.NewMetodoOuPanic("obterUm", func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
			return q.obterUm(args)
		}, "Retorna apenas o primeiro registro localizado."), nil

	case "inserir":
		// Insere dados no banco, validando tipos de campos dinamicamente se houver schema associado à tabela.
		return hrp.NewMetodoOuPanic("inserir", func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
			return q.inserir(args)
		}, "Insere um registro e roda validações estritas de esquema."), nil

	case "atualizar":
		// Modifica dados existentes no banco que batem com as condições da cláusula onde().
		return hrp.NewMetodoOuPanic("atualizar", func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
			return q.atualizar(args)
		}, "Atualiza registros em massa com base nas condições estabelecidas no onde()."), nil

	case "deletar":
		// Deleta registros correspondentes do banco de dados físico.
		return hrp.NewMetodoOuPanic("deletar", func(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
			return q.deletar(args)
		}, "Deleta linhas físicas filtradas pelas regras do onde()."), nil
	}

	return nil, hrp.NewErroF(hrp.AtributoErro, "Atributo '%s' não existe no QueryBuilder", nome)
}

// toGoType mapeia tipos nativos da VM Harpia para seus equivalentes Go de baixo nível.
func toGoType(obj hrp.Objeto) interface{} {
	if obj == nil || obj == hrp.Nulo || (obj.Tipo() != nil && obj.Tipo() == hrp.TipoNulo) {
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

// toPtObject converte tipos nativos retornados por drivers de banco Go (database/sql) em Objetos Harpia.
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

// converterPlaceholders ajusta dinamicamente a numeração de placeholders para drivers como Postgres ($1, $2, etc.).
func (q *QueryBuilder) converterPlaceholders(query string) string {
	if q.conexao == nil || q.conexao.driver != "postgres" {
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

// validarTipoCampo valida em tempo de execução os dados inseridos contra a definição do ORM do Harpia.
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

func converterArgumentosParaGo(args hrp.Tupla) []interface{} {
	var goArgs []interface{}
	if len(args) == 2 {
		if lista, ok := args[1].(*hrp.Lista); ok {
			for _, item := range lista.Itens {
				goArgs = append(goArgs, toGoType(item))
			}
			return goArgs
		}
		if tupla, ok := args[1].(hrp.Tupla); ok {
			for _, item := range tupla {
				goArgs = append(goArgs, toGoType(item))
			}
			return goArgs
		}
	}
	for _, arg := range args[1:] {
		goArgs = append(goArgs, toGoType(arg))
	}
	return goArgs
}
