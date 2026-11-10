# Guia Didático — Desenvolvimento Frontend SPA com Harpia (Fase 4)

O **Harpia** permite o desenvolvimento de interfaces de usuário reativas, dinâmicas e de alta performance de ponta a ponta na mesma linguagem em que você escreve o seu servidor backend, sem precisar de compiladores JS ou frameworks externos complexos.

A Fase 4 introduz a transpilação direta do Harpia para JavaScript moderno (ES6) acompanhado de um motor próprio e levíssimo de Virtual DOM com reatividade baseada em Sinais (~2.2KB final).

---

## 🚀 Setup Mínimo em 5 Linhas

Crie um arquivo chamado `main.hrp`:

```harpia
funcao MeuApp() {
	retorne <h1 classe="texto-azul itens-centro p-4"> Olá Mundo Reativo! </h1>;
}
```

E compile usando a ferramenta CLI:

```bash
harpia compilar --alvo=web --entrada=main.hrp --saida=dist
```

Isto gerará um diretório `/dist` com:

- `index.html` (ponto de entrada pronto)
- `app.js` (sua lógica de negócio transpilada em JS)
- `runtime-web.js` (motor Virtual DOM e reatividade de Sinais do Harpia)
- `estilos.css` (folhas de estilos unificadas e classes utilitárias PT extraídas)

Abra o `dist/index.html` no seu navegador favorito e veja a mágica acontecer!

---

## ⚡ Primitivas de Reatividade (Sinais)

A reatividade no Harpia é de granularidade fina (fine-grained), o que significa que o motor atualiza **exclusivamente os nós do DOM que mudaram**, sem re-renderizar a árvore inteira.

### 1. `sinal(valorInicial)`

Cria um estado observável. Retorna uma tupla contendo a função de leitura (índice `0`) e a função de escrita (índice `1`).

```harpia
var contadorSinal = sinal(0);
var contador = contadorSinal[0];     # função para LER o sinal
var setContador = contadorSinal[1];  # função para DEFINIR o valor
```

### 2. `efeito(funcao)`

Executa a função imediatamente e a registra de forma automática para ser re-executada sempre que os Sinais lidos lá dentro sofrerem alteração.

```harpia
efeito(funcao() {
	imprimir("O valor atual do contador é: " + contador());
});
```

### 3. `derivado(funcao)`

Cria um sinal computado baseado em outros sinais. Ele é memoizado de forma inteligente (só re-calcula se os sinais de dependência mudarem).

```harpia
var dobro = derivado(funcao() {
	retorne contador() * 2;
});
```

### 4. `armazem(objeto)`

Cria um estado global reativo compartilhado (estilo Pub/Sub) para sincronização de estados complexos entre múltiplos componentes de página.

```harpia
var carrinho = armazem({ itens: [], total: 0 });
```

### 5. Importações Curtas da Stdlib Web

O transpiler e runtime web suportam importações curtas e limpas da stdlib de reatividade:

```harpia
de "sinais" importe sinal, efeito, derivado
de "reatividade" importe armazem
de "web" importe renderizar
```

---

## 🏗️ Sintaxe de Componentes JSX-like

O Harpia permite mesclar código e marcação visual HTML de forma natural.

### Declaração de Tags

```harpia
var elemento = <div classe="p-4"> Ola Mundo </div>;
```

### Componentes como Funções

Qualquer função que retorne uma marcação JSX-like pode atuar como um componente reutilizável.

```harpia
funcao Cabecalho() {
	retorne <header classe="fundo-azul p-4 texto-branco"> Meu Site </header>;
}
```

#### ⚠️ Assinatura de Propriedades e Regra de Linter `HRP-JSX-001`

Os componentes JSX personalizados (cujo nome começa com letra maiúscula, como `Cabecalho`) recebem um **dicionário de propriedades único** (`propriedades`) do runtime VDOM, e não argumentos posicionais isolados.

Por isso, para garantir compatibilidade estrutural e prevenir bugs silenciosos em tempo de execução, o **Linter Estático do Harpia** possui a regra estrita `HRP-JSX-001` que impede a declaração de componentes JSX personalizados com múltiplos parâmetros posicionais.

* ❌ **Incorreto**:
  ```harpia
  funcao Aba(nome, rotulo) { ... }  # Falha no linter (HRP-JSX-001)
  ```
