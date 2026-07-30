# Harpia VM V2 — Motor Único Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Transform Harpia's execution engine into a unified, high-performance 16-bit Bytecode Virtual Machine with zero-alloc `switch` dispatch, $O(1)$ local slots, and full feature completeness.

**Architecture:** Replace closure-based threaded code with a flat Go `switch` loop over raw bytecode. Use 16-bit BigEndian operands for constants/jumps and static slot indexing for local variables in `Frame.Locais`. Unify execution across CLI, REPL, and runtime.

**Tech Stack:** Go 1.24, Harpia AST, `encoding/binary`, `sync.Pool`.

## Global Constraints
- Operands for constant pools and jump targets must be 16-bit BigEndian (`uint16`).
- Local variables must use $O(1)$ slot indexing (`Frame.Locais []hrp.Objeto`).
- Dispatch loop must be zero-alloc flat `switch` over `frame.Bytecode`.
- 100% test compatibility with `go test ./...`.

---

### Task 1: Refatorar Opcodes de 16 Bits e Estruturas da VM

**Files:**
- Modify: `vm/opcodes.go`
- Modify: `vm/vm.go`
- Modify: `vm/vm_test.go`

**Interfaces:**
- Consumes: AST da linguagem Harpia
- Produces: Novas constantes de `Opcode` em `vm/opcodes.go` de `0x01` a `0x28` e assinaturas auxiliares `NomeOpcode`

- [ ] **Step 1: Escrever teste focado para novos opcodes**

Criar `TestOpcodesV2` em `vm/vm_test.go` validando a resolução das strings e valores dos novos opcodes.

- [ ] **Step 2: Executar o teste e verificar falha**

Run: `go test ./vm -v -run TestOpcodesV2`

- [ ] **Step 3: Atualizar `vm/opcodes.go` com os opcodes de 16-bits e novas instruções**

Adicionar opcodes: `OP_CARREGAR_LOCAL`, `OP_ARMAZENAR_LOCAL`, `OP_CARREGAR_GLOBAL`, `OP_ARMAZENAR_GLOBAL`, `OP_CRIAR_LISTA`, `OP_CRIAR_MAPA`, `OP_OBTER_INDICE`, `OP_DEFINIR_INDICE`, `OP_OBTER_ATRIBUTO`, `OP_DEFINIR_ATRIBUTO`, `OP_COALESCE`, `OP_ACESSAR_OPCIONAL`, `OP_SETUP_TRY`, `OP_POP_TRY`, `OP_LANCAR`.

- [ ] **Step 4: Executar o teste para verificar aprovação**

Run: `go test ./vm -v -run TestOpcodesV2`

- [ ] **Step 5: Commit**

`git commit -m "refactor(vm): define 16-bit opcode set and new v2 instructions"`

---

### Task 2: Implementar Loop `switch` Zero-Alloc e Slots de Locais $O(1)$ na VM

**Files:**
- Modify: `vm/vm.go`
- Test: `vm/vm_test.go`

**Interfaces:**
- Consumes: `Frame` com `Bytecode []byte`, `Consts []hrp.Objeto`, `Locais []hrp.Objeto`
- Produces: `func (v *VM) Executar(frame *Frame) (hrp.Objeto, error)` sem alocações no loop principal

- [ ] **Step 1: Escrever teste focado em `Executar` com opcodes de 16 bits e slots de locais**

Adicionar teste em `vm/vm_test.go` testando `OP_PUSH_CONST` de 16 bits e `OP_CARREGAR_LOCAL`/`OP_ARMAZENAR_LOCAL` com slots.

- [ ] **Step 2: Executar teste e verificar falha**

Run: `go test ./vm -v -run TestVMSwitchLoop`

- [ ] **Step 3: Reescrever `vm/vm.go` com o loop `switch` plano e suporte a slots de locais**

Substituir `compilarThreadedCode` pelo loop `switch (op)` direto em `Executar`, lendo operandos de 16 bits com `binary.BigEndian.Uint16`.

- [ ] **Step 4: Executar teste e verificar aprovação**

Run: `go test ./vm -v -run TestVMSwitchLoop`

- [ ] **Step 5: Commit**

`git commit -m "feat(vm): implement zero-alloc flat switch execution loop with O(1) local slots"`

---

### Task 3: Atualizar `vm/compilador.go` com Operandos de 16-Bits e Resolução Estática de Escopo

**Files:**
- Modify: `vm/compilador.go`
- Test: `vm/vm_test.go`

**Interfaces:**
- Consumes: AST de `parser.BaseNode`
- Produces: `ProgramaCompilado` com bytecode codificado com operandos `u16` e opcodes `OP_CARREGAR_LOCAL` / `OP_CARREGAR_GLOBAL`

- [ ] **Step 1: Escrever teste para emissão de opcodes de 16 bits no compilador**

Adicionar teste em `vm/vm_test.go` para compilar atribuição local e operações matemáticas emitindo `u16`.

- [ ] **Step 2: Executar o teste e verificar falha**

Run: `go test ./vm -v -run TestCompiladorV2`

