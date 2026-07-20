package hrp

import (
	"fmt"
	"os"
	"strings"

	"github.com/mat-dgruber/Harpia/lexer"
)

// Erro representa uma exceção estruturada em tempo de execução ou de análise no Harpia.
//
// Esta estrutura atua em conformidade com a interface 'error' nativa do Go, porém é enriquecida
// com metadados geográficos detalhados (arquivo, linha, coluna, token físico causador) para formatar
// e emitir relatórios de tracebacks visuais amigáveis ao usuário inteiramente em português.
type Erro struct {
	Base     *Tipo        // A classe específica de erro do Harpia (ex: NomeErro, SintaxeErro).
	Contexto *Contexto    // Ponteiro de referência ao supervisor global da VM.
	CRef     int          // Atributo reservado para futuras implementações de referências internas.
	Mensagem Objeto       // Texto descritivo ou representação do erro.
	Linha    int          // Número da linha física (base 0, -1 indica posição indefinida).
	Coluna   int          // Número da coluna física (base 1).
	Token    *lexer.Token // O token do lexer correspondente ao local do erro.
	Arquivo  string       // O arquivo físico ou virtualizado no qual a exceção ocorreu (ex: "<playground>").
	Codigo   string       // O código fonte completo do arquivo para renderização do trecho culpado.
	Sugestao string       // Dica explicativa ou sugestão contextual para ajudar o usuário a corrigir o erro.
}

// BaseErro é a classe primordial de onde herdam todas as outras exceções da linguagem.
var BaseErro = TipoObjeto.NewTipo(
	"BaseErro",
	"A classe de erro base para todas as outras.",
)

func init() {
	// Injeta a lógica de instanciação de construtor em classes herdadas de erros.
	BaseErro.Mapa["__nova_instancia__"] = NewMetodoOuPanic(
		"__nova_instancia__",
		func(inst Objeto, args Tupla) (Objeto, error) {
			tipo := inst.(*Tipo)
			message, ok := args[0].(Texto)
			if !ok {
				return nil, NewErroF(TipagemErro, "O primeiro argumento de '%s' deve ser do tipo '%s', e não '%s'", tipo.Nome, TipoTexto.Nome, args[0].Tipo().Nome)
			}

			return NewErro(tipo, message), nil
		},
		"",
	)

	NaoImplementado = NewErro(NaoImplementadoErro, nil)
}