* ✅ **Correto** (Dicionário único de propriedades):
  ```harpia
  funcao Aba(propriedades) {
      var nome = propriedades["nome"]
      var rotulo = propriedades["rotulo"]
      ...
  }
  ```

### Propriedades e Eventos

Atributos comuns e escutas de eventos em português são automaticamente traduzidos sob o capô pelo motor de Virtual DOM do Harpia para serem compreendidos nativamente pelos navegadores.

```harpia
funcao BotaoIncrementar() {
	retorne <botao classe="fundo-azul" aoClicar={setContador(contador() + 1)}> Incrementar </botao>;
}
```

#### Tabela de Atributos Traduzidos (VDOM)

O Harpia mapeia as seguintes propriedades comuns escritas em português diretamente para os padrões do HTML:

| Atributo em Harpia                         | HTML/DOM Equivalente | Descrição                                                      |
| :----------------------------------------- | :------------------- | :--------------------------------------------------------------- |
| `classe`                                 | `class`            | Classe CSS do elemento                                           |
| `identificador`                          | `id`               | Identificador único do elemento                                 |
| `estilo`                                 | `style`            | Objeto ou string de estilo CSS inline                            |
| `tipo`                                   | `type`             | Tipo de input (ex:`"text"`, `"number"`, `"submit"`)        |
| `valor`                                  | `value`            | Valor atual do elemento de formulário                           |
| `desabilitado`                           | `disabled`         | Desabilita interativamente o elemento                            |
| `obrigatorio`                            | `required`         | Torna o campo de formulário obrigatório                        |
| `nome`                                   | `name`             | Nome do elemento de input                                        |
| `marcado`                                | `checked`          | Estado marcado para caixas de seleção (checkbox)               |
| `espacoReservado` / `espaco_reservado` | `placeholder`      | Texto flutuante de dica dentro de um input                       |
| `chave`                                  | `key`              | Identificador único interno para reconciliação fina de arrays |
| `alvo`                                   | `target`           | Destino de abertura de um link (ex:`"_blank"`)                 |
| `caminho` / `fonte`                    | `src`              | Origem da imagem ou mídia                                       |
| `link` / `destino` / `endereco`      | `href`             | Destino de ancoragem do link                                     |
| `autoFoco` / `auto_foco`               | `autofocus`        | Foca o elemento automaticamente ao montar a página              |

---

## 📦 Importações e Resolução de Módulos no Navegador

Ao compilar com `--alvo=web`, todas as declarações de importação sem prefixo relativo (bare module specifiers como `de "arquivos" importe ...` ou `de "seguranca" importe ...`) são redirecionadas automaticamente pelo transpiler para o `./runtime-web.js` relativo à estrutura de diretórios do projeto.

Isso garante duas coisas:
1. **Padrão ES Modules Nativo dos Navegadores**: Navegadores rejeitam bare module specifiers (`import { x } from "modulo"`) emitindo `Uncaught TypeError: Failed to resolve module specifier`. Ao resolver todas as importações para caminhos relativos (ex: `./runtime-web.js`), o código funciona sem necessitar de bundlers adicionais.
2. **Exportação de Stubs Compatíveis**: O `runtime-web.js` fornece e exporta automaticamente implementações e stubs para utilitários da biblioteca padrão (ex: `ler`, `escrever`, `sanitizarHtml`, `sanitizarSqlArgumento`, `conectarSqlite`, `app`, `arquivos`, `seguranca`), permitindo reutilização fluida de código entre backend e frontend.
| `autoCompletar` / `auto_completar`     | `autocomplete`     | Configura o recurso de preenchimento automático                 |
| `leituraApenas` / `somente_leitura`    | `readonly`         | Impede a digitação do usuário no campo de texto               |

#### Tabela de Eventos Traduzidos (VDOM)

Escutas de eventos no Harpia começam com o prefixo `ao` seguido do verbo em CamelCase:

| Evento em Harpia              | Evento de Navegador Equivalente |
| :---------------------------- | :------------------------------ |
| `aoClicar`                  | `click`                       |
| `aoMudar`                   | `change`                      |
| `aoEnviar` / `aoSubmeter` | `submit`                      |
| `aoFocar`                   | `focus`                       |
| `aoDesfocar`                | `blur`                        |
| `aoTeclar`                  | `keydown`                     |
| `aoSoltarTecla`             | `keyup`                       |
| `aoEntrarMouse`             | `mouseenter`                  |
| `aoSairMouse`               | `mouseleave`                  |
| `aoDigitar` / `aoInserir` | `input`                       |
| `aoCarregar`                | `load`                        |
| `aoDuploClique`             | `dblclick`                    |
| `aoPassarMouse`             | `mouseover`                   |

