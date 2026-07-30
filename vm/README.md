# Pacote `vm` (Máquina Virtual & Compilador de Bytecode do Harpia)

O pacote `vm` é o motor de execução nativo único de alta performance do **Harpia**. Ele encapsula o **Compilador de Bytecode** (que traduz a Árvore de Sintaxe Abstrata — AST para bytecode de 16 bits com resolução estática de escopo) e a **Máquina Virtual de Pilha** (com loop `switch` zero-alloc e acesso $O(1)$ a variáveis locais em slots numéricos `uint16`).

---

## 📖 Índice

1. [Visão Geral](#-visão-geral)
2. [Dependências e Arquitetura](#-dependências-e-arquitetura)
3. [Recursos de Performance e Otimizações](#-recursos-de-performance-e-otimizações)
4. [Especificação do Conjunto de Instruções (Opcodes de 16 Bits)](#-especificação-do-conjunto-de-instruções-opcodes-de-16-bits)
5. [Exemplo Completo de Uso Prático](#-exemplo-completo-de-uso-prático)

---

## 🎯 Visão Geral

Para executar programas Harpia com máxima performance sem o overhead de caminhamento em nós da AST, o pacote fornece uma esteira de compilação e execução unificada:

1. **Compilador (`compilador.go`)**: Realiza tradução single-pass dos nós de declarações, expressões e controle de fluxo para opcodes com operandos de 16 bits (`uint16` BigEndian). Resolve estaticamente variáveis locais para slots numéricos `uint16` em tempo de compilação via `EscopoCompilador`.
2. **Máquina Virtual (`vm.go`)**: Motor de execução com loop `switch` zero-alloc, acesso $O(1)$ a slots de variáveis locais em `frame.Locais[]`, pooling de pilhas via `sync.Pool`, e manipulação com contagem de referências de `hrp.Objeto`. Integra-se de forma transparente ao `hrp.Contexto` por meio do registrador `hrp.DefinirExecutadorVM`.

---

## 🏢 Dependências e Arquitetura

- **Dependências Internas**:
  - `github.com/mat-dgruber/Harpia/hrp`: Tipos de dados, objetos (`hrp.Objeto`), escopos e contexto global de execução.
  - `github.com/mat-dgruber/Harpia/parser`: Estruturas de nós AST consumidas pelo compilador.
  - `github.com/mat-dgruber/Harpia/compartilhado`: Utilitários de conversão segura.
- **Bibliotecas Go**:
  - `encoding/binary`: Decodificação BigEndian veloz para operandos de 16 bits (`uint16`).
  - `sync`: Reutilização de pilhas via `sync.Pool`.
  - `time`: Medições do Profiler integrado da VM.

---

## ⚡ Recursos de Performance e Otimizações

### 1. Loop `switch` Zero-Alloc Plano
O motor de despacho em `vm.go` utiliza um loop `switch` altamente otimizado por compiladores Go modernos, eliminando alocações no heap durante o ciclo de instrução e maximizando a localidade de cache da CPU.

### 2. Slots Locais $O(1)$ com Indexação `uint16`
Variáveis locais são resolvidas em tempo de compilação pelo `EscopoCompilador` e mapeadas para slots inteiros `uint16`. O acesso em runtime é $O(1)$ via indexação direta em `frame.Locais[slot]`, dispensando pesquisas de string em tabelas hash.

### 3. Internamento $O(1)$ de Constantes (`constMap`)
O compilador emprega uma tabela hash interna (`constMap map[string]uint16`) para dedup de constantes numéricas, textuais e booleanas em tempo de compilação, reduzindo o tempo de compilação de $O(N)$ linear para $O(1)$ constante em scripts com múltiplos literais.

### 4. Laços `para-em` (`BlocoPara`) e Controle de Fluxo (`pare` / `continue`)
Suporte nativo à compilação de laços de repetição sobre iteráveis (`para item em colecao { ... }`), com empilhamento de frames de laço (`LoopStack`) para desvios incondicionais e remendo automático de endereços em instruções `pare` (`PareNode`) e `continue` (`ContinueNode`).

### 5. Reatribuição Composta e Escritas em Coleções / Membros
Compilação direta de operadores compostos (`+=`, `-=`, `*=`, `/=`, `//=`, `%=`) em variáveis locais/globais e suporte completo a escritas diretas em membros de objetos (`obj.prop = val`) e elementos indexados (`lista[idx] = val`).

### 6. Suporte Expandido a Operandos 16-Bit BigEndian
Pools de constantes e endereçamento de pulos suportam até 65.535 entradas (`uint16`), permitindo compilar módulos e programas extensos sem estouro de limites de operandos.

### 7. Gestão de Memória via `sync.Pool`
Pilhas de operandos dos frames são gerenciadas via `sync.Pool` (`poolPilha`), reaproveitando slices alocados para reduzir drasticamente a pressão sobre o Garbage Collector do Go em chamadas de função frequentes e laços de repetição.

### 8. Tratamento Nativo de Exceções (`tente / capture`)
Suporte a blocos de exceção via opcodes `OP_SETUP_TRY` e `OP_POP_TRY`. Exceções em runtime ou disparadas por `OP_LANCAR` realizam unwinding automático da pilha de operandos até a profundidade registrada e desviam o apontador de instrução (`IP`) para o bloco `capture`.

---

## ⚙️ Especificação do Conjunto de Instruções (Opcodes de 16 Bits)

| Instrução (Opcode) | Hex | Operandos | Descrição Semântica |
|:---|:---|:---|:---|
| `OP_PUSH_CONST` | `0x01` | `[2 bytes u16]` | Empilha constante do pool no topo da pilha. |
| `OP_POP` | `0x02` | Nenhum | Desempilha e libera o elemento do topo da pilha. |
| `OP_DUP` | `0x03` | Nenhum | Duplica o elemento do topo da pilha. |
| `OP_ADD` / `OP_SUB` / `OP_MUL` / `OP_DIV` | `0x04..0x07` | Nenhum | Operações aritméticas binárias polimórficas. |
| `OP_DIV_INT` / `OP_MOD` | `0x08..0x09` | Nenhum | Divisão inteira (`//`) e resto de divisão (`%`). |
| `OP_EQ` / `OP_NEQ` / `OP_LT` / `OP_LTE` / `OP_GT` / `OP_GTE` | `0x0A..0x0F` | Nenhum | Operadores relacionais e de comparação. |
| `OP_JMP` | `0x10` | `[2 bytes u16]` | Salto incondicional para o endereço absoluto no bytecode. |
| `OP_JMP_FALSO` | `0x11` | `[2 bytes u16]` | Salto condicional se topo for `Falso` ou `Nulo`. |
| `OP_CARREGAR_LOCAL` | `0x12` | `[2 bytes u16]` | Carrega valor do slot local $O(1)$ na pilha. |
| `OP_ARMAZENAR_LOCAL` | `0x13` | `[2 bytes u16]` | Armazena topo da pilha no slot local $O(1)$. |
| `OP_CARREGAR_GLOBAL` | `0x14` | `[2 bytes u16]` | Carrega variável global por nome no pool de constantes. |
| `OP_ARMAZENAR_GLOBAL` | `0x15` | `[2 bytes u16]` | Armazena valor em variável global pelo nome. |
| `OP_CHAMAR` | `0x16` | `[1 byte arity]` | Invoca função ou chamável com N argumentos. |
| `OP_RETORNE` | `0x17` | Nenhum | Retorna o valor do topo da pilha do frame atual. |
| `OP_AWAIT` | `0x18` | Nenhum | Aguarda resolução de promessa assíncrona. |
| `OP_CRIAR_FUNCAO` | `0x19` | `[2 bytes u16]` | Instancia objeto de função do pool de constantes. |
| `OP_RETORNE_CONST` | `0x1A` | `[2 bytes u16]` | Retorna constante diretamente do pool de constantes. |
| `OP_RETORNE_LOCAL` | `0x1B` | `[2 bytes u16]` | Retorna valor do slot local $O(1)$ diretamente. |
| `OP_NEG` / `OP_NOT` | `0x1C..0x1D` | Nenhum | Operadores unários (negação aritmética e lógica). |
| `OP_CRIAR_LISTA` | `0x1E` | `[2 bytes u16]` | Cria lista com N elementos desempilhados. |
| `OP_CRIAR_MAPA` | `0x1F` | `[2 bytes u16]` | Cria mapa com N pares chave-valor. |
| `OP_OBTER_INDICE` / `OP_DEFINIR_INDICE` | `0x20..0x21` | Nenhum | Leitura e escrita de elemento por índice (`a[i]`). |
| `OP_OBTER_ATRIBUTO` / `OP_DEFINIR_ATRIBUTO` | `0x22..0x23` | `[2 bytes u16]` | Acesso e atribuição de membro (`obj.attr`). |
| `OP_COALESCE` | `0x24` | `[2 bytes u16]` | Operador de coalescência nula (`??`). |
| `OP_ACESSAR_OPCIONAL` | `0x25` | `[2 bytes u16]` | Acesso a membro opcional (`obj?.attr`). |
| `OP_SETUP_TRY` | `0x26` | `[2 bytes u16]` | Configura handler de exceção com IP do bloco catch. |
| `OP_POP_TRY` | `0x27` | Nenhum | Remove handler de exceção do topo da pilha de try. |
| `OP_LANCAR` | `0x28` | Nenhum | Lança exceção com erro desempilhado do topo. |

---

## 🛠️ Exemplo Completo de Uso Prático

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
	codigoFuente := `
	var a = 10
	var b = 20
	retorne a * b + 5
	`

	ast, err := parser.NewParserFromString(codigoFuente, "exemplo.hrp").Parse()
	if err != nil {
		log.Fatalf("Erro no parser: %v", err)
	}

	contexto := hrp.NewContexto(hrp.OpcsContexto{})
	compilador := vm.NewCompilador(contexto)
	programa, err := compilador.Compilar(ast)
	if err != nil {
		log.Fatalf("Erro na compilação do bytecode: %v", err)
	}

	escopoRaiz := hrp.NewEscopo(nil)
	maquinaVirtual := vm.NewVM(contexto)
	frameInicial := vm.NewFrame(programa.Bytecode, programa.Constantes, escopoRaiz, nil, programa.NomesLocais)

	resultado, err := maquinaVirtual.Executar(frameInicial)
	if err != nil {
		log.Fatalf("Erro em runtime na VM: %v", err)
	}

	fmt.Printf("✓ Resultado retornado pela VM: %v\n", resultado)
}
```