- [ ] **Step 3: Refatorar `Compilador` para manter escopo estático e emitir operandos de 2 bytes**

Implementar `EscopoCompilador`, gerenciar slots numéricos locais e atualizar `emitir` para lidar com operandos de 16 bits.

- [ ] **Step 4: Executar o teste e verificar aprovação**

Run: `go test ./vm -v -run TestCompiladorV2`

- [ ] **Step 5: Commit**

`git commit -m "feat(compilador): add static scope resolution and 16-bit operand emission"`

---

### Task 4: Expandir Compilador para Coleções, Atributos e Operadores Nulo-Seguros

**Files:**
- Modify: `vm/compilador.go`
- Modify: `vm/vm.go`
- Test: `vm/vm_test.go`

**Interfaces:**
- Consumes: Nós de AST para literais de lista/mapa, indexação, acesso a membro (`.`), `??` e `?.`
- Produces: Execução nativa dessas estruturas em bytecode na VM

- [ ] **Step 1: Escrever testes para coleções, membros e nulo-seguros na VM**

Adicionar testes em `vm/vm_test.go` para `[1, 2]`, `{ "a": 10 }`, `obj.membro`, `a ?? b` e `obj?.membro`.

- [ ] **Step 2: Executar testes e verificar falhas**

Run: `go test ./vm -v -run TestVMRecursosAvancados`

- [ ] **Step 3: Implementar visitação na AST e handlers no `switch` da VM**

Implementar compilação dos nós de coleções e operadores nulo-seguros no `compilador.go` e seus equivalentes no `switch` da `vm.go`.

- [ ] **Step 4: Executar testes e verificar aprovação**

Run: `go test ./vm -v -run TestVMRecursosAvancados`

- [ ] **Step 5: Commit**

`git commit -m "feat(vm): add bytecodes for collections, member access, and null-safe operators"`

---

### Task 5: Implementar Tratamento de Exceções (`tente / capture`) no Compilador e VM

**Files:**
- Modify: `vm/compilador.go`
- Modify: `vm/vm.go`
- Test: `vm/vm_test.go`

**Interfaces:**
- Consumes: Nós de AST `TenteCaptureFinalmente` e `LancarNode`
- Produces: Mecanismo de try/catch baseado em `TryStack` na VM

- [ ] **Step 1: Escrever teste de try/catch na VM**

Testar recuperação de exceção com `tente { lancar "erro" } capture (e) { ... }`.

- [ ] **Step 2: Executar teste e verificar falha**

Run: `go test ./vm -v -run TestVMExcecoes`

- [ ] **Step 3: Implementar `OP_SETUP_TRY`, `OP_POP_TRY` e `OP_LANCAR`**

Adicionar `TryStack` ao `Frame`, compilar `TenteCaptureFinalmente` e gerenciar o salto para o bloco `capture` na VM ao lançar exceção.

- [ ] **Step 4: Executar teste e verificar aprovação**

Run: `go test ./vm -v -run TestVMExcecoes`

- [ ] **Step 5: Commit**

`git commit -m "feat(vm): implement try/catch exception handling mechanism"`

---

### Task 6: Unificar a Execução do Harpia no Motor Único VM (`hrp/executar.go`)

**Files:**
- Modify: `hrp/executar.go`
- Modify: `cmd/executar.go`
- Test: `tests/` suíte geral

**Interfaces:**
- Consumes: Código-fonte Harpia
- Produces: Execução via Compilador + VM em todas as rotas do compilador

- [ ] **Step 1: Escrever/verificar teste de integração em `hrp/`**

Garantir que a chamada de `hrp.ExecutarCodigo` compile e execute via Bytecode VM.

- [ ] **Step 2: Executar testes da suíte principal do Harpia**

Run: `go test ./hrp/...`

- [ ] **Step 3: Atualizar `hrp/executar.go` para executar via `vm.VM`**

Redirecionar a execução padrão do Harpia para compilar para bytecode e rodar na VM V2.

- [ ] **Step 4: Executar todos os testes do Harpia**

Run: `go test ./...`

- [ ] **Step 5: Commit**

`git commit -m "feat(runtime): unify execution engine under Bytecode VM V2"`

---

### Task 7: Atualizar Documentação, READMEs, Manual e LLM Text Files

**Files:**
- Modify: `vm/README.md`
- Modify: `Manual.md`
- Modify: `ROADMAP.md`
- Modify: `llms.txt`
- Modify: `llms-full.txt`

- [ ] **Step 1: Atualizar `vm/README.md`**

Documentar os novos opcodes de 16 bits, slots de variáveis $O(1)$ e o motor único.

- [ ] **Step 2: Atualizar `Manual.md`, `ROADMAP.md`, `llms.txt` e `llms-full.txt`**

Refletir a unificação do motor de execução sob a Bytecode VM V2.

- [ ] **Step 3: Executar a suíte de testes final**

Run: `go test ./...`

- [ ] **Step 4: Commit das documentações**

`git commit -m "docs: update vm readme, manual, roadmap, and llms files for vm v2"`
