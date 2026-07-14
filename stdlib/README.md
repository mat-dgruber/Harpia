# Biblioteca Padrão (`stdlib` do Portuscript)

A **Biblioteca Padrão** (do inglês, *Standard Library* ou simplesmente `stdlib`) do **Portuscript** é um conjunto de módulos utilitários robustos implementados diretamente em Go. Eles são expostos nativamente para os scripts Portuscript de forma embutida ou via importação explícita.

Estes módulos estendem as capacidades básicas da linguagem, permitindo que os programadores realizem operações matemáticas avançadas, coletem dados do sistema operacional hospedeiro, estilizem saídas de terminal com cores e até gerenciem conectividade básica de redes de computadores.

---

## 📖 Índice

1. [Arquitetura de Registro Automático](#-arquitetura-de-registro-automático)
2. [Módulo Central: `embutidos`](#-módulo-central-embutidos)
3. [Módulo: `matematica`](#-módulo-matematica)
4. [Módulo: `sistema`](#-módulo-sistema)
5. [Módulo: `colorize`](#-módulo-colorize)
6. [Módulo: `soquete`](#-módulo-soquete)
7. [Exemplo Completo de Uso de Módulos](#-exemplo-completo-de-uso-de-módulos)

---

## 🏗️ Arquitetura de Registro Automático

A integração de novos pacotes em Go com a VM do Portuscript foi desenhada para ser o mais modular e desacoplada possível. 

No arquivo agregador central `stdlib.go`, é feito o uso do mecanismo de **importação anônima** (ou blank import `_`):

```go
package stdlib

import (
    _ "github.com/natanfeitosa/portuscript/stdlib/colorize"
    _ "github.com/natanfeitosa/portuscript/stdlib/embutidos"
    _ "github.com/natanfeitosa/portuscript/stdlib/matematica"
    _ "github.com/natanfeitosa/portuscript/stdlib/sistema"
    _ "github.com/natanfeitosa/portuscript/stdlib/soquete"
)
```

### Como e Por que Funciona:
- **Funções `init()`**: Cada subpacote declara uma função especial `init()` em seu código Go. Esta função é executada uma única vez, de forma prioritária, assim que o interpretador carrega o pacote correspondente.
- **Tabela de Módulos Globais**: Dentro do `init()`, cada módulo se autopopula com suas funções nativas e constantes, e se registra na máquina virtual chamando a função centralizada `ptst.RegistraModuloImpl()`.
- **Desacoplamento Máximo**: Adicionar ou remover um módulo nativo na distribuição oficial do Portuscript não exige alterações estruturais na VM ou no parser. Basta criar a pasta e declarar seu blank import em `stdlib.go`.

---

## 🧩 Módulo Central: `embutidos`

O subpacote `embutidos` é a base da linguagem. Diferente dos outros pacotes, seus símbolos e métodos **não necessitam de importação**. Eles são injetados diretamente na tabela global de símbolos e ficam imediatamente disponíveis em qualquer arquivo de código.

### Principais Recursos Embutidos:

- **`escreva(args...)`**: Imprime representações textuais na saída padrão do terminal.
- **`leia(mensagem?)`**: Pausa a execução para coletar uma entrada textual digitada pelo usuário no terminal.
- **`tamanho(objeto)`**: Retorna a contagem de elementos de uma lista, caracteres de uma string ou chaves de um dicionário.
- **`int(objeto)`**: Tenta fazer o casting (conversão) forçado do objeto para o tipo inteiro de 64 bits.
- **`texto(objeto)`**: Converte e devolve a representação de string de qualquer tipo primitivo da linguagem.

---

## 📐 Módulo: `matematica`

O módulo `matematica` disponibiliza constantes físicas e rotinas de aritmética avançada para complementar os operadores básicos de soma, subtração, multiplicação e divisão da linguagem.

- **Importação obrigatória**: `importar matematica`

### Constantes Disponíveis:

- **`matematica.PI`**: Constante matemática aproximada $\approx 3.141592653589793$.
- **`matematica.E`**: Constante de Euler e base dos logaritmos naturais $\approx 2.718281828459045$.

### Métodos e Funções:

| Função | Assinatura | Descrição Técnica |
| :--- | :--- | :--- |
| `absoluto` | `absoluto(numero)` | Retorna o valor absoluto (magnitude real sem sinal) de um número real (`math.Abs`). |
| `piso` | `piso(decimal)` | Arredonda o número de ponto flutuante para baixo, retornando o menor número inteiro mais próximo (`math.Floor`). |
| `teto` | `teto(decimal)` | Arredonda o número para cima, retornando o menor inteiro maior ou igual (`math.Ceil`). |
| `potencia` | `potencia(base, expoente)` | Calcula a potenciação correspondente a base elevada ao expoente informado (`math.Pow`). |
| `raiz` | `raiz(radicando, indice?)` | Retorna a raiz do radicando pelo índice. Se o índice for omitido, calcula a raiz quadrada por padrão. Implementado sob conversão de expoente fracionário ($radicando^{1.0/indice}$). |

---

## 💻 Módulo: `sistema`

O módulo `sistema` expõe informações sobre a infraestrutura e o ambiente de software em execução.

- **Importação obrigatória**: `importar sistema`

### Constantes Disponíveis:

- **`sistema.NOME`**: Retorna o identificador textual do sistema operacional hospedeiro (ex: `"darwin"` para macOS, `"linux"` para sistemas baseados em Linux, `"windows"` para Windows).
- **`sistema.ARQUITETURA`**: Retorna o tipo de processador/arquitetura onde o interpretador está rodando (ex: `"amd64"`, `"arm64"`, `"386"`).

---

## 🎨 Módulo: `colorize`

O módulo `colorize` (colorização de console) adiciona capacidades de estilização gráfica para que os desenvolvedores criem saídas ricas em cores no terminal usando sequências de escape ANSI.

- **Importação obrigatória**: `importar colorize`

---

## 🔌 Módulo: `soquete`

O módulo `soquete` fornece uma interface direta de controle de sockets de rede de baixo nível, permitindo criar scripts que conversam com servidores externos ou gerenciam conexões soquete ativas.

- **Importação obrigatória**: `importar soquete`

---

## 📝 Exemplo Completo de Uso de Módulos

Abaixo está um exemplo completo de um arquivo em Portuscript (`.pt`) ilustrando a sintaxe de carregamento e o uso conjunto de múltiplos recursos da `stdlib`:

```portuscript
# Exemplo de uso conjunto da biblioteca padrão

importar matematica
importar sistema

# 1. Usando recursos de sistema para diagnóstico
escreva("Rodando em: " + sistema.NOME + " (" + sistema.ARQUITETURA + ")")

# 2. Operações matemáticas de geometria básica
raio = 5.0
area = matematica.PI * matematica.potencia(raio, 2)
escreva("Área do círculo de raio 5: " + texto(area))

# 3. Arredondamento de valores decimais
escreva("Valor teto de area: " + texto(matematica.teto(area)))
escreva("Valor piso de area: " + texto(matematica.piso(area)))

# 4. Cálculo de raízes
escreva("Raiz quadrada de 144: " + texto(matematica.raiz(144)))
escreva("Raiz cúbica de 27: " + texto(matematica.raiz(27, 3)))
```
