# Módulo Web do Harpia (`stdlib/web`) — Plano de Implementação de Maturidade Industrial

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Elevar o ecossistema Web do Harpia (`stdlib/web/runtime-web.js` e `hrp/vdom_backend.go`) ao nível de maturidade industrial com suporte a primitivas de reatividade avançada (`semRastreio`, `lote`), renderização nativa de SVG e Fragmentos, gestão dinâmica de Head (`usarCabecalho`), DevTools e sincronização completa com o compilador Go.

**Architecture:** O runtime web JS embutido (`stdlib/web/runtime-web.js`) fornece o motor de reatividade e Virtual DOM do cliente. O compilador Harpia embute esse runtime via `cmd/runtime_web_code.go`. O backend em `hrp/vdom_backend.go` processa o VDOM em Go para renderização estática SSR e hidratação.

**Tech Stack:** JavaScript (ES2020+), Go (1.24+), Node.js (para testes do VDOM).

## Global Constraints

- Manter o runtime JS leve (~3.5KB minificado), sem dependências externas de terceiros.
- Garantir que `cmd/runtime_web_code.go` permaneça 100% idêntico a `stdlib/web/runtime-web.js`.
- Manter o escopo global do navegador (`window.*`) atualizado para exportação transparente no bundle Harpia.
- Todos os testes JS (`node stdlib/web/test_vdom.js`) e Go (`go test ./...`) devem passar com sucesso.

---

### Task 1: Primitivas Reativas Avançadas (`semRastreio`, `lote`, `aoMontar`, `aoDesmontar`)

**Files:**
- Modify: `stdlib/web/runtime-web.js`
- Test: `stdlib/web/test_vdom.js`

**Interfaces:**
- Produces: `semRastreio(fn)`, `untrack(fn)`, `lote(fn)`, `batch(fn)`, `aoMontar(fn)`, `aoDesmontar(fn)`

- [ ] **Step 1: Escrever teste com falha para `semRastreio` e `lote`**

No arquivo `stdlib/web/test_vdom.js`, adicione chamadas para `semRastreio` e `lote`:

```javascript
import { semRastreio, lote, aoMontar, aoDesmontar } from './runtime-web.js';

// Teste semRastreio e lote
```

- [ ] **Step 2: Executar o teste para verificar falha**

Run: `node stdlib/web/test_vdom.js`
Expected: FAIL com "semRastreio is not defined" ou "lote is not defined"

- [ ] **Step 3: Implementar `semRastreio`, `lote`, `aoMontar` e `aoDesmontar`**

No arquivo `stdlib/web/runtime-web.js`:

```javascript
/**
 * Executa uma função sem registrar dependências no efeito ativo.
 */
export function semRastreio(fn) {
  const anterior = efeitoAtivo;
  efeitoAtivo = null;
  try {
    return fn();
  } finally {
    efeitoAtivo = anterior;
  }
}
export const untrack = semRastreio;

/**
 * Agrupa Múltiplas mutações de sinais em um único lote síncrono.
 */
export function lote(fn) {
  const agendamentoAnterior = agendamentoLote;
  agendamentoLote = true;
  try {
    fn();
  } finally {
    agendamentoLote = agendamentoAnterior;
    if (!agendamentoLote && filaEfeitos.size > 0) {
      const loteEfeitos = Array.from(filaEfeitos);
      filaEfeitos.clear();
      loteEfeitos.forEach((ef) => ef.run());
    }
  }
}
export const batch = lote;

/**
 * Registra uma função de retorno para execução ao montar o componente.
 */
export function aoMontar(fn) {
  efeito(() => {
    fn();
  });
}

/**
 * Registra uma função de retorno para execução ao desmontar o componente.
 */
export function aoDesmontar(fn) {
  efeito(() => {
    return () => fn();
  });
}
```

- [ ] **Step 4: Executar o teste para verificar aprovação**

Run: `node stdlib/web/test_vdom.js`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add stdlib/web/runtime-web.js stdlib/web/test_vdom.js
git commit -m "feat(web): adiciona primitivas reativas semRastreio, lote, aoMontar e aoDesmontar"
```

---

### Task 2: Suporte a Elementos SVG (`createElementNS`) e `<Fragmento>`

**Files:**
- Modify: `stdlib/web/runtime-web.js`
- Test: `stdlib/web/test_vdom.js`

**Interfaces:**
- Produces: `<Fragmento>` / `<Fragment>`, suporte a `createElementNS` no VDOM para tags SVG.

- [ ] **Step 1: Escrever teste com falha para renderização de SVG e Fragmento**

No arquivo `stdlib/web/test_vdom.js`:

```javascript
import { Fragmento, h, montar } from './runtime-web.js';

