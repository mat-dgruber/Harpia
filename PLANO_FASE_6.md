# Plano de Implementação — Fase 6: Maturidade Corporativa

> Estado atual: 2026-07-16 | Resumo: 7/15 tarefas concluídas

---

## Legenda

- ✅ Concluído
- 🔄 Em andamento / parcial
- ⬜ Pendente
- 🔴 Bloqueado (dependência externa)

---

## 6.1 — Compilação AOT e Otimizações

| # | Tarefa | Status | Arquivos-chave |
|---|--------|--------|----------------|
| 1 | Empacotador Nativo (`harpia empacotar`) | ✅ | `cmd/empacotar.go` |
| 2 | Transpiler AOT (AST → Go) | ✅ | `cmd/transpiler_native.go` |
| 3 | Dead Code Elimination (DCE) | ⬜ | `cmd/transpiler_native.go` |

### 6.1.3 — Otimizações Estáticas (DCE)

**Escopo mínimo:** Percorrer AST removendo: declarações não referenciadas, branches com constante `falso`, expressões com efeitos colaterais nulos.

**Implementação:**
- Criar `cmd/otimizador.go` com `func Otimizar(ast *parser.Programa) *parser.Programa`
- Passo 1: Coletar todos os Identificadores referenciados
- Passo 2: Remover DeclVar sem referência
- Passo 3: Eliminar `se (falso) { ... }` e `enquanto (falso) { ... }`
- Integrar no pipeline do transpiler antes da geração Go
- Testes: `TestDCE_RemoveVarNaoReferenciada`, `TestDCE_RemoveBranchConstante`

**Esforço:** ~1 dia

---

## 6.2 — WASM/WASI

| # | Tarefa | Status | Bloqueio |
|---|--------|--------|----------|
| 1 | Alvo `--alvo=wasm` | ⬜ | Requer TinyGo ou wazero |
| 2 | WASI sandbox (arquivos/rede) | 🔴 | Go stdlib não suporta WASI nativamente |
| 3 | Segurança Runtime (XSS, secrets) | ⬜ | — |

### Recomendação

WASM é o item mais caro da Fase 6. Go compile WASM via `GOOS=js GOARCH=wasm` mas é pesado (~6MB). TinyGo gera binários menores (~100KB) mas tem compatibilidade parcial.

**Abordagem lazy:** Usar `wazero` (runtime WASM puro em Go, sem CGO) para executar WASM gerado. Começar com `--alvo=wasm` que gera `GOOS=js GOARCH=wasm` usando o transpiler existente.

**Esforço:** ~3-5 dias para MVP funcional

---

## 6.3 — LSP com Linter de Segurança

| # | Tarefa | Status | Arquivos-chave |
|---|--------|--------|----------------|
| 1 | Extensão VS Code | ✅ | `vscode-harpia/` |
| 2 | Linter de Segurança (SEC-001/002/003) | ✅ | `cmd/checar.go` |
| 3 | Copiloto Harpia Local (Ollama) | ⬜ | Novo: `cmd/copiloto.go` |

### 6.3.3 — Copiloto Local via Ollama

**Escopo mínimo:** Comando `harpia copiloto` que lê contexto do arquivo atual e envia para Ollama com prompt de completar código.

**Implementação:**
- `cmd/copiloto.go`: CLI que lê stdin/clipboard, monta prompt com contexto do projeto (imports, AST parcial), chama Ollama `/api/generate`
- Integrar ao LSP via `textDocument/completion` (sugestões contextuais)
- Prompt: "Complete o código Harpia abaixo de forma idiomática. Use a sintaxe portuguesa."

**Esforço:** ~2 dias

---

## 6.4 — SDK de IA

| # | Tarefa | Status | Arquivos-chave |
|---|--------|--------|----------------|
| 1 | Tipo nativo `agente` | ✅ | `stdlib/ia/agente.go` |
| 2 | Stdlib IA (Ollama/Gemini/OpenAI) | ✅ | `stdlib/ia/provedores.go` |
| 3 | Contratos Semânticos IA | ⬜ | Novo: `stdlib/ia/contratos.go` |

### 6.4.3 — Contratos Semânticos IA

**Escopo mínimo:** Função `validar_resposta(esquema, resposta)` que valida JSON de resposta da IA contra schema definido no código.

**Implementação:**
- `stdlib/ia/contratos.go`: `validar_resposta(esquema_mapa, texto_json) → (valido, erro)`
- Suporte a tipos: `texto`, `inteiro`, `decimal`, `booleano`, `lista`, `mapa`
- Validar campos obrigatórios, tipos, ranges
- Integrar ao agente: `agente.perguntar()` aceita parâmetro `esquema` opcional

**Esforço:** ~1 dia

---

