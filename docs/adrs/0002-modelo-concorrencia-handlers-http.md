<!--
================================================================================
LOG DE MANUTENÇÃO DE DOCUMENTAÇÃO
--------------------------------------------------------------------------------
Data       | Autor          | Descrição
--------------------------------------------------------------------------------
2026-11-04 | mat-dgruber    | Criação da ADR 0002 (serialização dos handlers HTTP).
================================================================================
-->

# ADR 0002: Serializar a Execução dos Handlers HTTP por Servidor

| Metadado | Detalhe |
| :--- | :--- |
| **Status** | **Implementada** (2026-11-04) |
| **Decisores** | mat-dgruber (aprovou o plano `docs/ai/plans/2026-11-04-correcoes-review-max.md`, Task 9) |
| **Tags** | `concorrencia`, `runtime`, `stdlib/http` |
| **Impacto Sistêmico** | Médio (`stdlib/http`: vazão de servidores Harpia) |
| **ADRs Relacionadas** | Estende ADR-0001 (paridade VM e concorrência) |

## Contexto

O `net/http` do Go atende cada requisição numa goroutine própria, e o `stdlib/http` do Harpia chamava o handler do script diretamente nessa goroutine. Só que as coleções do runtime (`hrp.Mapa` é um `map[string]Objeto` cru; `hrp.Lista` é uma fatia sem trava) não são seguras para uso concorrente.

### Riscos Identificados e Problemas a Resolver
1. **Corrupção silenciosa:** um contador num mapa global (`contagem["x"] = contagem["x"] + 1`) perdia 53 de 400 incrementos sob 32 conexões simultâneas (`stdlib/http_concorrencia_test.go`).
2. **Queda do processo:** escrita concorrente em `map` pode disparar `fatal error: concurrent map writes`, que não tem `recover` e derruba o servidor inteiro.
3. **Expectativa do usuário:** quem escreve um script Harpia não espera precisar de locks; a linguagem é voltada a ensino.

## 2. Decisão de Arquitetura

Cada `Servidor` tem um semáforo de capacidade 1 (`execSem`). A cadeia de middlewares mais o handler roda **um por vez por servidor**. Leitura do corpo, roteamento e escrita da resposta continuam concorrentes. Uma requisição espera no máximo `EsperaMaximaHandler` (5s) e, passado esse prazo, recebe `503 Service Unavailable` com `Retry-After: 1`.

### Matriz de Avaliação de Alternativas

| Critério | A: Semáforo por servidor (eleita) | B: Mutex interno em `Mapa`/`Lista` | C: Um interpretador por requisição |
| :--- | :--- | :--- | :--- |
| **Esforço** | Baixo (1 arquivo) | Alto (toda a hierarquia de tipos, hot path da VM) | Alto (estado global do script deixa de ser compartilhado) |
| **Vazão** | 1 handler por vez | Concorrente | Concorrente |
| **Correção** | Total para o código dos handlers | Parcial (operações compostas como `x = x + 1` continuam não atômicas) | Total, mas muda a semântica (globais não persistem entre requisições) |
| **Reversibilidade** | Alta | Média | Baixa |

**Por que A:** é a menor mudança que elimina corrupção e queda. B não resolve `x = x + 1` sem transações, e C quebra o modelo mental de "variável global do servidor".

## 3. Consequências

- **Vazão:** fica limitada a aproximadamente 1 / latência do handler. Um handler que espera I/O longo (ex.: chamada a um LLM) segura os demais.
- **Auto-chamada:** um handler que faz `requisitar` ao próprio servidor recebe 503 depois de 5s, em vez de travar até o timeout de 30s do cliente. Esse padrão não é suportado.
- **Limite conhecido:** funções `assincrono` disparadas dentro de um handler rodam fora do semáforo e continuam podendo mutar mapas em paralelo. O mesmo vale para o script principal que continua rodando depois de um `escutar` não bloqueante.
- **Gatilho de revisão:** quando houver meta de RPS, ou quando um usuário relatar 503 por handlers lentos, reavaliar B (coleções thread-safe com operações atômicas) ou um GIL por contexto liberado em chamadas de I/O.
