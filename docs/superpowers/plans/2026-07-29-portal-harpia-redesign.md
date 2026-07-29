# Portal Harpia Redesign Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Reformular a página principal do Portal Harpia (`projetos/portal-harpia/web/rotas/index.hrp`) transformando-a em um hub completo do ecossistema full-stack Harpia com Hero renovado, 8 cards de diferenciais, seção de Dev Tools (VS Code, TUI, CLI), Quick Start, Tabela Comparativa de linguagens, Demonstração de Código com abas, FAQ expandido com 8 perguntas e CTA de comunidade.

**Architecture:** Frontend SPA em Harpia (.hrp) com Virtual DOM leve, utilizando estilos inline CSS alinhados à paleta oficial em `BRAND_GUIDELINES.md` (`#0A0F0D`, `#171E26`, `#F2A900`, `#D45D34`, `#00A86B`, `#F3F6F4`, `#8C9B9E`), componentes reutilizáveis (`BlocoCodigo`, `Layout`) e dados estruturados em coleções Harpia.

**Tech Stack:** Harpia (.hrp), HTML5 semântico, CSS3 (Glassmorphism, Grid Layout, Flexbox), JavaScript runtime nativo Harpia.

## Global Constraints

- Fundo Primário: `#0A0F0D` (Sombra da Floresta)
- Fundo Secundário: `#171E26` (Rio Profundo)
- Destaque Principal: `#F2A900` (Olho da Harpia)
- Acento Secundário: `#D45D34` (Terracota)
- Acento Sucesso: `#00A86B` (Broto de Ipê)
- Texto Principal: `#F3F6F4` (Penagem Branca)
- Texto Apoio: `#8C9B9E` (Névoa da Manhã)
- Responsividade: Layout adaptável para telas móveis e desktop (`repeat(auto-fit, minmax(...))`).

---

### Task 1: Atualização das Coleções de Dados e Estrutura do Hero Banner em `index.hrp`

**Files:**
- Modify: `projetos/portal-harpia/web/rotas/index.hrp`

**Interfaces:**
- Consumes: Componentes `Layout`, `BlocoCodigo` e imagens em `/web/assets/img/`.
- Produces: Variáveis `conceitos` (8 itens), `diferenciais`, `devTools`, `faqs` e estrutura da rota `RotaIndex()`.

- [ ] **Step 1: Atualizar o array de conceitos e diferenciais em `index.hrp`**

Adicionar todos os 8 diferenciais e coleções de dados com os textos exatos da especificação de design:
```harpia
var conceitos = [
    { "icone": "🛡️", "titulo": "Clean Architecture & DDD", "descricao": "Diretórios dominio/, infra/, web/ e testes/ nativamente organizados com suporte a Domain Driven Design." },
    { "icone": "⚡", "titulo": "Reatividade por Sinais", "descricao": "Atualizações cirúrgicas do DOM estilo Solid.js com `sinal()` e `efeito()` sem re-renderizar a árvore inteira." },
    { "icone": "🖥️", "titulo": "SPA + SSR Nativo", "descricao": "Frontend reativo com Virtual DOM leve (~5KB) e SSR integrado com `srv.servir_app()`, sem frameworks externos." },
    { "icone": "🤖", "titulo": "Mentor IA Local", "descricao": "Diagnósticos amigáveis no terminal (`HRP-XXXX`) explicados por LLMs locais (Ollama / Llama.cpp) com `harpia erro explicar`." },
    { "icone": "📄", "titulo": "API Console Nativo (/docs)", "descricao": "Documentação interativa Swagger/OpenAPI 3.0 dinâmico gerado automaticamente pelas anotações de rota." },
    { "icone": "🔍", "titulo": "Linter Estático (harpia checar)", "descricao": "Análise estática da AST sem executar o código: validação de escopo, variáveis não declaradas e reatribuições de const." },
    { "icone": "🗄️", "titulo": "Multi-banco Integrado", "descricao": "Query Builder para SQLite/PostgreSQL e suporte nativo a NoSQL (MongoDB e Redis) diretamente na stdlib." },
    { "icone": "🔒", "titulo": "Segurança OWASP & Cifragem", "descricao": "Sanitização de entrada, cabeçalhos de segurança HTTP e cifragem de alta segurança AES-256-GCM na stdlib." }
]
```

- [ ] **Step 2: Implementar o Hero Banner Renovado**

