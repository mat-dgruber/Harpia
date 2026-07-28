# 🦅 Harpia

<p align="center">
  <img src="docs/assets/logo_harpia_refinado.jpg" alt="Logo Harpia" width="220" />
</p>

<p align="center">
  <strong>Linguagem de Programação Reativa, 100% Brasileira e Focada em Arquitetura Limpa</strong>
</p>

<p align="center">
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-00A86B.svg" alt="License" /></a>
  <a href="Manual.md"><img src="https://img.shields.io/badge/Manual_T%C3%A9cnico-Ver_Especifica%C3%A7%C3%A3o-007ACC.svg" alt="Manual Técnico" /></a>
  <a href="docs/BRAND_GUIDELINES.md"><img src="https://img.shields.io/badge/Diretrizes_de_Marca-Ver_Manual-F2A900.svg" alt="Brand Guidelines" /></a>
  <a href="ROADMAP.md"><img src="https://img.shields.io/badge/Desenvolvimento-Roadmap-D45D34.svg" alt="Roadmap" /></a>
</p>

---

## 📖 O que é a Harpia?

A **Harpia** é uma linguagem de programação brasileira moderna, focada em desenvolvimento ágil de ponta a ponta (Full Stack). Ela foi projetada para ir além do ensino de lógica de programação, permitindo a criação de sistemas profissionais de nível industrial — incluindo frontends SPA reativos, backends corporativos e APIs seguras — tudo utilizando a nossa língua nativa.

Representada pela imponente águia-real das Américas, a marca simboliza **soberania, precisão cirúrgica, força e foco absoluto**.

---

## 🚀 Instalação Rápida

Instale o compilador Harpia em segundos no seu sistema operacional executando o comando correspondente no terminal:

### Linux / macOS

```bash
curl -fsSL https://raw.githubusercontent.com/mat-dgruber/Harpia/main/Harpia/instalar.sh | bash
```

### Windows (PowerShell)

```powershell
irm https://raw.githubusercontent.com/mat-dgruber/Harpia/main/Harpia/instalar.ps1 | iex
```

---

## ⚡ Filosofia e Conceito: O Método Ponytail

A linguagem é orientada pela **Filosofia Ponytail** (preguiçosa com o código redundante, atenta com a leitura e pragmática com a solução):

