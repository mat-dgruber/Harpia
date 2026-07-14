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
- **Erros Didáticos e IA:** Mensagens de erro visuais com sublinhado e um explicador interativo nativo (`portuscript erro explicar`) integrado a LLM local.
- **Tratador de Canais (`|>`):** Operador nativo pipe para processamento sequencial de dados.
- **Reatividade Nativa:** Sinais, efeitos e estado global (`armazem`) nativos no núcleo.
- **Frontend SPA com WASM:** Compilação direta para WebAssembly com substituição de HTML/JS, estilização declarativa em português e roteador baseado em arquivos.
- **Estrutura de Arquitetura Assistida:** CLI que gera esqueletos de projetos separados por Clean Architecture e DDD.
- **Tradutor de Código:** CLI nativo (`portuscript traduzir`) para exportar o código PortuScript para JavaScript, Python ou Go.

---

## 🗺️ Roadmap de Desenvolvimento

O planejamento detalhado, justificativas técnicas e estratégias de evolução de cada fase estão documentados no nosso guia oficial:
👉 **Consulte o [ROADMAP.md](file:///Users/matheus.diniz_1/Documents/GitHub/portuscript/portuscript/ROADMAP.md)**

---

## 📦 Estrutura do CLI

A CLI do PortuScript foi desenhada em português brasileiro com atalhos fáceis:

- **`portuscript`**: Abre a TUI (Interface Gráfica de Terminal) com REPL, console de depuração e monitor de variáveis em tempo real.
- **`portuscript executar [arquivo.ptst]`**: Executa um arquivo diretamente na VM.
- **`portuscript testar [caminho]`**: Executa a suíte de testes nativa de forma automatizada.
- **`portuscript checar`**: Executa o validador semântico estático (linter) no código.
- **`portuscript novo-monolito [nome]`**: Inicializa um projeto Clean Arch/DDD completo.
- **`portuscript novo-backend [nome]`**: Inicializa um projeto contendo apenas lógica de domínio e banco de dados.
- **`portuscript novo-frontend [nome]`**: Inicializa um projeto contendo apenas componentes de página e estilos.
- **`portuscript servir`**: Roda o servidor local de desenvolvimento com Hot-Reload.
- **`portuscript traduzir [arquivo.ptst] --para=[javascript/python/go]`**: Traduz automaticamente o código.
- **`portuscript diagramar`**: Gera um diagrama de arquitetura do projeto.

---

## 🚀 Instalação e Contribuição

Consulte o arquivo [CONTRIBUTING.md](/CONTRIBUTING.md) para saber como contribuir e ajudar na construção da nossa linguagem brasileira.
