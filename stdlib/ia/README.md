# 🤖 Módulo `ia` (Biblioteca Padrão)

O módulo `ia` fornece conectores nativos e primitivas para integração de Inteligência Artificial e construção de **Agentes Autônomos** diretamente na linguagem Harpia.

---

## 🚀 Como Importar

```harpia
de "ia" importe Agente
```

---

## 🧩 Primitiva `Agente`

A classe `Agente` representa um agente inteligente autônomo. Cada agente possui seu próprio conjunto de instruções (diretrizes do sistema/comportamento), um provedor e modelo de linguagem configurados, e uma memória persistente integrada (histórico de conversas).

### Construtor

```harpia
var meuAgente = Agente(nome, instrucoes, provedor?, modelo?)
```

- **`nome`**: (Texto) Identificador do agente.
- **`instrucoes`**: (Texto) System prompt contendo o comportamento do agente.
- **`provedor`**: (Texto, Opcional) O provedor de IA a ser utilizado. Opções: `"ollama"`, `"gemini"`, `"openai"`. Padrão: `"ollama"`.
- **`modelo`**: (Texto, Opcional) O modelo de linguagem a ser utilizado (ex: `"llama3"`, `"gemini-1.5-flash"`, `"gpt-4o-mini"`). Padrão: `"llama3"`.

### Atributos

- **`nome`**: Retorna o nome do agente.
- **`instrucoes`**: Retorna o prompt de sistema do agente.
- **`provedor`**: Retorna o provedor configurado.
- **`modelo`**: Retorna o modelo configurado.
- **`historico`**: Retorna uma `Lista` contendo o histórico estruturado de mensagens (`[{"role": "user", "content": "..."}]`).

### Métodos

- **`perguntar(mensagem)`**: Envia uma pergunta ao agente, anexando todo o histórico de conversas anterior, e retorna a resposta em formato textual.
- **`limpar_memoria()`**: Zera completamente o histórico de mensagens salvas.
- **`comunicar(outroAgente, mensagem)`**: Envia uma instrução ou pergunta para outro agente, coleta a resposta dele, e a registra na própria memória do agente chamador de forma orquestrada (suporte nativo multi-agente).

---

## 🔌 Provedores Suportados e Variáveis de Ambiente

O runtime do Harpia gerencia de forma transparente a comunicação e a segurança ativa no tráfego das requisições de IA:

### 1. Ollama (Provedor Local / Padrão)
A linguagem prioriza a execução local para garantir a soberania e a privacidade dos dados.
- **Configuração**: Conecta-se à API local do Ollama no endereço definido na variável de ambiente `OLLAMA_HOST` (padrão: `http://localhost:11434`).
- **Fallback Transparente**: Se o Ollama local não estiver rodando no host e existirem chaves de APIs de nuvem configuradas (`GEMINI_API_KEY` ou `OPENAI_API_KEY`), o compilador redireciona a chamada para o modelo em nuvem correspondente automaticamente.

### 2. Google Gemini
- **Variável Necessária**: `GEMINI_API_KEY`
- **Modelo Padrão**: `gemini-1.5-flash`

### 3. OpenAI
- **Variável Necessária**: `OPENAI_API_KEY`
- **Modelo Padrão**: `gpt-4o-mini`

---

## 📝 Exemplos de Uso

### Exemplo 1: Conversa Simples com Memória

```harpia
de "ia" importe Agente

# Criamos um agente focado em desenvolvimento
var desenvolvedor = Agente("DevHelper", "Você é um programador sênior em Harpia", "ollama", "llama3")

var resposta1 = desenvolvedor.perguntar("como declarar uma lista no Harpia?")
imprimir(resposta1)

# O histórico mantém o contexto das mensagens seguintes
var resposta2 = desenvolvedor.perguntar("me dê um exemplo prático desse tipo de dado?")
imprimir(resposta2)
```

### Exemplo 2: Comunicação Multi-Agente

```harpia
de "ia" importe Agente

var escritor = Agente("Redator", "Você escreve poemas curtos", "ollama", "llama3")
var revisor = Agente("Editor", "Você analisa poemas e sugere melhorias formais", "ollama", "llama3")

# O redator gera e envia para o editor revisar
var poemaOriginal = "O código roda na floresta escura, a Harpia voa na altura..."
var revisao = escritor.comunicar(revisor, "Analise este poema: " + poemaOriginal)

imprimir("Revisão do Editor:")
imprimir(revisao)
```
