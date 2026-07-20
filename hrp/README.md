# Pacote `hrp` (Núcleo e Runtime do Harpia)

O pacote `hrp` é o **coração e a Máquina Virtual de runtime** do interpretador do **Harpia**. Ele implementa toda a infraestrutura de modelagem de dados dinâmicos, tabelas de símbolos e escopo, sistema de tipos orientados a objetos (Classes/Metaclasses) e os protocolos operacionais que governam a execução lógica de qualquer script.

---

## 📖 Índice

1. [O Modelo Primordial de Objetos (`Objeto`)](#-o-modelo-primordial-de-objetos-objeto)
2. [Sistema de Classes e Metaclasses (`Tipo`)](#-sistema-de-classes-e-metaclasses-tipo)
3. [Tabela de Protocolos e Métodos Mágicos](#-tabela-de-protocolos-e-métodos-mágicos)
   - [Protocolo de Coerção e Tipos](#protocolo-de-coerção-e-tipos)
   - [Protocolo de Cálculos Aritméticos](#protocolo-de-cálculos-aritméticos)
   - [Protocolo de Comparações Lógicas Ricas](#protocolo-de-comparações-lógicas-ricas)
   - [Protocolo de Loops e Iteradores](#protocolo-de-loops-e-iteradores)
4. [Tabela de Símbolos e Resolução de Nomes (`Escopo`)](#-tabela-de-símbolos-e-resolução-de-nomes-escopo)
5. [Orquestração do Ambiente da VM (`Contexto`)](#-orquestração-do-ambiente-da-vm-contexto)
6. [Mecânica Geral de Avaliação de AST](#-mecânica-geral-de-avaliação-de-ast)

---

## 🧩 O Modelo Primordial de Objetos (`Objeto`)

Toda e qualquer variável, constante, lista, mapa, função ou instância de classe tratada pela VM do Harpia deve obrigatoriamente implementar a interface primordial Go definida no arquivo `objeto.go`:

```go
type Objeto interface {
    Tipo() *Tipo
}
```

- **Por que é modelado assim?**  
  Como Go é uma linguagem estaticamente tipada, esta assinatura polimórfica abstrata unifica todas as instâncias sob um tipo genérico `Objeto`, permitindo que coleções de dados (como listas) guardem dados heterogêneos.
- **Reflexão Dinâmica**: O método `Tipo()` retorna um ponteiro para a classe estrutural (`*Tipo`) do objeto, permitindo que a VM verifique dinamicamente quais métodos e atributos o objeto expõe em tempo de execução.

---

## 🏗️ Sistema de Classes e Metaclasses (`Tipo`)

O arquivo `tipo.go` implementa a struct `Tipo`, que atua como a representação física de uma classe ou metaclasse na VM do Harpia:

```go
type Tipo struct {
    Nome       string         // Nome visual da classe (ex: "Decimal").
    Nova       NovaFunc       // Alocador estático de memória (__nova_instancia__).
    Inicializa InicializaFunc // Construtor inicializador de atributos (__inicializa__).
    Doc        string         // Bloco de documentação explicativo (Docstring).
    Base       *Tipo          // Ponteiro para a classe pai (Herança).
    Mapa       Mapa           // Tabela hash contendo métodos e constantes expostos.
}
```

### O Processo de Montagem Pré-Runtime:

Para evitar vazamentos de memória ou heranças quebradas, o interpretador realiza um processo de inicialização em duas etapas:

1. **Fila de Montagem**: Quando uma classe Go é alocada via `NewTipo()`, ela é armazenada temporariamente em um slice global chamado `filaMontagem`.
2. **Consolidação (`MontaOsTipos`)**: Antes de executar o primeiro byte de código do script do usuário, a VM dispara `MontaOsTipos()`, consolidando heranças, herdando métodos não substituídos, populando docstrings na propriedade interna `__doc__` e travando as assinaturas básicas na tabela hash do tipo.

---

## 🔌 Tabela de Protocolos e Métodos Mágicos

No Harpia, operadores e funções globais (como `+`, `==`, `tamanho()`) não são acoplados rigidamente a tipos estáticos. Eles funcionam de forma dinâmica por meio de **Protocolos** de interfaces Go (métodos mágicos, conceitualmente idênticos aos _dunder methods_ do Python) declarados em `interfaces.go`.

### Convenção de Nomenclatura Estrita:

- Interfaces de tipagem Go devem iniciar com o caractere **`I`**.
- Os métodos internos correspondentes que satisfazem a interface devem iniciar com a letra **`M`** (de "Método").

---

### Protocolo de Coerção e Tipos

| Interface Go    | Método Vinculado  | Equivalente no Harpia | Descrição Técnica                                             |
| :-------------- | :---------------- | :-------------------- | :------------------------------------------------------------ |
| `I__texto__`    | `M__texto__()`    | `texto(obj)`          | Converte o objeto para sua representação de string (`Texto`). |
| `I__bytes__`    | `M__bytes__()`    | `bytes(obj)`          | Converte o objeto para sua sequência de bytes raw (`Bytes`).  |
| `I__inteiro__`  | `M__inteiro__()`  | `int(obj)`            | Coerge o objeto para número inteiro de 64 bits (`Inteiro`).   |
| `I__decimal__`  | `M__decimal__()`  | `decimal(obj)`        | Coerge o objeto para float64 de dupla precisão (`Decimal`).   |
| `I__booleano__` | `M__booleano__()` | `booleano(obj)`       | Avalia a verdade ou falsidade lógica do objeto (`Booleano`).  |

---

### Protocolo de Cálculos Aritméticos

Quando o interpretador encontra operadores matemáticos na AST (ex: `a + b`), ele resolve o tipo de `a` e verifica se ele implementa o protocolo aritmético correspondente:

| Interface Go          | Método Vinculado             |   Operador   | Descrição e Comportamento                                  |
| :-------------------- | :--------------------------- | :----------: | :--------------------------------------------------------- |
| `I__adiciona__`       | `M__adiciona__(outro)`       |     `+`      | Adição aritmética ou concatenação textual.                 |
| `I__subtrai__`        | `M__subtrai__(outro)`        |     `-`      | Subtração matemática.                                      |
| `I__multiplica__`     | `M__multiplica__(outro)`     |     `*`      | Multiplicação matemática ou repetição múltipla de strings. |
| `I__divide__`         | `M__divide__(outro)`         |     `/`      | Divisão real com precisão de dízima ponto flutuante.       |
| `I__divide_inteiro__` | `M__divide_inteiro__(outro)` |     `//`     | Divisão por piso (retorna o inteiro truncado).             |
| `I__mod__`            | `M__mod__(outro)`            |     `%`      | Resto de divisão inteira de módulo.                        |
| `I__neg__`            | `M__neg__()`                 | `-` (unário) | Inversão unária aritmética de sinal (ex: `-x`).            |

---

### Protocolo de Comparações Lógicas Ricas

As comparações relacionais e de valor são governadas pelo conjunto de interfaces Go agrupadas na super-interface `I_comparacaoRica`:

| Interface Go          | Método Vinculado             | Operador | Função Lógica                                            |
| :-------------------- | :--------------------------- | :------: | :------------------------------------------------------- |
| `I__igual__`          | `M__igual__(outro)`          |   `==`   | Avalia se ambos os valores são semanticamente idênticos. |
| `I__diferente__`      | `M__diferente__(outro)`      |   `!=`   | Avalia se os valores são divergentes.                    |
| `I__menor_que__`      | `M__menor_que__(outro)`      |   `<`    | Comparador menor que.                                    |
| `I__menor_ou_igual__` | `M__menor_ou_igual__(outro)` |   `<=`   | Comparador menor ou igual.                               |
| `I__maior_que__`      | `M__maior_que__(outro)`      |   `>`    | Comparador maior que.                                    |
| `I__maior_ou_igual__` | `M__maior_ou_igual__(outro)` |   `>=`   | Comparador maior ou igual.                               |

---

### Protocolo de Loops e Iteradores

A iteração inteligente de laços de repetição `para x em colecao` é provida de forma desacoplada pela interface unificada `I_iterador`:

1. **`I__iter__`** ➔ `M__iter__() (Objeto, error)`: Chamado na inicialização do laço. Deve retornar uma estrutura que aja como o iterador ativo (geralmente ele mesmo ou uma nova instância de cursor).
2. **`I__proximo__`** ➔ `M__proximo__() (Objeto, error)`: Chamado a cada ciclo do loop. Deve retornar o valor do item atual e avançar o cursor de leitura interna.
   - **Fim da Iteração**: Assim que a coleção atinge o limite final, o método `M__proximo__` deve lançar o erro controlado estruturado especial **`FimIteracao`**. A VM intercepta essa exceção, encerra o laço `para` de forma elegante e continua a execução do script.

---

## 📇 Tabela de Símbolos e Resolução de Nomes (`Escopo`)

O arquivo `escopo.go` gerencia o encadeamento e o isolamento de variáveis de tempo de execução (tabelas de símbolos) em nível de blocos e funções de forma 100% thread-safe:

- **Encadeamento Léxico**: Cada instância de `Escopo` possui um ponteiro opcional para um escopo pai (`Pai *Escopo`).
- **Sincronização de Concorrência do Escopo**: Cada escopo individual conta com seu próprio `sync.RWMutex` que protege o mapa hash de símbolos de acessos concorrentes (reads/writes) por múltiplas goroutines Go rodando processos assíncronas do Event Loop.
- **Locks de Grão Fino em Símbolos**: Símbolos individuais do runtime (`hrp.Simbolo`) possuem um mutex local (`sync.RWMutex`) próprio que envolve as operações de consulta e gravação do seu campo `Valor`, garantindo atualizações atômicas e seguras.
- **Cópia Rasa Segura para o GC**: O escopo expõe o método `ObterSimbolosSeguro()`. Ele adquire um lock de leitura e devolve uma fatia estável de referências a símbolos para que o Garbage Collector faça varreduras cíclicas sem perigo de colisões de mapas Go em tempo de execução.
- **Algoritmo de Resolução de Símbolo (`ObterValor`)**:
  1. Tenta localizar a variável/constante no mapa hash do escopo local (sob lock de leitura seguro). Se encontrar, devolve o valor.
  2. Se não encontrar e o ponteiro `Pai` for diferente de `nil`, sobe recursivamente na hierarquia de escopos executando a busca no escopo superior.
  3. Se atingir o escopo primordial global sem sucesso, lança uma exceção estruturada de erro de nome (**`NomeErro`**).

---

## 🌐 Orquestração do Ambiente da VM (`Contexto`)

O arquivo `contexto.go` é o orquestrador macro da Máquina Virtual, responsável pelo ciclo de vida global de execução:

- **Busca de Módulos**: Mantém a lista de caminhos absolutos de busca (`CaminhosPadrao`) onde o interpretador varre o disco para localizar importações de arquivos locais (`importe "modulo"`).
- **Caches de Módulos**: Mantém o dicionário `ModulosCarregados` para garantir que o mesmo módulo não seja importado ou compilado repetidas vezes redundantes, acelerando o tempo de execução.
- **Isolamento de Estado**: Permite isolar instâncias completas da VM executando em paralelo sob contextos concorrentes separados de forma 100% thread-safe.

---

## ⚠️ Tratamento de Exceções (`tente / capture / finalmente`)

O Runtime do Harpia expõe um protocolo estruturado de tratamento de exceções totalmente em português, implementado em `interpretador.go` na rotina `visiteTenteCapture`:

| Componente           | Arquivo            | Responsabilidade                                                                                                   |
| :------------------- | :----------------- | :----------------------------------------------------------------------------------------------------------------- |
| `Erro` struct        | `erros.go`         | Representa uma exceção rica com `mensagem`, `linha`, `coluna`, `arquivo`, `token`, `sugestao`, `codigo`.           |
| `AdicionarContexto`  | `erros.go`         | Injeta metadados geográficos do `Contexto` da VM no erro (linha, coluna, arquivo).                                 |
| `visiteTenteCapture` | `interpretador.go` | Despachante: executa `tente`, desvia para `capture`, garante `finalmente`.                                         |
| `parseTenteCapture`  | `parser/parser.go` | Constrói o nó de AST `TenteCaptureFinalmente` a partir dos tokens `TokenTente`, `TokenCapture`, `TokenFinalmente`. |

**Semântica implementada (ver `excecoes_test.go`):**

1. `tente { ... } capture (erro) { ... }` – bloco `capture` recebe o erro em escopo léxico isolado.
2. `tente { ... }` – erros sem `capture` propagam após `finalmente` (se houver).
3. `finalmente` sempre roda via `defer` Go – mesmo em sucesso, captura ou propagação.
4. Erros em `finalmente` substituem o erro original.

**Importante**: `AdicionarContexto` propaga os metadados geográficos automaticamente, então o `erro.linha`/`erro.arquivo` deve ser totalmente funcional em qualquer bloco `capture`.

---

---

## 🗑️ Coletor de Memória por Contagem de Referências (`ObjetoGC`)

Adicionado no fechamento da **Fase 2.5**. O pacote `hrp` implementa gerenciamento determinístico de ciclo de vida de objetos complexos mutáveis (como `Lista` e `Mapa`) usando contagem de referências ativa:

- **Interface `ObjetoGC`**:
  ```go
  type ObjetoGC interface {
      Objeto
      Reter()
      Liberar()
      ObterRefs() int
      ObterFilhos() []Objeto
  }
  ```
- **Mixin `GCMixin`**: Estrutura leve e embutível contendo `RefsCount int`. Fornece comportamentos reusáveis de retenção (`Reter()`), liberação (`Liberar()`) e contagem (`ObterRefs()`).
- **Imunidade de Singletons**: Constantes fundamentais e tipos nativos (como `Nulo`, `Verdadeiro`, `Falso`) são inicializados com `RefsCount = -1` de forma permanente, tornando-os imunes a decrementos.
- **Quebra Simétrica de Ciclos (`ColetarCiclos`)**: Algoritmo de varredura _Trial Deletion_ baseado em alcance léxico. Ele detecta de forma recursiva grafos isolados circulares órfãos (ex: Lista A contendo Lista B, e Lista B contendo Lista A) e quebra a ciclicidade esvaziando seus filhos, permitindo que suas referências cheguem a zero e o Garbage Collector nativo do Go os libere da memória de forma definitiva.

---

## ⚙️ Mecânica Geral de Avaliação de AST

Durante a interpretação física de um script, as etapas operacionais de baixo nível em `hrp` ocorrem de forma fluída:

```
[Código Fonte Harpia]
           │
           ▼
[Parser] ➔ Compila para nós de AST BaseNode
           │
           ▼
[hrp.Contexto] ➔ Invoca AvaliarAst(ast, escopo)
           │
           ▼
   [Máquina de Estados de Avaliação (VM)]
           │
           ├─► Resolve referências em hrp.Escopo
           ├─► Realiza operações de cálculo via hrp.OperadoresAritmeticos
           ├─► Compara valores via hrp.OperadoresComparacao
           ├─► Executa blocos de fluxo de controle
           │
           ▼
     [Objetos Go de hrp.Objeto Gerados]
```
