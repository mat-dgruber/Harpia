# Diretrizes de Marca: Harpia 🦅

Este documento define a identidade visual, conceitual e o tom de voz da linguagem **Harpia** (anteriormente conhecida como _Harpia_).

---

## 1. Essência da Marca

A **Harpia** (gavião-real) é a maior águia das Américas e o predador aéreo mais imponente das florestas tropicais brasileiras. Ela representa **soberania, precisão cirúrgica, força e foco absoluto**.

### Pilares da Identidade

- **Precisão e Foco:** Como o olhar de uma águia antes do bote. O código em Harpia deve ser direto, eliminando redundâncias (filosofia YAGNI/Ponytail).
- **Soberania Nativa:** Uma tecnologia brasileira robusta e independente, focada em resolver o desenvolvimento moderno (monólitos, SPA e APIs) em nosso próprio idioma.
- **Arquitetura Elegante:** A imponência física da ave se traduz em Clean Architecture e DDD estruturados nativamente pela linguagem.

---

## 2. Identidade Visual & Design

A paleta de cores e o design devem refletir o habitat natural da Harpia (a copa das árvores da Amazônia) e as cores reais da ave (tons de cinza, preto, branco e detalhes sutis de amarelo nos olhos e garras).

### 2.1 Paleta de Cores (Cromia)

Utilizaremos uma paleta moderna e premium com foco em _Dark Mode_ nativo, substituindo os cinzas genéricos por tons que remetem à fauna e flora tropicais (folhagens amazônicas profundas, terra úmida e o azul-celeste de transição).

| Função                  | Cor                          | Hexadecimal | Conceito Visual (Brasilidade)                                                              |
| :---------------------- | :--------------------------- | :---------- | :----------------------------------------------------------------------------------------- |
| **Fundo Primário**      | Sombra da Floresta           | `#0A0F0D`   | Preto profundo com um subtom verde-musgo extremamente sutil                                |
| **Fundo Secundário**    | Rio Profundo (Cinza Azulado) | `#171E26`   | Um slate-blue (cinza-azulado) fechado, trazendo a sobriedade da água dos rios amazônicos   |
| **Destaque AEO / Foco** | Olho da Harpia               | `#F2A900`   | Amarelo-ouro vivo e quente, a cor dos olhos da ave e da riqueza mineral                    |
| **Acento Secundário**   | Terracota / Argila           | `#D45D34`   | Um laranja-argila/tijolo queimado, trazendo calor da nossa terra às interações secundárias |
| **Acento de Sucesso**   | Broto de Ipê (Verde)         | `#00A86B`   | Um verde folha vibrante para logs de sucesso e compilações bem-sucedidas                   |
| **Texto Principal**     | Penagem Branca               | `#F3F6F4`   | Branco suave com subtom cinza gesso, reduzindo o cansaço visual                            |
| **Texto de Apoio**      | Névoa da Manhã               | `#8C9B9E`   | Cinza levemente esverdeado/azulado para comentários e metadados                            |

### 2.2 Texturas e Elementos Gráficos

Para reforçar a identidade brasileira de forma sutil e elegante:

- **Geometria Indígena Modernizada:** Grafismos geométricos inspirados nas artes visuais dos povos nativos (Marajoara ou Yanomami), mas vetorizados com traços ultra-finos de opacidade baixa (10-15%) em fundos escuros do site e slides.
- **Gradientes de Transição:** Uso de gradientes que simulam o nascer do sol na mata (do `#171E26` transicionando suavemente para detalhes em `#D45D34` e `#F2A900`) (Veja a referência oficial em [fundo_harpia.jpg](./assets/fundo_harpia.jpg)).

O logotipo da Harpia deve ser minimalista, geométrico e agressivo na medida certa.