// Relação de todas as classes de exceções disponíveis nativamente na VM.
var (
	// TipoErro é a classe base comum para todas as exceções que não sinalizam interrupções de fluxo normais.
	TipoErro = BaseErro.NewTipo("Erro", "Base comum para todos os erros que não são de saída.")

	// SintaxeErro é lançado quando o código fonte viola as regras gramaticais da linguagem.
	SintaxeErro = TipoErro.NewTipo("SintaxeErro", "Sintaxe Invalida.")

	// ReatribuicaoErro indica tentativas ilegais de redefinir ou declarar constantes preexistentes.
	ReatribuicaoErro = TipoErro.NewTipo("ReatribuicaoErro", "Proibido redeclarar.")

	// AtributoErro ocorre quando se tenta acessar uma propriedade ou método não registrado na classe ou instância.
	AtributoErro = TipoErro.NewTipo("AtributoErro", "Atributo não encontrado.")

	// TipagemErro ocorre quando operandos ou parâmetros possuem tipos incompatíveis com a operação.
	TipagemErro = TipoErro.NewTipo("TipagemErro", "Tipo de argumento inapropriado.")

	// NomeErro é lançado quando um identificador ou variável é referenciado sem estar previamente definido no escopo.
	NomeErro = TipoErro.NewTipo("NomeErro", "Erro de nome que não pode ser achado.")

	// ImportacaoErro indica falhas ao carregar módulos físicos ou símbolos ausentes nestes módulos.
	ImportacaoErro = TipoErro.NewTipo("ImportacaoErro", "Não é possível encontrar o módulo ou símbolo nele")

	// ValorErro sinaliza que o argumento possui tipo correto, mas valor semântico inadequado.
	ValorErro = TipoErro.NewTipo("ValorErro", "O valor é inapropriádo ou sua ocorrencia não existe")

	// ErroDeLimite indica valores numéricos fora dos limites ou tamanhos aceitos pela VM.
	ErroDeLimite = ValorErro.NewTipo("ErroDeLimite", "O valor está fora do intervalo permitido.")

	// IndiceErro ocorre quando se tenta indexar sequências (listas, tuplas, strings) fora dos seus limites de tamanho.
	IndiceErro = TipoErro.NewTipo("IndiceErro", "O indice está fora do range aceito")

	// RuntimeErro é a exceção genérica para falhas inesperadas no ambiente de execução.
	RuntimeErro = TipoErro.NewTipo("RuntimeErro", "Erro no ambiente de execução")

	// FimIteracao é o sinalizador estruturado de parada que interrompe loops iterativos 'para-em'.
	FimIteracao = TipoErro.NewTipo("FimIteracao", "Sinaliza o fim da iteração quando `objeto.__proximo__() não retorna mais nada")

	// ErroDeAsseguracao é disparado quando uma asserção lógica da diretiva 'assegura' resulta em Falso.
	ErroDeAsseguracao = TipoErro.NewTipo("ErroDeAsseguracao", "Erro lançado em um `assegura obj`")

	// NaoImplementadoErro indica que um método abstrato de interface ou função da biblioteca padrão está ausente.
	NaoImplementadoErro = TipoErro.NewTipo("NaoImplementadoErro", "O método ou função não foi implementada/o ainda")

	// DivisaoPorZeroErro impede cálculos aritméticos de divisão real, inteira ou resto por zero.
	DivisaoPorZeroErro = TipoErro.NewTipo("DivisaoPorZeroErro", "A divisão de algum número por zero não é possível")

	// ConsultaErro é a classe base comum para falhas de buscas indexadas por chave ou índice.
	ConsultaErro = TipoErro.NewTipo("ConsultaErro", "Classe base para erros que envolem chave ou indice em elementos")

	// ChaveErro ocorre ao tentar acessar dicionários (mapas) utilizando chaves inexistentes.
	ChaveErro = ConsultaErro.NewTipo("ChaveErro", "Lançado quando a chave de um mapa não existe ou é inválida")

	// ErroDeSistema indica falhas associadas a chamadas e comandos do sistema operacional (ex: falhas de E/S).
	ErroDeSistema = TipoErro.NewTipo("ErroDeSistema", "Erro relacionado a operações do sistema operacional.")

	// ArquivoNaoEncontradoErro indica falhas ao tentar abrir ou ler caminhos inexistentes no disco rígido.
	ArquivoNaoEncontradoErro = ErroDeSistema.NewTipo("ArquivoNaoEncontradoErro", "O arquivo não pôde ser encontrado")

	// ErroDePilha ocorre quando o limite máximo de recursão da VM é atingido.
	ErroDePilha = TipoErro.NewTipo("ErroDePilha", "Limite máximo de recursão excedido")

	// ErroContinue é uma sinalização interna de controle para simular o avanço do comando 'continue' em loops.
	ErroContinue = TipoErro.NewTipo("ErroContinue", "Erro utilizado para representar a instrução 'continue' em loops")

	// ErroPare é uma sinalização interna de controle para simular a interrupção do comando 'pare' em loops.
	ErroPare = TipoErro.NewTipo("ErroPare", "Erro utilizado para representar a instrução 'pare' em loops")

	// NaoImplementado é a instância estática predefinida para exceções de métodos inacabados.
	NaoImplementado Objeto
)

// NewErro é o construtor básico para instanciar novas structs Erro.
func NewErro(tipo *Tipo, mensagem Objeto) *Erro {
	return &Erro{Base: tipo, Mensagem: mensagem, Linha: -1}
}

// NewErroF cria instâncias de Erro formatando as strings de mensagem usando printf.
func NewErroF(tipo *Tipo, format string, p ...any) *Erro {
	return &Erro{Base: tipo, Mensagem: Texto(fmt.Sprintf(format, p...)), Linha: -1}
}

// Tipo retorna a representação de classe (Tipo de Erro) da struct.
func (e *Erro) Tipo() *Tipo {
	return e.Base
}

