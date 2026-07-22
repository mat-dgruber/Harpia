# 🇧🇷 Manual de Referência Oficial do Harpia

Bem-vindo ao **Manual de Referência Oficial do Harpia**, uma especificação exaustiva, de nível de engenharia, que documenta todos os subsistemas, tipos primitivos, biblioteca padrão (stdlib), compilador, analisadores, runtime e filosofia de design da linguagem **Harpia**.

Este d

ocumento foi elaborado para servir tanto como um guia definitivo para desenvolvedores que escrevem programas em Harpia quanto para engenheiros de sistemas que contribuem com a evolução de sua máquina virtual e compilador.

---

## 📖 Índice Geral

1. [Filosofia de Design e Arquitetura](#1-filosofia-de-design-e-arquitetura)
2. [Interface de Linha de Comando (CLI)](#2-interface-de-linha-de-comando-cli)
3. [Análise Léxica (Lexer)](#3-análise-léxica-lexer)
4. [Análise Sintática (Parser &amp; AST)](#4-análise-sintática-parser--ast)
5. [Máquina Virtual e Runtime (hrp)](#5-máquina-virtual-e-runtime-hrp)
6. [Tipos de Dados Primitivos](#6-tipos-de-dados-primitivos)
7. [Biblioteca Padrão (Stdlib)](#7-biblioteca-padrão-stdlib)
8. [Recursos Avançados da Linguagem](#8-recursos-avançados-da-linguagem)
9. [Arcabouço de Testes Nativos (TDD)](#9-arcabouço-de-testes-nativos-tdd)
10. [Diagnósticos e Tratamento de Erros Ricos](#10-diagnósticos-e-tratamento-de-erros-ricos)
11. [Console Interativo (REPL / Playground)](#11-console-interativo-repl--playground)
12. [Guia de Sintaxe Rápida e Exemplos de Produção](#12-guia-de-sintaxe-rápida-e-exemplos-de-produção)
13. [Desenvolvimento Frontend Reativo e SPA](#13-desenvolvimento-frontend-reativo-e-spa-sinais-jsx-estilos-e-ssr)
14. [Segurança e Blindagem Corporativa](#capítulo-14--segurança-e-blindagem-corporativa-security-audit)
15. [Otimizações Avançadas e Desempenho da Máquina Virtual](#capítulo-15--otimizações-avançadas-e-desempenho-da-máquina-virtual)
16. [Novos Módulos Avançados da Biblioteca Padrão (Stdlib)](#capítulo-16--novos-módulos-avançados-da-biblioteca-padrão-stdlib)
17. [Extensões de CLI, DevOps e DevOps DX](#capítulo-17--extensões-de-cli-devops-e-devops-dx)
18. [Pacote de Inteligência Artificial e IA Generativa](#capítulo-18--pacote-de-inteligência-artificial-e-ia-generativa-de-ia)
19. [Conectores de Banco de Dados Corporativos](#capítulo-19--conectores-de-banco-de-dados-corporativos-de-bd)
20. [WebAssembly (WASM), WASI e Microsserviços Corporativos](#capítulo-20--webassembly-wasm-wasi-e-microsserviços-corporativos)
21. [Documentação Ininterrupta e Estilo de Contribuição](#capítulo-21--documentação-ininterrupta-e-estilo-de-contribuição)
22. [Práticas de Segurança e Programação Defensiva](#capítulo-22--práticas-de-segurança-e-programação-defensiva)

---

## 1. Filosofia de Design e Arquitetura

O Harpia foi construído sob uma **perspectiva de design dupla**:

1. **Ponte de Aprendizado:** Facilitar a transição suave de estudantes de programação no Brasil para linguagens de mercado (como JavaScript, Python, Go e C++), utilizando sintaxes modernas (blocos por chaves `{}`, escopo léxico estrito, corotinas assíncronas e tipagem dinâmica opcional) inteiramente em português.
2. **Poder e Identidade Própria:** Ser uma linguagem real de produção. Não é um mero tradutor de códigos ou interpretador didático lento. Oferece reatividade nativa de alto desempenho via **Sinais**, uma biblioteca de socket de baixo nível, suporte a componentes do tipo JSX, um gerador de diagramas de arquitetura e um interpretador robusto com suporte a plugins compartilhados binários em Go ou C/C++ (`.so`).

### Estrutura Geral do Workspace de Diretórios

```
harpia/
├── cmd/               -> Comandos de terminal da CLI (Cobra)
├── compartilhado/     -> Utilitários Unicode (UTF-8) e de casting de strings
├── gramatica/         -> Especificação formal da linguagem (ANTLR4 .g4)
├── lexer/             -> Analisador léxico escrito à mão em Go
├── parser/            -> Analisador sintático de descida recursiva à mão
├── playground/        -> Console interativo de terminal REPL (Liner)
├── hrp/              -> Núcleo do runtime (VM, tipos de dados, escopos)
├── stdlib/            -> Biblioteca padrão da linguagem (embutidos, matematica, etc.)
└── exemplos/          -> Demonstrações práticas e módulos externos
```

---

## 2. Interface de Linha de Comando (CLI)

O utilitário de terminal do Harpia foi construído usando a biblioteca **Cobra** (`github.com/spf13/cobra`). Toda a interface se comunica em português de forma natural.

### Instalação e Atualização Automática do PATH

O Harpia disponibiliza um script automatizado de instalação e build (`instalar.sh`) que compila o CLI atualizado e configura o seu ambiente local de desenvolvimento (Zsh/Bash):

```bash
# Compila e instala o Harpia na pasta ~/.harpia/bin e configura o PATH automaticamente
./instalar.sh
```

O script detecta o shell ativo (`~/.zshrc` ou `~/.bashrc`), adiciona a exportação `export PATH="$HOME/.harpia/bin:$PATH"` se ainda não existir, e substitui o binário anterior compilado sem necessidade de intervenção manual.

### Variáveis Globais de Build (Injeção via Linker)

O pipeline de CI/CD (usando GoReleaser) injeta metadados na compilação do executável do pacote `cmd` através das variáveis:

- `Commit`: Hash SHA-1 curta do commit Git que gerou o build.
- `Datetime`: Carimbo ISO-8601 que registra o instante de build.
- `Version`: Versão SemVer estável da release (ex: `0.3.1`). Se for compilado manualmente, assume o valor `"dev"`.

### Comandos Suportados

#### 1. `harpia` (ou `harpia executar` sem argumentos)

Abre o console de desenvolvimento interativo REPL (Playground) com realce de sintaxe, suporte multilinha e tratamento de sinais. Agora ele conta com ferramentas didáticas integradas como `ajuda <funcao>` (referência direta para todas as primitivas globais como `sinal`, `tamanho`, `tipo`, `sequencia`), `escopo` (visualização em tempo real das variáveis declaradas em memória) e `limpar` (para redefinir o buffer visual do terminal).

#### 2. `harpia executar [arquivo.hrp] [flags]` (Alias: `exec`)

Interpreta e executa um script físico ou código inline de forma síncrona.

- **Ordem de Carregamento**: Se uma string for fornecida pela flag `-c "codigo"`, o interpretador prioriza a execução do arquivo posicional e, em seguida, avalia o fragmento de código inline no mesmo contexto de execução.
- **Execução Zero-Config**: Pode ser invocado diretamente sem argumentos (`harpia executar`). O CLI auto-detecta o script de entrada (`servidor.hrp`, `main.hrp`, etc.) e exibe uma animação de spinner no terminal enquanto valida a sintaxe.
- **Flag `-c`, `--codigo`**: Executa um código direto no terminal (ex: `harpia executar -c "imprima('Olá!')"`).
- **Flag `--assistir`**: Modo Watch / Hot Reload nativo — monitora o arquivo fonte e o recarrega automaticamente ao salvar.
- **Flag `--estrito`**: Ativa a validação estrita de tipos em tempo de execução para anotações de tipo opcionais.

#### 3. `harpia compilar [flags]` (Alias: `compila`)

Transpila ou compila o código-fonte Harpia para alvos específicos (como a Web ou nativo).

- **Execução Zero-Config**: Executar `harpia compilar` sem flags auto-detecta a entrada (`main.hrp`, `index.hrp`, etc.), define a saída padrão como `dist` e o alvo como `web`.
- **Flag `-a`, `--alvo`**: Alvo da compilação. Valores suportados: `web` (padrão, transpila para Virtual DOM e JS puro), `nativo` (AOT via transpilação e build Go nativo), `wasm` (compilação para WebAssembly).
- **Flag `-e`, `--entrada`**: Ponto de entrada/arquivo principal do projeto.
- **Flag `-s`, `--saida`**: Pasta destino onde os arquivos estáticos ou binários serão salvos (padrão: `dist`).

#### 4. `harpia servir [flags]` (Alias: `serve`, `servidor`)

Inicializa o servidor de desenvolvimento local com Hot-Reload automatizado para sua aplicação SPA compilada para a web.

- **Execução Zero-Config**: Executar `harpia servir` sem parâmetros auto-localiza o ponto de entrada do projeto, compila para a pasta `dist` com visualização de spinner animado e levanta o servidor HTTP na porta 3000 (com fallback automático se estiver em uso).
- **Flag `-d`, `--diretorio`**: Diretório raiz de arquivos estáticos a servir (padrão: `dist`).
- **Flag `-p`, `--porta`**: Porta na qual o servidor HTTP escutará as requisições (padrão: `3000`).


#### 5. `harpia novo [command] [flags]` (Alias: `iniciar`, `inicializar`)

Inicializa uma nova estrutura de projeto corporativo pré-configurada seguindo as melhores práticas de Clean Architecture e DDD em Português.

- **Uso:** `harpia novo [backend | frontend | monolito] [nome-do-projeto]`
- **Subcomandos:**
  - `backend`: Estrutura minimalista voltada para microsserviços, APIs lógicas, conectores de banco de dados e concorrência orientada a canais.
  - `frontend`: Estrutura de cliente SPA reativa de alto desempenho baseada no motor de Sinais Reativos do Harpia.
  - `monolito`: Estrutura completa integrando frontend e backend, acompanhada por diretórios de documentação e READMEs explicativos de cada camada.

#### 6. `harpia crie [rota | componente | modelo] [nome]` (Alias: `criar`)

Assistente interativo de scaffolding que gera templates estruturados de arquivos seguindo os padrões de Clean Architecture e DDD definidos para o ecossistema Harpia dentro de um projeto existente.

- **Subcomandos:**
  - `rota`: Cria uma nova página de rota SPA (.hrp).
  - `componente`: Cria um componente de interface (.hrp) e seu correspondente arquivo de estilos dinâmicos (.estilo.hrp).
  - `modelo`: Cria um novo modelo/entidade de dados rico e tipado (.hrp) na camada de domínio.

#### 7. `harpia testar [caminho]`

Varre recursivamente o diretório em busca de arquivos com extensão `.hrp` e executa de forma isolada todos os blocos `testar` nativos definidos nos scripts, apresentando um relatório consolidado com o total de sucessos e falhas.

#### 8. `harpia atualize`

Executa o auto-update do executável a partir do repositório no GitHub.

- **Algoritmo de Resolução**: Monta o caminho de instalação sob o diretório do usuário (`~/.harpia/bin/harpia`). Compara a versão local (executando o binário com `-v`) com a última tag disponível via API do GitHub usando a biblioteca `semver/v3`. Se houver atualizações, usa o `curl` para baixar o binário comprimido adequado para a arquitetura do cliente (mapeando de forma inteligente arquiteturas como `amd64` para `x86_64` e SOs como `darwin` para `Darwin`) e o extrai. Se a versão local for `"dev"`, o processo de atualização automática é impedido para preservar builds de desenvolvimento.

#### 9. `harpia doc [entrada] [flags]`

Varre um diretório ou arquivo extraindo comentários iniciados com três barras (`///`) de funções, classes e métodos, gerando documentação estruturada exportada em formato Markdown (`--formato=markdown`) ou HTML (`--formato=html`).

#### 10. `harpia empacotar --entrada=[arquivo] --saida=[binario] [flags]`

Empacota um script Harpia e todos os seus recursos em um executável binário autônomo (Single Binary Bundle) sem dependências externas compilando dinamicamente o código Go subjacente via `go build` com suporte a cross-compilation (`--so` e `--arq`).

- **Suporte a WebAssembly (WASM)**: Se `--so=js` e `--arq=wasm` forem especificados, o comando compila o interpretador completo para WebAssembly (`docs/portal/harpia.wasm`) e extrai o carregador JavaScript portátil `wasm_exec.js` correspondente do GOROOT do sistema.

#### 11. `harpia diagramar [diretorio] [flags]`

Analisa recursivamente a estrutura física do projeto para mapear e validar a hierarquia de importações entre as camadas do Clean Architecture.

- **Flags**: `--formato` ou `-f` (`mermaid`, `html`, `svg`), `--saida` ou `-s`.
- **Diagrama Interativo**: Se o formato for `html` (ou `svg`), gera um arquivo HTML standalone contendo o visualizador interativo Mermaid.js que colore de verde as importações válidas, de **vermelho grossa as violações arquiteturais**, e emite um botão para exportar diretamente o arquivo `.svg` correspondente.

#### 12. `harpia instalar [nome-do-pacote] [versao-opcional]`

Gerenciador de pacotes e dependências assíncrono para o ecossistema Harpia.

- **Resolução Remota Semver**: Permite baixar pacotes públicos e resolver restrições de versão semver (ex: `banco-dados: 1.0.0`) diretamente de um registro JSON remoto central em português, gravando o módulo na pasta local `pt_modulos/`.

#### 13. `harpia stressar [arquivo] [flags]`

Utilitário CLI interno para benchmarking e testes de estresse concorrentes de aplicações locais ou remotas escritas em Harpia, detalhando estatísticas de tempo médio, mínimo, máximo e taxa de sucesso.

#### 14. `harpia depurar [flags]`

Inicializa o servidor TCP nativo compatível com o protocolo Debug Adapter Protocol (DAP) na porta `4711` (ou customizada via `--porta`), viabilizando a depuração interativa integrada com editores modernos (VS Code).

#### 15. `harpia lsp`

Inicia o servidor oficial LSP (Language Server Protocol) do Harpia via stdio, oferecendo suporte nativo para editores de código (como o VS Code) com autocomplete, hover lendo comentários de três barras (`///`), linter de arquitetura limpa e formatação automática de código ao salvar.

#### 16. `harpia migrar [subcomando] [flags]` (Alias: `migrations`)

Gerencia migrations SQL com SQLite para evolução do schema de banco. Não requer CGO (usa `glebarez/go-sqlite`).

- Subcomandos:
  - `criar <nome>`: cria arquivo timestamped em `infra/migracoes/AAAA-MM-DD-HHmmss-<nome>.sql` com marcadores `-- +migrar ParaCima` (subida) e `-- +migrar ParaBaixo` (descida).
  - `aplicar`: aplica todas as pendentes em ordem alfabética, dentro de uma transação por arquivo, registrando na tabela `_migracoes (versao, aplicada_em)`.
  - `status`: lista aplicadas e pendentes em formato tabular com caminho absoluto.
  - `reverter [N]`: reverte as últimas N migrations aplicadas (default 1) executando o bloco `ParaBaixo`.
- Flags:
  - `--banco` (default `dados.db`): caminho do arquivo SQLite alvo.

Cada bloco `ParaCima`/`ParaBaixo` é delimitado por comentários `-- +migrar ParaCima` e `-- +migrar ParaBaixo`. SQL entre os marcadores é extraído e executado; o resto do arquivo é ignorado.

#### 17. `harpia pwa [--dir=dist] [flags]`

Gera assets PWA (manifest + service worker) a partir do diretório de saída gerado por `harpia compilar --alvo=web`.

- Gera `manifest.webmanifest` (JSON com `name`, `short_name`, `theme_color`, `background_color`, `icons` 192/512).
- Gera `sw.js` com estratégia cache-first e precache de `/`, `index.html`, `app.js`, `runtime-web.js`, `estilos.css`.
- Patcha `index.html` (idempotente): injeta `<link rel="manifest">`, `<meta name="theme-color">` e, se `--registrar`, o `<script>` de registro do service worker.

Flags: `--dir` (default `dist`), `--nome`, `--curto`, `--cor-fundo`, `--cor-tema`, `--registrar`. Se `--nome`/`--curto` não forem passados, são lidos de `dependencias.json`.

#### 18. `harpia i18n [extrair|novo] [flags]`

Extrai e gerencia catálogos de tradução no formato gettext (`.pot`/`.po`), sem dependências externas.

- `extrair <arquivo|dir>`: varre `.ptst`/`.pt` recursivamente, identifica strings traduzíveis em chamadas `t("...")`, `i18n.texto("...")`, `tr(..."...")` e dedupa por `msgid`, gerando `<dir>/<dominio>.pot` com cabeçalho e referências `#: arquivo:linha`.
- `novo <idioma>`: cria `<idioma>.po` vazio no diretório de catálogos, copiando o cabeçalho do `.pot` quando existir.

Flags: `--dir` (default `traducoes`), `--dominio` (default `harpia`). Atenção: a detecção atual é por regex literal (heurística simples marcada com `ponytail:`); refinar quando o parser expor uma função pública de visita à AST sem custos adicionais de cache.

#### `harpia copiloto` (subcomandos)

Além do autocompletar via IA local (Ollama), o comando `copiloto` expõe dois subcomandos de análise estática textual, sem dependências extras:

- `copiloto revisar <arquivo.ptst>`: detecta funções com mais de 80 linhas, mais de 5 parâmetros, aninhamento > 4, variáveis prefixadas com `_` nunca referenciadas abaixo da declaração, e comentários `TODO`/`FIXME`. Saída PT-BR no formato `[ARQUIVO:linha] tipo → mensagem`, com sumário final.
- `copiloto refatorar <arquivo.ptst>`: para cada função > 80 linhas, mostra `[linha N – linha M] → nome_sugerido` com primeira/última linha do bloco. O nome é sugerido a partir de verbos/substantivos presentes nas primeiras linhas (verbos: validar/calcular/processar/...; substantivos: usuario/pedido/requisicao/...), caindo para `helper_N` como fallback.

---

## 3. Análise Léxica (Lexer)

O pacote `lexer` foi escrito inteiramente à mão em Go. Ele evita o uso de expressões regulares ou geradores automáticos para garantir a máxima velocidade de varredura e o tratamento preciso de strings Unicode multibyte.

### O Desafio de Strings UTF-8 em Go

Em Go, strings são slices de bytes UTF-8. Um único caractere Unicode (acentos ou emojis) pode ocupar entre 1 e 4 bytes. Acessos diretos por índice (ex: `str[i]`) podem quebrar runas ao meio.

- **Solução do Harpia**: O arquivo `compartilhado/strings.go` implementa a função `IndiceBytePorCarater(str string) []int`. Ela varre a string decodificando runas via `utf8.DecodeRuneInString` e pré-calcula uma tabela de mapeamento. Desse modo, o Lexer consegue fazer conversões e fatiamentos de caracteres de forma segura e rápida em tempo constante $O(1)$.
- **Cache Estático Thread-Safe Global**: Para suportar múltiplos interpretadores independentes rodando em paralelo sem colisões, o pacote `compartilhado` adota uma tabela de cache global protegida por um `sync.RWMutex`. Entradas são restritas a tamanhos menores que 4KB para evitar consumo excessivo de heap, e o cache inteiro é reciclado se ultrapassar 2048 registros, prevenindo estouros de memória.

### Estrutura Física de Coordenadas de Tokens

Rastrear coordenadas geográficas exatas é indispensável para gerar excelentes tracebacks de erros:

```go
type PosicaoToken struct {
    Coluna int // Coluna na linha física ativa (base 1)
    Linha  int // Número da linha física no arquivo (base 1)
    Indice int // Offset de byte absoluto desde o início (base 0)
}

type Token struct {
    Tipo   TokenType     // Identificador da categoria (iota)
    Valor  string        // Cadeia de texto literal
    Inicio *PosicaoToken // Ponteiro de início
    Fim    *PosicaoToken // Ponteiro de fechamento
}
```

### Palavras-Chave e Identificadores (Mapeamento em `helpers.go`)

O Lexer varre identificadores textuais e executa uma busca em tabela hash (`tokensIdentificadores`). Se o lexema coincidir com alguma chave reservada, o token genérico `TokenIdentificador` é promovido para a palavra-chave dedicada (ex: `TokenSe`, `TokenRetorne`, `TokenClasse`):

- **Estrutura Condicional e de Fluxo**: `se`, `senao`, `enquanto`, `para`, `em`, `retorne`, `pare`, `continue`, `assincrono`, `aguarde`
- **Definições e Escopos**: `var`, `const`, `func`, `funcao`, `classe`, `estende`, `self`, `estatico`, `enum`, `interface`, `exportar`
- **Módulos**: `importe`, `de`
- **Testes e Garantias**: `testar`, `assegura`
- **Constantes e Operadores**: `Verdadeiro`, `Falso`, `Nulo`, `ou`, `e`, `nao`, `nova`, `??` (coalescência nula), `?.` (encadeamento opcional)
- **Controle de Erros**: `tente`, `capture`, `finalmente`

---

## 4. Análise Sintática (Parser & AST)

O analisador sintático (`parser/parser.go`) é um **Parser de Descida Recursiva Manual** (_Manual Recursive Descent Parser_). Ele consome tokens lineares e monta a **Árvore de Sintaxe Abstrata (AST)**.

### Precedência e Hierarquia de Operadores

A descida recursiva força uma prioridade de resolução estrita. Os operadores são avaliados do nível de menor prioridade (resolvidos por último) até os de maior prioridade (resolvidos primeiro):

| Nível  | Operação / Categoria     | Operadores Relacionados                                |
| :----: | :----------------------- | :----------------------------------------------------- |
| **11** | Encadeamento Funcional   | `\|>` (Pipes)                                          |
| **10** | Disjunção Lógica         | `ou`                                                   |
| **9**  | Conjunção Lógica         | `e`                                                    |
| **8**  | Negação Lógica           | `nao`                                                  |
| **7**  | Comparadores Relacionais | `==`, `!=`, `<`, `<=`, `>`, `>=`, `em`, `instancia de` |
| **6**  | OU Bit a Bit             | `\|` (Bitwise OR)                                      |
| **5**  | XOR Bit a Bit            | `^` (Bitwise XOR)                                      |
| **4**  | E Bit a Bit              | `&` (Bitwise AND)                                      |
| **3**  | Deslocamento de Bits     | `<<`, `>>` (Bitwise Shifts)                            |
| **2**  | Soma e Subtração         | `+`, `-` (Concatenação textual também no `+`)          |
| **1**  | Multiplicação e Divisão  | `*`, `/`, `//` (divisão inteira), `%` (resto/módulo)   |
| **0**  | Sinais e Exponenciação   | `+`, `-`, `~` (unários); `**` (exponenciação)          |

### O Padrão Otimizado `parseEsqLst`

Para evitar redundância de código em todas as regras de operadores binários associativos à esquerda, o parser adota uma assinatura elegante baseada em funções de ordem superior:

```go
func (p *Parser) parseEsqLst(proximo func() (BaseNode, error), proxOp func() (string, bool)) (BaseNode, error) {
    esq, err := proximo()
    if err != nil { return nil, err }
    for {
        op, ok := proxOp()
        if !ok { return esq, nil }
        dir, err := proximo()
        if err != nil { return nil, err }
        esq = &OpBinaria{esq, op, dir}
    }
}
```

### Flexibilidade da Regra de Ponto e Vírgula (`;`)

O Harpia permite omitir o uso de ponto e vírgula. O analisador trata `\n` (quebras de linha) e `EOF` (fim de arquivo) como delimitadores implícitos de instrução. A verificação é unificada em `consome(";")`:

- Se o token corrente for de fato `";"`, consome-o e avança.
- Se for uma nova linha ou término de arquivo, valida a instrução como completa sem reclamar, garantindo um código limpo estilo Python ou Go.

---

## 5. Máquina Virtual e Runtime (hrp)

O pacote `hrp` gerencia a infraestrutura matemática, lógica, as tabelas de símbolos e a execução física da AST.

### Interface Primordial `Objeto`

Toda variável ou estrutura na VM do Harpia satisfaz a interface polimórfica:

```go
type Objeto interface {
    Tipo() *Tipo
}
```

Isso permite que fatias genéricas de estruturas Go (`[]Objeto`) armazenem qualquer tipo dinâmico da linguagem.

### Metaclassing e Ciclo de Vida de Tipos

A representação física de uma classe ou metaclasse na VM é modelada pela struct `Tipo`:

```go
type Tipo struct {
    Nome       string         // Nome visual do tipo (ex: "Booleano")
    Nova       NovaFunc       // Alocador de memória (__nova_instancia__)
    Inicializa InicializaFunc // Construtor (__inicializa__)
    Doc        string         // Bloco documental (Docstring)
    Base       *Tipo          // Classe pai (Herança)
    Mapa       Mapa           // Tabela hash contendo os métodos e constantes
}
```

- **Garantia de Montagem Consequente**: Para assegurar a resolução correta de heranças em tempo de carregamento, cada `Tipo` inicializado em Go é enfileirado na lista global `filaMontagem`. A VM dispara a rotina centralizada `MontaOsTipos()` antes de iniciar o processamento da AST, populando as tabelas e injetando as documentações na propriedade mágica `__doc__` de cada classe.

### Resolução de Métodos Mágicos via Reflexão (Reflection)

O Harpia adota uma convenção estrita de nomenclatura: interfaces Go de protocolos mágicos iniciam com **`I`** e seus métodos com **`M`** (ex: `I__texto__` com `M__texto__()`).

Durante o acesso a atributos (`ObtemAtributoS`), se o atributo solicitado iniciar e terminar com duas sublinhas (método mágico, ex: `__texto__`), a VM usa reflexão de pacotes em Go:

```go
ref := reflect.ValueOf(classe)
m := ref.MethodByName("M" + nome)
if m.IsValid() {
    metodo, _ := NewMetodoProxyDeNativo(nome, m.Interface())
    return metodo
}
```

Isso elimina a necessidade de registrar manualmente todos os métodos mágicos nativos de todas as classes, ligando-os dinamicamente ao interpretador em tempo de execução de forma extremamente limpa.

### Tabelas de Símbolos, Encadeamento e Escopo

As variáveis ativas e constantes são mantidas em estruturas `Escopo`:

- **Escopo Léxico (Lexical Scoping)**: Cada `Escopo` mantém um link de referência para seu escopo pai (`Pai *Escopo`).
- **Algoritmo de Busca (`ObterValor`)**: Busca primeiro na tabela local de símbolos. Se a chave não constar, sobe de forma recursiva investigando o escopo do pai. Se atingir a raiz primordial sem sucesso, verifica o módulo de embutidos antes de lançar o erro controlado `NomeErro` (PSC-0005).
- **Sincronização de Concorrência do Escopo**: Cada escopo de variáveis conta com seu próprio `sync.RWMutex` para sincronização fina de leitura e escrita concorrente. Isso previne colisões de mapas em Go durante execuções paralelas de corotinas em background que acessam variáveis comuns.
- **Locks de Grão Fino em Símbolos**: Símbolos individuais do runtime (`hrp.Simbolo`) contam com bloqueios de mutex específicos (`sync.RWMutex`) para ler e definir seu valor de forma atômica e segura.
- **Cooperação Segura com o Garbage Collector**: A listagem de símbolos pelo Garbage Collector utiliza o método seguro `ObterSimbolosSeguro()`, que gera uma cópia rasa estável da tabela de símbolos do escopo sob um lock de leitura, integrando com o algoritmo de varredura e quebra de ciclos de forma 100% thread-safe.

---

## 6. Tipos de Dados Primitivos

Todos os tipos de dados nativos no Harpia possuem comportamentos específicos sob a VM:

### Iteração Nativa e Protocolos de Coleção (`I_iterador`)

O Harpia oferece suporte a iteração unificada via laços `para-em`:

- **Listas e Tuplas**: Iteração sequencial pelos elementos contidos.
- **Mapas**: Iteração pelos pares `[Chave, Valor]` encapsulados em Tuplas, permitindo desestruturação fluida.
- **Texto (`Texto`)**: Iteração caractere a caractere (Unicode-safe via runas Go), preservando caracteres multibyte como acentos e emojis sem corromper a fatiagem física.
- **Bytes (`Bytes`)**: Iteração byte a byte (retornando inteiros para cada valor físico de byte). As tentativas de coerção direta de `Bytes` para `Inteiro` ou `Decimal` retornam um erro gracioso `NaoImplementadoErro` em português, orientando a conversão prévia para `Texto` ou o uso do método `tamanho()`.

### 1. `Inteiro` (int64)

- **Design**: Inteiro com sinal de 64 bits para evitar estouros aritméticos.
- **Casting**: `int(obj)` avalia o método mágico `__inteiro__`.
- **Coerção Booleana**: Retorna Falso se o valor for zero, e Verdadeiro do contrário.
- **Coerção Decimal**: Promove para `Decimal` quando somado, subtraído ou multiplicado por um membro do tipo `Decimal`.

### 2. `Decimal` (float64)

- **Design**: Número de ponto flutuante de dupla precisão (IEEE 754).
- **Representação Textual**: Se o valor numérico for inteiro (ex: `5.0`), o método `M__texto__()` anexa explicitamente `.0` ao texto para manter no console a distinção visual clara em relação aos Inteiros ordinários.

### 3. `Booleano` (bool)

- **Design**: Armazena as constantes globais estruturadas `Verdadeiro` ou `Falso`.
- **Casting de Inteiros**: `Verdadeiro` é coergido para `1` e `Falso` para `0` quando operado de forma aritmética.

### 4. `Texto` (string)

- **Design**: Cadeia imutável de caracteres UTF-8.
- **Comprimento Seguro**: O método `tamanho()` chama `utf8.RuneCountInString`, fornecendo a contagem de caracteres reais em vez de contagem de bytes físicos em disco.
- **Comparações Lexicográficas**: Suporta comparações relacionais ricas diretamente entre instâncias de Texto (`<`, `<=`, `>`, `>=`), comparando a ordem lexicográfica Unicode das strings.
- **Interpolação de Strings**: Implementada usando o operador de módulo `%`. Analisa e substitui de forma dinâmica marcadores de formatação:
  - `%i`: Formata para Inteiro.
  - `%d`: Formata para Decimal.
  - `%b`: Formata para Booleano.
  - `%s` (ou outro marcador): Formata chamando a representação textual do objeto.
- **Exemplo**: `"Eu tenho %i anos de idade e me chamo %s" % (23, "Carlos")`.


### 5. `Lista` (`[]Objeto` mutável)

- **Design**: Coleção indexada mutável ordenada de dados.
- **Métodos Embutidos**:
  - `adiciona(elemento)`: Insere um novo item no final.
  - `extende(outraLista)`: Concatena os elementos de outra coleção.
  - `remove(elemento)`: Busca, remove e retorna o elemento especificado.
  - `pop(indice?)`: Remove e retorna o item localizado no índice. Se omitido, assume o índice inicial `0`.
  - `indice(elemento)`: Retorna o índice da primeira ocorrência do item.
  - `limpa()`: Esvazia por completo a lista.

### 6. `Tupla` (`[]Objeto` imutável)

- **Design**: Coleção indexada ordenada e imutável de dados.
- **Imutabilidade**: Não fornece ou expõe métodos para mutabilidade ou alteração física após ser criada no script.

### 7. `Mapa` (`map[string]Objeto`)

- **Design**: Dicionário associativo do tipo chave-valor. As chaves são estritamente do tipo `Texto`.
- **Métodos Embutidos**:
  - `chaves()`: Retorna uma tupla imutável com todas as chaves registradas.
  - `valores()`: Retorna uma tupla contendo todos os valores dos objetos.
  - `atualizar(outroMapa, ignoreExistentes?)`: Copia e mescla os dados de outro mapa de forma mutável. Se `ignoreExistentes` for Verdadeiro, chaves repetidas não são sobrescritas.
- **Mecânica de Iteração**: O loop `para` sobre mapas retorna de forma consecutiva uma `Tupla` contendo o par `(chave, valor)`, simplificando a varredura e permitindo desestruturação fluida.

### 8. `Bytes` (`[]byte` mutável)

- **Design**: Array físico de bytes para controle de rede ou buffers de arquivo.
- **Comparações**: Permite comparações ricas (`==`, `!=`, `<=`, etc.) baseadas na contagem de bytes e no conteúdo literal através de `bytes.Equal` do Go.

### 9. `Nulo`

- **Design**: Representa a ausência física de valor. É do tipo de classe única `_Nulo`.

---

## 7. Biblioteca Padrão (Stdlib)

Os módulos nativos são desenvolvidos de forma desacoplada em Go. Eles se registram via funções `init()` acionadas por importações anônimas no arquivo agregador `stdlib/stdlib.go`.

### Módulo: `embutidos`

Símbolos e métodos injetados de forma global. Não requerem importação.

- `escreva(args...)` (Alias: `imprimir`): Concatena os argumentos textuais separando-os por espaço e exibe a mensagem na saída padrão.
- `leia(prompt?)`: Exibe o prompt textual se fornecido e pausa a VM aguardando digitação pelo usuário. Retorna sempre uma string.
- `tamanho(objeto)`: Retorna a contagem de elementos de coleções que implementam a interface `I__tamanho__`.
- `int(objeto)` / `texto(objeto)`: Construtores de coerção.
- `instanciaDe(obj, classes)`: Verifica se o objeto descende das classes informadas.
- `mesmoTipo(obj1, obj2)`: Compara as assinaturas de classe dos objetos.
- `tipo(obj)`: Retorna a representação de Tipo da classe do objeto.
- `doc(obj)`: Devolve o bloco explicativo de documentação (Docstring) do método ou classe.
- `sequencia(fim)` / `sequencia(inicio, fim, passo?)`: Retorna uma struct `SequenciaNumerica` que atua como um iterador numérico sob limites definidos, lançando `FimIteracao` ao término.

### Módulo: `matematica`

Recursos matemáticos de alta precisão. Requer `importar matematica`.

- **Constantes**: `matematica.PI`, `matematica.E`.
- **Métodos**:
  - `absoluto(n)`: Magnitude numérica sem sinal.
  - `piso(n)` / `teto(n)`: Arredondamento para baixo/cima.
  - `potencia(base, expoente)`: Calcula $base^{expoente}$.
  - `raiz(radicando, indice?)`: Calcula a raiz do número. Se o índice for omitido, calcula a raiz quadrada por potência fracionária de expoente ($radicando^{1.0/indice}$).

### Módulo: `sistema`

Acesso ao hardware e variáveis de ambiente. Requer `importar sistema`.

- `sistema.NOME`: String identificando o SO hospedeiro (`"darwin"`, `"linux"`, `"windows"`).
- `sistema.ARQUITETURA`: Tipo de arquitetura do processador (`"amd64"`, `"arm64"`).

### Módulo: `colorize`

Colorização de console com ANSI True Color de 24 bits. Requer `importar colorize`.

- **Objetos**: `colorize.TEXTO` (Foreground), `colorize.FUNDO` (Background).
- **Propriedades**: `colorize.SUPORTA` (Booleano dinâmico que detecta variáveis de escape como `NO_COLOR`).
- **Métodos**:
  - `converteRGB(r, g, b, background?)`: Retorna o código de escape ANSI correspondente.
  - `imprimac(args...)`: Imprime os argumentos com as cores aplicadas. Se o console não suportar cores, remove de forma limpa as sequências ANSI via expressão regular antes de imprimir.
- **Cores Mapeadas**: `vermelho`, `lima`, `azul`, `amarelo`, `agua`, `fuchsia`, `branco`, `preto` (disponíveis tanto em `TEXTO` quanto em `FUNDO`).
- **Exemplo**: `imprimac(colorize.TEXTO.azul(colorize.FUNDO.branco("Texto Colorido!")))`.

### Módulo: `arquivos`

Recursos e controle de sistema de arquivos e caminhos. Requer `de "arquivos" importe ...`.

- **Métodos**:
  - `ler(caminho)`: Lê o conteúdo de um arquivo em formato de texto.
  - `escrever(caminho, texto)`: Cria ou sobrescreve um arquivo gravando o texto especificado.
  - `acrescentar(caminho, texto)`: Adiciona o texto especificado ao final do arquivo.
  - `remover(caminho)`: Exclui o arquivo ou diretório especificado.
  - `renomear(origem, destino)`: Move ou altera o nome de um arquivo ou pasta.
  - `juntar(partes...)`: Concatena partes de caminhos físicos de arquivos de acordo com o SO.
  - `resolver(caminho)`: Devolve o caminho absoluto absoluto limpo.
  - `caminhar(caminho, callback)`: Varre recursivamente diretórios acionando a função de callback fornecida.

### Módulo: `json`

Serialização e desserialização de formato de dados JSON. Requer `de "json" importe ...`.

- **Métodos**:
  - `analisar(textoJson)`: Desserializa uma string JSON em estruturas nativas de dados do Harpia (Lista, Mapa, Inteiro, Decimal, Booleano, Nulo).
  - `serializar(objeto)`: Converte estruturas de dados recursivas do Harpia em string JSON representativa. Oferece suporte completo e nativo a instâncias de classes personalizadas (reunindo recursivamente seus atributos).

### Módulo: `yaml`

Serialização e desserialização de formato de dados YAML. Requer `de "yaml" importe ...`.

- **Métodos**:
  - `analisar(textoYaml)`: Desserializa uma string YAML em estruturas nativas de dados do Harpia.
  - `serializar(objeto)`: Converte estruturas de dados do Harpia em string YAML.

### Módulo: `xml`

Serialização e desserialização de formato de dados XML. Requer `de "xml" importe ...`.

- **Métodos**:
  - `analisar(textoXml)`: Desserializa uma string XML em estruturas nativas de dados do Harpia.
  - `serializar(mapa, tagRaiz?)`: Converte um Mapa do Harpia em string XML com a tag raiz opcional informada (padrão: "raiz").

### Módulo: `cripto`

Funções para criptografia, hashes e identificadores. Requer `de "cripto" importe ...`.

- **Métodos**:
  - `sha256(texto)`: Devolve o hash SHA-256 do texto fornecido em formato hexadecimal.
  - `codificarBase64(texto)`: Codifica um texto simples para o formato Base64.
  - `decodificarBase64(base64)`: Decodifica um texto de Base64 para o formato simples correspondente.
  - `uuid()`: Gera e retorna um identificador universal único (UUID v4) aleatório.

### Módulo: `http`

Protocolo de rede HTTP (Cliente e Servidor). Requer `de "http" importe ...`.

- **Classes**:
  - **`Servidor`**:
    - `obter(rota, handler)`: Registra um manipulador (handler) para requisições de método GET na rota. Aceita rotas dinâmicas com parâmetros nomeados, como `/ola/:nome`.
    - `postar(rota, handler)`: Registra um manipulador para o método POST na rota especificada.
    - `deletar(rota, handler)`: Registra um manipulador para o método DELETE na rota especificada.
    - `usar(middleware)`: Registra um middleware global (função `funcao(req, res)`) executado sequencialmente antes do handler de destino de cada requisição.
    - `escutar(porta, bloquear = Falso)`: Inicia a escuta e aceitação de requisições na porta informada. Se o segundo argumento for `Verdadeiro`, executa de forma síncrona bloqueando a thread principal. Do contrário, opera de forma assíncrona em background.
    - `fechar()`: Encerra a escuta do servidor HTTP liberando a porta local de forma limpa.
  - **`Requisicao`**:
    - Representa os metadados da requisição HTTP recebida. Atributos:
      - `metodo`: String que descreve o método HTTP usado (ex: `"GET"`, `"POST"`).
      - `caminho`: String contendo o caminho da rota requisitada (ex: `"/ola/harpia"`).
      - `cabecalho`: Mapa contendo os cabeçalhos recebidos.
      - `corpo`: Texto do corpo da mensagem HTTP.
      - `parametros`: Mapa dinâmico contendo as variáveis injetadas por rotas dinâmicas (ex: `req.parametros["nome"]` para a rota `/ola/:nome`).
      - `query`: Mapa contendo as variáveis de consulta (query string) passadas na URL (ex: `req.query["agora"]` para a URL `/?agora=2026-07-21`).
      - `corpoJson`: Objeto (Dicionário/Mapa) contendo o corpo da requisição automaticamente desserializado a partir de dados em formato JSON (nulo se não for um JSON válido).
  - **`Resposta`**:
    - Representa a resposta HTTP a ser enviada pelo servidor. Atributos:
      - `status`: Inteiro indicando o código de status HTTP (ex: `200`, `404`, `500`).
      - `corpo`: Texto a ser retornado no corpo da resposta.
      - `cabecalho`: Mapa com os cabeçalhos de resposta.
    - Métodos:
      - `definir_cabecalho(chave, valor)`: Define um cabeçalho customizado na resposta.
- **Funções**:
  - `requisitar(metodo, url, corpo?, cabecalhos?)`: Realiza uma chamada de requisição HTTP Cliente síncrona completa (suporta chamadas HTTPS) e retorna o respectivo objeto de `Resposta`.

- **Suporte Nativo a CORS**:
  - O servidor HTTP do Harpia possui suporte transparente e automático a CORS para facilitar o desenvolvimento de SPAs de frontend que consomem APIs localmente.
  - **Requisições OPTIONS (Preflight)**: São interceptadas e respondidas automaticamente com status `200 OK` e cabeçalhos de CORS padrão (`Access-Control-Allow-Origin: *`, `Access-Control-Allow-Methods: GET, POST, DELETE, OPTIONS, PUT, PATCH`, etc.).
  - **Customização**: Se o desenvolvedor desejar configurações específicas de CORS, ele pode simplesmente definir as chaves desejadas no mapa de cabeçalhos da resposta (`res.cabecalho` ou chamando `definir_cabecalho`). Estas definições têm precedência e substituem os cabeçalhos padrão:
    ```harpia
    app.obter("/api/dados", funcao(req, res) {
        res.cabecalho["Access-Control-Allow-Origin"] = "https://meu-site-confiavel.com"
        res.enviarJson({"dados": "confidenciais"})
    })
    ```

### Módulo: `bd`

Acesso e manipulação de bancos de dados relacionais e não-relacionais. Requer `de "bd" importe ...`.

- **Funções de Conexão**:
  - `conectarSqlite(caminho)`: Abre uma conexão SQLite pura em Go, retornando um objeto `ConexaoSQL`.
  - `conectarPostgres(url)`: Abre uma conexão PostgreSQL, retornando um objeto `ConexaoSQL`.
  - `conectarMysql(url)`: Abre uma conexão MySQL, retornando um objeto `ConexaoSQL`.
  - `conectarMongo(url)`: Abre uma conexão MongoDB, retornando um objeto `ConexaoMongo`.
  - `conectarRedis(url)`: Abre uma conexão Redis, retornando um objeto `ConexaoRedis`.
- **A Classe `ConexaoSQL`**:
  - `executar(sql, args...)`: Executa comandos SQL de mutação ou DDL (INSERT, UPDATE, DELETE, CREATE).
  - `consultar(sql, args...)`: Executa consultas SQL SELECT, retornando uma `Lista` de `Mapa`s.
  - `tabela(nome)`: Retorna uma instância de `QueryBuilder` ligada a essa conexão.
  - `fechar()`: Fecha a conexão.
- **A Classe `QueryBuilder`**:
  - `selecionar(colunas...)`: Define as colunas a serem selecionadas.
  - `onde(coluna, operador, valor)`: Adiciona uma cláusula de filtro.
  - `limite(n)`: Limita o número de registros.
  - `obterMuitos()`: Executa e retorna todos os registros correspondentes.
  - `obterUm()`: Executa e retorna o primeiro registro ou Nulo.
  - `inserir(mapaValores)`: Insere um novo registro com o Mapa fornecido.
  - `atualizar(mapaValores)`: Atualiza os registros filtrados com o Mapa de modificações.
  - `deletar()`: Remove os registros que coincidem com os filtros aplicados.
- **A Classe `ConexaoMongo`**:
  - `colecao(nome)`: Retorna uma coleção do MongoDB.
- **A Classe `ConexaoRedis`**:
  - `definir(chave, valor, expiracaoSegundos?)`: Define um valor para a chave.
  - `obter(chave)`: Obtém o valor da chave ou Nulo.
  - `remover(chave)`: Remove a chave.

### Módulo: `soquete`

Controle de sockets de baixo nível (TCP/IP). Requer `importar soquete`.

- **Constantes**: `AF_INET` (IPv4), `AF_INET6` (IPv6), `SOCK_STREAM` (TCP), `SOCK_DGRAM` (UDP).
- **A Classe `Soquete`**:
  - `nova Soquete(familia, tipo)`: Cria o socket chamando as APIs de syscall correspondentes do kernel do SO.
  - `associa(ip, porta)`: Vincula a conexão de rede local (Bind).
  - `ouve(backlog?)`: Ativa escuta de conexões com backlog de fila padrão de 1 se omitido.
  - `aceita()`: Aguarda de forma não-bloqueante (usando `unix.Poll` para gerenciar eventos do File Descriptor) e aceita conexões de clientes, retornando um novo objeto `Soquete` para a troca de dados.
  - `conecta(endereco, porta)`: Conecta o socket cliente ao servidor de destino (resolve o host dinamicamente de DNS via `net.LookupIP`).
  - `envia(bytes)`: Escreve e envia os bytes correspondentes (tipo `Bytes`).
  - `recebe(tamanho)`: Lê os dados disponíveis de rede até o limite do buffer e os devolve envelopados em um objeto `Bytes`.
  - `def_nao_bloqueante(booleano)`: Altera propriedades de espera de E/S do socket.
  - `define_opcoes(nivel, opcao, valor)`: Configura opções de soquete (SetsockoptInt).
  - `fecha()`: Encerra conexões e libera o File Descriptor do SO de forma segura.

### Módulo: `ia`

Integração nativa com inteligência artificial e primitivas de agentes autônomos. Requer `de "ia" importe ...`.

- **Classes**:
  - **`Agente`**:
    - `nova Agente(nome, instrucoes, provedor?, modelo?)`: Instancia um agente autônomo. O provedor padrão é `"ollama"` e o modelo padrão é `"llama3"`.
    - **Atributos**:
      - `nome`: Nome do agente.
      - `instrucoes`: System prompt com diretrizes de comportamento do agente.
      - `provedor`: Provedor configurado (`"ollama"`, `"gemini"`, `"openai"`).
      - `modelo`: Identificador do modelo de linguagem (ex: `"llama3"`, `"gemini-1.5-flash"`, `"gpt-4o-mini"`).
      - `historico`: Lista contendo a memória/histórico de mensagens trocadas (`"role"`, `"content"`).
    - **Métodos**:
      - `perguntar(mensagem)`: Envia uma pergunta ao agente, anexando o histórico de conversas anterior, e retorna a resposta de texto gerada.
      - `limpar_memoria()`: Apaga completamente o histórico de mensagens salvas.
      - `comunicar(outro_agente, mensagem)`: Envia uma instrução a outro agente, aguarda sua resposta e a registra na própria memória do agente chamador de forma orquestrada (suporte nativo multi-agente).

---

## 8. Recursos Avançados da Linguagem

O Harpia possui recursos modernos e engenhosos integrados nativamente em sua especificação gramatical e de runtime:

### 1. O Operador Pipe (`|>`)

Permite encadear transformações e chamadas consecutivas de dados de forma altamente legível.

- **Sintaxe**: `valor |> funcao_ou_metodo` ou `valor |> funcao(argumentoExtra)`.
- **Algoritmo de Injeção**:
  - Se o membro da direita for um identificador de função simples (ex: `texto |> maiusculo`), a VM avalia e executa a chamada simples passando o operando esquerdo como o único argumento: `maiusculo(texto)`.
  - Se o membro da direita for uma chamada parametrizada contendo argumentos extras (ex: `10 |> somar(5)`), o interpretador intercepta a chamada sintática, realiza o append do operando esquerdo na primeira posição da lista de argumentos e executa: `somar(10, 5)`.
- **Prevenção de Efeitos Colaterais**: O operando esquerdo é avaliado uma única vez de forma garantida antes da injeção, prevenindo execuções duplicadas e vazamento de estados (conforme validado nos testes em `pipe_test.go`).

### 1.1 Interpolação de Strings (Templates e Chaves `{}`)

Adicionada no **Sprint 8**, permite embutir expressões lógicas de Harpia diretamente em strings textuais (`TemplateLiteral`) e componentes delimitados por chaves `{ ... }`.

- **Sintaxe**: `"Olá, { nome }!"` ou `"Dobro: { valor |> duplicar }"`
- **Mecânica de Parsing**: O analisador sintático intercepta strings literais do tipo `lexer.TokenTexto` no parser (`parseAtomo`). Se detectar o padrão de chaves `{ ... }`, o parser segmenta a string em partes literais e expressões dinâmicas (`TemplateExpr`), parseando recursivamente com instâncias isoladas de Parser.
- **Operador Pipe em Interpolações**: O operador pipe `|>` pode ser empregado livremente dentro de chaves em interpolações para transformar dados de forma fluida (ex: `"Nome: { usuario.nome |> maiusculas }"`). No tempo de execução, os visitors da VM resolvem as sub-expressões dinâmicas e as concatenam em uma única string unificada do tipo `Texto`.

### 2. Parâmetros de Funções Avançados (Defaults & Nomeados)

As funções aceitam declarações de valores padrão e chamadas referenciando parâmetros nominalmente (em qualquer ordem de envio).

- **Parâmetros com Default**: `func calcular(a, b = 2) { retorne a * b }`.
- **Chamadas Nomeadas**: `calcular(b = 10, a = 5)`.
- **Mecânica de Resolução**: Na chamada de uma função, os argumentos posicionais são mapeados de forma ordenada. Argumentos nomeados são extraídos e armazenados na struct interna `ArgumentoNomeadoObj`. O método de chamada (`Funcao.M__chame__`) varre a lista de parâmetros formais esperados: preenche com os valores nomeados, avalia e injeta as expressões padrão caso falte algum parâmetro, e gera erro se um argumento obrigatório for omitido.

### 2.1. Anotações de Tipo Opcionais e Validação Estrita (`--estrito`)

Parâmetros, retorno de função e variáveis podem receber anotações de tipo estáticas:

```harpia
var idade: Inteiro = 18
const PI: Decimal = 3.14

funcao soma(a: Inteiro, b: Inteiro = 0): Inteiro {
    retorne a + b
}
```

- **Os tipos ficam registrados na AST** (`DeclVar.Tipo`, `DeclFuncaoParametro.Tipo`, `DeclFuncao.TipoRetorno`) e são validados ativamente se a flag `--estrito` estiver presente.
- **Validação em tempo de execução**: Ao executar o script com a flag `--estrito` (`harpia executar arquivo.hrp --estrito`), a VM valida se o valor atribuído a uma variável ou o retorno/parâmetros de uma chamada de função são compatíveis com os tipos anotados. Violações lançam erro do tipo `TipagemErro` (PSC-0004).
- **Tipos suportados**:
  - Primitivos: `Inteiro`, `Decimal`, `Texto`, `Booleano` (ou `Logico`), `Nulo`
  - Compostos: `Lista<T>`, `Mapa<C, V>`, `Tupla` (com validação profunda recursiva de elementos)
  - Assinaturas: `funcao` (ou `Funcao`) para qualquer objeto chamável.

### 2.2. Linter Estático — `harpia checar`

Comando que varre diretórios recursivamente em busca de arquivos `.hrp`/`.hrp` e os analisa sem executar.

```bash
$ harpia checar ./src --formato=json --estrito
```

#### Flags Suportadas

- `--formato`: Define o formato de saída do relatório.
  - `texto` (Padrão): Saída formatada agrupada por arquivo com sumário.
  - `json`: Emite diagnósticos formatados de acordo com o padrão `Diagnostic` de LSP (Language Server Protocol) em português, com as posições espaciais precisas de cada erro (linha, coluna, tamanho do token).
- `--estrito`: Ativa a verificação de compatibilidade de tipos estáticos anotados.

Verifica:

- **Redeclaração de nomes** no mesmo escopo local (Erro crítico, Severidade LSP 1).
- **Shadowing de nomes** em escopos filhos (Aviso educativo, Severidade LSP 2) - gera apenas um alerta amigável de sombreamento, mantendo o processo de build ativo e funcional.
- **Reatribuição de `const`** (preservando a regra de imutabilidade).
- **Identificadores não declarados** (com fallback para a stdlib via tabela `globalsLinter`).
- **Parâmetros duplicados** na mesma assinatura de função.

Para detalhes de implementação, ver `cmd/checar.go`.

### 3. Tratamento de Exceções Estruturado (`tente / capture / finalmente`)

Fluxo clássico de tratamento de erros nativo em português:

```harpia
tente {
    var resultado = 10 / 0
} capture (erro) {
    escreva("Erro capturado: " + erro.mensagem)
} finalmente {
    escreva("Sempre executa!")
}
```

- **Mecânica de Escopo**: O bloco `capture` cria um escopo léxico filho temporário e expõe o erro capturado sob a variável especificada em parênteses. O erro é uma instância rica contendo propriedades como `mensagem`, `linha`, `coluna` e `arquivo`. Essas coordenadas são injetadas automaticamente a partir do Contexto da VM, então `erro.arquivo` e `erro.linha` funcionam como inspetores de traceback.
- **Garantia do Finalmente**: O bloco `finalmente` é protegido com mecanismos de `defer` no runtime do Go. Ele é executado de forma garantida, mesmo que ocorram erros dentro do bloco de captura ou que exceções não tratadas se propaguem subindo na pilha de execução.
- **Sobrescrita do Erro Original pelo `finalmente`**: Se o bloco `finalmente` lançar uma exceção, essa exceção substitui o erro original (semântica Python/Java), refletindo em tracebacks. Use `tente { ... } capture (erro) { ... } finalmente { ... }` com cuidado ao manipular recursos propensos a falhar.
- **Bloco `finalmente` Opcional**: Apenas `tente { ... } capture (erro) { ... }` (sem `finalmente`) e `tente { ... } finalmente { ... }` (propagação sem captura — exigem reabertura para tratar) também são sintaxes válidas.

### 4. Plugins e Extensões Dinâmicas Go (`.so`)

Permite carregar dinamicamente bibliotecas compiladas na linguagem Go como extensões do interpretador, operando de forma nativa e rápida.

- **Orquestração de Carregamento**: Se um arquivo com extensão `.so` for importado, a VM usa o pacote `plugin` do Go para abrir a biblioteca, localiza via reflexão o símbolo público da função `InicializaModulo()`, executa-a e carrega seu respectivo escopo estruturado `ModuloImpl`.
- **Compilação**: `go build -buildmode=plugin -o modulo.so modulo.go`. (Suportado nativamente em Linux e macOS).

---

## 9. Arcabouço de Testes Nativos (TDD)

O Harpia estimula a escrita de testes de qualidade integrando as asserções e as suítes diretamente na sintaxe da linguagem.

### A Palavra-Chave `testar`

Permite declarar um bloco de teste nomeado no próprio script:

```harpia
testar "deve somar dois numeros corretamente" {
    assegura(soma(2, 2) == 4, "A soma deve ser quatro!")
}
```

- **Isolamento de Estado**: Ao rodar a suíte de testes (`harpia testar`), o compilador cria um escopo temporário para cada bloco `testar` que herda as variáveis, constantes e importações globais do arquivo original, mas previne colisões e vazamentos de estado de um teste para o outro.

### A Diretiva `assegura` (ou `assegure`)

Atua como a asserção padrão do TDD. Recebe uma expressão de verificação e uma mensagem textual opcional de erro:

```harpia
assegura condicao, "Mensagem caso falhe";
```

- Se a expressão lógica resultar em `Falso` (ou nulo/zero), lança a exceção estruturada `ErroDeAsseguracao` (PSC-0011).

---

## 10. Diagnósticos e Tratamento de Erros Ricos

Um dos recursos mais inovadores do Harpia é o seu sistema visual de diagnósticos educativos voltados ao ensino de programação.

### Estrutura Geral do Objeto `Erro`

```go
type Erro struct {
    Base     *Tipo        // Classe específica da exceção (ex: NomeErro)
    Contexto *Contexto    // Ponteiro ao supervisor global da VM
    Mensagem Objeto       // Texto descritivo
    Linha    int          // Linha do erro (base 0)
    Coluna   int          // Coluna do erro (base 1)
    Token    *lexer.Token // Token causador
    Arquivo  string       // Nome do arquivo de origem
    Codigo   string       // Código-fonte para renderizar o traceback
    Sugestao string       // Sugestão explicativa contextual
}
```

### Relação de Códigos Normatizados de Erros

Para facilitar pesquisas em fóruns e documentações de suporte, cada erro é associado a um código normatizado único:

|    Código    | Classe de Erro             | Causa Típica do Bug                                              |
| :----------: | :------------------------- | :--------------------------------------------------------------- |
| **PSC-0001** | `SintaxeErro`              | Violação de regras gramaticais e estruturas sintáticas.          |
| **PSC-0002** | `ReatribuicaoErro`         | Tentativa ilegal de redeclarar ou alterar constantes.            |
| **PSC-0003** | `AtributoErro`             | Acesso a propriedades ou métodos não existentes na instância.    |
| **PSC-0004** | `TipagemErro`              | Operandos de tipos incompatíveis com a operação solicitada.      |
| **PSC-0005** | `NomeErro`                 | Variável ou identificador não definido ou encontrado no escopo.  |
| **PSC-0006** | `ImportacaoErro`           | Falha ao localizar arquivos ou carregar módulos e símbolos.      |
| **PSC-0007** | `ValorErro`                | Argumento de tipo correto, mas valor inadequado.                 |
| **PSC-0008** | `ErroDeLimite`             | Valor numérico fora dos limites aceitáveis pela VM.              |
| **PSC-0009** | `IndiceErro`               | Indexação de sequências fora dos limites de tamanho.             |
| **PSC-0010** | `RuntimeErro`              | Falhas genéricas no ambiente de execução da VM.                  |
| **PSC-0011** | `ErroDeAsseguracao`        | Falha na validação lógica de uma asserção de teste (`assegura`). |
| **PSC-0012** | `DivisaoPorZeroErro`       | Tentativa matemática proibida de divisão por zero.               |
| **PSC-0013** | `ErroDeSistema`            | Falhas de chamadas e comandos de E/S do sistema operacional.     |
| **PSC-0014** | `ArquivoNaoEncontradoErro` | Tentativa de abrir ou acessar um caminho físico inexistente.     |

### Sugestões Educativas Inteligentes e Heurística de Digitação

O interpretador analisa heuristicamente as palavras causadoras do erro no momento de formatar a saída. Se o programador tiver cometido erros de digitação comuns, o compilador fornece a correção amigável em português:

- Se `NomeErro` for disparado e o lexema de falha for `"imrpimir"` ou `"imprimi"`, sugere: `Você quis dizer 'imprimir'?`.
- Se `SintaxeErro` for disparado com a palavra `"retornar"`, sugere: `Em Harpia, use a palavra-chave 'retorne' para retornar valores.`.
- Se `DivisaoPorZeroErro` for disparado, sugere: `Não é possível dividir um número por zero.`.

### O Renderizador de Traceback com Cores ANSI

O método `Error()` da struct `Erro` implementa um dos melhores formatadores visuais de console existentes:

1. **Deteção de Cores**: Se a variável de ambiente `NO_COLOR` não estiver definida, ativa estilizações cromáticas ANSI True Color de alto contraste (Vermelho para erros, Ciano para ponteiros de guia, Verde para sugestões didáticas).
2. **Desenho de Setas e Sublinhado**: Imprime a caixa identificadora do código (ex: `erro[PSC-0005]`), localiza as coordenadas `arquivo:linha:coluna`, abre o arquivo físico para extrair a linha culpada, desenha uma seta de traceback (`┌──>`) e **sublinha graficamente com acentos circunflexos na largura física exata (`^^^^`)** o token responsável pelo erro, facilitando imensamente a correção visual direta no console!

### Integração com IA Local (Ollama) para Explicação de Erros

A partir do fechamento da **Fase 1**, o comando `harpia erro` conta com o subcomando `explicar` para fornecer ajuda inteligente utilizando inteligência artificial local:

```bash
$ harpia erro explicar PSC-0005
```

- **Fluxo de Integração**: O comando realiza uma conexão HTTP local segura com a instância do Ollama (`127.0.0.1:11434/api/generate`) requisitando ao modelo `gemma` uma explicação didática do erro.
- **Fallback e DX Amigável**: Se o Ollama não estiver instalado ou ativo, o CLI detecta a ausência de conexão imediatamente e fornece um tutorial passo a passo em português ensinando como baixar e iniciar o Ollama, procedendo então a renderizar a explicação pedagógica estática catalogada do próprio dicionário local da linguagem, para que o desenvolvedor nunca fique sem auxílio.

---

## 11. Console Interativo (REPL / Playground)

O playground do Harpia é acessado digitando apenas `harpia` no terminal. Ele utiliza a biblioteca **Liner** para gerenciar entradas e manter um histórico persistente e inteligente.

### Máquina de Estados e Prompt Multilinha (Controle em `estado.go`)

Para suportar códigos multilinha, o REPL intercepta as linhas digitadas e avalia o pareamento dos delimitadores:

```go
strings.Count(codigo, "[") > strings.Count(codigo, "]") ||
strings.Count(codigo, "(") > strings.Count(codigo, ")") ||
strings.Count(codigo, "{") > strings.Count(codigo, "}")
```

- **Estado Normal (`>>> `)**: Ativo quando não há delimitadores abertos. Ao apertar Enter, o código acumulado é imediatamente enviado para a VM compilar e executar.
- **Estado Contínuo (`... `)**: Se houver delimitadores não pareados (ex: chaves abertas de uma função), o REPL não tenta executar e muda o prompt visual para `... `, indicando que a instrução lógica continua na próxima linha física do console.

### Persistência de Histórico de Comandos em Disco

O histórico de comandos não é perdido ao fechar a sessão. Na inicialização do playground, a VM localiza o diretório Home do usuário e abre/cria o arquivo oculto **`~/.historico_harpia`**, lendo e carregando os comandos anteriores. Ao fechar (via Ctrl+D ou comando `sair()`), o REPL atualiza e grava a lista de comandos de volta ao disco de forma persistente.

### O Arquivo Virtual `<playground>` e Persistência de Escopo

Para que o desenvolvedor declare uma variável em uma linha e ela continue visível na linha seguinte, o playground inicializa um módulo virtual sob o arquivo inexistente `<playground>`:

```go
exec.Modulo, _ = ctx.InicializarModulo(&hrp.ModuloImpl{
    Info: hrp.ModuloInfo{Arquivo: "<playground>"},
})
```

Todas as expressões avaliadas utilizam o mesmo escopo persistente deste módulo (`exec.Modulo.Escopo`), preservando o estado e evitando "perda de memória" entre as linhas digitadas.

---

## 11.1 Máquina Virtual de Pilha (Fase 2)

A partir da **Fase 2** (Fase de Otimização e Bytecode), o Harpia conta com uma máquina virtual de pilha altamente eficiente escrita em Go, substituindo a execução clássica de árvore (tree-walk).

### O Compilador de Bytecode

A AST do programa é compilada estaticamente para bytecode compacto (`.hrpc`) de passagem única:

- **Pool de Constantes**: Literais do programa (textos, números inteiros, decimais, booleanos, nulos) são internados de forma deduplicada no pool de constantes, otimizando alocações.
- **Opcodes de 1 Byte**: Instruções compactas que controlam a pilha (`OP_PUSH_CONST`, `OP_POP`, `OP_DUP`), execução aritmética (`OP_ADD`, `OP_SUB`), controle de fluxo (`OP_JMP`, `OP_JMP_FALSO`, `OP_RETORNE`) e escopo (`OP_CARREGAR_VAR`, `OP_ARMAZENAR_VAR`).
- **Super-Instruções & Fusão Estática (Fase D)**: Para maximizar o rendimento operacional, o compilador adota passagens de fusão de bytecodes. Ele identifica pares de instruções sequenciais comuns e os funde estaticamente em instruções atômicas compostas de alta velocidade:
  - `OP_RETORNE_CONST`: Funde `OP_PUSH_CONST` + `OP_RETORNE`, lendo do pool e retornando de forma direta e atômica. Otimiza inclusive retornos vazios (`retorne Nulo`).
  - `OP_RETORNE_VAR`: Funde `OP_CARREGAR_VAR` + `OP_RETORNE`, carregando o valor da variável e saindo do frame instantaneamente.
  - _Impacto_: Reduz pela metade a quantidade de decodificações e saltos da VM para operações de retorno.
- **Pulos e Loops Remendados**: Remendos inteligentes de endereçamento de 16 bits (`BigEndian uint16`) para gerenciar saltos de condicionais (`se/senao`) e laços (`enquanto`).

### Execução de Alta Performance via Flag `--vm`

Para rodar qualquer script na nova VM de bytecode, basta passar a flag `--vm` ao comando de execução:

```bash
$ harpia executar script.hrp --vm
```

### Motor JIT de Traço por Threaded Callbacks (Fase F)

Para atingir o limite máximo de velocidade de execução e aniquilar o custo clássico de decodificação de instruções de interpretadores virtuais (gargalos de loops de `switch/case`), o Harpia incorpora uma inovadora tecnologia de **Direct-Threaded Code JIT**:

- **Compilação Dinâmica "Just-In-Time"**: Ao carregar um frame de bytecode para execução, a VM de pilha realiza de forma transparente uma passagem de compilação threaded de passagem única, traduzindo o array plano de bytecodes em um array estável de ponteiros de funções Go (`[]InstrucaoThreaded`).
- **Currying de Operandos e Constantes**: Os operandos e constantes são pré-capturados no encerramento (closure) de cada callback Go em tempo de JIT. Isso elimina buscas de memória e incrementos de IP em tempo de execução, resolvendo os valores diretamente de referências estáticas.
- **Preservação de Pulos**: O array threaded de funções coincide perfeitamente em tamanho com o array plano original de bytes. Isso mantém os offsets e saltos de endereçamento de loops e desvios de condicionais 100% íntegros e compatíveis, sem necessidade de alterações estáticas na AST.
- **Impacto de Velocidade**: O loop principal executa chamadas diretas sequenciais de ponteiros de funções no array, contornando desvios e mispredictions de branch de CPU, entregando um ganho de performance colossal.

### Métricas de Benchmark (Interpretador vs VM)

Os testes de benchmark mostram ganhos espetaculares de performance medidos localmente:

- **Velocidade**: A VM de bytecode roda **2.18 vezes mais rápida** que o interpretador clássico de árvore sintática.
- **Redução de Alocação de Memória**: Consumo de heap reduzido em **74.8%** (de `74KB` por ciclo para apenas `18KB`).
- **Frequência de Alocações**: Redução de **64.0%** no número de alocações (GC nativo do Go é acionado muito menos vezes).

### Gerenciamento de Memória por Contagem de Referências (Fase 2.5)

A VM de pilha do Harpia conta com gerenciamento de memória explícito e determinístico:

- **Protocolo de Referências Ativo**: Utiliza as interfaces `ObjetoGC` e `GCMixin` (`hrp/gc.go`) para controlar as referências de forma ativa nas instruções de empilhamento (`push`), desempilhamento (`pop`) e armazenamento de variáveis (`OP_ARMAZENAR_VAR`).
- **Imunidade de Singletons**: Globais, classes nativas e constantes singleton (`Nulo`, `Verdadeiro`, `Falso`) são inicializadas com `-1` referências de forma imune, garantindo no-ops em retenções e prevenindo coletas acidentais.
- **Limpeza Ativa de Frames**: Símbolos locais e operandos remanescentes na pilha são desalocados e limpos de forma explícita imediatamente ao finalizar a execução de um frame (fim de chamada de função).
- **Coletor e Quebrador de Ciclos (Trial Deletion)**: Referências circulares fechadas órfãs (ex: lista A contém lista B, e lista B contém lista A) são detectadas a partir de varreduras no grafo léxico do escopo ativo e quebradas de forma simétrica (`hrp.ColetarCiclos(escopo)`), prevenindo vazamentos de memória (memory leaks) e preservando a integridade do sistema.
- **Pool de Alocação Rápida / Eden Space para Inteiros Curtos (Fase E)**: Para mitigar o estresse e overhead de alocações sobre o Garbage Collector do Go durante iterações intensas (loops), o runtime pré-aloca estaticamente interfaces do Go para inteiros na faixa de `-100` a `2000`. Sempre que um inteiro nessa faixa é instanciado na VM, a mesma interface imutável pré-alocada é retornada instantaneamente em tempo constante $O(1)$, evitando novas alocações no heap e acelerando operações matemáticas de contagem de loops.

### 11.2. Primitivas de Concorrência & Event Loop Cooperativo (Sprints 9 e 10)

A VM de pilha do Harpia integra suporte nativo a concorrência assíncrona baseada em corotinas de suspensão cooperativa:

- **Palavras-Chave**: `assincrono` e `aguarde`.
- **Mapeamento de Funções Assíncronas**: Funções marcadas com o modificador `assincrono funcao` têm seu flag `Assincrono` ativado pelo compilador de bytecode.
- **Inovação de Loop de Eventos Baseado em Goroutines**: Ao disparar uma chamada de função assíncrona (`OP_CHAMAR`), a VM detecta o flag ativo e, em vez de bloquear o fluxo principal de execução, delega a sua execução em background a uma nova goroutine leve do Go, retornando imediatamente um objeto `Promessa`.
- **Suspensão Cooperativa via `aguarde` (`OP_AWAIT`)**: Quando a instrução `aguarde` é executada sobre uma `Promessa` ativa, a execução do frame atual da VM cede cooperativamente. Ela registra um callback de encerramento na promessa (`prom.Registre`) e aguarda por meio de um canal seguro do Go (`chan hrp.Objeto`) até que a promessa seja resolvida com sucesso ou rejeitada por erro, garantindo que outras operações concorrentes em background progridam sem travar a VM.
- **Modelo CSP de Concorrência por Canais (Fase B)**: Integração do tipo nativo global `Canal` para troca sincronizada e thread-safe de dados entre goroutines (processos de background):
  - **`nova Canal()`**: Cria uma nova instância de canal de comunicação unificado.
  - **`meuCanal.enviar(dado)`**: Adiciona um dado no canal. Se houver algum processo assíncrono esperando na fila, entrega o dado instantaneamente (FIFO).
  - **`aguarde meuCanal.receber()`**: Retorna uma Promessa suspensa cooperativamente na VM ou no interpretador, que é resolvida com o primeiro dado disponível na fila (FIFO).

### 11.3. Robustez & Modo Sandbox de Segurança (Fase A)

Para garantir que o Harpia opere como um motor de backend profissional, seguro e de nível industrial, foram integradas proteções nativas no runtime e na biblioteca padrão:

- **Recovery Middleware (Prevenção de Pânicos em Goroutines)**: Todas as requisições tratadas pelo servidor HTTP em background são envelopadas por um tratador `defer recover()`. Se houver pânico lógico inesperado, ele é interceptado de forma segura, respondendo com HTTP 500 sem derrubar o processo e a execução global do interpretador Harpia.
- **Defesa contra Ataques Slowloris**: Configuração nativa de tempos de limite rígidos (`ReadTimeout: 5s`, `WriteTimeout: 10s`, `IdleTimeout: 120s`) no servidor HTTP para encerrar conexões obsoletas ou propositalmente lentas.
- **Modo Sandbox por Bloqueio de Acesso**: Adição das flags estruturais de restrição de segurança no contexto de execução do interpretador:
  - `BloquearArquivos`: Impede de forma física qualquer leitura, escrita, deleção ou modificação de arquivos do sistema operacional pelo módulo de `arquivos`.
  - `BloquearRede`: Impede abertura de conexões de escuta pelo `Servidor` HTTP ou chamadas de requisição cliente via `requisitar`.
  - Lança erros educativos ricos em português (`PSC-0005: ErroDeSistema - Acesso Negado`) caso os limites do sandbox sejam ultrapassados.

---

## 12. Guia de Sintaxe Rápida e Exemplos de Produção

Abaixo estão descritos snippets estruturados que consolidam as peculiaridades e a sintaxe operacional da linguagem Harpia.

### 1. Declaração de Variáveis, Constantes e Tipagem Opcional

```harpia
# Variável mutável com tipo inferido
var nome = "Carlos"

# Variável com tipagem estática opcional
var idade: Inteiro = 25

# Constante imutável obrigatória
const PI = 3.14159

# Tentativas de reatribuição causarão erro:
# PI = 3.14 (Gera erro PSC-0002)
```

### 2. Condicionais Limpas (Sem Parênteses)

```harpia
se idade >= 18 {
    imprimir("Maior de idade")
} senao se idade == 17 {
    imprimir("Quase lá! Falta pouco")
} senao {
    imprimir("Menor de idade")
}
```

### 3. Laços de Repetição (enquanto e para-em)

```harpia
# 1. Loop condicional 'enquanto'
var contador = 1
enquanto contador <= 5 {
    imprimir("Contador:", contador)
    contador = contador + 1
}

# 2. Loop iterativo 'para-em' com Tuplas
para item em (10, 20, 30) {
    imprimir("Item:", item)
}

# 3. Gerando sequências numéricas dinâmicas
para num em sequencia(1, 10, 2) {
    imprimir("Número ímpar:", num) # 1, 3, 5, 7, 9
}
```

### 4. Classes, Herança Simples e Enlace de 'self'

```harpia
classe Animal {
    # Construtor inicializador padrão
    func inicializar(self, nome) {
        self.nome = nome
    }

    func falar(self) {
        retorne "Som de animal"
    }
}

classe Cachorro estende Animal {
    # Construtor substituído
    func inicializar(self, nome, raca) {
        self.nome = nome
        self.raca = raca
    }

    # Método substituído
    func falar(self) {
        retorne "Au! Eu sou " + self.nome + " da raça " + self.raca
    }
}

# Instanciando e acionando métodos
var pet = nova Cachorro("Rex", "Pastor")
imprimir(pet.falar()) # Saída: "Au! Eu sou Rex da raça Pastor"

# Verificação de instância nativa
imprimir(pet instancia de Cachorro) # Saída: Verdadeiro
imprimir(pet instancia de Animal)   # Saída: Verdadeiro
```

### 5. Encadeamentos Fluídos com Operador Pipe (`|>`)

```harpia
func duplicar(x) {
    retorne x * 2
}

func somar(x, y) {
    retorne x + y
}

# O valor 10 é passado como primeiro argumento para duplicar,
# o resultado (20) é passado como primeiro argumento para somar(5) -> somar(20, 5)
var resultado = 10 |> duplicar |> somar(5)
imprimir(resultado) # Saída: 25
```

### 6. Servidor de Redes Assíncrono TCP Não-Bloqueante Completo

```harpia
# Implementação de um Servidor de Eco (Echo Server) rodando localmente na porta 3000
de "soquete" importe Soquete;

# 1. Cria socket TCP IPv4
const servidor = nova Soquete(2, 1);

# 2. Ativa reuso de portas e modo não-bloqueante
servidor.define_opcoes(1, 2, 1)
servidor.def_nao_bloqueante(Verdadeiro)

# 3. Liga e inicia escuta na porta 3000
servidor.associa("127.0.0.1", 3000)
servidor.ouve()

imprimir("Servidor Harpia rodando com sucesso em 127.0.0.1:3000")

enquanto Verdadeiro {
    # Aceita conexões entrantes sem travar o thread principal
    var cliente = servidor.aceita()

    enquanto Verdadeiro {
        # Lê pacotes de rede
        const dados = cliente.recebe(1024)

        # Se o cliente desconectar, encerra a conexão
        se nao dados {
            imprimir("Cliente desconectou do servidor.")
            pare
        }

        # Ecoa os mesmos dados recebidos de volta ao remetente
        cliente.envia(dados)
    }

    # Encerra socket do cliente
    cliente.fecha()
}
```

### 7. Integração e Contratos RPC Decoupled Nativos (`@backend/`)

O Harpia suporta a geração automática de contratos de API a partir de manifestos `dependencias.json` no ecossistema local do projeto.

#### O Manifesto `dependencias.json`

O arquivo `dependencias.json` deve ser colocado na raiz do projeto e mapeia o endereço e o caminho do backend local:

```json
{
  "conectarBackend": "../meu-backend",
  "urlBackend": "http://localhost:8083"
}
```

#### Uso das Importações RPC

Ao carregar scripts, o compilador intercepta importações iniciadas com o prefixo `@backend/`.

- **Análise Estática por AST (Fase C)**: O sistema de importações carrega o arquivo `.hrp` do backend correspondente e invoca o parser nativo do Harpia para gerar a sua árvore sintática abstrata (AST). Ele percorre as declarações de forma estática procurando nós reais de exportação (`DeclExportar` contendo `DeclFuncao`). Isso garante um mapeamento de contratos 100% preciso, imune a espaços, comentários ou quebras de linhas no arquivo original.
- **Geração Estática de Proxies**: Com base nas funções extraídas, o Harpia gera em tempo de execução um objeto de módulo proxy cujas propriedades são funções dinâmicas do Go. Ao serem executadas, elas realizam automaticamente uma requisição POST HTTP serializada para a URL mapeada em `dependencias.json`.

```harpia
# Importa de forma remota a função 'obterUsuario' definida no backend
de "@backend/usuarios" importe obterUsuario

var dados = obterUsuario("42")
imprimir(dados) # Realiza uma chamada HTTP POST de forma totalmente transparente!
```

---

## 13. Desenvolvimento Frontend Reativo e SPA (Sinais, JSX, Estilos e SSR)

A partir da **Fase 4**, o Harpia suporta de forma unificada o desenvolvimento de interfaces reativas e de alta performance que rodam diretamente no navegador do usuário final. O compilador transpila o código Harpia para JavaScript (ES6) otimizado e de tamanho mínimo acompanhado de um motor de Virtual DOM com reatividade baseada em Sinais (~2.2KB final).

### 13.1. Reatividade por Sinais (Fine-Grained)

A reatividade do Harpia atualiza de forma fina e cirúrgica apenas os nós do DOM que mudaram, prevenindo re-renderizações totais de página:

- **`sinal(valor)`**: Cria um estado reativo. Retorna um array `[ler, definir]`.
- **`sinalPersistente(chave, valorInicial)`**: Cria um estado reativo sincronizado automaticamente com o `localStorage` do navegador.
- **`usarFormulario(config)`**: Gestor de estado reativo, validações e submissões automáticas de formulários.
- **`usarConsulta(url, opcoes)`**: Cache global client-side, revalidação e mutação otimista (*SWR Pattern*).
- **`usarTema(chave, padrao)`**: Gestor de Dark/Light Mode com detecção automática do SO e persistência.
- **`usarArrastar(aoSoltar)`**: Hook reativo para gestos Drag-and-Drop e listas reordenáveis (Kanban).
- **`usarNotificacao()`**: Gerenciador global de Toast Notifications reativas (`notificar.sucesso(...)`, `notificar.erro(...)`).
- **`efeito(funcao)`**: Re-executa de forma automática sempre que os sinais dependentes sofrerem alteração.
- **`derivado(funcao)`**: Cria um sinal computado e memoizado.
- **`armazem(objeto)`**: Gerenciador de estado global sincronizado entre múltiplos componentes.



```harpia

var contadorSinal = sinal(0);
var contador = contadorSinal[0];
var setContador = contadorSinal[1];

efeito(funcao() {
    imprimir("Valor atual do contador: " + contador());
});
```

> ⚠️ **Nota Crítica sobre Escopo de Sinais**: Para garantir o funcionamento correto da reatividade fina, os **Sinais devem ser declarados no escopo global (fora da função do componente)** ou em módulos de gerenciamento de estado separados. Declarar sinais diretamente dentro do corpo da função do componente (`funcao MeuComponente() { var s = sinal(0); ... }`) fará com que o sinal seja recriado e reiniciado ao seu valor inicial a cada renderização e reconciliação disparada por efeitos reativos.

### 13.2. Componentes e Sintaxe JSX-like

Permite mesclar tags HTML e códigos de forma nativa e semantica:

- **Componentes funcionais**: Funções normais que retornam marcações JSX-like.
- **Eventos**: Atributos como `aoClicar` que mapeiam de forma nativa para eventos `onclick` de browser.
- **Componentes Nativos de UI**:
  - `<Link para="...">`: Navegação SPA de 1ª classe que previne recarregamentos da página e gerencia a URL via History API.
  - `<Escolha valor={...}>`: Pattern matching nativo de UI contendo blocos `<Caso valor="...">` e `<Padrao>`.
  - `<Aguardar recurso={...}>`: Tratamento declarativo de carregando/erro/sucesso de dados assíncronos.
  - `<Portal>`: Renderização de subárvores JSX diretamente no `document.body` (modais, popups e tooltips sem estouro de z-index).
  - `<BarraDeProgresso>`: Indicador de progresso de carregamento de rota animado no topo do navegador.

- **Estruturas Inline**:
  - `<se condicao={...}>...</se>`: Condicional dinâmico.
  - `<para item em lista={...} chave="id">...</para>`: Loops reativos com diffing de chaves (*keyed reconciliation*).

```harpia
funcao App() {
    retorne <div classe="p-4">
        <Link para="/sobre">Sobre nós</Link>
        <Escolha valor={TelaNav()}>
            <Caso valor="tarefas"><RotaTarefas /></Caso>
            <Padrao><div>404</div></Padrao>
        </Escolha>
    </div>;
}
```


### 13.3. Estilização Nativa e Classes Utilitárias em PT

O compilador suporta três pilares de estilização e extração sob demanda:

1. **Bloco `estilo`**: Palavra-chave para declarar folhas estáticas em português (ex: `corDeFundo → background-color`, `raio-grande → border-radius`) com suporte a aninhamento e pseudo-classes.
2. **Tailwind em PT**: O compilador extrai apenas as classes utilitárias PT usadas (como `flex-linha`, `p-4`, `fundo-azul`, `itens-centro`) e gera o arquivo `estilos.css` de saída.
3. **Estilo Inline Reativo em PT (`estilo={{...}}`)**: Se precisar aplicar estilos dinâmicos diretamente em um elemento usando propriedades em português, passe obrigatoriamente um **Objeto/Dicionário** (ex: `estilo={{ "corDeFundo": "#ff0000", "cor": "white" }}`). O runtime do Harpia traduzirá estas propriedades de forma dinâmica em tempo de execução. Strings em português escritas diretamente nos atributos (como `classe="corDeFundo: ..."` ou `style="corDeFundo: ..."`) **não são válidas** e serão ignoradas pelo navegador.
4. **Estilo Inline Tradicional (`style="..."`)**: Se desejar escrever estilos inline convencionais como string crua tradicional, utilize o atributo nativo HTML `style` acompanhado de regras padrão em inglês (ex: `style="background-color: #ff0000; color: white;"`).

### 13.4. Roteamento por Arquivos (File-system Routing)

A CLI detecta de forma automática pastas de rotas (`/web/rotas/` ou `/rotas/`) e cria o mapeamento SPA automático. O componente especial `<Link para="...">` intercepta cliques e navega de forma instantânea sem recarregar fisicamente a página.

### 13.5. SSR (Server-Side Rendering) e Hidratação

O servidor de backend do Harpia (`stdlib/http/http.go`) pode renderizar as páginas em HTML estático inicial instantâneo contendo metadados ricos de JSON-LD Schema.org (AEO) e OpenGraph. No navegador, o runtime web liga de forma invisível os Sinais existentes na estrutura física (Processo de Hidratação), ligando os fios de reatividade sem piscar ou destruir o DOM estático inicial.

### 13.6. Arquitetura SPA comparada ao Angular

Para desenvolvedores com experiência em Angular, o Harpia Web oferece equivalências diretas e simplificadas de design:

- **Pipes** ➔ Operador Pipe nativo (`|>`) para formatações visuais limpas em templates.
- **Directives** ➔ Tags de controle JSX (`<se>` para `*ngIf`, `<para>` para `*ngFor`).
- **Services** ➔ Exportações de estado global baseadas em `armazem()`.
- **Validators** ➔ Sinais compostos derivados (`derivado()`).

### 13.7. Modelo Híbrido de Desenvolvimento (Arquivos Separados)

Para suportar o desenvolvimento de sistemas complexos e evitar arquivos gigantescos, o Harpia permite separar de forma limpa as responsabilidades visuais, de estilo e de comportamento lógico:

1. **Estilos em Português (`.estilo.hrp`)**: Arquivos com extensão `.estilo.hrp` contêm exclusivamente blocos de estilo declarados em português (ex: `estilo Caixa { ... }`). Eles podem ser importados normalmente no seu arquivo de lógica.
2. **Layouts HTML Separados (`.html`)**: Você pode extrair a marcação JSX para arquivos `.html` separados e carregá-los de dentro da lógica do componente usando a chamada nativa `importarHtml("./template.html")`. O compilador em Go lê o arquivo e faz o inline dinâmico do HTML traduzido em tempo de compilação.

```harpia
# Exemplo de arquivo lógico: BotaoPersonalizado.hrp
de "web" importe sinal, importarHtml;
de "./BotaoPersonalizado.estilo.hrp" importe CaixaDeBotao; # Importa estilo do .estilo.hrp

funcao BotaoPersonalizado() {
    var [contador, setContador] = sinal(0);
    # Carrega e injeta o layout físico de forma transparente
    retorne importarHtml("./BotaoPersonalizado.html");
}
```

### 13.8. Recursos e Primitivas de Nível de Produção

O ecossistema frontend do Harpia inclui inovações de performance e facilidade de desenvolvimento para sustentar sistemas corporativos de grande porte:

- **Two-Way Data Binding (`ligar={sinal}`)**: Elimina o código repetitivo em formulários. Ao usar `<input ligar={nome} />`, o compilador e o runtime criam o vínculo bidirecional reativo automático entre o sinal de estado e o elemento físico de entrada do browser.
- **Modificadores de Eventos Declarativos**: Encadeamento direto na propriedade de eventos para manipulação do comportamento físico (ex: `aoEnviar_prevenir={submeter}` intercepta e executa `e.preventDefault()` de forma transparente antes do callback, e `aoClicar_parar` executa `e.stopPropagation()`).
- **Keyed Diffing (`chave`)**: Desempenho linear $O(N)$ em renderizações de listas. O algoritmo de diff do Virtual DOM no `runtime-web.js` utiliza o atributo `chave` em tags dentro de loops `<para>` para reutilizar e reposicionar nós físicos no DOM em vez de destruí-los.
- **Sinais Persistentes (`sinalPersistente`)**: Primitiva de reatividade sincronizada automaticamente com a API de `localStorage` do navegador do usuário final.
- **Sinais Assíncronos (`recurso`)**: Simplifica a gestão de requisições HTTP e consumo de APIs fornecendo flags de estado síncronas de progresso: `.carregando()`, `.erro()`, e `.ok()`.
- **Injeção de Dependências (`Provedor` & `injetar`)**: Permite prover instâncias de stores e serviços no topo da árvore de componentes e recuperá-los de forma limpa em componentes filhos profundos, evitando o acoplamento excessivo de propriedades (_prop-drilling_).
- **Componentes de UI Nativos e Acessíveis**:
  - `<FronteiraDeErro>`: Proteção de renderização que impede erros em componentes e widgets secundários de causarem tela branca no sistema inteiro, exibindo um componente de fallback amigável.
  - `<ListaVirtual>`: Renderiza estritamente os nós visíveis na tela para coleções de dados massivas (ex: 50.000 linhas), mantendo a performance de rolagem a 60fps constantes.
  - `<GradeDeDados>`: Tabela interativa com filtros rápidos de pesquisa em português, paginação automatizada e ordenamento rápido.

### 13.9. Criação de Projetos (Scaffolding)

Você pode inicializar uma estrutura padrão de projeto completa com suporte híbrido de forma automática utilizando a CLI Cobra:

```bash
harpia iniciar meu_app
```

O comando gerará os seguintes diretórios e arquivos de exemplo pré-configurados no disco:

- `/main.hrp` (ponto de entrada que monta a aplicação)
- `/web/rotas/rotas.hrp` (página de início demonstrando importações de arquivos)
- `/web/componentes/Botao.hrp` (componente visual lógico)
- `/web/componentes/Botao.estilo.hrp` (folha de estilo separada inteiramente em português)
- `/web/pages/Layout.html` (layout HTML separado demonstrando o uso de `importarHtml`)

### 13.10. Novas Primitivas Avançadas e Sinais de Tempo

- **Sinais com Debounce (`sinalDebounce`)**: O Harpia fornece a primitiva `sinalDebounce(valorInicial, tempoEmMs)` em seu runtime web. Ela atrasa de forma inteligente a atualização de estados reativos e expõe seu atualizador direto no getter (`ler.set`), integrando-se nativamente e sem boilerplates com o binding bidirecional `_ligar` em formulários de pesquisa.

### 13.11. SEO, Meta Tags, Scripts e Fontes na Web

O Harpia gerencia os aspectos cruciais de SEO, injeção de scripts externos, fontes e controle dinâmico de metadados na web de forma declarativa e sintonizada com a arquitetura SPA:

1. **O index.html Estático**: O comando `harpia compilar --alvo=web` gera o esqueleto mestre `index.html` na pasta de saída `dist/`, configurando a tag `<head>` com links para estilos unificados em `estilos.css`. Scripts corporativos, analytics e CDNs externas podem ser injetados diretamente na tag `<head>` desse HTML base.
2. **Importação de Fontes**: É possível fazer o carregamento de fontes (como Google Fonts) importando folha de estilos externas diretamente nas declarações de `estilo` do Harpia:
   ```harpia
   @import url("https://fonts.googleapis.com/css2?family=Inter:wght@400;700&display=swap");
   ```
3. **SEO Dinâmico Reativo**: Por meio do objeto nativo `documento` (DAP web), o Harpia permite alterar o título da aba e criar/atualizar meta tags (como Open Graph para redes sociais) em tempo de execução ao navegar de uma rota para outra:
   ```harpia
   documento.titulo = "Produto " + nome() + " | Loja Harpia";
   documento.definirMeta("description", "Compre agora o produto!");
   documento.definirMeta("og:title", nome());
   ```
4. **SSG e SSR**: Para otimização máxima de indexação em motores de busca mais simples que não rodam JavaScript, o compilador do Harpia possui suporte nativo para pré-renderização de rotas em HTML estático (Static Site Generation - SSG) ou renderização instantânea no servidor backend (Server-Side Rendering - SSR) com o processo de hidratação reativa síncrona no cliente de forma transparente.

### 13.12. Responsividade Nativa (@tela) na Web

O Harpia oferece suporte de fábrica para a criação de designs e layouts 100% responsivos adaptados para celulares, tablets e computadores, utilizando seletores em português:

1. **A Diretiva `@tela`**: Nos blocos de `estilo` do Harpia, o compilador traduz de forma estática a diretiva `@tela` para `@media` do CSS nativo, permitindo o aninhamento direto de media queries responsivas dentro das classes:

   ```harpia
   exportar estilo PainelDashboard {
       exibir: "grid";
       colunasGrid: "repetir(3, 1fr)";
       gap: "20px";

       # Tablets (telas menores que 1024px)
       @tela (larguraMaxima: 1024px) {
           colunasGrid: "repetir(2, 1fr)";
       }

       # Celulares (telas menores que 768px)
       @tela (larguraMaxima: 768px) {
           colunasGrid: "1fr";
       }
   }
   ```

2. **Breakpoints Utilitários**: No JSX e nas classes do Tailwind em português, é possível prefixar os breakpoints utilitários correspondentes (ex: `celular:flex-coluna`, `tablet:p-4`).
3. **Viewport de Fábrica**: Ao transpilar o projeto para a web (`harpia compilar`), o cabeçalho do `index.html` gerado recebe síncronamente a tag mestre `<meta name="viewport" content="width=device-width, initial-scale=1.0">` para garantir escala física real de 1:1 e impedir deformações visuais em dispositivos portáteis.

---

## Capítulo 14 — Segurança e Blindagem Corporativa (Security Audit)

As ferramentas e a CLI do Harpia foram submetidas a uma auditoria rigorosa de segurança de nível de produção, contando com defesas contra as vulnerabilidades mais críticas do mercado de software:

- **Prevenção de Zip Slip (Path Traversal)**: O comando de instalação de pacotes `harpia instalar` valida estaticamente todos os caminhos do arquivo ZIP extraídos no disco com `filepath.Clean` e `strings.HasPrefix(caminhoLimpo, pastaAlvo)`. Caso uma travessia ilegal com caminhos relativos (`..`) seja detectada, a extração é abortada com segurança na hora, impedindo corrupção física do sistema de arquivos.
- **Prevenção de Corridas de Dados (Anti-Race Condition)**: O interpretador web do playground local serializa de forma síncrona as execuções de código utilizando bloqueio de exclusão mútua (`sync.Mutex`). Isso garante que múltiplas requisições simultâneas não causem corridas de dados ao interceptar a saída padrão global de console (`os.Stdout`), isolando totalmente a saída de logs de cada usuário de forma segura.
- **Resiliência contra DoS de Rede**: O servidor web do playground limita síncronamente o tamanho do payload do editor de código para no máximo 1MB via `http.MaxBytesReader` e configura limites estritos de `ReadTimeout` e `WriteTimeout` no servidor HTTP Go.

---

## Capítulo 15 — Otimizações Avançadas e Desempenho da Máquina Virtual

A Máquina Virtual de bytecode e o runtime de execução do Harpia foram aprimorados com otimizações de baixo nível de classe mundial para sustentar aplicações de altíssima performance:

- **Recursion Guard (PSC-0015)**: Implementação de proteção ativa contra estouros físicos de pilha da VM. O interpretador rastreia a profundidade de execução das chamadas e interrompe loops recursivos infinitos ao ultrapassar o limite seguro de 1000 chamadas, lançando o erro estruturado `ErroDePilha` (PSC-0015).
- **Operand Stack Pre-allocation Pool**: Reaproveitamento agressivo de memória na VM. Utiliza um pool global sincronizado (`sync.Pool` em Go) para fornecer fatias pré-alocadas de operandos com capacidade fixa de 128 elementos. Ao fim da execução de cada bloco/função, os operandos são zerados e devolvidos ao pool, reduzindo a pressão do coletor de lixo (GC) de Go a zero para frames normais.
- **Morphic Inline Caching (MIC)**: Otimização em tempo de execução para a instrução de carregamento de variáveis (`OP_CARREGAR_VAR`). Símbolos resolvidos em loops quentes são cacheados de forma monomórfica em closures JIT. Se o escopo ou objeto de destino for idêntico ao do ciclo anterior, a VM extrai o valor diretamente em tempo constante $O(1)$ sem realizar buscas complexas de tabelas hash.
- **Profiler Embutido (`--perfil`)**: O comando `harpia executar --perfil` ativa a coleta síncrona de estatísticas e carimbos de tempo para cada instrução de bytecode (Opcode). Ao fim do programa, é exibida uma tabela de desempenho contendo a contagem exata de chamadas e hotspots lógicos de execução.

---

## Capítulo 16 — Novos Módulos Avançados da Biblioteca Padrão (Stdlib)

- **Logs Estruturados (`de "logs"`)**: Sistema de logging estruturado nativo com níveis (`info`, `alerta`, `erro`, `depurar`) e suporte a metadados dinâmicos (Mapas) e formatação selecionável (texto colorido amigável ou JSON para produção).
- **Métricas de Observabilidade (`de "metricas"`)**: Permite criar e registrar contadores e medidores (Gauges) dinâmicos compatíveis com o formato do Prometheus na rota `/metricas` para observabilidade de microsserviços.
- **Validador de Esquemas de Dados (`de "esquema"`)**: Permite declarar restrições de esquemas de dados complexos com validação em tempo de execução (Ex: `esquema.NovoEsquema({ "nome": esquema.Texto, "idade": esquema.Inteiro })`).
- **Agendador de Tarefas e Filas (`de "tarefas"`)**: Expõe o controle de filas concorrentes em memória e agendamento periódico baseado em Cron (Ex: `tarefas.agendar("*/5 * * * * *", funcao() { ... })`).
- **Foreign Function Interface (`ffi`)**: Ponte nativa bidirecional de baixo nível que permite carregar bibliotecas binárias compartilhadas C-compatíveis (`.so`, `.dll`, `.dylib`) e executar assinaturas externas diretamente no Harpia de forma síncrona e performática.
- **Resiliência e Estabilidade (`de "resiliencia"`)**: Padrões nativos de tolerância a falhas para microsserviços corporativos, incluindo _Disjuntor_ (Circuit Breaker com 3 estados), _Limite de Taxa_ (Rate Limiter via token bucket) e _Retentativa_ (Retry) com backoff exponencial.
- **Telemetria e Rastreamento (`de "telemetria"`)**: Observabilidade nativa e de baixo overhead compatível com a especificação OpenTelemetry, permitindo iniciar e finalizar Spans em formato JSON estruturado e registrar métricas com tags dinâmicas.

---

## Capítulo 17 — Extensões de CLI, DevOps e DevOps DX

- **Empacotamento Autônomo (`harpia empacotar`)**: Subcomando de compilação avançada de binários autônomos puros. O Harpia compila o código do usuário para bytecode `.hrpc` e o funde a um executável Go do interpretador, gerando um único executável nativo livre de dependências para o usuário final com suporte nativo a cross-compilation (via `--so` e `--arq`).
- **Testador de Estresse Concorrente (`harpia stressar`)**: Permite executar baterias massivas de requisições concorrentes e benchmarks automáticos para testar a resiliência de servidores e scripts Harpia.
- **Protocolo de Adaptador de Depurador (`harpia depurar`)**: Servidor TCP compatível com o protocolo oficial Debug Adapter Protocol (DAP). Permite a conexão e handshakes síncronos de IDEs modernas (como VS Code, Cursor) para depuração de nível profissional com breakpoints e inspeção de variáveis locais.
- **Extensão VS Code Oficial (`vscode-harpia`)**: Extensão oficial que habilita realce de sintaxe completo de alto nível, preenchimento rápido (snippets) para front/back, e se conecta via stdio/sockets diretamente aos servidores `lsp` (Language Server) e `depurar` (DAP) integrados na CLI.
  - **Gramática Multicores Enriquecida (TextMate)**:
    1. _Comentários Suaves_: Comentários iniciados por `#` (linha única) ou `<!-- -->` (bloco JSX/HTML) são renderizados de forma cinza suave. A regra de comentários foi movida para o topo da lista de prioridade, impedindo que palavras-chave como `funcao` ou `sinal` recebam cores indesejadas dentro de textos documentados.
    2. _Tags XML e JSX_: Elementos como `<div>`, `<section>`, `<main>`, `<button>` e os seus respectivos delimitadores (`<`, `/>`, `</`) são coloridos nativamente na IDE com o escopo oficial `entity.name.tag`.
    3. _Chamadas de Métodos e Funções_: Expressões como `somar(...)`, `imprimir(...)` ou `setCodigo(...)` são coloridas com o escopo de função `entity.name.function`.
    4. _Acessos e Propriedades de Objetos_: Atributos acessados através de ponto (como `self.nome`, `carrinho.itens` ou `res.json`) são destacados como `variable.other.property`.
    5. _Variáveis de Escopo OO_: Palavras-chave de orientação a objetos (`self` e `isto`) recebem o destaque clássico de linguagem (`variable.language.self`).
    6. _Suporte a Aspas Simples (`'textos'`)_: Textos delimitados por aspas simples são tratados e coloridos de forma idêntica a aspas duplas como strings de dados.
  - **Recomendação Automática de Ferramentas de DX**:
    A pasta `.vscode/` do plugin inclui o arquivo `extensions.json` recomendando a instalação do **Error Lens** e do **GitHub Copilot** de forma nativa para habilitar a exibição inline de erros estáticos LSP na própria linha de código físico do desenvolvedor.
  - **Formatação Síncrona On-Save via LSP**:
    1. _No Servidor LSP (Go)_: Durante a inicialização, o servidor (`cmd/lsp.go`) declara suporte nativo de formatação síncrona via `"documentFormattingProvider": true`. Quando recebe a requisição síncrona `textDocument/formatting` enviada pela IDE, ele intercepta o comando e retorna as edições do código limpo processadas pela função nativa `FormatarCodigoHarpia(codigo)` declarada em `cmd/formatar.go`.
    2. _Na Extensão do VS Code (`vscode-harpia`)_: No arquivo `vscode-harpia/extension.js`, o cliente LSP é instanciado via classe `LanguageClient`. Ao inicializar, a biblioteca padrão `vscode-languageclient` detecta a capacidade `"documentFormattingProvider": true` fornecida pelo servidor e registra automaticamente a capacidade de formatação nativa na IDE.
    3. _Como usar no VS Code_:
       - **Atalho de Formatação**: Pressionar `Shift + Alt + F` (Windows/Linux) ou `Shift + Option + F` (macOS) com um arquivo `.hrp` aberto.
       - **Formatação Automática ao Salvar**: Habilitar a configuração `"editor.formatOnSave": true` nas configurações do VS Code para disparar a formatação limpa automaticamente em todo `Cmd+S` or `Ctrl+S`.
  - **Publicação da Extensão no VS Code Marketplace (`vscode-harpia`)**:
    Caso você queira gerar e publicar atualizações da extensão oficial para a comunidade global de desenvolvedores do VS Code, siga os passos abaixo usando o utilitário oficial `vsce` (VS Code Extension Manager):
    1. **Instalação do CLI**: Instale o gerenciador de extensões da Microsoft de forma global via npm:
       ```bash
       npm install -g @vscode/vsce
       ```
    2. **Criação do Publicador (Publisher)**:
       - Crie uma conta de desenvolvedor no [Visual Studio Marketplace](https://marketplace.visualstudio.com/).
       - Crie um ID de Publicador exclusivo (ex: `harpia`).
       - Insira esse ID no campo `"publisher"` do arquivo `package.json` localizado dentro da pasta `vscode-harpia/`.
    3. **Token de Acesso Pessoal (PAT)**:
       - Crie uma conta no Azure DevOps (`dev.azure.com`) sob a mesma organização ou e-mail.
       - No painel superior direito do Azure DevOps, vá em **Personal Access Tokens**.
       - Adicione um novo token selecionando a organização "All accessible organizations", defina o escopo para **Marketplace (Publish)** com acessos de leitura e gravação (_Read & Write_). Salve o token (PAT) gerado em um local seguro.
    4. **Autenticação no Terminal**: Efetue login no publicador por meio do terminal:
       ```bash
       vsce login [seu-id-publicador]
       ```
       (Cole o PAT gerado no Azure DevOps quando solicitado).
    5. **Empacotamento e Publicação**:
       - Entre no diretório da extensão: `cd vscode-harpia`
       - Instale as dependências locais de desenvolvimento: `npm install`
       - **Publicar diretamente**: Execute `vsce publish` (ou incremente versões via `vsce publish patch` / `vsce publish minor`).
       - **Apenas empacotar localmente (offline)**: Para gerar um arquivo instalável `.vsix` localmente sem enviar para o Marketplace público, execute `vsce package`. O arquivo `.vsix` gerado pode ser compartilhado com qualquer desenvolvedor para instalação manual arrastando-o para a aba de extensões do VS Code.

---

## Capítulo 18 — Pacote de Inteligência Artificial e IA Generativa (`de "ia"`)

A Fase 6 introduz o pacote de IA padrão do Harpia, oferecendo uma interface declarativa, tipada e nativa em português para interação com modelos de linguagem locais e comerciais:

- **Agente de IA Declarativo**: A classe nativa `Agente(nome, instrucoes, provedor, modelo)` permite instanciar e parametrizar robôs autônomos com memória de conversação persistente nativa em Go, abstraindo a gestão do histórico e chamadas HTTP:

  ```harpia
  de "ia" importe Agente, validar_resposta

  # Instancia agente local usando Ollama
  var assistente = Agente("HarpiaHelper", "Você é um assistente sênior prestativo", "ollama", "llama3")

  # Envia prompts acumulando memória automaticamente no histórico do agente
  var resposta = assistente.perguntar("Qual é a capital do Brasil?")
  escreva(resposta)
  ```

- **Orquestração Multi-Agente Nativa**: O método `comunicar(outro_agente, mensagem)` permite que dois agentes troquem mensagens de forma autônoma e cooperativa, atualizando seus respectivos históricos locais de conversas de forma transparente.
- **Provedores de IA Integrados**: Conector nativo de alto desempenho que suporta `ollama` local (com fallback transparente para nuvens corporativas como `gemini` e `openai` utilizando chaves de ambiente).
- **Contratos Semânticos de IA**: A função `validar_resposta(esquema, resposta_json)` permite certificar que a resposta textual em formato JSON retornada pelo modelo de IA é totalmente válida e segura contra o esquema de dados declarado no código antes de processá-la.

---

## Capítulo 19 — Conectores de Banco de Dados Corporativos (`de "bd"`)

Drivers de persistência robustos prontos para escalar em ambientes de alta concorrência corporativa:

- **Drivers Nativos Expandidos**: Implementação nativa em Go de conectores de banco para **PostgreSQL**, **MySQL**, **SQLite** e **MongoDB**, com séries históricas de migrations encapsuladas e gestão de credenciais via contexto.
- **ORM Estático e Tipado**: Permite mapear schemas de tabelas de forma declarativa e síncrona diretamente no Query Builder via `conn.tabela("nome", schema)`. A VM realiza a validação de tipo e a detecção de colunas inexistentes em runtime com mensagens educativas em português.
- **Banco de Dados Vetorial Integrado**: O conector `conectarQdrant(url, colecao)` provê um cliente vetorial nativo completo e de altíssimo rendimento para Qdrant, habilitando as operações de `inserir`, `buscar` (por cossenos/L2) e `deletar` pontos vetoriais com payload.
- **Pool de Conexões Otimizado**: Reaproveitamento de conexões ativas a partir de um _pool_ thread-safe com backoff exponencial e detecção automatizada de timeouts do servidor.
- **Transações Atômicas**: Block declarativo `bd.transacao(funcao() { ... })` que garante commit atômico ou rollback completo em erros.

---

## Capítulo 20 — WebAssembly (WASM), WASI e Microsserviços Corporativos

A camada de execução WASM e WASI eleva o limite de processamento matemático pesado e de isolamento seguro no navegador e no backend, com interoperabilidade síncrona:

- **Alvo de Compilação `--alvo=wasm`**: O comando `harpia compilar --alvo=wasm` compila a AST do Harpia em código Go altamente otimizado e dispara o compilador Go com as variáveis `GOOS=js GOARCH=wasm`, gerando arquivos binários `.wasm` nativos leves prontos para execução no navegador.
- **Sandbox Segura via `--alvo=wasi`**: Compilação estrita direcionada para ambientes de microsserviços via `GOOS=wasip1 GOARCH=wasm` (WASI), gerando uma sandbox WebAssembly de alta performance que isola variáveis de ambiente, acessos de rede e escrita em arquivos.
- **Interoperabilidade Síncrona**: Bridge tipada de alta velocidade para expor e consumir de forma síncrona funções nativas da VM em WebAssembly, pulando os gargalos de marshalling de strings.
- **Stone of Dedicatória**: A Fase 6 fecha o Harpia como um ecossistema de linguagem corporativa de ponta, contando com compilador, IDE nativa, ferramenta de empacotamento de binários, runtime WASM, IA integrada e drivers de banco escaláveis.

---

## Capítulo 21 — Documentação Ininterrupta e Estilo de Contribuição

A linha mestra de desenvolvimento do Harpia se mantém desde a sua concepção:

- **Sintaxe Humana**: Construída sob medida para falantes nativos de português, com palavras-chave fonéticas sem ambiguidade (`funcao`, `retorne`, `classe`, `estende`).
- **DX como prioridade**: Erros didáticos, mensagens contextuais e o comando `harpia erro explicar` integrado com LLMs locais.
- **CLI consistente**: Nomes de comandos, flags e cláusulas escritos exclusivamente em português brasileiro (`--alvo`, `--estrito`, `--otimizar-assets`).
- **Segurança por padrão**: Toda nova feature é auditada contra _path traversal_, condicional de corrida, DoS de payload e _race conditions_ em pipes assíncronos durante os testes de aceitação.

---

## Capítulo 22 — Práticas de Segurança e Programação Defensiva

O ecossistema Harpia adota diretrizes estritas de desenvolvimento seguro de software para garantir conformidade em auditorias estáticas de código (SAST), auditorias de conformidade do GitHub Advanced Security (CodeQL) e mitigar vulnerabilidades comuns de infraestrutura e aplicação:

### 22.1. Confinamento Robusto de Diretórios (Anti-Path Traversal)

Operações de leitura ou escrita no sistema de arquivos a partir de parâmetros fornecidos pelo usuário representam o risco de ataques de "travessia de diretório" (Zip Slip ou Path Traversal).

- **Boas Práticas de Implementação**:
  - Nunca confie apenas na concatenação de strings para caminhos (ex: `baseDir + caminho`).
  - Sempre utilize a função `filepath.Rel` para calcular o caminho relativo entre o diretório raiz autorizado e o caminho final resolvido.
  - Aborte a execução imediatamente se o caminho relativo gerado apontar para cima (`..`) ou for absoluto, impedindo que o atacante escape da raiz da aplicação para ler arquivos sensíveis do sistema:
    ```go
    rel, err := filepath.Rel(baseDir, p)
    if err != nil || strings.HasPrefix(rel, "..") || filepath.IsAbs(rel) {
        return nil, fmt.Errorf("caminho de arquivo ilegal e fora da raiz")
    }
    ```

### 22.2. Prevenção de Estouro e Conversão Insegura de Inteiros (CWE-190)

Coerções de tipos inteiros de maior capacidade para tipos de menor capacidade (como de `int64` para `int`) sem validação de limites podem sofrer truncamento em arquiteturas de 32-bits, ocasionando loops infinitos ou falhas de estouro silencioso.

- **Boas Práticas de Implementação**:
  - **Validação Estrita de Limites (Range Check)**: Certifique-se de que o número de 64-bits esteja contido de forma garantida e síncrona dentro dos limites de capacidade máxima e mínima do tipo de destino (`math.MinInt` e `math.MaxInt` em Go) antes de realizar o cast.
  - **Evitar Casts Desnecessários**: Se o Query Builder, JSON, ou biblioteca de destino suportar o tipo de dados mais amplo (como `int64`), transmita-o diretamente de forma estática sem realizar coerções.
  - **Uso de Constantes para Análise Estática**: Ao criar pools ou fatias de cache de tamanho fixo, declare os limites superior e inferior usando constantes (`const`) em vez de variáveis (`var`). Isso permite que analisadores de segurança (como o CodeQL) comprovem de forma estática e sem falsos positivos a segurança matemática das expressões.

### 22.3. Princípio de Privilégio Mínimo de Tokens de CI/CD

Contas de automação e robôs de entrega contínua (GitHub Actions) devem rodar com o escopo de segurança mais restritivo aplicável para o seu respectivo fluxo de trabalho, mitigando o risco de comprometimento do repositório por dependências de terceiros maliciosas.

- **Boas Práticas de Implementação**:
  - Declare de forma explícita e minimalista o bloco de permissões globais em todos os arquivos de workflows do Actions (`.yml`), definindo o token do GitHub para permissão síncrona exclusiva de leitura de código para checkout:
    ```yaml
    permissions:
      contents: read
    ```

### 22.4. Manipulações Cirúrgicas de Strings e Layouts

Substituições globais de strings ou expressões regulares agressivas de limpeza de caracteres especiais (como `.replace(/\{/g, '')` em JavaScript) podem inadvertidamente corromper códigos legítimos e assinaturas do usuário que contêm o caractere de forma sã.

- **Boas Práticas de Implementação**:
  - Seja específico. Se o objetivo é remover apenas um caractere delimitador ou abertura de bloco no final ou início da declaração de linha, utilize métodos específicos de fronteira como `endsWith` combinado com fatiamento de string (`slice`) em vez de varreduras agressivas:
    ```javascript
    let assinatura = linha.trim();
    if (assinatura.endsWith("{")) {
      assinatura = assinatura.slice(0, -1).trim();
    }
    ```
