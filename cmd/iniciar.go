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
		Short:   "Inicializa uma nova estrutura de projeto Portuscript",
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
			}

			for _, dir := range diretorios {
				if err := os.MkdirAll(dir, 0755); err != nil {
					fmt.Fprintf(os.Stderr, "Erro ao criar pasta %s: %v\n", dir, err)
					os.Exit(1)
				}
			}

			// Escreve os arquivos padrões do monolito
			escreverArquivo(nomeProj, filepath.Join("dominio", "entidades", "usuario.ptst"), `
# Entidade de domínio tipada e segura
exportar classe Usuario {
	inicializar(self, id: Inteiro, nome: Texto, email: Texto) {
		self.id = id
		self.nome = nome
		self.email = email
	}
}
`)

			escreverArquivo(nomeProj, filepath.Join("infra", "banco", "conexao.ptst"), `
de "bd" importe conectar;

exportar funcao obterBanco() {
	# ponytail: SQLite embarcado de forma nativa e segura
	retorne conectar("sqlite", "dados.db");
}
`)

			escreverArquivo(nomeProj, filepath.Join("web", "rotas", "index.ptst"), `
de "web" importe sinal, importarHtml;

exportar funcao RotaIndex() {
	var [contador, setContador] = sinal(0);
	retorne importarHtml("../componentes/Layout.html");
}
`)

			escreverArquivo(nomeProj, filepath.Join("web", "componentes", "Layout.html"), `
<div class="p-6 max-w-lg mx-auto bg-white rounded-xl shadow-md space-y-4">
	<h1 class="text-2xl font-bold text-gray-900">Meu Monolito Portuscript</h1>
	<p class="text-gray-500">Desenvolvido em português com Clean Architecture e DDD.</p>
	<div class="contador-secao">
		<p>Cliques: <strong>{contador()}</strong></p>
		<button aoClicar={funcao() { setContador(contador() + 1); }} class="bg-blue-500 text-white px-4 py-2 rounded">
			Incrementar
		</button>
	</div>
</div>
`)

			escreverArquivo(nomeProj, "main.ptst", `
de "web" importe montar;
de "./web/rotas/index.ptst" importe RotaIndex;

funcao Aplicacao() {
	retorne <div class="App">
		<RotaIndex />
	</div>;
}

montar(Aplicacao, Nulo);
`)

			escreverArquivo(nomeProj, filepath.Join("testes", "usuario_test.ptst"), `
de "assegura" importe assegure;
de "../dominio/entidades/usuario.ptst" importe Usuario;

testar "deve instanciar um usuario de dominio com tipos corretos" {
	var user = nova Usuario(1, "Natan", "natan@portuscript.org")
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
			}

			for _, dir := range diretorios {
				if err := os.MkdirAll(dir, 0755); err != nil {
					fmt.Fprintf(os.Stderr, "Erro ao criar pasta %s: %v\n", dir, err)
					os.Exit(1)
				}
			}

			escreverArquivo(nomeProj, filepath.Join("dominio", "entidades", "produto.ptst"), `
exportar classe Produto {
	inicializar(self, id: Inteiro, nome: Texto, preco: Decimal) {
		self.id = id
		self.nome = nome
		self.preco = preco
	}
}
`)

			escreverArquivo(nomeProj, "main.ptst", `
de "http" importe Servidor;
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
			}

			for _, dir := range diretorios {
				if err := os.MkdirAll(dir, 0755); err != nil {
					fmt.Fprintf(os.Stderr, "Erro ao criar pasta %s: %v\n", dir, err)
					os.Exit(1)
				}
			}

			escreverArquivo(nomeProj, filepath.Join("web", "rotas", "index.ptst"), `
de "web" importe sinal, h;

exportar funcao RotaIndex() {
	var [nome, setNome] = sinal("");

	retorne <div class="p-6 font-sans">
		<h1 class="text-xl font-bold">Olá do Cliente SPA!</h1>
		<input ligar={nome} placeholder="Escreva seu nome..." class="border p-2 rounded m-2" />
		<se condicao={nome() != ""}>
			<p class="text-green-600">Seja muito bem-vindo, {nome()}!</p>
		</se>
	</div>;
}
`)

			escreverArquivo(nomeProj, "main.ptst", `
de "web" importe montar;
de "./web/rotas/index.ptst" importe RotaIndex;

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

			caminho := filepath.Join(pastaDest, strings.ToLower(nome)+".ptst")
			conteudo := fmt.Sprintf(`de "web" importe sinal, h;

exportar funcao Rota%s() {
	retorne <div class="p-6">
		<h1 class="text-2xl font-bold">Página %s</h1>
		<p>Gerado automaticamente com o assistente do Portuscript.</p>
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

			caminhoLogico := filepath.Join(pastaDest, nomeLimp+".ptst")
			caminhoEstilo := filepath.Join(pastaDest, nomeLimp+".estilo.ptst")

			conteudoLogico := fmt.Sprintf(`de "web" importe sinal, h;
de "./%s.estilo.ptst" importe CaixaDe%s;

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
		fmt.Println("  portuscript executar main.ptst")
	} else {
		fmt.Println("  portuscript compilar --alvo=web --entrada=main.ptst --saida=dist")
		fmt.Println("  portuscript servir --diretorio=dist")
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

			caminho := filepath.Join(pastaDest, strings.ToLower(nome)+".ptst")
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
