package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// comandoNovo gerencia a criação direcionada de novas aplicações (Monolito, Backend ou Frontend)
func comandoNovo() *cobra.Command {
	novo := &cobra.Command{
		Use:     "novo",
		Aliases: []string{"iniciar", "inicializar"},
		Short:   "Inicializa uma nova estrutura de projeto Harpia",
	}

	novo.AddCommand(comandoMonolito())
	novo.AddCommand(comandoBackend())
	novo.AddCommand(comandoFrontend())

	return novo
}

func comandoMonolito() *cobra.Command {
	return &cobra.Command{
		Use:   "monolito [nome-do-projeto]",
		Short: "Cria um novo monolito completo com Clean Architecture e DDD em português",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			nomeProj := args[0]
			// ponytail: impede sobrescrita acidental de projetos existentes
			if _, err := os.Stat(nomeProj); err == nil {
				fmt.Fprintf(os.Stderr, "Alerta: O diretório '%s' já existe! Abortando para evitar sobrescrita física.\n", nomeProj)
				os.Exit(1)
			}
			fmt.Printf("Inicializando monolito com Clean Architecture: %s...\n", nomeProj)

			diretorios := []string{
				nomeProj,
				filepath.Join(nomeProj, "dominio"),
				filepath.Join(nomeProj, "dominio", "entidades"),
				filepath.Join(nomeProj, "dominio", "repositorios"),
				filepath.Join(nomeProj, "infra"),
				filepath.Join(nomeProj, "infra", "banco"),
				filepath.Join(nomeProj, "web"),
				filepath.Join(nomeProj, "web", "rotas"),
				filepath.Join(nomeProj, "web", "pages"),
				filepath.Join(nomeProj, "web", "componentes"),
				filepath.Join(nomeProj, "testes"),
				filepath.Join(nomeProj, "docs"),
			}

			for _, dir := range diretorios {
				if err := os.MkdirAll(dir, 0755); err != nil {
					fmt.Fprintf(os.Stderr, "Erro ao criar pasta %s: %v\n", dir, err)
					os.Exit(1)
				}
			}

			// Escreve os arquivos padrões do monolito
			escreverArquivo(nomeProj, filepath.Join("dominio", "entidades", "usuario.hrp"), `# Entidade de domínio tipada e segura
exportar classe Usuario {
	inicializar(self, id: Inteiro, nome: Texto, email: Texto) {
		self.id = id
		self.nome = nome
		self.email = email
	}
}
`)

			escreverArquivo(nomeProj, filepath.Join("dominio", "README.md"), `# 🧬 Camada de Domínio (O Coração do Sistema)

Esta é a camada mais interna e importante do seu aplicativo. Ela abriga todas as regras e lógicas de negócios centrais da sua aplicação de forma isolada.

- **entidades/**: Modelos de dados ricos e tipados (.hrp).
- **repositorios/**: Interfaces e contratos de portas de comunicação para persistência (.hrp).

## 📌 Boas Práticas e Regras de Ouro:
1. **Isolamento Absoluto**: Esta camada nunca deve ter dependências de infraestrutura, bancos ou frameworks de terceiros. Deve ser puro código Harpia.
2. **Entidades Ricas**: Prefira entidades ricas (com lógicas de validação em seu inicializador) a estruturas de dados anêmicas e sem comportamento.
3. **Inversão de Dependência**: O domínio define os contratos de acesso (interfaces em repositories); a infraestrutura os implementa.
`)

			escreverArquivo(nomeProj, filepath.Join("infra", "banco", "conexao.hrp"), `de "bd" importe conectar;

exportar funcao obterBanco() {
	# SQLite embarcado de forma nativa e segura
	retorne conectar("sqlite", "dados.db");
}
`)

			escreverArquivo(nomeProj, filepath.Join("infra", "README.md"), `# 🔌 Camada de Infraestrutura (Detalhes de Tecnologia)

A camada de infraestrutura é responsável por conectar a lógica puramente conceitual do domínio com o mundo físico exterior (banco de dados, rede, etc.).

- **banco/**: Arquivos de conexão, drivers e repositórios concretos (ex: SQLite em .hrp).

## 📌 Boas Práticas e Regras de Ouro:
1. **Lógica de Persistência**: Isole queries, SQLs e ORM nesta camada. O domínio não deve saber se os dados vêm de arquivos locais, bancos SQL ou NoSQL.
2. **Resiliência Integrada**: Use disjuntores, limites de taxa ou retentativas na infraestrutura para blindar o sistema contra picos de erro de conexões externas.
`)

			escreverArquivo(nomeProj, filepath.Join("web", "rotas", "rotas.hrp"), `de "web" importe h;
de "../pages/Inicio.hrp" importe RotaInicio;

exportar funcao RotaIndex() {
	retorne <RotaInicio />;
}
`)

			escreverArquivo(nomeProj, filepath.Join("web", "rotas", "README.md"), conteudoRotasReadmeMd)

			criarAssetsPadrao(nomeProj)

			escreverArquivo(nomeProj, filepath.Join("web", "README.md"), `# 🌐 Camada Web e Apresentação (Frontend SPA)

Esta camada é responsável por expor a interface visual (UI) para o navegador de forma ultra veloz por meio do Virtual DOM reativo nativo do Harpia.

- **pages/**: Estrutura de telas, painéis e páginas em HTML, estilos específicos e lógica associada. Cada página possui a tríade de arquivos integrada: ".hrp" (lógica), ".estilo.hrp" (estilos) e ".html" (marcação visual).
- **componentes/**: Reservada para criar componentes visuais menores reutilizáveis (como botões, cards, modais, etc.) importados futuramente nas páginas.
- **global.estilos.hrp**: Arquivo na raiz de 'web/' contendo os estilos gerais da aplicação.
- **rotas/**: Telas, dashboards e páginas de roteamento SPA (.hrp).

## 📌 Boas Práticas e Regras de Ouro:
1. **Estado Fino com Sinais**: Use Sinais (sinal) no escopo global para gerenciar dados reativos, lendo seus valores via getter para atualizações cirúrgicas no DOM.
2. **Design Tokens Unificados**: Centralize cores e espaçamentos no seu estilos.hrp usando a palavra-chave 'estilo' do Harpia para preservar consistência.
`)

			escreverArquivo(nomeProj, filepath.Join("web", "global.estilos.hrp"), `exportar estilo TituloGlobal {
	tamanhoFonte: "2.5rem";
	alinhamentoTexto: "center";
	cor: "#171e26";
}

exportar estilo CorpoGlobal {
	margem: "0";
	preenchimento: "0";
	famíliaFonte: "sans-serif";
}
`)

			escreverArquivo(nomeProj, filepath.Join("web", "pages", "Inicio.hrp"), `de "web" importe sinal, importarHtml;
de "./Inicio.estilo.hrp" importe Aplicacao, BotaoPrincipal;

exportar funcao RotaInicio() {
	var contadorSinal = sinal(0);
	var contador = contadorSinal[0];
	var setContador = contadorSinal[1];

	var incrementar = funcao() {
		setContador(contador() + 1);
	};

	retorne importarHtml("./Inicio.html");
}
`)

			escreverArquivo(nomeProj, filepath.Join("web", "pages", "Inicio.estilo.hrp"), `exportar estilo Aplicacao {
	famíliaFonte: "sans-serif";
	padding: "40px";
	alinhamentoTexto: "center";
	cor: "#171e26";
}

exportar estilo BotaoPrincipal {
	corDeFundo: "#00A86B";
	cor: "#ffffff";
	borda: "none";
	padding: "10px 20px";
	raioBorda: "4px";
	cursor: "pointer";
}
`)

			escreverArquivo(nomeProj, filepath.Join("web", "pages", "Inicio.html"), `<div class={Aplicacao}>
	<h1>🦅 Olá, Mundo Harpia!</h1>
	<p>Desenvolvido em português com Clean Architecture e DDD.</p>

	<div style="margin-top: 30px; padding: 20px; background: #f3f4f6; border-radius: 8px; display: inline-block;">
		<p>Cliques reativos: <strong>{contador()}</strong></p>
		<button aoClicar={incrementar} class={BotaoPrincipal}>
			Incrementar
		</button>
	</div>
</div>
`)

			escreverArquivo(nomeProj, filepath.Join("docs", "arquitetura.md"), `# 🏛️ Guia de Arquitetura do Projeto Harpia (Clean Architecture + DDD)

Este projeto foi estruturado sob os princípios da **Arquitetura Limpa (Clean Architecture)** e do **Design Orientado a Domínio (DDD)** de forma nativa em português.

---

## 🧭 O Fluxo de Dependência (A Regra de Ouro)
O princípio central da Clean Architecture é que **as dependências de código devem apontar apenas para dentro**. As camadas externas (infraestrutura e web) conhecem o domínio, mas o domínio **nunca** deve importar nada da infraestrutura ou da web.

---

## 📁 Estrutura de Camadas e Convenções

### 1. Camada de Domínio ("dominio/")
Representa o núcleo do software. Ela contém as regras de negócios cruciais que seriam verdadeiras mesmo se o aplicativo não fosse um sistema computacional.

* **Convenções:**
  * **entidades/**: Modelos de domínio ricos que possuem identidade própria (ex: "Usuario"). Devem validar seus próprios estados durante a inicialização.
  * **repositorios/**: Contratos de persistência de dados (interfaces). Definem quais dados o domínio precisa salvar ou ler, sem saber qual banco será usado.

### 2. Camada de Infraestrutura ("infra/")
Contém os detalhes técnicos e implementações concretas que dão suporte ao funcionamento do domínio.

* **Convenções:**
  * **banco/**: Drivers, conexões físicas (ex: SQLite) e implementações de repositórios concretos.

### 3. Camada de Apresentação ("web/")
Responsável pela interface do usuário (UI) e controle de rotas no navegador (SPA).

* **Convenções:**
  * **pages/**: Estrutura de telas e páginas completas. Cada página segue a arquitetura tripla e sintonizada: um arquivo de marcação ".html", um arquivo de folha de estilos ".estilo.hrp" e um arquivo lógico ".hrp" que integra ambos e gerencia os sinais locais.
  * **componentes/**: Elementos e blocos visuais reutilizáveis menores (como botões, cards, barras) criados futuramente e importados nas páginas.
  * **global.estilos.hrp**: Arquivo na raiz de 'web/' contendo as classes de estilos globais da aplicação.
  * **rotas/**: Controladores de fluxo e arquivos de roteamento SPA ("rotas.hrp").
`)

			escreverArquivo(nomeProj, filepath.Join("docs", "comandos.md"), conteudoComandosMd)

			escreverArquivo(nomeProj, "main.hrp", `de "web" importe montar;
de "./web/rotas/rotas.hrp" importe RotaIndex;

funcao Aplicacao() {
	retorne <div class="App">
		<RotaIndex />
	</div>;
}

montar(Aplicacao, Nulo);
`)

			escreverArquivo(nomeProj, filepath.Join("testes", "usuario_test.hrp"), `de "assegura" importe assegure;
de "../dominio/entidades/usuario.hrp" importe Usuario;

testar "deve instanciar um usuario de dominio com tipos corretos" {
	var user = nova Usuario(1, "Natan", "natan@Harpia.org")
	assegure(user.id == 1)
	assegure(user.nome == "Natan")
}
`)

			escreverArquivo(nomeProj, "README.md", `# 🦅 Bem-vindo ao `+nomeProj+` (Monolito Harpia)

Este projeto monolito foi gerado de forma automática sob as premissas de **Clean Architecture** e **DDD** com termos em português.

---

## 🧭 Como Navegar
Consulte a pasta de documentação para guias e especificações de uso:
* 🏛️ [Manual de Arquitetura](docs/arquitetura.md) — Explicação de Domínio, Infra e Apresentação.
* 🎮 [Guia de Comandos](docs/comandos.md) — Comandos do CLI, flags e necessidades.

---

## ⚡ Como Rodar a Aplicação
Para compilar e servir o seu projeto no navegador, execute os seguintes comandos no terminal:

`+"```"+`bash
# 1. Compila para a web gerando o build estático
harpia compilar --alvo=web --entrada=main.hrp --saida=dist

# 2. Inicia o servidor local leve de hospedagem
harpia servir --diretorio=dist
`+"```"+`
`)

			exibirSucessoCompleto(nomeProj, "monolito")
		},
	}
}

