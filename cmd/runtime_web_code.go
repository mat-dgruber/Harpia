//go:build !js && !wasm

package cmd

// FallbackRuntimeWebJS contém o runtime JS do Harpia embutido no compilador.
const FallbackRuntimeWebJS = `/**
 * Harpia Web Runtime - VDOM & Reatividade Fina
 * ponytail: Solução ultraleve (~2.5KB) sem dependências externas.
 * Fornece Virtual DOM, reconciliação cirúrgica e reatividade fina baseada em Sinais.
 */

let efeitoAtivo = null;
let profundidadeLote = 0;

// DevTools metrics tracking
const devToolsMetrics = {
  sinaisCriados: 0,
  efeitosCriados: 0,
  renderizacoesVDOM: 0,
};

const registroSinais = new Set();
const registroEfeitos = new Set();

// ponytail: fila global e flag para agrupar e lotear execuções de efeitos síncronos na mesma microtask.
// Isso impede múltiplas reconciliações de VDOM desnecessárias e repaints repetitivos.
const filaEfeitos = new Set();
let agendamentoLote = false;

function executarFilaEfeitos() {
  while (filaEfeitos.size > 0) {
    const loteEfeitos = Array.from(filaEfeitos);
    filaEfeitos.clear();
    loteEfeitos.forEach((ef) => ef.run());
  }
}

function agendarEfeito(ef) {
  filaEfeitos.add(ef);
  if (profundidadeLote > 0) return;
  if (!agendamentoLote) {
    agendamentoLote = true;
    // queueMicrotask é um recurso nativo da plataforma web moderna
    queueMicrotask(() => {
      agendamentoLote = false;
      executarFilaEfeitos();
    });
  }
}

// ============================================================================
// 1. SISTEMA DE SINAIS E REATIVIDADE FINA
// ============================================================================

/**
 * Cria um sinal reativo contendo um valor.
 * @param {*} valor - Valor inicial
 * @returns {[Function, Function]} [ler, definir]
 */
export function sinal(valor) {
  devToolsMetrics.sinaisCriados++;
  const assinantes = new Set();

  const ler = () => {
    if (efeitoAtivo) {
      assinantes.add(efeitoAtivo);
      efeitoAtivo.deps.add(assinantes);
    }
    return valor;
  };

  const definir = (novoValor) => {
    if (valor !== novoValor) {
      valor = novoValor;
      // Clona para evitar recursão infinita se assinantes mutarem o sinal
      const paraDisparar = Array.from(assinantes);
      paraDisparar.forEach((sub) => agendarEfeito(sub));
    }
  };

  ler.set = definir; // ponytail: expõe atualizador direto no getter para binding bidirecional
  registroSinais.add(ler);
  return [ler, definir];
}

/**
 * Cria um sinal cujo atualizador possui atraso (debounce) síncrono.
 * @param {*} valorInicial - Valor inicial do sinal
 * @param {number} tempoEmMs - Tempo em milissegundos para atrasar o disparo
 */
export function sinalDebounce(valorInicial, tempoEmMs) {
  const [ler, definir] = sinal(valorInicial);
  let timeoutId = null;

  const definirDebounce = (novoValor) => {
    if (timeoutId) {
      clearTimeout(timeoutId);
    }
    timeoutId = setTimeout(() => {
      definir(novoValor);
    }, tempoEmMs);
  };

  ler.set = definirDebounce; // ponytail: expõe atualizador para binding reativo
  return [ler, definirDebounce];
}

/**
 * Cria um objeto de referência mutável e estável entre renderizações (ref).
 * @param {*} valorInicial - Valor inicial de ref.current
 * @returns {{ current: * }}
 */
export function criarRef(valorInicial = null) {
  return { current: valorInicial };
}

/**
 * Hook para temporizador de intervalo com limpeza automática.
 */
export function usarIntervalo(callback, tempoEmMs) {
  efeito(() => {
    if (tempoEmMs === null || tempoEmMs === undefined) return;
    const id = setInterval(callback, tempoEmMs);
    return () => clearInterval(id);
  });
}

/**
 * Hook para temporizador com atraso e limpeza automática.
 */
export function usarAtraso(callback, tempoEmMs) {
  efeito(() => {
    if (tempoEmMs === null || tempoEmMs === undefined) return;
    const id = setTimeout(callback, tempoEmMs);
    return () => clearTimeout(id);
  });
}

/**
 * Hook para transições assíncronas não-bloqueantes de baixa prioridade.
 * @returns {[Function, Function]} [estaPendente, iniciarTransicao]
 */
export function usarTransicao() {
  const [estaPendente, setPendente] = sinal(false);

  function iniciarTransicao(callback) {
    setPendente(true);
    queueMicrotask(() => {
      try {
        callback();
      } finally {
        queueMicrotask(() => setPendente(false));
      }
    });
  }

  return [estaPendente, iniciarTransicao];
}

/**
 * Executa uma função sem registrar dependências reativas no efeito ativo.
 * @param {Function} fn
 * @returns {*} Resultado de fn()
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
 * Agrupa múltiplas alterações de sinais para disparar efeitos apenas ao final do bloco.
 * @param {Function} fn
 * @returns {*} Resultado de fn()
 */
export function lote(fn) {
  profundidadeLote++;
  try {
    return fn();
  } finally {
    profundidadeLote--;
    if (profundidadeLote === 0) {
      executarFilaEfeitos();
    }
  }
}
export const batch = lote;

/**
 * Registra uma função de inicialização na montagem do componente/efeito.
 * Se fn retornar uma função de limpeza, ela será registrada para a desmontagem.
 * @param {Function} fn
 */
export function aoMontar(fn) {
  if (typeof fn === "function") {
    if (!efeitoAtivo || !efeitoAtivo.iniciado) {
      const limpa = fn();
      if (typeof limpa === "function") {
        aoDesmontar(limpa);
      }
    }
  }
}

/**
 * Registra uma função de limpeza (cleanup) para ser executada no desmonte/descarte do efeito ou componente.
 * @param {Function} fn
 */
export function aoDesmontar(fn) {
  if (typeof fn === "function" && efeitoAtivo) {
    if (!efeitoAtivo.iniciado) {
      if (!efeitoAtivo.desmontarCleanups) {
        efeitoAtivo.desmontarCleanups = new Set();
      }
      efeitoAtivo.desmontarCleanups.add(fn);
    }
  }
}

/**
 * Cria um efeito colateral que roda imediatamente e se reinscreve nas mudanças de sinais lidos dentro dele.
 * @param {Function} funcao
 * @returns {Function} Função de limpeza (cleanup / desmontagem)
 */
export function efeito(funcao) {
  devToolsMetrics.efeitosCriados++;
  const ef = {
    deps: new Set(),
    cleanup: null,
    cleanups: new Set(),
    desmontarCleanups: new Set(),
    iniciado: false,
    run() {
      if (typeof ef.cleanup === "function") {
        try {
          ef.cleanup();
        } catch (e) {
          console.error("Erro na limpeza do efeito:", e);
        }
        ef.cleanup = null;
      }
      for (const c of ef.cleanups) {
        try {
          c();
        } catch (e) {
          console.error("Erro na limpeza do efeito:", e);
        }
      }
      ef.cleanups.clear();

      limparDeps(ef);
      const anterior = efeitoAtivo;
      efeitoAtivo = ef;
      try {
        const ret = funcao();
        if (typeof ret === "function") {
          ef.cleanup = ret;
        }
      } finally {
        efeitoAtivo = anterior;
        ef.iniciado = true;
      }
    },
    desmontar() {
      registroEfeitos.delete(ef);
      if (typeof ef.cleanup === "function") {
        try {
          ef.cleanup();
        } catch (e) {
          console.error("Erro na limpeza do efeito:", e);
        }
        ef.cleanup = null;
      }
      for (const c of ef.cleanups) {
        try {
          c();
        } catch (e) {
          console.error("Erro na limpeza do efeito:", e);
        }
      }
      ef.cleanups.clear();

      for (const c of ef.desmontarCleanups) {
        try {
          c();
        } catch (e) {
          console.error("Erro ao desmontar efeito:", e);
        }
      }
      ef.desmontarCleanups.clear();

      limparDeps(ef);
    },
  };
  registroEfeitos.add(ef);
  ef.run();
  return () => ef.desmontar();
}

function limparDeps(ef) {
  for (const assinantes of ef.deps) {
    assinantes.delete(ef);
  }
  ef.deps.clear();
}

/**
 * Cria um valor reativo computado e memoizado a partir de outros sinais.
 * @param {Function} funcao
 * @returns {Function} Função de leitura reativa
 */
export function derivado(funcao) {
  const [ler, definir] = sinal();
  efeito(() => {
    definir(funcao());
  });
  return ler;
}

/**
 * Gerenciador de Estado Global reativo por chave-valor.
 * @param {Object} objeto - Estado inicial
 * @returns {Proxy} Objeto reativo global
 */
export function armazem(objeto) {
  const sinais = {};
  for (const chave in objeto) {
    if (objeto.hasOwnProperty(chave)) {
      sinais[chave] = sinal(objeto[chave]);
    }
  }
  return new Proxy(objeto, {
    get(target, propriedade) {
      if (sinais[propriedade]) {
        return sinais[propriedade][0]();
      }
      return target[propriedade];
    },
    set(target, propriedade, valor) {
      if (sinais[propriedade]) {
        sinais[propriedade][1](valor);
        target[propriedade] = valor;
        return true;
      }
      sinais[propriedade] = sinal(valor);
      target[propriedade] = valor;
      return true;
    },
  });
}

// ============================================================================
// 2. VIRTUAL DOM E RECONCILIAÇÃO (DIFF / PATCH)
// ============================================================================

/**
 * Cria um nó Virtual DOM (VNode).
 * @param {string|Function} tag - Nome da tag HTML ou componente funcional
 * @param {Object} props - Propriedades e atributos do elemento
 * @param {...*} filhos - Filhos do elemento
 */
export function h(tag, props, ...filhos) {
  props = props || {};
  const children = filhos
    .flat(Infinity)
    .map((f) => (f === undefined || f === false ? null : f));
  if (children.length > 0 && props.children === undefined) {
    props.children = children.length === 1 ? children[0] : children;
  }
  return { tag, props, children };
}

/**
 * Componente Fragmento para agrupar múltiplos nós sem criar elemento contêiner extra no DOM.
 */
export function Fragmento(props) {
  if (!props) return null;
  const c = props.children;
  if (c === undefined || c === null || c === false) return null;
  return Array.isArray(c) ? c : [c];
}
export const Fragment = Fragmento;

const TAGS_SVG = new Set([
  "svg",
  "path",
  "rect",
  "circle",
  "g",
  "polygon",
  "line",
  "polyline",
  "ellipse",
  "use",
  "text",
  "image",
  "defs",
  "clipPath",
  "clippath",
  "linearGradient",
  "lineargradient",
  "radialGradient",
  "radialgradient",
  "stop",
  "symbol",
  "marker",
  "mask",
  "pattern",
  "tspan",
  "filter",
  "feDropShadow",
  "fedropshadow",
  "feGaussianBlur",
  "fegaussianblur",
  "foreignObject",
  "foreignobject",
  "style",
  "animate",
  "animateTransform",
  "animatetransform",
]);

function eTagSVG(tag) {
  if (typeof tag !== "string") return false;
  return TAGS_SVG.has(tag) || TAGS_SVG.has(tag.toLowerCase());
}

/**
 * Instancia fisicamente um nó VDOM no DOM real do navegador.
 * @param {Object|string|number|Array} vno
 * @returns {Node} Elemento do DOM físico
 */
export function criarNo(vno) {
  if (vno === null || vno === undefined || vno === false) {
    return document.createTextNode("");
  }

  if (Array.isArray(vno)) {
    const frag = document.createDocumentFragment
      ? document.createDocumentFragment()
      : document.createElement("fragment");
    vno.forEach((filho) => {
      if (filho !== null && filho !== undefined && filho !== false) {
        frag.appendChild(criarNo(filho));
      }
    });
    return frag;
  }

  if (typeof vno === "function") {
    try {
      return criarNo(vno());
    } catch (e) {
      console.error("[Harpia VDOM] Erro ao executar função componente:", e);
      const errEl = document.createElement("div");
      errEl.style.cssText =
        "color:#EF4444; background:#1E1E2E; padding:8px; border-radius:4px; font-family:monospace; font-size:12px;";
      errEl.textContent = "[Erro no Componente: " + (e.message || e) + "]";
      return errEl;
    }
  }

  if (typeof vno === "string" || typeof vno === "number") {
    return document.createTextNode(vno);
  }

  if (!vno || typeof vno !== "object") {
    return document.createTextNode(String(vno || ""));
  }

  if (
    typeof vno.tag === "string" &&
    (vno.tag.toLowerCase() === "fragment" ||
      vno.tag.toLowerCase() === "fragmento" ||
      vno.tag === "")
  ) {
    return criarNo(vno.children || []);
  }

  if (typeof vno.tag === "function") {
    vno.props = vno.props || {};
    if (vno.props._ligar && !vno.props.ligar) {
      vno.props.ligar = vno.props._ligar;
    }
    try {
      const vnodeRendido = vno.tag(vno.props);
      vno._componenteInstancia = vnodeRendido;
      const no = criarNo(vnodeRendido);
      vno.el = no;
      return no;
    } catch (e) {
      console.error(
        "[Harpia VDOM] Erro ao renderizar componente <" + (vno.tag.name || "Anônimo") + ">:",
        e,
      );
      const errEl = document.createElement("div");
      errEl.style.cssText =
        "color:#EF4444; background:#1E1E2E; padding:8px; border-radius:4px; font-family:monospace; font-size:12px;";
      errEl.textContent = "<" + (vno.tag.name || "Componente") + "> Erro: " + (e.message || e);
      return errEl;
    }
  }

  const el = eTagSVG(vno.tag)
    ? (document.createElementNS
        ? document.createElementNS("http://www.w3.org/2000/svg", vno.tag)
        : document.createElement(vno.tag))
    : document.createElement(vno.tag || "div");

  vno.el = el;

  atualizarAtributos(el, {}, vno.props);

  if (vno.props && vno.props.ref) {
    const valRef = vno.props.ref;
    if (typeof valRef === "function") {
      valRef(el);
    } else if (typeof valRef === "object" && valRef !== null) {
      valRef.current = el;
    }
  }

  (vno.children || []).forEach((filho) => {
    if (filho !== null && filho !== undefined && filho !== false) {
      el.appendChild(criarNo(filho));
    }
  });

  return el;
}

function atualizarAtributos(el, velhosProps, novosProps) {
  velhosProps = velhosProps || {};
  novosProps = novosProps || {};

  // Remover velhos props
  for (const chave in velhosProps) {
    if (!chave || typeof chave !== "string") continue;
    const nomeLimpo = chave.trim();
    if (!nomeLimpo || nomeLimpo === "children" || nomeLimpo === "key" || nomeLimpo === "ref") continue;
    if (!(chave in novosProps)) {
      if (/^ao/i.test(nomeLimpo)) {
        const evento = mapearEvento(nomeLimpo);
        el.removeEventListener(evento, velhosProps[chave]);
      } else {
        const attrName = mapearAtributo(nomeLimpo);
        if (attrName) el.removeAttribute(attrName);
      }
    }
  }

  // Adicionar ou atualizar novos props
  for (const chave in novosProps) {
    if (!chave || typeof chave !== "string") continue;
    const nomeLimpo = chave.trim();
    if (!nomeLimpo || nomeLimpo === "children" || nomeLimpo === "key" || nomeLimpo === "ref") continue;
    const valor = novosProps[chave];
    const velhoValor = velhosProps[chave];

    if (valor !== velhoValor) {
      if (nomeLimpo === "_ligar" || nomeLimpo === "ligar") {
        const lerSinal = valor;
        efeito(() => {
          const val = typeof lerSinal === "function" ? lerSinal() : lerSinal;
          if (el.type === "checkbox") {
            el.checked = !!val;
          } else {
            el.value = val === null || val === undefined ? "" : val;
            if (el.tagName === "SELECT") {
              queueMicrotask(() => {
                el.value = val === null || val === undefined ? "" : val;
              });
            }
          }
        });
        const eventoTipos =
          el.type === "checkbox" || el.tagName === "SELECT"
            ? ["change", "input"]
            : ["input", "change"];
        const ouvinte = (e) => {
          const setter = lerSinal ? lerSinal.set || lerSinal[1] : null;
          if (setter) {
            setter(el.type === "checkbox" ? e.target.checked : e.target.value);
          }
        };
        eventoTipos.forEach((ev) => el.addEventListener(ev, ouvinte));
        continue;
      }

      if (nomeLimpo === "innerHTML") {
        el.innerHTML = valor;
        continue;
      }

      if (nomeLimpo === "estilo" || nomeLimpo === "style") {
        if (typeof valor === "object" && valor !== null) {
          el.style.cssText = "";
          Object.assign(el.style, valor);
        } else {
          el.setAttribute("style", valor || "");
        }
        continue;
      }

      if (/^ao/i.test(nomeLimpo)) {
        const evento = mapearEvento(nomeLimpo);
        if (el["_ao_" + evento])
          el.removeEventListener(evento, el["_ao_" + evento]);
        const listener = (e) => {
          if (typeof valor === "function") valor(e);
        };
        el["_ao_" + evento] = listener;
        el.addEventListener(evento, listener);
      } else {
        const attrName = mapearAtributo(nomeLimpo);
        if (!attrName) continue;
        if (valor === true) {
          el.setAttribute(attrName, "");
        } else if (valor === false || valor === null || valor === undefined) {
          el.removeAttribute(attrName);
        } else {
          el.setAttribute(attrName, valor);
        }
      }
    }
  }
}

function mapearEvento(nome) {
  // Remove o prefixo 'ao' case-insensitive (ex: ao, Ao, AO, aO)
  const semAo = (nome || "").trim().replace(/^ao/i, "").toLowerCase();

  // Regras de mapeamento adaptativo inteligente por Regex/Substrings conhecidas
  if (/^(clique|clicar|click)$/.test(semAo)) return "click";
  if (/^(mudar|alterar|change)$/.test(semAo)) return "change";
  if (/^(enviar|submeter|submit)$/.test(semAo)) return "submit";
  if (/^(focar|foco|focus)$/.test(semAo)) return "focus";
  if (/^(desfocar|desfoco|blur)$/.test(semAo)) return "blur";
  if (/^(teclar|tecla|keydown)$/.test(semAo)) return "keydown";
  if (/^(soltartecla|keyup)$/.test(semAo)) return "keyup";
  if (/^(entrarmouse|mouseenter)$/.test(semAo)) return "mouseenter";
  if (/^(sairmouse|mouseleave)$/.test(semAo)) return "mouseleave";
  if (/^(digitar|inserir|input)$/.test(semAo)) return "input";
  if (/^(carregar|load)$/.test(semAo)) return "load";
  if (/^(duploclique|dblclick)$/.test(semAo)) return "dblclick";
  if (/^(passarmouse|mouseover)$/.test(semAo)) return "mouseover";
  if (/^(erro|error)$/.test(semAo)) return "error";

  return semAo;
}

function mapearAtributo(nome) {
  if (!nome || typeof nome !== "string") return "";
  nome = nome.trim();
  if (!nome || /[\s\n\r<>"'/=]/.test(nome)) return "";
  const mapeamento = {
    classe: "class",
    identificador: "id",
    estilo: "style",
    corDeFundo: "bgcolor",
    largura: "width",
    altura: "height",
    tipo: "type",
    valor: "value",
    desabilitado: "disabled",
    obrigatorio: "required",
    nome: "name",
    marcado: "checked",
    espacoReservado: "placeholder",
    espaco_reservado: "placeholder",
    chave: "key",
    alvo: "target",
    caminho: "src",
    fonte: "src",
    link: "href",
    destino: "href",
    endereco: "href",
    autoFoco: "autofocus",
    auto_foco: "autofocus",
    autoCompletar: "autocomplete",
    auto_completar: "autocomplete",
    leituraApenas: "readonly",
    somente_leitura: "readonly",
  };
  return mapeamento[nome] || nome;
}

function contarNosFisicos(vno) {
  if (!vno) return 0;
  if (Array.isArray(vno)) {
    let sum = 0;
    for (const item of vno) sum += contarNosFisicos(item);
    return sum;
  }
  if (typeof vno === "object" && typeof vno.tag === "function" && vno._componenteInstancia) {
    return contarNosFisicos(vno._componenteInstancia);
  }
  return 1;
}

function obterNoFisico(vno, pai) {
  if (!vno) return null;
  if (Array.isArray(vno)) {
    for (const item of vno) {
      const el = obterNoFisico(item, pai);
      if (el) return el;
    }
    return null;
  }
  let el = vno.el;
  let curr = vno;
  while (!el && curr && typeof curr.tag === "function" && curr._componenteInstancia) {
    curr = curr._componenteInstancia;
    if (Array.isArray(curr)) {
      return obterNoFisico(curr, pai);
    }
    el = curr ? curr.el : null;
  }
  if (el && pai && el.parentNode !== pai) {
    return null;
  }
  return el || null;
}

/**
 * Reconcilia de forma cirúrgica um nó DOM real com as mudanças entre o VNode antigo e o novo.
 * @param {Node} pai - Elemento pai físico no DOM
 * @param {Object} velhoVNo - VNode anterior
 * @param {Object} novoVNo - Novo VNode
 * @param {number} index - Posição do nó físico no pai
 */
export function reconciliar(pai, velhoVNo, novoVNo, index = 0) {
  devToolsMetrics.renderizacoesVDOM++;
  if (!pai) return;

  if (typeof velhoVNo === "function") velhoVNo = velhoVNo();
  if (typeof novoVNo === "function") novoVNo = novoVNo();

  if (velhoVNo === novoVNo && typeof novoVNo !== "object") return;

  const noFisico =
    obterNoFisico(velhoVNo, pai) ||
    (pai.childNodes ? pai.childNodes[index] : null);

  if (velhoVNo === null || velhoVNo === undefined || velhoVNo === false) {
    if (novoVNo !== null && novoVNo !== undefined && novoVNo !== false) {
      const novoEl = criarNo(novoVNo);
      if (typeof pai.insertBefore === "function" && pai.childNodes && pai.childNodes[index]) {
        pai.insertBefore(novoEl, pai.childNodes[index]);
      } else if (typeof pai.appendChild === "function") {
        pai.appendChild(novoEl);
      }
    }
    return;
  }

  if (novoVNo === null || novoVNo === undefined || novoVNo === false) {
    if (Array.isArray(velhoVNo)) {
      for (let i = velhoVNo.length - 1; i >= 0; i--) {
        reconciliar(pai, velhoVNo[i], null, index + i);
      }
      return;
    }
    if (noFisico && (noFisico.parentNode === pai || !noFisico.parentNode) && typeof pai.removeChild === "function") {
      try { pai.removeChild(noFisico); } catch (_) {}
    }
    return;
  }

  if (mudou(velhoVNo, novoVNo)) {
    const novoEl = criarNo(novoVNo);
    const alvoPai = (noFisico && noFisico.parentNode) || pai;
    if (alvoPai && typeof alvoPai.replaceChild === "function" && noFisico) {
      alvoPai.replaceChild(novoEl, noFisico);
    } else if (typeof pai.insertBefore === "function" && pai.childNodes && pai.childNodes[index]) {
      pai.insertBefore(novoEl, pai.childNodes[index]);
    } else if (typeof pai.appendChild === "function") {
      pai.appendChild(novoEl);
    }
    return;
  }

  if (typeof novoVNo === "object" && typeof novoVNo.tag === "function") {
    novoVNo.props = novoVNo.props || {};
    if (novoVNo.props._ligar && !novoVNo.props.ligar) {
      novoVNo.props.ligar = novoVNo.props._ligar;
    }
    const novoRendido = novoVNo.tag(novoVNo.props);
    novoVNo._componenteInstancia = novoRendido;
    const velhoRendido =
      velhoVNo && velhoVNo._componenteInstancia !== undefined
        ? velhoVNo._componenteInstancia
        : velhoVNo;
    reconciliar(pai, velhoRendido, novoRendido, index);
    novoVNo.el = obterNoFisico(novoRendido, pai);
    return;
  }

  if (Array.isArray(novoVNo) || Array.isArray(velhoVNo)) {
    const velhosFilhos = Array.isArray(velhoVNo) ? velhoVNo : (velhoVNo ? [velhoVNo] : []);
    const novosFilhos = Array.isArray(novoVNo) ? novoVNo : (novoVNo ? [novoVNo] : []);
    if (velhosFilhos.length > novosFilhos.length) {
      let subIdx = index;
      for (let i = 0; i < novosFilhos.length; i++) {
        subIdx += contarNosFisicos(velhosFilhos[i]);
      }
      for (let i = velhosFilhos.length - 1; i >= novosFilhos.length; i--) {
        reconciliar(pai, velhosFilhos[i], null, subIdx);
      }
    }
    let domOffset = index;
    for (let i = 0; i < novosFilhos.length; i++) {
      reconciliar(pai, velhosFilhos[i], novosFilhos[i], domOffset);
      domOffset += contarNosFisicos(novosFilhos[i]);
    }
    return;
  }

  if (typeof novoVNo === "object") {
    if (noFisico && noFisico.nodeType === 1) {
      atualizarAtributos(
        noFisico,
        (velhoVNo && velhoVNo.props) || {},
        novoVNo.props || {},
      );
      novoVNo.el = noFisico;
    }

    const velhosFilhos = (velhoVNo && velhoVNo.children) || [];
    const novosFilhos = (novoVNo && novoVNo.children) || [];

    const temChaves = novosFilhos.some(
      (f) =>
        f &&
        f.props &&
        (f.props.chave !== undefined || f.props.key !== undefined),
    );

    if (temChaves && noFisico) {
      const mapaVelhos = new Map();
      velhosFilhos.forEach((vno, idx) => {
        let k = vno && vno.props ? (vno.props.chave ?? vno.props.key) : idx;
        if (k === undefined) k = idx;
        const noChild = noFisico.childNodes ? noFisico.childNodes[idx] : null;
        if (k !== undefined && noChild) {
          mapaVelhos.set(k, { vno, el: noChild });
        }
      });

      novosFilhos.forEach((novoChildVNo, i) => {
        let k =
          novoChildVNo && novoChildVNo.props
            ? (novoChildVNo.props.chave ?? novoChildVNo.props.key)
            : i;
        if (k === undefined) k = i;
        const correspondencia = mapaVelhos.get(k);

        if (correspondencia) {
          if (noFisico.childNodes && noFisico.childNodes[i] !== correspondencia.el && typeof noFisico.insertBefore === "function") {
            noFisico.insertBefore(correspondencia.el, noFisico.childNodes[i]);
          }
          reconciliar(noFisico, correspondencia.vno, novoChildVNo, i);
          mapaVelhos.delete(k);
        } else {
          const novoEl = criarNo(novoChildVNo);
          if (typeof noFisico.insertBefore === "function" && noFisico.childNodes && noFisico.childNodes[i]) {
            noFisico.insertBefore(novoEl, noFisico.childNodes[i]);
          } else if (typeof noFisico.appendChild === "function") {
            noFisico.appendChild(novoEl);
          }
        }
      });

      mapaVelhos.forEach(({ el }) => {
        if (el && el.parentNode === noFisico) {
          noFisico.removeChild(el);
        }
      });
    } else {
      if (velhosFilhos.length > novosFilhos.length) {
        let subIdx = 0;
        for (let i = 0; i < novosFilhos.length; i++) {
          subIdx += contarNosFisicos(velhosFilhos[i]);
        }
        for (let i = velhosFilhos.length - 1; i >= novosFilhos.length; i--) {
          reconciliar(noFisico, velhosFilhos[i], null, subIdx);
        }
      }
      let domOffset = 0;
      for (let i = 0; i < novosFilhos.length; i++) {
        reconciliar(noFisico, velhosFilhos[i], novosFilhos[i], domOffset);
        domOffset += contarNosFisicos(novosFilhos[i]);
      }
    }
  }
}

function mudou(no1, no2) {
  if (no1 === no2) return false;
  if (!no1 || !no2) return no1 !== no2;
  if (Array.isArray(no1) || Array.isArray(no2)) {
    return !Array.isArray(no1) || !Array.isArray(no2);
  }
  return (
    typeof no1 !== typeof no2 ||
    (typeof no1 === "string" && no1 !== no2) ||
    (typeof no1 === "number" && no1 !== no2) ||
    no1.tag !== no2.tag
  );
}

/**
 * Inicializa a montagem reativa da aplicação em um container DOM físico.
 * Suporta auto-hidratação transparente caso o elemento já contenha nós renderizados por SSR.
 * @param {Function} appComponente - Componente raiz
 * @param {HTMLElement} elementoAlvo - Elemento container no DOM
 */
export function montar(appComponente, elementoAlvo) {
  console.log("[Harpia Web] Inicializando montagem reativa da aplicação...");
  let velhoVNo = null;

  if (typeof elementoAlvo === "string") {
    let el = document.getElementById(elementoAlvo);
    if (!el && typeof document !== "undefined") {
      el = document.createElement("div");
      el.id = elementoAlvo;
      if (document.body) document.body.appendChild(el);
    }
    elementoAlvo = el;
  }

  if (!elementoAlvo && typeof document !== "undefined" && document.body) {
    let el = document.getElementById("app");
    if (!el) {
      el = document.createElement("div");
      el.id = "app";
      document.body.appendChild(el);
    }
    elementoAlvo = el;
  }

  if (elementoAlvo) {
    elementoAlvo.innerHTML = "";
  }

  const obterVNo =
    typeof appComponente === "function" ? appComponente : () => appComponente;

  efeito(() => {
    try {
      const novoVNo = obterVNo();
      console.log("[Harpia Web] Disparando efeito de renderização global...");
      reconciliar(elementoAlvo, velhoVNo, novoVNo);
      velhoVNo = novoVNo;
    } catch (err) {
      console.error("[Harpia Web] Erro de renderização no componente:", err);
      if (elementoAlvo) {
        elementoAlvo.innerHTML =
          '<div style="background:#1E1E2E; color:#F87171; font-family:monospace; padding:24px; border-radius:12px; margin:20px; border:1px solid #EF4444; box-shadow:0 10px 30px rgba(0,0,0,0.5);">' +
            '<h2 style="color:#FCA5A5; margin-top:0;">Erro de Renderização Harpia (Runtime)</h2>' +
            '<p style="color:#E5E7EB; margin-bottom:12px;">Ocorreu uma exceção ao executar o componente web:</p>' +
            '<pre style="background:#0F0F17; color:#F3F4F6; padding:16px; border-radius:8px; overflow-x:auto; white-space:pre-wrap; font-size:13px; line-height:1.5; border:1px solid #374151;">' + String(err && err.stack ? err.stack : err) + '</pre>' +
          '</div>';
      }
    }
  });
}

/**
  * Hidrata de forma não-destrutiva uma árvore DOM pré-renderizada via SSR no servidor.
  * Preserva os elementos HTML existentes e vincula reatividade e ouvintes de eventos dinâmicos.
  * @param {Function} appComponente - Componente raiz
  * @param {HTMLElement|string} elementoAlvo - Elemento container no DOM
  */
export function hidratar(appComponente, elementoAlvo) {
  if (typeof elementoAlvo === "string") {
    elementoAlvo = document.getElementById(elementoAlvo);
  }
  if (!elementoAlvo) return montar(appComponente, elementoAlvo);

  const obterVNo =
    typeof appComponente === "function" ? appComponente : () => appComponente;

  try {
    const vnoInicial = obterVNo();
    if (elementoAlvo.childNodes.length > 0) {
      vincularNos(elementoAlvo.firstElementChild || elementoAlvo, vnoInicial);
    } else {
      montar(appComponente, elementoAlvo);
    }
  } catch (e) {
    console.warn("[Harpia Web] Hidratação falhou, fallback para montagem limpa:", e);
    montar(appComponente, elementoAlvo);
  }
}

function vincularNos(elFisico, vno) {
  if (!vno || !elFisico) return;
  vno.el = elFisico;

  if (typeof vno === "object") {
    if (typeof vno.tag === "function") {
      const vnodeRendido = vno.tag(vno.props);
      vno._componenteInstancia = vnodeRendido;
      vincularNos(elFisico, vnodeRendido);
      return;
    }

    for (const chave in vno.props) {
      if (/^ao/i.test(chave)) {
        const evento = mapearEvento(chave);
        elFisico.addEventListener(evento, vno.props[chave]);
      }
    }

    const filhosFisicos = elFisico.childNodes;
    const max = Math.min(
      filhosFisicos.length,
      vno.children ? vno.children.length : 0,
    );
    for (let i = 0; i < max; i++) {
      vincularNos(filhosFisicos[i], vno.children[i]);
    }
  }
}

// ============================================================================
// 3. ROTEAMENTO SPA BASEADO EM SINAIS (HISTORY API)
// ============================================================================

const [urlAtiva, setUrlAtiva] = sinal(
  typeof window !== "undefined" ? window.location.pathname : "/",
);

/**
 * Navega dinamicamente para uma nova rota do SPA sem recarregar a página.
 * @param {string} destino - O caminho de URL de destino (ex: "/sobre")
 */
export function navegar(destino) {
  if (typeof window !== "undefined" && window.location.pathname !== destino) {
    window.history.pushState({}, "", destino);
    setUrlAtiva(destino);
  }
}

if (typeof window !== "undefined") {
  window.addEventListener("popstate", () => {
    setUrlAtiva(window.location.pathname);
  });

  window.addEventListener("click", (e) => {
    const a = e.target.closest && e.target.closest("a");
    if (a && a.href && a.origin === window.location.origin) {
      const target = a.getAttribute("target");
      if (!target || target === "_self") {
        const url = new URL(a.href);
        if (url.pathname !== window.location.pathname) {
          e.preventDefault();
          navegar(url.pathname + url.search + url.hash);
        }
      }
    }
  });
}

/**
 * Cria um componente funcional de Roteador baseado em Sinais.
 * @param {Object} rotas - Mapeamento de caminhos para componentes funcionais
 * @returns {Function} Componente do roteador
 */
export function roteador(rotas) {
  return () => {
    let path = urlAtiva();
    if (path.length > 1 && path.endsWith("/")) {
      path = path.slice(0, -1);
    }
    let componente = rotas[path];
    let parametros = {};

    if (!componente) {
      let pathDecodificado = path;
      try {
        pathDecodificado = decodeURIComponent(path);
      } catch (e) {}

      componente = rotas[pathDecodificado] || rotas[path.toLowerCase()] || rotas[pathDecodificado.toLowerCase()];

      if (!componente) {
        const pathLower = pathDecodificado.toLowerCase();
        for (const r in rotas) {
          if (r.toLowerCase() === pathLower) {
            componente = rotas[r];
            break;
          }
        }
      }
    }

    if (!componente) {
      for (const rotaDef in rotas) {
        if (!rotaDef.includes(":")) continue;
        const partesDef = rotaDef.split("/").filter(Boolean);
        const partesPath = path.split("/").filter(Boolean);

        if (partesDef.length === partesPath.length) {
          let corresponde = true;
          const paramsTemp = {};
          for (let i = 0; i < partesDef.length; i++) {
            if (partesDef[i].startsWith(":")) {
              paramsTemp[partesDef[i].slice(1)] = partesPath[i];
            } else if (partesDef[i] !== partesPath[i]) {
              corresponde = false;
              break;
            }
          }
          if (corresponde) {
            componente = rotas[rotaDef];
            parametros = paramsTemp;
            break;
          }
        }
      }
    }

    componente =
      componente ||
      rotas["/404"] ||
      (() => h("div", {}, "404 - Página Não Encontrada"));
    return h(componente, { parametros });
  };
}

// ============================================================================
// 4. PRIMITIVAS DE ESTADO CORPORATIVAS E COMPONENTES DE UI NATIVOS (Fase 4-C)
// ============================================================================

/**
 * Cria um sinal que é automaticamente persistido no localStorage do navegador.
 * @param {string} chave - Chave para armazenar no localStorage
 * @param {*} valorInicial - Valor padrão caso não exista nada persistido
 */
export function sinalPersistente(chave, valorInicial) {
  let valorSalvo = valorInicial;
  if (typeof window !== "undefined" && typeof localStorage !== "undefined") {
    const item = localStorage.getItem(chave);
    if (item !== null) {
      try {
        valorSalvo = JSON.parse(item);
      } catch (e) {
        valorSalvo = item;
      }
    }
  }
  const [ler, definir] = sinal(valorSalvo);
  const definirPersistente = (novoValor) => {
    definir(novoValor);
    if (typeof window !== "undefined" && typeof localStorage !== "undefined") {
      localStorage.setItem(chave, JSON.stringify(novoValor));
    }
  };
  ler.set = definirPersistente; // ponytail: expõe setter direto no getter para bind duplo
  return [ler, definirPersistente];
}

/**
 * Primitiva nativa de estado assíncrono para chamadas de rede/APIs.
 * @param {Function} funcaoAsync - Função que retorna uma Promise
 */
export function recurso(funcaoAsync) {
  const [dados, setDados] = sinal(null);
  const [carregando, setCarregando] = sinal(true);
  const [erro, setErro] = sinal(null);

  const executar = async () => {
    setCarregando(true);
    setErro(null);
    try {
      const res = await funcaoAsync();
      setDados(res);
    } catch (e) {
      setErro(e);
    } finally {
      setCarregando(false);
    }
  };

  executar();

  const ler = () => dados();
  const fnCarregando = () => carregando();
  fnCarregando.set = setCarregando;
  const fnErro = () => erro();
  fnErro.set = setErro;

  ler.dados = dados;
  ler.carregando = fnCarregando;
  ler.erro = fnErro;
  ler.ok = () => !carregando() && !erro();
  ler.recarregar = executar;

  const res = ler;
  res.dados = dados;
  res.carregando = fnCarregando;
  res.erro = fnErro;
  res.recarregar = executar;
  return res;
}

/**
 * Gestor de formulários reativos no Harpia (usarFormulario).
 */
export function usarFormulario(config) {
  const valoresIniciais = config.valoresIniciais || {};
  const [valores, setValores] = sinal({ ...valoresIniciais });
  const [erros, setErros] = sinal({});

  function campo(nome) {
    return {
      value: valores()[nome] || "",
      onInput: (e) => {
        const novos = { ...valores(), [nome]: e.target.value };
        setValores(novos);
        if (config.validar) {
          setErros(config.validar(novos) || {});
        }
      },
    };
  }

  function submeter(aoSubmeter) {
    return (e) => {
      if (e && e.preventDefault) e.preventDefault();
      const errs = config.validar ? config.validar(valores()) : {};
      setErros(errs);
      if (Object.keys(errs).length === 0) {
        aoSubmeter(valores());
      }
    };
  }

  return { valores, erros, campo, submeter };
}

/**
 * Cache global e hook de consulta HTTP inteligente (usarConsulta / SWR).
 */
const cacheGlobalConsultas = new Map();

export function usarConsulta(url, opcoes = {}) {
  const [dados, setDados] = sinal(cacheGlobalConsultas.get(url) || null);
  const [carregando, setCarregando] = sinal(!cacheGlobalConsultas.has(url));
  const [erro, setErro] = sinal(null);

  async function buscar() {
    try {
      setCarregando(true);
      const res = await fetch(url);
      const data = await res.json();
      cacheGlobalConsultas.set(url, data);
      setDados(data);
      setErro(null);
    } catch (err) {
      setErro(err.message || "Erro ao carregar consulta");
    } finally {
      setCarregando(false);
    }
  }

  efeito(() => {
    buscar();
  });

  return {
    dados,
    carregando,
    erro,
    mutarOtimista: (novo) => setDados(novo),
    recarregar: buscar,
  };
}

/**
 * Componente de animações e transições nativas (Animacao).
 */
export function Animacao(props) {
  const tipo = props.tipo || "fade";
  const duracao = props.duracao || "300ms";
  return h(
    "div",
    {
      estilo: {
        transition: "all " + duracao + " ease-in-out",
        animation: tipo + " " + duracao,
      },
    },
    props.children,
  );
}

/**
 * Componente de imagem otimizada com lazy loading (ImagemWeb).
 */
export function ImagemWeb(props) {
  return h("img", {
    src: props.caminho || props.src,
    alt: props.alt || "",
    loading: "lazy",
    width: props.largura,
    height: props.altura,
    classe: props.classe || "",
  });
}

/**
 * Carregador assíncrono de Micro-Frontends (moduloRemoto).
 */
export function moduloRemoto(url, nomeComponente) {
  return function ComponenteRemoto(props) {
    const [comp, setComp] = sinal(null);
    efeito(() => {
      import(/* webpackIgnore: true */ url).then((mod) => {
        setComp(() => mod[nomeComponente] || mod.default);
      });
    });
    return () => {
      const C = comp();
      return C
        ? h(C, props)
        : h(
            "div",
            { classe: "carregando-remoto" },
            "Carregando Módulo Remoto...",
          );
    };
  };
}

/**
 * Gestor de Tema Escuro/Claro (usarTema).
 */
export function usarTema(chave = "harpia_tema", padrao = "sistema") {
  const preferEscuro =
    window.matchMedia &&
    window.matchMedia("(prefers-color-scheme: dark)").matches;
  const salvo = localStorage.getItem(chave);
  const inicial =
    salvo ||
    (padrao === "sistema" ? (preferEscuro ? "escuro" : "claro") : padrao);
  const [tema, setTema] = sinal(inicial);

  efeito(() => {
    const atual = tema();
    localStorage.setItem(chave, atual);
    if (atual === "escuro") {
      document.documentElement.classList.add("escuro");
    } else {
      document.documentElement.classList.remove("escuro");
    }
  });

  const alternar = () => setTema(tema() === "escuro" ? "claro" : "escuro");
  return [tema, alternar, setTema];
}

/**
 * Componente Portal para renderização direta no document.body (Portal).
 */
export function Portal(props) {
  const el = document.createElement("div");
  el.className = "harpia-portal";
  efeito(() => {
    document.body.appendChild(el);
    montar(h("div", {}, props.children), el);
    return () => {
      if (el.parentNode) el.parentNode.removeChild(el);
    };
  });
  return null;
}

/**
 * Hook para Drag-and-Drop e Gestos Touch (usarArrastar).
 */
export function usarArrastar(aoSoltar) {
  let itemArrastado = null;
  return {
    aoIniciarArrasto: (item) => (e) => {
      itemArrastado = item;
      if (e.dataTransfer)
        e.dataTransfer.setData("text/plain", JSON.stringify(item));
    },
    aoSobrepor: (e) => e.preventDefault(),
    aoSoltar: (itemAlvo) => (e) => {
      if (e.preventDefault) e.preventDefault();
      if (aoSoltar && itemArrastado) {
        aoSoltar(itemArrastado, itemAlvo);
      }
    },
  };
}

/**
 * Barra de progresso visual no topo do navegador (BarraDeProgresso).
 */
export function BarraDeProgresso(props) {
  const [carregando, setCarregando] = sinal(false);
  return h("div", {
    estilo: {
      position: "fixed",
      top: "0",
      left: "0",
      height: "3px",
      width: carregando() ? "100%" : "0%",
      backgroundColor: props.cor || "#3b82f6",
      transition: "width 300ms ease-in-out",
      zIndex: "99999",
    },
  });
}

/**
 * Sistema de notificações tostadas reativas (usarNotificacao & NotificacaoToast).
 */
const [notificacoes, setNotificacoes] = sinal([]);

export function usarNotificacao() {
  function adicionar(
    mensagemOuOpcoes,
    tipo = "info",
    duracao = 4000,
    opcoes = {},
  ) {
    let config = {};
    if (typeof mensagemOuOpcoes === "object" && mensagemOuOpcoes !== null) {
      config = { ...mensagemOuOpcoes };
    } else {
      config = {
        mensagem: String(mensagemOuOpcoes || ""),
        tipo,
        duracao,
        ...opcoes,
      };
    }

    const id =
      config.id || Date.now() + Math.random().toString(36).substring(2, 7);
    const item = {
      id,
      mensagem: config.mensagem || "",
      tipo: config.tipo || "info",
      duracao: config.duracao !== undefined ? config.duracao : 4000,
      posicao: config.posicao || "topo-direita",
      icone: config.icone,
      acaoTexto: config.acaoTexto,
      aoClicarAcao: config.aoClicarAcao,
      saindo: false,
    };

    queueMicrotask(() => {
      setNotificacoes([...notificacoes().filter((n) => n.id !== id), item]);
    });

    if (item.duracao && item.duracao > 0) {
      item.timerId = setTimeout(() => {
        remover(id);
      }, item.duracao);
    }
    return id;
  }

  function remover(id) {
    const lista = notificacoes();
    const item = lista.find((n) => n.id === id);
    if (item) {
      if (item.timerId) clearTimeout(item.timerId);
      item.saindo = true;
      setNotificacoes([...lista]);
      setTimeout(() => {
        setNotificacoes(notificacoes().filter((n) => n.id !== id));
      }, 300);
    }
  }

  function limpar() {
    setNotificacoes([]);
  }

  return {
    adicionar,
    remover,
    limpar,
    sucesso: (msg, dur, opts) => adicionar(msg, "sucesso", dur, opts),
    erro: (msg, dur, opts) => adicionar(msg, "erro", dur, opts),
    info: (msg, dur, opts) => adicionar(msg, "info", dur, opts),
    avisar: (msg, dur, opts) => adicionar(msg, "aviso", dur, opts),
    lista: notificacoes,
  };
}

export function NotificacaoToast(props) {
  props = props || {};
  const { lista, remover } = usarNotificacao();
  const posProp = props.posicao || "topo-direita";

  if (
    typeof document !== "undefined" &&
    !document.getElementById("harpia-toast-styles")
  ) {
    const styleEl = document.createElement("style");
    styleEl.id = "harpia-toast-styles";
    styleEl.textContent =
      "@keyframes harpiaToastEntrada { 0% { opacity: 0; transform: translateX(60px) scale(0.92); } 100% { opacity: 1; transform: translateX(0) scale(1); } } @keyframes harpiaToastSaida { 0% { opacity: 1; transform: translateX(0) scale(1); max-height: 100px; margin-bottom: 0px; } 100% { opacity: 0; transform: translateX(60px) scale(0.92); max-height: 0px; margin-bottom: -10px; padding: 0; } } @keyframes harpiaToastProgresso { from { width: 100%; } to { width: 0%; } } .harpia-toast-item { animation: harpiaToastEntrada 0.4s cubic-bezier(0.16, 1, 0.3, 1) forwards; will-change: transform, opacity; } .harpia-toast-item.saindo { animation: harpiaToastSaida 0.35s cubic-bezier(0.16, 1, 0.3, 1) forwards; }";
    document.head.appendChild(styleEl);
  }

  const estilosPosicao = {
    position: "fixed",
    zIndex: "999999",
    display: "flex",
    flexDirection: "column",
    gap: "10px",
    maxWidth: "420px",
    width: "calc(100% - 32px)",
    pointerEvents: "none",
    top: posProp.includes("topo") ? "20px" : "auto",
    bottom: posProp.includes("rodape") ? "20px" : "auto",
    right: posProp.includes("direita") ? "20px" : "auto",
    left: posProp.includes("esquerda")
      ? "20px"
      : posProp.includes("centro")
        ? "50%"
        : "auto",
    transform: posProp.includes("centro") ? "translateX(-50%)" : "none",
  };

  const temasCores = {
    sucesso: {
      border: "#10b981",
      texto: "#ecfdf5",
      icone: "✓",
      barra: "#10b981",
    },
    erro: { border: "#f43f5e", texto: "#fff1f2", icone: "✕", barra: "#f43f5e" },
    aviso: {
      border: "#f59e0b",
      texto: "#fffbeb",
      icone: "⚡",
      barra: "#f59e0b",
    },
    info: { border: "#3b82f6", texto: "#eff6ff", icone: "ℹ", barra: "#3b82f6" },
  };

  return h(
    "div",
    { style: estilosPosicao },
    lista().map((item) => {
      const tema = temasCores[item.tipo] || temasCores.info;
      const dur = item.duracao !== undefined ? item.duracao : 4000;

      return h(
        "div",
        {
          key: item.id,
          className: "harpia-toast-item " + (item.saindo ? "saindo" : ""),
          style: {
            pointerEvents: "auto",
            position: "relative",
            overflow: "hidden",
            display: "flex",
            alignItems: "center",
            justifyContent: "space-between",
            padding: "14px 16px",
            borderRadius: "10px",
            background: "#18181b",
            borderLeft: "4px solid " + tema.border,
            boxShadow:
              "0 10px 25px -5px rgba(0, 0, 0, 0.5), 0 8px 10px -6px rgba(0, 0, 0, 0.3)",
            backdropFilter: "blur(12px)",
            color: tema.texto,
            fontSize: "14px",
            fontWeight: "500",
            fontFamily: "sans-serif",
          },
        },
        [
          h(
            "div",
            { style: { display: "flex", alignItems: "center", gap: "12px" } },
            [
              h(
                "span",
                {
                  style: {
                    fontSize: "16px",
                    color: tema.border,
                    fontWeight: "bold",
                  },
                },
                item.icone || tema.icone,
              ),
              h("span", {}, item.mensagem),
            ],
          ),
          h(
            "div",
            { style: { display: "flex", alignItems: "center", gap: "8px" } },
            [
              item.acaoTexto
                ? h(
                    "button",
                    {
                      onClick: () => {
                        if (item.aoClicarAcao) item.aoClicarAcao();
                        remover(item.id);
                      },
                      style: {
                        background: "transparent",
                        border: "1px solid " + tema.border,
                        color: tema.border,
                        padding: "4px 8px",
                        borderRadius: "4px",
                        cursor: "pointer",
                        fontSize: "12px",
                        fontWeight: "bold",
                      },
                    },
                    item.acaoTexto,
                  )
                : null,
              h(
                "button",
                {
                  onClick: () => remover(item.id),
                  style: {
                    background: "transparent",
                    border: "none",
                    color: "#9ca3af",
                    cursor: "pointer",
                    fontSize: "16px",
                    lineHeight: "1",
                    padding: "4px",
                  },
                },
                "✕",
              ),
            ],
          ),
          dur > 0
            ? h("div", {
                style: {
                  position: "absolute",
                  bottom: "0",
                  left: "0",
                  height: "3px",
                  background: tema.barra,
                  animation: "harpiaToastProgresso " + dur + "ms linear forwards",
                },
              })
            : null,
        ],
      );
    }),
  );
}

const mapaContextos = new Map();

/**
 * Provedor de contexto para injeção de dependências sem prop-drilling.
 */
export function Provedor(props) {
  const anterior = mapaContextos.get(props.chave);
  mapaContextos.set(props.chave, props.valor);
  const res = props.children || null;
  queueMicrotask(() => {
    if (anterior !== undefined) {
      mapaContextos.set(props.chave, anterior);
    } else {
      mapaContextos.delete(props.chave);
    }
  });
  return h("div", { estilo: { display: "contents" } }, res);
}

/**
 * Injeta/recupera um serviço provido por um componente Provedor superior.
 * @param {string} chave - Identificador do serviço
 */
export function injetar(chave) {
  return mapaContextos.get(chave);
}

/**
 * Fronteira de Erro / Barreira de Erro nativa para capturar falhas em componentes filhos secundários.
 */
export function FronteiraDeErro(props) {
  const [erro, setErro] = sinal(null);
  try {
    if (erro()) {
      return props.fallback || h("p", {}, "Erro ao carregar componente.");
    }
    return props.children || null;
  } catch (e) {
    setErro(e);
    return props.fallback || h("p", {}, "Erro ao carregar componente.");
  }
}

export const BarreiraDeErro = FronteiraDeErro;

/**
 * Componente nativo de navegação SPA sem recarregamento de página.
 */
export function Link(props) {
  const { para, precarregar, children, ...restanteProps } = props || {};
  const aoClicarOriginal = restanteProps.aoClicar;
  restanteProps.aoClicar = (e) => {
    if (e && e.preventDefault) e.preventDefault();
    if (para && para !== "#") navegar(para);
    if (aoClicarOriginal) aoClicarOriginal(e);
  };
  if (precarregar && para && para !== "#") {
    restanteProps.aoEntrarMouse = () => {
      if (
        typeof window !== "undefined" &&
        window.__harpiaRotas &&
        window.__harpiaRotas[para]
      ) {
        // Pré-carrega assincronamente a rota na memória antes do clique
        const rotaFn = window.__harpiaRotas[para];
        if (typeof rotaFn === "function") rotaFn();
      }
    };
  }
  restanteProps.href = para || "#";
  return h("a", restanteProps, children);
}

/**
 * Componente nativo de Suspense/Aguardar para dados assíncronos e recursos.
 */
export function Aguardar(props) {
  const { recurso: res, carregando, erro: renderErro, children } = props || {};
  if (!res) return typeof children === "function" ? children(null) : children;

  if (typeof res.carregando === "function" && res.carregando()) {
    return carregando || h("p", { class: "carregando" }, "⌛ Carregando...");
  }

  if (typeof res.erro === "function" && res.erro()) {
    const err = res.erro();
    return typeof renderErro === "function"
      ? renderErro(err)
      : renderErro || h("p", { class: "erro" }, "❌ Erro: " + err);
  }

  const dados = typeof res.dados === "function" ? res.dados() : res.dados;
  if (typeof children === "function") {
    return children(dados);
  }
  return children;
}

/**
 * Construto nativo de Pattern Matching de UI (Escolha/Caso/Padrao).
 */
export function Escolha(props) {
  const valorAlvo =
    typeof props.valor === "function" ? props.valor() : props.valor;
  const filhos = Array.isArray(props.children)
    ? props.children
    : [props.children];
  let casoCorrespondente = null;
  let casoPadrao = null;

  for (const filho of filhos) {
    if (!filho) continue;
    if (
      filho.tag === Caso ||
      (filho.props && filho.props.valor !== undefined)
    ) {
      if (filho.props.valor === valorAlvo) {
        casoCorrespondente = filho;
        break;
      }
    } else if (filho.tag === Padrao || (filho.props && filho.props.isPadrao)) {
      casoPadrao = filho;
    }
  }

  const escolhido = casoCorrespondente || casoPadrao;
  return escolhido ? escolhido.children || escolhido : null;
}

export function Caso(props) {
  return props.children || null;
}

export function Padrao(props) {
  return props.children || null;
}

/**
 * Componente de lista virtualizada de alta performance para renderizar coleções massivas.
 */
export function ListaVirtual(props) {
  const [inicio, setInicio] = sinal(0);
  const total = props.itens ? props.itens.length : 0;
  const alturaLinha = props.alturaLinha || 40;
  const alturaContainer = props.alturaContainer || 400;
  const visiveis = Math.ceil(alturaContainer / alturaLinha) + 2;

  const aoRolar = (e) => {
    const topo = e.target.scrollTop;
    const idx = Math.floor(topo / alturaLinha);
    setInicio(Math.max(0, Math.min(idx, total - visiveis)));
  };

  return () => {
    const itensVisiveis = (props.itens || []).slice(
      inicio(),
      inicio() + visiveis,
    );
    const espacoSuperior = inicio() * alturaLinha;
    const espacoInferior = Math.max(
      0,
      (total - inicio() - itensVisiveis.length) * alturaLinha,
    );

    return h(
      "div",
      {
        aoScroll: aoRolar,
        estilo: {
          height: alturaContainer + "px",
          overflowY: "auto",
          position: "relative",
        },
      },
      h("div", { estilo: { height: espacoSuperior + "px" } }),
      ...itensVisiveis.map((item, idx) =>
        props.children[0](item, inicio() + idx),
      ),
      h("div", { estilo: { height: espacoInferior + "px" } }),
    );
  };
}

/**
 * Grade de dados (Data Table) avançada com pesquisa, paginação e design responsivo.
 */
export function GradeDeDados(props) {
  const [pagina, setPagina] = sinal(1);
  const [filtro, setFiltro] = sinal("");
  const colunas = props.colunas || [];
  const dadosOriginais = props.dados || [];
  const limite = props.linhasPorPagina || 10;

  return () => {
    const dadosFiltrados = dadosOriginais.filter((item) => {
      const texto = filtro().toLowerCase();
      if (!texto) return true;
      return colunas.some((col) =>
        String(item[col] || "")
          .toLowerCase()
          .includes(texto),
      );
    });

    const totalPaginas = Math.ceil(dadosFiltrados.length / limite) || 1;
    const inicio = (pagina() - 1) * limite;
    const dadosPaginados = dadosFiltrados.slice(inicio, inicio + limite);

    return h(
      "div",
      { classe: "grade-dados-container" },
      h("input", {
        _ligar: filtro,
        placeholder: "Pesquisar...",
        classe: "grade-dados-pesquisa",
      }),
      h(
        "table",
        { classe: "grade-dados-tabela" },
        h("thead", {}, h("tr", {}, ...colunas.map((col) => h("th", {}, col)))),
        h(
          "tbody",
          {},
          ...dadosPaginados.map((item) =>
            h(
              "tr",
              {},
              ...colunas.map((col) => h("td", {}, String(item[col] || ""))),
            ),
          ),
        ),
      ),
      h(
        "div",
        { classe: "grade-dados-paginacao" },
        h(
          "button",
          {
            aoClicar: () => setPagina(Math.max(1, pagina() - 1)),
            disabled: pagina() === 1,
          },
          "Anterior",
        ),
        h("span", {}, " Página " + pagina() + " de " + totalPaginas + " "),
        h(
          "button",
          {
            aoClicar: () => setPagina(Math.min(totalPaginas, pagina() + 1)),
            disabled: pagina() === totalPaginas,
          },
          "Próxima",
        ),
      ),
    );
  };
}

/**
 * Carregador preguiçoso para Code Splitting e Lazy Loading de componentes.
 */
export function preguicoso(importarComponente) {
  return function ComponentePreguicoso(props) {
    const [comp, setComp] = sinal(null);
    importarComponente().then((Modulo) => {
      setComp(() => Modulo.default || Modulo);
    });
    return () => {
      const C = comp();
      return C
        ? h(C, props)
        : h("div", { classe: "carregando-lazy" }, "Carregando...");
    };
  };
}

/**
 * Componente Suspense reativo simples.
 */
export function Suspense(props) {
  return () => {
    const pronto =
      typeof props.pronto === "function" ? props.pronto() : props.pronto;
    return pronto
      ? props.children
      : props.fallback || h("div", {}, "Carregando...");
  };
}

/**
 * Realiza requisições HTTP assíncronas (fetch) e retorna o resultado.
 */
export async function requisitar(metodo, url, dados = null, cabecalhos = {}) {
  const opcoes = {
    method: metodo,
    headers: {
      "Content-Type": "application/json",
      ...cabecalhos,
    },
  };
  if (dados) {
    opcoes.body = typeof dados === "string" ? dados : JSON.stringify(dados);
  }
  const resposta = await fetch(url, opcoes);
  if (!resposta.ok) {
    throw new Error(
      "Erro na requisição: " + resposta.status + " " + resposta.statusText,
    );
  }
  const contentType = resposta.headers.get("content-type");
  if (contentType && contentType.includes("application/json")) {
    return resposta.json();
  }
  return resposta.text();
}

export function sanitizarHtml(str) {
  return str
    ? String(str)
        .replace(/&/g, "&amp;")
        .replace(/</g, "&lt;")
        .replace(/>/g, "&gt;")
        .replace(/"/g, "&quot;")
        .replace(/'/g, "&#39;")
    : "";
}
export function sanitizarSqlArgumento(str) {
  return str ? String(str).replace(/'/g, "''") : "";
}
export function ler(caminho) {
  return "";
}
export function escrever(caminho, conteudo) {
  return true;
}
export function existe(caminho) {
  return false;
}

/**
 * Hook para atualização reativa dinâmica do <title> e tags <meta> no document.head.
 * @param {Object|Function} opcoes - Configurações do cabeçalho ({ titulo, meta })
 */
export function usarCabecalho(opcoes) {
  efeito(() => {
    if (typeof document === "undefined") return;

    const opts = typeof opcoes === "function" ? opcoes() : (opcoes || {});
    const rawTitulo = opts.titulo !== undefined ? opts.titulo : opts.title;
    const titulo = typeof rawTitulo === "function" ? rawTitulo() : rawTitulo;

    if (titulo !== undefined) {
      document.title = String(titulo);
    }

    const rawMeta = opts.meta;
    const metaList = typeof rawMeta === "function" ? rawMeta() : rawMeta;

    if (Array.isArray(metaList)) {
      if (!document.head && typeof document.createElement === "function") {
        document.head = document.createElement("head");
      }
      const head = document.head;
      if (head) {
        metaList.forEach((m) => {
          if (!m) return;
          const nome = m.nome || m.name;
          const prop = m.property;
          const rawConteudo = m.conteudo !== undefined ? m.conteudo : m.content;
          const conteudo = typeof rawConteudo === "function" ? rawConteudo() : rawConteudo;

          let el = null;
          if (typeof head.querySelector === "function") {
            if (nome) el = head.querySelector('meta[name="' + nome + '"]');
            else if (prop) el = head.querySelector('meta[property="' + prop + '"]');
          } else if (head.childNodes) {
            for (const child of head.childNodes) {
              if (child.tag === "meta" && child.props) {
                if (nome && (child.props.name === nome || child.props.nome === nome)) {
                  el = child;
                  break;
                }
                if (prop && child.props.property === prop) {
                  el = child;
                  break;
                }
              }
            }
          }

          if (!el) {
            el = document.createElement("meta");
            if (nome) el.setAttribute("name", nome);
            if (prop) el.setAttribute("property", prop);
            head.appendChild(el);
          }

          if (conteudo !== undefined) {
            el.setAttribute("content", String(conteudo));
          }
        });
      }
    }
  });
}

/**
 * Componente para atualização reativa dinâmica do <title> e tags <meta> no document.head.
 */
export function CabecalhoPagina(props) {
  usarCabecalho(props);
  return null;
}

/**
 * Objeto de inspeção global e Métricas DevTools do Harpia Web.
 */
export const devTools = {
  get sinaisAtivos() {
    return registroSinais.size;
  },
  get totalSinais() {
    return devToolsMetrics.sinaisCriados;
  },
  get totalEfeitos() {
    return devToolsMetrics.efeitosCriados;
  },
  get efeitosAtivos() {
    return registroEfeitos.size;
  },
  get renderizacoesVDOM() {
    return devToolsMetrics.renderizacoesVDOM;
  },
  obterSinaisAtivos() {
    return registroSinais.size;
  },
  obterEfeitosRegistrados() {
    return devToolsMetrics.efeitosCriados;
  },
  obterMetricasRender() {
    return {
      renderizacoes: devToolsMetrics.renderizacoesVDOM,
      sinaisAtivos: registroSinais.size,
      totalEfeitos: devToolsMetrics.efeitosCriados,
      efeitosAtivos: registroEfeitos.size,
    };
  },
};

export const __HARPIA_DEVTOOLS__ = devTools;

if (typeof globalThis !== "undefined") {
  globalThis.__HARPIA_DEVTOOLS__ = devTools;
}

// ponytail: vincula as primitivas ao escopo global do navegador para compatibilidade de transpilação
if (typeof window !== "undefined") {
  Object.assign(window, {
    Nulo: null,
    Verdadeiro: true,
    Falso: false,

    sinal,
    sinalDebounce,
    semRastreio,
    untrack,
    lote,
    batch,
    aoMontar,
    aoDesmontar,
    efeito,
    derivado,
    armazem,
    h,
    criarNo,
    reconciliar,
    montar,
    navegar,
    roteador,
    sinalPersistente,
    recurso,
    Provedor,
    injetar,
    FronteiraDeErro,
    ListaVirtual,
    GradeDeDados,
    preguicoso,
    Suspense,
    requisitar,

    Link,
    Aguardar,
    Escolha,
    Caso,
    Padrao,
    usarFormulario,
    Animacao,
    ImagemWeb,
    moduloRemoto,
    usarTema,
    Portal,
    usarArrastar,
    BarraDeProgresso,
    usarNotificacao,
    NotificacaoToast,
    Fragmento,
    Fragment,

    usarCabecalho,
    CabecalhoPagina,
    devTools,
    __HARPIA_DEVTOOLS__: devTools,
  });
}
`
