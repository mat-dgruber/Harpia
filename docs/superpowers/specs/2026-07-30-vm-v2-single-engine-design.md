# Especificação do Motor Único Harpia VM V2

> **Data**: 2026-07-30  
> **Status**: Aprovado (Opção B - Motor Único Bytecode VM)  
> **Escopo**: Reformulação do motor de execução do Harpia para uma Máquina Virtual de Bytecode de alto desempenho, unificada de 16 bits.

---

## 1. Visão Geral e Arquitetura

O **Harpia VM V2** adota o modelo de **Motor Único de Execução (Single Engine Architecture)**. O interpretador legado baseado em caminhada direta na árvore sintática (AST tree-walk) é substituído por um compilador de bytecode em passagem única e um loop de execução plano `switch` com zero alocações por instrução.

### Metas Principais de Desempenho e Recursos
1. **Instruções e Operandos de 16 Bits (`uint16`)**:
   - Pools de constantes e deslocamentos de saltos (`OP_JMP`, `OP_JMP_FALSO`, etc.) passam a utilizar operandos de 16 bits em formato BigEndian (suportando até 65.536 constantes e instruções por função).
2. **Acesso $O(1)$ a Variáveis Locais por Slot**:
   - Resolução estática de escopo durante a compilação.
   - Variáveis locais são indexadas diretamente em uma fatia `Locais []hrp.Objeto` do `Frame`, eliminando buscas $O(N)$ em tabelas de símbolos de escopos encadeados.
3. **Loop de Despacho `switch` Plano Sem Alocações**:
   - Remoção do mecanismo de *threaded code* baseado em closures anonimas (`InstrucaoThreaded func(v *VM, frame *Frame)`).
   - Loop `switch` contínuo operando sobre `frame.Bytecode` com vetor de operandos e ponteiros de instrução numéricos diretos (`IP`).
4. **Completude Sintática de Bytecodes**:
   - Suporte nativo a coleções (`OP_CRIAR_LISTA`, `OP_CRIAR_MAPA`, `OP_OBTER_INDICE`, `OP_DEFINIR_INDICE`), acesso a membros (`OP_OBTER_ATRIBUTO`, `OP_DEFINIR_ATRIBUTO`), operadores nulo-seguros (`OP_COALESCE`, `OP_ACESSAR_OPCIONAL`), desestruturação, pattern matching (`escolha`) e tratamento de exceções (`OP_SETUP_TRY`, `OP_POP_TRY`, `OP_LANCAR`).

---

## 2. Estrutura de Dados e Opcodes do Harpia VM V2

### 2.1 Estrutura do Frame (`Frame`)

```go
type Frame struct {
    Pai          *Frame
    Funcao       *hrp.Funcao
    Bytecode     []byte
    IP           int
    Consts       []hrp.Objeto
    Pilha        []hrp.Objeto  // Pilha de operandos
    Locais       []hrp.Objeto  // Slots de variáveis locais O(1)
    Escopo       *hrp.Escopo   // Escopo fallback para globais e interop
    TryStack     []TryHandler  // Pilha de handlers de tente/capture
}

type TryHandler struct {
    CatchIP    int
    FinallyIP  int
    PilhaProf  int
}
```

### 2.2 Tabela Integrada de Opcodes (`vm/opcodes.go`)

