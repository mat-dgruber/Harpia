import { sinal, efeito, derivado, armazem, h, montar, sinalDebounce, criarRef, usarTransicao, usarAtraso, semRastreio, untrack, lote, batch, aoMontar, aoDesmontar } from './runtime-web.js';

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
          if (child) child.parentNode = this;
          this.childNodes.push(child);
        },
        insertBefore(newChild, refChild) {
          if (newChild) newChild.parentNode = this;
          const idx = this.childNodes.indexOf(refChild);
          if (idx > -1) {
            this.childNodes.splice(idx, 0, newChild);
          } else {
            this.childNodes.push(newChild);
          }
        },
        removeChild(child) {
          const idx = this.childNodes.indexOf(child);
          if (idx > -1) {
            this.childNodes.splice(idx, 1);
          }
        },
        replaceChild(newChild, oldChild) {
          if (newChild) newChild.parentNode = this;
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

  console.log("\n=== 5. TESTANDO MAPEAMENTO DE EVENTOS ADAPTATIVOS COM REGEX ===");
  const testContainer = document.createElement("div");
  let eventoDisparado = null;

  function AppEventos() {
    return h("div", {},
      h("button", {
        aoClicar: () => { eventoDisparado = 'click'; },
        AoMudar: () => { eventoDisparado = 'change'; }, // case-insensitive prefixo
        aoerror: () => { eventoDisparado = 'error'; } // case-insensitive completo
      }, "Eventos")
    );
  }

  montar(AppEventos, testContainer);
  await Promise.resolve();

  const botao = testContainer.childNodes[0].childNodes[0];

  // Simula 'click' (aoClicar)
  if (botao.listeners['click']) {
    botao.listeners['click']();
    console.log("Evento 'aoClicar' mapeou para 'click':", eventoDisparado === 'click' ? "OK" : "FALHA");
  } else {
    console.log("Erro: listener 'click' não encontrado!");
  }

  // Simula 'change' (AoMudar)
  if (botao.listeners['change']) {
    botao.listeners['change']();
    console.log("Evento 'AoMudar' mapeou para 'change':", eventoDisparado === 'change' ? "OK" : "FALHA");
  } else {
    console.log("Erro: listener 'change' não encontrado!");
  }

  // Simula 'error' (aoerror)
  if (botao.listeners['error']) {
    botao.listeners['error']();
    console.log("Evento 'aoerror' mapeou para 'error':", eventoDisparado === 'error' ? "OK" : "FALHA");
  } else {
    console.log("Erro: listener 'error' não encontrado!");
  }

  console.log("\n=== 6. TESTANDO REFS E HOOKS AVANÇADOS (criarRef, usarTransicao) ===");
  const meuInputRef = criarRef();
  const testRefContainer = document.createElement("div");

  function AppRef() {
    return h("input", { ref: meuInputRef, placeholder: "Digite aqui" });
  }

  montar(AppRef, testRefContainer);
  await Promise.resolve();

  console.log("Ref do elemento input atribuída com sucesso:", meuInputRef.current !== null ? "OK" : "FALHA");

  const [pendente, iniciarTransicao] = usarTransicao();
  let transicaoExecutada = false;

  iniciarTransicao(() => {
    transicaoExecutada = true;
  });

  await Promise.resolve();
  await new Promise(r => setTimeout(r, 10));

  console.log("Transição executada assincronamente:", transicaoExecutada ? "OK" : "FALHA");

  console.log("\n=== 7. TESTANDO FUNÇÃO DE LIMPEZA (CLEANUP) EM EFEITO ===");
  const [val, setVal] = sinal(1);
  let limpezaExecutada = false;

  const pararEfeito = efeito(() => {
    const v = val();
    return () => {
      limpezaExecutada = true;
    };
  });

  setVal(2);
  await Promise.resolve();
  console.log("Limpeza disparada ao re-executar efeito:", limpezaExecutada ? "OK" : "FALHA");

  limpezaExecutada = false;
  pararEfeito();
  console.log("Limpeza disparada ao cancelar efeito manualmente:", limpezaExecutada ? "OK" : "FALHA");

  console.log("\n=== 8. TESTANDO PRIMITIVAS REATIVAS AVANÇADAS (semRastreio, lote, aoMontar, aoDesmontar) ===");

  // 8.1 semRastreio / untrack
  const [sRastreavel, setSRastreavel] = sinal(10);
  const [sNaoRastreavel, setSNaoRastreavel] = sinal(100);
  let execucoesSemRastreio = 0;
  let valorNaoRastreado = 0;

  efeito(() => {
    execucoesSemRastreio++;
    const a = sRastreavel();
    valorNaoRastreado = semRastreio(() => sNaoRastreavel());
  });

  console.log("Inicial semRastreio - valor lido:", valorNaoRastreado === 100 ? "OK" : "FALHA");

  setSNaoRastreavel(200);
  await Promise.resolve();
  console.log("Alterar sinal não rastreado não dispara efeito:", execucoesSemRastreio === 1 ? "OK" : "FALHA");

  setSRastreavel(20);
  await Promise.resolve();
  console.log("Alterar sinal rastreado dispara efeito e lê novo valor não rastreado:", execucoesSemRastreio === 2 && valorNaoRastreado === 200 ? "OK" : "FALHA");

  // Testa alias untrack
  let valorUntrack = untrack(() => sNaoRastreavel());
  console.log("Alias untrack funciona:", valorUntrack === 200 ? "OK" : "FALHA");

  // 8.2 lote / batch
  const [aLote, setALote] = sinal(0);
  const [bLote, setBLote] = sinal(0);
  let execucoesLote = 0;

  efeito(() => {
    aLote();
    bLote();
    execucoesLote++;
  });

  execucoesLote = 0; // reset
  lote(() => {
    setALote(1);
    setBLote(2);
  });
  console.log("lote executa de forma agrupada (execuções do efeito = 1):", execucoesLote === 1 ? "OK" : "FALHA");

  execucoesLote = 0;
  batch(() => {
    setALote(10);
    setBLote(20);
  });
  console.log("Alias batch executa de forma agrupada (execuções do efeito = 1):", execucoesLote === 1 ? "OK" : "FALHA");

  // 8.3 aoMontar e aoDesmontar
  let montouContador = 0;
  let desmontouContador = 0;
  const [sMontar, setSMontar] = sinal(1);

  const pararAppMontar = efeito(() => {
    sMontar(); // cria dependência reativa
    aoMontar(() => {
      montouContador++;
      return () => {
        desmontouContador++;
      };
    });
    aoDesmontar(() => {
      desmontouContador++;
    });
  });

  console.log("aoMontar executou no início (1 vez):", montouContador === 1 ? "OK" : "FALHA");

  setSMontar(2); // re-executa efeito
  await Promise.resolve();
  console.log("aoMontar NÃO re-executa em atualizações de sinal (continua 1):", montouContador === 1 ? "OK" : "FALHA");

  pararAppMontar(); // desmonta/dispara cancelamento do efeito
  console.log("aoDesmontar executou ao desmontar efeito (2 callbacks):", desmontouContador === 2 ? "OK" : "FALHA");
}

rodarTestes().catch(console.error);
