# 🇧🇷 Portuscript — Roadmap Oficial

> Objetivo: transformar o Portuscript em uma linguagem de programação real, completa e em português — capaz de rodar no backend, no frontend e com sua própria VM.

## 💡 Filosofia de Design

O Portuscript foi concebido sob uma perspectiva dupla:

1. **Ponte de Aprendizado:** Facilitar a transição de novos desenvolvedores para linguagens consolidadas no mercado (como JavaScript, Python, Go e C#) utilizando sintaxes e conceitos familiares (escopo, blocos com chaves, OOP tradicional e corotinas).
2. **Poder e Identidade Própria:** Ser uma ferramenta altamente produtiva e usável por si só, oferecendo recursos únicos integrados diretamente à linguagem, como reatividade nativa via **sinais**, suporte a **componentes JSX-like embutidos**, e um compilador focado em **ensino e diagnósticos ricos em português**.

---

## 🧭 As 10 Decisões Fundamentais

| #   | Tema                  | Decisão                                                                                |
| --- | --------------------- | -------------------------------------------------------------------------------------- |
| 1   | **Motor**             | Tree-walk agora → VM de pilha + bytecode depois (quando o núcleo estiver estável)      |
| 2   | **Modelo de valores** | NaN-boxing híbrido: números/booleans inline, ponteiros para o resto                    |
| 3   | **Componentes**       | Funções + closures com sintaxe JSX-like (`<h1>...</h1>` embutido) — semântica primeiro |
| 4   | **Reatividade**       | Sinais agora (estilo Solid.js), decorators/classes depois da VM                        |
| 5   | **Estilo**            | B com pitada de C — moderna, familiar, chaves`{}` obrigatórias, sem `()` em condições  |
| 6   | **OOP**               | Classes com herança simples (estilo Python)                                            |
| 7   | **Tipos**             | Tipagem opcional — sem tipos funciona, com tipos o compilador valida                   |
| 8   | **GC**                | Contagem de referências (estilo CPython) — previsível e debugável                      |
| 9   | **Concorrência**      | Corotinas leves com`assincrono`/`aguarde` (estilo Python async/await)                  |
| 10  | **DX**                | Erros visuais em PT + sugestões (estilo Rust/Elm) + LSP desde cedo                     |
| 11  | **Ferramentas**       | Tradutor nativo (`portuscript traduzir`) e guia interativo de erros integrado no CLI   |

---

## 📦 Estado Atual (baseline v0.x)

- ✅ Lexer funcional (tokens em Go)
- ✅ Parser com AST completa (variáveis, funções, loops, condicionais, mapas, tuplas, listas, importação)
- ✅ Tree-walk interpreter funcional
- ✅ Stdlib inicial: `embutidos`, `matematica`, `sistema`, `soquete`, `colorize`
- ✅ Exemplos: aritmética, condicionais, loops, funções, soquetes, importação
- ⚠️ Classes marcadas como "sendo estudadas" — não implementadas
- ⚠️ Tipagem no parser não é validada em runtime
- ⚠️ Constantes declaráveis mas sem suporte completo
- ❌ GC próprio (usa GC do Go)
- ❌ VM / bytecode
- ❌ Sinais / reatividade
- ❌ LSP
- ❌ Concorrência / corotinas

---

## 🏗️ Arquitetura de Execução e Contextos

O Portuscript identifica onde e como está rodando por meio de **Alvos de Compilação (Targets)** informados no CLI ou deduzidos pelo ambiente:

### 1. Alvos Principais (CLI Flags)

- `portuscript exec`: Execução em tempo de desenvolvimento na **VM de pilha local** (ambiente nativo). Acesso total à stdlib local.
- `portuscript compilar --alvo=backend` (Padrão): Empacota o bytecode com a VM nativa Go gerando um **binário estático autônomo** para servidores.
- `portuscript compilar --alvo=web`: Transpila o código para JavaScript (ou WebAssembly) para rodar no **Navegador (Frontend)**. Desativa automaticamente módulos do sistema operacional (como `sistema` e `bd`).

### 2. Resolução de Módulos (Stdlib Dinâmica)

- **Módulos Restritos:** O compilador impede a importação de pacotes específicos do sistema (ex: `importar de "bd/sqlite"`) quando compila para o alvo `web`, gerando um erro de build em português explicando a restrição de segurança do navegador.
- **Abstração Dinâmica:** Módulos comuns de rede e dados (como `http` e `json`) se adaptam. `http.obter()` compila para as chamadas de rede nativas de Go no backend e para a API `fetch()` no navegador.

---

> **Meta:** Portuscript como linguagem completa de uso geral, com classes, tipos e erros humanos.

- **Por que fazer:** O interpretador de árvore (tree-walk) atual é ideal para prototipar a sintaxe básica sem a complexidade de compilar código, mas carece de estruturas necessárias para aplicações de verdade (como orientação a objetos estruturada, constantes reais e erros amigáveis).
- **Como fazer:** Continuaremos no motor em árvore atual. Vamos estender o parser para validar as constantes no tempo de execução, implementar o interpretador de classes com herança criando tabelas de métodos dinâmicos (vtable simplificada no runtime em Go) e criar um módulo centralizado de formatação de erros em português.

#### 1.1 — Sistema de tipos opcional

- [ ] Tipos primitivos: `Inteiro`, `Decimal`, `Texto`, `Booleano`, `Nulo`
- [ ] Tipos compostos: `Lista<T>`, `Mapa<C, V>`, `Tupla`
- [ ] Tipo `funcao` com assinatura tipada
- [ ] Verificação opcional em tempo de parse/execução (flag `--estrito`)

#### 1.2 — Classes com herança simples

- [ ] Sintaxe: `classe Animal { ... }`
- [ ] Construtor: `inicializar(self, ...)`
- [ ] Herança: `classe Cachorro estende Animal { ... }`
- [ ] Métodos de instância com `self`
- [ ] Métodos estáticos com `estatico`
- [ ] `instancia de` como operador nativo
- [ ] Conectar `NovaNode` (já no parser) ao runtime

#### 1.3 — Constantes e Módulos Unificados (Imports/Exports)

- [ ] `constante` com valor obrigatório e imutável em runtime
- [ ] Sintaxe unificada de importações/exportações para frontend/backend usando `importar { ... } de ...` e `exportar`.
- [ ] Sistema de análise estática de grafo de dependências para detecção e prevenção de dependências cíclicas (importações em loop) com erro amigável em tempo de compilação.
- [ ] Escopo léxico correto para closures aninhadas e shadowing controlado com aviso.

#### 1.4 — Erros visuais em português e guia de ajuda

- [ ] Struct `Erro` com: `mensagem`, `linha`, `coluna`, `trecho`, `sugestao`, `codigoErro`
- [ ] Output com sublinhado do trecho errado (estilo Elm)
- [ ] Sugestões contextuais (ex: "Você quis dizer `retorne`?")
- [ ] **Tratamento de Exceções em Runtime (`tente / capture / finalmente`):** Implementar o fluxo de tratamento de erros no interpretador para capturar erros em tempo de execução, permitindo que blocos `capture` tratem e previnam falhas do programa, com execução garantida do bloco `finalmente`.
- [ ] Sistema de ajuda interativa no CLI: `portuscript erro [codigo]` para obter explicação educativa
- [ ] **Integração com IA Local (Opcional):** Comando `portuscript erro explicar` que lê o contexto do último erro ocorrido e envia para um modelo de LLM local (via Ollama/Llama.cpp com modelo leve) para explicar de forma didática e em português como resolver o problema.
- [ ] Erros em PT em todo o lexer, parser e runtime

#### 1.5 — Parâmetros avançados de funções

- [ ] Valores padrão: `funcao soma(a, b = 0) { ... }`
- [ ] Parâmetros nomeados: `soma(a = 1, b = 2)`
- [ ] Tipagem opcional nos parâmetros

#### 1.6 — Testes nativos na linguagem

- [ ] Palavra-chave nativa `testar "nome do teste" { ... }`
- [ ] Integração do comando `portuscript testar [caminho]` para rodar todos os blocos de teste e mostrar relatório visual (passou/falhou) no terminal
- [ ] Função global `assegura(condicao)` (ou `assegure`) para validação de asserções nos testes

#### 1.7 — Operador de Canal (Pipes)

- [ ] Implementação do operador de canal `|>` (pipe operator) para fluxo de dados contínuo (Ex: `texto |> removerEspacos |> maiusculo`).
- [ ] Integração do operador na sintaxe de expressões de componentes (Ex: `<h1>{usuario.nome |> maiusculo}</h1>`).

#### 1.8 — Validador Semântico Estático (Linting e Segurança)

- [ ] Comando `portuscript checar` para varredura estática da AST sem executar o código.
- [ ] Validação prévia de escopos de variáveis, shadowing proibido, caminhos de importação incorretos e violações de regras de tipo estritas (quando tipagem estiver ativada).
- [ ] Relatórios de linting detalhados em português integrados ao LSP.

---

### 🟣 FASE 2 — VM de Pilha + Bytecode

> **Meta:** Compilar Portuscript para bytecode próprio (`.ptc`) e executar numa VM de pilha eficiente em Go.

- **Por que fazer:** O interpretador tree-walk é lento e gasta muita memória para projetos grandes porque ele visita a árvore sintática a cada execução de instrução ou loop. A VM de pilha e o bytecode garantem desempenho em nível de produção. O NaN-boxing permite representar qualquer tipo em um inteiro de 64 bits, reduzindo drasticamente a alocação de memória no heap.
- **Como fazer:** Escreveremos um compilador em Go que transforma a AST em um array plano de opcodes. A VM terá uma pilha de execução (`stack`) e lerá os opcodes sequencialmente. O Garbage Collector de contagem de referências incrementará as referências dos objetos no heap ao empilhar/armazenar e as reduzirá ao desempilhar ou sair do frame de função, liberando o objeto imediatamente quando chegar a zero.

#### 2.1 — Conjunto de instruções (ISA)

- [ ] Pilha: `PUSH`, `POP`, `DUP`, `SWAP`
- [ ] Aritmética: `ADD`, `SUB`, `MUL`, `DIV`, `MOD`
- [ ] Comparação: `EQ`, `NEQ`, `LT`, `GT`, `LTE`, `GTE`
- [ ] Controle: `JMP`, `JMP_SE_FALSO`, `RETORNE`
- [ ] Variáveis: `CARREGAR_LOCAL`, `ARMAZENAR_LOCAL`, `CARREGAR_GLOBAL`
- [ ] Funções: `CHAMAR`, `RETORNE`, `FECHAR` (closures)
- [ ] Objetos: `CRIAR_OBJETO`, `ACESSAR_MEMBRO`, `DEFINIR_MEMBRO`

#### 2.2 — Compilador AST → bytecode

- [ ] Visitor sobre a AST que emite instruções
- [ ] Pool de constantes internadas
- [ ] Formato `.ptc` (cabeçalho + versão + instruções)

#### 2.3 — VM de pilha em Go

- [ ] Frame de execução por chamada de função
- [ ] Loop `fetch → decode → execute`
- [ ] Pilha de operandos e variáveis locais (vetor indexado)

#### 2.4 — Modelo NaN-boxing

- [ ] `Valor` como `uint64` com bits de tag
- [ ] Tags: `NULO`, `BOOLEANO`, `INTEIRO`, `DECIMAL`, `PONTEIRO`
- [ ] Benchmarks vs. tree-walk

#### 2.5 — GC por contagem de referências

- [ ] `ObjetoGC` com campo `refs int`
- [ ] `Reter()` / `Liberar()` nas atribuições
- [ ] Detecção e quebra de ciclos simples
- [ ] Integração com a VM

---

### 🟢 FASE 3 — Stdlib Robusta (Backend Real)

> **Meta:** Portuscript como linguagem de backend de verdade.

- **Por que fazer:** Uma linguagem só é útil no mundo real se puder interagir com o sistema de arquivos, redes e bancos de dados. Concorrência leve via corotinas garante que a linguagem possa lidar com milhares de conexões simultâneas de I/O sem travar o thread principal.
- **Como fazer:** Mapear APIs do ecossistema do Go para módulos internos do Portuscript. As corotinas serão suspensas quando uma operação assíncrona for iniciada, registrando um callback no Event Loop interno em Go, e serão retomadas (com o estado da pilha restaurado na VM) assim que o Go sinalizar que a leitura/gravação foi concluída.

#### 3.1 — HTTP Servidor Completo (Middlewares & Injeção)

- [ ] Módulo `importar { Servidor } de "http"` com suporte a rotas e verbos HTTP (`obter()`, `postar()`, `deletar()`, `atualizar()`).
- [ ] Roteador avançado com parâmetros dinâmicos de URL (Ex: `/api/usuarios/:id` acessível via `req.parametros.id`).
- [ ] Pipeline de execução baseado em **Middlewares** (funções que interceptam e processam requisições com chamada de callback `proximo()`).
- [ ] Sistema nativo e leve de **Injeção de Dependências** no servidor HTTP para gerenciamento de ciclo de vida de classes e serviços do domínio.
- [ ] Cliente HTTP integrado com suporte a HTTPS e certificados.

#### 3.2 — Soquetes aprimorados

- [ ] API assíncrona com `aguarde`
- [ ] WebSocket nativos
- [ ] UDP além do TCP existente

#### 3.3 — Banco de Dados e Query Builder Nativo

- [ ] Interface de conexão unificada `Conexao` com métodos `consultar(sql, params)`, `executar(sql, params)` e `transacao(callback)`.
- [ ] Implementação de drivers embutidos: SQLite (nativo/embarcado) and PostgreSQL.
- [ ] Mapeamento automático de tabelas SQL para tipos de dados nativos (Listas e Mapas do Portuscript).
- [ ] Query Builder dinâmico integrado (Ex: `bd.tabela("usuarios").onde("idade", ">", 18).obterMuitos()`).
- [ ] **Suporte a Banco de Dados NoSQL:** Drivers integrados para MongoDB (repositório de documentos) e Redis (chave-valor/cache) com interface simplificada (Ex: `importar { Conectar } de "bd/mongodb"`, `importar { Cache } de "bd/redis"`).
- [ ] Gerenciamento de pool de conexões robusto e concorrência segura.

#### 3.4 — Sistema de arquivos aprimorado

- [ ] `ler()`, `escrever()`, `acrescentar()`, `remover()`, `renomear()`
- [ ] `caminhar(diretorio)` recursivo
- [ ] `Caminho.juntar()`, `Caminho.resolver()`

#### 3.5 — JSON/YAML/XML nativos

- [ ] `importar { analisar, serializar } de "json"`
- [ ] Tipos Portuscript ↔ JSON
- [ ] YAML e XML como módulos opcionais

#### 3.6 — Criptografia básica

- [ ] Hash: `sha256()`, `sha512()`, `md5()`
- [ ] HMAC, Base64, UUID
- [ ] `importar { criptografia } de "cripto"`

#### 3.7 — Concorrência com corotinas

- [ ] Palavras-chave: `assincrono funcao`, `aguarde`
- [ ] Event loop integrado na VM
- [ ] `Promessa` nativa
- [ ] `aguarde Promessa.todos([p1, p2, p3])`

#### 3.8 — Integração Decoupled e Clientes de API Autogerados (RPC)

- [ ] Mecanismo de leitura de contratos: O compilador lê as funções exportadas nas rotas do backend e gera assinaturas de chamada estáticas.
- [ ] Vinculação via `dependencias.json` permitindo que projetos de frontend consumam serviços de backend via importação direta (Ex: `importar { obterDados } de "@backend/dados"`), eliminando a necessidade de escrever requisições HTTP manuais e URLs absolutas.
- [ ] Geração dinâmica de clientes de API integrados na compilação do WebAssembly.

---

### 🟠 FASE 4 — Frontend + Sinais + JSX-like (SPA Completo)

> **Meta:** Portuscript rodando no browser, com reatividade, componentes, estilização embutida e roteamento de SPA.

- **Por que fazer:** Habilitar o desenvolvimento de interfaces reativas e dinâmicas com a mesma linguagem usada no backend, proporcionando uma experiência de desenvolvimento simples de ponta a ponta sem precisar configurar compiladores externos complexos.
- **Como fazer:** Criaremos um compilador alternativo de Portuscript para WebAssembly que traduz a sintaxe de sinais e componentes JSX-like para chamadas equivalentes a uma pequena biblioteca de Virtual DOM inclusiva, mapeando eventos HTML diretamente para funções reativas.

#### 4.1 — Compilação para WebAssembly (Substituição de HTML/JS)

- [ ] Compilador nativo: `portuscript compilar --alvo=web` gerando binário WASM autônomo
- [ ] Motor de montagem de página nativo a partir do ponto de entrada do arquivo Portuscript
- [ ] Abstração completa de manipulação do DOM pelo runtime WASM do Portuscript
- [ ] **Renderização no Servidor (SSR) e Hidratação:** O servidor de backend renderiza HTML estático inicial instantâneo, e o WebAssembly no navegador realiza a hidratação ligando os fios da reatividade sem recriar os elementos.

#### 4.2 — Estilização Nativa e Unificada (Três Pilares)

- [ ] **1. Bloco de Estilo Declarativo (`estilo`):** Palavra-chave nativa para blocos de estilização estruturada com suporte a seletores e pseudo-classes (Ex: `estilo MeuComponente { corDeFundo: "azul"; botao:hover { opacidade: 0.8; } }`).
- [ ] **2. Objetos de Estilo Reativos:** Mapas dinâmicos baseados no estado da aplicação (sinais) aplicados à propriedade `estilo` (Ex: `var estiloDinamico = { cor: ativo() ? "verde" : "vermelho" }`).
- [ ] **3. Classes Utilitárias Nativas ("Tailwind" em PT):** Utilitários embutidos de layout e cor baseados em strings brasileiras curtas na propriedade `classe` (Ex: `<div classe="flex-linha itens-centro p-4 cor-azul-500">...</div>`).
- [ ] Mapeamento e compilação otimizada dessas regras para arquivos CSS estáticos gerados em tempo de compilação do WebAssembly.

#### 4.3 — Roteamento SPA Baseado em Arquivos (File-system Routing)

- [ ] Detecção automática do diretório `/rotas` do projeto pelo compilador
- [ ] Criação de rotas automáticas baseadas em arquivos (Ex: `/rotas/sobre.ptst` vira rota `/sobre`)
- [ ] Navegação dinâmica via componente nativo `<Link para="/sobre">` sem recarregar a página

#### 4.4 — Metadados Semânticos Nativos (AEO & GEO)

- [ ] Objeto de configuração `metadados` exportável em rotas com suporte nativo a Schema.org (`esquema`) e OpenGraph.
- [ ] Geração automática de marcações JSON-LD estruturadas pelo servidor (SSR) para indexação eficiente em buscas por inteligência artificial (AEO) e serviços baseados em geolocalização (GEO).
- [ ] Mapeamento e renderização dinâmica de meta tags dinâmicas no cabeçalho da página durante a navegação do SPA.

#### 4.5 — Sinais e Estado Global Nativo

- [ ] `sinal(valorInicial)` → retorna `[ler, definir]`
- [ ] `efeito(funcao)` → re-executa quando sinais mudam
- [ ] `derivado(funcao)` → sinal computado (memoizado)
- [ ] **Gerenciador de Estado Global (`armazem`):** Primitiva para sincronização de dados globais entre múltiplos componentes (Ex: `var carrinho = armazem({ total: 0 })`).

#### 4.5 — Componentes JSX-like

- [ ] Sintaxe: `<div classe="app">...</div>` embutida
- [ ] Componentes como funções
- [ ] Props como parâmetros nomeados
- [ ] Eventos: `aoClicar={minhaFuncao}`
- [ ] Renderização condicional: `<se condicao>...</se>`
- [ ] Listas: `<para item em lista>...</para>`

#### 4.6 — Virtual DOM / Reconciliação

- [ ] Árvore virtual de nós
- [ ] Diff eficiente entre renders
- [ ] Atualização cirúrgica do DOM

---

### ⚡ FASE 5 — Tooling & Ecossistema

> **Meta:** Portuscript com ferramentas de classe mundial.

#### 5.1 — CLI e Scaffolding de Projetos (Comandos em PT)

- [ ] Comandos de criação de projeto direcionados:
  - `portuscript novo-monolito [nome]`: Inicializa a estrutura Clean Architecture e DDD completa (`/dominio`, `/infra`, `/web`, `/testes`).
  - `portuscript novo-backend [nome]`: Inicializa apenas a estrutura lógica de serviços e persistência (`/dominio`, `/infra`, `/testes`).
  - `portuscript novo-frontend [nome]`: Inicializa apenas a estrutura cliente reativa de páginas e estilos (`/web`, `/testes`).
- [ ] Comando `portuscript crie rota [nome]` / `portuscript crie componente [nome]`: Gerador de código assistido (scaffolding) que cria arquivos boilerplate estruturados prontos para uso no diretório correspondente da camada `/web`.
- [ ] Comando `portuscript servir`: Inicializa o servidor de desenvolvimento local com Hot-Reload (atualização instantânea no navegador ao salvar arquivos).

#### 5.2 — LSP (Language Server Protocol)

- [ ] `portuscript lsp` — servidor LSP em Go
- [ ] Autocompletar, diagnósticos, hover, go-to-definition
- [ ] Extensão VSCode oficial

#### 5.3 — Playground Interativo e Depurador Visual Local

- [ ] Comando `portuscript playground` que inicializa servidor web local
- [ ] Editor e interpretador web integrado rodando localmente
- [ ] Painel visual demonstrando passo a passo a pilha de execução (stack trace) e variáveis do runtime de forma didática
- [ ] **Formatação HTML de Erros:** Suporte nativo para exportação e renderização de erros ricos estruturados em HTML com tags de destaque, sublinhado e sugestões estilizadas para exibição visual rica na interface web do playground.

#### 5.4 — Formatador

- [ ] `portuscript formatar arquivo.ptst`
- [ ] Integração via LSP

#### 5.5 — Gerenciador de pacotes

- [ ] `ptst instalar nome-do-pacote`
- [ ] Registro central em PT
- [ ] `pacote.ptst` com dependências e semver

#### 5.6 — Console Interativo de Terminal (TUI Didática)

- [ ] Implementação de Interface de Usuário de Terminal (TUI) rica baseada no ecossistema Bubbletea (Go) ao executar `portuscript` sem argumentos.
- [ ] **Painéis Divididos:** Tela interativa contendo:
  - _Console/Editor de Código:_ REPL interativo com realce de sintaxe e autocompletar.
  - _Inspetor de Memória/Pilha:_ Exibição em tempo real do estado da VM, variáveis declaradas e stack trace didático.
  - _Painel de Saída:_ Exibição de logs de execução e erros estruturados com dicas.
- [ ] Atalhos interativos e atalhos rápidos integrados (Ex: Ajuda IA com F1, Executar com F2).

#### 5.7 — Documentação Assistida e Portal Oficial

- [ ] Documentação interna detalhada (`/docs`) contendo especificações técnicas completas de todos os comandos do CLI, comportamento da Stdlib, estruturas de imports/exports e sintaxes da linguagem, sempre detalhando o **Como** e o **Porquê** de cada design.
- [ ] Extração automática de documentação a partir de blocos de comentários especiais com três barras (`///`).
- [ ] Comando `portuscript doc arquivo.ptst` gerando documentação interativa rica em HTML ou Markdown.
- [ ] **Portal Oficial da Linguagem:** Site web estático contendo documentação estruturada amigável, guias de migração rápida para outras linguagens, e um **Playground interativo online** rodando Portuscript via WebAssembly direto no navegador.

#### 5.8 — Gerador de Diagramas de Arquitetura

- [ ] Comando `portuscript diagramar`: Mapeia graficamente as relações e fluxos de imports entre as pastas `/dominio`, `/infra` e `/web`.
- [ ] Detecção e aviso de violações das regras do Clean Architecture (ex: domínio importando infraestrutura).
- [ ] Exportação direta do diagrama para o formato Mermaid.md ou SVG para uso em documentação do projeto.

---

## 🔤 Sintaxe Alvo (referência)

```portuscript
// Classes com herança simples
classe Animal {
    inicializar(self, nome: Texto) {
        self.nome = nome
    }

    falar(self) -> Texto {
        retorne "..."
    }
}

classe Cachorro estende Animal {
    falar(self) -> Texto {
        retorne "Au! Eu sou " + self.nome
    }
}

// Tipagem opcional
funcao soma(a: Inteiro, b: Inteiro = 0) -> Inteiro {
    retorne a + b
}

// Tratamento de erros
tente {
    var resultado = operacaoArriscada()
} capture (e: Erro) {
    imprimir("Erro: " + e.mensagem)
} finalmente {
    limpar()
}

// Corotinas
assincrono funcao buscarDados(url: Texto) {
    var resposta = aguarde http.obter(url)
    retorne resposta.json()
}

// Sinais (reatividade)
var [contador, definirContador] = sinal(0)

efeito(funcao() {
    imprimir("Contador: " + contador())
})

definirContador(42)

// Componente JSX-like (Fase 4)
funcao BotaoContador() {
    var [n, definirN] = sinal(0)
    retorne <botao aoClicar={funcao() { definirN(n() + 1) }}>
        Cliques: {n()}
    </botao>
}
```

---

## 📅 Estimativa de Tempo (solo, part-time ~10–15h/semana)

| Fase | Descrição                      | Estimativa |
| ---- | ------------------------------ | ---------- |
| 1    | Núcleo (classes, tipos, erros) | 2–3 meses  |
| 2    | VM + bytecode + GC             | 3–4 meses  |
| 3    | Stdlib backend                 | 2–3 meses  |
| 4    | Frontend + sinais              | 3–4 meses  |
| 5    | Tooling + ecossistema          | Contínuo   |

> **Total para v1.0 completa:** ~10–14 meses de trabalho consistente

---

## 🚀 Primeiros passos imediatos

```
FASE 1 → Completar núcleo
├── [1] Erros visuais em PT com sugestões e sublinhado
├── [2] Classes com herança simples
├── [3] Constantes imutáveis em runtime
├── [4] Sistema de tipos opcional + TypeChecker
└── [5] Parâmetros de função completos (padrão + nomeados)
```

---

_Roadmap criado em 2026-07-14. Revisitar e atualizar após conclusão de cada fase._
