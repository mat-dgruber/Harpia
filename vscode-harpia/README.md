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

