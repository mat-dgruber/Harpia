/**
 * Portuscript Web Runtime - VDOM & Reatividade Fina
 * ponytail: Solução ultraleve (~2.5KB) sem dependências externas.
 * Fornece Virtual DOM, reconciliação cirúrgica e reatividade fina baseada em Sinais.
 */

let efeitoAtivo = null;

// ============================================================================
// 1. SISTEMA DE SINAIS E REATIVIDADE FINA
// ============================================================================

/**
 * Cria um sinal reativo contendo um valor.
 * @param {*} valor - Valor inicial
 * @returns {[Function, Function]} [ler, definir]
 */
export function sinal(valor) {
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
      paraDisparar.forEach(sub => sub.run());
    }
  };

  return [ler, definir];
}

/**
 * Cria um efeito colateral que roda imediatamente e se reinscreve nas mudanças de sinais lidos dentro dele.
 * @param {Function} funcao
 * @returns {Function} Função de limpeza (cleanup)
 */
export function efeito(funcao) {
  const ef = {
    deps: new Set(),
    run() {
      limparDeps(ef);
      const anterior = efeitoAtivo;
      efeitoAtivo = ef;
      try {
        funcao();
      } finally {
        efeitoAtivo = anterior;
      }
    }
  };
  ef.run();
  return () => limparDeps(ef);
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
    }
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
  const children = filhos.flat(Infinity).filter(f => f !== null && f !== undefined && f !== false);
  return { tag, props, children };
}

/**
 * Instancia fisicamente um nó VDOM no DOM real do navegador.
 * @param {Object|string|number} vno
 * @returns {Node} Elemento do DOM físico
 */
export function criarNo(vno) {
  if (typeof vno === 'string' || typeof vno === 'number') {
    return document.createTextNode(vno);
  }

  if (typeof vno.tag === 'function') {
    const vnodeRendido = vno.tag(vno.props);
    vno._componenteInstancia = vnodeRendido;
    return criarNo(vnodeRendido);
  }

  const el = document.createElement(vno.tag);
  vno.el = el;

  atualizarAtributos(el, {}, vno.props);

  vno.children.forEach(filho => {
    el.appendChild(criarNo(filho));
  });

  return el;
}

function atualizarAtributos(el, velhosProps, novosProps) {
  // Remover velhos props
  for (const chave in velhosProps) {
    if (!(chave in novosProps)) {
      if (chave.startsWith('ao')) {
        const evento = mapearEvento(chave);
        el.removeEventListener(evento, velhosProps[chave]);
      } else {
        el.removeAttribute(mapearAtributo(chave));
      }
    }
  }

  // Adicionar ou atualizar novos props
  for (const chave in novosProps) {
    const valor = novosProps[chave];
    const velhoValor = velhosProps[chave];

    if (valor !== velhoValor) {
      if (chave.startsWith('ao')) {
        const evento = mapearEvento(chave);
        if (velhoValor) el.removeEventListener(evento, velhoValor);
        el.addEventListener(evento, valor);
      } else if (chave === 'estilo' && typeof valor === 'object') {
        Object.assign(el.style, valor);
      } else {
        const attrName = mapearAtributo(chave);
        if (valor === true) {
          el.setAttribute(attrName, '');
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
  const mapeamento = {
    aoClicar: 'click',
    aoMudar: 'change',
    aoEnviar: 'submit',
    aoFocar: 'focus',
    aoDesfocar: 'blur',
    aoTeclar: 'keydown',
    aoSoltarTecla: 'keyup',
    aoEntrarMouse: 'mouseenter',
    aoSairMouse: 'mouseleave'
  };
  return mapeamento[nome] || nome.toLowerCase().replace(/^ao/, '');
}

function mapearAtributo(nome) {
  const mapeamento = {
    classe: 'class',
    identificador: 'id',
    corDeFundo: 'bgcolor',
    largura: 'width',
    altura: 'height'
  };
  return mapeamento[nome] || nome;
}

/**
 * Reconcilia de forma cirúrgica um nó DOM real com as mudanças entre o VNode antigo e o novo.
 * @param {Node} pai - Elemento pai físico no DOM
 * @param {Object} velhoVNo - VNode anterior
 * @param {Object} novoVNo - Novo VNode
 * @param {number} index - Posição do nó físico no pai
 */
export function reconciliar(pai, velhoVNo, novoVNo, index = 0) {
  const noFisico = pai.childNodes[index];

  if (!velhoVNo) {
    pai.appendChild(criarNo(novoVNo));
    return;
  }

  if (!novoVNo) {
    pai.removeChild(noFisico);
    return;
  }

  if (mudou(velhoVNo, novoVNo)) {
    pai.replaceChild(criarNo(novoVNo), noFisico);
    return;
  }

  if (typeof novoVNo === 'object') {
    if (typeof novoVNo.tag === 'function') {
      const vnodeRendido = novoVNo.tag(novoVNo.props);
      novoVNo._componenteInstancia = vnodeRendido;
      reconciliar(pai, velhoVNo._componenteInstancia, vnodeRendido, index);
      return;
    }

    atualizarAtributos(noFisico, velhoVNo.props, novoVNo.props);
    novoVNo.el = noFisico;

    const max = Math.max(velhoVNo.children.length, novoVNo.children.length);
    for (let i = 0; i < max; i++) {
      reconciliar(noFisico, velhoVNo.children[i], novoVNo.children[i], i);
    }
  }
}

function mudou(no1, no2) {
  return (
    typeof no1 !== typeof no2 ||
    (typeof no1 === 'string' && no1 !== no2) ||
    (typeof no1 === 'number' && no1 !== no2) ||
    no1.tag !== no2.tag
  );
}

/**
 * Inicializa a montagem reativa da aplicação em um container DOM físico.
 * @param {Function} appComponente - Componente raiz
 * @param {HTMLElement} elementoAlvo - Elemento container no DOM
 */
export function montar(appComponente, elementoAlvo) {
  let velhoVNo = null;
  efeito(() => {
    const novoVNo = typeof appComponente === 'function' ? appComponente() : appComponente;
    reconciliar(elementoAlvo, velhoVNo, novoVNo);
    velhoVNo = novoVNo;
  });
}
