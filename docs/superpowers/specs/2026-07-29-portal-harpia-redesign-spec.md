de continuidade

# Especificação de Design — Redesign da Página Principal do Portal Harpia

**Data**: 2026-07-29
**Status**: Aprovado
**Alvo**: `projetos/portal-harpia/web/rotas/index.hrp`

---

## 1. Visão Geral e Objetivos

Esta especificação define o redesign e enriquecimento da página principal do **Portal Harpia** (`index.hrp`). O objetivo é transformar a landing page principal em um hub completo do ecossistema full-stack Harpia, destacando soberania tecnológica brasileira, Clean Architecture, DDD, Sinais reativos, mentor de IA local, ferramentas do desenvolvedor, comparativos técnicos detalhados e FAQs abrangentes.

---

## 2. Identidade Visual & Cromia (Alinhada a BRAND_GUIDELINES.md)

A paleta oficial da Harpia segue o conceito de **Dark Mode Nativo**:

- **Fundo Primário**: `#0A0F0D` (Sombra da Floresta)
- **Fundo Secundário**: `#171E26` (Rio Profundo)
- **Destaque Principal / Foco**: `#F2A900` (Olho da Harpia)
- **Acento Secundário**: `#D45D34` (Terracota / Argila)
- **Acento de Sucesso**: `#00A86B` (Broto de Ipê)
- **Texto Principal**: `#F3F6F4` (Penagem Branca)
- **Texto de Apoio / Legendas**: `#8C9B9E` (Névoa da Manhã)
- **Tipografia**: Outfit / Inter para títulos e UI; JetBrains Mono / Fira Code para código.

---

## 3. Estrutura Detalhada das Seções

### Seção 1: Hero Banner (Cabeçalho de Impacto)

- **Logotipo Refinado**: Imagem do logo com aura iluminada em `#F2A900`.
- **Título**: "Linguagem de Programação Full-stack <span style="color: #F2A900">100% Brasileira</span>".
- **Subtítulo**: "Soberania, Precisão e Foco Absoluto. Full-stack do zero ao deploy com Clean Architecture, DDD e Reatividade nativa em português."
- **Badges**:
  - `v1.2.0 Estável`
  - `Clean Arch & DDD`
  - `Reatividade por Sinais`
  - `Mentor IA Local`
- **CTAs**: "Começar Agora" (`/docs/conceitos`) e "⭐ GitHub Oficial".

---

### Seção 2: Grid de Diferenciais & Inovações (8 Cards)

Cards interativos com efeito *glassmorphism* em `#171E26` com bordas sutis em `rgba(242, 169, 0, 0.15)`:

1. **Clean Architecture & DDD**: Estrutura nativa de diretórios (`dominio/`, `infra/`, `web/`, `testes/`).
2. **Reatividade por Sinais**: Mudanças cirúrgicas no DOM via `sinal()` e `efeito()` sem re-renderizar a árvore inteira.
3. **SPA + SSR Nativo**: Runtime VDOM ultraleve (~5KB) e renderização no servidor com `srv.servir_app()`.
4. **Mentor IA Local**: Diagnósticos amigáveis no terminal (`HRP-XXXX`) explicados por LLMs locais (Ollama / Llama.cpp) com `harpia erro explicar`.
5. **Harpia API Console (`/docs`)**: Documentação interativa e OpenAPI 3.0 dinâmico gerados automaticamente pelo backend.
6. **Linter Estático (`harpia checar`)**: Análise estática da AST sem execução (constantes, escopo, undeclared variables).
7. **Multi-banco Integrado (SQL & NoSQL)**: Query Builder para SQLite/PostgreSQL e suporte nativo a MongoDB e Redis na stdlib.
8. **Segurança OWASP & Cifragem**: Sanitização de entradas, headers de proteção HTTP, autorização e cifragem AES-256-GCM.

---

### Seção 3: Ecossistema Dev Tools (VS Code, TUI, CLI)

Apresentação visual e explicativa das ferramentas oficiais de desenvolvimento:

- **Extensão VS Code Harpia**: Syntax highlighting TextMate oficial, autocompletar e snippets.
- **TUI/REPL de 3 Painéis**: Editor/REPL, Inspetor de VM/Memória e Log/Saída no terminal.
- **Utilitários CLI**: `harpia migrar` (Migrations SQLite/Postgres), `harpia pwa` (Gerador PWA), `harpia i18n` (Extração de traduções), `harpia instalar` e `harpia pacote` (Gerenciador de Módulos e Lockfile Determinístico com SHA-256 `harpia.lock`).

---

### Seção 4: Quick Start (Instalação e Primeiros Passos em 3 Passos)

- Passo 1: Script de Instalação no Linux, macOS ou Windows.
- Passo 2: Criação de novo monolito DDD (`harpia novo meu-app --tipo=monolito`).
- Passo 3: Execução e desenvolvimento (`harpia executar`).

---

### Seção 5: Tabela Comparativa de Recursos Nativos

Comparação direta entre Harpia (.hrp), JavaScript/TypeScript, Go, Python e Rust/Elixir:

- Reatividade DOM por Sinais (~5KB)
- SSR + SPA Unificado na mesma linguagem
- Conectores Multi-banco na Stdlib
- Mentor IA Local no CLI (`harpia erro explicar`)
- Sintaxe, Compilador e Erros em Português (`HRP-XXXX`)
- Linter Estático sem dependências (`harpia checar`)

---

### Seção 6: Demonstração de Código Reativo & DDD

Exibição de exemplos limpos de código usando o componente `BlocoCodigo`:

- Exemplo 1: `dominio/usuario.hrp` (Entidade DDD, Sinais e Métodos).
- Exemplo 2: `servidor.hrp` (Rota HTTP, Query Builder e Injeção de Dependência).
- Exemplo 3: `web/componentes/contador.hrp` (Componente UI com Reatividade e Estilização Nativas).

---

### Seção 7: Perguntas Frequentes (FAQ Expandido - 8 Perguntas)

Perguntas sanadas em acordeões (`<details>`/`<summary>`):

1. O que é a Harpia e qual a sua proposta?
2. Como funciona o diagnóstico de erros com IA Local?
3. Quais alvos de compilação a Harpia suporta (VM, Web, WASM, WASI)?
4. Como funciona a reatividade por Sinais sem React/Solid?
5. A Harpia suporta bancos relacionais e NoSQL nativamente?
6. O que é o Harpia API Console (`/docs`)?
7. Como funciona a extensão oficial do VS Code?
8. Como posso contribuir para o projeto Harpia?

---

### Seção 8: Chamada para Ação & Comunidade

Card imponente convidando desenvolvedores para o GitHub, Guia de Contribuição e Licença Open-Source.

---

## 4. Plano de Verificação

Após a atualização de `projetos/portal-harpia/web/rotas/index.hrp`:

- Validar a compilação e execução do Portal Harpia.
- Verificar a integridade visual dos componentes e das cores oficiais do `BRAND_GUIDELINES.md`.
