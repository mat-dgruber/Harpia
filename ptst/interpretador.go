package ptst

import (
	"github.com/natanfeitosa/portuscript/lexer"
	"github.com/natanfeitosa/portuscript/parser"
)

// Interpretador representa o motor avaliador (Visitor) de nós da AST (Árvore de Sintaxe Abstrata) do Portuscript.
//
// O Interpretador varre os nós de forma recursiva, mantendo referências do contexto global,
// o escopo léxico corrente, buffers de valor de retorno e os metadados de posições físicas do código-fonte.
type Interpretador struct {
	Ast          parser.BaseNode                  // O nó sintático a ser analisado de forma macro.
	Contexto     *Contexto                        // Ponteiro para o supervisor global da VM.
	Escopo       *Escopo                          // O escopo léxico corrente sob execução.
	ValorRetorno Objeto                           // Buffer temporário para propagação lógica do valor de retorno ('retorne').
	Arquivo      string                           // Caminho lógico do arquivo em execução.
	Codigo       string                           // O código-fonte integral sob avaliação.
	Posicoes     map[parser.BaseNode]*lexer.Token // Mapa para decodificação de tracebacks geográficos de erro.
}

// entrarNoEscopo migra o contexto operacional do interpretador para uma nova tabela de símbolos (Escopo).
// Enlaça o escopo pai de forma a criar o isolamento léxico adequado.
func (i *Interpretador) entrarNoEscopo(escopo *Escopo) {
	if escopo == nil {
		escopo = i.Escopo.NewEscopo()
	}

	if escopo.Pai == nil && i.Escopo != nil {
		escopo.Pai = i.Escopo
	}

	i.Escopo = escopo
}

// sairDoEscopo restaura as referências operacionais de volta ao escopo hierárquico superior (Pai).
func (i *Interpretador) sairDoEscopo() {
	i.Escopo = i.Escopo.Pai
}

// Inicializa consome a raiz da AST do script (um Programa ou Bloco), iniciando a varredura linear do Visitor.
func (i *Interpretador) Inicializa() (Objeto, error) {
	var declaracoes []parser.BaseNode

	switch ast := i.Ast.(type) {
	case *parser.Programa:
		declaracoes = ast.Declaracoes
	case *parser.Bloco:
		declaracoes = ast.Declaracoes
	default:
		return nil, i.criarErroF(TipagemErro, "Quando usar o método `Inicializa()`, a ast deve ser do tipo `Programa` ou `Bloco`")
	}

	return i.Visite(declaracoes)
}

// Visite executa de forma sequencial um array ordenado de instruções lógicas.
//
// Interrompe o processamento imediatamente se alguma exceção (erro) for levantada,
// ou se um valor de retorno for definido no interpretador (via comando 'retorne').
func (i *Interpretador) Visite(nodes []parser.BaseNode) (Objeto, error) {
	var resultado Objeto
	var err error

	for _, node := range nodes {
		resultado, err = i.visite(node)
		adicionaContextoSeNaoTiver(err, i.Contexto)

		if err != nil {
			return nil, err
		}

		// Interrompe a execução local se encontrar instrução de retorno ('retorne')
		if i.ValorRetorno != nil {
			return i.ValorRetorno, nil
		}
	}

	return resultado, nil
}

