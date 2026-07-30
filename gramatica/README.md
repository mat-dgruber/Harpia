# Especificação Gramatical (`gramatica` do Harpia)

O diretório `gramatica` abriga os arquivos formais de especificação de sintaxe e léxico da linguagem **Harpia**. As regras são descritas no formato padrão da ferramenta geradora de compiladores **ANTLR4** (`.g4`), dividindo-se em:

1. **`HarpiaLexer.g4`**: Define a especificação léxica formal (tokens, literais, palavras-chave reservadas, operadores, delimitadores, bitwise, identificadores Unicode e ignoráveis).
2. **`HarpiaParser.g4`**: Define as regras sintáticas e a gramática de precedência de expressões (AST - Árvore de Sintaxe Abstrata), modelando desde a declaração raiz até estruturas complexas de POO, Pattern Matching, Async/Await e Estilos.

---

## 📖 Índice

1. [Papel da Gramática no Compilador](#-papel-da-gramática-no-compilador)
2. [Especificação Léxica Detalhada (`HarpiaLexer.g4`)](#-especificação-léxica-detalhada-harpialexerg4)
   - [Constantes e Literais Primitivos](#constantes-e-literais-primitivos)
   - [Palavras-Chave Reservadas](#palavras-chave-reservadas)
   - [Operadores de Reatribuição, Matemáticos e Especiais](#operadores-de-reatribuição-matemáticos-e-especiais)
   - [Delimitadores e Identificadores Unicode](#delimitadores-e-identificadores-unicode)
3. [Especificação Sintática Detalhada (`HarpiaParser.g4`)](#-especificação-sintática-detalhada-harpiaparserg4)
   - [Ponto de Entrada e Flexibilidade de Ponto e Vírgula](#ponto-de-entrada-e-flexibilidade-de-ponto-e-vírgula)
   - [Controle de Fluxo, Exceções e Pattern Matching](#controle-de-fluxo-exceções-e-pattern-matching)
   - [Orientação a Objetos, Enums e Interfaces](#orientação-a-objetos-enums-e-interfaces)
   - [Estilos e Testes Nativos](#estilos-e-testes-nativos)
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
- `FALSO`: Literal booleano `Falso`.
- `VERDADEIRO`: Literal booleano `Verdadeiro`.
- `NULO`: Constante nula `Nulo` (ausência explícita de valor).
- `LITERAL_INTEIRO`: Representações inteiras em decimal (`123`), hex (`0x1A`), binário (`0b101`) ou octal (`0o77`).
- `LITERAL_DECIMAL`: Números de ponto flutuante (`123.45`).
- `TEXTO`: Literais textuais entre aspas duplas (`"texto"`), aspas simples (`'texto'`) ou crase (`` `template ${var}` ``).

### Palavras-Chave Reservadas:
- **Declaração**: `var` (mutáveis), `const` (imutáveis), `func` (funções).
- **Controle de Fluxo**: `se`, `senao`, `enquanto`, `para`, `em`.
- **Sub-Instruções**: `retorne`, `pare`, `continue`, `assegura`.
- **Tratamento de Erros**: `tente`, `capture`, `finalmente`.
- **Pattern Matching**: `escolha`, `caso`, `quando`.
- **Modularidade**: `importe`, `de`, `exportar`.
- **POO & Tipagem**: `nova`, `classe`, `estende`, `self`, `estatico`, `enum`, `interface`.
- **Concorrência**: `assincrono`, `aguarde`.
- **Framework & Estilo**: `estilo`, `testar`.

### Operadores de Reatribuição, Matemáticos e Especiais:
- **Reatribuição Agregada**: `=`, `+=`, `-=`, `*=`, `/=`, `//=`, `%=`, `&=`, `|=`, `^=`, `<<=`, `>>=`, `**=`, `??=`, `||=`, `&&=`.
- **Operadores Aritméticos**: `+`, `-`, `*`, `**`, `/`, `//`, `%`.
- **Operadores Relacionais**: `<`, `<=`, `>`, `>=`, `==`, `!=`, `em`.
- **Operadores Lógicos e Nulidade**: `ou`, `e`, `nao`, `??` (coalescência nula), `?.` (chaining opcional), `?` (ternário).
- **Operadores Bitwise**: `|`, `^`, `&`, `~`, `<<`, `>>`.
- **Operadores Especiais**: `|>` (pipe), `...` (spread/rest), `@` (decorador), `->` (seta/lambda).

### Delimitadores e Identificadores Unicode:
- Delimitadores: `(`, `)`, `;`, `,`, `{`, `}`, `:`, `.`, `[`, `]`.
- Identificadores (`ID`): Suporte a letras ASCII e caracteres acentuados Unicode em português (`\u00C0..\u024F`), como `seção`, `módulos`, `área`.

---

## 🌳 Especificação Sintática Detalhada (`HarpiaParser.g4`)

### Ponto de Entrada e Flexibilidade de Ponto e Vírgula:
- `programa`: Regra raiz do compilador terminada por `EOF`.
- `fim_instrucao`: Suporta tanto `;` quanto quebras de linha (`\n`), tornando o ponto e vírgula 100% opcional.

### Controle de Fluxo, Exceções e Pattern Matching:
- `declaracao_se`: Condicionais `se / senao se / senao`.
- `declaracao_enquanto`: Laço condicional `enquanto (condicao) bloco`.
- `declaracao_para`: Laço iterativo `para item em colecao bloco`.
- `declaracao_tente`: Blocos de tratamento de exceção `tente ... capture (erro) ... finalmente ...`.
- `declaracao_escolha`: Pattern matching `escolha (expressao) { caso val quando cond: ... senao: ... }`.

### Orientação a Objetos, Enums e Interfaces:
- `declaracao_classe`: Definição de classes com herança (`classe Filha estende Pai`).
- `declaracao_enum`: Enumerações tipadas (`enum Cor { Vermelho, Verde, Azul }`).
- `declaracao_interface`: Contratos de interface para verificação estática.

### Estilos e Testes Nativos:
- `declaracao_estilo`: Blocos de estilização nativa (`estilo Botao { cor: "azul"; }`).
- `declaracao_teste`: Blocos de testes unitários integrados (`testar "soma de valores" { ... }`).

### Precedência de Operadores (Hierarquia de AST):

| Nível | Categoria | Regra ANTLR4 | Operadores | Descrição Sintática |
| :---: | :--- | :--- | :--- | :--- |
| **13** | Ternário | `ternario` | `? :` | Operador condicional ternário. |
| **12** | Pipe | `pipe` | `\|>` | Encadeamento de funções. |
| **11** | Coalescência Nula | `coalescencia` | `??` | Fallback para valores nulos. |
| **10** | Disjunção Lógica | `disjuncao` | `ou` | Operador OU lógico. |
| **9** | Conjunção Lógica | `conjuncao` | `e` | Operador E lógico. |
| **8** | Negação Lógica | `inversao` | `nao` | Operador unário de negação. |
| **7** | Comparação Relacional | `comparacao` | `==`, `!=`, `<`, `<=`, `>`, `>=`, `em` | Comparações relacionais e pertencimento. |
| **6** | OU Bit a Bit | `ou_bitabit` | `\|` | Bitwise OR. |
| **5** | XOR Bit a Bit | `exou_bitabit` | `^` | Bitwise XOR. |
| **4** | E Bit a Bit | `e_bitabit` | `&` | Bitwise AND. |
| **3** | Deslocamento | `deslocamento` | `<<`, `>>` | Deslocamento de bits. |
| **2** | Aritmética Aditiva | `arit_basica` | `+`, `-` | Adição/Subtração/Concatenação. |
| **1** | Multiplicativa | `termo` | `*`, `/`, `//`, `%` | Multiplicação, divisão real/inteira e resto. |
| **0** | Unário / Concorrência | `fator`, `aguarde_expr` | `+`, `-`, `~`, `aguarde` | Sinais unários e await assíncrono. |
| **Expon.** | Exponenciação | `potencia` | `**` | Exponenciação (base elevado a expoente). |

---

## 🛠️ Dicas para Visualização e Desenvolvimento

Para programadores que desejam analisar o grafo de sintaxe visualmente no editor **VS Code**, utilize a extensão oficial:

- **Nome**: ANTLR4 grammar syntax support and VM
- **ID**: `mike-lischke.vscode-antlr4`
