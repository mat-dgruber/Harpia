# Especificação de Design: Harpia CLI - Fase A (DX, Formatação AST & Stream UI)

## Objetivo
Elevar a qualidade e maturidade da experiência de desenvolvimento no terminal da CLI do Harpia, implementando um formatador de código robusto guiado por AST/Tokens (`harpia formatar`), gerador de diffs unificados (`--diff`), verificação de integridade para CI (`--verificar`) e suporte centralizado a estilos e detecção de TTY/`NO_COLOR`.

## Componentes

### 1. Centralizador de Terminal & UI (`cmd/ui.go`)
- **Responsabilidade**: Abstrair e gerenciar a saída formatada de terminal, cores ANSI e animações.
- **Funcionalidades**:
  - `IsTerminal(w io.Writer) bool`: Verifica se o writer é um terminal TTY interativo.
  - `ColorEnabled() bool`: Checa variáveis de ambiente `NO_COLOR` e `FORCE_COLOR`.
  - Helpers de formatação com atalhos de cores (`UI.Verde()`, `UI.Vermelho()`, `UI.Amarelo()`, `UI.Negrito()`).

### 2. Formatador de Código Baseado em AST & Tokens (`cmd/formatar.go`)
- **Responsabilidade**: Formatar o código Harpia sem alterar a árvore sintática ou o comportamento semântico.
- **Funcionalidades**:
  - Analisa a AST do arquivo `.hrp`/`.pt` e reestrutura recuos (2 espaços por nível de profundidade de bloco `{}`).
  - Remove espaços em branco ao final das linhas (*trailing spaces*).
  - Garante apenas uma quebra de linha no final do arquivo.
  - Normaliza declarações de variáveis, funções e estruturas de controle (`se`, `enquanto`, `para`, `escolha`).

### 3. Exibição de Diff Unificado (`--diff`)
- **Responsabilidade**: Comparar visualmente o arquivo original e o arquivo formatado.
- **Saída**: Formato Git Unified Diff (`--- a/arquivo.hrp`, `+++ b/arquivo.hrp`, `@@ -1,5 +1,5 @@`), com coloração sintática caso o terminal suporte cores (linhas removidas em vermelho, adicionadas em verde).

### 4. Integração com CI (`--verificar`)
- **Responsabilidade**: Garantir conformidade em esteiras de integração contínua sem alterar arquivos em disco.
- **Comportamento**: Retorna código de erro `1` com o número de arquivos desformatados caso encontre divergências.

## Testes Automatizados
- Testes unitários em `cmd/formatar_test.go` validando:
  - Formatação de blocos aninhados (`se`, `enquanto`, `funcao`).
  - Geração de unified diffs com `--diff`.
  - Comportamento de saída com `--verificar` em código formatado vs desformatado.
  - Respeito à flag `NO_COLOR`.