Atualizar o HTML/JSX do Hero com títulos, cores oficiais (`#F2A900`), subtítulo e badges:
```harpia
<!-- Hero Section -->
<section style="position: relative; min-height: 85vh; display: flex; align-items: center; justify-content: center; overflow: hidden; padding: 120px 24px 80px; text-align: center; background: linear-gradient(to bottom, rgba(10, 15, 13, 0.75) 0%, rgba(10, 15, 13, 0.95) 100%), url('/web/assets/img/hero_bg_floresta_harpia.jpg') center/cover no-repeat;">
    <div style="max-width: 900px; margin: 0 auto; position: relative; z-index: 10;">
        <div style="margin-bottom: 24px; display: flex; justify-content: center;">
            <img src="/web/assets/img/logo_harpia_refinado.jpg" alt="Harpia Logo Refinado" style="width: 140px; height: 140px; border-radius: 50%; border: 3px solid #F2A900; box-shadow: 0 0 50px rgba(242, 169, 0, 0.45); object-fit: cover;" />
        </div>
        <div style="display: flex; gap: 8px; justify-content: center; flex-wrap: wrap; margin-bottom: 20px;">
            <span style="background: rgba(242, 169, 0, 0.15); border: 1px solid rgba(242, 169, 0, 0.4); color: #F2A900; font-size: 0.8rem; font-weight: 700; padding: 4px 12px; border-radius: 12px;">v1.2.0 Estável</span>
            <span style="background: rgba(0, 168, 107, 0.15); border: 1px solid rgba(0, 168, 107, 0.4); color: #00A86B; font-size: 0.8rem; font-weight: 700; padding: 4px 12px; border-radius: 12px;">Clean Arch & DDD</span>
            <span style="background: rgba(212, 93, 52, 0.15); border: 1px solid rgba(212, 93, 52, 0.4); color: #D45D34; font-size: 0.8rem; font-weight: 700; padding: 4px 12px; border-radius: 12px;">Reatividade por Sinais</span>
        </div>
        <h1 style="font-size: 3.5rem; font-weight: 800; color: #F3F6F4; margin-bottom: 20px; line-height: 1.15; letter-spacing: -1px;">
            Linguagem de Programação Full-stack <br/><span style="color: #F2A900; text-shadow: 0 0 25px rgba(242, 169, 0, 0.4);">100% Brasileira</span>
        </h1>
        <p style="font-size: 1.35rem; color: #8C9B9E; max-width: 750px; margin: 0 auto 40px; line-height: 1.6;">
            Soberania, Precisão e Foco Absoluto. Full-stack do zero ao deploy com Clean Architecture, DDD e Reatividade nativa em português.
        </p>
        <div style="display: flex; justify-content: center; gap: 20px; flex-wrap: wrap;">
            <a href="/docs/conceitos" class="BotaoPrimario" style="text-decoration: none; font-size: 1.1rem; padding: 14px 36px; box-shadow: 0 0 25px rgba(242, 169, 0, 0.35);">
                Começar Agora
            </a>
            <a href="https://github.com/mat-dgruber/Harpia" target="_blank" class="BotaoSecundario" style="text-decoration: none; font-size: 1.1rem; padding: 14px 36px; display: flex; align-items: center; gap: 8px;">
                ⭐ GitHub
            </a>
        </div>
    </div>
</section>
```

- [ ] **Step 3: Testar e verificar compilação inicial**

Executar a compilação do portal para garantir que o arquivo `index.hrp` não possui erros de sintaxe.
Run: `cd /Users/matheus.diniz_1/Documents/GitHub/Harpia/projetos/portal-harpia && go run . checar` (ou compilar no diretório Harpia principal).

- [ ] **Step 4: Commit das alterações do Hero Banner**

```bash
git add projetos/portal-harpia/web/rotas/index.hrp
git commit -m "feat(portal): renovar hero banner e colecoes de dados da pagina principal"
```

---

### Task 2: Implementação das Seções de Diferenciais (8 Cards), Dev Tools e Quick Start

**Files:**
- Modify: `projetos/portal-harpia/web/rotas/index.hrp`

**Interfaces:**
- Consumes: Ícones e layout responsivo CSS.
- Produces: Seções de grid de diferenciais, ecossistema dev tools e bloco de comandos CLI em 3 passos.

- [ ] **Step 1: Renderizar o Grid dos 8 Diferenciais Nativos**

Renderizar os 8 cards de diferenciais com layout responsivo e efeito glassmorphic:
```harpia
<!-- Diferenciais & Inovações Grid (8 Cards) -->
<section style="padding: 64px 24px; max-width: 1200px; margin: 0 auto;">
    <div style="text-align: center; margin-bottom: 48px;">
        <h2 style="font-size: 2.25rem; font-weight: 800; color: #F3F6F4; margin-bottom: 12px;">
            Recursos & <span style="color: #F2A900;">Diferenciais Nativos</span>
        </h2>
        <p style="color: #8C9B9E; font-size: 1.15rem; max-width: 700px; margin: 0 auto;">
            Tudo o que você precisa para construir software moderno e robusto sem dependências externas.
        </p>
    </div>
    <div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(260px, 1fr)); gap: 24px;">
        <!-- Iteração sobre conceitos[0] até conceitos[7] -->
    </div>
</section>
```

- [ ] **Step 2: Adicionar a Seção de Dev Tools (Extensão VS Code, TUI e CLI)**

