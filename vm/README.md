# Pacote `vm` (Máquina Virtual & Compilador de Bytecode do Harpia)

O pacote `vm` é o mecanismo de execução nativo de alta performance do interpretador **Harpia**. Ele encapsula tanto o **Compilador de Bytecode** (que traduz a Árvore de Sintaxe Abstrata - AST da linguagem em instruções planas de 1 byte) quanto a **Máquina Virtual de Pilha** (que executa esse bytecode de forma extremamente otimizada usando despacho rosqueado e cache polimórfico).

---

## 📖 Índice

1. [Visão Geral](#-visão-geral)
2. [Dependências e Arquitetura](#-dependências-e-arquitetura)
3. [Recursos de Performance e Otimizações](#-recursos-de-performance-e-otimizações)
4. [Especificação do Conjunto de Instruções (Opcodes)](#-especificação-do-conjunto-de-instruções-opcodes)
5. [Exemplo Completo de Uso Prático](#-exemplo-completo-de-uso-prático)

---

## 🎯 Visão Geral

Para executar programas Harpia no backend sem o overhead da interpretação direta da AST, o pacote fornece uma esteira de execução de bytecode:

1. **Compilador (`compilador.go`)**: Realiza uma tradução recursiva de passagem única (single-pass) dos nós de declarações, operações e fluxo de controle em opcodes compactos (1 byte) e gerencia o pool global de constantes do programa.
2. **Máquina Virtual (`vm.go`)**: Um motor de execução de bytecode baseado em pilha de operandos que manipula objetos canônicos do Harpia (`hrp.Objeto`) sob uma estratégia rígida de contagem de referências.

---

## 🏢 Dependências e Arquitetura

O pacote foi desenhado de forma minimalista e autônoma, dependendo apenas do ecossistema do próprio compilador e da biblioteca padrão de Go:

- **Dependências Internas**:
  - `github.com/mat-dgruber/Harpia/hrp`: Modelo de dados e sistema de tipos unificado do Harpia (Inteiro, Decimal, Texto, Lista, Mapa, etc.).
  - `github.com/mat-dgruber/Harpia/parser`: Fornece as estruturas de nós da AST para o compilador.
  - `github.com/mat-dgruber/Harpia/compartilhado`: Utilitários compartilhados de conversão segura de tipos de dados.
- **Bibliotecas Go Estritamente Utilizadas**:
  - `encoding/binary`: Codificação e decodificação rápida Big-Endian para operandos numéricos multianinhados.
  - `sync`: Pooling de pilha de operandos por meio de `sync.Pool`.
  - `time`: Medições de carimbos temporais de alta precisão do Profiler integrado.

---

## ⚡ Recursos de Performance e Otimizações

A VM implementa vários conceitos modernos de engenharia de compiladores e máquinas virtuais de nível industrial para otimizar o throughput e a latência de execução:

### 1. JIT de Traço / Despacho Rosqueado (Threaded Code JIT - Fase F)
Em vez de utilizar um tradicional loop centralizado contendo um `switch-case` (que causa frequentes falhas de predição de desvios - *branch misprediction* na CPU física), a VM do Harpia compila o bytecode plano em tempo de carregamento em uma fatia de ponteiros de funções (`[]InstrucaoThreaded`). Cada callback executa sua operação e retorna o próximo desvio diretamente, reduzindo as bolhas no pipeline do processador.

### 2. Monomorphic Inline Cache (MIC) para Variáveis
A resolução dinâmica de escopos e tabelas de símbolos (escopos léxicos aninhados) é uma das operações mais caras em linguagens dinâmicas. O opcode `OP_CARREGAR_VAR` implementa um mecanismo de cache em linha monomórfico: se o escopo atual for idêntico ao escopo do último acesso, a variável é retornada diretamente da memória cache rápida, pulando a varredura linear recursiva na árvore de escopos.

### 3. Fusão de Opcodes e Super-Instruções (Fase D)
O otimizador e compilador detecta sequências frequentes de instruções de retorno e as colapsa em uma única instrução composta (Super-Instruções). Por exemplo, a sequência que carrega uma variável ou literal e imediatamente executa o retorno é compilada diretamente nos opcodes dedicados `OP_RETORNE_VAR` e `OP_RETORNE_CONST`, reduzindo pela metade as chamadas e incrementos de ponteiro de instrução.

### 4. Reutilização de Pilha via `sync.Pool`
Para anular alocações repetidas de coleções dinâmicas de fatias no heap do Go (`make([]hrp.Objeto)`) em funções recursivas de alta concorrência, a VM utiliza um pool global sincronizado (`poolPilha`) que retém e reinicializa as pilhas locais dos frames, reduzindo drasticamente a carga sobre o garbage collector do runtime de Go.

---

## ⚙️ Especificação do Conjunto de Instruções (Opcodes)

| Instrução (Opcode) | Valor Hex | Operandos | Descrição da Operação Semântica |
|:---|:---|:---|:---|
| `OP_PUSH_CONST` | `0x01` | `[1 byte index]` | Empilha uma constante do pool local de constantes na pilha de operandos. |
| `OP_POP` | `0x02` | Nenhum | Desempilha o elemento situado no topo da pilha. |
| `OP_DUP` | `0x03` | Nenhum | Duplica o elemento situado no topo da pilha de operandos. |
| `OP_ADD` | `0x04` | Nenhum | Desempilha `a` e `b`, executa `a + b` e empilha o resultado polimórfico. |
| `OP_SUB` | `0x05` | Nenhum | Desempilha `a` e `b`, executa `a - b` e empilha o resultado polimórfico. |
| `OP_MUL` | `0x06` | Nenhum | Desempilha `a` e `b`, executa `a * b` e empilha o resultado polimórfico. |
| `OP_DIV` | `0x07` | Nenhum | Desempilha `a` e `b`, executa `a / b` e empilha o resultado polimórfico. |
| `OP_DIV_INT` | `0x08` | Nenhum | Divisão de inteiros desempilhando dois números operandos (`a // b`). |
| `OP_MOD` | `0x09` | Nenhum | Resto da divisão matemática de operandos desempilhados (`a % b`). |
| `OP_EQ` | `0x0A` | Nenhum | Avalia igualdade polimórfica (`a == b`) e empilha `Booleano`. |
| `OP_NEQ` | `0x0B` | Nenhum | Avalia desigualdade polimórfica (`a != b`) e empilha `Booleano`. |
| `OP_LT` | `0x0C` | Nenhum | Avalia se `a < b` e empilha `Booleano`. |
| `OP_LTE` | `0x0D` | Nenhum | Avalia se `a <= b` e empilha `Booleano`. |
| `OP_GT` | `0x0E` | Nenhum | Avalia se `a > b` e empilha `Booleano`. |
| `OP_GTE` | `0x0F` | Nenhum | Avalia se `a >= b` e empilha `Booleano`. |
| `OP_JMP` | `0x10` | `[2 bytes offset]` | Realiza um desvio (salto) incondicional para o endereço absoluto especificado. |
| `OP_JMP_FALSO` | `0x11` | `[2 bytes offset]` | Realiza um desvio (salto) se o topo da pilha de operandos for `Falso` ou `Nulo`. |
| `OP_CARREGAR_VAR` | `0x12` | `[1 byte index]` | Carrega na pilha o valor associado à variável identificada pelo nome no índice. |
| `OP_ARMAZENAR_VAR` | `0x13` | `[1 byte index]` | Vincula ou reatribui o valor do topo da pilha à variável identificada pelo nome. |
| `OP_CHAMAR` | `0x14` | `[1 byte arity]` | Invoca um objeto chamável com N argumentos desempilhados. |
| `OP_RETORNE` | `0x15` | Nenhum | Sái do frame de execução retornando o valor posicionado no topo da pilha. |
| `OP_AWAIT` | `0x16` | Nenhum | Suspende e aguarda de forma não-bloqueante a resolução de uma Promessa/Canal. |
| `OP_CRIAR_FUNCAO` | `0x17` | `[1 byte index]` | Instancia um objeto de função executável em tempo de execução no escopo. |
| `OP_RETORNE_CONST`| `0x18` | `[1 byte index]` | Carrega constante do pool e encerra o frame retornando-a de forma atômica. |
| `OP_RETORNE_VAR`  | `0x19` | `[1 byte index]` | Carrega variável local e encerra o frame retornando-a de forma atômica. |

---

## 🛠️ Exemplo Completo de Uso Prático

O exemplo a seguir demonstra o pipeline completo de inicialização, compilação, estruturação de frame de execução e despacho de bytecode na máquina virtual de alta performance do Harpia utilizando a linguagem Go:

```go
package main

import (
	"fmt"
	"log"

	"github.com/mat-dgruber/Harpia/hrp"
	"github.com/mat-dgruber/Harpia/parser"
	"github.com/mat-dgruber/Harpia/vm"
)

func main() {
	// 1. Código fonte em Harpia (Sintaxe em Português)
	codigoFonte := `
	var a = 10
	var b = 20
	retorne a * b + 5
	`

	// 2. Parseamento do código fonte para Árvore de Sintaxe Abstrata (AST)
	ast, err := parser.NewParserFromString(codigoFonte, "exemplo.hrp").Parse()
	if err != nil {
		log.Fatalf("Erro de parser: %v", err)
	}

	// 3. Inicialização do compilador e tradução para Bytecode
	compilador := vm.NewCompilador()
	programa, err := compilador.Compilar(ast)
	if err != nil {
		log.Fatalf("Erro de compilação: %v", err)
	}

	// 4. Criação do contexto global e escopo do interpretador Harpia
	contexto := hrp.NewContexto(hrp.OpcsContexto{})
	escopoRaiz := hrp.NewEscopo(nil)

	// 5. Inicialização da VM e do Frame inicial de ativação da pilha
	maquinaVirtual := vm.NewVM(contexto)
	frameInicial := vm.NewFrame(programa.Bytecode, programa.Constantes, escopoRaiz, nil)

	// 6. Execução veloz do bytecode compilado
	resultado, err := maquinaVirtual.Executar(frameInicial)
	if err != nil {
		log.Fatalf("Erro em runtime na VM: %v", err)
	}

	// 7. Impressão do objeto de resultado obtido
	fmt.Printf("✓ Resultado retornado pela VM: %v (Tipo: %T)\n", resultado, resultado)
}
```