- **Símbolo Principal:** A silhueta estilizada da cabeça da Harpia vista de perfil, destacando a sua **crista bipartida de penas** levantada (símbolo de atenção e prontidão) e o bico curvado imponente (Veja a referência conceitual em [logo_harpia.jpg](./assets/logo_harpia.jpg) e a versão refinada com proporções anatômicas realistas em [logo_harpia_refinado.jpg](./assets/logo_harpia_refinado.jpg)).
- **Aplicação:**
  - O símbolo deve funcionar perfeitamente em cor única (monocromático) para o CLI e ícones de arquivo do sistema.
  - Em interfaces ricas (Web), o olho da ave é destacado no amarelo dourado vibrante (`#FFC72C`), agindo como o foco central.

---

## 3. Tipografia

Para manter o aspecto premium, limpo e legível do ecossistema:

- **Títulos e Landing Page:** **Outfit** ou **Inter** (Google Fonts). Geométricas, modernas e altamente legíveis.
- **Código e Terminal (Mono):** **JetBrains Mono** ou **Fira Code**. Perfeitas para destacar a sintaxe em português com excelente espaçamento e suporte a ligaduras.

---

## 4. Tom de Voz e Escrita

A Harpia fala com **autoridade amigável**. Como uma mentora de desenvolvimento:

- **Confiante, mas não arrogante:** Explica conceitos complexos de Clean Architecture e DDD de forma simples e pragmática.
- **Identidade Própria:** Evita modismos exagerados de tecnologia (jargões em inglês desnecessários) e prefere usar termos claros em português, sem soar arcaico.
- **Foco na Solução:** As mensagens de erro (`HRP-XXXX`) nunca devem apenas apontar o erro; elas devem ser instrutivas, sugerindo o caminho de correção (explicadas pela IA local integrada).

---

## 5. Diretrizes para a Comunidade & Ecossistema

- **Licenciamento:** Mantido sob a licença MIT, incentivando forks, contribuições e uso comercial livre.
- **O Nome:** Sempre grafado como **Harpia** (com H maiúsculo), nunca _harpia-lang_ ou _HarpiaJS_.
- **Extensão de Arquivo Padrão:** Os arquivos de código-fonte da linguagem utilizam oficialmente a extensão **`.hrp`**.

---

## 6. Diretrizes de Design no Código e Ferramental

Abaixo estão as especificações visuais de como a identidade da Harpia se traduz em código escrito, terminal (CLI) e editores de texto.

### 6.1 Tema de Sintaxe Oficial ("Harpia Dark")

O esquema de cores para IDEs (como VS Code) e renderizadores de código deve usar a paleta nativa para criar um ambiente de baixo cansaço visual e alta legibilidade:

| Elemento de Código                                            | Cor                | Hexadecimal | Conceito                              |
| :------------------------------------------------------------ | :----------------- | :---------- | :------------------------------------ |
| **Fundo Geral**                                               | Sombra da Floresta | `#0A0F0D`   | Visual profundo e confortável         |
| **Palavras-chave de Controle** (`se`, `retornar`, `exportar`) | Terracota          | `#D45D34`   | Destaque quente e estrutural          |
| **Definições e Tipos** (`funcao`, `classe`, `constante`)      | Azul Celeste       | `#38B6FF`   | Tonalidades do horizonte sobre a copa |
| **Chamadas de Funções** (`imprimir`, `sinal`)                 | Olho da Harpia     | `#F2A900`   | Foco de ação                          |
| **Strings / Literais de Texto**                               | Verde Broto        | `#00A86B`   | Conteúdo orgânico e seguro            |
| **Comentários de Código**                                     | Névoa da Manhã     | `#8C9B9E`   | Código inativo e legível              |

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

- **Borda Lateral:** A linha vertical (`│`) guia o olho do programador, destacando o contexto da linha de código.
- **Identificador Amigável:** O código `HRP-XXXX` facilita buscas na documentação ou ajuda com a IA integrada (`harpia explicar HRP-1002`).

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

_(O design frontal destaca a crista dupla no topo e o bico imponente afunilado ao centro)._

