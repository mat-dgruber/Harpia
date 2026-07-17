import { sinal, efeito, derivado, armazem, h, montar, sinalDebounce } from './runtime-web.js';

// Mock ultra-simples e corrigido do DOM para rodar os testes em ambiente Node.js / Deno sem browser
if (typeof document === 'undefined') {
  globalThis.document = {
    createTextNode(text) {
      return {
        nodeType: 3,
        textContent: text,
        toString() { return this.textContent; }
      };
    },
    createElement(tag) {
      const el = {
        nodeType: 1,
        tag,
        props: {},
        style: {},
        childNodes: [],
        addEventListener(event, fn) {
          this.listeners = this.listeners || {};
          this.listeners[event] = fn;
        },
        removeEventListener(event, fn) {
          if (this.listeners) delete this.listeners[event];
        },
        setAttribute(name, value) {
          this.props[name] = value;
        },
        removeAttribute(name) {
          delete this.props[name];
        },
        appendChild(child) {
          this.childNodes.push(child);
        },
        removeChild(child) {
          const idx = this.childNodes.indexOf(child);
          if (idx > -1) {
            this.childNodes.splice(idx, 1);
          }
        },
        replaceChild(newChild, oldChild) {
          const idx = this.childNodes.indexOf(oldChild);
          if (idx > -1) {
            this.childNodes[idx] = newChild;
          }
        },
        toString() {
          const propsStr = Object.entries(this.props)
            .map(([k, v]) => ` ${k}="${v}"`).join('');
          const styleStr = Object.keys(this.style).length
            ? ` estilo="${JSON.stringify(this.style)}"`
            : '';
          return `<${this.tag}${propsStr}${styleStr}>${this.childNodes.map(c => c.toString()).join('')}</${this.tag}>`;
        }
      };
      // ponytail: alias simples e direto para childrens
      Object.defineProperty(el, 'children', {
        get() { return this.childNodes; }
      });
      return el;
    }
  };
}

// ponytail: assevera a reatividade e o loteamento de microtarefas de forma assíncrona
async function rodarTestes() {
  console.log("=== 1. TESTANDO SINAIS E REATIVIDADE FINA ===");
  const [count, setCount] = sinal(10);
  const dobro = derivado(() => count() * 2);

  let logsEfeito = [];
  efeito(() => {
    logsEfeito.push(`Efeito observador: count=${count()} | dobro=${dobro()}`);
  });

  console.log("Log inicial:", logsEfeito[0]);

  // Testando o loteamento (batching) síncrono:
  // Mudamos o sinal três vezes seguidas na mesma execução síncrona.
  // Graças ao agendamento em microtasks, o efeito de render só deve rodar UMA única vez!
  setCount(15);
  setCount(18);
  setCount(20);

  console.log("Síncrono imediatamente após setCount():", logsEfeito[1]); // Deve ser undefined (efeito ainda não disparado)

  // Aguarda a resolução das microtarefas
  await Promise.resolve();

  console.log("Após microtask resolvida (Loteamento consolidado):", logsEfeito[1]); // Deve exibir o valor consolidado '20'
  console.log("Total de execuções do efeito (esperava 2 - inicial + consolidado):", logsEfeito.length);

  console.log("\n=== 2. TESTANDO ARMAZÉM GLOBAL ===");
  const carrinho = armazem({ total: 0, itens: [] });
  efeito(() => {
    console.log(`Efeito do armazém: carrinho.total mudou para = ${carrinho.total}`);
  });
  carrinho.total = 150;
  await Promise.resolve();

  console.log("\n=== 3. TESTANDO VIRTUAL DOM E MONTAGEM REATIVA ===");
  const [titulo, setTitulo] = sinal("Olá Harpia");
  const container = document.createElement("div");

  function MeuApp() {
    return h("div", { classe: "app" },
      h("h1", {}, titulo()),
      h("button", { aoClicar: () => setTitulo("Novo Título!") }, "Clique aqui")
    );
  }

  // Montagem inicial reativa do app
  montar(MeuApp, container);
  await Promise.resolve();
  console.log("HTML inicial rendido:\n", container.toString());

  // Mutando sinal deve disparar o renderizador e a reconciliação cirúrgica do DOM
  setTitulo("Título Atualizado Reativamente!");
  await Promise.resolve();
  console.log("\nHTML após setTitulo():\n", container.toString());

  // Simulando evento físico de clique
  console.log("\nSimulando clique no botão...");
  container.childNodes[0].childNodes[1].listeners.click();
  await Promise.resolve();
  console.log("HTML após clique:\n", container.toString());

  console.log("\n=== 4. TESTANDO SINAL COM DEBOUNCE ===");
  const [pesquisa, setPesquisa] = sinalDebounce("abc", 50);
  let totalDebounce = 0;
  efeito(() => {
    console.log(`Efeito debounce: pesquisa mudou para = ${pesquisa()}`);
    totalDebounce++;
  });

  setPesquisa("def");
  setPesquisa("ghi");
  setPesquisa("jkl");

  console.log("Imediato (deve ser 'abc'):", pesquisa());

  await new Promise(r => setTimeout(r, 100));
  console.log("Após 100ms (deve ser 'jkl'):", pesquisa());
  console.log("Total execuções efeito (esperava 2 - inicial + jkl):", totalDebounce);
}

rodarTestes().catch(console.error);
