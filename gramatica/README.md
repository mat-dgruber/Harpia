# Especificação Gramatical (`gramatica` do Harpia)

O diretório `gramatica` abriga os arquivos formais de especificação de sintaxe e léxico da linguagem **Harpia**. As regras são descritas no formato padrão da ferramenta geradora de compiladores **ANTLR4** (`.g4`), dividindo-se em:

1. **`HarpiaLexer.g4`**: Define a especificação léxica (tokens, literais, operadores e identificadores).
2. **`HarpiaParser.g4`**: Define as regras sintáticas e a gramática de precedência gramatical (AST - Árvore de Sintaxe Abstrata).

---

## 📖 Índice

1. [Papel da Gramática no Compilador](#-papel-da-gramática-no-compilador)
2. [Análise Léxica (Palavras Reservadas e Constantes)](#-análise-léxica-palavras-reservadas-e-constantes)
3. [Árvore Sintática e Precedência de Operadores](#-árvore-sintática-e-precedência-de-operadores)
4. [Estruturas de Controle de Fluxo](#-estruturas-de-controle-de-fluxo)
5. [Dicas para Visualização e Desenvolvimento](#-dicas-para-visualização-e-desenvolvimento)

---

## 🎯 Papel da Gramática no Compilador

> **Nota de Design Importante**:  
> Embora os arquivos `.g4` descrevam formalmente a gramática no padrão ANTLR4, o compilador físico do Harpia **não utiliza código gerado automaticamente** pelo ANTLR. 
> 
> Por questões de performance de execução, flexibilidade de recursos de rede, tratamento correto de strings UTF-8 multibyte e, crucialmente, para emitir mensagens de diagnósticos de erros ricos inteiramente em Português, o **Lexer e o Parser foram escritos totalmente à mão em Go** (localizados nos pacotes correspondentes `/lexer` e `/parser`).

Portanto, os arquivos contidos nesta pasta servem como a **especificação técnica formal de referência** de como a linguagem é projetada e estruturada, devendo ser estritamente seguidos por qualquer implementação de interpretador.

---

## 🔤 Análise Léxica (Palavras Reservadas e Constantes)

O analisador léxico (`HarpiaLexer.g4`) quebra o código fonte em pequenas unidades atômicas chamadas **Tokens**.

### Constantes Primordiais:
- `Verdadeiro`: Literal booleano de valor positivo.
- `Falso`: Literal booleano de valor negativo.
- `Nulo`: Representação de ausência de tipo ou valor.

### Palavras-Chave Reservadas:
- **Declaração**: `var` (variáveis mutáveis), `const` (constantes imutáveis), `func` (definição de funções).
- **Controle de Fluxo**: `se` (if), `senao` (else / else if), `para` (laço de repetição), `em` (conector iterativo para-em).
- **Sub-Instruções**: `retorne` (retorno de funções), `pare` (break), `continue` (continue), `assegura` (asserção de teste - assert).
- **Modularidade**: `importe` (import), `de` (from).
- **Classes**: `nova` (instanciação de classe).

---

## 🌳 Árvore Sintática e Precedência de Operadores

A gramática do interpretador (`HarpiaParser.g4`) organiza as expressões matemáticas e lógicas seguindo uma hierarquia estrita de precedência gramatical para evitar ambiguidades em cálculos lógicos e aritméticos.

Abaixo está o mapeamento de **Precedência Operacional**, listado do operador de menor prioridade (resolvido por último) até o de maior prioridade (resolvido primeiro):

| Nível de Precedência | Categoria | Operadores | Descrição Sintática |
| :---: | :--- | :--- | :--- |
| **11** | Disjunção Lógica | `ou` | Operador OU lógico. |
| **10** | Conjunção Lógica | `e` | Operador E lógico. |
| **9** | Negação Lógica | `nao` | Operador unário de negação lógica. |
| **8** | Comparação Relacional | `==`, `!=`, `<`, `<=`, `>`, `>=`, `em` | Comparações lógicas e verificação de presença (`em`). |
| **7** | OU Bit a Bit | `\|` | Operador binário bitwise OR. |
| **6** | XOR Bit a Bit | `^` | Operador binário bitwise XOR (Ou Exclusivo). |
| **5** | E Bit a Bit | `&` | Operador binário bitwise AND. |
| **4** | Deslocamento de Bits | `<<`, `>>` | Deslocamento de bits binário para esquerda/direita. |
| **3** | Aritmética Aditiva | `+`, `-` | Soma, subtração ou concatenação. |
| **2** | Aritmética Multiplicativa | `*`, `/`, `//`, `%` | Multiplicação, divisão real, divisão inteira e resto. |
| **1** | Sinais Unários | `+`, `-`, `~` | Identidade unária, sinal negativo e inversão binária. |
| **0** | Exponenciação | `**` | Exponenciação aritmética (base elevado a expoente). |

---

## 🔄 Estruturas de Controle de Fluxo

### 1. Declaração Condicional (`se` / `senao`)
A estrutura de blocos condicionais não utiliza parênteses na expressão de validação lógica (estilo Go/Rust), exigindo apenas chaves para delimitar o escopo:

```harpia
se x > 10 {
    escreva("Maior que dez")
} senao se x == 10 {
    escreva("É dez")
} senao {
    escreva("Menor que dez")
}
```

### 2. Laço Iterativo (`para` / `em`)
O laço iterativo varre coleções (como listas ou sequências numéricas geradas) de forma simplificada:

```harpia
para item em sequencia(10) {
    escreva(item)
}
```

---

## 🛠️ Dicas para Visualização e Desenvolvimento

Para programadores que desejam analisar o grafo de sintaxe visualmente ou editar as regras da gramática no editor **VS Code**, é altamente recomendada a instalação da extensão oficial:

- **Nome**: ANTLR4 grammar syntax support and VM
- **ID**: `mike-lischke.vscode-antlr4`
- **Link**: [Extension Marketplace](https://marketplace.visualstudio.com/items?itemName=mike-lischke.vscode-antlr4)

### Recursos da Extensão:
- Realce de sintaxe colorido para arquivos `.g4`.
- Geração e exibição gráfica dinâmica de grafos sintáticos (Parse Tree) para depurar e validar novas regras gramaticais da linguagem de forma visual e em tempo de projeto.
