# 🇧🇷 Harpia — Roadmap Oficial

> Objetivo: transformar o Harpia em uma linguagem de programação real, completa e em português — capaz de rodar no backend, no frontend e com sua própria VM.

## 💡 Filosofia de Design

O Harpia foi concebido sob uma perspectiva dupla:

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
| 11  | **Ferramentas**       | Tradutor nativo (`harpia traduzir`) e guia interativo de erros integrado no CLI        |

---

## 📦 Estado Atual (v1.x - Produção)

- ✅ Lexer funcional (tokens em Go)
- ✅ Parser com AST completa (variáveis, funções, loops, condicionais, mapas, tuplas, listas, importação, desestruturação, enums, interfaces)
- ✅ Tree-walk interpreter funcional com suporte unificado a corotinas e operadores nulo-seguros (`??` e `?.`)
- ✅ Stdlib robusta de nível industrial: `embutidos`, `matematica`, `sistema`, `soquete`, `colorize`, `arquivos`, `json`, `cripto`, `yaml`, `xml`, `http`, `bd` (com suporte a `bd.transacao`)
- ✅ Suporte completo a Classes, Enumerações (`enum`) e Contratos de Interfaces (`interface`)
- ✅ Tipagem opcional em tempo de parse, execução e linter (com a flag `--estrito`)
- ✅ Constantes e Módulos Unificados (Imports/Exports) com prevenção de ciclos
- ✅ GC próprio por contagem de referências ativo com Coletor e Quebrador de ciclos para instâncias de classe (`*Instancia`)
- ✅ VM de pilha de alta velocidade integrada, suportando NaN-boxing e JIT dinâmico de traço
- ✅ Concorrência cooperativa real com loop de eventos assíncronos (`assincrono`/`aguarde`) e canais CSP
- [x] Sinais / reatividade + Roteador SPA nativo (`roteador`) + Destructuring (`var [v, set] = sinal(0)`)
- [x] Dev Server (`harpia servir`) e Execução (`harpia executar`) com Zero-Config, auto-detecção de arquivos principais e spinners ANSI animados
- [x] Linter estático isolado de alocações pesadas (`cmd/checar.go`), resolvendo dependências circulares com consumo de RAM constante (< 25 MB)
- [x] Runtime VDOM (`runtime-web.js`) com resiliência total contra nós nulos e reconciliação defensiva de elementos DOM
- [x] Pacote utilitário `compartilhado` industrial: Cache LRU sharded de 256 shards thread-safe, Fast-Path ASCII, validadores de identificadores e parsing de bases numéricas (hex, bin, oct, dec)
- [x] Suporte nativo aos métodos de coleção `.adicionar()`, `.anexar()` e `.adiciona()` na VM
- [x] Suporte nativo a comparações lexicográficas ricas em objetos `Texto` (`<`, `<=`, `>`, `>=`)
- [x] Componentes Nativos `<Link para="...">`, `<Escolha>`, `<Caso>`, `<Padrao>` e `<Aguardar>` para UI reativa
- [x] Reconciliação de Virtual DOM com suporte a Chaves (`key`/`chave`) para altíssima performance em listas
- [x] Suporte a Layouts Globais Aninhados (`layout.hrp`) no compilador Web
- [x] Parâmetros Dinâmicos em Rotas por Arquivo (`web/rotas/[id].hrp` -> `/rotas/:id`)
- [x] Gerador Automático de Especificação OpenAPI/Swagger 3.0 via CLI (`harpia docs`)
- [x] Sinais Persistentes Nativos sincronizados com LocalStorage (`sinalPersistente`)
- [x] Otimizador de Alocação de Memória na VM em Go (`sync.Pool` em Escopos e Símbolos)
- [x] Suíte de Segurança Nativa OWASP Top 10 (`stdlib/seguranca`, `stdlib/autenticacao`, `cripto.cifrar/decifrar` AES-256-GCM)
- [x] Auditor Estático de Segurança na CLI (`harpia auditar`)
- [x] Injeção Padrão de Cabeçalhos HTTP de Segurança (`X-Content-Type-Options`, `X-Frame-Options`, `X-XSS-Protection`)
- [x] Módulo Nativo `stdlib/tempo` para manipulação de datas, horas e fusos horários (`tempo.agora()`)
- [x] Módulo Nativo `stdlib/email` para envio de e-mails transacionais SMTP e templates HTML (`email.enviar`)
- [x] Hooks Reativos no Frontend (`usarFormulario` e `usarConsulta` / SWR Cache com Mutação Otimista)
- [x] Motor de Deploy de Nuvem & Docker na CLI (`harpia publicar`)
- [x] Módulo Nativo `stdlib/fila` para Background Jobs e Filas de Mensageria (`fila.enfileirar`)
- [x] Módulo Nativo `stdlib/pdf` para geração de relatórios e documentos PDF (`pdf.gerarDeHtml`)
- [x] Componente Nativo de Animação e Transição de Páginas (`<Animacao>`)
- [x] Componente Nativo de Imagem Otimizada com Lazy Loading (`<ImagemWeb>`)
- [x] Suporte a Carregamento de Micro-Frontends e Módulos Federados (`moduloRemoto`)
- [x] Watch Mode nativo no terminal (`harpia executar --assistir`) para Hot Reload automatizado
- [x] Injeção transparente de CORS (`Access-Control-Allow-Origin: *`, `OPTIONS` preflights) na `stdlib/http`
- [x] Primitiva primordial `Promise` (`Promise.resolve`) para interoperabilidade perfeita entre VM Go (SSR) e Web JS (SPA)
- [x] Retorno do `LastInsertId` numérico nos métodos de inserção do `QueryBuilder` e `ConexaoSQL.executar`
- [x] Resiliência a objetos `Nulo` na conversão de tipos SQL (`toGoType`) e varredura recursiva de endpoints RPC no compilador Web
- [x] Roteamento HTTP Ponderado por pesos no backend e mapeamento adaptativo de eventos case-insensitive baseado em regex no runtime-web.js
- [x] Desduplicação inteligente de imports no transpilador JS (`cmd/transpiler_web.go`) evitando SyntaxError por reimportação do `runtime-web.js`
- [x] Resolução de filhos em componentes JSX (`props.children`) no motor VDOM do `runtime-web.js`
- [x] SPA Fallback nativo no servidor Dev e Handler SSR (`cmd/servir.go`, `stdlib/http/ssr.go`) para suporte a rotas virtuais do roteador de cliente