### 6.4 Metáforas da Biblioteca Padrão (Stdlib)

Os módulos nativos incorporados na linguagem devem usar nomes que remetam a elementos naturais e estruturas do ecossistema brasileiro:

- **`raiz`** (Em vez de `core` ou `kernel`): Ponto de entrada e utilitários principais do sistema.
- **`copa`** (Em vez de `view` ou `layout`): Recursos voltados à camada visual (Frontend SPA e estilo).
- **`ninho`** (Em vez de `sandbox` ou `environment`): Gerenciamento de processos isolados e variáveis de ambiente.
- **`correnteza`** (Em vez de `stream` ou `pipes`): Processamento de fluxos de dados e canais de comunicação.

---

## 7. Modo Claro (Light Mode)

O modo claro mantém a identidade da Harpia em ambientes com fundo claro, priorizando legibilidade e contraste WCAG AA+. A paleta inverte a hierarquia de fundo/texto mas preserva as cores de destaque (Olho da Harpia, Terracota, Broto de Ipê) para consistência entre temas.

### 7.1 Paleta de Cores — Modo Claro

| Função                  | Cor                | Hexadecimal | Nota                                                                      |
| :---------------------- | :----------------- | :---------- | :------------------------------------------------------------------------ |
| **Fundo Primário**      | Farinha de Trigo   | `#FAF5EE`   | Creme quente de farinha, inspirado na mesa brasileira — nunca branco puro |
| **Fundo Secundário**    | Palha Seca         | `#F1E8DA`   | Tonalidade palha dos ninhos da Harpia para cards e painéis laterais       |
| **Fundo Terciário**     | Pau-Brasil Claro   | `#E8DDD0`   | Madeira clara do Pau-Brasil para bordas e separadores                     |
| **Destaque AEO / Foco** | Olho da Harpia     | `#C48A00`   | Amarelo-ouro escuro (suficiente contraste sobre creme)                    |
| **Acento Secundário**   | Terracota Profunda | `#B84A2A`   | Versão escura da terracota para manter contraste                          |
| **Acento de Sucesso**   | Ipê Maduro         | `#1A7A4D`   | Verde escuro — broto de Ipê com bom contraste sobre creme                 |
| **Acento de Alerta**    | Tucunaré           | `#E07A2F`   | Laranja-avesso do peixe Tucunaré — aviso visível mas acolhedor            |
| **Texto Principal**     | Grauite            | `#1A2320`   | Quase-preto com subtom verde, fiel à identidade                           |
| **Texto de Apoio**      | Pedra              | `#5F6B65`   | Cinza médio para comentários e legendas                                   |
| **Borda Padrão**        | Areia              | `#C8BBA8`   | Borda com subtom bege-dourado — remete à areia de praia                   |

> **Regra de contraste:** A cor de destaque (`#C48A00`) atinge ratio 4.3:1 sobre Farinha de Trigo (`#FAF5EE`). Texto principal (`#1A2320`) atinge 14.2:1 — bem acima do AA (4.5:1).

### 7.2 Mapeamento: Dark → Light

| Escuro               | Claro                                 |
| :------------------- | :------------------------------------ |
| `bg-[#0A0F0D]`       | `bg-[#FAF5EE]` (Farinha de Trigo)     |
| `bg-[#171E26]`       | `bg-[#F1E8DA]` (Palha Seca)           |
| `border-slate-900`   | `border-[#C8BBA8]` (Areia)            |
| `text-[#F3F6F4]`     | `text-[#1A2320]` (Grauite)            |
| `text-[#8C9B9E]`     | `text-[#5F6B65]` (Pedra)              |
| `text-[#F2A900]`     | `text-[#C48A00]` (Olho da Harpia)     |
| `text-[#D45D34]`     | `text-[#B84A2A]` (Terracota Profunda) |
| `text-[#00A86B]`     | `text-[#1A7A4D]` (Ipê Maduro)         |
| `hover:bg-[#0A0F0D]` | `hover:bg-[#E8DDD0]` (Pau-Brasil)     |