func comandoBackend() *cobra.Command {
	return &cobra.Command{
		Use:   "backend [nome-do-projeto]",
		Short: "Cria uma estrutura de backend focada em APIs lógicas, banco e concorrência",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			nomeProj := args[0]
			// ponytail: impede sobrescrita acidental de projetos existentes
			if _, err := os.Stat(nomeProj); err == nil {
				fmt.Fprintf(os.Stderr, "Alerta: O diretório '%s' já existe! Abortando para evitar sobrescrita física.\n", nomeProj)
				os.Exit(1)
			}
			fmt.Printf("Inicializando projeto de backend: %s...\n", nomeProj)

			diretorios := []string{
				nomeProj,
				filepath.Join(nomeProj, "dominio"),
				filepath.Join(nomeProj, "dominio", "entidades"),
				filepath.Join(nomeProj, "infra"),
				filepath.Join(nomeProj, "infra", "banco"),
				filepath.Join(nomeProj, "testes"),
				filepath.Join(nomeProj, "docs"),
			}

			for _, dir := range diretorios {
				if err := os.MkdirAll(dir, 0755); err != nil {
					fmt.Fprintf(os.Stderr, "Erro ao criar pasta %s: %v\n", dir, err)
					os.Exit(1)
				}
			}

			escreverArquivo(nomeProj, filepath.Join("dominio", "entidades", "produto.hrp"), `exportar classe Produto {
	inicializar(self, id: Inteiro, nome: Texto, preco: Decimal) {
		self.id = id
		self.nome = nome
		self.preco = preco
	}
}
`)

			escreverArquivo(nomeProj, filepath.Join("dominio", "README.md"), `# 🧬 Camada de Domínio do Backend (O Coração da API)

Camada central onde residem as regras de negócio puras (entidades e modelos de domínio), livre de conexões com banco de dados ou protocolo de rede HTTP/gRPC.

## 📌 Boas Práticas:
1. **Modelagem de Domínio Rito**: Agregue lógicas, regras fiscais e validações diretamente nos métodos das entidades em "entidades/".
2. **Independência Total**: O domínio define o que o sistema faz; ele nunca deve saber onde ou como os dados são salvos ou enviados.
`)

			escreverArquivo(nomeProj, filepath.Join("infra", "README.md"), `# 🔌 Camada de Infraestrutura do Backend (Detalhes Técnicos)

Aqui residem as implementações concretas e adaptadores de entrada/saída (conexões de banco, clientes HTTP, envio de e-mails, etc.).

## 📌 Boas Práticas:
1. **Consultas e queries**: Mantenha as consultas brutas SQL e integrações externas focadas nesta camada.
2. **Robustez de Rede**: Implemente disjuntores de circuito (circuit breakers) e lógicas de retentativa para garantir alta disponibilidade.
`)

			escreverArquivo(nomeProj, filepath.Join("docs", "arquitetura.md"), `# 🏛️ Arquitetura de Backend Harpia (Clean Architecture + DDD)

Uma estrutura robusta e performática voltada para o desenvolvimento de microsserviços, APIs corporativas e concorrência orientada a eventos em português.

- **dominio/**: Entidades de negócio ricas e tipadas, e definições de contratos.
- **infra/**: Conectores e adaptadores físicos para banco de dados e APIs externas.
- **main.hrp**: Ponto de entrada do executável que liga as rotas HTTP e inicializa o servidor.
`)

			escreverArquivo(nomeProj, filepath.Join("docs", "comandos.md"), conteudoComandosMd)

			escreverArquivo(nomeProj, "main.hrp", `de "http" importe Servidor;
de "json" importe analisar;

var api = novo Servidor();

api.obter("/api/produtos", funcao(req, res) {
	var resposta = [
		{ "id": 1, "nome": "Caneta", "preco": 1.50 },
		{ "id": 2, "nome": "Caderno", "preco": 15.90 }
	]
	res.status(200).json(resposta)
});

imprimir("Servidor backend ativo na porta 8080...")
api.ouvir(8080)
`)

			escreverArquivo(nomeProj, "README.md", `# 🦅 Bem-vindo ao `+nomeProj+` (Backend Harpia)

Este projeto de backend foi gerado de forma automática com foco em APIs lógicas, conectores de banco de dados e concorrência leve em português.

---

## 🧭 Como Navegar
Consulte a pasta de documentação para guias e especificações de uso:
* 🏛️ [Manual de Arquitetura](docs/arquitetura.md) — Explicação do fluxo de dados e portas de banco.
* 🎮 [Guia de Comandos](docs/comandos.md) — Comandos do CLI, flags e necessidades.

---

## ⚡ Como Rodar o Servidor Backend
Para executar o seu servidor backend imediatamente, execute no terminal:

`+"```"+`bash
# Executa o servidor na máquina local
harpia executar main.hrp
`+"```"+`
`)

			exibirSucessoCompleto(nomeProj, "backend")
		},
	}
}