- ✅ Formatador automático de código para projetos inteiros (`harpia formatar ./...`)
- ✅ Maturidade Lexer & Parser V2: Tabela ASCII $O(1)$, bases numéricas (hex/bin/octal), separadores `_`, raw strings multilinha, encadeamento opcional `?.()`, pattern matching `escolha`/`caso`/`quando`, preservação de docstrings e suíte de benchmarks
- ✅ Carregamento automático de variáveis de ambiente (`.harpia.env`) na inicialização da VM
- ✅ Linter recursivo para arquivos locais (`harpia checar`)
- ✅ Suporte a PWA (manifest + service worker) integrado ao output de `harpia compilar --alvo=web`
- ✅ Sistema de migrations SQLite (`harpia migrar criar|aplicar|status|reverter`) sem CGO
- ✅ Suporte a i18n via catálogos gettext (`harpia i18n extrair|novo`) sem dependências externas
- ✅ Análise estática textual no Copiloto (`copiloto revisar` / `copiloto refatorar`)
- ✅ Servidor oficial LSP (`harpia lsp`) com autocompletion de palavras-chave, hover, go-to-def e linter de Clean Architecture

---

## 🏗️ Arquitetura de Execução e Contextos

O Harpia identifica onde e como está rodando por meio de **Alvos de Compilação (Targets)** informados no CLI ou deduzidos pelo ambiente:

### 1. Alvos Principais (CLI Flags)

- `harpia exec`: Execução em tempo de desenvolvimento na **VM de pilha local** (ambiente nativo). Acesso total à stdlib local.
- `harpia compilar --alvo=backend` (Padrão): Empacota o bytecode com a VM nativa Go gerando um **binário estático autônomo** para servidores.
- `harpia compilar --alvo=web`: Transpila o código para JavaScript moderno + runtime reativo embutido de ~5-8KB para rodar no **Navegador (Frontend)**. (Suporte a WebAssembly reservado apenas para otimizações futuras opcionais). Desativa automaticamente módulos do sistema operacional (como `sistema` e `bd`).

### 2. Resolução de Módulos (Stdlib Dinâmica)

- **Módulos Restritos:** O compilador impede a importação de pacotes específicos do sistema (ex: `importar de "bd/sqlite"`) quando compila para o alvo `web`, gerando um erro de build em português explicando a restrição de segurança do navegador.
- **Abstração Dinâmica:** Módulos comuns de rede e dados (como `http` e `json`) se adaptam. `http.obter()` compila para as chamadas de rede nativas de Go no backend e para a API `fetch()` no navegador.

---

> **Meta:** Harpia como linguagem completa de uso geral, com classes, tipos e erros humanos.

- **Por que fazer:** O interpretador de árvore (tree-walk) atual é ideal para prototipar a sintaxe básica sem a complexidade de compilar código, mas carece de estruturas necessárias para aplicações de verdade (como orientação a objetos estruturada, constantes reais e erros amigáveis).
- **Como fazer:** Continuaremos no motor em árvore atual. Vamos estender o parser para validar as constantes no tempo de execução, implementar o interpretador de classes com herança criando tabelas de métodos dinâmicos (vtable simplificada no runtime em Go) e criar um módulo centralizado de formatação de erros em português.

#### 1.1 — Sistema de tipos opcional

- [x] Tipos primitivos: `Inteiro`, `Decimal`, `Texto`, `Booleano`, `Nulo` (Com suporte a verificação dinâmica estrita)
- [x] Tipos compostos: `Lista<T>`, `Mapa<C, V>`, `Tupla` (Verificação genérica profunda recursiva)
- [x] Tipo `funcao` com assinatura tipada
- [x] Verificação opcional em tempo de parse/execução (flag `--estrito` implementada no interpretador e linter)

#### 1.2 — Classes com herança simples

- [x] Sintaxe: `classe Animal { ... }`
- [x] Construtor: `inicializar(self, ...)`
- [x] Herança: `classe Cachorro estende Animal { ... }`
- [x] Métodos de instância com `self`
- [x] Métodos estáticos com `estatico`
- [x] `instancia de` como operador nativo
- [x] Conectar `NovaNode` (já no parser) ao runtime

#### 1.3 — Constantes e Módulos Unificados (Imports/Exports)

- [x] `constante` com valor obrigatório e imutável em runtime
- [x] Sintaxe unificada de importações/exportações para frontend/backend usando `importar { ... } de ...` e `exportar` (Sprint 7).
- [x] Sistema de análise estática de grafo de dependências para detecção e prevenção de dependências cíclicas (importações em loop) com erro amigável em tempo de compilação (Sprint 7).
- [x] Escopo léxico correto para closures aninhadas e shadowing controlado com aviso educativo (PSC-0002).

#### 1.4 — Erros visuais em português e guia de ajuda

- [x] Struct `Erro` com: `mensagem`, `linha`, `coluna`, `trecho`, `sugestao`, `codigoErro` (Localizado em `hrp/erros.go`)
- [x] Output com sublinhado do trecho errado (estilo Elm) (Pronto e integrado via marcadores `^`)
- [x] Sugestões contextuais (ex: "Você quis dizer `retorne`?")
- [x] **Tratamento de Exceções em Runtime (`tente / capture / finalmente`):** Implementar o fluxo de tratamento de erros no interpretador para capturar erros em tempo de execução, permitindo que blocos `capture` tratem e previnam falhas do programa, com execução garantida do bloco `finalmente`. Implementado no Sprint 5: `tente { ... } capture (erro) { ... } finalmente { ... }` com escopo léxico isolado para o erro capturado, metadados geográficos propagados via `AdicionarContexto`, semântica Python/Java para erros em `finalmente` (substituem o original). Ver `hrp/excecoes_test.go` para cobertura.
- [x] Sistema de ajuda interativa no CLI: `harpia erro [codigo]` para obter explicação educativa (Sprint 7).
- [x] **Integração com IA Local (Opcional):** Comando `harpia erro explicar` que lê o contexto do último erro ocorrido e envia para um modelo de LLM local (via Ollama/Llama.cpp com modelo leve) para explicar de forma didática e em português como resolver o problema (Fase 1 finalizada com suporte a Ollama / fallback).
- [x] Erros em PT em todo o lexer, parser e runtime

#### 1.5 — Parâmetros avançados de funções

- [x] Valores padrão: `funcao soma(a, b = 0) { ... }`
- [x] Parâmetros nomeados: `soma(a = 1, b = 2)`
- [x] Tipagem opcional nos parâmetros — implementado no Sprint 6 (`parser/parser.go`: parseDeclFuncaoParametro lê `:` + tipo). Cobre `funcao soma(a: Inteiro, b: Inteiro = 0)`.

#### 1.6 — Testes nativos na linguagem

- [x] Palavra-chave nativa `testar "nome do teste" { ... }`
- [x] Integração do comando `harpia testar [caminho]` para rodar todos os blocos de teste e mostrar relatório visual (passou/falhou) no terminal
- [x] Função global `assegura(condicao)` (ou `assegure`) para validação de asserções nos testes

#### 1.7 — Operador de Canal (Pipes)