### Estruturas de Controle Inline (JSX)

#### Renderização Condicional (`<se>`)

Use a tag `<se condicao={...}>` para renderização condicional dinâmica:

```harpia
<se condicao={contador() > 5}>
	<span> O valor está alto! </span>
</se>
```

#### Renderização de Listas (`<para>`)

Use a tag `<para item em lista={...}>` para renderizar loops de listas ou arrays:

```harpia
<ul class="flex-coluna py-2">
	<para item em lista={lerTarefas()}>
		<li classe="p-2 borda-abaixo">
			<span> {item} </span>
		</li>
	</para>
</ul>
```

---

## 🎨 Estilização Unificada e Classes Utilitárias PT

O Harpia oferece duas formas unificadas de estilização de nível profissional:

### 1. Bloco de Estilo Declarativo em Português (`estilo`)

Permite declarar estilos estruturados diretamente em português com suporte a pseudo-classes e chaves mapeadas de forma canônica:

```harpia
estilo MeuCard {
	corDeFundo: "#ffffff";   # Mapeia para background-color
	padding: "1.5rem";       # Mantém propriedades normais
	largura: "400px";        # Mapeia para width
	raio-grande: true;       # Mapeia para border-radius de canto arredondado

	botao:hover {            # Suporta pseudo-seletores aninhados
		opacidade: 0.8;
	}
}
```

#### Concatenação Dinâmica em Tags `<style>` JSX

Tags `<style>` embutidas em componentes JSX aceitam expressões de concatenação multilinhas de strings:

```harpia
<style>{
	".card { background: white; }" +
	".card:hover { transform: scale(1.02); }"
}</style>
```

### 2. Classes Utilitárias Nativas ("Tailwind" em PT)

O compilador de Harpia possui um motor inteligente de extração sob demanda que escaneia o atributo `classe="..."` e gera no `estilos.css` apenas as classes utilitárias que você realmente utilizou.

Exemplos de classes em português suportadas:

- **Layout/Flexbox**: `flex-linha`, `flex-coluna`, `itens-centro`, `conteudo-centro`, `conteudo-espacado`.
- **Dimensões**: `largura-cheia`, `altura-cheia`, `largura-tela`, `altura-tela`.
- **Padding/Margin**: `p-1`, `p-4`, `px-2`, `py-4`, `m-1`, `m-4`, `mx-2`, `my-4`.
- **Cores de Texto**: `texto-branco`, `texto-preto`, `texto-azul`, `texto-vermelho`, `texto-verde`.
- **Cores de Fundo**: `fundo-branco`, `fundo-cinza`, `fundo-azul`, `fundo-vermelho`, `fundo-verde`.
- **Bordas**: `borda`, `raio-pequeno`, `raio-medio`, `raio-grande`, `raio-cheio`.

```harpia
# Exemplo de uso de utilitários
var card = <div classe="flex-linha itens-centro p-4 fundo-cinza raio-medio borda"> ... </div>;
```

---

## 🛣️ Roteamento SPA por Arquivos (File-system Routing)

O compilador do Harpia web detecta de forma inteiramente automatizada o diretório `/web/rotas/` (ou `/rotas/`) do projeto.

### Mapeamento Automático

- `web/rotas/index.hrp` → Mapeia para a URL `/`
- `web/rotas/sobre.hrp` → Mapeia para a URL `/sobre`
- `web/rotas/blog/artigo.hrp` → Mapeia para a URL `/blog/artigo`

### Navegação Dinâmica sem Recarregar (`<Link>`)

Para navegar entre rotas de forma instantânea sem recarregar fisicamente a página, use o componente nativo `<Link>`:

```harpia
<Link para="/sobre"> Conheça mais sobre nós </Link>
```

---

## 🌐 Renderização no Servidor (SSR) e Hidratação

O servidor HTTP nativo do Harpia (`stdlib/http/http.go`) pode renderizar a árvore de componentes em HTML estático inicial instantâneo para fins de SEO de alto nível e carregamento imediato.

