// Package stdlib agrupa, orquestra e gerencia a inicialização e o registro de todos os módulos nativos
// que compõem a biblioteca padrão da linguagem Harpia.
//
// O pacote funciona como o agregador central e declarativo de componentes. Ao ser importado em qualquer
// parte do programa (normalmente no pacote cmd de execução), ele aciona os mecanismos de inicialização
// de seus subpacotes via funções init().
package stdlib

import (
	// A importação blank (_) do pacote colorize registra funções utilitárias para estilização,
	// formatação e coloração de saídas de texto no console do usuário de forma nativa.
	_ "github.com/mat-dgruber/Harpia/stdlib/colorize"

	// A importação blank (_) de embutidos injeta as palavras-chave globais e funções integradas essenciais
	// diretamente na tabela global de símbolos (como 'escreva', 'leia', 'tamanho', etc.).
	_ "github.com/mat-dgruber/Harpia/stdlib/embutidos"

	// A importação blank (_) de matematica expõe recursos e operações matemáticas nativas de alta precisão
	// (como calculo de teto, piso, potência, raiz quadrada, absoluto, etc.).
	_ "github.com/mat-dgruber/Harpia/stdlib/matematica"

	// A importação blank (_) de sistema expõe dados e comandos para interação direta com as variáveis de ambiente,
	// argumentos CLI e propriedades do sistema operacional hospedeiro.
	_ "github.com/mat-dgruber/Harpia/stdlib/sistema"

	// A importação blank (_) do pacote soquete expõe facilidades de rede de baixo nível, permitindo que scripts
	// Harpia abram conexões de rede TCP/IP e criem servidores/clientes básicos diretamente no terminal.
	_ "github.com/mat-dgruber/Harpia/stdlib/soquete"

	// Novos módulos da Fase 3
	_ "github.com/mat-dgruber/Harpia/stdlib/arquivos"
	_ "github.com/mat-dgruber/Harpia/stdlib/cripto"
	_ "github.com/mat-dgruber/Harpia/stdlib/http"
	_ "github.com/mat-dgruber/Harpia/stdlib/json"
	_ "github.com/mat-dgruber/Harpia/stdlib/bd"
	_ "github.com/mat-dgruber/Harpia/stdlib/yaml"
	_ "github.com/mat-dgruber/Harpia/stdlib/xml"
	_ "github.com/mat-dgruber/Harpia/stdlib/logs"
	_ "github.com/mat-dgruber/Harpia/stdlib/metricas"
	_ "github.com/mat-dgruber/Harpia/stdlib/esquema"
	_ "github.com/mat-dgruber/Harpia/stdlib/resiliencia"
	_ "github.com/mat-dgruber/Harpia/stdlib/tarefas"
	_ "github.com/mat-dgruber/Harpia/stdlib/ffi"
	_ "github.com/mat-dgruber/Harpia/stdlib/ia"
)