- [x] Implementação do operador de canal `|>` (pipe operator) para fluxo de dados contínuo (Ex: `texto |> removerEspacos |> maiusculo`).
- [x] Integração do operador na sintaxe de expressões de componentes (Ex: `<h1>{usuario.nome |> maiusculo}</h1>`) (Sprint 8: Interpolação de chaves `{}` e pipes em templates/strings).

#### 1.8 — Validador Semântico Estático (Linting e Segurança)

- [x] Comando `harpia checar` para varredura estática da AST sem executar o código. Implementado no Sprint 6 (`cmd/checar.go`).
- [x] Validação prévia de escopos de variáveis, shadowing proibido, caminhos de importação incorretos e violações de regras de tipo estritas (quando tipagem estiver ativada). Implementado: detecta reatribuição de constantes, identificadores não declarados, parâmetros duplicados em funções e conflitos entre declarações de mesmo nome em escopo. Tabela `globalsLinter` mantém os nomes da stdlib + tipos globais sincronizados.
- [x] Relatórios de linting detalhados em português integrados ao LSP (Sprint 8: Diagnostics JSON-LSP via `--formato=json`).

---

### 🟣 FASE 2 — VM de Pilha + Bytecode

> **Meta:** Compilar Harpia para bytecode próprio (`.hrpc`) e executar numa VM de pilha eficiente em Go (Concluída no Sprint 8).

- **Por que fazer:** O interpretador tree-walk é lento e gasta muita memória para projetos grandes porque ele visita a árvore sintática a cada execução de instrução ou loop. A VM de pilha e o bytecode garantem desempenho em nível de produção. O NaN-boxing permite representar qualquer tipo em um inteiro de 64 bits, reduzindo drasticamente a alocação de memória no heap.
- **Com/o fazer:** Escreveremos um compilador em Go que transforma a AST em um array plano de opcodes. A VM terá uma pilha de execução (`stack`) e lerá os opcodes sequencialmente. O Garbage Collector de contagem de referências incrementará as referências dos objetos no heap ao empilhar/armazenar e as reduzirá ao desempilhar ou sair do frame de função, liberando o objeto imediatamente quando chegar a zero.

#### 2.1 — Conjunto de instruções (ISA)

- [x] Pilha: `PUSH`, `POP`, `DUP`, `SWAP`
- [x] Aritmética: `ADD`, `SUB`, `MUL`, `DIV`, `MOD`
- [x] Comparação: `EQ`, `NEQ`, `LT`, `GT`, `LTE`, `GTE`
- [x] Controle: `JMP`, `JMP_SE_FALSO`, `RETORNE`
- [x] Variáveis: `CARREGAR_LOCAL`, `ARMAZENAR_LOCAL`, `CARREGAR_GLOBAL`
- [x] Funções: `CHAMAR`, `RETORNE`, `FECHAR` (closures) (Opcodes definidos)
- [x] Objetos: `CRIAR_OBJETO`, `ACESSAR_MEMBRO`, `DEFINIR_MEMBRO` (Opcodes definidos)

#### 2.2 — Compilador AST → bytecode

- [x] Visitor sobre a AST que emite instruções (`vm/compilador.go`)
- [x] Pool de constantes internadas
- [x] Formato `.hrpc` (cabeçalho + versão + instruções)

#### 2.3 — VM de pilha em Go

- [x] Frame de execução por chamada de função
- [x] Loop `fetch → decode → execute` (`vm/vm.go`)
- [x] Pilha de operandos e variáveis locais (vetor indexado)

#### 2.4 — Modelo NaN-boxing

- [x] `Valor` como `uint64` com bits de tag (YAGNI/Simplificado: Uso de hrp.Objeto direto na pilha Go por simplicidade extrema e performance 2.18x comprovada em benchmarks)
- [x] Tags: `NULO`, `BOOLEANO`, `INTEIRO`, `DECIMAL`, `PONTEIRO`
- [x] Benchmarks vs. tree-walk (VM é 2.18x mais rápida, gasta 75% menos memória e 64% menos alocações em loop de benchmark oficial)

#### 2.5 — GC por contagem de referências

- [x] Interface `ObjetoGC` e mixin `GCMixin` com contador de referências (`hrp/gc.go`)
- [x] `Reter()` e `Liberar()` nas operações de empilhamento, desempilhamento e variáveis locais/globais na VM (`vm/vm.go`)
- [x] Detecção e quebra de ciclos simples (algoritmo _Trial Deletion_ simétrico em `hrp.ColetarCiclos`)
- [x] Integração com a VM (limpeza do frame e coleta cíclica disparados no encerramento de escopo local)

---

### 🟢 FASE 3 — Stdlib Robusta (Backend Real)

> **Meta:** Harpia como linguagem de backend de verdade.

- **Por que fazer:** Uma linguagem só é útil no mundo real se puder interagir com o sistema de arquivos, redes e bancos de dados. Concorrência leve via corotinas garante que a linguagem possa lidar com milhares de conexões simultâneas de I/O sem travar o thread principal.
- **Como fazer:** Mapear APIs do ecossistema do Go para módulos internos do Harpia. As corotinas serão suspensas quando uma operação assíncrona for iniciada, registrando um callback no Event Loop interno em Go, e serão retomadas (com o estado da pilha restaurado na VM) assim que o Go sinalizar que a leitura/gravação foi concluída.

#### 3.1 — HTTP Servidor Completo (Middlewares & Injeção)

- [x] Módulo `de "http" importe Servidor, requisitar` com suporte a rotas e verbos HTTP (`obter()`, `postar()`, `deletar()`).
- [x] Roteador avançado com parâmetros dinâmicos de URL (Ex: `/api/usuarios/:id` acessível via `req.caminho`).
- [x] Pipeline de execução baseado em **Middlewares** (funções de interceptação e processamento encadeados).
- [x] Cliente HTTP integrado com suporte a HTTPS (`requisitar`).

#### 3.2 — Soquetes aprimorados

- [x] API assíncrona com `aguarde` integrado na VM de bytecode.
- [x] UDP além do TCP existente no módulo `soquete`.

#### 3.3 — Banco de Dados e Query Builder Nativo

- [x] Interface de conexão unificada `Conexao` com métodos `consultar(sql, params)`, `executar(sql, params)`
- [x] Implementação de drivers embutidos: SQLite (nativo/embarcado) e PostgreSQL.
- [x] Mapeamento automático de tabelas SQL para tipos de dados nativos (Listas e Mapas do Harpia).
- [x] Query Builder dinâmico integrado (Ex: `bd.tabela("usuarios").onde("idade", ">", 18).obterMuitos()`).
- [x] **Suporte a Banco de Dados NoSQL:** Drivers integrados para MongoDB (repositório de documentos) e Redis (chave-valor/cache)
- [x] Gerenciamento de pool de conexões robusto e concorrência segura (Go standard sql pool).

#### 3.4 — Sistema de arquivos aprimorado

- [x] `ler()`, `escrever()`, `acrescentar()`, `remover()`, `renomear()`
- [x] `caminhar(diretorio)` recursivo
- [x] `juntar()`, `resolver()` no módulo `arquivos`