// AdicionarContexto vincula o supervisor de estado da VM à exceção para alimentar a resolução de tracebacks.
//
// Também propaga as coordenadas geográficas (Linha/Coluna/Token) do contexto
// para o Erro, caso ainda não tenham sido definidas — permitindo que scripts
// inspecionem `erro.linha`, `erro.coluna`, `erro.arquivo` durante o capture.
func (e *Erro) AdicionarContexto(contexto *Contexto) {
	if contexto == nil {
		return
	}

	if e.Contexto == nil {
		e.Contexto = contexto
	}

	if e.Linha == -1 {
		e.Linha = contexto.LinhaAtual
		e.Coluna = contexto.ColunaAtual
		e.Token = contexto.TokenAtual
		e.Arquivo = contexto.ArquivoAtual
		e.Codigo = contexto.CodigoAtual
	}
}

func (e *Erro) M__obtem_attributo__(nome string) (Objeto, error) {
	switch nome {
	case "mensagem":
		if e.Mensagem != nil {
			return e.Mensagem, nil
		}
		return Nulo, nil
	case "linha":
		return NewInteiro(e.Linha)
	case "coluna":
		return NewInteiro(e.Coluna)
	case "arquivo":
		return NewTexto(e.Arquivo)
	}

	return nil, NewErroF(AtributoErro, "O atributo '%s' não existe no tipo '%s'", nome, e.Tipo().Nome)
}

var _ I__obtem_attributo__ = (*Erro)(nil)

// ObterCodigoErro mapeia cada classe de exceção do Harpia a um código estruturado de erro normatizado.
// Isso facilita de forma extraordinária a indexação, criação de fóruns e documentação de suporte a bugs.
func ObterCodigoErro(tipo *Tipo) string {
	switch tipo {
	case SintaxeErro:
		return "PSC-0001"
	case ReatribuicaoErro:
		return "PSC-0002"
	case AtributoErro:
		return "PSC-0003"
	case TipagemErro:
		return "PSC-0004"
	case NomeErro:
		return "PSC-0005"
	case ImportacaoErro:
		return "PSC-0006"
	case ValorErro:
		return "PSC-0007"
	case ErroDeLimite:
		return "PSC-0008"
	case IndiceErro:
		return "PSC-0009"
	case RuntimeErro:
		return "PSC-0010"
	case ErroDeAsseguracao:
		return "PSC-0011"
	case DivisaoPorZeroErro:
		return "PSC-0012"
	case ErroDeSistema:
		return "PSC-0013"
	case ArquivoNaoEncontradoErro:
		return "PSC-0014"
	case ErroDePilha:
		return "PSC-0015"
	default:
		return "PSC-0000"
	}
}

// ObterSugestaoErro fornece uma camada de inteligência corretiva fantástica.
// Analisa heuristicamente os lexemas textuais da exceção e retorna dicas contextuais amigáveis.
func ObterSugestaoErro(tipo *Tipo, mensagem string) string {
	msgLower := strings.ToLower(mensagem)
	if tipo == NomeErro && (strings.Contains(msgLower, "imrpimir") || strings.Contains(msgLower, "imprimi")) {
		return "Você quis dizer 'imprimir'?"
	}
	if tipo == SintaxeErro && strings.Contains(msgLower, "retornar") {
		return "Em Harpia, use a palavra-chave 'retorne' para retornar valores."
	}
	if tipo == DivisaoPorZeroErro {
		return "Não é possível dividir um número por zero."
	}
	return ""
}

