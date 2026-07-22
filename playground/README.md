# Pacote `playground` (TUI e REPL Interativo do Harpia)

O pacote `playground` é o responsável por fornecer a Interface de Usuário de Terminal (TUI) e o ambiente **REPL** (Read-Eval-Print Loop) interativo do **Harpia**.

Ele permite que programadores testem expressões, declarem variáveis, criem funções e experimentem os recursos da linguagem em tempo real diretamente no console, de forma rápida e intuitiva, sem a necessidade de criar arquivos físicos no disco.

---

## 📖 Índice

1. [Visão Geral e Banner](#-visão-geral-e-banner)
2. [Máquina de Estado do Prompt (Entrada Multilinha)](#-máquina-de-estado-do-prompt-entrada-multilinha)
   - [Mecânica de Delimitadores Abertos](#mecânica-de-delimitadores-abertos)
3. [Executor de Expressões (Persistência de Escopo)](#-executor-de-expressões-persistência-de-escopo)
4. [Persistência de Histórico de Comandos](#-persistência-de-histórico-de-comandos)
5. [Injeção Dinâmica de Funções Auxiliares](#-injeção-dinâmica-de-funções-auxiliares)
6. [API de Métodos e Estruturas Internas](#-api-de-métodos-e-estruturas-internas)
7. [Diagrama do Ciclo do Loop do REPL](#-diagrama-do-ciclo-do-loop-do-repl)
8. [Exemplos de Interações no Console](#-exemplos-de-interações-no-console)

---

## 🌟 Visão Geral e Banner

Ao acionar o Harpia sem parâmetros de arquivos (ex: digitando apenas `harpia executar`), o console REPL é acionado. Na inicialização, a função pública `Inicializa()` é disparada, imprimindo um banner com metadados injetados de versão e compilação do binário:

```
Bem vindos ao Harpia v0.3.0.

(2026-07-14T00:00:00Z) [abcdef]
>>>
```

---

## 🔄 Máquina de Estado do Prompt (Entrada Multilinha)

O arquivo `estado.go` implementa um mecanismo inteligente de análise gramatical preliminar para suportar comandos que se estendem por múltiplas linhas física de código (como laços `enquanto`, condicionais `se`, listas e mapas).

### Mecânica de Delimitadores Abertos

O interpretador REPL conta o pareamento de caracteres delimitadores principais:

- Colchetes: `[` e `]`
- Parênteses: `(` e `)`
- Chaves: `{` e `}`

#### Regra de Transição de Estado:

```
                                [Usuário envia linha]
                                          │
                                 RecalcularEstado()
                                          │
                    strings.Count(abertura) > strings.Count(fechamento)?
                                          │
                        ┌─────────────────┴─────────────────┐
                        ▼ sim                               ▼ não
                Ativa estado.Continua               Desativa estado.Continua
                        │                                   │
                Prompt muda para "... "              Prompt muda para ">>> "
              VM aguarda mais digitação             VM executa código acumulado
```

Se o número de símbolos de abertura for **maior** do que os de fechamento correspondentes, o REPL entende que a declaração está incompleta e não tenta executá-la. Ele muda o prompt visual do usuário de `>>> ` para `... ` e adiciona a nova linha ao buffer temporário `Estado.Codigo`.

Assim que o usuário fechar todos os delimitadores pendentes (igualando a contagem de abertura e fechamento), o prompt retorna para o modo de indicação normal `>>> ` e envia o buffer completo para a VM avaliar.

---

## ⚡ Executor de Expressões (Persistência de Escopo)

A execução em REPLs comuns sofre de "perda de memória" se cada comando for rodado em um ambiente estéril. O Harpia resolve isso em `executor.go` virtualizando um arquivo persistente.

### O Arquivo Virtual `<playground>`

No momento em que o playground inicia, a função `NovoExecutor()` instancia uma estrutura `Executor` que cria um módulo especial virtualizado no interpretador sob o arquivo físico inexistente `<playground>`:

```go
exec.Modulo, _ = ctx.InicializarModulo(&hrp.ModuloImpl{
    Info: hrp.ModuloInfo{
        Arquivo: "<playground>",
    },
})
```

Toda instrução digitada pelo usuário é compilada para AST e avaliada usando o **mesmo escopo** (`e.Modulo.Escopo`).

> **Vantagem de Negócio**:  
> Graças a essa persistência, se o usuário digitar `a = 10` na primeira linha, a variável `a` será gravada na tabela de símbolos do escopo do módulo `<playground>`. Na linha seguinte, se ele digitar `escreva(a)`, a variável estará acessível e o valor `10` será impresso com sucesso.

---

## 💾 Persistência de Histórico de Comandos

O playground utiliza a biblioteca externa **Liner** (`github.com/peterh/liner`) para controlar o terminal, fornecendo uma experiência profissional:

- Edição em linha (uso de Backspace, Delete e posicionamento de cursor com setas `←` e `→`);
- Histórico de comandos anteriores usando as setas `↑` e `↓`;
- Atalhos comuns de terminal (como `Ctrl+A` para ir ao início ou `Ctrl+E` para ir ao fim).

### Ciclo de Persistência em Disco

O histórico de comandos não é perdido ao fechar o REPL. Ele é gravado de forma transparente no disco no arquivo oculto `~/.historico_harpia` (localizado sob a pasta Home do usuário):

1. **Abertura/Leitura**: Ao iniciar (`Inicializa`), o playground localiza a Home do usuário corrente, abre ou cria o arquivo `.historico_harpia` e popula o buffer de histórico da biblioteca Liner via `line.ReadHistory()`.
2. **Registro de Comandos**: A cada linha não-vazia digitada pelo usuário, o REPL chama `line.AppendHistory()` para incluir o comando na pilha de histórico ativa.
3. **Escrita/Persistência**: Ao encerrar o terminal (por Ctrl+D ou comando `sair()`), o REPL garante a gravação de todos os comandos de volta para o disco rígido chamando `line.WriteHistory()`, mantendo o histórico de digitação intacto para futuras sessões de programação.

---

## 💉 Injeção Dinâmica de Funções Auxiliares

O playground injeta métodos de utilidade específicos no escopo de execução que não existem por padrão em scripts comuns de arquivos.

### O Método `sair()`

O principal método injetado é a função `sair()`.

- **Implementação**:
  ```go
  exec.RegistrarMetodo(hrp.NewMetodoOuPanic("sair", func(_ hrp.Objeto, args hrp.Objeto) (hrp.Objeto, error) {
      finalizar()
      return nil, nil
  }, ""))
  ```
- **Comportamento**: A chamada para `sair()` no console altera uma flag booleana local (`finalizou = true`), fazendo com que o loop principal do terminal termine de maneira limpa, salvando o histórico e encerrando o processo graciosamente.

---

## 💻 API de Métodos e Estruturas Internas

O ecossistema interno do `playground` está dividido estruturalmente em três arquivos Go e um arquivo Harpia, detalhados abaixo:

### 1. Orquestrador do Loop REPL (`playground.go`)

*   **`Inicializa(ctx *hrp.Contexto, version, datetime, commit string)`**:
    *   **Função**: Orquestra a montagem, exibição do banner e o loop iterativo principal de captura e avaliação do console.
    *   **Parâmetros**:
        *   `ctx *hrp.Contexto`: O contexto unificado de execução da VM do Harpia.
        *   `version string`, `datetime string`, `commit string`: Metadados injetados no build do compilador.
*   **`homeDirectory() string`**:
    *   **Função**: Obtém de forma resiliente o caminho absoluto da pasta home do usuário utilizando `user.Current()` com fallback na variável de ambiente `$HOME`.
*   **`ArquivoHistorico(escrita bool) *os.File`**:
    *   **Função**: Abre ou cria o arquivo oculto de histórico `~/.historico_harpia` no diretório home do usuário. Se `escrita == true` abre em modo append; caso contrário, em modo leitura.

### 2. Máquina de Estado do Buffer (`estado.go`)

*   **`struct Estado`**:
    *   **Campos**:
        *   `Indicador indicador`: Prompt visual corrente (`Normal` ou `Continua`).
        *   `Continua bool`: Sinaliza se existem blocos de código pendentes de fechamento.
        *   `Codigo string`: Buffer de string acumulado.
*   **`NewEstado() *Estado`**:
    *   **Função**: Cria e retorna uma nova instância limpa de monitoramento do prompt REPL.
*   **`RecalcularEstado(cod string)`**:
    *   **Função**: Anexa a nova entrada ao buffer de código e atualiza as flags de continuação e indicadores visuais com base no balanceamento de símbolos.
*   **`continuaEmNovaLinha(abre, fecha string) bool`**:
    *   **Função**: Retorna verdadeiro se a contagem do delimitador `abre` for maior que `fecha` no buffer.

### 3. Máquina Virtual e Escopo Persistente (`executor.go`)

*   **`struct Executor`**:
    *   **Campos**:
        *   `Contexto *hrp.Contexto`: Referência à VM do Harpia.
        *   `Modulo *hrp.Modulo`: Módulo virtual `<playground>` com tabela de símbolos persistentes.
*   **`NovoExecutor(ctx *hrp.Contexto) *Executor`**:
    *   **Função**: Construtor padrão que inicializa a estrutura do executor com o escopo do módulo virtual.
*   **`ExecutarCodigo(codigo string)`**:
    *   **Função**: Compila a string de código para AST e a executa sob o escopo isolado e persistente do módulo virtual do playground, imprimindo retornos ou erros no terminal.
*   **`RegistrarMetodo(metodo *hrp.Metodo) error`**:
    *   **Função**: Permite a injeção dinâmica de métodos nativos do Go no escopo global da sessão interativa (como `sair()`).
*   **`Terminar()`**:
    *   **Função**: Libera caches e encerra a VM interna do interpretador.

---

## 🔄 Diagrama do Ciclo do Loop do REPL

O ciclo operacional completo do REPL, integrando a coleta de dados da biblioteca Liner, o controle de estado e a execução física na máquina virtual do Harpia pode ser sumarizado no seguinte fluxo estrutural:

```
    +--------------------------------------------------+
    |           [Início: Inicializa()]                 |
    +--------------------------------------------------+
                             |
                             v
             Carrega histórico de comandos
             e inicia console interativo Liner
                             |
                             v
+---> [Prompt Liner: Aguarda entrada do programador]
|                            |
|                            v
|                     Recebe a linha
|                            |
|                            v
|           Saves à pilha de histórico (Liner)
|                            |
|                            v
|         Acumula e analisa no Estado (estado.go)
|             RecalcularEstado(codigo_linha)
|                            |
|                            v
|                [Existe bloco em aberto?]
|                     /            \
|                    /              \
|             Sim   /                \ Não
|                  v                  v
|         Prompt muda para "... "    Prompt retorna para ">>> "
|         Aguarda próxima linha      VM executa o código acumulado
|                  |                  ExecutarCodigo(estado.Codigo)
|                  |                  Zera o buffer de código
|                  |                         |
+------------------+-------------------------+
                             |
                             v
                   [Comando foi sair()?]
                     /            \
                    /              \
             Sim   /                \ Não (Ctrl+D / Erro)
                  v                  v
         Encerra o Loop TUI        Continua no Loop de prompt
         e salva o histórico
         de comandos no disco
```

---

## 🛠️ Exemplos de Interações no Console

Abaixo estão alguns exemplos práticos que demonstram o comportamento do estado e da persistência de escopo do playground do Harpia:

### 1. Declarações Simples e Persistência de Escopo

```
>>> x = 42
>>> escreva(x)
42
```

### 2. Declarações Multilinha (Controle de Prompt `... `)

Se o usuário declarar uma lista ou abrir um bloco estruturado como uma função, o prompt se transforma automaticamente:

```
>>> lista = [
...   "Harpia",
...   "Linguagem",
...   "Brasileira"
... ]
>>> escreva(lista)
['Harpia', 'Linguagem', 'Brasileira']
```

No exemplo acima, o terminal permaneceu em modo contínuo (`... `) nas linhas 2, 3 e 4 porque a abertura de colchetes na linha 1 não havia sido pareada com o fechamento correspondente. Ao digitar `]` na linha 5, o loop processou e avaliou toda a expressão.

### 3. Tratamento de Erros Sem Interrupção

Se houver erros no código digitado, o REPL imprime o erro detalhado mas continua ativo para novas tentativas:

```
>>> escreva(y)
Erro de Nome: o nome 'y' não foi definido
>>> y = 100
>>> escreva(y)
100
```

### 4. Saindo do Ambiente

Para sair, basta chamar a função embutida `sair()` ou interromper via teclado com `Ctrl+D`:

```
>>> sair()
Saindo...
```

---

## 🖥️ TUI Bubbletea de Alta Fidelidade (`harpia tui`)

O Harpia disponibiliza uma Interface de Usuário de Terminal (TUI) rica e didática desenvolvida sobre o ecossistema **Bubbletea** em Go:

- **Painéis Divididos**: A tela de terminal exibe simultaneamente:
  1.  _Console/Editor_: Digite seu código Harpia de forma interativa.
  2.  _Inspetor de Escopo/VM_: Visualize em tempo real as variáveis locais criadas e seus tipos.
  3.  _Saída/Erros_: Logs de saída de terminal e erros humanos ricos.
- **Navegação e Foco Estético**: Alterne facilmente o foco entre o editor e o inspetor utilizando a tecla `Tab`. O painel ativo ganha destaque visual com bordas de cor personalizada do Lipgloss.
- **Depurador Passo-a-Passo (F7/F8)**: Pressione `F8` para iniciar uma sessão de depuração síncrona, e aperte `F7` para executar o código linha por linha, acompanhando a atualização síncrona das variáveis locais do escopo na VM em tempo real.

---

## 🌐 Dogfooding Supremo: Playground Web escrito em Harpia SPA!

O playground interativo baseado em navegador (`harpia playground`) é o exemplo definitivo de maturidade e poder da linguagem, sendo **integralmente desenvolvido em Harpia Reativo**:

- **Lógica e Layout (`playground/interface.hrp`)**: Todo o design, formulário de escrita de código, e lógica de atualização são um aplicativo SPA Harpia.
- **Interações Reativas**: Utiliza Two-Way Binding síncrono (`_ligar`) para monitorar o código digitado e as tabelas inteligentes `<GradeDeDados>` em português do `runtime-web.js` para renderizar as variáveis locais e constantes retornadas de forma reativa.
- **Transpilação On-Demand**: O servidor Go compila dinamicamente esse arquivo `.hrp` para JavaScript e CSS em tempo de refresh (<5ms) servindo os arquivos finais na rota raiz `/` com suporte a renderização de erros estruturados com sublinhado ANSI-to-HTML no VDOM.

---

## 🛡️ Segurança e Blindagem Corporativa (Security Audit)

O playground e as ferramentas de rede do Harpia foram submetidos a uma auditoria de segurança rigorosa e contam com as seguintes defesas nativas:

1.  **Sincronização de Execuções Concorrentes (Anti-Stdout Race)**: O endpoint `/api/executar` utiliza bloqueio de exclusão mútua síncrono (`sync.Mutex`) para gerenciar as execuções de código. Isso impede que duas requisições simultâneas causem corridas de dados ao interceptar a variável global `os.Stdout`, garantindo isolamento total de logs entre sessões de usuários concorrentes.
2.  **Mitigação de DoS e OOM**: O tamanho do payload de código enviado ao servidor é limitado síncronamente a no máximo 1MB via `http.MaxBytesReader`, impedindo ataques de negação de serviço ou estouro de memória física.
3.  **Timeouts de Rede Rígidos**: O servidor HTTP do playground utiliza limites estritos de `ReadTimeout` de 5 segundos e `WriteTimeout` de 10 segundos, mitigando riscos de estouro de conexões e ataques de Slowloris.