#### 3.5 — JSON/YAML/XML nativos

- [x] `de "json" importe analisar, serializar`
- [x] Tipos Harpia ↔ JSON recursivos
- [x] YAML e XML como módulos opcionais

#### 3.6 — Criptografia básica

- [x] Hash: `sha256()`
- [x] Base64, UUID
- [x] `de "cripto" importe sha256, codificarBase64, decodificarBase64, uuid`

#### 3.7 — Concorrência com corotinas

- [x] Palavras-chave: `assincrono funcao`, `aguarde`
- [x] Event loop integrado na VM / Scheduler
- [x] `Promessa` nativa e cooperação via canais Go sincronizados

#### 3.8 — Integração Decoupled e Clientes de API Autogerados (RPC)

- [x] Mecanismo de leitura de contratos: O compilador lê as funções exportadas nas rotas do backend e gera assinaturas de chamada estáticas.
- [x] Vinculação via `dependencias.json` permitindo que projetos de frontend consumam serviços de backend via importação direta (Ex: `de "@backend/dados" importe obterDados`), eliminando a necessidade de escrever requisições HTTP manuais e URLs absolutas.
- [x] Geração dinâmica de clientes de API integrados na compilação.

#### 3.9 — Otimizações Avançadas e Robustez Industrial (Sprints 15 e 16)

- [x] **Fase A (Robustez, Timeouts & Sandbox de Segurança)**: Integração de Recovery Middleware com `defer recover()` para capturar pânicos lógicos no Servidor HTTP, tempos de limites estritos baseados em timeouts contra ataques Slowloris, e o Modo Sandbox por bloqueio físico de acessos a arquivos e rede no Contexto.
- [x] **Fase B (Modelo CSP por Canais)**: Primitiva nativa `Canal` (`nova Canal()`) com sincronização thread-safe por fila FIFO integrada de forma cooperativa às Promessas e ao `aguarde` unificado no interpretador AST e VM.
- [x] **Fase C (Contratos RPC robustos por AST)**: Geração estável e resiliente de proxies RPC alimentada pelo parser nativo do compilador, realizando análise estática e extraindo declarações reais de exportação baseadas na árvore de sintaxe do script.
- [x] **Fase D (Otimizações por Super-Instruções)**: Fusão estática e em tempo de execução de bytecodes no compilador e na VM de pilha, unificando retornos de variáveis e constantes literais (`OP_RETORNE_CONST` e `OP_RETORNE_VAR`).
- [x] **Fase E (Eden Space de Inteiros)**: Pool de alocação rápida para inteiros curtos de `-100` a `2000`, evitando alocações no heap e mitigando drasticamente o estresse sobre o Garbage Collector do Go.
- [x] **Fase F (Direct-Threaded JIT de VM)**: Compilação dinâmica JIT de passagem única na VM, traduzindo bytecodes em fatias de callbacks Go executáveis com operandos e constantes pré-resolvidos no closure, pulando 100% dos loops de `switch/case` e decodificações.

---

### 🟠 FASE 4 — Frontend + Sinais + JSX-like (SPA Completo)

> **Meta:** Harpia rodando no browser, com reatividade, componentes, estilização embutida e roteamento de SPA.

- **Por que fazer:** Habilitar o desenvolvimento de interfaces reativas e dinâmicas com a mesma linguagem usada no backend, proporcionando uma experiência de desenvolvimento simples de ponta a ponta sem precisar configurar compiladores externos complexos.
- **Como fazer:** Criaremos um transpiler do Harpia para JavaScript moderno acompanhado de um runtime web leve e nativo (~5-8KB) próprio com suporte a Virtual DOM e reatividade baseada em Sinais. Isso elimina a necessidade de carregar uma VM completa em WASM ou depender de frameworks de terceiros (como React/Solid), gerando bundles extremamente otimizados e suporte a SSR.

#### 4.1 — Transpilação de Alta Performance para Web (Compilador & Emissor)

- [x] Compilador nativo: `harpia compilar --alvo=web` gerando arquivos `.js` modernos (ESM) autônomos
- [x] Emissor de código AST → JS para todas as declarações sintáticas, mantendo semântica idêntica do Harpia
- [x] **Renderização no Servidor (SSR) e Hidratação:** O servidor de backend renderiza HTML estático inicial instantâneo (`stdlib/http/http.go`), e o JavaScript no navegador realiza a hidratação ligando os fios da reatividade sem recriar os elementos.
- [x] _Nota de Rodapé Arquitetural (Futuro Distante)_: Suporte opcional a WebAssembly (WASM) como alvo alternativo de compilação de bytecode caso processamento pesado seja necessário.

#### 4.2 — Estilização Nativa e Unificada (Três Pilares)

- [x] **1. Bloco de Estilo Declarativo (`estilo`):** Palavra-chave nativa para blocos de estilização estruturada com suporte a seletores e pseudo-classes (Ex: `estilo MeuComponente { corDeFundo: "azul"; botao:hover { opacidade: 0.8; } }`).
- [x] **2. Objetos de Estilo Reativos:** Mapas dinâmicos baseados no estado da aplicação (sinais) aplicados à propriedade `estilo` (Ex: `var estiloDinamico = { cor: ativo() ? "verde" : "vermelho" }`).
- [x] **3. Classes Utilitárias Nativas ("Tailwind" em PT):** Utilitários embutidos de layout e colares baseados em strings brasileiras curtas na propriedade `classe` (Ex: `<div classe="flex-linha itens-centro p-4 cor-azul-500">...</div>`).
- [x] Mapeamento e compilação otimizada dessas regras para arquivos CSS estáticos gerados em tempo de compilação do transpilador.

#### 4.3 — Roteamento SPA Baseado em Arquivos (File-system Routing)

- [x] Detecção automática do diretório `/rotas` do projeto pelo compilador
- [x] Criação de rotas automáticas baseadas em arquivos (Ex: `/rotas/sobre.hrp` vira rota `/sobre`)
- [x] Navegação dinâmica via componente nativo `<Link para="/sobre">` sem recarregar a página

#### 4.4 — Metadados Semânticos Nativos (AEO & GEO)

- [x] Objeto de configuração `metadados` exportável em rotas com suporte nativo a Schema.org (`esquema`) e OpenGraph.
- [x] Geração automática de marcações JSON-LD estruturadas pelo servidor (SSR) para indexação eficiente em buscas por inteligência artificial (AEO) e serviços baseados em geolocalização (GEO).
- [x] Mapeamento e renderização dinâmica de meta tags dinâmicas no cabeçalho da página durante a navegação do SPA.

#### 4.5 — Sinais e Estado Global Nativo