1. **YAGNI (You Aren't Gonna Need It):** Eliminar qualquer código ou feature que não seja estritamente necessário.
2. **Modularização sem Cíclicos:** O compilador barra dependências cíclicas estaticamente antes da execução.
3. **Erros Didáticos e IA:** Mensagens estruturadas (`HRP-XXXX`) com dicas em português e suporte a explicações guiadas por IA local (`harpia erro explicar`).
4. **Legibilidade Semântica:** Uso do operador de canais (`|>`) e termos nativos do ecossistema brasileiro.
5. **Suíte Completa de Produção:** Suporte nativo a Segurança OWASP Top 10 (`seguranca`, `autenticacao`, `cripto.cifrar/decifrar`), manipulação de datas (`tempo`), e-mails (`email`), background jobs (`fila`), gerador PDF (`pdf`) e UI avançada (`usarTema`, `<Portal>`, `usarArrastar`, `usarNotificacao`).

---

## 🏗️ Clean Architecture e DDD Nativo

A Harpia foi desenhada sob a premissa de estruturar projetos robustos por padrão. Ao iniciar novos projetos corporativos, o compilador gera a árvore organizacional baseada em **DDD (Domain-Driven Design)** e **Clean Architecture**:

```text
meu-app/
├── dependencias.json        -> Manifesto de dependências e configurações
├── main.hrp                 -> Ponto de entrada (Bootstrapper)
├── dominio/                 -> Regras de Negócio Isoladas
│   ├── modelos/             -> Entidades e Objetos de Valor (ex: usuario.hrp)
│   └── servicos/            -> Validações e regras de domínio (ex: validador.hrp)
├── infra/                   -> Detalhes de Tecnologia (Banco de Dados, APIs)
│   ├── bd/                  -> Conexões e repositórios (ex: sqlite.hrp)
│   └── api/                 -> Clientes de requisição externa
├── web/                     -> Camada de Apresentação (Frontend SPA)
│   ├── rotas/               -> Páginas com File-system Routing (ex: index.hrp)
│   ├── componentes/         -> Componentes UI reutilizáveis (ex: botao.hrp)
│   └── estilos/             -> Folhas de estilo locais ou globais
└── testes/                  -> Camada de Testes Automatizados
```

---

## ✨ Características Principais & Performance

- **Direct-Threaded JIT VM:** Bytecodes dinamicamente traduzidos em chamadas Go nativas em tempo de execução, otimizando o loop de decodificação.
- **Pool de Alocação Eden:** Pré-boxeamento de inteiros curtos (de `-100` a `2000`) em $O(1)$ para aniquilar pressões desnecessárias do Garbage Collector.
- **Reatividade Nativa (SPA):** Transpilação reativa eficiente para a web (`--alvo=web`) baseada em Sinais (`var [contador, definirContador] = sinal(0)`), Efeitos e Estado Global.
- **Estilização Nativa:** Blocos de estilo CSS integrados nativamente e classes utilitárias na estrutura de marcação.
- **Contrato RPC Automático:** Comunicação simplificada entre o Front-end e o Back-end sem a necessidade de APIs manuais complexas.

---

## 🛠️ Caixa de Ferramentas (CLI)

A ferramenta de linha de comando (`harpia`) oferece utilitários robustos de ponta a ponta para gerenciar, compilar, testar e estruturar seus projetos. Abaixo estão os 27 comandos disponíveis com suas especificações e flags:

### 1. `harpia atualize`

Atualiza a CLI do Harpia consultando o GitHub Releases API, comparando a tag local com a remota e substituindo o binário em `~/.harpia/bin/harpia`.

- **Uso:** `harpia atualize`
- **Observação:** Comando totalmente Zero-Config — sem argumentos nem flags.

### 2. `harpia executar` (ou `harpia exec`)

Executa um script físico Harpia ou inicia o console de desenvolvimento interativo (REPL).

- **Uso:** `harpia executar [caminho-do-arquivo.hrp] [flags]`
- **Flags Principais:**
  - `-c, --codigo`: Executa um trecho de código diretamente no terminal (ex: `harpia executar -c "imprimir('Olá!')"`).
  - `--assistir`: Modo Watch Mode (Hot Reload) — recarrega e reexecuta o script automaticamente ao detectar alterações no disco.
  - `--estrito`: Ativa a validação estrita de anotações de tipo em tempo de execução.
- **Necessidades:** Sem argumentos, inicia o REPL com ajuda e inspetor de memória embutidos; com argumento, executa o arquivo `.hrp` imediatamente.

### 3. `harpia testar`

Varre e isola testes unitários lógicos declarados na cláusula sintática `testar "nome" { ... }` nativa da linguagem, fornecendo um relatório consolidado de acertos e falhas.

- **Uso:** `harpia testar [caminho_ou_pasta]`

### 4. `harpia checar`

Linter semântico offline-friendly que realiza análise estática de sintaxe e semântica no arquivo especificado sem executá-lo.

- **Uso:** `harpia checar [caminho-do-arquivo.hrp]`
- **Saída:** Diagnósticos detalhados de variáveis não declaradas, reatribuição de constantes, assinaturas de funções incorretas ou erros de sintaxe.

### 5. `harpia erro [codigo]`

Fornece explicações didáticas em português sobre os códigos de erros do Harpia.

- **Uso:** `harpia erro [PSC-xxxx]`
- **Catálogo:** 14 códigos de erro (`PSC-0001` a `PSC-0014`) cobrindo Sintaxe, Reatribuição, Atributo, Tipagem, Nome, Importação, Valor, Limite, Índice, Runtime, Asseguração, DivisãoPorZero, Sistema e ArquivoNãoEncontrado.
- **Subcomando:** `harpia erro explicar [codigo]` — usa IA local (Ollama, modelo `gemma`) para gerar explicação pedagógica personalizada; fallback automático para o catálogo estático se Ollama estiver offline.

### 6. `harpia compilar`

Transpila o seu código fonte Harpia para outras plataformas (como a Web com suporte a Virtual DOM e JavaScript, ou executáveis nativos via AOT).

- **Uso:** `harpia compilar --alvo=web --entrada=main.hrp --saida=dist`
- **Flags Principais:**
  - `-a, --alvo`: Alvo da compilação. Opções: `web` (padrão), `nativo`, `wasm`.
  - `-e, --entrada`: Ponto de entrada/arquivo principal do projeto (ex: `main.hrp`).
  - `-s, --saida`: Pasta destino onde serão gravados os arquivos transpilados/compilados (padrão: `dist`).

### 7. `harpia servir`

Inicia o Dev Server de desenvolvimento integrado com compilação sob demanda para hospedar os arquivos compilados da sua aplicação SPA reativa e visualizá-la no navegador.

- **Uso:** `harpia servir --diretorio=dist --porta=8080`
- **Flags Principais:**
  - `-d, --diretorio`: Pasta que contém os arquivos que serão servidos (padrão: `dist`).
  - `-p, --porta`: Porta na qual o servidor web será escutado (padrão: `8080`).

### 8. `harpia novo` (ou `iniciar`, `inicializar`)

Inicializa uma nova estrutura de projeto baseada em Clean Architecture e DDD com termos em português.

- **Uso:** `harpia novo [backend | frontend | monolito] [nome-do-projeto]`
- **Subcomandos:**
  - `backend`: Cria uma estrutura enxuta de backend focada em APIs lógicas, conectores de banco de dados e concorrência leve.
  - `frontend`: Cria uma estrutura reativa cliente puramente SPA de alto desempenho baseada em Sinais e Virtual DOM.
  - `monolito`: Cria um novo monolito completo de Frontend + Backend com pastas explicativas e documentação interna de arquitetura.

### 9. `harpia crie`

Assistente de geração dinâmica de novos arquivos de templates seguindo a Clean Architecture dentro de um projeto existente.

- **Uso:** `harpia crie [rota | componente | modelo] [nome]`
- **Subcomandos:**
  - `rota`: Cria um novo arquivo de rota SPA (`.hrp`) na pasta correspondente.
  - `componente`: Cria um componente de interface (`.hrp`) e seu respectivo arquivo de estilos dinâmicos (`.estilo.hrp`).
  - `modelo`: Cria um novo modelo/entidade de dados rico e tipado na camada de domínio.

### 10. `harpia lsp`

Inicia o servidor oficial LSP (Language Server Protocol) do Harpia via stdio, oferecendo suporte nativo para editores de código (como o VS Code) com autocomplete, hover lendo comentários de três barras (`///`), linter de arquitetura limpa e formatação automática de código ao salvar.

- **Uso:** `harpia lsp`

### 11. `harpia playground`

Inicia o servidor web local do Playground Interativo do Harpia para experimentação de código no navegador.

- **Uso:** `harpia playground`
- **Flags:**
  - `-p, --porta`: Porta de escuta do servidor do playground (padrão: `8090`).

### 12. `harpia formatar`

Formata a indentação e o estilo visual de arquivos Harpia, aplicando indentação canônica de 4 espaços, removendo trailing whitespace e higienizando quebras de linha.

- **Uso:** `harpia formatar [arquivo.hrp | diretorio]` (mínimo 1 argumento)
- **Flags:**
  - `-w, --escrever`: Salva as alterações in-place no arquivo original.
  - `--verificar`: Modo CI — retorna exit≠0 se houver drift de formatação (mensagem `HRP-FMT-001` por arquivo).

### 13. `harpia instalar` (ou `instale`, `install`)

Instala dependências do projeto registradas no `pacote.hrp` ou `pacote.json`, buscando metadados no índice de pacotes centralizado em `./pt_modulos/`.

- **Uso:** `harpia instalar [pacote-opcional] [url-ou-versao-opcional]`
- **Flags:**
  - `-f, --arquivo`: Caminho do arquivo de manifesto personalizado.

### 14. `harpia diagramar`

Mapeia as relações de importações e valida regras da Clean Architecture, gerando diagramas em grafo Mermaid.

- **Uso:** `harpia diagramar [diretorio]` (opcional, default `.`)
- **Flags:**
  - `-f, --formato`: Formato de saída do diagrama: `mermaid` (padrão), `html`, `svg`.
  - `-s, --saida`: Caminho do arquivo para salvar a saída.

### 15. `harpia tui`

Inicia a TUI Interativa e REPL didático do Harpia no terminal (baseado em Bubbletea/Lipgloss).

- **Uso:** `harpia tui`

### 16. `harpia doc`

Gera documentação automática a partir de comentários especiais `///` precedendo declarações de `funcao`, `classe` (com `estende`), `constante` e `var`.

- **Uso:** `harpia doc [arquivo.hrp]` (1 argumento obrigatório)
- **Flags:**
  - `-f, --formato`: Formato de saída: `markdown` (padrão) ou `html`.
  - `-s, --saida`: Caminho do arquivo de saída (vazio = stdout).
- **Distinto de `harpia docs`** (OpenAPI/Swagger de rotas HTTP).

### 17. `harpia empacotar`

Empacota um script Harpia em um executável nativo autônomo (Single Binary Bundle) ou WASM.

- **Uso:** `harpia empacotar`
- **Flags:**
  - `-e, --entrada`: Arquivo Harpia de entrada.
  - `-s, --saida`: Caminho de saída do binário gerado.
  - `--so`: Sistema Operacional alvo (padrão: SO atual; opções: `linux`, `windows`, `darwin`, `js`).
  - `--arq`: Arquitetura alvo (padrão: arquitetura atual; opções: `amd64`, `arm64`, `wasm`).

### 18. `harpia stressar`

Executa testes de estresse concorrentes em um script Harpia, medindo throughput e latência.

- **Uso:** `harpia stressar`
- **Flags:**
  - `-a, --arquivo`: Caminho do arquivo Harpia.
  - `-c, --concorrencia`: Quantidade de instâncias concorrentes (padrão: `10`).
  - `-r, --requisicoes`: Número total de execuções de teste (padrão: `100`).

### 19. `harpia depurar`

Inicia o servidor de depuração DAP (Debug Adapter Protocol) para conexão com IDEs e editores.

- **Uso:** `harpia depurar`
- **Flags:**
  - `-p, --porta`: Porta TCP para o servidor DAP (padrão: `4711`).

### 20. `harpia copiloto [arquivo]`

Sugere complementação de código inteligente via modelo de IA local (Ollama).

- **Uso:** `harpia copiloto [arquivo]`
- **Flags:**
  - `-m, --modelo`: Modelo do Ollama a ser utilizado (padrão: `llama3`).
  - `-u, --url`: URL da API do Ollama local (padrão: `http://localhost:11434/api/generate`).
- **Subcomandos:**
  - `copiloto revisar <arquivo>`: Análise estática — detecta funções longas, muitos parâmetros, aninhamento excessivo, variáveis não usadas e TODO/FIXME.
  - `copiloto refatorar <arquivo>`: Sugere extração de helpers a partir de funções/métodos com mais de 80 linhas.

### 21. `harpia gerar`

Gera dados fake brasileiros (pessoas, empresas, endereços) para testes.

- **Uso:** `harpia gerar <pessoa | empresa | endereco>`
- **Flags:**
  - `-q, --quantidade`: Quantos registros gerar (padrão: `1`).
  - `-f, --formato`: Formato de saída: `json` (padrão) ou `texto` (chave: valor separado por `---`).
- **Subcomandos:**
  - `pessoa`: Gera uma pessoa física brasileira completa (nome, CPF, email, endereço).
  - `empresa`: Gera uma empresa brasileira (razão social, CNPJ simulado).
  - `endereco`: Gera um endereço brasileiro (logradouro, cidade, UF, CEP).

### 22. `harpia migrar`

Gerencia migrations SQL com SQLite.

- **Uso:** `harpia migrar <criar | aplicar | status | reverter>`
- **Flags Globais:**
  - `--banco`: Banco SQLite alvo (padrão: `dados.db`).
- **Subcomandos:**
  - `criar <nome>`: Cria arquivo de migration com timestamp.
  - `aplicar`: Aplica todas as migrations pendentes em ordem.
  - `status`: Lista migrations aplicadas e pendentes.
  - `reverter [n]`: Reverte as últimas N migrations aplicadas (default 1).

### 23. `harpia pwa`

Gera `manifest.webmanifest` e `sw.js` (PWA) a partir de `dist/`.

- **Uso:** `harpia pwa`
- **Flags:**
  - `--dir`: Diretório alvo (padrão: `dist`).
  - `--nome`: Nome do app (lido de `dependencias.json` se vazio).
  - `--curto`: Nome curto (lido de `dependencias.json` se vazio).
  - `--cor-fundo`: Cor de fundo (padrão: `#ffffff`).
  - `--cor-tema`: Cor de tema (padrão: `#3b82f6`).
  - `--registrar`: Injeta registro do service worker em `index.html`.

### 24. `harpia i18n`

Extrai e gerencia catálogos de tradução (`.pot`/`.po`).

- **Uso:** `harpia i18n <extrair | novo>`
- **Flags Globais:**
  - `--dir`: Diretório de catálogos (padrão: `traducoes`).
  - `--dominio`: Domínio gettext (padrão: `harpia`).
- **Subcomandos:**
  - `extrair <arquivo|dir>`: Extrai strings traduzíveis e gera `<dir>/<dominio>.pot`.
  - `novo <idioma>`: Cria catálogo `.po` vazio para o idioma (cabeçalho copiado do `.pot`).

### 25. `harpia docs` (ou `swagger`, `openapi`)

Analisa a AST do servidor HTTP e gera automaticamente a especificação OpenAPI 3.0.0 (`swagger.json`) para documentação de APIs REST.

- **Uso:** `harpia docs [arquivo.hrp]`

### 26. `harpia auditar` (ou `audit`, `seguranca`)

Executa análise estática de segurança varrendo a AST por segredos expostos, concatenações de SQL brutas e falhas OWASP Top 10.

- **Uso:** `harpia auditar [arquivo_ou_pasta]`

### 27. `harpia publicar` (ou `deploy`, `publish`)

Gera Dockerfiles otimizados e artefatos compilados de produção para deploy em nuvem ou contêineres em 1 clique.

- **Uso:** `harpia publicar [alvo]`

### 28. `harpia completar` (ou `autocompletar`, `completion`)

Gera o script de autocompletar para o shell informado.

- **Uso:** `harpia completar [bash|zsh|fish|powershell]`
- **Exemplos:**
  - `harpia completar bash > ~/.local/share/bash-completion/completions/harpia`
  - `source <(harpia completar bash)`

---

## 📚 Documentação & Guias

Para mais informações sobre as regras de desenvolvimento do projeto:

- 📖 **[Manual de Referência Técnico Oficial](Manual.md)**
- 🎨 **[Diretrizes de Marca & Identidade Visual](docs/BRAND_GUIDELINES.md)**
- 💡 **[Exemplos Práticos de Aplicação e Código](docs/harpia-brand-examples.md)**
- 🚀 **[Guia de Contribuição](CONTRIBUTING.md)**
- 🗺️ **[Roadmap de Evolução](ROADMAP.md)**

---

<p align="center">
  <img src="docs/assets/fundo_harpia.jpg" alt="Wallpaper Harpia" width="100%" />
</p>

<p align="center">
  Feito com ❤️ pela Comunidade Brasileira de Programação 🇧🇷
</p>
