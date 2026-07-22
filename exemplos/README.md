# Galeria de Exemplos (`exemplos` do Harpia)

Este diretório abriga um catálogo abrangente de **Scripts Demonstrativos** (`.hrp`), **Módulos de Extensões** e **Aplicações Frontend Reativas** desenhados para servir de guia de aprendizado pedagógico prático das capacidades do **Harpia**.

Aqui você encontrará desde o clássico Olá Mundo até programas de redes assíncronas, gerenciamento de estado reativo e extensões nativas compiladas de alto desempenho.

---

## 📖 Índice

1. [Relação Geral de Demonstrações](#-relação-geral-de-demonstrações)
2. [Recursos Avançados Frontend Reativos (`frontend/`)](#-recursos-avançados-frontend-reativos-frontend)
3. [Recurso Avançado: Conectividade de Redes (`soquetes`)](#-recurso-avançado-conectividade-de-redes-soquetes)
4. [Recurso Avançado: Extensões Dinâmicas Go (`modExterno`)](#-recurso-avançado-extensões-dinâmicas-go-modexterno)
5. [Como Executar os Exemplos](#-como-executar-os-exemplos)

---

## 📋 Relação Geral de Demonstrações

Os scripts estão organizados por nível de complexidade e categoria técnica de recursos:

| Arquivo de Exemplo         | Categoria Técnica  | Recurso de Linguagem Demonstrado                                                |
| :------------------------- | :----------------- | :------------------------------------------------------------------------------ |
| **`olaMundo.hrp`**         | Introdução Básica  | Uso da função de saída de console `imprima()`.                                  |
| **`variaveis.hrp`**        | Fundamentos        | Declaração mutável (`var`) vs. imutável constante (`const`).                    |
| **`aritmetica.hrp`**       | Matemática         | Operadores matemáticos básicos e precedência algébrica.                         |
| **`booleanos.hrp`**        | Lógica             | Operadores booleanos de curto-circuito (`e`, `ou`) e comparadores (`==`, `!=`). |
| **`condicionais.hrp`**     | Controle de Fluxo  | Desvios lógicos de decisão condicionais (`se`, `senao se`, `senao`).            |
| **`entradaSaida.hrp`**     | Interatividade     | Coleta de dados via teclado com `leia()` e coerções primitivas.                 |
| **`funcao.hrp`**           | Sub-Rotinas        | Declaração e chamada de funções parametrizadas com retorno.                     |
| **`fatorial.hrp`**         | Recursividade      | Funções recursivas completas calculando o fatorial de um número.                |
| **`acessaMembros.hrp`**    | POO e Reflexão     | Acesso dinâmico a propriedades de classes via atributo especial `__doc__`.      |
| **`lacosDeRepeticao.hrp`** | Iteração           | Loops condicionais `enquanto` aninhados gerando tabuadas completas.             |
| **`testeFor.hrp`**         | Iteração           | Loops iteradores baseados na sintaxe para-em (`para num em colecao`).           |
| **`importacao.hrp`**       | Modulos / Stdlib   | Importações desestruturadas e qualificadas do core da stdlib (`matematica`).    |
| **`importacaoRelativa.hrp`**| Modulos Locais    | Importações relativas locais e leitura de metadados (`__arquivo__`, `__nome__`).|
| **`atm.hrp`**              | Aplicação Completa | Caixa Eletrônico simulado com menu, saques, depósitos e interrupção `pare`.     |

---

## 💻 Recursos Avançados Frontend Reativos (`frontend/`)

A subpasta **`frontend/`** demonstra a capacidade do Harpia como uma linguagem voltada para desenvolvimento web moderno e reativo de página única (SPA). Essas aplicações rodam diretamente no navegador, compiladas para JavaScript e estruturadas com reconciliação de Virtual DOM:

- **`contador/main.hrp`**: Demonstra a primitiva fundamental de reatividade `sinal` para criar um contador simples e interativo. Aplica classes e estilizações de design por meio do bloco `definirEstilo`.
- **`formulario/main.hrp`**: Implementa um formulário de contato robusto que calcula validações de tamanho de nome e estrutura de email em tempo real através da primitiva `derivado`, desabilitando o botão de envio condicionalmente.
- **`tarefas/main.hrp`**: Uma aplicação completa de lista de tarefas (Todo list) que demonstra a manipulação reativa de arrays dinâmicos, adicionando e removendo tarefas, renderizando listas dinamicamente através do bloco sintático `<para item em lista={...}>`.

---

## 🔌 Recurso Avançado: Conectividade de Redes (`soquetes`)

A subpasta **`soquetes/`** abriga uma demonstração prática de programação de rede soquete TCP/IP orientada a eventos assíncronos:

- **`servidor.hrp`**: Inicializa um socket IPv4 TCP, define-o como não-bloqueante, inicia escuta na porta `3000` e atua como um servidor de eco (echo server) que recebe e devolve bytes ao cliente conectado.
- **`cliente.hrp`**: Conecta ao servidor local, solicita uma mensagem textual do console via `leia()`, envia-a encapsulada em bytes e imprime a resposta retornada.

---

## ⚙️ Recurso Avançado: Extensões Dinâmicas Go (`modExterno`)

O Harpia permite que programadores criem módulos compilados nativamente em Go que são carregados de forma assíncrona como plug-ins compartilhados de alto desempenho (`.so`):

- **`modExterno/main.go`**: Define a extensão nativa Go contendo o método de performance `.exiba()`.
- **`modExterno/main.hrp`**: Demonstra a importação e o uso da extensão.
- **`modExterno/README.md`**: Explica as instruções e comandos para compilar a extensão.

---

## 🚀 Como Executar os Exemplos

### 1. Exemplos de Terminal / Console:
Para rodar qualquer um dos scripts de demonstração locais, utilize o interpretador oficial acionando o subcomando `executar` (ou o atalho `exec`) a partir do diretório raiz do projeto:

```bash
# Executa o exemplo do Caixa Eletrônico interativo
harpia executar exemplos/atm.hrp

# Executa o exemplo de fatorial recursivo
harpia executar exemplos/fatorial.hrp
```

### 2. Exemplos de Frontend Web:
Para rodar e visualizar qualquer exemplo contido na pasta `frontend/`, compile-o para a plataforma web:

```bash
# Navegue até a pasta do exemplo (ex: contador)
cd exemplos/frontend/contador

# Compile a aplicação para o diretório de distribuição web
harpia compilar --alvo=web --entrada=main.hrp --saida=dist

# Abra o arquivo index.html gerado no seu navegador
open dist/index.html
```
