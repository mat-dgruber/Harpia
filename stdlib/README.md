# Biblioteca Padrão (`stdlib` do Harpia)

A **Biblioteca Padrão** (do inglês, *Standard Library* ou simplesmente `stdlib`) do **Harpia** é um conjunto de módulos utilitários robustos implementados diretamente em Go. Eles são expostos nativamente para os scripts Harpia de forma embutida ou via importação explícita.

Estes módulos estendem as capacidades básicas da linguagem, permitindo que os programadores realizem operações matemáticas avançadas, coletem dados do sistema operacional hospedeiro, estilizem saídas de terminal com cores e até gerenciem conectividade básica de redes de computadores.

---

## 📖 Índice

1. [Arquitetura de Registro Automático](#-arquitetura-de-registro-automático)
2. [Módulo Central: `embutidos`](#-módulo-central-embutidos)
3. [Módulo: `matematica`](#-módulo-matematica)
4. [Módulo: `sistema`](#-módulo-sistema)
5. [Módulo: `colorize`](#-módulo-colorize)
6. [Módulo: `soquete`](#-módulo-soquete)
7. [Exemplo Completo de Uso de Módulos](#-exemplo-completo-de-uso-de-módulos)

---

## 🏗️ Arquitetura de Registro Automático

A integração de novos pacotes em Go com a VM do Harpia foi desenhada para ser o mais modular e desacoplada possível. 

No arquivo agregador central `stdlib.go`, é feito o uso do mecanismo de **importação anônima** (ou blank import `_`):

```go
package stdlib

import (
	_ "github.com/mat-dgruber/Harpia/stdlib/colorize"
	_ "github.com/mat-dgruber/Harpia/stdlib/embutidos"
	_ "github.com/mat-dgruber/Harpia/stdlib/matematica"
	_ "github.com/mat-dgruber/Harpia/stdlib/sistema"
	_ "github.com/mat-dgruber/Harpia/stdlib/soquete"
)
```

### Como e Por que Funciona:
- **Funções `init()`**: Cada subpacote declara uma função especial `init()` em seu código Go. Esta função é executada uma única vez, de forma prioritária, assim que o interpretador carrega o pacote correspondente.
- **Tabela de Módulos Globais**: Dentro do `init()`, cada módulo se autopopula com suas funções nativas e constantes, e se registra na máquina virtual chamando a função centralizada `ptst.RegistraModuloImpl()`.
- **Desacoplamento Máximo**: Adicionar ou remover um módulo nativo na distribuição oficial do Harpia não exige alterações estruturais na VM ou no parser. Basta criar a pasta e declarar seu blank import em `stdlib.go`.

---

## 🧩 Módulo Central: `embutidos`

O subpacote `embutidos` é a base da linguagem. Diferente dos outros pacotes, seus símbolos e métodos **não necessitam de importação**. Eles são injetados diretamente na tabela global de símbolos e ficam imediatamente disponíveis em qualquer arquivo de código.

### Principais Recursos Embutidos:

- **`escreva(args...)`**: Imprime representações textuais na saída padrão do terminal.
- **`leia(mensagem?)`**: Pausa a execução para coletar uma entrada textual digitada pelo usuário no terminal.
- **`tamanho(objeto)`**: Retorna a contagem de elementos de uma lista, caracteres de uma string ou chaves de um dicionário.
- **`int(objeto)`**: Tenta fazer o casting (conversão) forçado do objeto para o tipo inteiro de 64 bits.
- **`texto(objeto)`**: Converte e devolve a representação de string de qualquer tipo primitivo da linguagem.

---

## 📐 Módulo: `matematica`

O módulo `matematica` disponibiliza constantes físicas e rotinas de aritmética avançada para complementar os operadores básicos de soma, subtração, multiplicação e divisão da linguagem.

- **Importação obrigatória**: `importar matematica`

### Constantes Disponíveis:

- **`matematica.PI`**: Constante matemática aproximada $\approx 3.141592653589793$.
- **`matematica.E`**: Constante de Euler e base dos logaritmos naturais $\approx 2.718281828459045$.

### Métodos e Funções:

| Função | Assinatura | Descrição Técnica |
| :--- | :--- | :--- |
| `absoluto` | `absoluto(numero)` | Retorna o valor absoluto (magnitude real sem sinal) de um número real (`math.Abs`). |
| `piso` | `piso(decimal)` | Arredonda o número de ponto flutuante para baixo, retornando o menor número inteiro mais próximo (`math.Floor`). |
| `teto` | `teto(decimal)` | Arredonda o número para cima, retornando o menor inteiro maior ou igual (`math.Ceil`). |
| `potencia` | `potencia(base, expoente)` | Calcula a potenciação correspondente a base elevada ao expoente informado (`math.Pow`). |
| `raiz` | `raiz(radicando, indice?)` | Retorna a raiz do radicando pelo índice. Se o índice for omitido, calcula a raiz quadrada por padrão. Implementado sob conversão de expoente fracionário ($radicando^{1.0/indice}$). |

---

## 💻 Módulo: `sistema`

O módulo `sistema` expõe informações sobre a infraestrutura e o ambiente de software em execução.

- **Importação obrigatória**: `importar sistema`

### Constantes Disponíveis:

- **`sistema.NOME`**: Retorna o identificador textual do sistema operacional hospedeiro (ex: `"darwin"` para macOS, `"linux"` para sistemas baseados em Linux, `"windows"` para Windows).
- **`sistema.ARQUITETURA`**: Retorna o tipo de processador/arquitetura onde o interpretador está rodando (ex: `"amd64"`, `"arm64"`, `"386"`).

---

## 🎨 Módulo: `colorize`

O módulo `colorize` (colorização de console) adiciona capacidades de estilização gráfica para que os desenvolvedores criem saídas ricas em cores no terminal usando sequências de escape ANSI.

- **Importação obrigatória**: `importar colorize`

---

## 🔌 Módulo: `soquete`

O módulo `soquete` fornece uma interface direta de controle de sockets de rede de baixo nível, permitindo criar scripts que conversam com servidores externos ou gerenciam conexões soquete ativas.

- **Importação obrigatória**: `importar soquete`

---

## 📦 Novos Módulos de Backend (v1.x)

Para complementar o ecossistema profissional de backend (Fase 3), a biblioteca padrão do Harpia incorpora módulos nativos adicionais extremamente poderosos e integrados:

### 1. Módulo: `arquivos`
Permite a manipulação direta e segura do sistema de arquivos físico (I/O). Conta com proteção nativa contra acessos não autorizados por meio do Sandbox de Segurança (`BloquearArquivos`).
* **Funções**: `ler()`, `escrever()`, `acrescentar()`, `remover()`, `renomear()`, `caminhar()`, `resolver()`.

### 2. Módulo: `http`
Protocolo completo HTTP para cliente e servidor de alto desempenho. Conta com proteção nativa contra pânicos de execução (Recovery), timeouts slowloris e Sandbox de Rede (`BloquearRede`). Suporta assinaturas HMAC SHA-256 e geração de especificações OpenAPI 3.0 para o servidor.
* **Classes**: `Servidor` (suporta `obter()`, `postar()`, `deletar()`, `usar()`, `escutar()`, `fechar()`), `Requisicao`, `Resposta`.
* **Funções**: `requisitar(metodo, url, ...)`, `assinar_hmac(chave, mensagem)`, `verificar_hmac(chave, mensagem, assinatura)`, `gerar_openapi(servidor)`.

### 3. Módulo: `bd`
Interface unificada e query builder para bancos de dados relacionais, não-relacionais (NoSQL) e vetoriais de alto rendimento.
* **SQL & ORM**: Drivers integrados para `SQLite`, `PostgreSQL` e `MySQL`. Suporta pool de conexões, Query Builder fluído `bd.tabela("usuarios").onde(...).obterMuitos()`, e ORM Tipado opcional passando esquema na tabela: `bd.tabela("usuarios", {"nome": "texto", "idade": "inteiro"})`.
* **NoSQL**: Conectores e mapeadores para coleções de documentos no `MongoDB` e chaves-valores/cache rápido no `Redis`.
* **Vetorial**: Conector `conectarQdrant(url, colecao)` de alto rendimento para bancos vetoriais, suportando operações de `inserir`, `buscar` (por cosseno/L2) e `deletar` pontos.

### 4. Módulo: `json`, `yaml` e `xml`
Módulos de serialização e desserialização ultra-velozes para tráfego e formatação estruturada de dados.
* **`json`**: `analisar(string)` (parse de JSON para objetos e dicionários nativos) e `serializar(objeto)` (gera string JSON compacta).
* **`yaml`**: Manipulação de manifestos e configurações estruturadas.
* **`xml`**: Conversão bidirecional entre textos XML e dicionários nativos.

### 5. Módulo: `cripto`
Suporte nativo a assinaturas criptográficas, hashes de integridade e geração de identificadores únicos seguros.
* **Funções**: `sha256()`, `codificarBase64()`, `decodificarBase64()`, `uuid()`.

### 6. Módulo: `logs`
Logging estruturado com suporte a níveis (`info`, `alerta`, `erro`, `depurar`), metadados e formatação de texto colorido ou JSON industrial.

### 7. Módulo: `metricas`
Criação dinâmica de contadores e medidores (Gauges) no padrão Prometheus para observabilidade.

### 8. Módulo: `esquema`
Mecanismo de validação declarativa de dados baseado em esquemas e tipagens (estilo Zod).

### 9. Módulo: `tarefas`
Agendador de tarefas periódicas via Cron e filas assíncronas concorrentes em memória.

### 10. Módulo: `ffi`
FFI portátil para carregar dinamicamente bibliotecas C (.so, .dll, .dylib) e chamar assinaturas de forma síncrona.

### 11. Módulo: `ia`
Integração nativa com inteligência artificial para criação de Agentes autônomos com memória, conectores Ollama/nuvem, orquestração de diálogos e contratos semânticos de validação de esquemas de resposta.
* **Funções**: `validar_resposta(esquema, resposta_json)` para certificar o formato da IA em runtime.

### 12. Módulo: `resiliencia`
Padrões nativos de resiliência e estabilidade para microsserviços.
* **Funções**: `novo_disjuntor(limite_falhas, timeout_segundos)`, `novo_limite_de_taxa(max_tokens, tokens_seg)`, `nova_retentativa(tentativas, base_ms, fator)`.

### 13. Módulo: `telemetria`
Observabilidade nativa e leve compatível com as especificações do OpenTelemetry para exportação de dados.
* **Funções**: `novo_tracer(servico)` (retorna Tracer com `iniciar_span`), `nova_metrica(nome, tipo)` (retorna Metrica com `registrar`).

---

## 📝 Exemplo Completo de Uso de Módulos

Abaixo está um exemplo completo de um arquivo em Harpia (`.hrp`) ilustrando a sintaxe de carregamento e o uso conjunto de múltiplos recursos da `stdlib`:

```harpia
# Exemplo de uso conjunto da biblioteca padrão

importar matematica
importar sistema

# 1. Usando recursos de sistema para diagnóstico
escreva("Rodando em: " + sistema.NOME + " (" + sistema.ARQUITETURA + ")")

# 2. Operações matemáticas de geometria básica
raio = 5.0
area = matematica.PI * matematica.potencia(raio, 2)
escreva("Área do círculo de raio 5: " + texto(area))

# 3. Arredondamento de valores decimais
escreva("Valor teto de area: " + texto(matematica.teto(area)))
escreva("Valor piso de area: " + texto(matematica.piso(area)))

# 4. Cálculo de raízes
escreva("Raiz quadrada de 144: " + texto(matematica.raiz(144)))
escreva("Raiz cúbica de 27: " + texto(matematica.raiz(27, 3)))
```
