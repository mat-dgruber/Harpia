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

			escreverArquivo(nomeProj, filepath.Join("dominio", "README.md"), `# Camada de Domínio

Esta é a camada mais importante do sistema. Ela abriga a lógica central de negócios (regras que definem o sistema) e é totalmente independente de banco de dados, APIs ou frameworks.

- **entidades/**: Modelos de dados ricos e tipados (.hrp).
- **repositorios/**: Interfaces/portas de comunicação para persistência de dados (.hrp).
`)

			escreverArquivo(nomeProj, filepath.Join("infra", "banco", "conexao.hrp"), `de "bd" importe conectar;

exportar funcao obterBanco() {
	# ponytail: SQLite embarcado de forma nativa e segura
	retorne conectar("sqlite", "dados.db");
}
`)

			escreverArquivo(nomeProj, filepath.Join("infra", "README.md"), `# Camada de Infraestrutura

Aqui residem os adaptadores concretos e tecnologias externas que suportam o domínio.

- **banco/**: Arquivos de conexão, drivers e implementações específicas de consultas (ex: SQLite em .hrp).
`)

			escreverArquivo(nomeProj, filepath.Join("web", "rotas", "index.hrp"), `de "web" importe sinal, importarHtml;
de "../componentes/estilos.hrp" importe Aplicacao, BotaoPrincipal;

exportar funcao RotaIndex() {
	var contadorSinal = sinal(0);
	var contador = contadorSinal[0];
	var setContador = contadorSinal[1];
	retorne importarHtml("../componentes/Layout.html");
}
`)

			escreverArquivo(nomeProj, filepath.Join("web", "README.md"), `# Camada Web / Interface

Contém a camada visual, componentes de interface Virtual DOM, folhas de estilo e gerenciamento de rotas.

- **componentes/**: Blocos visuais reutilizáveis em HTML e estilos dinâmicos (.hrp).
- **rotas/**: Telas, dashboards e páginas de roteamento SPA (.hrp).
`)

			escreverArquivo(nomeProj, filepath.Join("web", "componentes", "estilos.hrp"), `exportar estilo Aplicacao {
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

			escreverArquivo(nomeProj, filepath.Join("web", "componentes", "Layout.html"), `<div class={Aplicacao}>
	<h1>🦅 Olá, Mundo Harpia!</h1>
	<p>Desenvolvido em português com Clean Architecture e DDD.</p>

	<div style="margin-top: 30px; padding: 20px; background: #f3f4f6; border-radius: 8px; display: inline-block;">
		<p>Cliques reativos: <strong>{contador()}</strong></p>
		<button aoClicar={funcao() { setContador(contador() + 1); }} class={BotaoPrincipal}>
			Incrementar
		</button>
	</div>
</div>
`)

			escreverArquivo(nomeProj, filepath.Join("docs", "arquitetura.md"), `# Arquitetura do Projeto Harpia

Este projeto foi estruturado utilizando **Clean Architecture** (Arquitetura Limpa) combinada com conceitos de **DDD** (Domain-Driven Design) em Português.

## Camadas do Sistema:
1. **dominio/**: Contém a lógica de negócios e entidades essenciais (independente de frameworks ou banco de dados).
2. **infra/**: Adaptadores de dados, repositórios concretos e conexões (como banco de dados SQLite).
3. **web/**: Roteamento, componentes de interface, visual e controladores.
`)

			escreverArquivo(nomeProj, filepath.Join("docs", "comandos.md"), conteudoComandosMd)

			escreverArquivo(nomeProj, "main.hrp", `de "web" importe montar;
de "./web/rotas/index.hrp" importe RotaIndex;

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

			escreverArquivo(nomeProj, filepath.Join("dominio", "README.md"), `# Domínio do Backend

Camada central onde estão as regras de negócio puras (entidades e modelos de domínio), sem acoplamento com banco de dados ou protocolo de transporte (HTTP/gRPC).
`)

			escreverArquivo(nomeProj, filepath.Join("infra", "README.md"), `# Infraestrutura do Backend

Adaptadores de entrada/saída, conexões de banco de dados, drivers e clientes de APIs externas.
`)

			escreverArquivo(nomeProj, filepath.Join("docs", "arquitetura.md"), `# Arquitetura de Backend Harpia

Estrutura limpa voltada para o desenvolvimento de APIs robustas, microsserviços e concorrência orientada a eventos.

- **dominio/**: Entidades de negócio e contratos de repositório.
- **infra/**: Conectores de banco de dados e APIs externas.
- **main.hrp**: Ponto de entrada que inicializa o servidor de microsserviço/API HTTP.
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

			escreverArquivo(nomeProj, filepath.Join("web", "rotas", "index.hrp"), `de "web" importe sinal, h;
de "../componentes/estilos.hrp" importe Aplicacao, EntradaTexto;

exportar funcao RotaIndex() {
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

			escreverArquivo(nomeProj, filepath.Join("web", "componentes", "estilos.hrp"), `exportar estilo Aplicacao {
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

			escreverArquivo(nomeProj, filepath.Join("web", "README.md"), `# Web / Frontend

Esta pasta abriga os elements de renderização dinâmica no navegador (SPA), como rotas e componentes reutilizáveis baseados em Sinais Reativos e Virtual DOM.

- **componentes/**: Blocos visuais reutilizáveis em HTML e estilos dinâmicos (.hrp).
- **rotas/**: Telas, painéis e páginas de navegação com Sinais Reativos de estado (.hrp).
`)

			escreverArquivo(nomeProj, filepath.Join("docs", "arquitetura.md"), `# Arquitetura Frontend Harpia

Uma estrutura reativa cliente puramente SPA (Single Page Application) de altíssimo desempenho, alimentada pelo motor Virtual DOM do Harpia.

- **web/componentes/**: Blocos de interface declarativos em HTML e arquivos de estilo (.hrp / .html).
- **web/rotas/**: Telas, painéis e páginas de navegação com Sinais Reativos de estado.
- **main.hrp**: Ponto de entrada que monta o componente raiz no documento.
`)

			escreverArquivo(nomeProj, filepath.Join("docs", "comandos.md"), conteudoComandosMd)

			escreverArquivo(nomeProj, "main.hrp", `de "web" importe montar;
de "./web/rotas/index.hrp" importe RotaIndex;

montar(RotaIndex, Nulo);
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
		Short: "Gera uma nova rota de SPA no diretório correspondente",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			nome := args[0]
			nomeLimp := strings.Title(strings.ToLower(nome))

			// Detecta se existe a pasta web/rotas ou rotas
			pastaDest := "rotas"
			if _, err := os.Stat("web/rotas"); err == nil {
				pastaDest = "web/rotas"
			} else {
				os.MkdirAll("rotas", 0755)
			}

			caminho := filepath.Join(pastaDest, strings.ToLower(nome)+".hrp")
			conteudo := fmt.Sprintf(`de "web" importe sinal, h;

exportar funcao Rota%s() {
	retorne <div class="p-6">
		<h1 class="text-2xl font-bold">Página %s</h1>
		<p>Gerado automaticamente com o assistente do Harpia.</p>
	</div>;
}
`, nomeLimp, nomeLimp)

			if err := os.WriteFile(caminho, []byte(conteudo), 0644); err != nil {
				fmt.Fprintf(os.Stderr, "Erro ao criar rota %s: %v\n", caminho, err)
				os.Exit(1)
			}

			fmt.Printf("✅ Rota '%s' gerada com sucesso em: %s\n", nomeLimp, caminho)
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

