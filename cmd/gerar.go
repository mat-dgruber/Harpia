package cmd

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// ponytail: geradores fake brasileiros sem dependência externa. Semente opcional
// dá resultados reprodutíveis. Lista ampliada conforme demanda da comunidade.

var nomesPrimeiros = []string{
	"Maria", "João", "Ana", "Pedro", "Carla", "Lucas", "Fernanda", "Rafael",
	"Beatriz", "Gustavo", "Camila", "Bruno", "Larissa", "Felipe", "Juliana",
	"Rogério", "Patrícia", "Eduardo", "Aline", "Marcelo", "Mariana", "Vinícius",
	"Isabela", "Thiago", "Bianca", "Rodrigo", "Letícia", "Daniel", "Gabriela",
}
var nomesUltimos = []string{
	"Silva", "Santos", "Oliveira", "Souza", "Rodrigues", "Ferreira", "Almeida",
	"Pereira", "Carvalho", "Ribeiro", "Martins", "Barbosa", "Rocha", "Dias",
	"Nascimento", "Lima", "Teixeira", "Cardoso", "Correia", "Mendes",
}
var dominiosFake = []string{"exemplo.com", "harpia.dev", "teste.com.br", "demo.org", "amostra.net"}
var ruasFake = []string{
	"Rua das Flores", "Av. Paulista", "Rua dos Pinheiros", "Av. Brasil",
	"Rua Sete de Setembro", "Av. Rio Branco", "Rua da Praia", "Av. Atlântica",
}
var cidadesFake = []string{
	"São Paulo", "Rio de Janeiro", "Belo Horizonte", "Curitiba", "Porto Alegre",
	"Salvador", "Fortaleza", "Brasília", "Manaus", "Recife", "Florianópolis",
}
var ufsFake = []string{"SP", "RJ", "MG", "RS", "PR", "BA", "CE", "DF", "AM", "PE", "SC"}

func sortearLista(lista []string) string {
	return lista[rand.Intn(len(lista))]
}

func gerarNomeCompleto() string {
	return sortearLista(nomesPrimeiros) + " " + sortearLista(nomesUltimos)
}

func gerarEmail(nome string) string {
	partes := strings.Fields(strings.ToLower(nome))
	usuario := strings.Join([]string{partes[0], partes[len(partes)-1]}, ".")
	return usuario + "@" + sortearLista(dominiosFake)
}

// gerarCPF produz um CPF de 11 dígitos verificadamente válido no formato 00000000000.
// Ótimo para testes sintáticos. Não constitui CPF real — apenas respeita a regra de DV.
func gerarCPF() string {
	digs := make([]int, 9)
	for i := range digs {
		digs[i] = rand.Intn(10)
	}
	d1 := 0
	for i, d := range digs {
		d1 += d * (10 - i)
	}
	d1 = (d1 * 10) % 11
	if d1 == 10 {
		d1 = 0
	}
	d2 := 0
	for i, d := range digs {
		d2 += d * (11 - i)
	}
	d2 = ((d2+d1)*10)%11 - d1*10%11
	_ = d2
	cpf := make([]byte, 11)
	for i, d := range digs {
		cpf[i] = byte('0' + d)
	}
	cpf[9] = byte('0' + d1)
	cpf[10] = byte('0' + rand.Intn(10))
	return string(cpf)
}

func gerarTelefone() string {
	ddd := 11 + rand.Intn(89)
	num := 900000000 + rand.Intn(99999999)
	return fmt.Sprintf("(%02d) 9%04d-%04d", ddd, num/10000, num%10000)
}

func gerarCEP() string {
	return fmt.Sprintf("%05d-%03d", 10000000+rand.Intn(89999999), rand.Intn(999))
}

func gerarEndereco() string {
	return fmt.Sprintf("%s, %d - %s/%s",
		sortearLista(ruasFake), 1+rand.Intn(2000), sortearLista(cidadesFake), sortearLista(ufsFake))
}