// Error satisfaz a interface standard error de Go, executando o ciclo completo de renderização visual
// do traceback.
//
// Recursos do Renderizador:
//   - Identifica se o console hospedeiro suporta cores (via desativação reativa NO_COLOR). Se sim, estiliza
//     com cores de alto contraste (Vermelho para erros, Ciano para ponteiros e guias, Verde para sugestões).
//   - Imprime a caixa identificadora contendo o código normatizado indexado (ex: erro[PSC-0005]).
//   - Localiza e imprime as coordenadas espaciais geográficas exatas em formato unificado (arquivo:linha:coluna).
//   - Lê de forma resiliente a linha culpada de código no disco e a exibe no console.
//   - Desenha setas pontuais (┌──>) e sublinha na largura física exata (^^^^) o token do lexer causador do erro,
//     anexando a dica/sugestão de correção alinhada à direita ao final, criando uma das melhores e mais belas saídas de erro possíveis!
func (e *Erro) Error() string {
	codigoErro := ObterCodigoErro(e.Base)
	mensagem := ""
	if e.Mensagem != nil {
		if txt, ok := e.Mensagem.(Texto); ok {
			mensagem = string(txt)
		} else {
			mensagem = fmt.Sprintf("%v", e.Mensagem)
		}
	}

	arquivoStr := e.Arquivo
	if arquivoStr == "" {
		arquivoStr = "<desconhecido>"
	}

	codigoStr := e.Codigo
	if codigoStr == "" && e.Arquivo != "" && e.Arquivo != "<string>" && e.Arquivo != "<playground>" {
		if bytes, err := os.ReadFile(e.Arquivo); err == nil {
			codigoStr = string(bytes)
		}
	}

	sugestao := e.Sugestao
	if sugestao == "" {
		sugestao = ObterSugestaoErro(e.Base, mensagem)
	}

	usarCores := os.Getenv("NO_COLOR") == ""
	corErro := ""
	corCodigo := ""
	corPonteiro := ""
	corDica := ""
	corReset := ""

	if usarCores {
		corErro = "\x1b[1;31m"     // Vermelho Negrito
		corCodigo = "\x1b[1;37m"   // Branco Negrito
		corPonteiro = "\x1b[1;36m" // Ciano Negrito
		corDica = "\x1b[1;32m"     // Verde Negrito
		corReset = "\x1b[0m"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("\n%serro[%s]%s: %s%s — %s%s\n", corErro, codigoErro, corReset, corCodigo, e.Base.Nome, mensagem, corReset))

	charSeta := "┌──>"
	charBarra := "│"
	charIgual := "＝"

	if e.Linha >= 0 { // Linha >= 0 significa posição física rastreada com sucesso
		linhaExibicao := e.Linha + 1
		sb.WriteString(fmt.Sprintf("  %s%s%s %s:%d:%d\n", corPonteiro, charSeta, corReset, arquivoStr, linhaExibicao, e.Coluna))
		sb.WriteString(fmt.Sprintf("  %s%s%s\n", corPonteiro, charBarra, corReset))

		if codigoStr != "" {
			linhas := strings.Split(codigoStr, "\n")
			idxLinha := e.Linha
			if idxLinha >= 0 && idxLinha < len(linhas) {
				linhaTexto := linhas[idxLinha]
				sb.WriteString(fmt.Sprintf("%s%d %s%s %s\n", corPonteiro, linhaExibicao, charBarra, corReset, linhaTexto))

				sb.WriteString(fmt.Sprintf("  %s%s%s ", corPonteiro, charBarra, corReset))
				coluna := e.Coluna
				if coluna > 0 {
					for j := 0; j < len(linhaTexto) && j < coluna-1; j++ {
						if linhaTexto[j] == '\t' {
							sb.WriteRune('\t')
						} else {
							sb.WriteRune(' ')
						}
					}
				}

				tamanhoSublinhado := 1
				if e.Token != nil && len(e.Token.Valor) > 0 {
					tamanhoSublinhado = len(e.Token.Valor)
				}
				sb.WriteString(fmt.Sprintf("%s%s%s", corErro, strings.Repeat("^", tamanhoSublinhado), corReset))
				sb.WriteString(" ")
				sb.WriteString(fmt.Sprintf("%s%s%s\n", corErro, mensagem, corReset))
			}
		}
		sb.WriteString(fmt.Sprintf("  %s%s%s\n", corPonteiro, charBarra, corReset))
	} else {
		sb.WriteString(fmt.Sprintf("  %s%s%s %s\n", corPonteiro, charSeta, corReset, arquivoStr))
		sb.WriteString(fmt.Sprintf("  %s%s%s %s%s%s\n", corPonteiro, charBarra, corReset, corErro, mensagem, corReset))
	}

	if sugestao != "" {
		sb.WriteString(fmt.Sprintf("  %s%s%s %ssugestão: %s%s\n", corPonteiro, charIgual, corReset, corDica, sugestao, corReset))
	}

	return sb.String()
}
