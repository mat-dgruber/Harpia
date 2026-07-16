# Diretrizes de Marca: Harpia 🦅

Este documento define a identidade visual, conceitual e o tom de voz da linguagem **Harpia** (anteriormente conhecida como *Harpia*).

---

## 1. Essência da Marca

A **Harpia** (gavião-real) é a maior águia das Américas e o predador aéreo mais imponente das florestas tropicais brasileiras. Ela representa **soberania, precisão cirúrgica, força e foco absoluto**.

### Pilares da Identidade

* **Precisão e Foco:** Como o olhar de uma águia antes do bote. O código em Harpia deve ser direto, eliminando redundâncias (filosofia YAGNI/Ponytail).
* **Soberania Nativa:** Uma tecnologia brasileira robusta e independente, focada em resolver o desenvolvimento moderno (monólitos, SPA e APIs) em nosso próprio idioma.
* **Arquitetura Elegante:** A imponência física da ave se traduz em Clean Architecture e DDD estruturados nativamente pela linguagem.

---

## 2. Identidade Visual & Design

A paleta de cores e o design devem refletir o habitat natural da Harpia (a copa das árvores da Amazônia) e as cores reais da ave (tons de cinza, preto, branco e detalhes sutis de amarelo nos olhos e garras).

### 2.1 Paleta de Cores (Cromia)

Utilizaremos uma paleta moderna e premium com foco em *Dark Mode* nativo, substituindo os cinzas genéricos por tons que remetem à fauna e flora tropicais (folhagens amazônicas profundas, terra úmida e o azul-celeste de transição).

| Função                      | Cor                          | Hexadecimal | Conceito Visual (Brasilidade)                                                                  |
| :---------------------------- | :--------------------------- | :---------- | :--------------------------------------------------------------------------------------------- |
| **Fundo Primário**     | Sombra da Floresta           | `#0A0F0D` | Preto profundo com um subtom verde-musgo extremamente sutil                                    |
| **Fundo Secundário**   | Rio Profundo (Cinza Azulado) | `#171E26` | Um slate-blue (cinza-azulado) fechado, trazendo a sobriedade da água dos rios amazônicos     |
| **Destaque AEO / Foco** | Olho da Harpia               | `#F2A900` | Amarelo-ouro vivo e quente, a cor dos olhos da ave e da riqueza mineral                        |
| **Acento Secundário**  | Terracota / Argila           | `#D45D34` | Um laranja-argila/tijolo queimado, trazendo calor da nossa terra às interações secundárias |
| **Acento de Sucesso**   | Broto de Ipê (Verde)        | `#00A86B` | Um verde folha vibrante para logs de sucesso e compilações bem-sucedidas                     |
| **Texto Principal**     | Penagem Branca               | `#F3F6F4` | Branco suave com subtom cinza gesso, reduzindo o cansaço visual                               |
| **Texto de Apoio**      | Névoa da Manhã             | `#8C9B9E` | Cinza levemente esverdeado/azulado para comentários e metadados                               |

### 2.2 Texturas e Elementos Gráficos

Para reforçar a identidade brasileira de forma sutil e elegante:

* **Geometria Indígena Modernizada:** Grafismos geométricos inspirados nas artes visuais dos povos nativos (Marajoara ou Yanomami), mas vetorizados com traços ultra-finos de opacidade baixa (10-15%) em fundos escuros do site e slides.
* **Gradientes de Transição:** Uso de gradientes que simulam o nascer do sol na mata (do `#171E26` transicionando suavemente para detalhes em `#D45D34` e `#F2A900`) (Veja a referência oficial em [fundo_harpia.jpg](./assets/fundo_harpia.jpg)).

O logotipo da Harpia deve ser minimalista, geométrico e agressivo na medida certa.

* **Símbolo Principal:** A silhueta estilizada da cabeça da Harpia vista de perfil, destacando a sua **crista bipartida de penas** levantada (símbolo de atenção e prontidão) e o bico curvado imponente (Veja a referência conceitual em [logo_harpia.jpg](./assets/logo_harpia.jpg) e a versão refinada com proporções anatômicas realistas em [logo_harpia_refinado.jpg](./assets/logo_harpia_refinado.jpg)).
* **Aplicação:**
  * O símbolo deve funcionar perfeitamente em cor única (monocromático) para o CLI e ícones de arquivo do sistema.
  * Em interfaces ricas (Web), o olho da ave é destacado no amarelo dourado vibrante (`#FFC72C`), agindo como o foco central.

---

## 3. Tipografia

Para manter o aspecto premium, limpo e legível do ecossistema:

* **Títulos e Landing Page:** **Outfit** ou **Inter** (Google Fonts). Geométricas, modernas e altamente legíveis.
* **Código e Terminal (Mono):** **JetBrains Mono** ou **Fira Code**. Perfeitas para destacar a sintaxe em português com excelente espaçamento e suporte a ligaduras.

---

## 4. Tom de Voz e Escrita