No navegador do usuário final, o runtime web liga de forma invisível os ouvintes de eventos e os Sinais na estrutura existente (Processo de Hidratação), ligando os fios de interatividade sem piscar, recriar ou destruir os elementos do DOM.

### Metadados Inteligentes para IA e SEO (AEO & GEO)

Cada arquivo de rota pode exportar um objeto de configuração `metadados` contendo Schema.org (`esquema`) e OpenGraph. O servidor SSR gerará marcações de JSON-LD ricas estruturadas para indexação de buscas por IAs (AEO) e serviços de localização (GEO).

```harpia
exportar var metadados = {
	titulo: "Minha Empresa",
	descricao: "Desenvolvimento rápido em português corporativo",
	esquema: "Organization",
	localizacao: "São Paulo, SP"
};
```

---

## 🏛️ Comparação Arquitetural: Para Desenvolvedores Angular

Se você tem experiência com **Angular**, o desenvolvimento web em Harpia parecerá bastante familiar devido ao uso de **Sinais (Sinais)** e componentes modulares. O ecossistema do Harpia Web, porém, adota uma abordagem minimalista de arquivo único (estilo Solid.js ou Svelte), eliminando cerimônias de injeção de dependência complexas e decoradores pesados.

### Tabela de Equivalências

| Conceito no Angular        | Equivalente em Harpia                     | Descrição Prática                                                                       |
| :------------------------- | :---------------------------------------- | :----------------------------------------------------------------------------------------- |
| **Pipes**            | **Operador Pipe Nativo (`\|>`)**   | Formatação nativa elegante de expressões em templates JSX.                              |
| **Directives**       | **Tags de Controle JSX**            | Condicionais`<se condicao={...}>` e loops de repetição `<para item em lista={...}>`. |
| **Services & State** | **Módulos com `armazem`**        | Exportação de primitivas`armazem` de estado global reativo compartilhado.              |
| **Validators**       | **Sinais Derivados (`derivado`)** | Validações finas escritas como funções derivadas que recalcularão sob demanda.        |
| **Models**           | **`classe` ou Mapas**             | Uso de classes nativas de orientação a objetos com o método`inicializar()`.           |

### Exemplo Comparativo Prático

```harpia
# 1. MÓDULO DE SERVIÇO (carrinho.hrp)
var carrinho = armazem({ itens: [], total: 0 });

funcao adicionarAoCarrinho(nome, preco) {
	carrinho.itens.adicionar({ nome: nome, preco: preco });
	carrinho.total = carrinho.total + preco;
}

# 2. O COMPONENTE (MeuApp.hrp)
funcao formatarMoeda(valor) {
	retorne "R$ " + valor.texto();
}

estilo MeuCarrinho {
	corDeFundo: "#f9fafb";
	padding: "1rem";
}

funcao MeuApp() {
	retorne <div classe="MeuCarrinho p-4 raio-medio">
		<h2> Seu Carrinho de Compras </h2>
	
		# Diretiva de controle (Substitui *ngIf)
		<se condicao={carrinho.itens.tamanho() == 0}>
			<p classe="texto-cinza"> Carrinho vazio! </p>
		</se>

		# Diretiva de loop (Substitui *ngFor)
		<ul>
			<para item em lista={carrinho.itens}>
				<li class="flex-linha conteudo-espacado p-2">
					<span> {item.nome} </span>
					# Uso do Operador Pipe Nativo
					<span> {item.preco |> formatarMoeda} </span>
				</li>
			</para>
		</ul>

		<h3> Total: {carrinho.total |> formatarMoeda} </h3>
	</div>;
}
```

---

## 📦 Modelo Híbrido de Arquivos (`.hrp`, `.html`, `.estilo.hrp`)

Para aplicações corporativas, manter lógica, estilo e estrutura física dentro de um único arquivo pode causar poluição visual. O Harpia permite organizar o projeto de forma híbrida e modular:

1. **Estilos Separados em Português (`.estilo.hrp`)**: Declare seus blocos `estilo Nome { ... }` em um arquivo de extensão `.estilo.hrp` e importe no arquivo lógico. O compilador em Go lê as importações e escreve no `estilos.css` final de forma automática, criando constantes string locais de referência no seu JS transpilado para evitar erros de variável indefinida.
2. **Templates HTML Físicos Separados (`.html`)**: Desenvolva o layout visual do seu componente de forma isolada em um arquivo `.html` tradicional (com interpolações de chaves e condicionais reativas) e importe de dentro da função lógica utilizando a chamada nativa `importarHtml("./layout.html")`. O compilador faz o inline dinâmico e a transpilação do HTML em tempo de compilação.