func comandoFrontend() *cobra.Command {
	return &cobra.Command{
		Use:   "frontend [nome-do-projeto]",
		Short: "Cria uma estrutura reativa cliente puramente SPA de alto desempenho",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			nomeProj := args[0]
			// ponytail: impede sobrescrita acidental de projetos existentes
			if _, err := os.Stat(nomeProj); err == nil {
				fmt.Fprintf(os.Stderr, "Alerta: O diretório '%s' já existe! Abortando para evitar sobrescrita física.\n", nomeProj)
				os.Exit(1)
			}
			fmt.Printf("Inicializando projeto de frontend SPA: %s...\n", nomeProj)

			diretorios := []string{
				nomeProj,
				filepath.Join(nomeProj, "web"),
				filepath.Join(nomeProj, "web", "rotas"),
				filepath.Join(nomeProj, "web", "pages"),
				filepath.Join(nomeProj, "web", "componentes"),
				filepath.Join(nomeProj, "testes"),
				filepath.Join(nomeProj, "docs"),
			}

			for _, dir := range diretorios {
				if err := os.MkdirAll(dir, 0755); err != nil {
					fmt.Fprintf(os.Stderr, "Erro ao criar pasta %s: %v\n", dir, err)
					os.Exit(1)
				}
			}

			escreverArquivo(nomeProj, filepath.Join("web", "rotas", "rotas.hrp"), `de "web" importe h;
de "../pages/Inicio.hrp" importe RotaInicio;

exportar funcao RotaIndex() {
	retorne <RotaInicio />;
}
`)

			escreverArquivo(nomeProj, filepath.Join("web", "rotas", "README.md"), conteudoRotasReadmeMd)

			criarAssetsPadrao(nomeProj)

			escreverArquivo(nomeProj, filepath.Join("web", "global.estilos.hrp"), `exportar estilo TituloGlobal {
	tamanhoFonte: "2.5rem";
	alinhamentoTexto: "center";
	cor: "#171e26";
}

exportar estilo CorpoGlobal {
	margem: "0";
	preenchimento: "0";
	famíliaFonte: "sans-serif";
}
`)

			escreverArquivo(nomeProj, filepath.Join("web", "pages", "Inicio.hrp"), `de "web" importe sinal, h;
de "./Inicio.estilo.hrp" importe Aplicacao, EntradaTexto;

exportar funcao RotaInicio() {
	var nomeSinal = sinal("");
	var nome = nomeSinal[0];
	var setNome = nomeSinal[1];

	retorne <div class={Aplicacao}>
		<h1>🦅 Cliente SPA Harpia!</h1>
		<p>Modifique o nome abaixo para ver a reatividade em tempo real:</p>

		<input ligar={nomeSinal} placeholder="Escreva seu nome..." class={EntradaTexto} />

		<se condicao={nome() != ""}>
			<p style="color: #00A86B; font-weight: bold; font-size: 1.2rem;">Seja muito bem-vindo, {nome()}!</p>
		</se>
	</div>;
}
`)

			escreverArquivo(nomeProj, filepath.Join("web", "pages", "Inicio.estilo.hrp"), `exportar estilo Aplicacao {
	famíliaFonte: "sans-serif";
	padding: "40px";
	alinhamentoTexto: "center";
	cor: "#171e26";
}

exportar estilo EntradaTexto {
	borda: "1px solid #ccc";
	padding: "10px";
	raioBorda: "4px";
	margem: "15px 0";
	tamanhoFonte: "1rem";
	alinhamentoTexto: "center";
}
`)

			escreverArquivo(nomeProj, filepath.Join("web", "README.md"), `# 🌐 Camada Web e Apresentação (Frontend SPA)

Esta pasta abriga os elementos de renderização dinâmica no navegador (SPA), como rotas e componentes de interface Virtual DOM com suporte a folhas de estilo declaradas em português.

- **pages/**: Estrutura de telas, painéis e páginas em HTML, estilos específicos e lógica associada. Cada página possui a tríade de arquivos integrada: ".hrp" (lógica), ".estilo.hrp" (estilos) e ".html" (marcação visual).
- **componentes/**: Pasta vazia reservada para criar componentes visuais menores reutilizáveis (como botões, cards, modais, etc.) importados futuramente nas páginas.
- **global.estilos.hrp**: Arquivo na raiz de 'web/' contendo os estilos gerais da aplicação.
- **rotas/**: Telas, dashboards e páginas de roteamento SPA (.hrp).

## 📌 Boas Práticas:
1. **Controle de Estado Cirúrgico**: Faça o controle de dados reativos por meio de Sinais (sinal) para obter alta precisão e evitar renderizações pesadas do DOM.
2. **Estilizações Reutilizáveis**: Evite escrever estilos inline extensos; centralize cores e layouts em "web/global.estilos.hrp" para reuso.
`)

			escreverArquivo(nomeProj, filepath.Join("docs", "arquitetura.md"), `# 🏛️ Arquitetura Frontend Harpia (SPA)

Uma estrutura reativa focada no cliente SPA (Single Page Application) de altíssimo desempenho, alimentada integralmente pelo motor de Virtual DOM e Sinais Reativos do Harpia.

- **web/pages/**: Estrutura de telas e páginas completas. Cada página segue a arquitetura tripla e sintonizada: um arquivo de marcação ".html", um arquivo de folha de estilos ".estilo.hrp" e um arquivo lógico ".hrp" que integra ambos e gerencia os sinais locais.
- **web/componentes/**: Blocos de interface declarativos e componentes menores reutilizáveis criados futuramente.
- **web/global.estilos.hrp**: Arquivo na raiz de 'web/' contendo as classes de estilos globais da aplicação.
- **web/rotas/**: Telas e fluxos de navegação reativos baseados em rotas amigáveis ("rotas.hrp").
- **main.hrp**: Ponto de entrada que carrega os estilos e monta o componente raiz na interface.
`)

			escreverArquivo(nomeProj, filepath.Join("docs", "comandos.md"), conteudoComandosMd)

			escreverArquivo(nomeProj, "main.hrp", `de "web" importe montar;
de "./web/rotas/rotas.hrp" importe RotaIndex;

montar(RotaIndex, Nulo);
`)

			escreverArquivo(nomeProj, "README.md", `# 🦅 Bem-vindo ao `+nomeProj+` (Frontend SPA Harpia)

Este projeto de frontend reativo SPA foi gerado de forma automática com base em Virtual DOM e Sinais reativos de alto desempenho em português.

---

## 🧭 Como Navegar
Consulte a pasta de documentação para guias e especificações de uso:
* 🏛️ [Manual de Arquitetura](docs/arquitetura.md) — Explicação de componentes, rotas e sinais.
* 🎮 [Guia de Comandos](docs/comandos.md) — Comandos do CLI, flags e necessidades.

---

## ⚡ Como Rodar a Aplicação
Para compilar e servir o seu projeto no navegador, execute no seu terminal:

`+"```"+`bash
# 1. Compila para a web gerando o build estático
harpia compilar --alvo=web --entrada=main.hrp --saida=dist

# 2. Inicia o servidor local leve de hospedagem
harpia servir --diretorio=dist
`+"```"+`
`)

			exibirSucessoCompleto(nomeProj, "frontend")
		},
	}
}