### 7.3 Tema de Sintaxe — "Harpia Light"

| Elemento de Código                    | Cor                | Hexadecimal |
| :------------------------------------ | :----------------- | :---------- |
| **Fundo Geral**                       | Farinha de Trigo   | `#FAF5EE`   |
| **Palavras-chave** (`se`, `retornar`) | Terracota Profunda | `#B84A2A`   |
| **Definições e Tipos** (`funcao`)     | Azul Profundo      | `#1A6FB5`   |
| **Chamadas de Funções** (`imprimir`)  | Olho da Harpia     | `#C48A00`   |
| **Strings / Literais**                | Ipê Maduro         | `#1A7A4D`   |
| **Comentários**                       | Pedra              | `#5F6B65`   |

### 7.4 Diretrizes de Componentes — Modo Claro

- **Cards e painéis:** Fundo `#F1E8DA` (Palha Seca) com borda `#C8BBA8` (Areia). Sombra sutil `0 1px 3px rgba(0,0,0,0.06)`.
- **Inputs e textareas:** Fundo `#FAF5EE` (Farinha), borda `#C8BBA8`, focus ring `#C48A00` com opacidade 30%.
- **Botões primários:** Fundo `#C48A00`, texto `#FAF5EE` (contraste 4.3:1). Hover: escurecer 10%.
- **Botões secundários:** Fundo transparente, borda `#C8BBA8`, texto `#5F6B65`. Hover: fundo `#F1E8DA`.
- **Toasts e alerts:** Fundo `#FAF5EE` com borda lateral colorida (sucesso: `#1A7A4D`, erro: `#B84A2A`, aviso: `#E07A2F` Tucunaré).
- **Modais:** Fundo `#FAF5EE` com backdrop `rgba(0,0,0,0.3)`.
- **Badges de método:** Fundo transparente e texto colorido (ex: `text-[#1A6FB5]` para GET).
- **Sidebar:** Fundo `#F1E8DA`, borda direita `#C8BBA8`, links inativos em `#5F6B65`, link ativo com borda `#C48A00` e fundo `rgba(196,138,0,0.08)`.

---

## 7. Modo Claro ("Harpia Light")

A Harpia nasce com vocação para _Dark Mode_, mas respeita quem prefere programar à luz do dia. O **Harpia Light** traduz os mesmos pilares (precisão, soberania, elegância) para um ambiente diurno, mantendo a cromia brasileira — agora pensada como o amanhecer sobre a copa das árvores, a palha clara dos ninhos e a brancura da penagem ventral da ave.

### 7.1 Paleta de Cores — Harpia Light

Inspirada no amanhecer na copa da floresta e na penagem ventral clara da Harpia. Os mesmos papéis funcionais do tema escuro, com valores invertidos e dessaturados:

| Função                  | Cor                   | Hexadecimal | Conceito Visual (Brasilidade)                                                                   |
| :---------------------- | :-------------------- | :---------- | :---------------------------------------------------------------------------------------------- |
| **Fundo Primário**      | Penagem Clara         | `#FAF7F0`   | Off-white com subtom creme-areia, remetendo à luz filtrada pelas folhas                         |
| **Fundo Secundário**    | Ninho de Palha        | `#F1ECE2`   | Bege muito suave para cartões, modais e sidebars — a textura do ninho da Harpia                 |
| **Superfície Elevada**  | Papel Kraft           | `#FFFFFF`   | Branco puro para inputs, dropdowns e blocos em destaque                                         |
| **Destaque AEO / Foco** | Olho da Harpia        | `#C68A00`   | Mesmo amarelo-ouro, mas **-15% luminosidade** para manter contraste WCAG AA sobre fundos claros |
| **Acento Secundário**   | Terracota Suave       | `#B14A22`   | Laranja-argila menos saturado para não cansar em ambiente claro                                 |
| **Acento de Sucesso**   | Broto de Ipê (Verde)  | `#0F8F5C`   | Verde levemente mais escuro para contraste sobre fundo claro                                    |
| **Erro / Offline**      | Argila Queimada       | `#A8412B`   | Vermelho-tijolo coerente com o tema escuro                                                      |
| **Texto Principal**     | Casca de Imbuia       | `#1A1410`   | Marrom-pretado suave (não puro `#000`), mais quente que cinza, remetendo à madeira da floresta  |
| **Texto de Apoio**      | Somra da Copa         | `#6B6258`   | Marrom-acinzentado para comentários e metadados                                                 |
| **Bordas Sutis**        | Linha de Sombra Suave | `#E5DFD2`   | Bege-acinzentado para divisores (substitui o `slate-900` do tema escuro)                        |