// visite é a central de desvio do Visitor Pattern.
//
// Antes de avaliar as propriedades do nó, ele atualiza as coordenadas da VM no 'Contexto'
// utilizando o mapa de posições físicas do parser. Isso garante que tracebacks tenham as coordenadas físicas reais do erro.
func (i *Interpretador) visite(astNode parser.BaseNode) (Objeto, error) {
	if i.Posicoes != nil {
		if tok, ok := i.Posicoes[astNode]; ok && tok != nil {
			i.Contexto.TokenAtual = tok
			if tok.Inicio != nil {
				i.Contexto.LinhaAtual = tok.Inicio.Linha
				i.Contexto.ColunaAtual = tok.Inicio.Coluna
			}
		}
	}

	switch node := astNode.(type) {
	case *parser.DeclVar:
		return i.visiteDeclVar(node)
	case *parser.DeclFuncao:
		return i.visiteDeclFuncao(node)
	case *parser.ChamadaFuncao:
		return i.visiteChamadaFuncao(node)
	case *parser.TextoLiteral:
		return i.visiteTextoLiteral(node)
	case *parser.InteiroLiteral:
		return i.visiteInteiroLiteral(node)
	case *parser.DecimalLiteral:
		return i.visiteDecimalLiteral(node)
	case *parser.TuplaLiteral:
		return i.visiteTuplaLiteral(node)
	case *parser.ListaLiteral:
		return i.visiteListaLiteral(node)
	case *parser.OpBinaria:
		return i.visiteOpBinaria(node)
	case *parser.OpUnaria:
		return i.visiteOpUnaria(node)
	case *parser.Identificador:
		return i.visiteIdentificador(node)
	case *parser.Reatribuicao:
		return i.visiteReatribuicao(node)
	case *parser.ExpressaoSe:
		return i.visiteExpressaoSe(node)
	case *parser.Bloco:
		return i.visiteBloco(node)
	case *parser.RetorneNode:
		return i.visiteRetorneNode(node)
	case *parser.Enquanto:
		return i.visiteEnquanto(node)
	case *parser.AcessoMembro:
		return i.visiteAcessoMembro(node)
	case *parser.BlocoPara:
		return i.visiteBlocoPara(node)
	case *parser.PareNode:
		return i.visitePareNode(node)
	case *parser.ContinueNode:
		return i.visiteContinueNode(node)
	case *parser.ImporteDe:
		return i.visiteImporteDe(node)
	case *parser.Indexacao:
		return i.visiteIndexacao(node)
	case *parser.MapaLiteral:
		return i.visiteMapaLiteral(node)
	case *parser.NovaNode:
		return i.visiteNovaNode(node)
	case *parser.AsseguraNode:
		return i.visiteAsseguraNode(node)
	case *parser.DeclClasse:
		return i.visiteDeclClasse(node)
	case *parser.OpPipe:
		return i.visiteOpPipe(node)
	case *parser.ArgumentoNomeado:
		return i.visiteArgumentoNomeado(node)
	case *parser.DeclTeste:
		return i.visiteDeclTeste(node)
	}

	return nil, nil
}

// visiteDeclVar avalia a expressão inicializadora (se fornecida) e cria o novo símbolo no escopo.
func (i *Interpretador) visiteDeclVar(node *parser.DeclVar) (Objeto, error) {
	var valor Objeto = Nulo

	if node.Inicializador != nil {
		val, err := i.visite(node.Inicializador)

		if err != nil {
			err.(*Erro).AdicionarContexto(i.Contexto)
			return nil, err
		}

		valor = val
	}

	simbolo := NewVarSimbolo(node.Nome, valor)

	if node.Constante {
		simbolo.Constante = true
	}

	if err := i.Escopo.DefinirSimbolo(simbolo); err != nil {
		return nil, err
	}

	return nil, nil
}