// Teste de criação de elemento <svg> e <Fragmento>
```

- [ ] **Step 2: Executar o teste para verificar se falha ou se SVG é criado sem namespace**

Run: `node stdlib/web/test_vdom.js`
Expected: FAIL ou falta de suporte a `createElementNS`

- [ ] **Step 3: Implementar suporte a SVG e `Fragmento` em `runtime-web.js`**

No arquivo `stdlib/web/runtime-web.js`:

```javascript
const TAGS_SVG = new Set([
  "svg", "path", "rect", "circle", "g", "polygon", "line",
  "polyline", "ellipse", "use", "text", "image", "defs", "clipPath"
]);

export function Fragmento(props) {
  return props.children || null;
}
export const Fragment = Fragmento;

// Em criarNo:
// Se TAGS_SVG.has(tagLower), usar document.createElementNS("http://www.w3.org/2000/svg", vno.tag)
```

- [ ] **Step 4: Executar os testes para verificar se passam**

Run: `node stdlib/web/test_vdom.js`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add stdlib/web/runtime-web.js stdlib/web/test_vdom.js
git commit -m "feat(web): adiciona suporte a elementos SVG com namespace e componente Fragmento"
```

---

### Task 3: Gestão Dinâmica do Head (`usarCabecalho`) e DevTools (`__HARPIA_DEVTOOLS__`)

**Files:**
- Modify: `stdlib/web/runtime-web.js`
- Test: `stdlib/web/test_vdom.js`

**Interfaces:**
- Produces: `usarCabecalho(opcoes)`, `window.__HARPIA_DEVTOOLS__`

- [ ] **Step 1: Escrever teste com falha para `usarCabecalho` e DevTools**

No arquivo `stdlib/web/test_vdom.js`:

```javascript
import { usarCabecalho } from './runtime-web.js';

// Testar atualização de título e metadados
```

- [ ] **Step 2: Executar teste**

Run: `node stdlib/web/test_vdom.js`
Expected: FAIL com "usarCabecalho is not defined"

- [ ] **Step 3: Implementar `usarCabecalho` e `window.__HARPIA_DEVTOOLS__`**

No arquivo `stdlib/web/runtime-web.js`:

```javascript
export function usarCabecalho(opcoes = {}) {
  efeito(() => {
    if (typeof document === "undefined") return;
    if (opcoes.titulo) {
      document.title = opcoes.titulo;
    }
    if (opcoes.meta && Array.isArray(opcoes.meta)) {
      opcoes.meta.forEach(({ nome, conteudo, property }) => {
        let el = document.querySelector(`meta[${nome ? 'name="' + nome + '"' : 'property="' + property + '"'}]`);
        if (!el) {
          el = document.createElement("meta");
          if (nome) el.setAttribute("name", nome);
          if (property) el.setAttribute("property", property);
          document.head.appendChild(el);
        }
        el.setAttribute("content", conteudo || "");
      });
    }
  });
}

if (typeof window !== "undefined") {
  window.__HARPIA_DEVTOOLS__ = {
    versao: "1.0.0",
    sinaisAtivos: 0,
    efEfeitosAtivos: 0,
    obterStatus() {
      return { sinais: this.sinaisAtivos, efeitos: this.efEfeitosAtivos };
    }
  };
}
```

- [ ] **Step 4: Executar testes**

Run: `node stdlib/web/test_vdom.js`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add stdlib/web/runtime-web.js stdlib/web/test_vdom.js
git commit -m "feat(web): adiciona hook usarCabecalho e registro global DevTools"
```

---

### Task 4: Sincronização de Runtime Go (`cmd/runtime_web_code.go`), SSR Backend (`hrp/vdom_backend.go`) e Testes de Integração

**Files:**
- Modify: `cmd/runtime_web_code.go`, `hrp/vdom_backend.go`
- Test: `tests/ssr_test.go`

**Interfaces:**
- Consumes: `stdlib/web/runtime-web.js`
- Produces: Runtime web Go 100% atualizado, renderização SSR de SVG em Go.

- [ ] **Step 1: Copiar conteúdo atualizado de `stdlib/web/runtime-web.js` para `cmd/runtime_web_code.go`**

Atualizar o valor de `FallbackRuntimeWebJS` em `cmd/runtime_web_code.go` para ser idêntico a `stdlib/web/runtime-web.js`.

- [ ] **Step 2: Atualizar `hrp/vdom_backend.go` para suporte a SVG no SSR**

No arquivo `hrp/vdom_backend.go`:

Adicionar tags SVG na sanitização de tags Void e preservar atributos SVG em caixa mista.

- [ ] **Step 3: Executar a suíte de testes do Go**

Run: `go test ./...`
Expected: PASS em todos os pacotes.

- [ ] **Step 4: Commit**

```bash
git add cmd/runtime_web_code.go hrp/vdom_backend.go tests/ssr_test.go
git commit -m "chore(web): sincroniza runtime web JS no compilador Go e aprimora SSR backend"
```
