# Especificação Gramatical (`gramatica` do Harpia)

O diretório `gramatica` abriga os arquivos formais de especificação de sintaxe e léxico da linguagem **Harpia**. As regras são descritas no formato padrão da ferramenta geradora de compiladores **ANTLR4** (`.g4`), dividindo-se em:

1. **`HarpiaLexer.g4`**: Define a especificação léxica formal (tokens, literais, operadores, delimitadores, bitwise, identificadores e ignoráveis).
2. **`HarpiaParser.g4`**: Define as regras sintáticas e a gramática de precedência gramatical (AST - Árvore de Sintaxe Abstrata), modelando desde a declaração raiz até as expressões aninhadas.

---

## 📖 Índice

1. [Papel da Gramática no Compilador](#-papel-da-gramática-no-compilador)
2. [Especificação Léxica Detalhada (`HarpiaLexer.g4`)](#-especificação-léxica-detalhada-harpialexerg4)
   - [Constantes e Literais Primitivos](#constantes-e-literais-primitivos)
   - [Palavras-Chave Reservadas](#palavras-chave-reservadas)
   - [Operadores de Reatribuição e Matemáticos](#operadores-de-reatribuição-e-matemáticos)
   - [Delimitadores de Sintaxe e Bitwise](#delimitadores-de-sintaxe-e-bitwise)
3. [Especificação Sintática Detalhada (`HarpiaParser.g4`)](#-especificação-sintática-detalhada-harpiaparserg4)
   - [Ponto de Entrada e Declarações](#ponto-de-entrada-e-declarações)
   - [Controle de Fluxo, Funções e Imports](#controle-de-fluxo-funções-e-imports)
   - [Precedência de Operadores (Hierarquia de AST)](#precedência-de-operadores-hierarquia-de-ast)
4. [Dicas para Visualização e Desenvolvimento](#-dicas-para-visualização-e-desenvolvimento)

---

## 🎯 Papel da Gramática no Compilador

> **Nota de Design Importante**:  
> Embora os arquivos `.g4` descrevam formalmente a gramática no padrão ANTLR4, o compilador físico do Harpia **não utiliza código gerado automaticamente** pelo ANTLR. 
> 
> Por questões de performance de execução extraordinária, flexibilidade de recursos de rede, tratamento correto de strings UTF-8 multibyte de forma resiliente e, crucialmente, para emitir mensagens de diagnósticos de erros ricos inteiramente em Português (com indicação precisa de linha, coluna e token sob falha), o **Lexer e o Parser foram escritos totalmente à mão em Go** (localizados nos pacotes correspondentes `/lexer` e `/parser`).

Portanto, os arquivos contidos nesta pasta servem como a **especificação técnica formal de referência** de como a linguagem é projetada e estruturada, devendo ser estritamente seguidos por qualquer implementação de compilador, interpretador, ferramenta de análise estática ou servidor LSP.

---

## 🔤 Especificação Léxica Detalhada (`HarpiaLexer.g4`)

O analisador léxico (`HarpiaLexer.g4`) quebra o código fonte em pequenas unidades lógicas atômicas chamadas **Tokens**.

### Constantes e Literais Primitivos:
- `FALSO`: Representa o literal booleano `Falso`. No runtime do Go, é resolvido como um objeto de escopo singleton e imutável.
- `VERDADEIRO`: Representa o literal booleano `Verdadeiro`. No runtime do Go, é resolvido como um objeto de escopo singleton e imutável.
- `NULO`: Representa o valor nulo `Nulo`, denotando a ausência de valor ou tipo estrutural.

### Palavras-Chave Reservadas:
- **Declaração**: `var` (variáveis mutáveis), `const` (constantes imutáveis), `func` (definição de funções).
- **Controle de Fluxo**: `se` (condicional), `senao` (ramificação alternativa), `para` (laço de repetição), `em` (conector para iteração em coleções).
- **Sub-Instruções**: `retorne` (retorno de funções), `pare` (break), `continue` (continue), `assegura` (asserção de teste/garantia lúdica).
- **Modularidade**: `importe` (diretiva de importação), `de` (from - qualificador de pacotes).
- **Orientação a Objetos**: `nova` (instanciação de classes).

### Operadores de Reatribuição e Matemáticos:
- `OPERADOR_REATRIBUICAO`: Reúne operadores compostos como `=`, `+=`, `-=`, `*=`, `@=`, `/=`, `%=`, `&=`, `|=`, `^=`, `<<=`, `>>=`, `**=`, `//=`.
- Operadores Básicos: `MAIS` (`+`), `MENOS` (`-`), `ASTERISCO` (`*`), `POTENCIA` (`**`), `DIVISAO` (`/`), `DIVISAO_INTEIRA` (`//`), `MODULO` (`%`).
- Operadores Relacionais: `MENOR_QUE` (`<`), `MENOR_OU_IGUAL` (`<=`), `MAIOR_QUE` (`>`), `MAIOR_OU_IGUAL` (`>=`), `IGUAL_IGUAL` (`==`), `DIFERENTE` (`!=`).

### Delimitadores de Sintaxe e Bitwise:
- Delimitadores: `ABRE_PARENTESES` (`(`), `FECHA_PARENTESES` (`)`), `PONTO_E_VIRGULA` (`;`), `VIRGULA` (`,`), `ABRE_CHAVES` (`{`), `FECHA_CHAVES` (`}`), `DOIS_PONTOS` (`:`), `PONTO` (`.`), `ABRE_COLCHETES` (`[`), `FECHA_COLCHETES` (`]`).
- Operadores de Bits (Bitwise): `OU_BIT_A_BIT` (`|`), `EX_OU_BIT_A_BIT` (`^`), `E_BIT_A_BIT` (`&`), `NAO_BIT_A_BIT` (`~`), `DESLOC_ESQUERDA` (`<<`), `DESLOC_DIREITA` (`>>`).
- Identificadores (`ID`): Permite letras ou sublinhas (`_`) na primeira posição, seguidos por letras ou dígitos decimais.
- Literais de Texto (`TEXTO`): Strings entre aspas duplas, com suporte completo a barras de escape.
- Espaços em Branco (`WS`): Espaços, tabulações e quebras descartados via diretiva `-> skip`.

---

## 🌳 Especificação Sintática Detalhada (`HarpiaParser.g4`)

A gramática do interpretador (`HarpiaParser.g4`) recebe os tokens produzidos pelo Lexer e constrói a Árvore de Sintaxe Abstrata (AST), validando o pareamento de delimitadores e a precedência operacional.

### Ponto de Entrada e Declarações:
- `programa`: A regra de parsing raiz, que aceita uma sequência opcional de declarações terminada pelo token especial `EOF`.
- `declaracao`: Bifurca-se em `declaracao_composta` (estruturas complexas com blocos de escopo delimitados por chaves) ou `declaracao_simples` (instruções unilinha ou terminadas em ponto e vírgula).
- `atribuicao`: Unifica `atribuicao_variavel` (`var x = 10;` ou com indicação de tipo `var x: Inteiro;`) e `atribuicao_constante` (`const PI = 3.14;`).

### Controle de Fluxo, Funções e Imports:
- `declaracao_importacao`: Mapeia importações simples (`importe "matematica";`) ou parciais/desestruturadas (`de "matematica" importe PI, raiz;`).
- `declaracao_funcao`: Define funções com parâmetros tipados opcionalmente (`func somar(a, b: Inteiro) { ... }`).
- `declaracao_se` / `declaracao_senao_se` / `declaracao_senao`: Modela o comportamento condicional de ramificação.
- `declaracao_para`: Modela laços iterativos de repetição de coleções (ex: `para item em lista`).
- `bloco`: Representa um escopo fechado por chaves `{}` que abriga uma sequência de declarações.

### Precedência de Operadores (Hierarquia de AST):
A gramática organiza expressões aninhadas do operador de menor prioridade (resolvido por último) até o de maior prioridade (resolvido primeiro):

| Nível de Precedência | Categoria | Regra ANTLR4 | Operadores | Descrição Sintática |
| :---: | :--- | :--- | :--- | :--- |
| **11** | Disjunção Lógica | `disjuncao` | `ou` | Operador OU lógico. |
| **10** | Conjunção Lógica | `conjuncao` | `e` | Operador E lógico. |
| **9** | Negação Lógica | `inversao` | `nao` | Operador unário de negação lógica. |
| **8** | Comparação Relacional | `comparacao` | `==`, `!=`, `<`, `<=`, `>`, `>=`, `em` | Comparações de igualdade, magnitude e presença (`em`). |
| **7** | OU Bit a Bit | `ou_bitabit` | `\|` | Operador binário bitwise OR. |
| **6** | XOR Bit a Bit | `exou_bitabit`| `^` | Operador binário bitwise XOR (Ou Exclusivo). |
| **5** | E Bit a Bit | `e_bitabit` | `&` | Operador binário bitwise AND. |
| **4** | Deslocamento de Bits | `deslocamento`| `<<`, `>>` | Deslocamento de bits binário para esquerda/direita. |
| **3** | Aritmética Aditiva | `arit_basica` | `+`, `-` | Soma, subtração ou concatenação de strings. |
| **2** | Aritmética Multiplicativa| `termo` | `*`, `/`, `//`, `%` | Multiplicação, divisão real, divisão inteira e resto. |
| **1** | Sinais Unários | `fator` | `+`, `-`, `~` | Identidade unária, sinal negativo e inversão binária. |
| **0** | Exponenciação | `potencia` | `**` | Exponenciação aritmética (base elevado a expoente). |

---

## 🛠️ Dicas para Visualização e Desenvolvimento

Para programadores que desejam analisar o grafo de sintaxe visualmente ou editar as regras da gramática no editor **VS Code**, é altamente recomendada a instalação da extensão oficial:

- **Nome**: ANTLR4 grammar syntax support and VM
- **ID**: `mike-lischke.vscode-antlr4`
- **Link**: [Extension Marketplace](https://marketplace.visualstudio.com/items?itemName=mike-lischke.vscode-antlr4)

### Recursos da Extensão:
- Realce de sintaxe colorido para arquivos `.g4`.
- Geração e exibição gráfica dinâmica de grafos sintáticos (Parse Tree) para depurar e validar novas regras gramaticais da linguagem de forma visual e em tempo de projeto.