// visiteDeclFuncao instancia um novo objeto Funcao ligando o corpo e os escopos e o registra na tabela local.
func (i *Interpretador) visiteDeclFuncao(node *parser.DeclFuncao) (Objeto, error) {
	funcao := NewFuncao(node.Nome, node.Corpo, i.Contexto, i.Escopo)

	nomes := make([]string, len(node.Parametros))
	for idx, param := range node.Parametros {
		nomes[idx] = param.Nome
		if param.Padrao != nil {
			funcao.definirDefault(param.Nome, param.Padrao)
		}
	}
	funcao.definirArgs(nomes)

	err := i.Escopo.DefinirSimbolo(NewVarSimbolo(node.Nome, funcao))
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// visiteChamadaFuncao avalia todos os argumentos passados na chamada e invoca o chamável via ptst.Chamar.
func (i *Interpretador) visiteChamadaFuncao(node *parser.ChamadaFuncao) (Objeto, error) {
	objeto, err := i.visite(node.Identificador)

	if err != nil {
		return nil, err
	}

	var args Tupla

	for _, argnode := range node.Argumentos {
		arg, err := i.visite(argnode)

		if err != nil {
			return nil, err
		}

		args = append(args, arg)
	}

	return Chamar(objeto, args)
}

// visiteTextoLiteral converte a string física limpa de aspas em um Texto nativo.
func (i *Interpretador) visiteTextoLiteral(node *parser.TextoLiteral) (Objeto, error) {
	return NewTexto(node.Valor[1 : len(node.Valor)-1])
}

// visiteInteiroLiteral converte o lexema numérico em um Inteiro nativo da VM.
func (i *Interpretador) visiteInteiroLiteral(node *parser.InteiroLiteral) (Objeto, error) {
	return NewInteiro(node.Valor)
}

// visiteDecimalLiteral converte o lexema real em um Decimal nativo da VM.
func (i *Interpretador) visiteDecimalLiteral(node *parser.DecimalLiteral) (Objeto, error) {
	return NewDecimal(node.Valor)
}

// visiteTuplaLiteral instancia uma nova coleção de dados imutável do tipo Tupla.
func (i *Interpretador) visiteTuplaLiteral(node *parser.TuplaLiteral) (Objeto, error) {
	var tupla Tupla

	for _, elemento := range node.Elementos {
		item, err := i.visite(elemento)
		if err != nil {
			return nil, err
		}

		tupla = append(tupla, item)
	}
	return tupla, nil
}

// visiteListaLiteral instancia uma nova coleção mutável do tipo Lista.
func (i *Interpretador) visiteListaLiteral(node *parser.ListaLiteral) (Objeto, error) {
	lista := &Lista{}

	for _, elemento := range node.Elementos {
		item, err := i.visite(elemento)
		if err != nil {
			return nil, err
		}

		lista.Adiciona(item)
	}
	return lista, nil
}

// visiteOpUnaria desvia a operação unária para as rotinas matemáticas centrais (Neg, Pos, Nao).
func (i *Interpretador) visiteOpUnaria(node *parser.OpUnaria) (Objeto, error) {
	operando, err := i.visite(node.Expressao)

	if err != nil {
		return nil, err
	}

	switch node.Operador {
	case "-":
		return Neg(operando)
	case "+":
		return Pos(operando)
	case "nao":
		return Nao(operando)
	}

	return nil, NewErroF(TipagemErro, "A operação '%s' não é suportada para o tipo '%s'", node.Operador, operando.Tipo().Nome)
}

// visiteOpBinaria resolve expressões matemáticas e lógicas binárias, executando avaliações em short-circuit para 'e' / 'ou'.
func (i *Interpretador) visiteOpBinaria(node *parser.OpBinaria) (Objeto, error) {
	esquerda, err := i.visite(node.Esq)

	if err != nil {
		return nil, err
	}

	// Curto-circuito lógicos: 'ou' (OR) e 'e' (AND) evitam avaliar o operando da direita se possível
	if node.Operador == "ou" {
		if v, err := NewBooleano(esquerda); err != nil {
			return nil, err
		} else if v.(Booleano) {
			return esquerda, nil
		}

		return i.visite(node.Dir)
	}

	if node.Operador == "e" {
		if v, err := NewBooleano(esquerda); err != nil {
			return nil, err
		} else if !v.(Booleano) {
			return esquerda, nil
		}

		return i.visite(node.Dir)
	}

	direita, err := i.visite(node.Dir)

	if err != nil {
		return nil, err
	}

	switch node.Operador {
	case "+":
		return Adiciona(esquerda, direita)
	case "*":
		return Multiplica(esquerda, direita)
	case "-":
		return Subtrai(esquerda, direita)
	case "/":
		return Divide(esquerda, direita)
	case "//":
		return DivideInteiro(esquerda, direita)
	case "%":
		return Mod(esquerda, direita)
	case "<":
		return MenorQue(esquerda, direita)
	case "<=":
		return MenorOuIgual(esquerda, direita)
	case "==":
		return Igual(esquerda, direita)
	case "!=":
		return Diferente(esquerda, direita)
	case ">":
		return MaiorQue(esquerda, direita)
	case ">=":
		return MaiorOuIgual(esquerda, direita)
	case "|":
		return Ou(esquerda, direita)
	case "&":
		return E(esquerda, direita)
	case "em":
		return Em(direita, esquerda)
	case "instancia":
		// Trata o operador 'instancia de'
		inst, ok := esquerda.(*Instancia)
		if !ok {
			return Falso, nil
		}
		classeObj, ok := direita.(*ClasseObj)
		if !ok {
			return Falso, nil
		}

		curr := inst.Classe
		for curr != nil {
			if curr == classeObj {
				return Verdadeiro, nil
			}
			curr = curr.Base
		}
		return Falso, nil
	}

	return nil, NewErroF(TipagemErro, "A operação '%s' não é suportada entre os tipos '%s' e '%s'", node.Operador, esquerda.Tipo().Nome, direita.Tipo().Nome)
}

// visiteIdentificador recupera o valor de uma variável no escopo, ou no catálogo global de embutidos como fallback.
func (i *Interpretador) visiteIdentificador(node *parser.Identificador) (Objeto, error) {
	objeto, err := i.Escopo.ObterValor(node.Nome)

	if err != nil {
		if objeto, err = i.Contexto.Modulos.Embutidos.M__obtem_attributo__(node.Nome); err == nil {
			return objeto, nil
		}

		return nil, err
	}

	return objeto, nil
}

// visiteReatribuicao realiza atribuição ordinária ou acumulações aritméticas compostas (+=, -=).
func (i *Interpretador) visiteReatribuicao(node *parser.Reatribuicao) (Objeto, error) {
	var direita, esquerda, valor Objeto
	var err error

	if direita, err = i.visite(node.Expressao); err != nil {
		return nil, err
	}

	if node.Operador == "=" {
		if obj, ok := node.Objeto.(*parser.AcessoMembro); ok {
			pai, err := i.visite(obj.Dono)
			if err != nil {
				return nil, err
			}
			membro := obj.Membro.(*parser.Identificador).Nome
			err = DefineAtributo(pai, membro, direita)
			return direita, err
		}
	}

	if esquerda, err = i.visite(node.Objeto); err != nil {
		if obj, ok := node.Objeto.(*parser.Indexacao); ok && node.Operador == "=" {
			nodePai := obj.Objeto
			nodeFilho := obj.Argumento

			for {
				if a, ok := nodeFilho.(*parser.Indexacao); ok {
					nodePai = a.Objeto
					nodeFilho = a.Argumento
					continue
				}

				break
			}

			var pai, filho Objeto
			if pai, err = i.visite(nodePai); err != nil {
				return nil, err
			}

			if filho, err = i.visite(nodeFilho); err != nil {
				return nil, err
			}

			return DefineItem(pai, filho, direita)
		}

		return nil, err
	}

	valor = direita

	switch node.Operador {
	case "+=":
		valor, err = AdicionaEAtribui(esquerda, direita)
	case "*=":
		valor, err = MultiplicaEAtribui(esquerda, direita)
	case "-=":
		valor, err = SubtraiEAtribui(esquerda, direita)
	case "/=":
		valor, err = DivideEAtribui(esquerda, direita)
	case "//=":
		valor, err = DivideInteiroEAtribui(esquerda, direita)
	}

	if err != nil {
		return nil, err
	}

	switch obj := node.Objeto.(type) {
	case *parser.AcessoMembro:
		pai, err := i.visite(obj.Dono)
		if err != nil {
			return nil, err
		}
		membro := obj.Membro.(*parser.Identificador).Nome
		err = DefineAtributo(pai, membro, valor)
		return valor, err
	case *parser.Indexacao:
		if esquerda, err = i.visite(obj.Objeto); err != nil {
			return nil, err
		}
		chave, err := i.visite(obj.Argumento)
		if err != nil {
			return nil, err
		}
		return DefineItem(esquerda, chave, valor)
	case *parser.Identificador:
		return nil, i.Escopo.RedefinirValor(obj.Nome, valor)
	}

	return nil, nil
}

// visiteExpressaoSe avalia a condicional booleana e encaminha a execução para a ramificação correta.
func (i *Interpretador) visiteExpressaoSe(node *parser.ExpressaoSe) (Objeto, error) {
	condicao, err := i.visite(node.Condicao)
	if err != nil {
		return nil, err
	}

	if condicao, err = NewBooleano(condicao); err != nil {
		return nil, err
	}

	if condicao.(Booleano) {
		return i.visite(node.Corpo)
	}

	return i.visite(node.Alternativa)
}

// visiteBloco aloca um escopo léxico temporário filho, executa as declarações e o destrói ao final.
func (i *Interpretador) visiteBloco(node *parser.Bloco) (Objeto, error) {
	i.entrarNoEscopo(nil)
	defer i.sairDoEscopo()

	for _, decl := range node.Declaracoes {
		if _, err := i.visite(decl); err != nil {
			return nil, err
		}
	}
	return nil, nil
}

// visiteRetorneNode avalia a expressão e popula o registrador 'ValorRetorno' para interromper e sair de funções.
func (i *Interpretador) visiteRetorneNode(node *parser.RetorneNode) (Objeto, error) {
	valor, err := i.visite(node.Expressao)
	if err != nil {
		return nil, err
	}
	i.ValorRetorno = valor
	return valor, nil
}

// visiteEnquanto roda o laço condicional em Go. Intercepta erros estruturados 'ErroContinue' e 'ErroPare'.
func (i *Interpretador) visiteEnquanto(node *parser.Enquanto) (Objeto, error) {
	for {
		condicao, err := i.visite(node.Condicao)
		if err != nil {
			return nil, err
		}

		if condicao, err = NewBooleano(condicao); err != nil {
			return nil, err
		}

		if !condicao.(Booleano) {
			break
		}

		_, err = i.visite(node.Corpo)
		if err != nil {
			if objErr, ok := err.(*Erro); ok {
				switch objErr.Tipo() {
				case ErroContinue:
					continue
				case ErroPare:
					return nil, nil
				}
			}

			return nil, err
		}
	}

	return nil, nil
}

// visiteAcessoMembro extrai propriedades dinâmicas das tabelas de instâncias ou tipos (classes).
func (i *Interpretador) visiteAcessoMembro(node *parser.AcessoMembro) (Objeto, error) {
	dono, err := i.visite(node.Dono)
	if err != nil {
		return nil, err
	}

	membro := node.Membro.(*parser.Identificador).Nome
	return ObtemAtributoS(dono, membro)
}

// visiteBlocoPara gerencia e executa loops de repetição iterativos (para-em), tratando exceções FimIteracao graciosamente.
func (i *Interpretador) visiteBlocoPara(node *parser.BlocoPara) (Objeto, error) {
	i.Escopo.DefinirSimbolo(NewVarSimbolo(node.Identificador, Nulo))
	defer func() {
		i.Escopo.ExcluirSimbolo(node.Identificador)
	}()

	var item, iterador Objeto
	var err error

	if iterador, err = i.visite(node.Iterador); err != nil {
		return nil, err
	}

	if iterador, err = Iter(iterador); err != nil {
		return nil, err
	}

	for {
		if item, err = Proximo(iterador); err != nil {
			if objErr, ok := err.(*Erro); ok {
				if objErr.Tipo() == FimIteracao {
					return nil, nil
				}
			}

			return nil, err
		}

		i.Escopo.RedefinirValor(node.Identificador, item)

		_, err = i.visite(node.Corpo)
		if err != nil {
			if objErr, ok := err.(*Erro); ok {
				switch objErr.Tipo() {
				case ErroContinue:
					continue
				case ErroPare:
					return nil, nil
				}
			}

			return nil, err
		}
	}
}

// visitePareNode retorna um erro controlado sinalizando à VM a interrupção imediata do laço.
func (i *Interpretador) visitePareNode(_ *parser.PareNode) (Objeto, error) {
	return nil, NewErro(ErroPare, Nulo)
}

// visiteContinueNode retorna um erro controlado sinalizando o avanço de iteração de laço.
func (i *Interpretador) visiteContinueNode(_ *parser.ContinueNode) (Objeto, error) {
	return nil, NewErro(ErroContinue, Nulo)
}

// visiteImporteDe executa a importação parcial de constantes e métodos de módulos qualificados.
func (i *Interpretador) visiteImporteDe(node *parser.ImporteDe) (Objeto, error) {
	caminho, err := i.visiteTextoLiteral(node.Caminho)
	if err != nil {
		return nil, err
	}

	modulo, err := MaquinarioImporteModulo(i.Contexto, string(caminho.(Texto)), i.Escopo)
	if err != nil {
		return nil, err
	}

	for _, nome := range node.Nomes {
		obj, err := ObtemAtributoS(modulo, nome)
		if err != nil {
			return nil, err
		}

		i.Escopo.DefinirSimbolo(NewVarSimbolo(nome, obj))
	}

	return nil, nil
}

// visiteIndexacao resolve indexações de coleções (como fatiar strings ou listas).
func (i *Interpretador) visiteIndexacao(node *parser.Indexacao) (Objeto, error) {
	var obj, arg Objeto
	var err error

	if obj, err = i.visite(node.Objeto); err != nil {
		return nil, err
	}

	if arg, err = i.visite(node.Argumento); err != nil {
		return nil, err
	}

	return ObtemItem(obj, arg)
}

// visiteMapaLiteral cria um dicionário vazio e popula os seus respectivos pares de chaves e valores.
func (i *Interpretador) visiteMapaLiteral(node *parser.MapaLiteral) (Objeto, error) {
	mapa := NewMapaVazio()

	for _, entrada := range node.Entradas {
		var chave Objeto
		var err error

		if id, ok := entrada.Chave.(*parser.Identificador); ok {
			chave, err = NewTexto(id.Nome)
			if err != nil {
				return nil, err
			}
		} else {
			chave, err = i.visite(entrada.Chave)
			if err != nil {
				return nil, err
			}
		}

		valor, err := i.visite(entrada.Valor)
		if err != nil {
			return nil, err
		}

		if _, err := mapa.M__define_item__(chave, valor); err != nil {
			return nil, err
		}
	}

	return mapa, nil
}

// visiteNovaNode executa a instanciação de classes do usuário (ClasseObj) ou primitivos nativos da VM.
func (i *Interpretador) visiteNovaNode(node *parser.NovaNode) (Objeto, error) {
	chamada, ok := node.Objeto.(*parser.ChamadaFuncao)
	if !ok {
		return nil, NewErroF(SintaxeErro, "era esperada uma sintaxe similar a chamada de função após o token 'nova'")
	}

	var obj Objeto
	var err error

	if obj, err = i.visite(chamada.Identificador); err != nil {
		return nil, err
	}

	var elementos []parser.BaseNode = chamada.Argumentos
	var argsTupla Tupla
	for _, elem := range elementos {
		val, err := i.visite(elem)
		if err != nil {
			return nil, err
		}
		argsTupla = append(argsTupla, val)
	}

	if classe, ok := obj.(*ClasseObj); ok {
		return classe.M__nova_instancia__(nil, argsTupla)
	}

	return NovaInstancia(obj, argsTupla)
}

// visiteDeclClasse registra uma nova definição de classe estruturada e seus métodos no escopo local.
func (i *Interpretador) visiteDeclClasse(node *parser.DeclClasse) (Objeto, error) {
	classe := &ClasseObj{
		Nome:    node.Nome,
		Metodos: make(map[string]*Funcao),
	}

	if node.Heranca != "" {
		baseVal, err := i.Escopo.ObterValor(node.Heranca)
		if err != nil {
			return nil, err
		}
		baseClasse, ok := baseVal.(*ClasseObj)
		if !ok {
			return nil, NewErroF(TipagemErro, "a classe base '%s' deve ser uma classe", node.Heranca)
		}
		classe.Base = baseClasse
	}

	for _, metodoNode := range node.Metodos {
		funcao := NewFuncao(metodoNode.Nome, metodoNode.Corpo, i.Contexto, i.Escopo)
		funcao.Estatico = metodoNode.Estatico
		nomes := make([]string, len(metodoNode.Parametros))
		for idx, param := range metodoNode.Parametros {
			nomes[idx] = param.Nome
			if param.Padrao != nil {
				funcao.definirDefault(param.Nome, param.Padrao)
			}
		}
		funcao.definirArgs(nomes)
		classe.Metodos[metodoNode.Nome] = funcao
	}

	err := i.Escopo.DefinirSimbolo(NewVarSimbolo(node.Nome, classe))
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// visiteAsseguraNode realiza a validação de asserções lógica e emite ErroDeAsseguracao caso resulte em Falso.
func (i *Interpretador) visiteAsseguraNode(node *parser.AsseguraNode) (Objeto, error) {
	var condicao, mensagem Objeto = Verdadeiro, Texto("")
	var err error

	if condicao, err = i.visite(node.Condicao); err != nil {
		return nil, err
	}

	if node.Mensagem != nil {
		if mensagem, err = i.visite(node.Mensagem); err != nil {
			return nil, err
		}
	}

	boolCond, err := NewBooleano(condicao)
	if err != nil {
		return nil, err
	}

	if boolCond == Falso {
		return nil, NewErro(ErroDeAsseguracao, mensagem)
	}

	return nil, nil
}

// visiteOpPipe executa o encadeamento funcional do operador pipe (|>).
// Garante avaliar e injetar o operando da esquerda prioritariamente para prevenir múltiplos efeitos colaterais.
func (i *Interpretador) visiteOpPipe(node *parser.OpPipe) (Objeto, error) {
	esquerda, err := i.visite(node.Esq)
	if err != nil {
		return nil, err
	}

	switch dir := node.Dir.(type) {
	case *parser.ChamadaFuncao:
		objeto, err := i.visite(dir.Identificador)
		if err != nil {
			return nil, err
		}

		args := make(Tupla, 0, len(dir.Argumentos)+1)
		args = append(args, esquerda)

		for _, argnode := range dir.Argumentos {
			arg, err := i.visite(argnode)
			if err != nil {
				return nil, err
			}
			args = append(args, arg)
		}

		return Chamar(objeto, args)

	default:
		funcao, err := i.visite(node.Dir)
		if err != nil {
			return nil, err
		}

		return Chamar(funcao, Tupla{esquerda})
	}
}

func (i *Interpretador) visiteArgumentoNomeado(node *parser.ArgumentoNomeado) (Objeto, error) {
	val, err := i.visite(node.Valor)
	if err != nil {
		return nil, err
	}
	// Retornamos um par estruturado temporário (ou um tipo interno) que o Chamar / Funcao.M__chame__
	// possa detectar. Podemos simplesmente retornar uma struct ou usar um Mapa com chave especial,
	// ou criar um tipo runtime específico.
	// Vamos criar um tipo runtime interno para representar um argumento nomeado avaliado
	return &ArgumentoNomeadoObj{Nome: node.Nome, Valor: val}, nil
}

// visiteDeclTeste executa o bloco de teste. Se falhar, captura e lança erro amigável.
func (i *Interpretador) visiteDeclTeste(node *parser.DeclTeste) (Objeto, error) {
	// Se estivermos em modo de execução normal e não de teste, podemos apenas ignorar
	// ou podemos executar. Para compatibilidade e facilidade, vamos executar o teste.
	// Se ocorrer um erro de asserção, propagamos.
	_, err := i.visite(node.Corpo)
	if err != nil {
		if objErr, ok := err.(*Erro); ok && objErr.Tipo() == ErroDeAsseguracao {
			// Adiciona o nome do teste ao erro para identificação
			return nil, NewErroF(ErroDeAsseguracao, "Teste '%s' falhou: %v", node.Nome, objErr.Mensagem)
		}
		return nil, err
	}
	return nil, nil
}

// criarErroF aloca e formata erros associando automaticamente os metadados do contexto da VM.
func (i *Interpretador) criarErroF(tipo *Tipo, format string, args ...any) error {
	erro := NewErroF(tipo, format, args...)
	erro.Contexto = i.Contexto
	return erro
}
