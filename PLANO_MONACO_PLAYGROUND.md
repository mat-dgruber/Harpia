# Plano de Implementação — Monaco Editor no Playground do Harpia

## Visão Geral

**Objetivo:** Substituir o `<textarea>` cru do playground web por Monaco Editor com syntax highlighting (Monarch), hover com docs, e two-way binding reativo — resolvendo os 3 sintomas reportados: identificadores sem cor, `estilo`/chaves sem cor, e hover sem tooltip.

**Status atual:** Editor é uma `<textarea>` sem nenhum tokenização. Não há syntax highlighting, não há hover, não há autocomplete.

---

## Arquitetura Atual vs Proposta

### Atual

```
interface.hrp → textarea puro (sem highlight)
cmd/playground.go → htmlInterfacePlayground (servi estático)
                   → /api/executar (única rota API)
```

### Proposta

```
cmd/playground.go → htmlInterfacePlayground (com Monaco loader CDN + CSS)
                   → /api/executar (execução existente)
                   → /api/editor-config (keywords + default code + docs)
                   → /api/docs?palavra=X (hover docs)
                   → /editor-monaco.js (JS do Monaco: linguagem, tema, montagem, binding)

interface.hrp → <div id="editor-mount"> (container para Monaco)
              → remove <textarea> + bindings de código
```

---

## Task 1: Trocar textarea por Monaco Editor no playground

### O que fazer

- No `htmlInterfacePlayground` (cmd/playground.go:257), adicionar:
  - `<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/monaco-editor/0.45.0/min/vs/editor/editor.main.css">` no `<head>`
  - `<script src="https://cdnjs.cloudflare.com/ajax/libs/monaco-editor/0.45.0/min/vs/loader.js"></script>` no `<body>`
  - `<script src="/editor-monaco.js"></script>` antes do fechamento do `<body>`
  - Remover o estilo `.editor-textarea` que não será mais usado
- Criar a rota `/editor-monaco.js` em `iniciarServidorPlayground()` que devolve JS estático inline

### Por quê

Monaco Editor é o editor que VS Code usa. Ele suporta Monarch (gramática declarativa), HoverProvider (tooltips), e bind reativo via eventos. CDN elimina a necessidade de bundler.

---

## Task 2: Mapear tokens Portuscript para Monarch Language

### O que fazer

Em `editor-monaco.js` (JS literal servido pelo servidor), registrar uma linguagem Monarch com:

```
palavrasChave: [
  se, senao, enquanto, para, retorne, pare, continue,
  de, importe, Verdadeiro, Falso, Nulo,
  var, const, constante, func, funcao,
  ou, e, nao,
  nova, classe, estende, self, estatico,
  assegura, testar,
  tente, capture, finalmente,
  exportar, em, assincrono, aguarde, estilo
]

operadores: [=, +, -, *, **, /, //, %, <, <=, ==, !=, >, >=, +=, -=, *=, /=, //=, |, ^, &, ~, <<, >>, ., |>]

delimitadores: [(, ), {, }, [, ], ;, ,, :]

comentario: # ate fim de linha

comentarioHTML: <!-- ... -->
```

Regras Monarch:

- Comentários `#` → token `comment` (cinza)
- Comentários `<!-- ... -->` → token `comment` (cinza)
- Strings `"..."` e `'...'` → token `string` (verde)
- Números (int e decimal) → token `number` (laranja)
- Palavras-chave → token `keyword` (roxo)
- Operadores → token `operator` (branco/ciano)
- Delimitadores `{`, `}` → token `delimiter.bracket` (amarelo suave)
- Delimitadores `(`, `)`, `[`, `]` → token `delimiter` (amarelo suave)
- Identificadores (qualquer outra coisa) → token `identifier` (azul claro)

### Por quê

Mapa direto dos `tokensIdentificadores` + `tokensSimples` do lexer Go. Mantém o syntax highlighting consistente com a linguagem real. A keyword `estilo` e as chaves `{ }` passam a ter cor.

---

## Task 3: Registrar HoverProvider com docs do Harpia

### O que fazer

- Criar endpoint `GET /api/docs?palavra=X` em `cmd/playground.go`
- Recebe uma palavra, retorna JSON `{ "doc": "markdown string" }`
- Doc map hardcoded no servidor (inline): `funcao` → assinatura + exemplo, `var` → explicação, `sinal` → "Cria um sinal reativo...", etc.
- Em `editor-monaco.js`, registrar `monaco.languages.registerHoverProvider('portuscript', ...)` que:
  - Obtém a palavra sob o cursor (`wordAtPosition`)
  - Faz `fetch('/api/docs?palavra=' + word)`
  - Retorna `monaco.languages.Hover` com markdown formatado

### Por quê

Hover é nativo do Monaco e dá DX profissional. Mostra assinatura de funções e built-ins quando o mouse paira sobre `funcao`, `sinal`, `retorne`, etc. Sem isso, o playground é só um textarea colorido.

---

## Task 4: Escolher e aplicar tema Monaco dark

### O que fazer

Em `editor-monaco.js`, antes de criar o editor:

```js
monaco.editor.defineTheme("harpia-dark", {
  base: "vs-dark",
  inherit: true,
  rules: [
    { token: "keyword", foreground: "c678dd" }, // roxo
    { token: "string", foreground: "98c379" }, // verde esmeralda
    { token: "number", foreground: "d19a66" }, // laranja
    { token: "comment", foreground: "7f848e", fontStyle: "italic" }, // cinza
    { token: "operator", foreground: "56b6c2" }, // ciano
    { token: "identifier", foreground: "61afef" }, // azul claro
    { token: "delimiter.bracket", foreground: "e5c07b" }, // amarelo (chaves)
    { token: "delimiter", foreground: "e5c07b" }, // amarelo (parênteses/colchetes)
  ],
  colors: {
    "editor.background": "#0f172a", // slate-950
    "editor.foreground": "#e2e8f0", // slate-200
    "editor.lineHighlightBackground": "#1e293b", // slate-800
    "editor.selectionBackground": "#334155", // slate-700
    "editorCursor.foreground": "#38bdf8", // sky-400
    "editorLineNumber.foreground": "#475569", // slate-600
    "editorLineNumber.activeForeground": "#94a3b8", // slate-400
  },
});
monaco.editor.setTheme("harpia-dark");
```

### Por quê

Tema dark consistente com o `bg-slate-950` do Tailwind já usado no playground. Cores do One Dark Pro (popular) adaptadas para portasuglês legível.

---

## Task 5: Wirear two-way binding Monaco ↔ sinal codigo

### O que fazer

Em `editor-monaco.js`:

1. **Init**: buscar `/api/editor-config` → `{ defaultCode: "..." }`. Criar editor com `value: defaultCode`.

2. **Monaco → Sinal**: registrar `editor.onDidChangeModelContent(() => { window.__psiSetCodigo(editor.getValue()); })`. Flag `__psiLock` evita loop.

3. **Sinal → Monaco**: expor `window.__psiSetCodigo` e `window.__psiEditor` para que `interface.hrp` possa chamar. Se `codigo()` muda externamente, `editor.setValue(newVal)` com flag de lock.

4. **Ctrl+Enter**: `editor.addCommand(monaco.KeyMod.CtrlCmd | monaco.KeyCode.Enter, () => window.__psiExecutar())`. `window.__psiExecutar` será ligado pelo `interface.hrp`.

### Por quê

Hoje `interface.hrp` usa `ligar={codigo}` no textarea para two-way binding. Com Monaco, o binding é manual via eventos. O `window.__psi*` bridge conecta o SPA Portuscript ao editor vanilla JS sem modificar o runtime.

---

## Task 6: Manter app_teste.hrp como conteúdo padrão do editor

### O que fazer

Em `cmd/playground.go`, handler `/api/editor-config`:

```go
func apiEditorConfig(w http.ResponseWriter, r *http.Request) {
    // Ler app_teste.hrp do diretório de trabalho
    defaultCode, err := os.ReadFile("app_teste.hrp")
    if err != nil {
        defaultCode = []byte("# Escreva seu código aqui...\n")
    }

    config := map[string]string{
        "defaultCode": string(defaultCode),
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(config)
}
```

### Por quê

O usuário pediu explicitamente manter o `app_teste.hrp` como default. Hoje o snippet default é inline no `interface.hrp:13`. O endpoint serve o conteúdo real do arquivo, que pode ser atualizado sem recompilar.

---

## Ordem de Execução

| Passo | Task                    | Dependência | Arquivos                                               |
| ----- | ----------------------- | ----------- | ------------------------------------------------------ |
| 1     | Task 6                  | Nenhuma     | `cmd/playground.go`                                    |
| 2     | Task 2 + 3 + 4 (juntas) | Passo 1     | `cmd/playground.go` (handlers), novo arquivo JS inline |
| 3     | Task 1                  | Passos 1+2  | `cmd/playground.go` (HTML template)                    |
| 4     | Task 5                  | Passos 1+3  | `cmd/playground.go` (handlers), `interface.hrp`        |

### Arquivos modificados

- `Harpia/cmd/playground.go` — HTML template + 3 novos handlers + JS inline
- `Harpia/playground/interface.hrp` — trocar textarea por div#editor-mount

### Arquivos NÃO modificados

- `Harpia/lexer/*` — sem alteração (tokens já mapeados estáticos no JS)
- `Harpia/playground/playground.go` — sem alteração
- `Harpia/playground/executor.go` — sem alteração

---

## Riscos e Mitigações

| Risco                                                | Mitigação                                                                         |
| ---------------------------------------------------- | --------------------------------------------------------------------------------- |
| Monaco CDN lento / offline                           | Usar versão fixa (0.45.0), fallback para textarea se loader falhar                |
| Loop infinito Monaco ↔ sinal                         | Flag `__psiLock` em cada direção de binding                                       |
| Hover API lento                                      | Cache de docs no cliente (Map em JS), debounce de 300ms                           |
| `interface.hrp` não compila (extensão .hrp vs .ptst) | Verificar se `servePlaygroundJS` lê `.ptst` — se `.hrp` renomear, ajustar handler |

---

## Checklist de Aceitação

- [ ] Variáveis como `contadorSinal`, `definirContador` aparecem em azul claro
- [ ] Keywords `funcao`, `var`, `se`, `retorne`, `estilo` aparecem em roxo
- [ ] Chaves `{ }` aparecem em amarelo suave
- [ ] Strings `"#f4f4f9"` aparecem em verde
- [ ] Números `10`, `1`, `15` aparecem em laranja
- [ ] Comentários `# ...` aparecem em cinza itálico
- [ ] Hover sobre `funcao` mostra tooltip com assinatura
- [ ] Hover sobre `sinal` mostra doc do built-in
- [ ] Ctrl+Enter executa o código
- [ ] Conteúdo padrão do editor é `app_teste.hrp`
- [ ] Two-way binding funciona (sinal `codigo` ↔ Monaco)
- [ ] Tema dark consistente com Tailwind slate-950