> **Regra de Contraste:** Todo texto principal sobre fundo deve atingir no mínimo **WCAG AA (4.5:1)**. Componentes interativos (botão primário, badge de método) devem usar o par _fundo claro + texto escuro_ ou _fundo escuro + texto claro_, nunca mistura parcial.

### 7.2 Tema de Sintaxe — "Harpia Light"

Espelha os papéis da seção 6.1, recalibrados para fundos claros. O tema escuro continua sendo o oficial; o Light é a alternativa diurna:

| Elemento de Código                                            | Cor            | Hexadecimal | Conceito                                 |
| :------------------------------------------------------------ | :------------- | :---------- | :--------------------------------------- |
| **Fundo Geral**                                               | Penagem Clara  | `#FAF7F0`   | Base clara e confortável para o dia      |
| **Palavras-chave de Controle** (`se`, `retornar`, `exportar`) | Terracota      | `#B14A22`   | Destaque quente, mesma função do dark    |
| **Definições e Tipos** (`funcao`, `classe`, `constante`)      | Azul Celeste   | `#1E7AB8`   | Mais saturado para legibilidade em claro |
| **Chamadas de Funções** (`imprimir`, `sinal`)                 | Olho da Harpia | `#C68A00`   | Foco de ação, versão diurna              |
| **Strings / Literais de Texto**                               | Verde Broto    | `#0F8F5C`   | Conteúdo seguro e legível                |
| **Comentários de Código**                                     | Sombra da Copa | `#6B6258`   | Inativo, mas nunca ilegível              |
| **Números**                                                   | Indigo Açúcar  | `#5A4FCF`   | Tom frio para destacar valores numéricos |

### 7.3 Princípios de Aplicação do Light Mode

1. **Não inverter cegamente.** Não basta trocar preto por branco. A Harpia usa marrom-pretado (`#1A1410`) como texto principal — não `#000` puro — para preservar a temperatura visual da marca.
2. **Destaques continuam quentes.** O amarelo `#C68A900` (light) é o mesmo olho da Harpia do tema escuro; só perde 15% de luminosidade. Não usar amarelo-claro (#FFD700) nem amarelo-enxofre.
3. **Bordas substituem sombras.** No dark, profundidade vem de `bg-[#171E26]` sobre `bg-[#0A0F0D]`. No light, a separação vem de **bordas finas** (`#E5DFD2`) — sombras pesadas (drop-shadow) devem ser usadas com moderação, evitando a aparência "fofa" de UI infantil.
4. **Superfícies elevadas via branco, não via cinza.** Dropdowns, modais e tooltips devem usar `#FFFFFF` sobre o fundo `#FAF7F0` — não tons de cinza frio.
5. **Toasts e banners de erro.** Devem manter o ícone ⚠️ em `#A8412B` (Argila Queimada) sobre `#FFFFFF` com borda esquerda `#A8412B` 3px. Nunca usar fundo rosa-pastel.
6. **Toggle de tema.** Toda UI Harpia deve expor um seletor claro/escuro (ícones ☀️/🌙) na barra superior, persistindo a escolha em `localStorage` na chave `harpia_theme`.