// comandoCrie gerencia a geração assistida de novos arquivos (Rotas ou Componentes) em projetos existentes
func comandoCrie() *cobra.Command {
	crie := &cobra.Command{
		Use:   "crie",
		Short: "Gera novos templates estruturados de arquivos (rota, componente ou modelo)",
	}

	crie.AddCommand(comandoCrieRota())
	crie.AddCommand(comandoCrieComponente())
	crie.AddCommand(comandoCrieModelo())

	return crie
}

func comandoCrieRota() *cobra.Command {
	return &cobra.Command{
		Use:   "rota [nome]",
		Short: "Gera uma nova tríade de arquivos para uma página/rota de SPA",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			nome := args[0]
			nomeLimp := strings.Title(strings.ToLower(nome))

			// Detecta se existe a pasta web/pages ou pages
			pastaPages := "pages"
			if _, err := os.Stat("web/pages"); err == nil {
				pastaPages = "web/pages"
			} else {
				os.MkdirAll("pages", 0755)
			}

			caminhoLogico := filepath.Join(pastaPages, nomeLimp+".hrp")
			caminhoEstilo := filepath.Join(pastaPages, nomeLimp+".estilo.hrp")
			caminhoVisual := filepath.Join(pastaPages, nomeLimp+".html")

			conteudoEstilo := fmt.Sprintf(`exportar estilo Conteiner%s {
	padding: "30px";
	famíliaFonte: "sans-serif";
}
`, nomeLimp)

			conteudoVisual := fmt.Sprintf(`<div class={Conteiner%s}>
	<h1>Página %s ativa!</h1>
	<p>Modifique a lógica em "%s.hrp", os estilos em "%s.estilo.hrp" e o layout em "%s.html".</p>
</div>
`, nomeLimp, nomeLimp, nomeLimp, nomeLimp, nomeLimp)

			conteudoLogico := fmt.Sprintf(`de "web" importe sinal, importarHtml;
de "./%s.estilo.hrp" importe Conteiner%s;

exportar funcao Rota%s() {
	retorne importarHtml("./%s.html");
}
`, nomeLimp, nomeLimp, nomeLimp, nomeLimp)

			if err := os.WriteFile(caminhoEstilo, []byte(conteudoEstilo), 0644); err != nil {
				fmt.Fprintf(os.Stderr, "Erro ao criar estilo %s: %v\n", caminhoEstilo, err)
				os.Exit(1)
			}

			if err := os.WriteFile(caminhoVisual, []byte(conteudoVisual), 0644); err != nil {
				fmt.Fprintf(os.Stderr, "Erro ao criar visual %s: %v\n", caminhoVisual, err)
				os.Exit(1)
			}

			if err := os.WriteFile(caminhoLogico, []byte(conteudoLogico), 0644); err != nil {
				fmt.Fprintf(os.Stderr, "Erro ao criar lógica %s: %v\n", caminhoLogico, err)
				os.Exit(1)
			}

			injetarRotaEmRotasHrp(nomeLimp)

			fmt.Printf("✅ Tríade de arquivos da página '%s' criada com sucesso em '%s/':\n", nomeLimp, pastaPages)
			fmt.Printf("  - %s.hrp (Lógica e controle de sinais)\n", nomeLimp)
			fmt.Printf("  - %s.estilo.hrp (Estilos locais em português)\n", nomeLimp)
			fmt.Printf("  - %s.html (Marcação estrutural HTML)\n", nomeLimp)
			fmt.Println("\nA rota foi injetada em web/rotas/rotas.hrp como <Rota" + nomeLimp + " />.")
		},
	}
}