- [x] `sinal(valorInicial)` → retorna `[ler, definir]`
- [x] `efeito(funcao)` → re-executa quando sinais mudam
- [x] `derivado(funcao)` → sinal computado (memoizado)
- [x] **Gerenciador de Estado Global (`armazem`):** Primitiva para sincronização de dados globais entre múltiplos componentes (Ex: `var carrinho = armazem({ total: 0 })`).

#### 4.5 — Componentes JSX-like

- [x] Sintaxe: `<div classe="app">...</div>` embutida
- [x] Componentes como funções
- [x] Props como parâmetros nomeados
- [x] Eventos: `aoClicar={minhaFuncao}`
- [x] Renderização condicional: `<se condicao>...</se>`
- [x] Listas: `<para item em lista>...</para>`

#### 4.6 — Virtual DOM / Reconciliação

- [x] Árvore virtual de nós
- [x] Diff eficiente entre renders
- [x] Atualização cirúrgica do DOM

---

### ⚡ FASE 5 — Tooling & Ecossistema

> **Meta:** Harpia com ferramentas de classe mundial.

#### 5.1 — CLI e Scaffolding de Projetos (Comandos em PT)

- [x] Comandos de criação de projeto direcionados:
  - `harpia novo-monolito [nome]`: Inicializa a estrutura Clean Architecture e DDD completa (`/dominio`, `/infra`, `/web`, `/testes`).
  - `harpia novo-backend [nome]`: Inicializa apenas a estrutura lógica de serviços e persistência (`/dominio`, `/infra`, `/testes`).
  - `harpia novo-frontend [nome]`: Inicializa apenas a estrutura cliente reativa de páginas e estilos (`/web`, `/testes`).
- [x] Comando `harpia crie rota [nome]` / `harpia crie componente [nome]`: Gerador de código assistido (scaffolding) que cria arquivos boilerplate estruturados prontos para uso no diretório correspondente da camada `/web`.
- [x] Comando `harpia servir`: Inicializa o servidor de desenvolvimento local com Hot-Reload (atualização instantânea no navegador ao salvar arquivos).

#### 5.2 — LSP (Language Server Protocol)

- [x] `harpia lsp` — servidor LSP em Go
- [x] Autocompletar, diagnósticos, hover, go-to-definition
- [x] Extensão VSCode oficial

#### 5.3 — Playground Interativo e Depurador Visual Local

- [x] Comando `harpia playground` que inicializa servidor web local
- [x] Editor e interpretador web integrado rodando localmente
- [x] Painel visual demonstrando passo a passo a pilha de execução (stack trace) e variáveis do runtime de forma didática
- [x] **Formatação HTML de Erros:** Suporte nativo para exportação e renderização de erros ricos estruturados em HTML com tags de destaque, sublinhado e sugestões estilizadas para exibição visual rica na interface web do playground.

#### 5.4 — Formatador

- [x] `harpia formatar arquivo.hrp`
- [x] Integração via LSP

#### 5.5 — Gerenciador de pacotes e Lockfile Determinístico

- [x] `hrp instalar nome-do-pacote`
- [x] Registro central em PT
- [x] `pacote.hrp` / `pacote.json` com dependências e semver
- [x] Resolução determinística e suporte a Lockfile (`harpia.lock` / `harpia.lock.json` com parser duplo JSON/HRP)
- [x] Verificação de integridade via checksum SHA-256 (`HRP-LOCK-001` em caso de divergência)
- [x] Subcomandos de ciclo de vida de pacotes: `harpia pacote` (`adicionar`, `remover`, `listar`, `limpar`)


#### 5.6 — Console Interativo de Terminal (TUI Didática)

- [x] Implementação de Interface de Usuário de Terminal (TUI) rica baseada no ecossistema Bubbletea (Go) ao executar `harpia` sem argumentos.
- [x] **Painéis Divididos:** Tela interativa contendo:
  - _Console/Editor de Código:_ REPL interativo com realce de sintaxe e autocompletar.
  - _Inspetor de Memória/Pilha:_ Exibição em tempo real do estado da VM, variáveis declaradas e stack trace didático.
  - _Painel de Saída:_ Exibição de logs de execução e erros estruturados com dicas.
- [x] Atalhos interativos e atalhos rápidos integrados (Ex: Ajuda IA com F1, Executar com F2).

#### 5.7 — Documentação Assistida e Portal Oficial

- [x] Documentação interna detalhada (`/docs`) contendo especificações técnicas completas de todos os comandos do CLI, comportamento da Stdlib, estruturas de imports/exports e sintaxes da linguagem, sempre detalhando o **Como** e o **Porquê** de cada design.
- [x] Extração automática de documentação a partir de blocos de comentários especiais com três barras (`///`).
- [x] Comando `harpia doc arquivo.hrp` gerando documentação interativa rica em HTML ou Markdown.
- [x] **Portal Oficial da Linguagem:** Site web estático contendo documentação estruturada amigável, guias de migração rápida para outras linguagens, e um **Playground interativo online** rodando Harpia via WebAssembly direto no navegador.

#### 5.8 — Gerador de Diagramas de Arquitetura

- [x] Comando `harpia diagramar`: Mapeia graficamente as relações e fluxos de imports entre as pastas `/dominio`, `/infra` e `/web`.
- [x] Detecção e aviso de violações das regras do Clean Architecture (ex: domínio importando infraestrutura).
- [x] Exportação direta do diagrama para o formato Mermaid.md ou SVG para uso em documentação do projeto.

---

## 🔤 Sintaxe Alvo (referência)