Criar a seção de ferramentas do ecossistema:
```harpia
<!-- Dev Tools Section -->
<section style="padding: 64px 24px; max-width: 1200px; margin: 0 auto; background: rgba(23, 30, 38, 0.4); border-radius: 24px; border: 1px solid rgba(242, 169, 0, 0.15);">
    <div style="text-align: center; margin-bottom: 40px;">
        <h2 style="font-size: 2rem; font-weight: 800; color: #F3F6F4; margin-bottom: 12px;">
            Ecossistema de <span style="color: #F2A900;">Ferramentas Dev</span>
        </h2>
        <p style="color: #8C9B9E; font-size: 1.1rem;">Produtividade máxima da IDE ao terminal.</p>
    </div>
    <div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); gap: 24px;">
        <div style="background: rgba(10, 15, 13, 0.6); padding: 28px; border-radius: 16px; border: 1px solid rgba(242, 169, 0, 0.2);">
            <h3 style="color: #F2A900; font-size: 1.25rem; margin-bottom: 8px;">🧩 Extensão VS Code</h3>
            <p style="color: #8C9B9E; font-size: 0.95rem; line-height: 1.6;">Realce de sintaxe TextMate, autocompletar inteligente de palavras-chave em português e suporte nativo a snippets .hrp.</p>
        </div>
        <div style="background: rgba(10, 15, 13, 0.6); padding: 28px; border-radius: 16px; border: 1px solid rgba(242, 169, 0, 0.2);">
            <h3 style="color: #F2A900; font-size: 1.25rem; margin-bottom: 8px;">🖥️ TUI & REPL (3 Painéis)</h3>
            <p style="color: #8C9B9E; font-size: 0.95rem; line-height: 1.6;">Interface interativa de terminal com Editor REPL, Inspetor de Memória/VM e Logs de execução em tempo real.</p>
        </div>
        <div style="background: rgba(10, 15, 13, 0.6); padding: 28px; border-radius: 16px; border: 1px solid rgba(242, 169, 0, 0.2);">
            <h3 style="color: #F2A900; font-size: 1.25rem; margin-bottom: 8px;">🛠️ Utilitários CLI Nativos</h3>
            <p style="color: #8C9B9E; font-size: 0.95rem; line-height: 1.6;">Gerenciador de Migrations, gerador de PWA para dist/ e extrator i18n automatizado incorporados no compilador.</p>
        </div>
    </div>
</section>
```

- [ ] **Step 3: Commit das seções de diferenciais e dev tools**

```bash
git add projetos/portal-harpia/web/rotas/index.hrp
git commit -m "feat(portal): adicionar grid de 8 diferenciais e secao de dev tools no portal"
```

---

### Task 3: Implementação da Tabela Comparativa, Demonstrações de Código, FAQ Expandido e CTA Final

**Files:**
- Modify: `projetos/portal-harpia/web/rotas/index.hrp`

**Interfaces:**
- Consumes: Componente `BlocoCodigo`, tags `<details>` / `<summary>`.
- Produces: Tabela comparativa com Rust/Elixir, exemplos de código DDD/HTTP e FAQ com 8 perguntas.

- [ ] **Step 1: Atualizar a Tabela Comparativa Tecnológica**

Adicionar as linhas comparativas com JavaScript/TS, Go, Python e Rust/Elixir com destaque verde `#00A86B`:
- Reatividade DOM (Sinais) (~5KB)
- SSR + SPA na mesma linguagem
- Multi-banco na stdlib (SQL/NoSQL/Redis)
- Mentor IA Local no CLI (`harpia erro explicar`)
- Sintaxe & Erros em Português (`HRP-XXXX`)
- Linter Estático Integrado (`harpia checar`)

- [ ] **Step 2: Adicionar os Exemplos de Código Reativo e DDD**

Adicionar os blocos de código usando o componente `BlocoCodigo`:
1. `dominio/usuario.hrp` (Entidade DDD & Sinais)
2. `servidor.hrp` (API REST & Query Builder)
3. `web/componentes/contador.hrp` (Componente UI com Reatividade e Estilo)

- [ ] **Step 3: Implementar o FAQ Expandido com 8 Perguntas**

Adicionar todas as 8 perguntas em formato `<details>`/`<summary>` estilizados com bordas em `rgba(242, 169, 0, 0.15)` e destaque dourado `#F2A900` para a pergunta expandida.

- [ ] **Step 4: Verificar a integridade total do arquivo e testar compilação**

Executar a checagem no compilador Harpia para garantir 100% de conformidade técnica.
Run: `cd /Users/matheus.diniz_1/Documents/GitHub/Harpia/Harpia && go run . checar ../projetos/portal-harpia/web/rotas/index.hrp`

- [ ] **Step 5: Commit final da reestruturação do Portal Harpia**

```bash
git add projetos/portal-harpia/web/rotas/index.hrp
git commit -m "feat(portal): finalizar redesign completo da landing page do portal harpia"
```

---

## Execution Handoff

Plan complete and saved to `docs/superpowers/plans/2026-07-29-portal-harpia-redesign.md`.