func comandoCrieComponente() *cobra.Command {
	return &cobra.Command{
		Use:   "componente [nome]",
		Short: "Gera um componente lógico e seu arquivo de estilizações híbridas em português",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			nome := args[0]
			nomeLimp := strings.Title(strings.ToLower(nome))

			pastaDest := "componentes"
			if _, err := os.Stat("web/componentes"); err == nil {
				pastaDest = "web/componentes"
			} else {
				os.MkdirAll("componentes", 0755)
			}

			caminhoLogico := filepath.Join(pastaDest, nomeLimp+".hrp")
			caminhoEstilo := filepath.Join(pastaDest, nomeLimp+".estilo.hrp")

			conteudoLogico := fmt.Sprintf(`de "web" importe sinal, h;
de "./%s.estilo.hrp" importe CaixaDe%s;

exportar funcao %s(props) {
	retorne <div class={CaixaDe%s}>
		<p>Componente '%s' ativo!</p>
		{props.children}
	</div>;
}
`, nomeLimp, nomeLimp, nomeLimp, nomeLimp, nomeLimp)

			conteudoEstilo := fmt.Sprintf(`exportar estilo CaixaDe%s {
	corDeFundo: "#f3f4f6";
	padding: "1rem";
	raio-medio: Verdadeiro;
	borda: "1px solid #e5e7eb";
}
`, nomeLimp)

			if err := os.WriteFile(caminhoLogico, []byte(conteudoLogico), 0644); err != nil {
				fmt.Fprintf(os.Stderr, "Erro ao criar componente %s: %v\n", caminhoLogico, err)
				os.Exit(1)
			}

			if err := os.WriteFile(caminhoEstilo, []byte(conteudoEstilo), 0644); err != nil {
				fmt.Fprintf(os.Stderr, "Erro ao criar estilo %s: %v\n", caminhoEstilo, err)
				os.Exit(1)
			}

			fmt.Printf("✅ Componente '%s' (Lógico e Estilos) gerado em: %s\n", nomeLimp, pastaDest)
		},
	}
}

