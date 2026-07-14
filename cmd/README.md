# Pacote `cmd` (Comandos da CLI do Portuscript)

O pacote `cmd` é o núcleo de controle da Interface de Linha de Comando (CLI) do **Portuscript**. Ele serve como ponte de integração estrutural entre a inicialização inicial do executável (no ponto de entrada `main.go`) e a biblioteca de gerenciamento de comandos **Cobra** (`github.com/spf13/cobra`).

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
6. [Exemplo de Uso Prático](#-exemplo-de-uso-prático)

---

## 🎯 Filosofia de Design

A CLI do Portuscript foi desenhada para colocar a língua portuguesa em primeiro lugar. Por isso, toda a experiência no terminal:
- Descrições curtas (`Short`) e longas (`Long`) dos comandos;
- Mensagens explicativas de help/ajuda;
- Diagnósticos e mensagens de erro do sistema;
- Aliases (apelidos) de comandos (ex: `exec` como apelido para `executar`);
são escritos nativamente em **Português (PT-BR)**, criando consistência com o propósito educacional e de acessibilidade da linguagem Portuscript.

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

| Variável | Tipo | Valor Padrão (Local) | Propósito de Negócio |
| :--- | :--- | :--- | :--- |
| `Commit` | `string` | `"-"` | Identifica a hash curta (SHA-1) do Git referente ao código fonte usado na geração deste binário. Utilizado para auditoria técnica precisa de bugs. |
| `Datetime` | `string` | `"0000-00-00T00:00:00"` | Armazena o carimbo de data/hora (formato ISO-8601) que marca o instante exato em que a build do binário foi executada. |
| `Version` | `string` | `"dev"` | Armazena a versão semântica oficial da release (ex: `0.3.0`). Usada pelo módulo `atualize` para comparar com o GitHub e verificar se há novas atualizações disponíveis. |

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

O subcomando `executar` (apelidado de `exec`) é o ponto de entrada primário para a execução física de códigos na máquina virtual Portuscript. 

### Mecânica de Funcionamento

O interpretador inicializa o ambiente importando implicitamente a biblioteca padrão e configurando as variáveis de escopo local. No arquivo `executar.go`, a importação blank (`_`) é crítica:

```go
import _ "github.com/natanfeitosa/portuscript/stdlib"
```

> **Por que isso é necessário?**  
> A importação anônima força a execução das funções `init()` de todos os arquivos no pacote `stdlib`. Isso registra dinamicamente todos os módulos internos e funções embutidas (como `escreva()`, `leia()`, etc.) na tabela global de símbolos do interpretador antes de qualquer código de usuário rodar.

### Fluxo de Decisão do interpretador

Quando o usuário digita `portuscript executar [arquivo] [flags]`, o seguinte algoritmo de decisão é acionado em `executar.go`:

```
                    [Início do Comando Executar]
                                 │
                 Obtém diretório atual (os.Getwd())
                                 │
                   Cria novo Contexto de Execução
                 (ptst.NewContexto com diretório local)
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
               │       Executa ptst.ExecutarArquivo()
               │       Se falhar: chama ptst.LancarErro() e encerra.
               │
               └─► [2] Se há flag -c/--codigo (script inline):
                       Executa ptst.ExecutarString()
                       Se falhar: chama ptst.LancarErro() e encerra.
```

1. **Obtenção do Diretório de Trabalho**: A chamada para `os.Getwd()` define qual o diretório corrente do processo do usuário. Esse caminho é adicionado à lista de `CaminhosPadrao` do contexto da VM, permitindo que a importação de módulos locais (`importar ...`) seja resolvida corretamente a partir do local de chamada.
2. **Ciclo de Vida do Contexto**: O contexto da VM (`ptst.Contexto`) é protegido por `defer ctx.Terminar()`. O método `Terminar()` limpa a memória, remove referências circulares, realiza o flush (escoamento) de buffers de escrita ativos e finaliza o coletor interno, prevenindo memory leaks.
3. **Precedência de Execução**: Se o usuário fornecer um arquivo E código inline (`-c`), o arquivo é executado **primeiro**. O código inline é avaliado em seguida no mesmo contexto. Esse comportamento permite usar arquivos locais como scripts de "setup" de variáveis e ambientes antes de rodar comandos de teste rápidos passados inline no terminal.
4. **Tratamento Amigável de Falhas**: Exceções e erros levantados pela VM são capturados e tratados via `ptst.LancarErro()`. Em vez de simplesmente exibir um stacktrace feio do Go, o Portuscript formata um traceback amigável em português com ponteiros visuais (sublinhado) indicando exatamente o token incorreto ou a linha física falha.

---

## 🔄 Subcomando: `atualize`

O subcomando `atualize` automatiza o download e a instalação de novas releases binárias oficiais diretamente do repositório no GitHub.

### Mapeamento de Sistema e Arquitetura

Como os arquivos binários compilados pelo GoReleaser no GitHub seguem uma nomenclatura padronizada de SO e arquitetura de CPU, o módulo realiza a conversão das variáveis de runtime do Go (`runtime.GOOS` e `runtime.GOARCH`) para coincidir com as nomenclaturas de distribuição usando as funções utilitárias:

- **`nomeOS() string`**:
  - `darwin` ➔ `Darwin` (Evita erros comuns como o typo "Darwind")
  - `linux` ➔ `Linux`
  - `windows` ➔ `Windows`
  - *Fallback*: Retorna `Linux` por segurança operacional.
- **`nomeArch() string`**:
  - `amd64` ➔ `x86_64` (Compatibilidade padrão Unix)
  - `386` ➔ `i386`
  - *Fallback*: Retorna a própria string gerada pelo runtime (ex: `arm64`).

---

### Fluxo de Atualização Semântica

Ao acionar `portuscript atualize`, o interpretador executa as seguintes etapas estruturadas:

```
                  [Comando portuscript atualize]
                                 │
                   Identifica Home do Usuário
                                 │
              Monta caminho do binário local executável
             (~/.portuscript/bin/portuscript[.exe])
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

1. **Localização**: Descobre o diretório de instalação padrão montando o caminho absoluto sob a pasta home do usuário (`~/.portuscript/bin/portuscript` ou `~/.portuscript/bin/portuscript.exe` se for Windows).
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
    url := "https://github.com/natanfeitosa/portuscript/releases/latest/download/"
    url += nomeOS() + "_" + nomeArch()
    if isWindows() { return url + ".zip" }
    return url + ".tar.gz"
}
```

1. **Download Seguro**: Cria um arquivo temporário em disco protegido com prefixo `-ptst` no diretório padrão temporário do sistema operacional (limpo e apagado ao final via `defer os.Remove(...)`).
2. **Invocação do `curl`**: Executa o comando utilitário `curl` do sistema de forma nativa para realizar o download com suporte a redirecionamento (`--location`), interrupção caso haja falhas HTTP (`--fail`) e exibindo uma barra de progresso elegante no próprio terminal (`--progress-bar`).
3. **Descompactação Inteligente**: 
   - **Sistemas Unix (Linux/macOS)**: Utiliza o utilitário nativo de sistema `tar` (`tar -xf [arquivo] -C [destino]`) para extrair e sobrescrever o binário diretamente em `~/.portuscript/bin/`.
   - **Windows**: Tenta primeiramente invocar o utilitário `unzip` do sistema. Se este não for localizado na variável de ambiente PATH, faz um fallback e tenta acionar a ferramenta `7z` (7-Zip) com parâmetros automáticos silenciosos de sobrescrita.

---

## 🛠️ Exemplo de Uso Prático

### Executando Códigos Através do Terminal

#### Rodando um arquivo físico `.pt`:
```bash
# Executa um script local chamado ola_mundo.pt
portuscript executar ola_mundo.pt
```

#### Rodando código de forma rápida inline (flag `-c`):
```bash
# Executa o interpretador sem criar arquivos locais
portuscript executar -c "escreva('Olá, Portuscript!')"
```

#### Executando arquivo local e complementando com código inline:
```bash
# Roda as definições contidas em setup.pt e executa a função de teste inline
portuscript executar setup.pt -c "inicializarSessao()"
```

#### Rodando o Playground (TUI Interativo):
```bash
# Sem parâmetros, inicia o modo REPL console interativo
portuscript executar
```

### Atualizando o Interpretador

```bash
# Verifica se há novas versões estáveis no GitHub e instala se houver
portuscript atualize
```