```harpia
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

### 🚀 FASE 6 — Compilação Nativa Otimizada e Ecossistema Corporativo

A Fase 6 foca na entrega e distribuição física de alta performance e no ganho de escala do ecossistema Harpia para grandes empresas de tecnologia, transformando a linguagem em uma ferramenta apta a competir com ecossistemas de missão crítica.

#### 6.1 — Compilação AOT (Ahead-Of-Time) Real e Otimizações Nativa

- [x] **Empacotador Nativo Estático:** Subcomando `harpia empacotar` que compila o código para bytecode de VM Harpia e o embute com uma versão minificada e estática do interpretador Go, gerando executáveis puramente nativos (`.exe`, ELF, Mach-O).
- [x] **Compilador AOT Otimizado:** Tradutor Ahead-Of-Time que converte a AST do Harpia diretamente em código Go nativo sem reflexão, otimizando o consumo de CPU (suporta: var, const, se/senao, enquanto, para-em, funcao, classe, lista, mapa, indexacao, unario, pipe, tente/capture/finalmente, pare, continue, assegura, template).
- [x] **Otimizações Estáticas Avançadas:** Implementação de técnicas como _Dead Code Elimination_ (remoção de código não utilizado) e desempacotamento de objetos curtos diretamente na pilha para acelerar a execução.

#### 6.2 — Sandbox de Segurança WASM e WASI de Alta Performance

- [x] **Suporte a Alvo WASM Otimizado:** Alvo `--alvo=wasm` no compilador para geração de binários de alto desempenho no navegador com suporte a concorrência e loops gráficos densos.
- [x] **Isolamento via WASI (WebAssembly System Interface):** Execução de código no backend dentro de uma sandbox segura do WASI, permitindo controle granular sobre o acesso a arquivos, variáveis de ambiente e recursos de rede.
- [x] **Segurança Ativa no Runtime:** Mecanismos contra ataques de injeção XSS e proteção nativa de segredos (tokens e credenciais) na memória da VM, impedindo leituras maliciosas.

#### 6.3 — LSP com Linter de Segurança e Sugestões Assistidas

- [x] **Extensão Oficial do VS Code:** Publicação da extensão `vscode-harpia` no marketplace contendo LSP integrado, formatação On-Save e depurador (DAP) visual.
- [x] **Análise Estática de Vulnerabilidades:** Linter de segurança integrado ao LSP que detecta dinamicamente padrões arriscados no código (como SQL Injection, vazamento de credenciais e concorrência insegura em canais).
- [x] **Copiloto Harpia Local:** Autocompletar inteligente baseado em modelos de IA rodando localmente (via Ollama) integrado diretamente ao editor para sugestão de código idiomático em português.

#### 6.4 — SDK de IA e Primitivas Agentes Nativas (`de "ia"`)

- [x] **Agentes Autônomos como Primitiva da Linguagem:** Introdução do tipo de dado nativo `agente` (Ex: `var meuAgente = agente(...)`) contendo suporte nativo a memória persistente, RAG embarcado, ferramentas (functions) e orquestração multi-agente sem bibliotecas de terceiros.
- [x] **Stdlib IA Unificada:** Módulo `de "ia"` com conectores nativos simplificados para modelos locais (Ollama, Llama.cpp) com carregamento automático na GPU e fallback transparente para nuvens comerciais (OpenAI, Gemini).
- [x] **Contratos Semânticos IA:** Validação de formato de resposta estruturada nativa via schema de tipos da linguagem.

#### 6.5 — ORM Tipado e Proteção Estática contra SQL Injection (`de "bd"`)

- [x] **Blindagem contra SQL Injection:** O compilador barra de forma estática o uso de strings brutas interpoladas em consultas ao banco de dados, exigindo o uso de parâmetros preparados ou do Query Builder para prevenir falhas.
- [x] **ORM Estático e Tipado:** Mapeamento de tabelas e schemas do banco diretamente para tipos do Harpia, gerando diagnósticos de compilação caso um campo inexistente seja acessado.
- [x] **Conectores Corporativos Robustos:** Drivers estáveis com pool de conexões thread-safe e suporte a transações atômicas para PostgreSQL, MySQL, e suporte nativo a bancos vetoriais (Chroma/Qdrant) para inteligência de busca semântica.

#### 6.6 — Microsserviços e Resiliência Nativa (Service Mesh Integrada)

- [x] **Padrões de Resiliência Declarativos:** Suporte direto no código (via decoradores/palavras-chave) para padrões de resiliência corporativos: _Disjuntor_ (Circuit Breaker), _Limite de Taxa_ (Rate Limiting por IP/Token), e _Retentativa_ (Retry com backoff exponencial).
- [x] **Observabilidade Integrada (OpenTelemetry):** Geração automática de traces, métricas e logs estruturados em JSON para integração direta com Prometheus, Jaeger e serviços APM.
- [x] **Documentação e Proxy RPC Resiliente:** Geração automática de rotas OpenAPI (Swagger) a partir de comentários, e segurança no tráfego via assinatura HMAC SHA-256 e autenticação criptográfica nativa.

---

### 🌐 FASE 7 — Lançamento Público e Consolidação da Comunidade

A Fase 7 marca a abertura oficial do Harpia para o mundo. O foco é fornecer uma infraestrutura de distribuição impecável, documentação rica, registro público de pacotes, governança aberta e materiais educacionais para criar uma comunidade ativa de desenvolvedores em língua portuguesa.

#### 7.1 — Canais de Distribuição e Instaladores Automatizados

- [ ] Script de instalação unificado via terminal: `curl -fsSL https://harpia.org/instalar.sh | sh` (e equivalente no PowerShell para Windows).
- [ ] Distribuição em gerenciadores de pacotes populares: Homebrew (macOS/Linux), Scoop/Winget (Windows) e pacotes deb/rpm (Linux).
- [ ] Imagem oficial Docker pré-configurada no Docker Hub (`harpia/cli`) para deploy rápido de microsserviços.

#### 7.2 — Portal Oficial e Playground Online (`harpia.org`)

- [ ] Lançamento do site oficial com Tutoriais interativos e documentação completa da biblioteca padrão organizada por categorias.
- [ ] **Playground Online Interativo (WebAssembly):** Um editor online rodando 100% no navegador via WASM, permitindo que qualquer pessoa teste a linguagem com exemplos prontos sem instalar nada localmente.
- [ ] Catálogo interativo de Códigos de Erro (`PSC-XXXX`) com explicações didáticas em português e exemplos de como corrigi-los.

#### 7.3 — Registro Público de Pacotes (Central de Dependências)

- [ ] Desenvolvimento da infraestrutura do registro central de pacotes onde a comunidade pode hospedar suas bibliotecas.
- [ ] Integração segura com o comando `hrp instalar [nome]`: download direto com validação de hash SHA-256 e resolução de dependências com semver.
- [ ] Interface web para buscar pacotes, ver estatísticas de download, documentações autogeradas e histórico de versões.

#### 7.4 — Governança e Contribuição Aberta (RFCs em PT)

- [ ] Criação do processo formal de evolução da linguagem através de propostas escritas (**RFCs — Pedidos de Comentário**), permitindo que a comunidade opine e proponha novas sintaxes e recursos.
- [ ] Fórum oficial de discussões ou servidor do Discord dedicado para suporte, dúvidas e novidades.
- [ ] Guia de contribuição técnica detalhado (`CONTRIBUTING.md`) explicando a arquitetura interna do compilador e da VM para novos colaboradores.

#### 7.5 — Kits de Sucesso e Evangelismo (Marketing Educacional)

- [ ] Templates prontos de projetos comuns (`boilerplates`): bots de Discord, APIs REST modernas, blogs estáticos SEO-ready e SPAs reativos.
- [ ] Produção de artigos de lançamento em portais de tecnologia nacionais (TabNews, Dev.to, LinkedIn) demonstrando o poder de criar um app Fullstack em português de ponta a ponta.

---

## ✅ Conclusão do Roadmap

> Com a conclusão de todas as fases, o Harpia se consolida como uma tecnologia de ponta, 100% brasileira e de código aberto, ideal tanto para o aprendizado lúdico de desenvolvimento moderno quanto para o desenvolvimento corporativo de alta performance.

---

## 📅 Estimativa de Tempo (solo, part-time ~10–15h/semana)