```harpia
# web/componentes/MeuCard.hrp
de "web" importe sinal, importarHtml;
de "./MeuCard.estilo.hrp" importe EstiloDoCard; # Estilos vêm do .estilo.hrp

exportar funcao MeuCard() {
    var [expandido, setExpandido] = sinal(falso);
  
    # Carrega e transpila o layout físico HTML de forma transparente
    retorne importarHtml("./MeuCard.html");
}
```

---

## 🛠️ Recursos de Produção e Otimizações de Performance (Fase 4-C)

A Fase 4-C introduz primitivas de alta performance e utilitários projetados para viabilizar sistemas de grande porte:

### 1. Two-Way Data Binding (`ligar={sinal}`)

Elimina o boilerplate tradicional de ligar valores e escutar mudanças físicas. Ao declarar `<input ligar={nome} />`, o compilador gera as propriedades reativas necessárias e o runtime do browser faz o bind bidirecional instantâneo.

### 2. Modificadores de Eventos Declarativos

Controle o comportamento físico de eventos diretamente do JSX-like sem precisar injetar chamadas manuais nas funções:

* `aoEnviar_prevenir={submeter}` ➔ Executa `e.preventDefault()` de forma transparente.
* `aoClicar_parar={acao}` ➔ Executa `e.stopPropagation()` automaticamente.

### 3. Keyed Diffing (`chave`) no Virtual DOM

Para loops complexos utilizando `<para>`, o algoritmo de reconciliação de Virtual DOM do Harpia analisa o atributo `chave`. Se a chave coincidir entre renderizações, o nó físico do DOM é preservado e apenas movido de lugar, reduzindo drasticamente o custo computacional de redesenho na tela de $O(N^2)$ para $O(N)$ linear.

### 4. Sinais Persistentes (`sinalPersistente`)

Cria estados reativos normais que persistem de forma inteiramente síncrona e transparente no `localStorage` do navegador do usuário, sobrevivendo a recarregamentos de página:

```harpia
var [tema, setTema] = sinalPersistente("tema", "claro");
```

### 5. Sinais Assíncronos (`recurso`)

Unifica chamadas assíncronas HTTP e consumo de APIs com o fluxo de renderização reativo do frontend, fornecendo variáveis de controle reativas de progresso:

* `usuario.carregando()` ➔ Retorna `Verdadeiro` durante a requisição.
* `usuario.erro()` ➔ Retorna o erro capturado caso a promessa falhe.
* `usuario.ok()` ➔ Retorna `Verdadeiro` quando os dados forem obtidos com sucesso.

### 6. Injeção de Dependências (`Provedor` & `injetar`)

Instancie e prove serviços e stores no topo da sua árvore de componentes usando `<Provedor chave="servico" valor={instancia}>` e recupere-os em qualquer componente filho profundo com `var servico = injetar("servico")`, evitando o acoplamento de propriedades.

### 7. Componentes de UI Corporativos Nativos

O Harpia Web disponibiliza componentes utilitários de alta usabilidade e performance embutidos no runtime:

* `<FronteiraDeErro>`: Barreira protetora que captura falhas em widgets secundários e exibe um fallback personalizado, impedindo que erros derrubem a aplicação inteira.
* `<ListaVirtual>`: Renderiza estritamente os nós físicos que estão visíveis na tela para listas massivas de dados (ex: 50.000 linhas), mantendo a rolagem macia e leve.
* `<GradeDeDados>`: Tabela interativa robusta com paginação, filtro de pesquisa em tempo real e ordenação rápida de colunas em português.

---

## 🚀 Inicializador de Projetos (Scaffolding)

Para começar a desenvolver imediatamente na nova arquitetura híbrida, utilize a CLI Cobra do Harpia:

```bash
harpia iniciar nome-do-seu-projeto
```

Isso criará uma estrutura de pastas limpa e organizada no disco contendo exemplos práticos de arquivos `.hrp`, `.html` e `.estilo.hrp` integrados e prontos para rodar.