func exibirSucessoCompleto(nomeProj, tipo string) {
	fmt.Printf("\n🎉 Sucesso! Projeto do tipo '%s' criado em ./%s\n", tipo, nomeProj)
	fmt.Println("Para rodar:")
	fmt.Printf("  cd %s\n", nomeProj)
	if tipo == "backend" {
		fmt.Println("  harpia executar main.hrp")
	} else {
		fmt.Println("  harpia compilar --alvo=web --entrada=main.hrp --saida=dist")
		fmt.Println("  harpia servir --diretorio=dist")
	}
}

func escreverArquivo(basePath, subPath, conteudo string) {
	fullPath := filepath.Join(basePath, subPath)
	err := os.WriteFile(fullPath, []byte(conteudo), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao criar arquivo %s: %v\n", fullPath, err)
		os.Exit(1)
	}
}

const conteudoAssetsReadmeMd = `# 🖼️ Pasta de Assets Estáticos (web/assets/)

Esta pasta centraliza todos os arquivos estáticos consumidos diretamente pelo navegador (imagens, fontes, favicons, manifestos visuais, etc.).

## 📂 Estrutura Recomendada

- **img/**: Imagens de uso geral do site (logos, banners, ilustrações, fotos).
- **fontes/**: Arquivos de fontes externas (.woff2, .ttf, .otf). Importe via bloco 'estilo' ou '@import url(...)'.
- **favicon/**: Ícones do site (favicon.ico, apple-touch-icon.png, etc.).

## ⚙️ Como Funciona no Build

Ao rodar 'harpia compilar --alvo=web' (ou 'harpia dev'), todo o conteúdo desta pasta é copiado automaticamente para 'dist/assets/', sem qualquer alteração. PNG/JPG podem ainda ser otimizados com a flag --otimizar-assets.

## 📌 Boas Práticas

1. **Nomenclatura**: Use kebab-case em português (ex: 'logo-principal.png', 'banner-promo.webp').
2. **Formatos Leves**: Prefira '.svg' para ilustrações vetoriais e '.webp' para fotos (a flag --otimizar-assets cuida disso).
3. **Sem Mistura com Código**: Esta pasta existe para isolar binários e mídias; nunca coloque '.hrp' aqui.
`

