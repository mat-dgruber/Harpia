# 📝 Registro de Desenvolvimento — 2026-07-28 (Backlog Geral e Sprint 6)

**Escopo:** Organização, refinamento e consolidação de todo o backlog acumulado contendo correções de bugs, testes de integração e novas capacidades da linguagem.
**Commits gerados:** 4
**Arquivos modificados:** 45

---

## 1. Visão Geral das Alterações

Consolidação de uma série de melhorias em quatro áreas fundamentais do ecossistema Harpia:
1. **VM & Interpretador (`hrp/`):** Correção do pânico de ponteiros nulos no acesso a propriedades de objetos `Nulo`, suporte expandido a variáveis de ambiente (.env e .harpia.env), preservação dinâmica de coordenadas de arquivo para tracebacks de erro, operadores de desigualdade personalizados e métodos embutidos de string.
2. **Biblioteca Padrão & Web Runtime (`stdlib/`):** API de eventos baseada em Regex no runtime JS (case-insensitive), vinculação de atributos nativos, novo componente visual `NotificacaoToast`, exportação OpenAPI 3.0 estruturada e classe de compatibilidade `Promise`.
3. **CLI & Compilação (`cmd/`):** Resolução recursiva de caminhos relativos em imports web, flag `-d` / `--diretorio` no comando `servir` para servidores estáticos de produção e desenvolvimento, melhorias de UX no terminal (REPL e TUI) e parser resiliente a quebras de linha em assinaturas de funções.
4. **Testes de Integração (`tests/`):** Cobertura ponta a ponta de requisições CORS / preflight OPTIONS, documentação automática de endpoints `/docs` e `/docs.json`, e testes robustos das novas regras estáticas de linter.

---

## 2. Detalhemento por Commit

### 1. `refactor(vm): aprimora tracebacks, suporte a .env, tratamento de Nulo e rpc`
- **Razão:** Prevenir pânicos inesperados de Go (nil pointer errors) ao acessar propriedades ou índices em variáveis nulas, e assegurar que erros gerados dentro de módulos referenciem os arquivos originais.
- **Solução:**
  - Adicionado suporte nativo a `.env` caso `.harpia.env` não esteja disponível.
  - Implementação de interceptador na VM para que ao tentar ler ou escrever propriedades/índices em `Nulo`, um erro explícito do tipo `AtributoErro` ou `TipagemErro` seja lançado de forma amigável ao desenvolvedor.
  - Snapshot de coordenadas do código em `Funcao` na fase de declaração, garantindo tracebacks precisos.

### 2. `feat(stdlib): adiciona api de eventos regex, openapi estruturado e promise`
- **Razão:** Habilitar suporte completo a eventos case-insensitive no frontend (ex: `AoClicar`, `aoerror`) e garantir geração limpa e tipada de OpenAPI 3.0 a partir das assinaturas de rotas e metadados no backend.
- **Solução:**
  - Tratador inteligente regex no `runtime-web.js` para ligar eventos livremente.
  - Motor embutido no compilador/stdlib para ler anotações de comentários `///` no Go e injetar automaticamente no OpenAPI JSON e console `/docs`.
  - Implementação do objeto `Promise` nativo no backend Go para espelhamento de contratos assíncronos.

### 3. `feat(cli): estende compilacao de web, dev server e depuracao`
- **Razão:** Tornar a experiência de desenvolvimento e compilação de SPAs Harpia idêntica à do ecossistema Node moderno, mantendo a leveza sem dependências pesadas.
- **Solução:**
  - Comando `servir` agora aceita a flag `-d`/`--diretorio` para agir como servidor estático robusto (com CORS integrado por padrão).
  - Resolução inteligente de caminhos relativos em subpastas de componentes transpiliados em `dist/`.
  - Tratador do parser que ignora novas linhas entre argumentos de funções, permitindo formatações flexíveis e legíveis de código longo.

### 4. `test(http): adiciona testes integrados de rpc, cors e docs`
- **Razão:** Validar todas as implementações anteriores de forma robusta e persistente.
- **Solução:**
  - `TestRPCCORS` valida requisições CORS com Preflight OPTIONS retornando headers de origem cruzada corretos.
  - `TestDocumentacaoAutomaticaHttp` inicia um servidor HTTP simulado, requisita `/docs` e `/docs.json` e assevera a integridade de seus retornos.

---

## 3. ✅ Status de Validação

- `go test ./...` executado na raiz com sucesso absoluto (100% dos testes passando em < 25s).
- Correção de bugs visuais de syntax highlight e do cURL na geração de snippets totalmente resolvida.
