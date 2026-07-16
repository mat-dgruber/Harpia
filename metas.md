# Metas e Direcionamento do Harpia

Todas as metas, planejamentos detalhados de curto/médio prazo e escolhas de design arquitetural foram unificados no nosso documento oficial de Roadmap.

Consulte o arquivo completo de planejamento em:
👉 **[ROADMAP.md](file:///Users/matheus.diniz_1/Documents/GitHub/harpia/harpia/ROADMAP.md)**

## Resumo das Fases do Roadmap
*   **Fase 1 — Núcleo Sólido:** Classes com herança simples, sistema de tipos opcional, testes nativos, constantes, erros detalhados em português com IA e operador de canal (pipes `|>`).
*   **Fase 2 — VM de Pilha + Bytecode:** Compilador AST para bytecode `.hrpc`, modelo de valor eficiente NaN-boxing e Garbage Collector por contagem de referências.
*   **Fase 3 — Stdlib Backend Real:** Módulo HTTP completo com middlewares e injeção de dependências, banco de dados (SQLite/Postgres) com Query Builder nativo e corotinas assíncronas.
*   **Fase 4 — Frontend SPA (Concluída):** Transpilação para JavaScript com runtime web próprio de ~5-8KB (sem dependências), reatividade via sinais/efeitos, estilização declarativa em português, roteamento de SPA baseado em arquivos, SSR com hidratação e estado global nativo (`armazem`).
*   **Fase 5 — Tooling & Ecossistema:** CLI com scaffolding de Clean Arch/DDD, console interativo TUI (Bubbletea), LSP oficial e gerador de diagramas de arquitetura.