// injetarRotaEmRotasHrp adiciona o import e o uso da rotaRota<nome> dentro de web/rotas/rotas.hrp,
// caso ainda não exista. É idempotente: rodar duas vezes não duplica import nem uso.
func injetarRotaEmRotasHrp(nomeLimp string) {
	caminho := "web/rotas/rotas.hrp"
	conteudo, err := os.ReadFile(caminho)
	if err != nil {
		return
	}
	texto := string(conteudo)
	linhaImport := fmt.Sprintf("de \"../pages/%s.hrp\" importe Rota%s;", nomeLimp, nomeLimp)
	uso := fmt.Sprintf("<Rota%s />", nomeLimp)

	if strings.Contains(texto, linhaImport) {
		return
	}
	novo := texto
	if idx := strings.Index(novo, "exportar funcao RotaIndex()"); idx != -1 {
		novo = novo[:idx] + linhaImport + "\n\n" + strings.TrimLeft(novo[idx:], "\n")
	} else {
		novo += "\n" + linhaImport + "\n"
	}
	if !strings.Contains(novo, uso) {
		chave := "retorne <RotaInicio />;"
		novo = strings.Replace(novo, chave, "retorne "+uso+";", 1)
		if !strings.Contains(novo, uso) {
			novo += "\n\tuso adicional: " + uso + "\n"
		}
	}
	if err := os.WriteFile(caminho, []byte(novo), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Aviso: não foi possível auto-injetar em %s: %v\n", caminho, err)
	}
}

// criarAssetsPadrao cria a pasta web/assets/ e suas subpastas convencionais.
func criarAssetsPadrao(nomeProj string) {
	diretorios := []string{
		filepath.Join(nomeProj, "web", "assets"),
		filepath.Join(nomeProj, "web", "assets", "img"),
		filepath.Join(nomeProj, "web", "assets", "fontes"),
		filepath.Join(nomeProj, "web", "assets", "favicon"),
	}
	for _, dir := range diretorios {
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Erro ao criar pasta %s: %v\n", dir, err)
			os.Exit(1)
		}
	}
	escreverArquivo(nomeProj, filepath.Join("web", "assets", "README.md"), conteudoAssetsReadmeMd)
}

func comandoCrieModelo() *cobra.Command {
	return &cobra.Command{
		Use:   "modelo [nome]",
		Short: "Gera um novo modelo de domínio tipado",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			nome := args[0]
			nomeLimp := strings.Title(strings.ToLower(nome))

			pastaDest := filepath.Join("dominio", "modelos")
			if _, err := os.Stat("web"); err == nil {
				pastaDest = filepath.Join("dominio", "modelos")
			}
			os.MkdirAll(pastaDest, 0755)

			caminho := filepath.Join(pastaDest, strings.ToLower(nome)+".hrp")
			conteudo := fmt.Sprintf(`exportar classe %s {
	inicializar(self, id: Inteiro) {
		self.id = id
	}
}
`, nomeLimp)

			if err := os.WriteFile(caminho, []byte(conteudo), 0644); err != nil {
				fmt.Fprintf(os.Stderr, "Erro ao criar modelo %s: %v\n", caminho, err)
				os.Exit(1)
			}

			fmt.Printf("✅ Modelo '%s' gerado com sucesso em: %s\n", nomeLimp, caminho)
		},
	}
}

