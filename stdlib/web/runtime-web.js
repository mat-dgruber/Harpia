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
  if (typeof vno === 'function') {
    return criarNo(vno());
  }

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
      if (chave === '_ligar') {
        const lerSinal = valor;
        efeito(() => {
          const val = lerSinal();
          if (el.type === 'checkbox') {
            el.checked = !!val;
          } else {
            el.value = val === null || val === undefined ? '' : val;
          }
        });
        const eventoTipo = el.type === 'checkbox' ? 'change' : 'input';
        el.addEventListener(eventoTipo, (e) => {
          const setter = lerSinal.set || lerSinal[1];
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
  if (typeof velhoVNo === 'function') velhoVNo = velhoVNo();
  if (typeof novoVNo === 'function') novoVNo = novoVNo();

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

    const velhosFilhos = velhoVNo.children || [];
    const novosFilhos = novoVNo.children || [];
    const max = Math.max(velhosFilhos.length, novosFilhos.length);

    const chavesVelhas = {};
    velhosFilhos.forEach((c, idx) => {
      if (c && c.props && c.props.chave !== undefined) {
        chavesVelhas[c.props.chave] = { no: c, idx };
      }
    });

    for (let i = 0; i < max; i++) {
      let velhoF = velhosFilhos[i];
      let novoF = novosFilhos[i];

      if (novoF && novoF.props && novoF.props.chave !== undefined) {
        const achado = chavesVelhas[novoF.props.chave];
        if (achado) {
          velhoF = achado.no;
          const noFis = noFisico.childNodes[achado.idx];
          if (noFis && noFisico.childNodes[i] !== noFis) {
            noFisico.insertBefore(noFis, noFisico.childNodes[i]);
          }
        }
      }
      reconciliar(noFisico, velhoF, novoF, i);
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

  if (elementoAlvo && elementoAlvo.childNodes.length > 0) {
    const vnoInicial = typeof appComponente === 'function' ? appComponente() : appComponente;
    vincularNos(elementoAlvo.firstElementChild || elementoAlvo, vnoInicial);
    velhoVNo = vnoInicial;
  }

  efeito(() => {
    const novoVNo = typeof appComponente === 'function' ? appComponente() : appComponente;
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
    const max = Math.min(filhosFisicos.length, vno.children.length);
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
    const componente = rotas[path] || rotas['/404'] || (() => h('div', {}, '404 - Página Não Encontrada'));
    return h(componente, {});
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
  ler.carregando = carregando;
  ler.erro = erro;
  ler.ok = () => !carregando() && !erro();
  ler.recarregar = executar;
  return [ler];
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

