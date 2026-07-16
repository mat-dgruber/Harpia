# 🇧🇷 Manual de Referência Oficial do Portuscript

Bem-vindo ao **Manual de Referência Oficial do Portuscript**, uma especificação exaustiva, de nível de engenharia, que documenta todos os subsistemas, tipos primitivos, biblioteca padrão (stdlib), compilador, analisadores, runtime e filosofia de design da linguagem **Portuscript**.

Este d

ocumento foi elaborado para servir tanto como um guia definitivo para desenvolvedores que escrevem programas em Portuscript quanto para engenheiros de sistemas que contribuem com a evolução de sua máquina virtual e compilador.

---

## 📖 Índice Geral

1. [Filosofia de Design e Arquitetura](#1-filosofia-de-design-e-arquitetura)
2. [Interface de Linha de Comando (CLI)](#2-interface-de-linha-de-comando-cli)
3. [Análise Léxica (Lexer)](#3-análise-léxica-lexer)
4. [Análise Sintática (Parser &amp; AST)](#4-análise-sintática-parser--ast)
5. [Máquina Virtual e Runtime (ptst)](#5-máquina-virtual-e-runtime-ptst)
6. [Tipos de Dados Primitivos](#6-tipos-de-dados-primitivos)
7. [Biblioteca Padrão (Stdlib)](#7-biblioteca-padrão-stdlib)
8. [Recursos Avançados da Linguagem](#8-recursos-avançados-da-linguagem)
9. [Arcabouço de Testes Nativos (TDD)](#9-arcabouço-de-testes-nativos-tdd)
10. [Diagnósticos e Tratamento de Erros Ricos](#10-diagnósticos-e-tratamento-de-erros-ricos)
11. [Console Interativo (REPL / Playground)](#11-console-interativo-repl--playground)
12. [Guia de Sintaxe Rápida e Exemplos de Produção](#12-guia-de-sintaxe-rápida-e-exemplos-de-produção)

---

## 1. Filosofia de Design e Arquitetura

O Portuscript foi construído sob uma **perspectiva de design dupla**:

1. **Ponte de Aprendizado:** Facilitar a transição suave de estudantes de programação no Brasil para linguagens de mercado (como JavaScript, Python, Go e C++), utilizando sintaxes modernas (blocos por chaves `{}`, escopo léxico estrito, corotinas assíncronas e tipagem dinâmica opcional) inteiramente em português.
2. **Poder e Identidade Própria:** Ser uma linguagem real de produção. Não é um mero tradutor de códigos ou interpretador didático lento. Oferece reatividade nativa de alto desempenho via **Sinais**, uma biblioteca de socket de baixo nível, suporte a componentes do tipo JSX, um gerador de diagramas de arquitetura e um interpretador robusto com suporte a plugins compartilhados binários em Go ou C/C++ (`.so`).

### Estrutura Geral do Workspace de Diretórios

```
portuscript/
├── cmd/               -> Comandos de terminal da CLI (Cobra)
├── compartilhado/     -> Utilitários Unicode (UTF-8) e de casting de strings
├── gramatica/         -> Especificação formal da linguagem (ANTLR4 .g4)
├── lexer/             -> Analisador léxico escrito à mão em Go
├── parser/            -> Analisador sintático de descida recursiva à mão
├── playground/        -> Console interativo de terminal REPL (Liner)
├── ptst/              -> Núcleo do runtime (VM, tipos de dados, escopos)
├── stdlib/            -> Biblioteca padrão da linguagem (embutidos, matematica, etc.)
└── exemplos/          -> Demonstrações práticas e módulos externos
```

---

## 2. Interface de Linha de Comando (CLI)

O utilitário de terminal do Portuscript foi construído usando a biblioteca **Cobra** (`github.com/spf13/cobra`). Toda a interface se comunica em português de forma natural.

### Variáveis Globais de Build (Injeção via Linker)

O pipeline de CI/CD (usando GoReleaser) injeta metadados na compilação do executável do pacote `cmd` através das variáveis:

* `Commit`: Hash SHA-1 curta do commit Git que gerou o build.
* `Datetime`: Carimbo ISO-8601 que registra o instante de build.
* `Version`: Versão SemVer estável da release (ex: `0.3.1`). Se for compilado manualmente, assume o valor `"dev"`.

### Comandos Suportados

#### 1. `portuscript` (ou `portuscript executar` sem argumentos)

Abre o REPL interativo com realce de sintaxe e controle de buffers multilinha.

#### 2. `portuscript executar [arquivo.pt] [flags]` (Alias: `exec`)

Interpreta e executa um script físico.

* **Ordem de Carregamento**: Se uma string for fornecida pela flag `-c "codigo"`, o interpretador prioriza a execução do arquivo posicional e, em seguida, avalia o fragmento de código inline no mesmo contexto de execução.
* **Flag `-c`, `--codigo`**: Executa um código direto no terminal (ex: `portuscript executar -c "imprima('Olá!')"`).

#### 3. `portuscript testar [caminho]`

Varre recursivamente o diretório em busca de arquivos com extensões `.pt` ou `.ptst` e executa de forma isolada todos os blocos `testar` nativos definidos nos scripts, apresentando um relatório consolidado com o total de sucessos e falhas.

#### 4. `portuscript atualize`

Executa o auto-update do executável a partir do repositório no GitHub.

* **Algoritmo de Resolução**: Monta o caminho de instalação sob o diretório do usuário (`~/.portuscript/bin/portuscript`). Compara a versão local (executando o binário com `-v`) com a última tag disponível via API do GitHub usando a biblioteca `semver/v3`. Se houver atualizações, usa o `curl` para baixar o binário comprimido adequado para a arquitetura do cliente (mapeando de forma inteligente arquiteturas como `amd64` para `x86_64` e SOs como `darwin` para `Darwin`) e o extrai. Se a versão local for `"dev"`, o processo de atualização automática é impedido para preservar builds de desenvolvimento.

#### 5. `portuscript doc [entrada] [flags]`

Varre um diretório ou arquivo extraindo comentários iniciados com três barras (`///`) de funções, classes e métodos, gerando documentação estruturada exportada em formato Markdown (`--formato=markdown`) ou HTML (`--formato=html`).

#### 6. `portuscript empacotar --entrada=[arquivo] --saida=[binario] [flags]`

Empacota um script Portuscript e todos os seus recursos em um executável binário autônomo (Single Binary Bundle) sem dependências externas compilando dinamicamente o código Go subjacente via `go build` com suporte a cross-compilation (`--so` e `--arq`).
* **Suporte a WebAssembly (WASM)**: Se `--so=js` e `--arq=wasm` forem especificados, o comando compila o interpretador completo para WebAssembly (`docs/portal/portuscript.wasm`) e extrai o carregador JavaScript portátil `wasm_exec.js` correspondente do GOROOT do sistema.

#### 6.1. `portuscript diagramar [diretorio] [flags]`

Analisa recursivamente a estrutura física do projeto para mapear e validar a hierarquia de importações entre as camadas do Clean Architecture.
* **Flags**: `--formato` ou `-f` (`mermaid`, `html`, `svg`), `--saida` ou `-s`.
* **Diagrama Interativo**: Se o formato for `html` (ou `svg`), gera um arquivo HTML standalone contendo o visualizador interativo Mermaid.js que colore de verde as importações válidas, de **vermelho grossa as violações arquiteturais**, e emite um botão para exportar diretamente o arquivo `.svg` correspondente.

#### 6.2. `portuscript instalar [nome-do-pacote] [versao-opcional]`

Gerenciador de pacotes e dependências assíncrono para o ecossistema Portuscript.
* **Resolução Remota Semver**: Permite baixar pacotes públicos e resolver restrições de versão semver (ex: `banco-dados: 1.0.0`) diretamente de um registro JSON remoto central em português, gravando o módulo na pasta local `pt_modulos/`.

#### 7. `portuscript stressar [arquivo] [flags]`

Utilitário CLI interno para benchmarking e testes de estresse concorrentes de aplicações locais ou remotas escritas em Portuscript, detalhando estatísticas de tempo médio, mínimo, máximo e taxa de sucesso.

#### 8. `portuscript depurar [flags]`

Inicializa o servidor TCP nativo compatível com o protocolo Debug Adapter Protocol (DAP) na porta `4711` (ou customizada via `--porta`), viabilizando a depuração interativa integrada com editores modernos (VS Code).

#### 9. `portuscript crie [rota | componente | modelo] [nome]`

Assistente interativo de scaffolding que gera templates estruturados de arquivos seguindo os padrões de Clean Architecture e DDD definidos para o ecossistema Portuscript.

---

## 3. Análise Léxica (Lexer)

O pacote `lexer` foi escrito inteiramente à mão em Go. Ele evita o uso de expressões regulares ou geradores automáticos para garantir a máxima velocidade de varredura e o tratamento preciso de strings Unicode multibyte.

### O Desafio de Strings UTF-8 em Go

Em Go, strings são slices de bytes UTF-8. Um único caractere Unicode (acentos ou emojis) pode ocupar entre 1 e 4 bytes. Acessos diretos por índice (ex: `str[i]`) podem quebrar runas ao meio.

* **Solução do Portuscript**: O arquivo `compartilhado/strings.go` implementa a função `IndiceBytePorCarater(str string) []int`. Ela varre a string decodificando runas via `utf8.DecodeRuneInString` e pré-calcula uma tabela de mapeamento. Desse modo, o Lexer consegue fazer conversões e fatiamentos de caracteres de forma segura e rápida em tempo constante $O(1)$.
* **Cache Estático Thread-Safe Global**: Para suportar múltiplos interpretadores independentes rodando em paralelo sem colisões, o pacote `compartilhado` adota uma tabela de cache global protegida por um `sync.RWMutex`. Entradas são restritas a tamanhos menores que 4KB para evitar consumo excessivo de heap, e o cache inteiro é reciclado se ultrapassar 2048 registros, prevenindo estouros de memória.

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

* **Estrutura Condicional e de Fluxo**: `se`, `senao`, `enquanto`, `para`, `em`, `retorne`, `pare`, `continue`
* **Definições e Escopos**: `var`, `const`, `func`, `funcao`, `classe`, `estende`, `self`, `estatico`
* **Módulos**: `importe`, `de`
* **Testes e Garantias**: `testar`, `assegura`
* **Constantes e Operadores**: `Verdadeiro`, `Falso`, `Nulo`, `ou`, `e`, `nao`, `nova`
* **Controle de Erros**: `tente`, `capture`, `finalmente`

---

## 4. Análise Sintática (Parser & AST)

O analisador sintático (`parser/parser.go`) é um **Parser de Descida Recursiva Manual** (*Manual Recursive Descent Parser*). Ele consome tokens lineares e monta a **Árvore de Sintaxe Abstrata (AST)**.

### Precedência e Hierarquia de Operadores

A descida recursiva força uma prioridade de resolução estrita. Os operadores são avaliados do nível de menor prioridade (resolvidos por último) até os de maior prioridade (resolvidos primeiro):

|    Nível    | Operação / Categoria     | Operadores Relacionados                                                |
| :----------: | :------------------------- | :--------------------------------------------------------------------- |
| **11** | Encadeamento Funcional     | `\|>` (Pipes)                                                         |
| **10** | Disjunção Lógica        | `ou`                                                                 |
| **9** | Conjunção Lógica        | `e`                                                                  |
| **8** | Negação Lógica          | `nao`                                                                |
| **7** | Comparadores Relacionais   | `==`, `!=`, `<`, `<=`, `>`, `>=`, `em`, `instancia de` |
| **6** | OU Bit a Bit               | `\|` (Bitwise OR)                                                     |
| **5** | XOR Bit a Bit              | `^` (Bitwise XOR)                                                    |
| **4** | E Bit a Bit                | `&` (Bitwise AND)                                                    |
| **3** | Deslocamento de Bits       | `<<`, `>>` (Bitwise Shifts)                                        |
| **2** | Soma e Subtração         | `+`, `-` (Concatenação textual também no `+`)                 |
| **1** | Multiplicação e Divisão | `*`, `/`, `//` (divisão inteira), `%` (resto/módulo)         |
| **0** | Sinais e Exponenciação   | `+`, `-`, `~` (unários); `**` (exponenciação)               |

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

O Portuscript permite omitir o uso de ponto e vírgula. O analisador trata `\n` (quebras de linha) e `EOF` (fim de arquivo) como delimitadores implícitos de instrução. A verificação é unificada em `consome(";")`:

* Se o token corrente for de fato `";"`, consome-o e avança.
* Se for uma nova linha ou término de arquivo, valida a instrução como completa sem reclamar, garantindo um código limpo estilo Python ou Go.

---

## 5. Máquina Virtual e Runtime (ptst)

O pacote `ptst` gerencia a infraestrutura matemática, lógica, as tabelas de símbolos e a execução física da AST.

### Interface Primordial `Objeto`

Toda variável ou estrutura na VM do Portuscript satisfaz a interface polimórfica:

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

* **Garantia de Montagem Consequente**: Para assegurar a resolução correta de heranças em tempo de carregamento, cada `Tipo` inicializado em Go é enfileirado na lista global `filaMontagem`. A VM dispara a rotina centralizada `MontaOsTipos()` antes de iniciar o processamento da AST, populando as tabelas e injetando as documentações na propriedade mágica `__doc__` de cada classe.

### Resolução de Métodos Mágicos via Reflexão (Reflection)

O Portuscript adota uma convenção estrita de nomenclatura: interfaces Go de protocolos mágicos iniciam com **`I`** e seus métodos com **`M`** (ex: `I__texto__` com `M__texto__()`).

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

* **Escopo Léxico (Lexical Scoping)**: Cada `Escopo` mantém um link de referência para seu escopo pai (`Pai *Escopo`).
* **Algoritmo de Busca (`ObterValor`)**: Busca primeiro na tabela local de símbolos. Se a chave não constar, sobe de forma recursiva investigando o escopo do pai. Se atingir a raiz primordial sem sucesso, verifica o módulo de embutidos antes de lançar o erro controlado `NomeErro` (PSC-0005).
* **Sincronização de Concorrência do Escopo**: Cada escopo de variáveis conta com seu próprio `sync.RWMutex` para sincronização fina de leitura e escrita concorrente. Isso previne colisões de mapas em Go durante execuções paralelas de corotinas em background que acessam variáveis comuns.
* **Locks de Grão Fino em Símbolos**: Símbolos individuais do runtime (`ptst.Simbolo`) contam com bloqueios de mutex específicos (`sync.RWMutex`) para ler e definir seu valor de forma atômica e segura.
* **Cooperação Segura com o Garbage Collector**: A listagem de símbolos pelo Garbage Collector utiliza o método seguro `ObterSimbolosSeguro()`, que gera uma cópia rasa estável da tabela de símbolos do escopo sob um lock de leitura, integrando com o algoritmo de varredura e quebra de ciclos de forma 100% thread-safe.

---

## 6. Tipos de Dados Primitivos

Todos os tipos de dados nativos no Portuscript possuem comportamentos específicos sob a VM:

### 1. `Inteiro` (int64)

* **Design**: Inteiro com sinal de 64 bits para evitar estouros aritméticos.
* **Casting**: `int(obj)` avalia o método mágico `__inteiro__`.
* **Coerção Booleana**: Retorna Falso se o valor for zero, e Verdadeiro do contrário.
* **Coerção Decimal**: Promove para `Decimal` quando somado, subtraído ou multiplicado por um membro do tipo `Decimal`.

### 2. `Decimal` (float64)

* **Design**: Número de ponto flutuante de dupla precisão (IEEE 754).
* **Representação Textual**: Se o valor numérico for inteiro (ex: `5.0`), o método `M__texto__()` anexa explicitamente `.0` ao texto para manter no console a distinção visual clara em relação aos Inteiros ordinários.

### 3. `Booleano` (bool)

* **Design**: Armazena as constantes globais estruturadas `Verdadeiro` ou `Falso`.
* **Casting de Inteiros**: `Verdadeiro` é coergido para `1` e `Falso` para `0` quando operado de forma aritmética.

### 4. `Texto` (string)

* **Design**: Cadeia imutável de caracteres UTF-8.
* **Comprimento Seguro**: O método `tamanho()` chama `utf8.RuneCountInString`, fornecendo a contagem de caracteres reais em vez de contagem de bytes físicos em disco.
* **Interpolação de Strings**: Implementada usando o operador de módulo `%`. Analisa e substitui de forma dinâmica marcadores de formatação:
  * `%i`: Formata para Inteiro.
  * `%d`: Formata para Decimal.
  * `%b`: Formata para Booleano.
  * `%s` (ou outro marcador): Formata chamando a representação textual do objeto.
* **Exemplo**: `"Eu tenho %i anos de idade e me chamo %s" % (23, "Carlos")`.

### 5. `Lista` (`[]Objeto` mutável)

* **Design**: Coleção indexada mutável ordenada de dados.
* **Métodos Embutidos**:
  * `adiciona(elemento)`: Insere um novo item no final.
  * `extende(outraLista)`: Concatena os elementos de outra coleção.
  * `remove(elemento)`: Busca, remove e retorna o elemento especificado.
  * `pop(indice?)`: Remove e retorna o item localizado no índice. Se omitido, assume o índice inicial `0`.
  * `indice(elemento)`: Retorna o índice da primeira ocorrência do item.
  * `limpa()`: Esvazia por completo a lista.

### 6. `Tupla` (`[]Objeto` imutável)

* **Design**: Coleção indexada ordenada e imutável de dados.
* **Imutabilidade**: Não fornece ou expõe métodos para mutabilidade ou alteração física após ser criada no script.

### 7. `Mapa` (`map[string]Objeto`)

* **Design**: Dicionário associativo do tipo chave-valor. As chaves são estritamente do tipo `Texto`.
* **Métodos Embutidos**:
  * `chaves()`: Retorna uma tupla imutável com todas as chaves registradas.
  * `valores()`: Retorna uma tupla contendo todos os valores dos objetos.
  * `atualizar(outroMapa, ignoreExistentes?)`: Copia e mescla os dados de outro mapa de forma mutável. Se `ignoreExistentes` for Verdadeiro, chaves repetidas não são sobrescritas.
* **Mecânica de Iteração**: O loop `para` sobre mapas retorna de forma consecutiva uma `Tupla` contendo o par `(chave, valor)`, simplificando a varredura e permitindo desestruturação fluida.

### 8. `Bytes` (`[]byte` mutável)

* **Design**: Array físico de bytes para controle de rede ou buffers de arquivo.
* **Comparações**: Permite comparações ricas (`==`, `!=`, `<=`, etc.) baseadas na contagem de bytes e no conteúdo literal através de `bytes.Equal` do Go.

### 9. `Nulo`

* **Design**: Representa a ausência física de valor. É do tipo de classe única `_Nulo`.

---

## 7. Biblioteca Padrão (Stdlib)

Os módulos nativos são desenvolvidos de forma desacoplada em Go. Eles se registram via funções `init()` acionadas por importações anônimas no arquivo agregador `stdlib/stdlib.go`.

### Módulo: `embutidos`

Símbolos e métodos injetados de forma global. Não requerem importação.

* `escreva(args...)` (Alias: `imprimir`): Concatena os argumentos textuais separando-os por espaço e exibe a mensagem na saída padrão.
* `leia(prompt?)`: Exibe o prompt textual se fornecido e pausa a VM aguardando digitação pelo usuário. Retorna sempre uma string.
* `tamanho(objeto)`: Retorna a contagem de elementos de coleções que implementam a interface `I__tamanho__`.
* `int(objeto)` / `texto(objeto)`: Construtores de coerção.
* `instanciaDe(obj, classes)`: Verifica se o objeto descende das classes informadas.
* `mesmoTipo(obj1, obj2)`: Compara as assinaturas de classe dos objetos.
* `tipo(obj)`: Retorna a representação de Tipo da classe do objeto.
* `doc(obj)`: Devolve o bloco explicativo de documentação (Docstring) do método ou classe.
* `sequencia(fim)` / `sequencia(inicio, fim, passo?)`: Retorna uma struct `SequenciaNumerica` que atua como um iterador numérico sob limites definidos, lançando `FimIteracao` ao término.

### Módulo: `matematica`

Recursos matemáticos de alta precisão. Requer `importar matematica`.

* **Constantes**: `matematica.PI`, `matematica.E`.
* **Métodos**:
  * `absoluto(n)`: Magnitude numérica sem sinal.
  * `piso(n)` / `teto(n)`: Arredondamento para baixo/cima.
  * `potencia(base, expoente)`: Calcula $base^{expoente}$.
  * `raiz(radicando, indice?)`: Calcula a raiz do número. Se o índice for omitido, calcula a raiz quadrada por potência fracionária de expoente ($radicando^{1.0/indice}$).

### Módulo: `sistema`

Acesso ao hardware e variáveis de ambiente. Requer `importar sistema`.

* `sistema.NOME`: String identificando o SO hospedeiro (`"darwin"`, `"linux"`, `"windows"`).
* `sistema.ARQUITETURA`: Tipo de arquitetura do processador (`"amd64"`, `"arm64"`).

### Módulo: `colorize`

Colorização de console com ANSI True Color de 24 bits. Requer `importar colorize`.

* **Objetos**: `colorize.TEXTO` (Foreground), `colorize.FUNDO` (Background).
* **Propriedades**: `colorize.SUPORTA` (Booleano dinâmico que detecta variáveis de escape como `NO_COLOR`).
* **Métodos**:
  * `converteRGB(r, g, b, background?)`: Retorna o código de escape ANSI correspondente.
  * `imprimac(args...)`: Imprime os argumentos com as cores aplicadas. Se o console não suportar cores, remove de forma limpa as sequências ANSI via expressão regular antes de imprimir.
* **Cores Mapeadas**: `vermelho`, `lima`, `azul`, `amarelo`, `agua`, `fuchsia`, `branco`, `preto` (disponíveis tanto em `TEXTO` quanto em `FUNDO`).
* **Exemplo**: `imprimac(colorize.TEXTO.azul(colorize.FUNDO.branco("Texto Colorido!")))`.

### Módulo: `arquivos`

Recursos e controle de sistema de arquivos e caminhos. Requer `de "arquivos" importe ...`.

* **Métodos**:
  * `ler(caminho)`: Lê o conteúdo de um arquivo em formato de texto.
  * `escrever(caminho, texto)`: Cria ou sobrescreve um arquivo gravando o texto especificado.
  * `acrescentar(caminho, texto)`: Adiciona o texto especificado ao final do arquivo.
  * `remover(caminho)`: Exclui o arquivo ou diretório especificado.
  * `renomear(origem, destino)`: Move ou altera o nome de um arquivo ou pasta.
  * `juntar(partes...)`: Concatena partes de caminhos físicos de arquivos de acordo com o SO.
  * `resolver(caminho)`: Devolve o caminho absoluto absoluto limpo.
  * `caminhar(caminho, callback)`: Varre recursivamente diretórios acionando a função de callback fornecida.

### Módulo: `json`

Serialização e desserialização de formato de dados JSON. Requer `de "json" importe ...`.

* **Métodos**:
  * `analisar(textoJson)`: Desserializa uma string JSON em estruturas nativas de dados do Portuscript (Lista, Mapa, Inteiro, Decimal, Booleano, Nulo).
  * `serializar(objeto)`: Converte estruturas de dados recursivas do Portuscript em string JSON representativa.

### Módulo: `yaml`

Serialização e desserialização de formato de dados YAML. Requer `de "yaml" importe ...`.

* **Métodos**:
  * `analisar(textoYaml)`: Desserializa uma string YAML em estruturas nativas de dados do Portuscript.
  * `serializar(objeto)`: Converte estruturas de dados do Portuscript em string YAML.

### Módulo: `xml`

Serialização e desserialização de formato de dados XML. Requer `de "xml" importe ...`.

* **Métodos**:
  * `analisar(textoXml)`: Desserializa uma string XML em estruturas nativas de dados do Portuscript.
  * `serializar(mapa, tagRaiz?)`: Converte um Mapa do Portuscript em string XML com a tag raiz opcional informada (padrão: "raiz").

### Módulo: `cripto`

Funções para criptografia, hashes e identificadores. Requer `de "cripto" importe ...`.

* **Métodos**:
  * `sha256(texto)`: Devolve o hash SHA-256 do texto fornecido em formato hexadecimal.
  * `codificarBase64(texto)`: Codifica um texto simples para o formato Base64.
  * `decodificarBase64(base64)`: Decodifica um texto de Base64 para o formato simples correspondente.
  * `uuid()`: Gera e retorna um identificador universal único (UUID v4) aleatório.

### Módulo: `http`

Protocolo de rede HTTP (Cliente e Servidor). Requer `de "http" importe ...`.

* **Classes**:
  * **`Servidor`**:
    * `obter(rota, handler)`: Registra um manipulador (handler) para requisições de método GET na rota. Aceita rotas dinâmicas com parâmetros nomeados, como `/ola/:nome`.
    * `postar(rota, handler)`: Registra um manipulador para o método POST na rota especificada.
    * `deletar(rota, handler)`: Registra um manipulador para o método DELETE na rota especificada.
    * `usar(middleware)`: Registra um middleware global (função `funcao(req, res)`) executado sequencialmente antes do handler de destino de cada requisição.
    * `escutar(porta)`: Inicia a escuta e aceitação de requisições na porta informada, operando de forma assíncrona e concorrente em background.
    * `fechar()`: Encerra a escuta do servidor HTTP liberando a porta local de forma limpa.
  * **`Requisicao`**:
    * Representa os metadados da requisição HTTP recebida. Atributos:
      * `metodo`: String que descreve o método HTTP usado (ex: `"GET"`, `"POST"`).
      * `caminho`: String contendo o caminho da rota requisitada (ex: `"/ola/portuscript"`).
      * `cabecalho`: Mapa contendo os cabeçalhos recebidos.
      * `corpo`: Texto do corpo da mensagem HTTP.
      * `parametros`: Mapa dinâmico contendo as variáveis injetadas por rotas dinâmicas (ex: `req.parametros["nome"]` para a rota `/ola/:nome`).
  * **`Resposta`**:
    * Representa a resposta HTTP a ser enviada pelo servidor. Atributos:
      * `status`: Inteiro indicando o código de status HTTP (ex: `200`, `404`, `500`).
      * `corpo`: Texto a ser retornado no corpo da resposta.
      * `cabecalho`: Mapa com os cabeçalhos de resposta.
    * Métodos:
      * `definir_cabecalho(chave, valor)`: Define um cabeçalho customizado na resposta.
* **Funções**:
  * `requisitar(metodo, url, corpo?, cabecalhos?)`: Realiza uma chamada de requisição HTTP Cliente síncrona completa (suporta chamadas HTTPS) e retorna o respectivo objeto de `Resposta`.

### Módulo: `bd`

Acesso e manipulação de bancos de dados relacionais e não-relacionais. Requer `de "bd" importe ...`.

* **Funções de Conexão**:
  * `conectarSqlite(caminho)`: Abre uma conexão SQLite pura em Go, retornando um objeto `ConexaoSQL`.
  * `conectarPostgres(url)`: Abre uma conexão PostgreSQL, retornando um objeto `ConexaoSQL`.
  * `conectarMongo(url)`: Abre uma conexão MongoDB, retornando um objeto `ConexaoMongo`.
  * `conectarRedis(url)`: Abre uma conexão Redis, retornando um objeto `ConexaoRedis`.
* **A Classe `ConexaoSQL`**:
  * `executar(sql, args...)`: Executa comandos SQL de mutação ou DDL (INSERT, UPDATE, DELETE, CREATE).
  * `consultar(sql, args...)`: Executa consultas SQL SELECT, retornando uma `Lista` de `Mapa`s.
  * `tabela(nome)`: Retorna uma instância de `QueryBuilder` ligada a essa conexão.
  * `fechar()`: Fecha a conexão.
* **A Classe `QueryBuilder`**:
  * `selecionar(colunas...)`: Define as colunas a serem selecionadas.
  * `onde(coluna, operador, valor)`: Adiciona uma cláusula de filtro.
  * `limite(n)`: Limita o número de registros.
  * `obterMuitos()`: Executa e retorna todos os registros correspondentes.
  * `obterUm()`: Executa e retorna o primeiro registro ou Nulo.
  * `inserir(mapaValores)`: Insere um novo registro com o Mapa fornecido.
  * `atualizar(mapaValores)`: Atualiza os registros filtrados com o Mapa de modificações.
  * `deletar()`: Remove os registros que coincidem com os filtros aplicados.
* **A Classe `ConexaoMongo`**:
  * `colecao(nome)`: Retorna uma coleção do MongoDB.
* **A Classe `ConexaoRedis`**:
  * `definir(chave, valor, expiracaoSegundos?)`: Define um valor para a chave.
  * `obter(chave)`: Obtém o valor da chave ou Nulo.
  * `remover(chave)`: Remove a chave.

### Módulo: `soquete`

Controle de sockets de baixo nível (TCP/IP). Requer `importar soquete`.

* **Constantes**: `AF_INET` (IPv4), `AF_INET6` (IPv6), `SOCK_STREAM` (TCP), `SOCK_DGRAM` (UDP).
* **A Classe `Soquete`**:
  * `nova Soquete(familia, tipo)`: Cria o socket chamando as APIs de syscall correspondentes do kernel do SO.
  * `associa(ip, porta)`: Vincula a conexão de rede local (Bind).
  * `ouve(backlog?)`: Ativa escuta de conexões com backlog de fila padrão de 1 se omitido.
  * `aceita()`: Aguarda de forma não-bloqueante (usando `unix.Poll` para gerenciar eventos do File Descriptor) e aceita conexões de clientes, retornando um novo objeto `Soquete` para a troca de dados.
  * `conecta(endereco, porta)`: Conecta o socket cliente ao servidor de destino (resolve o host dinamicamente de DNS via `net.LookupIP`).
  * `envia(bytes)`: Escreve e envia os bytes correspondentes (tipo `Bytes`).
  * `recebe(tamanho)`: Lê os dados disponíveis de rede até o limite do buffer e os devolve envelopados em um objeto `Bytes`.
  * `def_nao_bloqueante(booleano)`: Altera propriedades de espera de E/S do socket.
  * `define_opcoes(nivel, opcao, valor)`: Configura opções de soquete (SetsockoptInt).
  * `fecha()`: Encerra conexões e libera o File Descriptor do SO de forma segura.

---

## 8. Recursos Avançados da Linguagem

O Portuscript possui recursos modernos e engenhosos integrados nativamente em sua especificação gramatical e de runtime:

### 1. O Operador Pipe (`|>`)

Permite encadear transformações e chamadas consecutivas de dados de forma altamente legível.

* **Sintaxe**: `valor |> funcao_ou_metodo` ou `valor |> funcao(argumentoExtra)`.
* **Algoritmo de Injeção**:
  * Se o membro da direita for um identificador de função simples (ex: `texto |> maiusculo`), a VM avalia e executa a chamada simples passando o operando esquerdo como o único argumento: `maiusculo(texto)`.
  * Se o membro da direita for uma chamada parametrizada contendo argumentos extras (ex: `10 |> somar(5)`), o interpretador intercepta a chamada sintática, realiza o append do operando esquerdo na primeira posição da lista de argumentos e executa: `somar(10, 5)`.
* **Prevenção de Efeitos Colaterais**: O operando esquerdo é avaliado uma única vez de forma garantida antes da injeção, prevenindo execuções duplicadas e vazamento de estados (conforme validado nos testes em `pipe_test.go`).

### 1.1 Interpolação de Strings (Templates e Chaves `{}`)

Adicionada no **Sprint 8**, permite embutir expressões lógicas de Portuscript diretamente em strings textuais (`TemplateLiteral`) e componentes delimitados por chaves `{ ... }`.

* **Sintaxe**: `"Olá, { nome }!"` ou `"Dobro: { valor |> duplicar }"`
* **Mecânica de Parsing**: O analisador sintático intercepta strings literais do tipo `lexer.TokenTexto` no parser (`parseAtomo`). Se detectar o padrão de chaves `{ ... }`, o parser segmenta a string em partes literais e expressões dinâmicas (`TemplateExpr`), parseando recursivamente com instâncias isoladas de Parser.
* **Operador Pipe em Interpolações**: O operador pipe `|>` pode ser empregado livremente dentro de chaves em interpolações para transformar dados de forma fluida (ex: `"Nome: { usuario.nome |> maiusculas }"`). No tempo de execução, os visitors da VM resolvem as sub-expressões dinâmicas e as concatenam em uma única string unificada do tipo `Texto`.

### 2. Parâmetros de Funções Avançados (Defaults & Nomeados)

As funções aceitam declarações de valores padrão e chamadas referenciando parâmetros nominalmente (em qualquer ordem de envio).

* **Parâmetros com Default**: `func calcular(a, b = 2) { retorne a * b }`.
* **Chamadas Nomeadas**: `calcular(b = 10, a = 5)`.
* **Mecânica de Resolução**: Na chamada de uma função, os argumentos posicionais são mapeados de forma ordenada. Argumentos nomeados são extraídos e armazenados na struct interna `ArgumentoNomeadoObj`. O método de chamada (`Funcao.M__chame__`) varre a lista de parâmetros formais esperados: preenche com os valores nomeados, avalia e injeta as expressões padrão caso falte algum parâmetro, e gera erro se um argumento obrigatório for omitido.

### 2.1. Anotações de Tipo Opcionais e Validação Estrita (`--estrito`)

Parâmetros, retorno de função e variáveis podem receber anotações de tipo estáticas:

```portuscript
var idade: Inteiro = 18
const PI: Decimal = 3.14

funcao soma(a: Inteiro, b: Inteiro = 0): Inteiro {
    retorne a + b
}
```

* **Os tipos ficam registrados na AST** (`DeclVar.Tipo`, `DeclFuncaoParametro.Tipo`, `DeclFuncao.TipoRetorno`) e são validados ativamente se a flag `--estrito` estiver presente.
* **Validação em tempo de execução**: Ao executar o script com a flag `--estrito` (`portuscript executar arquivo.ptst --estrito`), a VM valida se o valor atribuído a uma variável ou o retorno/parâmetros de uma chamada de função são compatíveis com os tipos anotados. Violações lançam erro do tipo `TipagemErro` (PSC-0004).
* **Tipos suportados**:
  * Primitivos: `Inteiro`, `Decimal`, `Texto`, `Booleano` (ou `Logico`), `Nulo`
  * Compostos: `Lista<T>`, `Mapa<C, V>`, `Tupla` (com validação profunda recursiva de elementos)
  * Assinaturas: `funcao` (ou `Funcao`) para qualquer objeto chamável.

### 2.2. Linter Estático — `portuscript checar`

Comando que varre diretórios recursivamente em busca de arquivos `.ptst`/`.pt` e os analisa sem executar.

```bash
$ portuscript checar ./src --formato=json --estrito
```

#### Flags Suportadas

* `--formato`: Define o formato de saída do relatório.
  * `texto` (Padrão): Saída formatada agrupada por arquivo com sumário.
  * `json`: Emite diagnósticos formatados de acordo com o padrão `Diagnostic` de LSP (Language Server Protocol) em português, com as posições espaciais precisas de cada erro (linha, coluna, tamanho do token).
* `--estrito`: Ativa a verificação de compatibilidade de tipos estáticos anotados.

Verifica:

- **Redeclaração de nomes** no mesmo escopo local (Erro crítico, Severidade LSP 1).
- **Shadowing de nomes** em escopos filhos (Aviso educativo, Severidade LSP 2) - gera apenas um alerta amigável de sombreamento, mantendo o processo de build ativo e funcional.
- **Reatribuição de `const`** (preservando a regra de imutabilidade).
- **Identificadores não declarados** (com fallback para a stdlib via tabela `globalsLinter`).
- **Parâmetros duplicados** na mesma assinatura de função.

Para detalhes de implementação, ver `cmd/checar.go`.

### 3. Tratamento de Exceções Estruturado (`tente / capture / finalmente`)

Fluxo clássico de tratamento de erros nativo em português:

```portuscript
tente {
    var resultado = 10 / 0
} capture (erro) {
    escreva("Erro capturado: " + erro.mensagem)
} finalmente {
    escreva("Sempre executa!")
}
```

* **Mecânica de Escopo**: O bloco `capture` cria um escopo léxico filho temporário e expõe o erro capturado sob a variável especificada em parênteses. O erro é uma instância rica contendo propriedades como `mensagem`, `linha`, `coluna` e `arquivo`. Essas coordenadas são injetadas automaticamente a partir do Contexto da VM, então `erro.arquivo` e `erro.linha` funcionam como inspetores de traceback.
* **Garantia do Finalmente**: O bloco `finalmente` é protegido com mecanismos de `defer` no runtime do Go. Ele é executado de forma garantida, mesmo que ocorram erros dentro do bloco de captura ou que exceções não tratadas se propaguem subindo na pilha de execução.
* **Sobrescrita do Erro Original pelo `finalmente`**: Se o bloco `finalmente` lançar uma exceção, essa exceção substitui o erro original (semântica Python/Java), refletindo em tracebacks. Use `tente { ... } capture (erro) { ... } finalmente { ... }` com cuidado ao manipular recursos propensos a falhar.
* **Bloco `finalmente` Opcional**: Apenas `tente { ... } capture (erro) { ... }` (sem `finalmente`) e `tente { ... } finalmente { ... }` (propagação sem captura — exigem reabertura para tratar) também são sintaxes válidas.

### 4. Plugins e Extensões Dinâmicas Go (`.so`)

Permite carregar dinamicamente bibliotecas compiladas na linguagem Go como extensões do interpretador, operando de forma nativa e rápida.

* **Orquestração de Carregamento**: Se um arquivo com extensão `.so` for importado, a VM usa o pacote `plugin` do Go para abrir a biblioteca, localiza via reflexão o símbolo público da função `InicializaModulo()`, executa-a e carrega seu respectivo escopo estruturado `ModuloImpl`.
* **Compilação**: `go build -buildmode=plugin -o modulo.so modulo.go`. (Suportado nativamente em Linux e macOS).

---

## 9. Arcabouço de Testes Nativos (TDD)

O Portuscript estimula a escrita de testes de qualidade integrando as asserções e as suítes diretamente na sintaxe da linguagem.

### A Palavra-Chave `testar`

Permite declarar um bloco de teste nomeado no próprio script:

```portuscript
testar "deve somar dois numeros corretamente" {
    assegura(soma(2, 2) == 4, "A soma deve ser quatro!")
}
```

* **Isolamento de Estado**: Ao rodar a suíte de testes (`portuscript testar`), o compilador cria um escopo temporário para cada bloco `testar` que herda as variáveis, constantes e importações globais do arquivo original, mas previne colisões e vazamentos de estado de um teste para o outro.

### A Diretiva `assegura` (ou `assegure`)

Atua como a asserção padrão do TDD. Recebe uma expressão de verificação e uma mensagem textual opcional de erro:

```portuscript
assegura condicao, "Mensagem caso falhe";
```

* Se a expressão lógica resultar em `Falso` (ou nulo/zero), lança a exceção estruturada `ErroDeAsseguracao` (PSC-0011).

---

## 10. Diagnósticos e Tratamento de Erros Ricos

Um dos recursos mais inovadores do Portuscript é o seu sistema visual de diagnósticos educativos voltados ao ensino de programação.

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

|      Código      | Classe de Erro               | Causa Típica do Bug                                                    |
| :----------------: | :--------------------------- | :---------------------------------------------------------------------- |
| **PSC-0001** | `SintaxeErro`              | Violação de regras gramaticais e estruturas sintáticas.              |
| **PSC-0002** | `ReatribuicaoErro`         | Tentativa ilegal de redeclarar ou alterar constantes.                   |
| **PSC-0003** | `AtributoErro`             | Acesso a propriedades ou métodos não existentes na instância.        |
| **PSC-0004** | `TipagemErro`              | Operandos de tipos incompatíveis com a operação solicitada.          |
| **PSC-0005** | `NomeErro`                 | Variável ou identificador não definido ou encontrado no escopo.       |
| **PSC-0006** | `ImportacaoErro`           | Falha ao localizar arquivos ou carregar módulos e símbolos.           |
| **PSC-0007** | `ValorErro`                | Argumento de tipo correto, mas valor inadequado.                        |
| **PSC-0008** | `ErroDeLimite`             | Valor numérico fora dos limites aceitáveis pela VM.                   |
| **PSC-0009** | `IndiceErro`               | Indexação de sequências fora dos limites de tamanho.                 |
| **PSC-0010** | `RuntimeErro`              | Falhas genéricas no ambiente de execução da VM.                      |
| **PSC-0011** | `ErroDeAsseguracao`        | Falha na validação lógica de uma asserção de teste (`assegura`). |
| **PSC-0012** | `DivisaoPorZeroErro`       | Tentativa matemática proibida de divisão por zero.                    |
| **PSC-0013** | `ErroDeSistema`            | Falhas de chamadas e comandos de E/S do sistema operacional.            |
| **PSC-0014** | `ArquivoNaoEncontradoErro` | Tentativa de abrir ou acessar um caminho físico inexistente.           |

### Sugestões Educativas Inteligentes e Heurística de Digitação

O interpretador analisa heuristicamente as palavras causadoras do erro no momento de formatar a saída. Se o programador tiver cometido erros de digitação comuns, o compilador fornece a correção amigável em português:

* Se `NomeErro` for disparado e o lexema de falha for `"imrpimir"` ou `"imprimi"`, sugere: `Você quis dizer 'imprimir'?`.
* Se `SintaxeErro` for disparado com a palavra `"retornar"`, sugere: `Em Portuscript, use a palavra-chave 'retorne' para retornar valores.`.
* Se `DivisaoPorZeroErro` for disparado, sugere: `Não é possível dividir um número por zero.`.

### O Renderizador de Traceback com Cores ANSI

O método `Error()` da struct `Erro` implementa um dos melhores formatadores visuais de console existentes:

1. **Deteção de Cores**: Se a variável de ambiente `NO_COLOR` não estiver definida, ativa estilizações cromáticas ANSI True Color de alto contraste (Vermelho para erros, Ciano para ponteiros de guia, Verde para sugestões didáticas).
2. **Desenho de Setas e Sublinhado**: Imprime a caixa identificadora do código (ex: `erro[PSC-0005]`), localiza as coordenadas `arquivo:linha:coluna`, abre o arquivo físico para extrair a linha culpada, desenha uma seta de traceback (`┌──>`) e **sublinha graficamente com acentos circunflexos na largura física exata (`^^^^`)** o token responsável pelo erro, facilitando imensamente a correção visual direta no console!

### Integração com IA Local (Ollama) para Explicação de Erros

A partir do fechamento da **Fase 1**, o comando `portuscript erro` conta com o subcomando `explicar` para fornecer ajuda inteligente utilizando inteligência artificial local:

```bash
$ portuscript erro explicar PSC-0005
```

* **Fluxo de Integração**: O comando realiza uma conexão HTTP local segura com a instância do Ollama (`127.0.0.1:11434/api/generate`) requisitando ao modelo `gemma` uma explicação didática do erro.
* **Fallback e DX Amigável**: Se o Ollama não estiver instalado ou ativo, o CLI detecta a ausência de conexão imediatamente e fornece um tutorial passo a passo em português ensinando como baixar e iniciar o Ollama, procedendo então a renderizar a explicação pedagógica estática catalogada do próprio dicionário local da linguagem, para que o desenvolvedor nunca fique sem auxílio.

---

## 11. Console Interativo (REPL / Playground)

O playground do Portuscript é acessado digitando apenas `portuscript` no terminal. Ele utiliza a biblioteca **Liner** para gerenciar entradas e manter um histórico persistente e inteligente.

### Máquina de Estados e Prompt Multilinha (Controle em `estado.go`)

Para suportar códigos multilinha, o REPL intercepta as linhas digitadas e avalia o pareamento dos delimitadores:

```go
strings.Count(codigo, "[") > strings.Count(codigo, "]") ||
strings.Count(codigo, "(") > strings.Count(codigo, ")") ||
strings.Count(codigo, "{") > strings.Count(codigo, "}")
```

* **Estado Normal (`>>> `)**: Ativo quando não há delimitadores abertos. Ao apertar Enter, o código acumulado é imediatamente enviado para a VM compilar e executar.
* **Estado Contínuo (`... `)**: Se houver delimitadores não pareados (ex: chaves abertas de uma função), o REPL não tenta executar e muda o prompt visual para `... `, indicando que a instrução lógica continua na próxima linha física do console.

### Persistência de Histórico de Comandos em Disco

O histórico de comandos não é perdido ao fechar a sessão. Na inicialização do playground, a VM localiza o diretório Home do usuário e abre/cria o arquivo oculto **`~/.historico_portuscript`**, lendo e carregando os comandos anteriores. Ao fechar (via Ctrl+D ou comando `sair()`), o REPL atualiza e grava a lista de comandos de volta ao disco de forma persistente.

### O Arquivo Virtual `<playground>` e Persistência de Escopo

Para que o desenvolvedor declare uma variável em uma linha e ela continue visível na linha seguinte, o playground inicializa um módulo virtual sob o arquivo inexistente `<playground>`:

```go
exec.Modulo, _ = ctx.InicializarModulo(&ptst.ModuloImpl{
    Info: ptst.ModuloInfo{Arquivo: "<playground>"},
})
```

Todas as expressões avaliadas utilizam o mesmo escopo persistente deste módulo (`exec.Modulo.Escopo`), preservando o estado e evitando "perda de memória" entre as linhas digitadas.

---

## 11.1 Máquina Virtual de Pilha (Fase 2)

A partir da **Fase 2** (Fase de Otimização e Bytecode), o Portuscript conta com uma máquina virtual de pilha altamente eficiente escrita em Go, substituindo a execução clássica de árvore (tree-walk).

### O Compilador de Bytecode

A AST do programa é compilada estaticamente para bytecode compacto (`.ptc`) de passagem única:

* **Pool de Constantes**: Literais do programa (textos, números inteiros, decimais, booleanos, nulos) são internados de forma deduplicada no pool de constantes, otimizando alocações.
* **Opcodes de 1 Byte**: Instruções compactas que controlam a pilha (`OP_PUSH_CONST`, `OP_POP`, `OP_DUP`), execução aritmética (`OP_ADD`, `OP_SUB`), controle de fluxo (`OP_JMP`, `OP_JMP_FALSO`, `OP_RETORNE`) e escopo (`OP_CARREGAR_VAR`, `OP_ARMAZENAR_VAR`).
* **Super-Instruções & Fusão Estática (Fase D)**: Para maximizar o rendimento operacional, o compilador adota passagens de fusão de bytecodes. Ele identifica pares de instruções sequenciais comuns e os funde estaticamente em instruções atômicas compostas de alta velocidade:
  * `OP_RETORNE_CONST`: Funde `OP_PUSH_CONST` + `OP_RETORNE`, lendo do pool e retornando de forma direta e atômica. Otimiza inclusive retornos vazios (`retorne Nulo`).
  * `OP_RETORNE_VAR`: Funde `OP_CARREGAR_VAR` + `OP_RETORNE`, carregando o valor da variável e saindo do frame instantaneamente.
  * *Impacto*: Reduz pela metade a quantidade de decodificações e saltos da VM para operações de retorno.
* **Pulos e Loops Remendados**: Remendos inteligentes de endereçamento de 16 bits (`BigEndian uint16`) para gerenciar saltos de condicionais (`se/senao`) e laços (`enquanto`).

### Execução de Alta Performance via Flag `--vm`

Para rodar qualquer script na nova VM de bytecode, basta passar a flag `--vm` ao comando de execução:

```bash
$ portuscript executar script.ptst --vm
```

### Motor JIT de Traço por Threaded Callbacks (Fase F)

Para atingir o limite máximo de velocidade de execução e aniquilar o custo clássico de decodificação de instruções de interpretadores virtuais (gargalos de loops de `switch/case`), o Portuscript incorpora uma inovadora tecnologia de **Direct-Threaded Code JIT**:

* **Compilação Dinâmica "Just-In-Time"**: Ao carregar um frame de bytecode para execução, a VM de pilha realiza de forma transparente uma passagem de compilação threaded de passagem única, traduzindo o array plano de bytecodes em um array estável de ponteiros de funções Go (`[]InstrucaoThreaded`).
* **Currying de Operandos e Constantes**: Os operandos e constantes são pré-capturados no encerramento (closure) de cada callback Go em tempo de JIT. Isso elimina buscas de memória e incrementos de IP em tempo de execução, resolvendo os valores diretamente de referências estáticas.
* **Preservação de Pulos**: O array threaded de funções coincide perfeitamente em tamanho com o array plano original de bytes. Isso mantém os offsets e saltos de endereçamento de loops e desvios de condicionais 100% íntegros e compatíveis, sem necessidade de alterações estáticas na AST.
* **Impacto de Velocidade**: O loop principal executa chamadas diretas sequenciais de ponteiros de funções no array, contornando desvios e mispredictions de branch de CPU, entregando um ganho de performance colossal.

### Métricas de Benchmark (Interpretador vs VM)

Os testes de benchmark mostram ganhos espetaculares de performance medidos localmente:

- **Velocidade**: A VM de bytecode roda **2.18 vezes mais rápida** que o interpretador clássico de árvore sintática.
- **Redução de Alocação de Memória**: Consumo de heap reduzido em **74.8%** (de `74KB` por ciclo para apenas `18KB`).
- **Frequência de Alocações**: Redução de **64.0%** no número de alocações (GC nativo do Go é acionado muito menos vezes).

### Gerenciamento de Memória por Contagem de Referências (Fase 2.5)

A VM de pilha do Portuscript conta com gerenciamento de memória explícito e determinístico:

* **Protocolo de Referências Ativo**: Utiliza as interfaces `ObjetoGC` e `GCMixin` (`ptst/gc.go`) para controlar as referências de forma ativa nas instruções de empilhamento (`push`), desempilhamento (`pop`) e armazenamento de variáveis (`OP_ARMAZENAR_VAR`).
* **Imunidade de Singletons**: Globais, classes nativas e constantes singleton (`Nulo`, `Verdadeiro`, `Falso`) são inicializadas com `-1` referências de forma imune, garantindo no-ops em retenções e prevenindo coletas acidentais.
* **Limpeza Ativa de Frames**: Símbolos locais e operandos remanescentes na pilha são desalocados e limpos de forma explícita imediatamente ao finalizar a execução de um frame (fim de chamada de função).
* **Coletor e Quebrador de Ciclos (Trial Deletion)**: Referências circulares fechadas órfãs (ex: lista A contém lista B, e lista B contém lista A) são detectadas a partir de varreduras no grafo léxico do escopo ativo e quebradas de forma simétrica (`ptst.ColetarCiclos(escopo)`), prevenindo vazamentos de memória (memory leaks) e preservando a integridade do sistema.
* **Pool de Alocação Rápida / Eden Space para Inteiros Curtos (Fase E)**: Para mitigar o estresse e overhead de alocações sobre o Garbage Collector do Go durante iterações intensas (loops), o runtime pré-aloca estaticamente interfaces do Go para inteiros na faixa de `-100` a `2000`. Sempre que um inteiro nessa faixa é instanciado na VM, a mesma interface imutável pré-alocada é retornada instantaneamente em tempo constante $O(1)$, evitando novas alocações no heap e acelerando operações matemáticas de contagem de loops.

### 11.2. Primitivas de Concorrência & Event Loop Cooperativo (Sprints 9 e 10)

A VM de pilha do Portuscript integra suporte nativo a concorrência assíncrona baseada em corotinas de suspensão cooperativa:

* **Palavras-Chave**: `assincrono` e `aguarde`.
* **Mapeamento de Funções Assíncronas**: Funções marcadas com o modificador `assincrono funcao` têm seu flag `Assincrono` ativado pelo compilador de bytecode.
* **Inovação de Loop de Eventos Baseado em Goroutines**: Ao disparar uma chamada de função assíncrona (`OP_CHAMAR`), a VM detecta o flag ativo e, em vez de bloquear o fluxo principal de execução, delega a sua execução em background a uma nova goroutine leve do Go, retornando imediatamente um objeto `Promessa`.
* **Suspensão Cooperativa via `aguarde` (`OP_AWAIT`)**: Quando a instrução `aguarde` é executada sobre uma `Promessa` ativa, a execução do frame atual da VM cede cooperativamente. Ela registra um callback de encerramento na promessa (`prom.Registre`) e aguarda por meio de um canal seguro do Go (`chan ptst.Objeto`) até que a promessa seja resolvida com sucesso ou rejeitada por erro, garantindo que outras operações concorrentes em background progridam sem travar a VM.
* **Modelo CSP de Concorrência por Canais (Fase B)**: Integração do tipo nativo global `Canal` para troca sincronizada e thread-safe de dados entre goroutines (processos de background):
  * **`nova Canal()`**: Cria uma nova instância de canal de comunicação unificado.
  * **`meuCanal.enviar(dado)`**: Adiciona um dado no canal. Se houver algum processo assíncrono esperando na fila, entrega o dado instantaneamente (FIFO).
  * **`aguarde meuCanal.receber()`**: Retorna uma Promessa suspensa cooperativamente na VM ou no interpretador, que é resolvida com o primeiro dado disponível na fila (FIFO).

### 11.3. Robustez & Modo Sandbox de Segurança (Fase A)

Para garantir que o Portuscript opere como um motor de backend profissional, seguro e de nível industrial, foram integradas proteções nativas no runtime e na biblioteca padrão:

* **Recovery Middleware (Prevenção de Pânicos em Goroutines)**: Todas as requisições tratadas pelo servidor HTTP em background são envelopadas por um tratador `defer recover()`. Se houver pânico lógico inesperado, ele é interceptado de forma segura, respondendo com HTTP 500 sem derrubar o processo e a execução global do interpretador Portuscript.
* **Defesa contra Ataques Slowloris**: Configuração nativa de tempos de limite rígidos (`ReadTimeout: 5s`, `WriteTimeout: 10s`, `IdleTimeout: 120s`) no servidor HTTP para encerrar conexões obsoletas ou propositalmente lentas.
* **Modo Sandbox por Bloqueio de Acesso**: Adição das flags estruturais de restrição de segurança no contexto de execução do interpretador:
  * `BloquearArquivos`: Impede de forma física qualquer leitura, escrita, deleção ou modificação de arquivos do sistema operacional pelo módulo de `arquivos`.
  * `BloquearRede`: Impede abertura de conexões de escuta pelo `Servidor` HTTP ou chamadas de requisição cliente via `requisitar`.
  * Lança erros educativos ricos em português (`PSC-0005: ErroDeSistema - Acesso Negado`) caso os limites do sandbox sejam ultrapassados.

---

## 12. Guia de Sintaxe Rápida e Exemplos de Produção

Abaixo estão descritos snippets estruturados que consolidam as peculiaridades e a sintaxe operacional da linguagem Portuscript.

### 1. Declaração de Variáveis, Constantes e Tipagem Opcional

```portuscript
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

```portuscript
se idade >= 18 {
    imprimir("Maior de idade")
} senao se idade == 17 {
    imprimir("Quase lá! Falta pouco")
} senao {
    imprimir("Menor de idade")
}
```

### 3. Laços de Repetição (enquanto e para-em)

```portuscript
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

```portuscript
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

```portuscript
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

```portuscript
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

imprimir("Servidor Portuscript rodando com sucesso em 127.0.0.1:3000")

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

O Portuscript suporta a geração automática de contratos de API a partir de manifestos `dependencias.json` no ecossistema local do projeto.

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

* **Análise Estática por AST (Fase C)**: O sistema de importações carrega o arquivo `.ptst` do backend correspondente e invoca o parser nativo do Portuscript para gerar a sua árvore sintática abstrata (AST). Ele percorre as declarações de forma estática procurando nós reais de exportação (`DeclExportar` contendo `DeclFuncao`). Isso garante um mapeamento de contratos 100% preciso, imune a espaços, comentários ou quebras de linhas no arquivo original.
* **Geração Estática de Proxies**: Com base nas funções extraídas, o Portuscript gera em tempo de execução um objeto de módulo proxy cujas propriedades são funções dinâmicas do Go. Ao serem executadas, elas realizam automaticamente uma requisição POST HTTP serializada para a URL mapeada em `dependencias.json`.

```portuscript
# Importa de forma remota a função 'obterUsuario' definida no backend
de "@backend/usuarios" importe obterUsuario

var dados = obterUsuario("42")
imprimir(dados) # Realiza uma chamada HTTP POST de forma totalmente transparente!
```

---

## 13. Desenvolvimento Frontend Reativo e SPA (Sinais, JSX, Estilos e SSR)

A partir da **Fase 4**, o Portuscript suporta de forma unificada o desenvolvimento de interfaces reativas e de alta performance que rodam diretamente no navegador do usuário final. O compilador transpila o código Portuscript para JavaScript (ES6) otimizado e de tamanho mínimo acompanhado de um motor de Virtual DOM com reatividade baseada em Sinais (~2.2KB final).

### 13.1. Reatividade por Sinais (Fine-Grained)

A reatividade do Portuscript atualiza de forma fina e cirúrgica apenas os nós do DOM que mudaram, prevenindo re-renderizações totais de página:

* **`sinal(valor)`**: Cria um estado reativo. Retorna um array `[ler, definir]`.
* **`efeito(funcao)`**: Re-executa de forma automática sempre que os sinais dependentes sofrerem alteração.
* **`derivado(funcao)`**: Cria um sinal computado e memoizado.
* **`armazem(objeto)`**: Gerenciador de estado global sincronizado entre múltiplos componentes.

```portuscript
var contadorSinal = sinal(0);
var contador = contadorSinal[0];
var setContador = contadorSinal[1];

efeito(funcao() {
    imprimir("Valor atual do contador: " + contador());
});
```

### 13.2. Componentes e Sintaxe JSX-like

Permite mesclar tags HTML e códigos de forma nativa e semantica:

* **Componentes funcionais**: Funções normais que retornam marcações JSX-like.
* **Eventos**: Atributos como `aoClicar` que mapeiam de forma nativa para eventos `onclick` de browser.
* **Estruturas Inline**:
  * `<se condicao={...}>...</se>`: Condicional dinâmico.
  * `<para item em lista={...}>...</para>`: Loops reativos eficientes.

```portuscript
funcao App() {
    retorne <div classe="p-4">
        <h1>Contador: {contador()}</h1>
        <botao aoClicar={setContador(contador() + 1)}>Incrementar</botao>
    </div>;
}
```

### 13.3. Estilização Nativa e Classes Utilitárias em PT

O compilador suporta dois pilares de estilização e extração sob demanda:

1. **Bloco `estilo`**: Palavra-chave para declarar folhas estáticas em português (ex: `corDeFundo → background-color`, `raio-grande → border-radius`) com suporte a aninhamento e pseudo-classes.
2. **Tailwind em PT**: O compilador extrai apenas as classes utilitárias PT usadas (como `flex-linha`, `p-4`, `fundo-azul`, `itens-centro`) e gera o arquivo `estilos.css` de saída.

### 13.4. Roteamento por Arquivos (File-system Routing)

A CLI detecta de forma automática pastas de rotas (`/web/rotas/` ou `/rotas/`) e cria o mapeamento SPA automático. O componente especial `<Link para="...">` intercepta cliques e navega de forma instantânea sem recarregar fisicamente a página.

### 13.5. SSR (Server-Side Rendering) e Hidratação

O servidor de backend do Portuscript (`stdlib/http/http.go`) pode renderizar as páginas em HTML estático inicial instantâneo contendo metadados ricos de JSON-LD Schema.org (AEO) e OpenGraph. No navegador, o runtime web liga de forma invisível os Sinais existentes na estrutura física (Processo de Hidratação), ligando os fios de reatividade sem piscar ou destruir o DOM estático inicial.

### 13.6. Arquitetura SPA comparada ao Angular

Para desenvolvedores com experiência em Angular, o Portuscript Web oferece equivalências diretas e simplificadas de design:

* **Pipes** ➔ Operador Pipe nativo (`|>`) para formatações visuais limpas em templates.
* **Directives** ➔ Tags de controle JSX (`<se>` para `*ngIf`, `<para>` para `*ngFor`).
* **Services** ➔ Exportações de estado global baseadas em `armazem()`.
* **Validators** ➔ Sinais compostos derivados (`derivado()`).

### 13.7. Modelo Híbrido de Desenvolvimento (Arquivos Separados)

Para suportar o desenvolvimento de sistemas complexos e evitar arquivos gigantescos, o Portuscript permite separar de forma limpa as responsabilidades visuais, de estilo e de comportamento lógico:

1. **Estilos em Português (`.estilo.ptst`)**: Arquivos com extensão `.estilo.ptst` contêm exclusivamente blocos de estilo declarados em português (ex: `estilo Caixa { ... }`). Eles podem ser importados normalmente no seu arquivo de lógica.
2. **Layouts HTML Separados (`.html`)**: Você pode extrair a marcação JSX para arquivos `.html` separados e carregá-los de dentro da lógica do componente usando a chamada nativa `importarHtml("./template.html")`. O compilador em Go lê o arquivo e faz o inline dinâmico do HTML traduzido em tempo de compilação.

```portuscript
# Exemplo de arquivo lógico: BotaoPersonalizado.ptst
de "web" importe sinal, importarHtml;
de "./BotaoPersonalizado.estilo.ptst" importe CaixaDeBotao; # Importa estilo do .estilo.ptst

funcao BotaoPersonalizado() {
    var [contador, setContador] = sinal(0);
    # Carrega e injeta o layout físico de forma transparente
    retorne importarHtml("./BotaoPersonalizado.html");
}
```

### 13.8. Recursos e Primitivas de Nível de Produção

O ecossistema frontend do Portuscript inclui inovações de performance e facilidade de desenvolvimento para sustentar sistemas corporativos de grande porte:

* **Two-Way Data Binding (`ligar={sinal}`)**: Elimina o código repetitivo em formulários. Ao usar `<input ligar={nome} />`, o compilador e o runtime criam o vínculo bidirecional reativo automático entre o sinal de estado e o elemento físico de entrada do browser.
* **Modificadores de Eventos Declarativos**: Encadeamento direto na propriedade de eventos para manipulação do comportamento físico (ex: `aoEnviar_prevenir={submeter}` intercepta e executa `e.preventDefault()` de forma transparente antes do callback, e `aoClicar_parar` executa `e.stopPropagation()`).
* **Keyed Diffing (`chave`)**: Desempenho linear $O(N)$ em renderizações de listas. O algoritmo de diff do Virtual DOM no `runtime-web.js` utiliza o atributo `chave` em tags dentro de loops `<para>` para reutilizar e reposicionar nós físicos no DOM em vez de destruí-los.
* **Sinais Persistentes (`sinalPersistente`)**: Primitiva de reatividade sincronizada automaticamente com a API de `localStorage` do navegador do usuário final.
* **Sinais Assíncronos (`recurso`)**: Simplifica a gestão de requisições HTTP e consumo de APIs fornecendo flags de estado síncronas de progresso: `.carregando()`, `.erro()`, e `.ok()`.
* **Injeção de Dependências (`Provedor` & `injetar`)**: Permite prover instâncias de stores e serviços no topo da árvore de componentes e recuperá-los de forma limpa em componentes filhos profundos, evitando o acoplamento excessivo de propriedades (*prop-drilling*).
* **Componentes de UI Nativos e Acessíveis**:
  * `<FronteiraDeErro>`: Proteção de renderização que impede erros em componentes e widgets secundários de causarem tela branca no sistema inteiro, exibindo um componente de fallback amigável.
  * `<ListaVirtual>`: Renderiza estritamente os nós visíveis na tela para coleções de dados massivas (ex: 50.000 linhas), mantendo a performance de rolagem a 60fps constantes.
  * `<GradeDeDados>`: Tabela interativa com filtros rápidos de pesquisa em português, paginação automatizada e ordenamento rápido.

### 13.9. Criação de Projetos (Scaffolding)

Você pode inicializar uma estrutura padrão de projeto completa com suporte híbrido de forma automática utilizando a CLI Cobra:

```bash
portuscript iniciar meu_app
```

O comando gerará os seguintes diretórios e arquivos de exemplo pré-configurados no disco:

* `/main.ptst` (ponto de entrada que monta a aplicação)
* `/web/rotas/index.ptst` (página de início demonstrando importações de arquivos)
* `/web/componentes/Botao.ptst` (componente visual lógico)
* `/web/componentes/Botao.estilo.ptst` (folha de estilo separada inteiramente em português)
* `/web/componentes/Layout.html` (layout HTML separado demonstrando o uso de `importarHtml`)

### 13.10. Novas Primitivas Avançadas e Sinais de Tempo

* **Sinais com Debounce (`sinalDebounce`)**: O Portuscript fornece a primitiva `sinalDebounce(valorInicial, tempoEmMs)` em seu runtime web. Ela atrasa de forma inteligente a atualização de estados reativos e expõe seu atualizador direto no getter (`ler.set`), integrando-se nativamente e sem boilerplates com o binding bidirecional `_ligar` em formulários de pesquisa.

---

## Capítulo 14 — Segurança e Blindagem Corporativa (Security Audit)

As ferramentas e a CLI do Portuscript foram submetidas a uma auditoria rigorosa de segurança de nível de produção, contando com defesas contra as vulnerabilidades mais críticas do mercado de software:

* **Prevenção de Zip Slip (Path Traversal)**: O comando de instalação de pacotes `portuscript instalar` valida estaticamente todos os caminhos do arquivo ZIP extraídos no disco com `filepath.Clean` e `strings.HasPrefix(caminhoLimpo, pastaAlvo)`. Caso uma travessia ilegal com caminhos relativos (`..`) seja detectada, a extração é abortada com segurança na hora, impedindo corrupção física do sistema de arquivos.
* **Prevenção de Corridas de Dados (Anti-Race Condition)**: O interpretador web do playground local serializa de forma síncrona as execuções de código utilizando bloqueio de exclusão mútua (`sync.Mutex`). Isso garante que múltiplas requisições simultâneas não causem corridas de dados ao interceptar a saída padrão global de console (`os.Stdout`), isolando totalmente a saída de logs de cada usuário de forma segura.
* **Resiliência contra DoS de Rede**: O servidor web do playground limita síncronamente o tamanho do payload do editor de código para no máximo 1MB via `http.MaxBytesReader` e configura limites estritos de `ReadTimeout` e `WriteTimeout` no servidor HTTP Go.

---

## Capítulo 15 — Otimizações Avançadas e Desempenho da Máquina Virtual

A Máquina Virtual de bytecode e o runtime de execução do Portuscript foram aprimorados com otimizações de baixo nível de classe mundial para sustentar aplicações de altíssima performance:

* **Recursion Guard (PSC-0015)**: Implementação de proteção ativa contra estouros físicos de pilha da VM. O interpretador rastreia a profundidade de execução das chamadas e interrompe loops recursivos infinitos ao ultrapassar o limite seguro de 1000 chamadas, lançando o erro estruturado `ErroDePilha` (PSC-0015).
* **Operand Stack Pre-allocation Pool**: Reaproveitamento agressivo de memória na VM. Utiliza um pool global sincronizado (`sync.Pool` em Go) para fornecer fatias pré-alocadas de operandos com capacidade fixa de 128 elementos. Ao fim da execução de cada bloco/função, os operandos são zerados e devolvidos ao pool, reduzindo a pressão do coletor de lixo (GC) de Go a zero para frames normais.
* **Morphic Inline Caching (MIC)**: Otimização em tempo de execução para a instrução de carregamento de variáveis (`OP_CARREGAR_VAR`). Símbolos resolvidos em loops quentes são cacheados de forma monomórfica em closures JIT. Se o escopo ou objeto de destino for idêntico ao do ciclo anterior, a VM extrai o valor diretamente em tempo constante $O(1)$ sem realizar buscas complexas de tabelas hash.
* **Profiler Embutido (`--perfil`)**: O comando `portuscript executar --perfil` ativa a coleta síncrona de estatísticas e carimbos de tempo para cada instrução de bytecode (Opcode). Ao fim do programa, é exibida uma tabela de desempenho contendo a contagem exata de chamadas e hotspots lógicos de execução.

---

## Capítulo 16 — Novos Módulos Avançados da Biblioteca Padrão (Stdlib)

* **Logs Estruturados (`de "logs"`)**: Sistema de logging estruturado nativo com níveis (`info`, `alerta`, `erro`, `depurar`) e suporte a metadados dinâmicos (Mapas) e formatação selecionável (texto colorido amigável ou JSON para produção).
* **Métricas de Observabilidade (`de "metricas"`)**: Permite criar e registrar contadores e medidores (Gauges) dinâmicos compatíveis com o formato do Prometheus na rota `/metricas` para observabilidade de microsserviços.
* **Validador de Esquemas de Dados (`de "esquema"`)**: Permite declarar restrições de esquemas de dados complexos com validação em tempo de execução (Ex: `esquema.NovoEsquema({ "nome": esquema.Texto, "idade": esquema.Inteiro })`).
* **Agendador de Tarefas e Filas (`de "tarefas"`)**: Expõe o controle de filas concorrentes em memória e agendamento periódico baseado em Cron (Ex: `tarefas.agendar("*/5 * * * * *", funcao() { ... })`).
* **Foreign Function Interface (`ffi`)**: Ponte nativa bidirecional de baixo nível que permite carregar bibliotecas binárias compartilhadas C-compatíveis (`.so`, `.dll`, `.dylib`) e executar assinaturas externas diretamente no Portuscript de forma síncrona e performática.

---

## Capítulo 17 — Extensões de CLI, DevOps e DevOps DX

* **Empacotamento Autônomo (`portuscript empacotar`)**: Subcomando de compilação avançada de binários autônomos puros. O Portuscript compila o código do usuário para bytecode `.ptc` e o funde a um executável Go do interpretador, gerando um único executável nativo livre de dependências para o usuário final com suporte nativo a cross-compilation (via `--so` e `--arq`).
* **Testador de Estresse Concorrente (`portuscript stressar`)**: Permite executar baterias massivas de requisições concorrentes e benchmarks automáticos para testar a resiliência de servidores e scripts Portuscript.
* **Protocolo de Adaptador de Depurador (`portuscript depurar`)**: Servidor TCP compatível com o protocolo oficial Debug Adapter Protocol (DAP). Permite a conexão e handshakes síncronos de IDEs modernas (como VS Code, Cursor) para depuração de nível profissional com breakpoints e inspeção de variáveis locais.
* **Extensão VS Code Oficial (`vscode-portuscript`)**: Extensão oficial que habilita realce de sintaxe completo, preenchimento rápido (snippets) para front/back, e se conecta via stdio/sockets diretamente aos servidores `lsp` (Language Server) e `depurar` (DAP) integrados na CLI.
  * **Formatação Síncrona On-Save via LSP**:
    1. *No Servidor LSP (Go)*: Durante a inicialização, o servidor (`cmd/lsp.go`) declara suporte nativo de formatação síncrona via `"documentFormattingProvider": true`. Quando recebe a requisição síncrona `textDocument/formatting` enviada pela IDE, ele intercepta o comando e retorna as edições do código limpo processadas pela função nativa `FormatarCodigoPortuscript(codigo)` declarada em `cmd/formatar.go`.
    2. *Na Extensão do VS Code (`vscode-portuscript`)*: No arquivo `vscode-portuscript/extension.js`, o cliente LSP é instanciado via classe `LanguageClient`. Ao inicializar, a biblioteca padrão `vscode-languageclient` detecta a capacidade `"documentFormattingProvider": true` fornecida pelo servidor e registra automaticamente a capacidade de formatação nativa na IDE.
    3. *Como usar no VS Code*:
       - **Atalho de Formatação**: Pressionar `Shift + Alt + F` (Windows/Linux) ou `Shift + Option + F` (macOS) com um arquivo `.ptst` aberto.
       - **Formatação Automática ao Salvar**: Habilitar a configuração `"editor.formatOnSave": true` nas configurações do VS Code para disparar a formatação limpa automaticamente em todo `Cmd+S` or `Ctrl+S`.
  * **Publicação da Extensão no VS Code Marketplace (`vscode-portuscript`)**:
    Caso você queira gerar e publicar atualizações da extensão oficial para a comunidade global de desenvolvedores do VS Code, siga os passos abaixo usando o utilitário oficial `vsce` (VS Code Extension Manager):
    1. **Instalação do CLI**: Instale o gerenciador de extensões da Microsoft de forma global via npm:
       ```bash
       npm install -g @vscode/vsce
       ```
    2. **Criação do Publicador (Publisher)**:
       - Crie uma conta de desenvolvedor no [Visual Studio Marketplace](https://marketplace.visualstudio.com/).
       - Crie um ID de Publicador exclusivo (ex: `portuscript`).
       - Insira esse ID no campo `"publisher"` do arquivo `package.json` localizado dentro da pasta `vscode-portuscript/`.
    3. **Token de Acesso Pessoal (PAT)**:
       - Crie uma conta no Azure DevOps (`dev.azure.com`) sob a mesma organização ou e-mail.
       - No painel superior direito do Azure DevOps, vá em **Personal Access Tokens**.
       - Adicione um novo token selecionando a organização "All accessible organizations", defina o escopo para **Marketplace (Publish)** com acessos de leitura e gravação (*Read & Write*). Salve o token (PAT) gerado em um local seguro.
    4. **Autenticação no Terminal**: Efetue login no publicador por meio do terminal:
       ```bash
       vsce login [seu-id-publicador]
       ```
       (Cole o PAT gerado no Azure DevOps quando solicitado).
    5. **Empacotamento e Publicação**:
       - Entre no diretório da extensão: `cd vscode-portuscript`
       - Instale as dependências locais de desenvolvimento: `npm install`
       - **Publicar diretamente**: Execute `vsce publish` (ou incremente versões via `vsce publish patch` / `vsce publish minor`).
       - **Apenas empacotar localmente (offline)**: Para gerar um arquivo instalável `.vsix` localmente sem enviar para o Marketplace público, execute `vsce package`. O arquivo `.vsix` gerado pode ser compartilhado com qualquer desenvolvedor para instalação manual arrastando-o para a aba de extensões do VS Code.

---

## Capítulo 18 — Pacote de Inteligência Artificial e IA Generativa (`de "ia"`)

A Fase 6 introduz o pacote de IA padrão do Portuscript, oferecendo uma interface declarativa e humana em português para interação com modelos de linguagem:

* **Conector Local com Ollama**: A função `ia.conectarLocal("llama3")` estabelece um cliente nativo apontando para uma instância local do Ollama, abstraindo requisições HTTP e gestão de tokens em chamadas simples:
    ```portuscript
    de "ia" importe conectarLocal, completar, transcreverAudio;

    var modelo = conectarLocal("llama3");
    var resposta = modelo.completar("Qual é a capital do Brasil?");
    imprimir(resposta);
    ```
* **Conector Remoto com OpenAI/Anthropic**: Para desenvolvedores que optam por provedores em nuvem, basta instanciar `ia.conectarRemoto("openai", "sua-chave-api")` e escolher o modelo. O mesmo prompt programático é reutilizado.
* **Compreensão Multimodal**: Funções nativas para transcrição de áudio (`transcreverAudio(blob)`), sumarização de PDFs e classificação de intenção semântica em frases.
* **Agentes de IA Declarativos**: Permite a construção de agentes em poucas linhas com a função `criarAgente`, integrando memória de curto prazo e ferramentas nativas (calendário, banco de dados, arquivos) via *function calling* estruturado.

---

## Capítulo 19 — Conectores de Banco de Dados Corporativos (`de "bd"`)

Drivers de persistência robustos prontos para escalar em ambientes de alta concorrência corporativa:

* **Drivers Nativos Expandidos**: Implementação nativa em Go de conectores de banco para **PostgreSQL**, **MySQL** e **MongoDB**, com séries históricas de migrations encapsuladas e gestão de credenciais via contexto.
* **Pool de Conexões Otimizado**: Reaproveitamento de conexões ativas a partir de um *pool* thread-safe com backoff exponencial e detecção automatizada de timeouts do servidor.
* **Transações Atômicas**: Block declarativo `bd.transacao(funcao() { ... })` que garante commit atômico ou rollback completo em erros, validando shutdown do Node independente da thread de origem.
* **Query Builder Refinado**: Sintaxe encadeada fluente ainda mais ergonômica para filtros complexos, subqueries e joins múltiplos com autodetecção de colunas.
* **Detecção de Esquema Automática**: Mapeamento automático entre tipos SQL (varchar, timestamp, jsonb) e o sistema de tipos da Stdlib `esquema`.

---

## Capítulo 20 — WebAssembly (WASM) e Microsserviços Corporativos

A camada de execução WASM eleva o limite de processamento matemático pesado no navegador, com interoperabilidade síncrona e tipada com a plataforma host:

* **Alvo de Compilação `--alvo=wasm`**: O comando `portuscript compilar --alvo=wasm` gera arquivos binários `.wasm` autônomos com tamanho mínimo.
* **Interoperabilidade Síncrona**: Bridge tipada em TypeScript (`portuscript/exporta.ts`) para expor funções da VM (rotas de filtro, ordenação pesada, parser JSON) e consumi-las via `WebAssembly.instantiate` na thread principal sem gargalos de marshalling.
* **Stone of Dedicatória**: A Fase 6 fecha o Portuscript como um ecossistema de linguagem corporativa de ponta, contando com compilador, IDE nativa, ferramenta de empacotamento de binários, runtime WASM, IA integrada e drivers de banco escaláveis.

---

## Capítulo 21 — Documentação Ininterrupta e Estilo de Contribuição

A linha mestra de desenvolvimento do Portuscript se mantém desde a sua concepção:

* **Sintaxe Humana**: Construída sob medida para falantes nativos de português, com palavras-chave fonéticas sem ambiguidade (`funcao`, `retorne`, `classe`, `estende`).
* **DX como prioridade**: Erros didáticos, mensagens contextuais e o comando `portuscript erro explicar` integrado com LLMs locais.
* **CLI consistente**: Nomes de comandos, flags e cláusulas escritos exclusivamente em português brasileiro (`--alvo`, `--estrito`, `--otimizar-assets`).
* **Segurança por padrão**: Toda nova feature é auditada contra *path traversal*, condicional de corrida, DoS de payload e *race conditions* em pipes assíncronos durante os testes de aceitação.
