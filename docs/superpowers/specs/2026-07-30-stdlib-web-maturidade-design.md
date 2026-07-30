# Especificação de Design — Evolução e Maturidade do Módulo Web do Harpia (`stdlib/web`)

> **Data:** 2026-07-30  
> **Status:** Aprovado  
> **Escopo:** `stdlib/web/runtime-web.js`, `cmd/runtime_web_code.go`, `hrp/vdom_backend.go` e suíte de testes.

---

## 1. Visão Geral e Objetivos

O módulo `stdlib/web` fornece o runtime de renderização do Virtual DOM, reatividade fina por Sinais e suporte a SSR do Harpia. Esta especificação eleva o ecossistema Web do Harpia para maturidade total de nível industrial, trazendo paridade com frameworks como SolidJS, React 19 e Vue 3.

---

## 2. Arquitetura e Componentes

### 2.1 Reatividade Fina Avançada
- **`semRastreio(fn)` / `untrack(fn)`**: Executa uma função sem registrar dependências de leitura de sinais no `efeitoAtivo` atual.
- **`lote(fn)` / `batch(fn)`**: Agrupa múltiplas mutações de sinais em um único ciclo de atualização síncrono, evitando re-execuções intermediárias antes do término do bloco.
- **Gerenciamento de Ciclo de Vida de Efeitos**: `efeito(fn)` armazena e executa qualquer função de limpeza (`cleanup`) retornada por `fn` quando o efeito re-executa ou é cancelado.

### 2.2 Virtual DOM, SVG e Fragmentos
- **Suporte Nativo a SVG**: Reconhecimento automático de tags SVG (`svg`, `path`, `rect`, `circle`, `g`, `polygon`, `line`, `polyline`, `ellipse`, `use`, `text`, `image`) no VDOM frontend (`criarNo`) utilizando `document.createElementNS("http://www.w3.org/2000/svg", tag)`.
- **`Fragmento` / `Fragment`**: Suporte a agrupamento de múltiplos nós filhos sem adicionar elementos contêineres HTML extras no DOM (`document.createDocumentFragment()`).
- **Atributo `ref` e Hooks de Ciclo de Vida**:
  - `criarRef(valorInicial)` cria um objeto mutável `{ current: valorInitial }`.
  - Atribuição do elemento DOM nativo ao atributo `ref` (suportando refs do tipo função ou objeto ref).
  - Lifecycle hooks: `aoMontar(fn)` e `aoDesmontar(fn)`.

### 2.3 SSR, Hidratação e Gestão do Head (`usarCabecalho`)
- **`usarCabecalho(opcoes)`**: Hook e componente nativo para manipular dinamicamente `<title>`, `<meta name="...">` e tags OpenGraph no cliente e renderizar no SSR.
- **Hidratação Resiliente (`hidratar`)**: Procura elementos HTML existentes do SSR, vincula reatividade e ouvintes de eventos dinâmicos e reporta avisos amigáveis em caso de divergência sem quebrar a aplicação.

### 2.4 Ferramentas de Desenvolvedor (DevTools)
- **`window.__HARPIA_DEVTOOLS__`**: Ponto de entrada para inspeção de sinais ativos, componentes montados, métricas de renderização e erros em tempo real.

---

## 3. Pleno Alinhamento entre Frontend JS e Backend Go SSR

O compilador Harpia embute o runtime web no binário via `cmd/runtime_web_code.go` (`FallbackRuntimeWebJS`). As alterações em `stdlib/web/runtime-web.js` são mantidas estritamente sincronizadas com `cmd/runtime_web_code.go`. O renderizador backend Go em `hrp/vdom_backend.go` suporta auto-fechamento de tags SVG e sanitização de atributos contra XSS.

---

## 4. Plano de Testes

1. **Testes JS (`stdlib/web/test_vdom.js`)**:
   - `semRastreio` e `lote`.
   - `criarRef` e binding do atributo `ref`.
   - Renderização de elementos SVG via `createElementNS`.
   - Fragmentos com múltiplos nós.
   - Limpeza de efeitos e temporizadores (`usarIntervalo`, `usarAtraso`).
2. **Testes Go (`go test ./...`)**:
   - Testes de compilação, execução e SSR em `cmd`, `hrp` e `tests`.