A Harpia fala com **autoridade amigável**. Como uma mentora de desenvolvimento:

* **Confiante, mas não arrogante:** Explica conceitos complexos de Clean Architecture e DDD de forma simples e pragmática.
* **Identidade Própria:** Evita modismos exagerados de tecnologia (jargões em inglês desnecessários) e prefere usar termos claros em português, sem soar arcaico.
* **Foco na Solução:** As mensagens de erro (`HRP-XXXX`) nunca devem apenas apontar o erro; elas devem ser instrutivas, sugerindo o caminho de correção (explicadas pela IA local integrada).

---

## 5. Diretrizes para a Comunidade & Ecossistema

* **Licenciamento:** Mantido sob a licença MIT, incentivando forks, contribuições e uso comercial livre.
* **O Nome:** Sempre grafado como **Harpia** (com H maiúsculo), nunca *harpia-lang* ou *HarpiaJS*.
* **Extensão de Arquivo Padrão:** Os arquivos de código-fonte da linguagem utilizam oficialmente a extensão **`.hrp`**.

---

## 6. Diretrizes de Design no Código e Ferramental

Abaixo estão as especificações visuais de como a identidade da Harpia se traduz em código escrito, terminal (CLI) e editores de texto.

### 6.1 Tema de Sintaxe Oficial ("Harpia Dark")

O esquema de cores para IDEs (como VS Code) e renderizadores de código deve usar a paleta nativa para criar um ambiente de baixo cansaço visual e alta legibilidade:

| Elemento de Código                                                       | Cor                | Hexadecimal | Conceito                              |
| :------------------------------------------------------------------------ | :----------------- | :---------- | :------------------------------------ |
| **Fundo Geral**                                                     | Sombra da Floresta | `#0A0F0D` | Visual profundo e confortável        |
| **Palavras-chave de Controle** (`se`, `retornar`, `exportar`) | Terracota          | `#D45D34` | Destaque quente e estrutural          |
| **Definições e Tipos** (`funcao`, `classe`, `constante`)    | Azul Celeste       | `#38B6FF` | Tonalidades do horizonte sobre a copa |
| **Chamadas de Funções** (`imprimir`, `sinal`)                 | Olho da Harpia     | `#F2A900` | Foco de ação                        |
| **Strings / Literais de Texto**                                     | Verde Broto        | `#00A86B` | Conteúdo orgânico e seguro          |
| **Comentários de Código**                                         | Névoa da Manhã   | `#8C9B9E` | Código inativo e legível            |

### 6.2 Estrutura Visual de Erros (`HRP-XXXX`)

Erros em tempo de compilação ou execução devem ser tratados como uma oportunidade de aprendizado, estruturados de forma limpa e imponente no terminal:

```ansi
[HRP-1002] Erro de Sintaxe: Chave não fechada
 ──> web/componentes/botao.hrp:12:5
  │
11│  estilo Botao {
12│      largura: "100%"
  │                     ^ esperado '}' para fechar o bloco estilo
  │
 Dica da Harpia: Você abriu um bloco de estilo na linha 11, mas esqueceu 
                 de fechá-lo. Adicione um '}' na linha 13.
```

* **Borda Lateral:** A linha vertical (`│`) guia o olho do programador, destacando o contexto da linha de código.
* **Identificador Amigável:** O código `HRP-XXXX` facilita buscas na documentação ou ajuda com a IA integrada (`harpia explicar HRP-1002`).

### 6.3 Assinatura do CLI (Assinatura de Alta Resolução)
Quando o comando `harpia` é executado sem argumentos, a tela de boas-vindas deve exibir por padrão a assinatura de alta resolução em blocos de caracteres Unicode (`utf-8`). Ela forma a silhueta geométrica e frontal de uma harpia (destacando a crista dupla ao topo e o bico afunilado):

```text
      ▲     ▲
     ▀█▀   ▀█▀       HARPIA (v1.0.0)
    ▄█████████▄      Linguagem reativa para Clean Architecture
   ▐████▀ ▀████▌    
    ▀████▄████▀      > Digite 'ajuda' para comandos ou 'ia' para suporte local.
      ▀█████▀        
        ▀█▀          
```
*(O design frontal destaca a crista dupla no topo e o bico imponente afunilado ao centro).*


### 6.4 Metáforas da Biblioteca Padrão (Stdlib)

Os módulos nativos incorporados na linguagem devem usar nomes que remetam a elementos naturais e estruturas do ecossistema brasileiro:

* **`raiz`** (Em vez de `core` ou `kernel`): Ponto de entrada e utilitários principais do sistema.
* **`copa`** (Em vez de `view` ou `layout`): Recursos voltados à camada visual (Frontend SPA e estilo).
* **`ninho`** (Em vez de `sandbox` ou `environment`): Gerenciamento de processos isolados e variáveis de ambiente.
* **`correnteza`** (Em vez de `stream` ou `pipes`): Processamento de fluxos de dados e canais de comunicação.
