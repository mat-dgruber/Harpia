// Package ptst implementa o núcleo estrutural e de runtime (Máquina Virtual, Tipos e Símbolos)
// do interpretador Harpia.
//
// Este pacote define a modelagem polimórfica de objetos dinâmicos (Objeto, Tipo), as tabelas de escopo,
// tratamento de erros ricos em português e toda a infraestrutura matemática e lógica executável de runtime.
package ptst

// Nota Arquitetural / Histórico:
// Os métodos originais de inicialização legados contidos neste arquivo foram totalmente substituídos
// e migrados de forma madura para os arquivos executores centrais `contexto.go`, `interpretador.go` e `executar.go`.
// Este arquivo permanece reservado para agrupamento global conceitual de metadados e documentações de runtime.