## 6.5 — ORM Tipado e SQL Injection

| # | Tarefa | Status | Arquivos-chave |
|---|--------|--------|----------------|
| 1 | Conectores (MySQL, PostgreSQL, MongoDB) | ✅ | `stdlib/bd/bd.go` |
| 2 | Blindagem SQL Injection (linter) | ✅ | `cmd/checar.go` (HRP-SEC-001) |
| 3 | ORM Estático e Tipado | ⬜ | `stdlib/bd/orm.go` |
| 4 | Bancos Vetoriais (Chroma/Qdrant) | ⬜ | `stdlib/bd/vetorial.go` |

### 6.5.3 — ORM Estático e Tipado

**Escopo mínimo:** Mapeamento declarativo de tabelas com validação de campos em tempo de compilação.

**Implementação:**
- `stdlib/bd/orm.go`: tipo `Modelo` com campos tipados
- `bd.tabela("usuarios", bd.Campo("nome", bd.Texto), bd.Campo("idade", bd.Inteiro))`
- Query Builder fluente: `bd.tabela("usuarios").selecionar("nome").onde("idade", ">", 18).executar(conn)`
- Validação estática: campo inexistente gera erro de compilação

### 6.5.4 — Bancos Vetoriais

**Escopo mínimo:** Conector para Qdrant (mais simples que Chroma, API REST limpa).

**Implementação:**
- `stdlib/bd/vetorial.go`: `novo_vetorial(url, colecao) → cliente`
- `cliente.inserir(id, vetor, metadados)`, `cliente.buscar(vetor, limite)`, `cliente.deletar(id)`
- HTTP REST client para Qdrant API

**Esforço ORM:** ~2 dias | **Esforço Vetorial:** ~1 dia

---

## 6.6 — Microsserviços e Resiliência

| # | Tarefa | Status | Arquivos-chave |
|---|--------|--------|----------------|
| 1 | Resiliência (CB/RL/Retry) | ✅ | `stdlib/resiliencia/resiliencia.go` |
| 2 | OpenTelemetry (traces/métricas) | ⬜ | `stdlib/telemetria/telemetria.go` |
| 3 | OpenAPI + HMAC | ⬜ | `stdlib/http/openapi.go` |

### 6.6.2 — OpenTelemetry

**Escopo mínimo:** Wrapper leve que gera traces em JSON compatível com Jaeger/Prometheus.

**Implementação:**
- `stdlib/telemetria/telemetria.go`: `novo_tracer(nome)`, `iniciar_span(nome)`, `finalizar_span(span, status)`
- `nova_metrica(nome, tipo)` com counters e histograms
- Exportar em JSON para stdout ou arquivo (compatível com OpenTelemetry Collector via OTLP HTTP)
- Não depende do SDK oficial OpenTelemetry (evita dependência pesada)

### 6.6.3 — OpenAPI + HMAC

**Escopo mínimo:** Gerar spec OpenAPI 3.0 a partir de comentários `///` no código, e assinar requisições com HMAC SHA-256.

**Implementação:**
- `stdlib/http/openapi.go`: `gerar_openapi(servidor) → texto_json` — lê rotas registradas e gera spec
- `stdlib/http/hmac.go`: `assinar_hmac(chave, mensagem) → hex`, `verificar_hmac(chave, mensagem, assinatura) → bool`
- Integrar `assinar_hmac` como middleware do servidor HTTP

**Esforço Telemetria:** ~1.5 dias | **Esforço OpenAPI/HMAC:** ~1 dia

---

## Cronograma Resumido

| Subfase | Itens Pendentes | Esforço |
|---------|----------------|---------|
| 6.1 DCE | 1 | ~1 dia |
| 6.2 WASM/WASI | 3 | ~3-5 dias |
| 6.3 Copiloto | 1 | ~2 dias |
| 6.4 Contratos IA | 1 | ~1 dia |
| 6.5 ORM + Vetorial | 2 | ~3 dias |
| 6.6 Telemetria + OpenAPI | 2 | ~2.5 dias |
| **Total** | **10** | **~12-15 dias** |

### Ordem Recomendada (impacto/esforço)

1. **6.4.3 Contratos IA** — 1 dia, fecha módulo IA
2. **6.1.3 DCE** — 1 dia, melhora AOT
3. **6.6.3 OpenAPI/HMAC** — 1 dia, fecha microsserviços
4. **6.6.2 OpenTelemetry** — 1.5 dias, observabilidade
5. **6.5.4 Bancos Vetoriais** — 1 dia, fecha BD
6. **6.5.3 ORM Tipado** — 2 dias, melhora DX do BD
7. **6.3.3 Copiloto** — 2 dias, melhora DX do editor
8. **6.2 WASM/WASI** — 3-5 dias, o mais custoso, deixar por último
