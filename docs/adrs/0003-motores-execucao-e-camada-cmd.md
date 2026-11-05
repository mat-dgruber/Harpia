<!--
================================================================================
LOG DE MANUTENÇÃO DE DOCUMENTAÇÃO
--------------------------------------------------------------------------------
Data       | Autor          | Descrição
--------------------------------------------------------------------------------
2026-11-05 | mat-dgruber    | Criação da ADR 0003 (proposta exploratória).
================================================================================
-->

# ADR 0003: Estratégia para os Motores de Execução e a Camada `cmd/`

| Metadado | Detalhe |
| :--- | :--- |
| **Status** | **Proposta** (2026-11-05). Exploratória: nenhuma refatoração decidida |
| **Decisores** | mat-dgruber |
| **Tags** | `arquitetura`, `runtime`, `compilador` |
| **Impacto Sistêmico** | Alto (`hrp/`, `vm/`, `cmd/transpiler_*`, `cmd/lsp.go`, `cmd/checar.go`) |
| **ADRs Relacionadas** | Estende ADR-0001 (paridade VM) e ADR-0002 (concorrência HTTP) |

## Contexto

O mesmo programa Harpia pode rodar por **quatro back-ends** com semânticas implementadas separadamente:

1. **Interpretador AST** (`hrp/interpretador.go`): caminho de fallback e referência informal.
2. **VM de bytecode** (`vm/`): caminho padrão de `executar`, com fallback para o AST quando o compilador não cobre uma construção (`hrp.ErrNaoSuportadoVM`).
3. **Transpilador web** (`cmd/transpiler_web.go` + `runtime-web.js`): JavaScript no navegador.
4. **Transpilador nativo** (`cmd/transpiler_native.go`): Go compilado.

O review de 2026-11-04 mostrou o custo dessa paridade manual:
- o operador `**` existia no lexer e no parser, mas em nenhum motor;
- o fallback VM→AST dependia de texto de erro e reexecutava programas;
- o overflow de `Inteiro` foi corrigido no AST/VM, mas o transpilador nativo ainda dá a volta (`int64` do Go) e o web perde precisão acima de 2^53 (`Number`).

Além disso, `cmd/` (cerca de 16 mil linhas) concentra estágios do compilador (transpiladores, LSP, formatador, linter, otimizador) dentro da camada de CLI. Isso dificulta testá-los como biblioteca e reaproveitá-los, por exemplo no playground WASM.

### Problemas a resolver
1. **Divergência semântica silenciosa** entre os motores (cada correção precisa ser feita até 4 vezes).
2. **Testes de paridade inexistentes:** não há uma suíte que rode o mesmo programa nos 4 motores e compare as saídas.
3. **Acoplamento do compilador à CLI.**

## 2. Opções em avaliação

| Critério | A: Suíte golden de paridade | B: Motor único (VM) + transpiladores a partir do bytecode | C: Extrair `compilador/` de `cmd/` |
| :--- | :--- | :--- | :--- |
| **Esforço** | Baixo/médio | Muito alto | Médio |
| **Reduz divergência** | Detecta (não previne) | Previne | Não muda |
| **Risco** | Baixo | Alto (reescreve os transpiladores) | Baixo (movimentação de código) |
| **Pré-requisito de** | B e C | — | Testes dos estágios como biblioteca |

**Direção sugerida (a confirmar):** começar por **A**, uma pasta `tests/paridade/` com programas `.hrp` e a saída esperada, executada pelo AST, pela VM, pelo nativo (compilando com `go run`) e pelo web (Node ou `goja`). Ela torna a divergência visível no CI. **C** pode vir em paralelo, sem mudar comportamento. **B** só se o custo de manter os 4 motores continuar alto depois de A.

## 3. Critério de decisão e próximo passo

- **Spike:** ter 20 programas golden cobrindo aritmética (incluindo overflow), texto, coleções, laços, exceções, classes e módulos, e medir quantos divergem entre os motores hoje.
- Se mais de 20% divergirem, priorizar A e corrigir; reavaliar B depois.
- Esta ADR passa a **Aprovada** quando a opção for escolhida após o spike.