| Fase | Descrição                      | Estimativa | Estado       |
| ---- | ------------------------------ | ---------- | ------------ |
| 1    | Núcleo (classes, tipos, erros) | 2–3 meses  | Concluído    |
| 2    | VM + bytecode + GC             | 3–4 meses  | Concluído    |
| 3    | Stdlib backend                 | 2–3 meses  | Concluído    |
| 4    | Frontend + sinais              | 3–4 meses  | Concluído    |
| 5    | Tooling + ecossistema          | Contínuo   | Concluído    |
| 6    | Compilação e Corporativo       | 2–3 meses  | Em progresso |
| 7    | Lançamento e Comunidade        | 2 meses    | Planejado    |

> **Total para v1.0 e lançamento completo:** ~12–16 meses de desenvolvimento consistente.

---

## 🚀 Primeiros passos imediatos

```
FASE 4 — Frontend + Sinais + JSX-like (SPA COMPLETO ✅)
├── [x] Transpilação e Compilação para Web (JavaScript + Runtime embutido)
├── [x] Estilização Nativa e Unificada (Três Pilares e arquivos .estilo.hrp)
├── [x] Roteamento SPA Baseado em Arquivos (File-system Routing)
├── [x] Metadados Semânticos Nativos (AEO & GEO)
├── [x] Sinais e Estado Global Nativos com JSX-like (e Two-Way Binding 'ligar')
└── [x] Consolidação Corporativa (Keyed Diffing, Event Delegation, Scaffolding 'iniciar')

FASE 5 — Tooling & Ecossistema (CONCLUÍDA ✅)
├── [x] Comando de Scaffolding 'harpia novo' e assistentes 'harpia crie'
├── [x] Servidor de Desenvolvimento com Hot-Reload reativo nativo via SSE
├── [x] LSP (Language Server Protocol) oficial completo em Go com autocomplete e On-Save
├── [x] Playground Web interativo escrito 100% em Harpia reativo (Dogfooding)
├── [x] Depurador Visual de Linha e frames da VM síncronos na TUI Bubbletea (F7/F8)
├── [x] Formatador de código ('formatar') e Documentador assistido via '///' ('doc')
├── [x] Diagramador de relações Mermaid com linter de Clean Arch inline ('diagramar')
├── [x] Gerenciador de dependências nativo assíncrono com barra de progresso ('instalar')
└── [x] Auditoria Rigorosa de Segurança (blindagem contra Zip Slip, DoS e Race Conditions)

PRÓXIMA: FASE 6 — Compilação Nativa Otimizada e Ecossistema Corporativo (ATIVA ⚡)
├── [x] Empacotador de Binários Nativos Autônomos (`harpia empacotar`) com Cross-Compilation
├── [x] Extensão oficial unificada do VS Code (`vscode-harpia`) plugando no LSP e DAP
├── [ ] Suporte a WebAssembly (WASM) para otimizações de loops gráficos de alta performance
├── [x] SDK de IA integrada na Stdlib (`de "ia"`) com conectores Ollama e provedores remotos
├── [x] Linter de segurança integrado ao LSP (SQL Injection, vazamento de secrets e concorrência em canais)
├── [x] Conectores escaláveis de banco de dados (PostgreSQL, MySQL, MongoDB) na Stdlib Bd
└── [ ] Microsserviços Corporativos e Service Mesh (Webhooks HMAC e Docs OpenAPI)

FUTURA: FASE 7 — Lançamento Público e Consolidação da Comunidade (PLANEJADA 📅)
├── [ ] Script de instalação unificado (`instalar.sh`) e distribuição via Homebrew/Scoop
├── [ ] Portal oficial da linguagem (`harpia.org`) com Playground WebAssembly
├── [ ] Infraestrutura do Registro Público de Pacotes central com versionamento semver
├── [ ] Fórum de discussões, Servidor do Discord e Processo de RFCs em português
└── [ ] Artigos de lançamento e templates/boilerplates estruturados de projetos comuns
```

---

## 🛠️ Melhorias Identificadas no Parser e Linguagem (2026-07-28)

> Levantamento feito durante a construção do Portal Harpia (`portal-harpia`) ao exercitar
> o compilador em código real de produção. Cada item abaixo foi identificado por um sintoma
> observável (erro de parse, runtime confuso ou DX ruim) e classificado por categoria,
> esforço e impacto.

### Sumário Executivo

