/**
 * Harpia Web Runtime - VDOM & Reatividade Fina
 * ponytail: Solução ultraleve (~2.5KB) sem dependências externas.
 * Fornece Virtual DOM, reconciliação cirúrgica e reatividade fina baseada em Sinais.
 */

let efeitoAtivo = null;

// ponytail: fila global e flag para agrupar e lotear execuções de efeitos síncronos na mesma microtask.
// Isso impede múltiplas reconciliações de VDOM desnecessárias e repaints repetitivos.
const filaEfeitos = new Set();
let agendamentoLote = false;

function agendarEfeito(ef) {
  filaEfeitos.add(ef);
  if (!agendamentoLote) {
    agendamentoLote = true;
    // queueMicrotask é um recurso nativo da plataforma web moderna
    queueMicrotask(() => {
      agendamentoLote = false;
      const lote = Array.from(filaEfeitos);
      filaEfeitos.clear();
      lote.forEach(ef => ef.run());
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
      paraDisparar.forEach(sub => agendarEfeito(sub));
    }
  };

  ler.set = definir; // ponytail: expõe atualizador direto no getter para binding bidirecional
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
  if (vno === null || vno === undefined || vno === false) {
    return document.createTextNode('');
  }

  if (typeof vno === 'function') {
    return criarNo(vno());
  }

  if (typeof vno === 'string' || typeof vno === 'number') {
    return document.createTextNode(vno);
  }

  if (!vno || typeof vno !== 'object') {
    return document.createTextNode(String(vno || ''));
  }

  if (typeof vno.tag === 'function') {
    const vnodeRendido = vno.tag(vno.props);
    vno._componenteInstancia = vnodeRendido;
    return criarNo(vnodeRendido);
  }

  const el = document.createElement(vno.tag || 'div');

  vno.el = el;

  atualizarAtributos(el, {}, vno.props);

  (vno.children || []).forEach(filho => {
    if (filho !== null && filho !== undefined && filho !== false) {
      el.appendChild(criarNo(filho));
    }
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
      if (chave === '_ligar' || chave === 'ligar') {
        const lerSinal = valor;
        efeito(() => {
          const val = typeof lerSinal === 'function' ? lerSinal() : lerSinal;
          if (el.type === 'checkbox') {
            el.checked = !!val;
          } else {
            el.value = val === null || val === undefined ? '' : val;
          }
        });
        const eventoTipo = el.type === 'checkbox' ? 'change' : 'input';
        el.addEventListener(eventoTipo, (e) => {
          const setter = lerSinal ? (lerSinal.set || lerSinal[1]) : null;
          if (setter) {
            setter(el.type === 'checkbox' ? e.target.checked : e.target.value);
          }
        });
        continue;
      }


      if (chave === 'innerHTML') {
        el.innerHTML = valor;
        continue;
      }

      if (chave.startsWith('ao')) {
        const evento = mapearEvento(chave);
        if (velhoValor) el.removeEventListener(evento, velhoValor);
        const listener = (e) => {
          if (el.tagName === 'A') {
            e.preventDefault();
          }
          valor(e);
        };
        el.addEventListener(evento, listener);
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
  if (!pai) return;

  if (typeof velhoVNo === 'function') velhoVNo = velhoVNo();
  if (typeof novoVNo === 'function') novoVNo = novoVNo();

  if (velhoVNo && typeof velhoVNo.tag === 'function') {
    velhoVNo = velhoVNo.tag(velhoVNo.props);
  }
  if (novoVNo && typeof novoVNo.tag === 'function') {
    novoVNo = novoVNo.tag(novoVNo.props);
  }

  const noFisico = pai.childNodes ? pai.childNodes[index] : null;

  if (!velhoVNo) {
    if (novoVNo) pai.appendChild(criarNo(novoVNo));
    return;
  }

  if (!novoVNo) {
    if (noFisico && pai.removeChild) pai.removeChild(noFisico);
    return;
  }

  if (mudou(velhoVNo, novoVNo)) {
    if (noFisico && pai.replaceChild) {
      pai.replaceChild(criarNo(novoVNo), noFisico);
    } else {
      pai.appendChild(criarNo(novoVNo));
    }
    return;
  }

  if (typeof novoVNo === 'object') {
    if (noFisico && noFisico.nodeType === 1) {
      atualizarAtributos(noFisico, (velhoVNo && velhoVNo.props) || {}, novoVNo.props || {});
      novoVNo.el = noFisico;
    }

    const velhosFilhos = (velhoVNo && velhoVNo.children) || [];
    const novosFilhos = (novoVNo && novoVNo.children) || [];

    const temChaves = novosFilhos.some(f => f && f.props && (f.props.chave !== undefined || f.props.key !== undefined));

    if (temChaves && noFisico) {
      const mapaVelhos = new Map();
      velhosFilhos.forEach((vno, idx) => {
        const k = vno && vno.props ? (vno.props.chave ?? vno.props.key) : idx;
        const noChild = noFisico.childNodes ? noFisico.childNodes[idx] : null;
        if (k !== undefined && noChild) {
          mapaVelhos.set(k, { vno, el: noChild });
        }
      });

      novosFilhos.forEach((novoChildVNo, i) => {
        const k = novoChildVNo && novoChildVNo.props ? (novoChildVNo.props.chave ?? novoChildVNo.props.key) : i;
        const correspondencia = mapaVelhos.get(k);

        if (correspondencia) {
          reconciliar(noFisico, correspondencia.vno, novoChildVNo, i);
          mapaVelhos.delete(k);
        } else {
          const novoEl = criarNo(novoChildVNo);
          if (noFisico.childNodes && noFisico.childNodes[i]) {
            noFisico.insertBefore(novoEl, noFisico.childNodes[i]);
          } else {
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
      const max = Math.max(velhosFilhos.length, novosFilhos.length);
      for (let i = 0; i < max; i++) {
        reconciliar(noFisico, velhosFilhos[i], novosFilhos[i], i);
      }
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
 * Suporta auto-hidratação transparente caso o elemento já contenha nós renderizados por SSR.
 * @param {Function} appComponente - Componente raiz
 * @param {HTMLElement} elementoAlvo - Elemento container no DOM
 */
export function montar(appComponente, elementoAlvo) {
  let velhoVNo = null;

  if (typeof elementoAlvo === 'string') {
    elementoAlvo = document.getElementById(elementoAlvo);
  }

  if (elementoAlvo) {
    elementoAlvo.innerHTML = '';
  }

  const obterVNo = typeof appComponente === 'function' ? appComponente : () => appComponente;

  efeito(() => {
    const novoVNo = obterVNo();
    reconciliar(elementoAlvo, velhoVNo, novoVNo);
    velhoVNo = novoVNo;
  });
}



function vincularNos(elFisico, vno) {
  if (!vno || !elFisico) return;
  vno.el = elFisico;

  if (typeof vno === 'object') {
    if (typeof vno.tag === 'function') {
      const vnodeRendido = vno.tag(vno.props);
      vno._componenteInstancia = vnodeRendido;
      vincularNos(elFisico, vnodeRendido);
      return;
    }

    for (const chave in vno.props) {
      if (chave.startsWith('ao')) {
        const evento = mapearEvento(chave);
        elFisico.addEventListener(evento, vno.props[chave]);
      }
    }

    const filhosFisicos = elFisico.childNodes;
    const max = Math.min(filhosFisicos.length, vno.children ? vno.children.length : 0);
    for (let i = 0; i < max; i++) {
      vincularNos(filhosFisicos[i], vno.children[i]);
    }
  }
}

// ============================================================================
// 3. ROTEAMENTO SPA BASEADO EM SINAIS (HISTORY API)
// ============================================================================

const [urlAtiva, setUrlAtiva] = sinal(typeof window !== 'undefined' ? window.location.pathname : '/');

/**
 * Navega dinamicamente para uma nova rota do SPA sem recarregar a página.
 * @param {string} destino - O caminho de URL de destino (ex: "/sobre")
 */
export function navegar(destino) {
  if (typeof window !== 'undefined' && window.location.pathname !== destino) {
    window.history.pushState({}, '', destino);
    setUrlAtiva(destino);
  }
}

if (typeof window !== 'undefined') {
  window.addEventListener('popstate', () => {
    setUrlAtiva(window.location.pathname);
  });
}

/**
 * Cria um componente funcional de Roteador baseado em Sinais.
 * @param {Object} rotas - Mapeamento de caminhos para componentes funcionais
 * @returns {Function} Componente do roteador
 */
export function roteador(rotas) {
  return () => {
    const path = urlAtiva();
    let componente = rotas[path];
    let parametros = {};

    if (!componente) {
      for (const rotaDef in rotas) {
        if (!rotaDef.includes(':')) continue;
        const partesDef = rotaDef.split('/').filter(Boolean);
        const partesPath = path.split('/').filter(Boolean);

        if (partesDef.length === partesPath.length) {
          let corresponde = true;
          const paramsTemp = {};
          for (let i = 0; i < partesDef.length; i++) {
            if (partesDef[i].startsWith(':')) {
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

    componente = componente || rotas['/404'] || (() => h('div', {}, '404 - Página Não Encontrada'));
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
  if (typeof window !== 'undefined' && typeof localStorage !== 'undefined') {
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
    if (typeof window !== 'undefined' && typeof localStorage !== 'undefined') {
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
      value: valores()[nome] || '',
      onInput: (e) => {
        const novos = { ...valores(), [nome]: e.target.value };
        setValores(novos);
        if (config.validar) {
          setErros(config.validar(novos) || {});
        }
      }
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
      setErro(err.message || 'Erro ao carregar consulta');
    } finally {
      setCarregando(false);
    }
  }

  efeito(() => {
    buscar();
  });

  return { dados, carregando, erro, mutarOtimista: (novo) => setDados(novo), recarregar: buscar };
}

/**
 * Componente de animações e transições nativas (Animacao).
 */
export function Animacao(props) {
  const tipo = props.tipo || 'fade';
  const duracao = props.duracao || '300ms';
  return h('div', {
    estilo: {
      transition: `all ${duracao} ease-in-out`,
      animation: `${tipo} ${duracao}`
    }
  }, props.children);
}

/**
 * Componente de imagem otimizada com lazy loading (ImagemWeb).
 */
export function ImagemWeb(props) {
  return h('img', {
    src: props.caminho || props.src,
    alt: props.alt || '',
    loading: 'lazy',
    width: props.largura,
    height: props.altura,
    classe: props.classe || ''
  });
}

/**
 * Carregador assíncrono de Micro-Frontends (moduloRemoto).
 */
export function moduloRemoto(url, nomeComponente) {
  return function ComponenteRemoto(props) {
    const [comp, setComp] = sinal(null);
    efeito(() => {
      import(/* webpackIgnore: true */ url).then(mod => {
        setComp(() => mod[nomeComponente] || mod.default);
      });
    });
    return () => {
      const C = comp();
      return C ? h(C, props) : h('div', { classe: 'carregando-remoto' }, 'Carregando Módulo Remoto...');
    };
  };
}

/**
 * Gestor de Tema Escuro/Claro (usarTema).
 */
export function usarTema(chave = 'harpia_tema', padrao = 'sistema') {
  const preferEscuro = window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches;
  const salvo = localStorage.getItem(chave);
  const inicial = salvo || (padrao === 'sistema' ? (preferEscuro ? 'escuro' : 'claro') : padrao);
  const [tema, setTema] = sinal(inicial);

  efeito(() => {
    const atual = tema();
    localStorage.setItem(chave, atual);
    if (atual === 'escuro') {
      document.documentElement.classList.add('escuro');
    } else {
      document.documentElement.classList.remove('escuro');
    }
  });

  const alternar = () => setTema(tema() === 'escuro' ? 'claro' : 'escuro');
  return [tema, alternar, setTema];
}

/**
 * Componente Portal para renderização direta no document.body (Portal).
 */
export function Portal(props) {
  const el = document.createElement('div');
  el.className = 'harpia-portal';
  efeito(() => {
    document.body.appendChild(el);
    montar(h('div', {}, props.children), el);
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
      if (e.dataTransfer) e.dataTransfer.setData('text/plain', JSON.stringify(item));
    },
    aoSobrepor: (e) => e.preventDefault(),
    aoSoltar: (itemAlvo) => (e) => {
      if (e.preventDefault) e.preventDefault();
      if (aoSoltar && itemArrastado) {
        aoSoltar(itemArrastado, itemAlvo);
      }
    }
  };
}

/**
 * Barra de progresso visual no topo do navegador (BarraDeProgresso).
 */
export function BarraDeProgresso(props) {
  const [carregando, setCarregando] = sinal(false);
  return h('div', {
    estilo: {
      position: 'fixed',
      top: '0',
      left: '0',
      height: '3px',
      width: carregando() ? '100%' : '0%',
      backgroundColor: props.cor || '#3b82f6',
      transition: 'width 300ms ease-in-out',
      zIndex: '99999'
    }
  });
}

/**
 * Sistema de notificações tostadas reativas (usarNotificacao).
 */
const [notificacoes, setNotificacoes] = sinal([]);

export function usarNotificacao() {
  function adicionar(mensagem, tipo = 'info', duracao = 3000) {
    const id = Date.now();
    const item = { id, mensagem, tipo };
    setNotificacoes([...notificacoes(), item]);
    setTimeout(() => {
      setNotificacoes(notificacoes().filter(n => n.id !== id));
    }, duracao);
  }

  return {
    sucesso: (msg) => adicionar(msg, 'sucesso'),
    erro: (msg) => adicionar(msg, 'erro'),
    info: (msg) => adicionar(msg, 'info'),
    avisar: (msg) => adicionar(msg, 'aviso'),
    lista: notificacoes
  };
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
  return h('div', { estilo: { display: 'contents' } }, res);
}

/**
 * Injeta/recupera um serviço provido por um componente Provedor superior.
 * @param {string} chave - Identificador do serviço
 */
export function injetar(chave) {
  return mapaContextos.get(chave);
}

/**
 * Fronteira de Erro nativa para capturar falhas em componentes filhos secundários.
 */
export function FronteiraDeErro(props) {
  const [erro, setErro] = sinal(null);
  try {
    if (erro()) {
      return props.fallback || h('p', {}, 'Erro ao carregar componente.');
    }
    return props.children || null;
  } catch (e) {
    setErro(e);
    return props.fallback || h('p', {}, 'Erro ao carregar componente.');
  }
}

/**
 * Componente nativo de navegação SPA sem recarregamento de página.
 */
export function Link(props) {
  const { para, children, ...restanteProps } = props || {};
  const aoClicarOriginal = restanteProps.aoClicar;
  restanteProps.aoClicar = (e) => {
    if (e && e.preventDefault) e.preventDefault();
    if (para) navegar(para);
    if (aoClicarOriginal) aoClicarOriginal(e);
  };
  restanteProps.href = para || '#';
  return h('a', restanteProps, children);
}

/**
 * Componente nativo de Suspense/Aguardar para dados assíncronos e recursos.
 */
export function Aguardar(props) {
  const { recurso: res, carregando, erro: renderErro, children } = props || {};
  if (!res) return typeof children === 'function' ? children(null) : children;

  if (typeof res.carregando === 'function' && res.carregando()) {
    return carregando || h('p', { class: 'carregando' }, '⌛ Carregando...');
  }

  if (typeof res.erro === 'function' && res.erro()) {
    const err = res.erro();
    return typeof renderErro === 'function'
      ? renderErro(err)
      : (renderErro || h('p', { class: 'erro' }, `❌ Erro: ${err}`));
  }

  const dados = typeof res.dados === 'function' ? res.dados() : res.dados;
  if (typeof children === 'function') {
    return children(dados);
  }
  return children;
}

/**
 * Construto nativo de Pattern Matching de UI (Escolha/Caso/Padrao).
 */
export function Escolha(props) {
  const valorAlvo = typeof props.valor === 'function' ? props.valor() : props.valor;
  const filhos = Array.isArray(props.children) ? props.children : [props.children];
  let casoCorrespondente = null;
  let casoPadrao = null;

  for (const filho of filhos) {
    if (!filho) continue;
    if (filho.tag === Caso || (filho.props && filho.props.valor !== undefined)) {
      if (filho.props.valor === valorAlvo) {
        casoCorrespondente = filho;
        break;
      }
    } else if (filho.tag === Padrao || (filho.props && filho.props.isPadrao)) {
      casoPadrao = filho;
    }
  }

  const escolhido = casoCorrespondente || casoPadrao;
  return escolhido ? (escolhido.children || escolhido) : null;
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
    const itensVisiveis = (props.itens || []).slice(inicio(), inicio() + visiveis);
    const espacoSuperior = inicio() * alturaLinha;
    const espacoInferior = Math.max(0, (total - inicio() - itensVisiveis.length) * alturaLinha);

    return h('div', {
      aoScroll: aoRolar,
      estilo: { height: `${alturaContainer}px`, overflowY: 'auto', position: 'relative' }
    },
      h('div', { estilo: { height: `${espacoSuperior}px` } }),
      ...itensVisiveis.map((item, idx) => props.children[0](item, inicio() + idx)),
      h('div', { estilo: { height: `${espacoInferior}px` } })
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
    const dadosFiltrados = dadosOriginais.filter(item => {
      const texto = filtro().toLowerCase();
      if (!texto) return true;
      return colunas.some(col => String(item[col] || "").toLowerCase().includes(texto));
    });

    const totalPaginas = Math.ceil(dadosFiltrados.length / limite) || 1;
    const inicio = (pagina() - 1) * limite;
    const dadosPaginados = dadosFiltrados.slice(inicio, inicio + limite);

    return h('div', { classe: 'grade-dados-container' },
      h('input', {
        _ligar: filtro,
        placeholder: 'Pesquisar...',
        classe: 'grade-dados-pesquisa'
      }),
      h('table', { classe: 'grade-dados-tabela' },
        h('thead', {},
          h('tr', {}, ...colunas.map(col => h('th', {}, col)))
        ),
        h('tbody', {},
          ...dadosPaginados.map(item =>
            h('tr', {}, ...colunas.map(col => h('td', {}, String(item[col] || ""))))
          )
        )
      ),
      h('div', { classe: 'grade-dados-paginacao' },
        h('button', {
          aoClicar: () => setPagina(Math.max(1, pagina() - 1)),
          disabled: pagina() === 1
        }, 'Anterior'),
        h('span', {}, ` Página ${pagina()} de ${totalPaginas} `),
        h('button', {
          aoClicar: () => setPagina(Math.min(totalPaginas, pagina() + 1)),
          disabled: pagina() === totalPaginas
        }, 'Próxima')
      )
    );
  };
}

/**
 * Carregador preguiçoso para Code Splitting e Lazy Loading de componentes.
 */
export function preguicoso(importarComponente) {
  return function ComponentePreguicoso(props) {
    const [comp, setComp] = sinal(null);
    importarComponente().then(Modulo => {
      setComp(() => Modulo.default || Modulo);
    });
    return () => {
      const C = comp();
      return C ? h(C, props) : h('div', { classe: 'carregando-lazy' }, 'Carregando...');
    };
  };
}

/**
 * Componente Suspense reativo simples.
 */
export function Suspense(props) {
  return () => {
    const pronto = typeof props.pronto === 'function' ? props.pronto() : props.pronto;
    return pronto ? props.children : (props.fallback || h('div', {}, 'Carregando...'));
  };
}

/**
 * Realiza requisições HTTP assíncronas (fetch) e retorna o resultado.
 */
export async function requisitar(metodo, url, dados = null, cabecalhos = {}) {
  const opcoes = {
    method: metodo,
    headers: {
      'Content-Type': 'application/json',
      ...cabecalhos
    }
  };
  if (dados) {
    opcoes.body = typeof dados === 'string' ? dados : JSON.stringify(dados);
  }
  const resposta = await fetch(url, opcoes);
  if (!resposta.ok) {
    throw new Error(`Erro na requisição: ${resposta.status} ${resposta.statusText}`);
  }
  const contentType = resposta.headers.get('content-type');
  if (contentType && contentType.includes('application/json')) {
    return resposta.json();
  }
  return resposta.text();
}

// ponytail: vincula as primitivas ao escopo global do navegador para compatibilidade de transpilação
if (typeof window !== 'undefined') {
  Object.assign(window, {
    Nulo: null,
    Verdadeiro: true,
    Falso: false,

    sinal,
    sinalDebounce,
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
    usarNotificacao
  });
}