| Opcode | Hex | Operandos | Descrição |
|---|---|---|---|
| `OP_PUSH_CONST` | `0x01` | `u16` (2 bytes) | Empilha constante do pool pelo índice. |
| `OP_POP` | `0x02` | - | Desempilha e libera o elemento no topo. |
| `OP_DUP` | `0x03` | - | Duplica o elemento no topo da pilha. |
| `OP_ADD` | `0x04` | - | Adição / concatenação. |
| `OP_SUB` | `0x05` | - | Subtração. |
| `OP_MUL` | `0x06` | - | Multiplicação. |
| `OP_DIV` | `0x07` | - | Divisão. |
| `OP_DIV_INT` | `0x08` | - | Divisão inteira (`//`). |
| `OP_MOD` | `0x09` | - | Resto da divisão (`%`). |
| `OP_EQ` | `0x0A` | - | Igualdade (`==`). |
| `OP_NEQ` | `0x0B` | - | Diferença (`!=`). |
| `OP_LT` | `0x0C` | - | Menor que (`<`). |
| `OP_LTE` | `0x0D` | - | Menor ou igual (`<=`). |
| `OP_GT` | `0x0E` | - | Maior que (`>`). |
| `OP_GTE` | `0x0F` | - | Maior ou igual (`>=`). |
| `OP_JMP` | `0x10` | `u16` (2 bytes) | Salto incondicional. |
| `OP_JMP_FALSO` | `0x11` | `u16` (2 bytes) | Salto se topo for falso ou nulo. |
| `OP_CARREGAR_LOCAL` | `0x12` | `u16` (2 bytes) | Carrega variável local por slot $O(1)$. |
| `OP_ARMAZENAR_LOCAL` | `0x13` | `u16` (2 bytes) | Armazena variável local no slot $O(1)$. |
| `OP_CARREGAR_GLOBAL` | `0x14` | `u16` (2 bytes) | Carrega variável global por nome no pool. |
| `OP_ARMAZENAR_GLOBAL` | `0x15` | `u16` (2 bytes) | Armazena variável global por nome no pool. |
| `OP_CHAMAR` | `0x16` | `u8` (1 byte aridade) | Invoca chamável. |
| `OP_RETORNE` | `0x17` | - | Retorna o valor no topo. |
| `OP_AWAIT` | `0x18` | - | Await de promessas. |
| `OP_CRIAR_FUNCAO` | `0x19` | `u16` (const idx) | Instancia objeto de função compilada. |
| `OP_RETORNE_CONST` | `0x1A` | `u16` (2 bytes) | Retorna constante diretamente do pool. |
| `OP_RETORNE_LOCAL` | `0x1B` | `u16` (2 bytes) | Retorna local diretamente do slot. |
| `OP_NEG` | `0x1C` | - | Negação numérica (`-x`). |
| `OP_NOT` | `0x1D` | - | Negação lógica (`nao x`). |
| `OP_CRIAR_LISTA` | `0x1E` | `u16` (qtd elementos) | Cria `Lista` a partir de N elementos da pilha. |
| `OP_CRIAR_MAPA` | `0x1F` | `u16` (qtd pares) | Cria `Mapa` a partir de 2*N elementos da pilha. |
| `OP_OBTER_INDICE` | `0x20` | - | Acesso por índice `alvo[idx]`. |
| `OP_DEFINIR_INDICE` | `0x21` | - | Atribuição por índice `alvo[idx] = val`. |
| `OP_OBTER_ATRIBUTO` | `0x22` | `u16` (const idx) | Acesso a membro `alvo.membro`. |
| `OP_DEFINIR_ATRIBUTO` | `0x23` | `u16` (const idx) | Atribuição a membro `alvo.membro = val`. |
| `OP_COALESCE` | `0x24` | `u16` (jmp offset) | Operador nulo-seguro `??`. |
| `OP_ACESSAR_OPCIONAL` | `0x25` | `u16` (const idx) | Operador de acesso opcional `?.`. |
| `OP_SETUP_TRY` | `0x26` | `u16` (catch IP) | Registra handler de exceção. |
| `OP_POP_TRY` | `0x27` | - | Remove handler de exceção ativo. |
| `OP_LANCAR` | `0x28` | - | Lança exceção do topo da pilha. |

---

## 3. Análise Estática de Escopos e Compilador

O `Compilador` em `vm/compilador.go` mantém uma tabela de símbolos estáticos em tempo de compilação:

```go
type EscopoCompilador struct {
    Pai      *EscopoCompilador
    Locais   []LocalVar
    Count    int
}

type LocalVar struct {
    Nome  string
    Depth int
    Slot  int
}
```

Ao encontrar uma declaração `var x`, o compilador resolve se `x` é uma variável local (se estiver em uma função/bloco) ou global, atribuindo o slot numérico correspondente (`0, 1, 2...`). As instruções emitidas para variáveis locais passam a ser estritamente `OP_CARREGAR_LOCAL <slot_u16>` e `OP_ARMAZENAR_LOCAL <slot_u16>`.

---

## 4. Integração do Motor Único (`hrp/executar.go`)

Todas as vias de execução do Harpia (CLI `Harpia arquivo.hrp`, REPL, testes e servidores) passam a utilizar a compilação para Bytecode e a execução na VM:

```go
func ExecutarCodigo(codigo string, nomeArquivo string, ctx *hrp.Contexto) (hrp.Objeto, error) {
    p := parser.NewParserFromString(codigo, nomeArquivo)
    ast, err := p.Parse()
    if err != nil {
        return nil, err
    }

    comp := vm.NewCompilador()
    prog, err := comp.Compilar(ast)
    if err != nil {
        return nil, err
    }

    execVM := vm.NewVM(ctx)
    frame := vm.NewFrame(prog.Bytecode, prog.Constantes, ctx.EscopoGlobal, nil)
    return execVM.Executar(frame)
}
```

---

## 5. Plano de Validação e Testes

1. **Testes Unitários da VM (`vm/vm_test.go`)**:
   - Validação de opcodes de 16 bits.
   - Resolução e atibuição de variáveis locais por slot $O(1)$.
   - Criação e indexação de listas e mapas na VM.
   - Tratamento de exceções `tente / capture`.
2. **Suíte Completa do Harpia (`go test ./...`)**:
   - Garantir 100% de aprovação de todos os testes sintáticos e semânticos do projeto no novo motor.
3. **Benchmarks de Desempenho**:
   - Comparação de throughput e alocações antes e depois da eliminação do *threaded code*.
