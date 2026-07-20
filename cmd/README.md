# Pacote `cmd` (Comandos da CLI do Harpia)

O pacote `cmd` é o núcleo de controle da Interface de Linha de Comando (CLI) do **Harpia**. Ele serve como ponte de integração estrutural entre a inicialização inicial do executável (no ponto de entrada `main.go`) e a biblioteca de gerenciamento de comandos **Cobra** (`github.com/spf13/cobra`).

Este pacote encapsula toda a lógica de tratamento de entrada de terminal, orquestração de contextos de execução do interpretador, inicialização de interfaces interativas e gerenciamento de atualizações automatizadas do binário.

---

## 📖 Índice

1. [Filosofia de Design](#-filosofia-de-design)
2. [Variáveis de Build (Injeção de Metadados)](#-variáveis-de-build-injeção-de-metadados)
3. [Estrutura do Ponto de Entrada Principal](#-estrutura-do-ponto-de-entrada-principal)
4. [Subcomando: `executar` (`exec`)](#-subcomando-executar-exec)
   - [Mecânica de Funcionamento](#mecânica-de-funcionamento)
   - [Fluxo de Decisão do interpretador](#fluxo-de-decisão-do-interpretador)
5. [Subcomando: `atualize`](#-subcomando-atualize)
   - [Mapeamento de Sistema e Arquitetura](#mapeamento-de-sistema-e-arquitetura)
   - [Fluxo de Atualização Semântica](#fluxo-de-atualização-semântica)
   - [Baixando e Instalando Releases](#baixando-e-instalando-releases)
6. [Subcomando: `atualize`](#-subcomando-atualize)
   - [Mapeamento de Sistema e Arquitetura](#mapeamento-de-sistema-e-arquitetura)
   - [Fluxo de Atualização Semântica](#fluxo-de-atualização-semântica)
   - [Baixando e Instalando Releases](#baixando-e-instalando-releases)
7. [Subcomando: `testar`](#-subcomando-testar)
8. [Subcomando: `checar` — Linter Estático](#-subcomando-checar--linter-estático)
9. [Exemplo de Uso Prático](#-exemplo-de-uso-prático)

---

## 🎯 Filosofia de Design

A CLI do Harpia foi desenhada para colocar a língua portuguesa em primeiro lugar. Por isso, toda a experiência no terminal:

- Descrições curtas (`Short`) e longas (`Long`) dos comandos;
- Mensagens explicativas de help/ajuda;
- Diagnósticos e mensagens de erro do sistema;
- Aliases (apelidos) de comandos (ex: `exec` como apelido para `executar`);
  são escritos nativamente em **Português (PT-BR)**, criando consistência com o propósito educacional e de acessibilidade da linguagem Harpia.

---

## 🏗️ Variáveis de Build (Injeção de Metadados)

O arquivo `cmd.go` expõe três variáveis globais de estado que rastreiam metadados cruciais do binário compilado. Elas são declaradas com valores sentinela para permitir desenvolvimento local (ex: via `go run`) e são substituídas pelo **GoReleaser** no pipeline de CI/CD utilizando injeção via flags do linker (`ldflags`):

```go
var (
    Commit   string = "-"
    Datetime string = "0000-00-00T00:00:00"
    Version  string = "dev"
)
```

### Detalhamento das Variáveis:

| Variável   | Tipo     | Valor Padrão (Local)    | Propósito de Negócio                                                                                                                                                    |
| :--------- | :------- | :---------------------- | :---------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `Commit`   | `string` | `"-"`                   | Identifica a hash curta (SHA-1) do Git referente ao código fonte usado na geração deste binário. Utilizado para auditoria técnica precisa de bugs.                      |
| `Datetime` | `string` | `"0000-00-00T00:00:00"` | Armazena o carimbo de data/hora (formato ISO-8601) que marca o instante exato em que a build do binário foi executada.                                                  |
| `Version`  | `string` | `"dev"`                 | Armazena a versão semântica oficial da release (ex: `0.3.0`). Usada pelo módulo `atualize` para comparar com o GitHub e verificar se há novas atualizações disponíveis. |

---

## 🔌 Estrutura do Ponto de Entrada Principal

### `InstalarComandos(raiz *cobra.Command)`

Esta é a **única função exportada** e pública do pacote `cmd`. Ela serve como o único ponto de montagem da árvore hierárquica de comandos CLI. Centralizar a instalação aqui simplifica a leitura da arquitetura, permitindo identificar todos os comandos suportados em um único local.

- **Parâmetro**: `raiz` (`*cobra.Command`) - O comando base pré-configurado no `main.go`.
- **Implementação**:
  ```go
  func InstalarComandos(raiz *cobra.Command) {
      raiz.AddCommand(comandoAtualize())
      raiz.AddCommand(comandoExecutar())
  }
  ```

---

## 🚀 Subcomando: `executar` (`exec`)

O subcomando `executar` (apelidado de `exec`) é o ponto de entrada primário para a execução física de códigos na máquina virtual Harpia.

### Mecânica de Funcionamento

O interpretador inicializa o ambiente importando implicitamente a biblioteca padrão e configurando as variáveis de escopo local. No arquivo `executar.go`, a importação blank (`_`) é crítica:

```go
import _ "github.com/natanfeitosa/harpia/stdlib"
```

> **Por que isso é necessário?**  
> A importação anônima força a execução das funções `init()` de todos os arquivos no pacote `stdlib`. Isso registra dinamicamente todos os módulos internos e funções embutidas (como `escreva()`, `leia()`, etc.) na tabela global de símbolos do interpretador antes de qualquer código de usuário rodar.

### Fluxo de Decisão do interpretador

Quando o usuário digita `harpia executar [arquivo] [flags]`, o seguinte algoritmo de decisão é acionado em `executar.go`:

```
                    [Início do Comando Executar]
                                 │
                 Obtém diretório atual (os.Getwd())
                                 │
                   Cria novo Contexto de Execução
                 (hrp.NewContexto com diretório local)
                                 │
              Garante destruição com defer ctx.Terminar()
                                 │
               ┌─────────────────┴─────────────────┐
               ▼                                   ▼
   Há arquivo OU flag -c?                     Sem argumentos e sem -c
               │                                   │
               │                                   ▼
               │                        Inicializa o Playground
               │                         (TUI interativa/REPL)
               │                                   │
               ▼                                   ▼
      [Modo de Execução]                        [Fim]
               │
               ├─► [1] Se há argumento posicional (caminho do arquivo):
               │       Se a flag --vm estiver presente, compila o código para bytecode
               │       plano e despacha para a Máquina Virtual de Pilha (vm/vm.go).
               │       Caso contrário, executa via interpretador clássico hrp.ExecutarArquivo().
               │       Se falhar: chama hrp.LancarErro() e encerra.
               │
               └─► [2] Se há flag -c/--codigo (script inline):
                       Executa hrp.ExecutarString()
                       Se falhar: chama hrp.LancarErro() e encerra.
```

1. **Obtenção do Diretório de Trabalho**: A chamada para `os.Getwd()` define qual o diretório corrente do processo do usuário. Esse caminho é adicionado à lista de `CaminhosPadrao` do contexto da VM, permitindo que a importação de módulos locais (`importar ...`) seja resolvida corretamente a partir do local de chamada.
2. **Ciclo de Vida do Contexto**: O contexto da VM (`hrp.Contexto`) é protegido por `defer ctx.Terminar()`. O método `Terminar()` limpa a memória, remove referências circulares, realiza o flush (escoamento) de buffers de escrita ativos e finaliza o coletor interno, prevenindo memory leaks.
3. **Precedência de Execução**: Se o usuário fornecer um arquivo E código inline (`-c`), o arquivo é executado **primeiro**. O código inline é avaliado em seguida no mesmo contexto. Esse comportamento permite usar arquivos locais como scripts de "setup" de variáveis e ambientes antes de rodar comandos de teste rápidos passados inline no terminal.
4. **Tratamento Amigável de Falhas**: Exceções e erros levantados pela VM são capturados e tratados via `hrp.LancarErro()`. Em vez de simplesmente exibir um stacktrace feio do Go, o Harpia formata um traceback amigável em português com ponteiros visuais (sublinhado) indicando exatamente o token incorreto ou a linha física falha.

---

## 🔄 Subcomando: `atualize`

O subcomando `atualize` automatiza o download e a instalação de novas releases binárias oficiais diretamente do repositório no GitHub.

### Mapeamento de Sistema e Arquitetura

Como os arquivos binários compilados pelo GoReleaser no GitHub seguem uma nomenclatura padronizada de SO e arquitetura de CPU, o módulo realiza a conversão das variáveis de runtime do Go (`runtime.GOOS` e `runtime.GOARCH`) para coincidir com as nomenclaturas de distribuição usando as funções utilitárias:

- **`nomeOS() string`**:
  - `darwin` ➔ `Darwin` (Evita erros comuns como o typo "Darwind")
  - `linux` ➔ `Linux`
  - `windows` ➔ `Windows`
  - _Fallback_: Retorna `Linux` por segurança operacional.
- **`nomeArch() string`**:
  - `amd64` ➔ `x86_64` (Compatibilidade padrão Unix)
  - `386` ➔ `i386`
  - _Fallback_: Retorna a própria string gerada pelo runtime (ex: `arm64`).

---

### Fluxo de Atualização Semântica

Ao acionar `harpia atualize`, o interpretador executa as seguintes etapas estruturadas:

```
                  [Comando harpia atualize]
                                 │
                   Identifica Home do Usuário
                                 │
              Monta caminho do binário local executável
             (~/.harpia/bin/harpia[.exe])
                                 │
              Verifica versão local (executa -v)
                                 │
         Busca tags mais recentes via API do GitHub
               (GET api.github.com/repos/.../tags)
                                 │
             ┌───────────────────┴───────────────────┐
             ▼                                       ▼
    Versão local é "dev"                  Versão local é SemVer válida
             │                                       │
             ▼                                       ▼
     Emite erro e para                   Compara versões com SemVer
    (versão dev não atualiza)             (jaAtualizado(local, github))
                                                     │
                                       ┌─────────────┴─────────────┐
                                       ▼                           ▼
                                  Já atualizado?            Nova versão disponível?
                                       │                           │
                                       ▼                           ▼
                                 Fim do fluxo             Inicia Download
                             (Mensagem amigável)              e Instalação
```

1. **Localização**: Descobre o diretório de instalação padrão montando o caminho absoluto sob a pasta home do usuário (`~/.harpia/bin/harpia` ou `~/.harpia/bin/harpia.exe` se for Windows).
2. **Extração de Versão Local**: Executa o próprio binário de forma isolada disparando o subcomando `-v` através do pacote nativo `os/exec`. Extrai a string correspondente à versão. Se retornar a string `"dev"`, o processo é abortado com erro informativo, já que ambientes de desenvolvimento local/compilações manuais não devem sofrer sobreposição automática de releases do GitHub.
3. **Consulta Remota**: Utiliza um cliente HTTP reutilizável (`httpClient`) configurado com um timeout preventivo de **10 segundos** para obter a lista de tags da API do GitHub. O JSON é desserializado para a struct `Tag`:
   ```go
   type Tag struct {
       Name string `json:"name"`
   }
   ```
4. **Comparação de Versões**: Com a tag mais recente capturada, usa a biblioteca `github.com/Masterminds/semver/v3` para validar o cenário:
   ```go
   func jaAtualizado(a, b string) bool {
       i, _ := semver.NewConstraint("< " + b)
       n, _ := semver.NewVersion(a)
       return i.Check(n)
   }
   ```
   Se a versão instalada for menor que a tag remota, o processo inicia o download.

---

### Baixando e Instalando Releases

Se uma nova versão for detectada, o fluxo inicia a substituição do binário atual pelo novo baixado do GitHub:

```go
func urlDaVersao() string {
    url := "https://github.com/natanfeitosa/harpia/releases/latest/download/"
    url += nomeOS() + "_" + nomeArch()
    if isWindows() { return url + ".zip" }
    return url + ".tar.gz"
}
```

1. **Download Seguro**: Cria um arquivo temporário em disco protegido com prefixo `-hrp` no diretório padrão temporário do sistema operacional (limpo e apagado ao final via `defer os.Remove(...)`).
2. **Invocação do `curl`**: Executa o comando utilitário `curl` do sistema de forma nativa para realizar o download com suporte a redirecionamento (`--location`), interrupção caso haja falhas HTTP (`--fail`) e exibindo uma barra de progresso elegante no próprio terminal (`--progress-bar`).
3. **Descompactação Inteligente**:
   - **Sistemas Unix (Linux/macOS)**: Utiliza o utilitário nativo de sistema `tar` (`tar -xf [arquivo] -C [destino]`) para extrair e sobrescrever o binário diretamente em `~/.harpia/bin/`.
   - **Windows**: Tenta primeiramente invocar o utilitário `unzip` do sistema. Se este não for localizado na variável de ambiente PATH, faz um fallback e tenta acionar a ferramenta `7z` (7-Zip) com parâmetros automáticos silenciosos de sobrescrita.

---

## 🛠️ Exemplo de Uso Prático

### Executando Códigos Através do Terminal

#### Rodando um arquivo físico `.hrp`:

```bash
# Executa um script local chamado ola_mundo.hrp
harpia executar ola_mundo.hrp
```

#### Rodando código de forma rápida inline (flag `-c`):

```bash
# Executa o interpretador sem criar arquivos locais
harpia executar -c "escreva('Olá, Harpia!')"
```

#### Executando arquivo local e complementando com código inline:

```bash
# Roda as definições contidas em setup.hrp e executa a função de teste inline
harpia executar setup.hrp -c "inicializarSessao()"
```

#### Rodando o Playground (TUI Interativo):

```bash
# Sem parâmetros, inicia o modo REPL console interativo
harpia executar
```

### Atualizando o Interpretador

```bash
# Verifica se há novas versões estáveis no GitHub e instala se houver
harpia atualize
```

---

## 🧪 Subcomando: `testar`

Varre o diretório em busca de arquivos `.hrp`/`.hrp` (recursivo), executa todos os blocos `testar`/blocos nativos e apresenta relatório visual com `✅ [PASSOU]`/`❌ [FALHOU]`. Implementação em `cmd/testar.go`.

---

## 🧹 Subcomando: `checar` — Linter Estático

Novo subcomando adicionado no **Sprint 6** e estendido no **Sprint 8**. Permite varrer o código sem executá-lo, identificando erros _preventivos_ da AST:

```bash
harpia checar ./src
```

### Flags Suportadas

- `--formato`: Define o formato de saída do relatório de erros.
  - `texto` (Padrão): Saída formatada amigável agrupada por arquivo com sumário.
  - `json`: Saída padronizada no formato `Diagnostic` de LSP (Language Server Protocol) em português.

#### Exemplo de Saída JSON-LSP (`--formato=json`)

```json
[
  {
    "range": {
      "start": { "line": 2, "character": 4 },
      "end": { "line": 2, "character": 10 }
    },
    "severity": 1,
    "code": "PSC-0005",
    "source": "harpia-linter",
    "message": "Identificador 'x' não encontrado no escopo"
  }
]
```

| Verificação                              | Descrição                                                                                 |
| :--------------------------------------- | :---------------------------------------------------------------------------------------- |
| Redeclaração de identificadores          | Detecta nomes repetidos em um mesmo escopo                                                |
| Reatribuição de `const`                  | Impede escrita em constantes                                                              |
| Identificadores indefinidos              | Cobre builtins via tabela `globalsLinter`                                                 |
| Parâmetros duplicados                    | Detecta `func f(a, a, b)`                                                                 |
| SQL Injection (`HRP-SEC-001`)            | Detecta concatenação de strings em consultas SQL (`consultar`/`executar`)                 |
| Vazamento de Credenciais (`HRP-SEC-002`) | Detecta segredos estáticos declarados no código em variáveis como `senha`, `token`, `key` |
| Canais Inseguros (`HRP-SEC-003`)         | Detecta operações síncronas de canal (`enviar`/`receber`) fora de contexto assíncrono     |

**Arquitetura interna (`cmd/checar.go`):**

- `EscopoLinter` — pilha de escopos léxicos encadeada via `Pai`.
- `Linter.Checar(BaseNode)` — visitor recursivo que decide qual caso tratar com base no tipo dinâmico do nó AST.
- `globalsLinter` — mapa constante sincronizado manualmente com a stdlib (evita falso positivo em `imprimir`, `assegura`, etc.).
- **Mapeamento Espacial (Sprint 8)**: O linter resgata as posições absolutas físicas (linha, coluna e tamanho do token) de cada nó utilizando o mapa de posições gerado no analisador sintático, permitindo gerar o objeto `range` preciso no JSON-LSP.
- Saída tradicional categoriza erros em **Erros de Sintaxe** (do parser) e **Erros Semânticos** (do linter), encerrando com sumário `Total: %d erro(s) semântico(s) em %d arquivo(s).`.

**Limitação conhecida**: `globalsLinter` precisa ser updated manualmente quando novas funções stdlib forem adicionadas. Alternativa futura: injetar via `Contexto` da VM.

---

## 🛑 Subcomando: `erro` — Dicionário e Guia de Ajuda

Adicionado no **Sprint 7** e estendido no **Sprint 8**. Permite consultar explicações pedagógicas sobre erros do Harpia diretamente do terminal:

```bash
harpia erro PSC-0005
```

### Subcomando: `explicar` (com IA Local)

Fornece uma explicação inteligente e personalizada sobre o erro consultado utilizando um LLM local:

```bash
harpia erro explicar PSC-0005
```

- **IA Local (Ollama)**: Se conecta via HTTP ao serviço Ollama (`127.0.0.1:11434`) usando o modelo leve `gemma` para gerar explicações pedagógicas e interativas em português brasileiro, completas com exemplos de erro e correções recomendadas.
- **Fallback Inteligente**: Caso o Ollama não esteja instalado ou rodando localmente, o comando intercepta o erro de rede de forma graciosa, exibe instruções amigáveis passo a passo de como instalar o Ollama, e em seguida renderiza a explicação didática estática local do dicionário para que o desenvolvedor não fique parado.

---

## ⚡ Flag Global: `--estrito`

Adicionada no fechamento da **Fase 1**. Habilita verificação estrita de tipagem opcional declarada em variáveis, parâmetros de funções e assinaturas de retorno.

- **No comando `executar`**:
  ```bash
  harpia executar meu_script.hrp --estrito
  ```
  Se ativado, qualquer atribuição de valor incompatível com o tipo anotado ou passagem incorreta de parâmetros em runtime lança um erro `TipagemErro` (PSC-0004).
- **No comando `checar`**:
  ```bash
  harpia checar meu_script.hrp --estrito
  ```
  O linter analisa estaticamente as atribuições e emite diagnósticos preventivos de conflito de tipo.

---

## 🏗️ Novas Ferramentas de Tooling e Ecossistema (Fase 5, 5-B & 5-C)

### 11. Subcomando: `novo` — Scaffolding Estruturado

Inicializa uma nova árvore física padrão de projeto baseada no escopo e topologia desejada:

```bash
harpia novo monolito meu_app
harpia novo backend meu_back
harpia novo frontend meu_front
```

- **Proteção Física**: O comando verifica se a pasta do projeto já existe e aborta de forma síncrona com aviso para impedir sobrescritas acidentais de arquivos do usuário.

### 12. Subcomando: `crie` — Assistente de Templates (Generators)

Gera boilerplates estruturados prontos para uso em projetos existentes:

```bash
harpia crie rota sobre
harpia crie componente botao
```

- **Inteligência de Detecção**: O assistente detecta automaticamente a presença da pasta `/web` e direciona a escrita física de arquivos para `/web/rotas` e `/web/componentes`, criando o arquivo lógico `.hrp` e folha de estilos correspondente `.estilo.hrp` em português.

### 13. Subcomando: `lsp` — Servidor de Linguagem Integrado

Inicia o servidor oficial de Language Server Protocol (LSP) em Go via stdio para IDEs (VS Code/Cursor):

- **Autocompletar (`textDocument/completion`)**: Fornece sugestões instantâneas no editor de palavras-chave da linguagem (como `funcao`, `classe`, `se`) e embutidos reativos (como `sinal`, `sinalPersistente`, `recurso`).
- **Formatação "On-Save" (`textDocument/formatting`)**: Permite que o editor formate arquivos `.hrp` de forma automatizada ao salvar, usando a nossa heurística síncrona de alinhamento de blocos.
- **Diagnósticos de Clean Arch Inline**: Linter reativo que emite sublinhados vermelhos de erro na IDE em tempo de digitação caso um arquivo sob `/dominio/` tente importar algo de `/infra/` ou `/web/`.
- **Dicas de Passar o Mouse (`textDocument/hover`)**: Exibe informações ricas e assinaturas estruturadas de variáveis, funções e classes ao passar o mouse por cima do símbolo, integrando automaticamente blocos de documentações `///` e provendo documentações didáticas integradas para builtins (`imprimir`, `sinal`, `efeito`, etc.).
- **Ir Para a Definição (`textDocument/definition` / F12)**: Permite saltar instantaneamente do local de chamada de uma função, classe ou variável diretamente para a linha do arquivo onde o símbolo foi declarado de forma nativa.

### 14. Subcomando: `playground` — Editor e Depurador Web

Inicia o servidor de interface web do playground interativo:

```bash
harpia playground -p 8090
```

- **Dogfooding Reativo**: A própria página do playground (`playground/interface.hrp`) é escrita **100% em Harpia reativo**, usando Two-Way Binding (`_ligar`) e o componente de tabela inteligente `<GradeDeDados>` para renderizar as variáveis locais e logs capturados em tempo real.

### 15. Subcomando: `formatar` — Pretty-Printer Heurístico

Higieniza e alinha recuos de arquivos Harpia:

```bash
harpia formatar meu_codigo.hrp -w
```

- **Controle de Blocos**: Corrige a indentação baseando-se em contagens de delimitadores (`{}`, `[]`, `()`), remove linhas vazias duplicadas e preserva 100% de comentários e JSX.

### 16. Subcomando: `doc` — Geração de Documentações

Varre arquivos extraindo comentários especiais iniciados com três barras (`///`):

```bash
harpia doc calculos.hrp --formato=html
```

- Gera APIs e documentações estruturadas de funções, classes, constantes e variáveis em formato Markdown ou páginas responsivas estilizadas em HTML.

### 17. Subcomando: `diagramar` — Mapeador Arquitetural

Analisa o fluxo de importações entre as camadas do projeto:

```bash
harpia diagramar
```

- Gera grafos de relacionamento textual em formato Mermaid, validando se as dependências de isolamento do Clean Architecture foram violadas.
- **Formatos HTML e SVG Interativos**: Adicionada a flag `--formato html` ou `-f html` (e suporte a `svg`) junto de `--saida` ou `-s` para gerar visualizações ricas locais.
- **Colorização de Violações**: Identifica e colore dinamicamente o fluxo do diagrama gerado: linhas verdes representam conexões arquiteturais saudáveis, enquanto **linhas vermelhas grossas destacam violações** de dependência no mapa visual.
- **Exportação de SVG em 1 Clique**: O arquivo HTML interativo gerado traz embutido um script de renderização interativa do Mermaid.js via CDN de alta velocidade que inclui um botão para baixar a imagem vetorial `.svg` com as proporções perfeitas direto pelo navegador de forma portátil.

### 18. Subcomando: `instalar` — Gerenciador de Módulos

Download e resolvedor de dependências assíncrono:

```bash
harpia instalar [nome-do-pacote] [versao-opcional]
```

- **Resolução Remota e Registro Central**: Agora permite buscar pacotes públicos diretamente pelo nome do registro central em português (`URL_REGISTRO_CENTRAL`).
- **Gestão de Versões Semver**: Suporta a resolução de dependências baseadas em restrições de versão semver diretamente no arquivo de manifesto `pacote.hrp` ou `pacote.json` (ex: `"banco-dados": "1.0.0"`, `"banco-dados": "latest"`), resolvendo os caminhos remotos dos arquivos ZIP compactados automaticamente.
- Lê o arquivo de manifesto em português `pacote.hrp` (ou `pacote.json`), baixa os pacotes zip associados e os extrai de forma automatizada sob a pasta local `pt_modulos/` exibindo relatórios de progresso de download em tempo real.
- **Segurança de Extração (Anti-Zip Slip)**: Implementa um validador de sanidade de caminhos (`filepath.Clean` e verificação de prefixo) que bloqueia na hora a extração de arquivos maliciosos contendo caminhos de travessia de diretório (`..`), impedindo que pacotes ZIP corrompidos sobrescrevam ou criem arquivos fora da pasta alvo do projeto.

### 19. Subcomando: `empacotar` — Compilador e Gerador de Binários Standalone / WASM

O comando `empacotar` permite compilar scripts Harpia unificando interpretador e código de usuário em um executável nativo compilado autônomo (Single Binary Bundle) ou compilar o interpretador completo para WebAssembly:

- **Executáveis Standalone**: Compila e embute um script Harpia específico de forma a rodá-lo diretamente como binário executável local do SO (Windows, Linux ou macOS).
- **Compilação Cruzada para WebAssembly (WASM)**: Suporta `--so=js --arq=wasm` para compilar o interpretador Harpia completo para rodar de forma dinâmica 100% no cliente direto no navegador (`docs/portal/harpia.wasm`).
- **Loader Portátil `wasm_exec.js`**: Copia dinamicamente do GOROOT do seu sistema a biblioteca Javascript correta correspondente à versão instalada de Go na sua máquina para carregar o binário compilado.