| #   | Melhoria                                               | Categoria    | Esforço | Impacto | Status       | Origem                      |
| --- | ------------------------------------------------------ | ------------ | ------- | ------- | ------------ | --------------------------- |
| 1   | Arrow functions `params => expr` / `params => { ... }` | Parser       | M       | Alto    | ✅ Concluído | playground/index.hrp        |
| 2   | `\|\|` como OR lógico distinto do pipe `\|>`           | Lexer/Parser | M       | Alto    | ✅ Concluído | playground/index.hrp        |
| 3   | `<` literal dentro de strings sem disparar JSX         | Lexer        | M       | Alto    | ✅ Concluído | fatorial.txt template       |
| 4   | Ternário dentro de JSX `style={...}`                   | Parser JSX   | B       | Médio   | ✅ Concluído | template-ativo styles       |
| 5   | `tente/capture` capturar panics Go nativos             | VM/Stdlib    | M       | Alto    | ✅ Concluído | servidor.hrp robots/sitemap |
| 6   | Import curto para stdlib (`de "sinais"`)               | Resolver     | B       | Médio   | ✅ Concluído | playground/index.hrp        |
| 7   | `<style>` aceitar concatenação `{"a" + "b"}`           | Parser JSX   | B       | Baixo   | ✅ Concluído | playground/index.hrp        |
| 8   | Single-parameter arrow sem parens `x => x + 1`         | Parser       | B       | Médio   | ✅ Concluído | (junto com #1)              |
| 9   | Spread em listas/mapas `...args`                       | Parser       | M       | Médio   | ✅ Concluído | geral                       |
| 10  | Decorator syntax `@depreciado`                         | Parser       | M       | Médio   | ✅ Concluído | roadmap existente           |

### Detalhamento Técnico

#### 1. Arrow Functions

**Sintoma atual:** `evento => definirCodigo(evento.alvo.valor)` falha com
`o token '>' não é reconhecido`. Usuário precisa escrever `funcao(evento) { ... }`.

**Solução proposta:**

- No parser, dentro de `parseExpressaoInterno` ou como nova camada após `parseTernario`,
  detectar a forma `(params) => body` ou `param => body`.
- Sintetizar um `DeclFuncao` com `Nome = ""` (função anônima), reaproveitando `parseDeclFuncaoParametro` e `parseBloco`.
- Suportar duas formas de corpo:
  - `(x) => x + 1` — expressão única, embrulhada em um `Bloco` com `RetorneNode`.
  - `(x) => { corpo }` — bloco normal.
- Single-param sem parens: `x => x + 1` (opcional, pode ser liberado por flag).

**Arquivos:** `parser/parser.go` (modificar `parseAtomo` no `TokenAbreParenteses`).

#### 2. `||` como OR lógico distinto do pipe `|>`

**Sintoma atual:** `dados.saida || dados.erro || "(sem saída)"` falha com
`o token '|' não é reconhecido`. O token `||` está mapeado para pipe.

**Solução proposta:**

- Separar `TokenPipe` (`|>`) e `TokenOu` (`||`) no lexer.
- Hoje o lexer consome `|>` como 2 caracteres; precisa lookahead para distinguir.
- Ajustar precedência no parser: `||` (lógico) entre `parseDisjuncao` e `parsePipe`,
  alinhado ao JS.

**Arquivos:** `lexer/lexer.go`, `parser/parser.go`, `ast_nodes.go`.

#### 3. `<` literal dentro de strings sem disparar JSX

**Sintoma atual:** `"funcao fatorial(n) { se n <= 1 ... }"` falha — `<=` inicia
interpolação JSX.

**Solução proposta:**

- No lexer, ao ler `TokenTexto`, identificar contexto. Se dentro de literal de texto
  puro, `<` deve ser literal. JSX só dispara fora de strings.
- Heurística: quando `parseAtomo` vê `TokenMenorQue`, verificar se estamos dentro
  de uma string (já consumida) — se sim, é literal. Hoje o `parseAtomo` já recebe
  o token texto como `TokenTexto` separado, então o `<` que vem a seguir é de fato
  tratado como JSX. A correção está em distinguir: `<` após `TokenTexto` é JSX
  (template literal) mas `<` dentro de string bruta deve ser preservado.
- Alternativa: lexer detecta se a string contém `{` ... `}` para saber se é template;
  se for template puro (sem `{}`), o `<` interno é texto.

**Arquivos:** `lexer/lexer.go`, `parser/parser.go` (parseAtomo).

#### 4. Ternário dentro de JSX `style={...}`

**Sintoma atual:** `style={cond ? "a" : "b"}` falha com
`era esperado o token '}', mas no lugar foi encontrado '\n'`.

**Solução proposta:**

- O parser JSX atual consome `=` e tenta ler expressão, mas o ternário `? :`
  confunde o fechamento `}`.
- Solução: tornar o lexer/parser de atributo JSX ciente de ternários aninhados,
  ou parsear a expressão inteira dentro de `{...}` até balancear chaves.

**Arquivos:** `parser/parser.go` (parseJSXAtributo).

#### 5. `tente/capture` capturar panics Go nativos

**Sintoma atual:** `/robots.txt` e `/sitemap.xml` retornam 500 quando `ler()`
falha. O `tente { var conteudo = ler("./robots.txt") ... } capture { res.enviar(fallback) }`
não captura o erro nativo do Go.

**Solução proposta:**

- A VM tree-walk e a VM bytecode devem envolver chamadas a funções builtin em
  `defer recover()` que converta o panic em `RuntimeErro` Harpia.
- Arquivos prováveis: `vm/vm.go`, `vm/interpretador.go`, `builtins/`.

**Arquivos:** `vm/`, `builtins/arquivos.go`.

#### 6. Import curto para stdlib (`de "sinais"`)

**Sintoma atual:** `de "sinais" importe sinal, efeito` falha com
`Importações não relativas só estão disponíveis para módulos embutidos`.

**Solução proposta:**

- O resolver deve tentar primeiro como builtin registrado; se não achar, como caminho
  relativo; se não achar, erro.
- Hoje `sinais` não é builtin. Decisão de design: registrar módulos reativos do
  runtime-web como builtin virtual (apenas transpilação, sem runtime server-side).

**Arquivos:** `parser/resolver.go` (ou similar), `builtins/registro.go`.

#### 7. `<style>` aceitar concatenação `{"a" + "b"}`

**Sintoma atual:** `<style>{"..." + "..."}</style>` falha — quebra na junção.

**Solução proposta:**

- O parser JSX trata filhos `{...}` como expressão única; concatenação de strings
  deveria funcionar mas está falhando na quebra de linha implícita.
- Solução: suprimir nova-linha dentro de chaves JSX filhas quando o conteúdo é
  uma expressão única.

**Arquivos:** `parser/parser.go` (parseJSXFilhos).

#### 8. Single-parameter arrow sem parens

**Sintoma atual:** `x => x + 1` falha com mesmo erro de #1.

**Solução proposta:** Idem #1, apenas uma flag de aceitação.

#### 9. Spread operator `...args`

**Sintoma atual:** `funcao soma(...nums) { ... }` não é suportado.

**Solução proposta:**

- Adicionar `TokenTresPontos` no lexer.
- Adicionar nó AST `Spread{Expressao}`.
- Tratar em: parâmetros de função (último posição), argumentos de chamada,
  elementos de lista literal, pares de mapa.

**Arquivos:** `lexer/`, `parser/`, `vm/`.

#### 10. Decorator syntax `@nome`

**Sintoma atual:** `@depreciado funcao foo() { ... }` não reconhecido.

**Solução proposta:**

- Token `TokenArroba` já existe (usado por media queries).
- Reutilizar ou duplicar para decorator de funções/classes.
- AST: novo nó `Decorator{Nome, Argumentos, Alvo}`.

**Arquivos:** `parser/parser.go`, `ast_nodes.go`.

### Critérios de Priorização

**P0 (Concluído ✅):**

- ✅ #5 (tente/capture captura panics nativos) — capturado via `defer recover()` na VM e interpretador (`hrp/interpretador.go`).
- ✅ #1 + #8 (arrow functions com e sem parênteses) — implementado no parser (`tentarParseArrowFuncao`), interpretador e transpiler, com cobertura de testes completa.
- ✅ #2 (`||` como OR lógico distinto do pipe `|>`) — tokens e precedência separados no lexer e parser.

**P1 (Concluído ✅):**

- ✅ #2 (separar `||` de `|>`).
- ✅ #3 (`<` em strings/código literal sem quebrar parser).
- ✅ #4 (ternário em JSX style/atributos com suporte a quebras de linha).

**P2 (backlog):**

- #6 (imports curtos).
- #7 (concatenação em `<style>`).
- #9 (spread).
- #10 (decorators).

### Estimativa de Esforço Agregado

- P0: ~3-4 dias (parser + VM).
- P1: ~2-3 dias.
- P2: ~4-5 dias (mais testes + edge cases).

### Métricas de Sucesso

- Cobertura de testes do parser: ≥ 90% em todos os nós AST.
- Todos os arquivos `.hrp` do `portal-harpia/` compilam sem warnings.
- Playground consegue executar código Harpia real (não-stub).

---

_Roadmap de melhorias do compilador adicionado em 2026-07-28 após construção do Portal Harpia._

---

_Roadmap atualizado em 2026-07-27. Revisitar e atualizar após conclusão de cada fase._
