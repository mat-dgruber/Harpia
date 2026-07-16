// Package stdlib agrupa, orquestra e gerencia a inicialização e o registro de todos os módulos nativos
// que compõem a biblioteca padrão da linguagem Portuscript.
//
// O pacote funciona como o agregador central e declarativo de componentes. Ao ser importado em qualquer
// parte do programa (normalmente no pacote cmd de execução), ele aciona os mecanismos de inicialização
// de seus subpacotes via funções init().
package stdlib

import (
	// A importação blank (_) do pacote colorize registra funções utilitárias para estilização,
	// formatação e coloração de saídas de texto no console do usuário de forma nativa.
	_ "github.com/natanfeitosa/portuscript/stdlib/colorize"

	// A importação blank (_) de embutidos injeta as palavras-chave globais e funções integradas essenciais
	// diretamente na tabela global de símbolos (como 'escreva', 'leia', 'tamanho', etc.).
	_ "github.com/natanfeitosa/portuscript/stdlib/embutidos"

	// A importação blank (_) de matematica expõe recursos e operações matemáticas nativas de alta precisão
	// (como calculo de teto, piso, potência, raiz quadrada, absoluto, etc.).
	_ "github.com/natanfeitosa/portuscript/stdlib/matematica"

	// A importação blank (_) de sistema expõe dados e comandos para interação direta com as variáveis de ambiente,
	// argumentos CLI e propriedades do sistema operacional hospedeiro.
	_ "github.com/natanfeitosa/portuscript/stdlib/sistema"

	// A importação blank (_) do pacote soquete expõe facilidades de rede de baixo nível, permitindo que scripts
	// Portuscript abram conexões de rede TCP/IP e criem servidores/clientes básicos diretamente no terminal.
	_ "github.com/natanfeitosa/portuscript/stdlib/soquete"

	// Novos módulos da Fase 3
	_ "github.com/natanfeitosa/portuscript/stdlib/arquivos"
	_ "github.com/natanfeitosa/portuscript/stdlib/cripto"
	_ "github.com/natanfeitosa/portuscript/stdlib/http"
	_ "github.com/natanfeitosa/portuscript/stdlib/json"
	_ "github.com/natanfeitosa/portuscript/stdlib/bd"
	_ "github.com/natanfeitosa/portuscript/stdlib/yaml"
	_ "github.com/natanfeitosa/portuscript/stdlib/xml"
	_ "github.com/natanfeitosa/portuscript/stdlib/logs"
	_ "github.com/natanfeitosa/portuscript/stdlib/metricas"
	_ "github.com/natanfeitosa/portuscript/stdlib/esquema"
	_ "github.com/natanfeitosa/portuscript/stdlib/tarefas"
	_ "github.com/natanfeitosa/portuscript/stdlib/ffi"
)