const conteudoComandosMd = `# 🎮 Guia de Comandos do CLI do Harpia

Este guia apresenta todos os comandos disponíveis na ferramenta de linha de comando (CLI) do Harpia, suas flags, necessidades e especificações de uso.

---

## 1. harpia executar
Executa um arquivo de código Harpia diretamente no interpretador ou inicia o console de desenvolvimento interativo.

* **Uso:**
  ` + "`" + `harpia executar [caminho-do-arquivo.hrp]` + "`" + `
* **Necessidades:**
  * Sem argumentos: Inicia o REPL (playground de teste interativo).
  * Com argumento: Executa o script síncrono imediatamente.

---

## 2. harpia checar
Realiza uma análise estática (linter sintático e semântico) no arquivo de entrada sem de fato executá-lo. Excelente para validar a consistência e integridade do seu código em pipelines de CI/CD ou antes de fazer commit.

* **Uso:**
  ` + "`" + `harpia checar [caminho-do-arquivo.hrp]` + "`" + `
* **Saída:**
  * Exibe avisos e erros detalhados de variáveis não declaradas, chamadas incorretas ou erros de sintaxe.

---

## 3. harpia compilar
Transpila o seu código fonte Harpia para outras plataformas (como a Web com suporte a Virtual DOM e JavaScript, ou targets nativos).

* **Uso:**
  ` + "`" + `harpia compilar --alvo=web --entrada=main.hrp --saida=dist` + "`" + `
* **Flags Principais:**
  * ` + "`" + `-a, --alvo` + "`" + `: Alvo da compilação (padrão: ` + "`" + `web` + "`" + `). Opções: ` + "`" + `web` + "`" + `, ` + "`" + `nativo` + "`" + `, ` + "`" + `wasm` + "`" + `.
  * ` + "`" + `-e, --entrada` + "`" + `: Arquivo principal/ponto de entrada do seu código (ex: ` + "`" + `main.hrp` + "`" + `).
  * ` + "`" + `-s, --saida` + "`" + `: Pasta onde os arquivos estáticos compilados (HTML, JS, CSS) serão gravados (padrão: ` + "`" + `dist` + "`" + `).

---

## 4. harpia servir
Inicia um servidor web local extremamente leve e rápido para hospedar os arquivos compilados e visualizar a sua aplicação SPA reativa diretamente no navegador.

* **Uso:**
  ` + "`" + `harpia servir --diretorio=dist` + "`" + `
* **Flags Principais:**
  * ` + "`" + `-d, --diretorio` + "`" + `: Pasta que contém os arquivos que serão servidos (padrão: ` + "`" + `dist` + "`" + `).
  * ` + "`" + `-p, --porta` + "`" + `: Porta na qual o servidor será escutado (padrão: ` + "`" + `8080` + "`" + `).

---

## 5. harpia novo
Inicializa uma nova estrutura de projeto baseada em arquiteturas de mercado prontas em português (Clean Architecture + DDD).

* **Subcomandos disponíveis:**
  * ` + "`" + `harpia novo monolito [nome-do-projeto]` + "`" + `: Estrutura completa de Frontend + Backend organizada em Clean Architecture e DDD.
  * ` + "`" + `harpia novo frontend [nome-do-projeto]` + "`" + `: Estrutura reativa voltada exclusivamente para o cliente SPA de alto desempenho.
  * ` + "`" + `harpia novo backend [nome-do-projeto]` + "`" + `: Estrutura enxuta de backend focada em APIs lógicas, conectores de banco e concorrência.

---

## 6. harpia crie
Assistente de geração rápida e assistida de novos arquivos estruturados de templates de código dentro de um projeto existente.

* **Subcomandos de criação:**
  * ` + "`" + `harpia crie rota [nome]` + "`" + `: Gera uma nova tela/página de roteamento SPA na pasta correspondente.
  * ` + "`" + `harpia crie componente [nome]` + "`" + `: Cria uma estrutura lógica e seu respectivo arquivo de estilo dinâmico (.hrp).
  * ` + "`" + `harpia crie modelo [nome]` + "`" + `: Gera um novo modelo de dados rico e tipado na camada de domínio.
`

const conteudoRotasReadmeMd = `# 🛣️ Pasta de Rotas (Roteamento SPA)

Esta pasta gerencia as telas, painéis e páginas principais da sua aplicação Single Page Application (SPA), controlando quais componentes e páginas são renderizados de acordo com o fluxo de navegação.

## 📁 O arquivo "rotas.hrp"
Este é o controlador mestre das rotas visuais da sua aplicação. Nele você pode importar as páginas da pasta "pages/" e os componentes de "componentes/" para decidir o que exibir na interface.

---

## 💡 Exemplos Práticos de Roteamento

### 1. Roteamento Simples Baseado em Estado (Sinais)
Você pode gerenciar qual tela exibir usando um sinal de texto reativo simples no arquivo "rotas.hrp":

` + "```" + `harpia
de "web" importe sinal, h;

exportar funcao RotaMestra() {
	var telaSinal = sinal("inicio");
	var tela = telaSinal[0];
	var irPara = telaSinal[1];

	retorne <div class="App">
		<nav>
			<button aoClicar={funcao() { irPara("inicio"); }}>Início</button>
			<button aoClicar={funcao() { irPara("sobre"); }}>Sobre</button>
		</nav>

		<main>
			<se condicao={tela() == "inicio"}>
				<div class="PaginaInicio">
					<h1>Página Inicial</h1>
				</div>
			</se>
			<se condicao={tela() == "sobre"}>
				<div class="PaginaSobre">
					<h1>Sobre Nós</h1>
				</div>
			</se>
		</main>
	</div>;
}
` + "```" + `

### 2. Importação e Renderização de Páginas HTML
Recomenda-se manter o código visual separado na pasta "web/pages/" e carregá-lo dinamicamente de dentro das suas rotas com "importarHtml":

` + "```" + `harpia
de "web" importe sinal, importarHtml;

exportar funcao RotaIndex() {
	# Carrega a página estática de forma ultra veloz e acopla a lógica reativa local
	retorne importarHtml("../pages/Layout.html");
}
` + "```" + `
`
