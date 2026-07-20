# Extensão Harpia para VS Code

Esta extensão ativa o suporte oficial de desenvolvimento à linguagem Harpia diretamente no VS Code.

## Recursos Disponíveis

- **Colorização de Sintaxe Rica**: Palavras-chave de fluxo de controle, números, strings, comentários, tipos estritos (`Texto`, `Inteiro`, `Decimal`), tags JSX estruturais compiladas (`<se>` e `<para>`) e modificadores de ações reativas (`_prevenir`, `_parar`) coloridos automaticamente.
- **Autocomplete Inteligente**: Autocompletação (`CompletionItemProvider`) contextual de diretivas de importação de módulos nativos da stdlib (`web`, `resiliencia`, `telemetria`, `bd`, `ia`) e suas funções ao digitar.
- **Ajuda de Parâmetros (Signature Help)**: Dicas flutuantes interativas (`Ctrl+Shift+Space`) que auxiliam no preenchimento de argumentos de funções nativas ou locais do usuário.
- **Hover de Documentação & Resolução Local**: Mostra a documentação detalhada das palavras-chave padrão em português, analisa assinaturas e comentários (`#` ou `//`) de tipos locais do usuário, e resolve caminhos de importação e cadeias de herança de classes (`estende`) em múltiplos arquivos.
- **Navegação Rápida (F12 — Go to Definition)**: Pule instantaneamente para a declaração original de variáveis, funções ou classes no arquivo e na linha exatos (suportando definições importadas).
- **Code Lenses para Testes Unitários**: Link flutuante interativo **`▶️ Executar teste`** gerado acima de cada bloco `testar` do código para rodar cenários isoladamente com um clique.
- **Seletor de Cores Dinâmico (Color Picker)**: Detecta cores hexadecimais, formatos `rgb()`, `rgba()`, `hsl()`, `hsla()` e nomes de cores nomeadas CSS clássicas diretamente no código, exibindo uma prévia colorida e abrindo o seletor nativo da IDE para ajustes interativos de cor.
- **Formatação Automática**: Ao salvar ou via `Alt+Shift+F`, delega síncronamente para `harpia formatar` na CLI para garantir estilo perfeito do código.
- **Depurador Interativo (DAP)**: Comunicação e pontes de breakpoint via DAP (`harpia depurar`).
- **Painel de Ferramentas & Atalhos**: Webview lateral (ativável em `Ctrl+Alt+H` ou `Cmd+Alt+H`) para gerenciar o servidor de desenvolvimento (`servir`), empacotar a aplicação, simular cargas (`stressar`) e gerar scaffolding completo de rota, componente ou modelo de domínio com abertura automática do arquivo gerado.
- **Tema Harpia**: Paleta Catppuccin otimizada para Harpia, com cores distintas para classes, componentes, funções, propriedades e estilos.

## Como Testar Localmente

1. Abra a pasta `vscode-harpia` no VS Code.
2. Abra um terminal dentro da pasta `vscode-harpia` e instale as dependências:

   ```bash
   npm install
   ```

3. Pressione `F5` para iniciar uma nova janela do VS Code (Janela de Desenvolvimento de Extensão) com a extensão ativa.
4. Crie ou abra qualquer arquivo com extensão `.hrp` ou `.ptst` e comece a programar em Português!

## Como Publicar e Atualizar

A extensão é distribuída oficialmente no **VS Code Marketplace** sob o publisher `harpia`.

### 1. Requisitos para Publicadores
Para publicar atualizações, você precisa de:
- Acesso à conta de desenvolvedor no [Marketplace da Microsoft](https://marketplace.visualstudio.com/manage) vinculada ao publisher `harpia`.
- Um **Personal Access Token (PAT)** criado no [Azure DevOps](https://dev.azure.com/) configurado com:
  - **Organization**: `All accessible organizations`
  - **Scopes**: `Custom defined` -> `Marketplace (Publish)`

### 2. Primeiro Login (Realizado com Sucesso ✅)
Para associar sua sessão local com a credencial de publicação:
```bash
npm install -g @vscode/vsce
vsce login harpia
```

### 3. Publicando Atualizações (SemVer)
Para gerar uma nova tag e publicar atualizações automáticas:
```bash
# Para correções de bugs (ex: 0.1.0 -> 0.1.1)
vsce publish patch

# Para novos recursos retrocompatíveis (ex: 0.1.0 -> 0.2.0)
vsce publish minor

# Para grandes mudanças ou quebras de compatibilidade (ex: 0.1.0 -> 1.0.0)
vsce publish major
```

### 4. Empacotamento Local (Distribuição Manual)
Se preferir testar ou distribuir offline através de um arquivo `.vsix`:
```bash
vsce package
```
O arquivo `.vsix` gerado poderá ser instalado arrastando-o para dentro do VS Code ou clicando em *Instalar de VSIX...* no painel de extensões.