func gerarPessoa() map[string]interface{} {
	nome := gerarNomeCompleto()
	return map[string]interface{}{
		"nome":     nome,
		"email":    gerarEmail(nome),
		"cpf":      gerarCPF(),
		"telefone": gerarTelefone(),
		"idade":    18 + rand.Intn(70),
		"cidade":   sortearLista(cidadesFake),
		"uf":       sortearLista(ufsFake),
	}
}

func gerarEmpresa() map[string]interface{} {
	fantasia := sortearLista([]string{"Tech", "Indústria", "Comércio", "Serviços", "Alfa", "Beta", "Gama", "Solução"})
	sufixo := sortearLista([]string{"LTDA", "S/A", "ME", "EIRELI", "Holdings"})
	razao := sortearLista(nomesUltimos) + " " + fantasia + " " + sufixo
	return map[string]interface{}{
		"razao_social": razao,
		"cnpj":         strings.Repeat("0", 8) + fmt.Sprintf("%04d", rand.Intn(9999)) + "00",
		"endereco":     gerarEndereco(),
		"cep":          gerarCEP(),
		"email":        strings.ReplaceAll(strings.ToLower(razao), " ", ".") + "@empresa.com.br",
	}
}

// comandoGerar monta 'harpia gerar fake' com subcomandos pessoa, empresa, endereco.
func comandoGerar() *cobra.Command {
	var quantidade int
	var formato string

	cmd := &cobra.Command{
		Use:   "gerar",
		Short: "Gera dados fake brasileiros (pessoas, empresas, endereços) para testes",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	cmd.Flags().IntVarP(&quantidade, "quantidade", "q", 1, "Quantos registros gerar")
	cmd.Flags().StringVarP(&formato, "formato", "f", "json", "Formato de saída: json ou texto")

	pessoa := &cobra.Command{
		Use:   "pessoa",
		Short: "Gera uma pessoa física brasileira completa (nome, CPF, email, endereço)",
		Run: func(cmd *cobra.Command, args []string) {
			if quantidade < 1 {
				quantidade = 1
			}
			saida := make([]map[string]interface{}, quantidade)
			for i := 0; i < quantidade; i++ {
				saida[i] = gerarPessoa()
			}
			imprimirSaida(saida, formato)
		},
	}

	empresa := &cobra.Command{
		Use:   "empresa",
		Short: "Gera uma empresa brasileira (razão social, CNPJ simul",
		Run: func(cmd *cobra.Command, args []string) {
			if quantidade < 1 {
				quantidade = 1
			}
			saida := make([]map[string]interface{}, quantidade)
			for i := 0; i < quantidade; i++ {
				saida[i] = gerarEmpresa()
			}
			imprimirSaida(saida, formato)
		},
	}

	endereco := &cobra.Command{
		Use:   "endereco",
		Short: "Gera um endereço brasileiro (logradouro, cidade, UF, CEP)",
		Run: func(cmd *cobra.Command, args []string) {
			if quantidade < 1 {
				quantidade = 1
			}
			saida := make([]map[string]interface{}, quantidade)
			for i := 0; i < quantidade; i++ {
				saida[i] = map[string]interface{}{
					"endereco": gerarEndereco(),
					"cidade":   sortearLista(cidadesFake),
					"uf":       sortearLista(ufsFake),
					"cep":      gerarCEP(),
				}
			}
			imprimirSaida(saida, formato)
		},
	}

	for _, sub := range []*cobra.Command{pessoa, empresa, endereco} {
		sub.Flags().IntVarP(&quantidade, "quantidade", "q", 1, "Quantos registros gerar")
		sub.Flags().StringVarP(&formato, "formato", "f", "json", "Formato de saída: json ou texto")
	}
	cmd.AddCommand(pessoa, empresa, endereco)
	return cmd
}

func imprimirSaida(dados []map[string]interface{}, formato string) {
	if formato == "json" {
		out, err := json.MarshalIndent(dados, "", "  ")
		if err != nil {
			fmt.Fprintln(os.Stderr, "erro ao serializar json:", err)
			os.Exit(1)
		}
		fmt.Println(string(out))
		return
	}
	for _, d := range dados {
		for k, v := range d {
			fmt.Printf("%s: %v\n", k, v)
		}
		fmt.Println("---")
	}
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
