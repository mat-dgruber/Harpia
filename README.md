/

# 🇧🇷 PortuScript

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Documentation Status](https://readthedocs.org/projects/portudoc/badge/?version=latest)](https://portudoc.readthedocs.io/pt/latest/?badge=latest)

**PortuScript** é uma linguagem de programação brasileira, desenvolvida por brasileiros, totalmente em português. Mais do que uma simples linguagem para treino de lógica, o PortuScript visa proporcionar uma experiência de programação acessível, envolvente e extremamente poderosa para a comunidade de língua portuguesa.

A linguagem é projetada sob uma perspectiva de **Ponte de Aprendizado** (facilitando a migração posterior para linguagens como JavaScript, Python e Go) e ao mesmo tempo como um **Ecossistema Completo** capaz de criar aplicações web profissionais (Frontend SPA), servidores robustos (Backend com injeção e banco de dados) e execução local rápida através de sua própria VM.

---

## ✨ Características Principais

- **Brasileira por Natureza:** Sintaxe e palavras-chave totalmente em português.
- **Acessível e Moderna:** Sintaxe de blocos `{}` limpa, sem parênteses em condições.
- **Erros Didáticos e IA:** Mensagens de erro visuais com sublinhado e um explicador interativo nativo (`harpia erro explicar`) integrado a LLM local.
- **Tratador de Canais (`|>`):** Operador nativo pipe para processamento sequencial de dados.
- **Reatividade Nativa:** Sinais, efeitos e estado global (`armazem`) nativos no núcleo.
- **Frontend SPA Reativo:** Transpilação de alta performance para JavaScript moderno com runtime web próprio (~2.2KB), suporte nativo a Sinais, JSX, roteamento SPA por arquivos, estilos declarativos em português e SSR integrado.
- **Estrutura de Arquitetura Assistida:** CLI que gera esqueletos de projetos separados por Clean Architecture e DDD.
- **Tradutor de Código:** CLI nativo (`harpia traduzir`) para exportar o código PortuScript para JavaScript, Python ou Go.

---

## ⚡ Inovações Industriais de Performance e Segurança (v1.x)
A versão `1.x` do Harpia incorpora tecnologias de ponta em engenharia de compiladores e runtimes, garantindo velocidade de execução e segurança de nível industrial:
* **Direct-Threaded JIT VM**: Traduz dinamicamente bytecodes planos em callbacks Go de alta velocidade em tempo de execução, contornando desvios e pulando 100% dos loops de `switch/case` e decodificações na VM de pilha.
* **Super-Instruções e Fusão de Opcodes**: Otimização estática no compilador para fundir operações sequenciais comuns de retorno de variáveis e literais (`OP_RETORNE_CONST`/`OP_RETORNE_VAR`), encolhendo os scripts compilados.
* **Eden Space para Inteiros Curtos**: Pool de alocação rápida para inteiros de `-100` a `2000`, reutilizando instâncias imutáveis pré-boxeadas do Go em $O(1)$, aniquilando alocações redundantes no heap e poupando o Garbage Collector.
* **Modelo Concorrente CSP por Canais**: Suporte nativo à primitiva `Canal` (`nova Canal()`) integrado ao `aguarde` assíncrono para tráfego thread-safe e reativo de mensagens entre processos de background.
* **Modo Sandbox por Bloqueio Físico**: Proteção de isolamento estrito no contexto de execução por meio das flags de restrição ativa `BloquearArquivos` e `BloquearRede`, blindando o sistema operacional contra acessos não autorizados.
* **Recovery Middleware & Timeouts contra Slowloris**: Servidor HTTP imune a interrupções lógicas graças a interceptadores `defer recover()`, e configuração rígida de tempos de limite de leitura e gravação de sockets.

---

## 🗺️ Roadmap de Desenvolvimento

O planejamento detalhado, justificativas técnicas e estratégias de evolução de cada fase estão documentados no nosso guia oficial:
👉 **Consulte o [ROADMAP.md](file:///Users/matheus.diniz_1/Documents/GitHub/harpia/harpia/ROADMAP.md)**

---

## 📦 Estrutura do CLI

A CLI do PortuScript foi desenhada em português brasileiro com atalhos e ferramentas corporativas integradas de fábrica:

*   **`harpia` (ou `tui`)**: Abre a TUI (Interface Gráfica de Terminal) Bubbletea com painéis para REPL, inspetor de VM e console de erros, com atalhos de depuração síncrona passo-a-passo (`F7`/`F8`) e navegação facilitada com `Tab`.
*   **`harpia novo monolito/backend/frontend [nome]`**: Inicializa a árvore de pastas padrão de projetos corporativos com proteção contra sobrescritas.
*   **`harpia crie rota/componente [nome]`**: Assistente assistido (generator) que cria boilerplates estruturados e acoplados prontos para uso.
*   **`harpia executar [arquivo.hrp]`**: Executa códigos sob o interpretador tradicional ou VM de bytecode de alta performance se a flag `--vm` for passada.
*   **`harpia testar [caminho] [--html]`**: Executa testes nativos, e opcionalmente gera o relatório visual `cobertura.html` com as linhas cobertas (fundo verde) e não cobertas (fundo vermelho).
*   **`harpia checar [caminho] [--formato=json]`**: Linter semântico estático preventivo com suporte a diagnósticos no formato JSON de IDE.
*   **`harpia lsp`**: Inicia o servidor de Language Server Protocol com suporte a autocompletar, formatação "On-Save" e linter de Clean Arch inline em tempo de digitação na IDE.
*   **`harpia playground`**: Abre o servidor web local com editor de código e depurador web **escrito 100% em Harpia SPA Reativo (Dogfooding Supremo)**.
*   **`harpia formatar [arquivo.hrp] [-w]`**: Pretty-printer de indentação de 4 espaços com preservação total de comentários e JSX.
*   **`harpia doc [arquivo.hrp] [--formato=html]`**: Extrai comentários com três barras (`///`) e gera relatórios de API interativos em HTML ou Markdown.
*   **`harpia diagramar`**: Varre as relações de importações e cospe o diagrama no formato Mermaid textual, alertando contra quebras de regras de Clean Architecture.
*   **`harpia instalar`**: Resolvedor assíncrono que lê manifestos em português (`pacote.hrp`) e baixa zips de pacotes na pasta local `pt_modulos/`.
*   **`harpia compilar [entrada.hrp] --alvo=web [--otimizar-assets]`**: Transpila o projeto para o browser e opcionalmente comprime e converte imagens de assets locais de forma síncrona para o diretório `/dist`.
*   **`harpia servir [saida_dir] [--porta=3000]`**: Sobe o Dev Server de desenvolvimento integrado com **Hot-Reload em tempo real nativo via Server-Sent Events (SSE)**.
*   **`harpia erro [código] [explicar]`**: Dicionário interativo de erros amigáveis em português, integrado com IA Local (Ollama) para explicações pedagógicas personalizadas.

---

## 🚀 Instalação e Contribuição

Consulte o arquivo [CONTRIBUTING.md](/CONTRIBUTING.md) para saber como contribuir e ajudar na construção da nossa linguagem brasileira.
