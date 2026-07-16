# Exemplos Práticos de Aplicação — Guia de Marca Harpia

Documento complementar com casos de uso reais e modelos para implementar a identidade visual, tom de voz e diretrizes em diferentes contextos.

---

## 1. Exemplos de Código e Tema de Sintaxe

### 1.1 Código Harpia com Tema "Harpia Dark"

```harpia
// usuario.hrp — Módulo de domínio (Domain Layer)
// Exemplo de código bem estruturado em Harpia

usar raiz::resultado
usar tronco::conexao

// ============= Tipos de Domínio =============

estrutura Usuario {
  id: Texto,
  nome: Texto,
  email: Texto,
  ativo: Booleano,
}

contrato RepositorioUsuario {
  obter(id: Texto) -> Resultado<Usuario>
  salvar(usuario: Usuario) -> Resultado<Usuario>
  listar() -> Resultado<Lista<Usuario>>
}

// ============= Casos de Uso =============

funcao criar_usuario(nome: Texto, email: Texto) -> Resultado<Usuario> {
  // Validar entrada
  se nome::vazio?() {
    retornar Resultado::erro("Nome não pode ser vazio")
  }
  
  // Criar entidade de domínio
  usuario = Usuario {
    id: gerar_id_unico(),
    nome: nome,
    email: email,
    ativo: verdadeiro,
  }
  
  // Persistir (retorna Resultado)
  retornar repositorio::salvar(usuario)
}

// Exemplo de teste unitário
teste "deve validar email duplicado" {
  usuario1 = Usuario { email: "joao@harpia.dev", ... }
  resultado = criar_usuario("João", "joao@harpia.dev")
  
  afirmar resultado::erro? == verdadeiro
  afirmar resultado::mensagem contém "duplicado"
}
```

**Destaques de Sintaxe (Cores):**

- **Keywords** (`se`, `retornar`, `funcao`): Terracota (#D45D34)
- **Types** (`Usuario`, `Resultado`, `Texto`): Azul Celeste (#38B6FF)
- **Strings** (`"Nome não pode ser vazio"`): Broto de Ipê (#00A86B)
- **Comments**: Névoa da Manhã (#8C9B9E)
- **Function calls** (`criar_usuario`, `salvar`): Penagem Branca (#F3F6F4)

---

## 2. Exemplos de Mensagens CLI

### 2.1 Mensagem de Boas-vindas (Comando `harpia`)

**Versão Final (Padrão Unificado - Premium):**

```text
 🦅  HARPIA v1.2.0  •  Linguagem de Programação 100% Brasileira
 ──────────────────────────────────────────────────────────────
 "O código bonito é como uma boa bossa nova: simples e fluido."
 ──────────────────────────────────────────────────────────────
 Sistema: macOS (arm64)  |  Projeto ativo: meu-app (v1.0.0)
 
 › harpia ajuda   — Lista de todos os comandos do CLI
 › harpia novo    — Cria uma nova estrutura Clean/DDD (.hrp)
 › harpia ia      — Inicia o suporte interativo via IA local
 
 [F1] Ajuda  │  [F2] Executar  │  [F3] Mentor IA  │  [ESC]/[ctrl+d] Sair
 ──────────────────────────────────────────────────────────────
 » 
```

**Cores ANSI Recomendadas para o Terminal:**
*   **Título (`🦅 HARPIA`):** Olho da Harpia em Negrito (`\033[1;33m` ou `#F2A900`)
*   **Slogan/Ditado:** Névoa da Manhã em Itálico (`\033[3;90m` ou `#8C9B9E`)
*   **Comandos (`› harpia ...`):** Terracota (`\033[38;2;212;93;52m` ou `#D45D34`)
*   **Barra de Atalhos (`[F1] ...`):** Fundo Rio Profundo e texto em Penagem Branca (`\033[48;2;23;30;38m\033[38;2;243;246;244m`)


**Cores Utilizadas:**

- Verde da bandeira (■): Broto de Ipê (`#00A86B`)
- Losango (◆): Olho da Harpia (`#F2A900`)
- Emoji 🦅: Olho da Harpia (`#F2A900`)
- Texto: Penagem Branca (`#F3F6F4`)

### 2.2 Sucesso de Compilação

```
[✓] Harpia 1.2.0 — Compilação iniciada
 ─ Verificando dependências... OK (3 módulos)
 ─ Verificando tipos... OK (42 funções, 8 estruturas)
 ─ Gerando código... OK (512 KB)
 ─ Otimizando... OK (+15% performance)

✨ Compilado com sucesso em 2.34s
   Saída: ./dist/app.hrpx (8.2 MB)
   Próximas etapas: harpia executar ./dist/app.hrpx
```

**Cores:**

- Checkmark: Broto de Ipê (#00A86B)
- Título: Olho da Harpia (#F2A900)
- Texto geral: Penagem Branca (#F3F6F4)
- Itens: Névoa da Manhã (#8C9B9E)

### 2.3 Erro de Compilação (Formatado)

```
[HRP-3001] Erro de Tipo: Incompatibilidade de argumentos
 ──> src/modelos/pedido.hrp:45:12
  │
44│  resultado = usuario::obter(id, nome)
45│                            ^^^^^^^^
  │                            esperado 1 argumento, encontrado 2
  │
 Dica: A função usuario::obter() aceita apenas um argumento (id).
       Se deseja buscar por nome também, use usuario::buscar_por_nome(nome).

 Saiba mais: harpia doc HRP-3001
            harpia ia "como buscar usuário por nome"
```

**Componentes coloridos:**

- `[HRP-3001]`: Olho da Harpia (#F2A900)
- Localização (`src/...`): Névoa da Manhã (#8C9B9E)
- Indicador de erro (`^^^^^^^^`): Terracota (#D45D34)
- Linha de contexto (`│`): Rio Profundo (#171E26)
- "Dica:": Broto de Ipê (#00A86B)

### 2.4 Aviso de Dependência Desatualizada

```
⚠️  Aviso: 2 dependências desatualizadas encontradas

  • copa (v1.0.2 → v1.1.0)
    └ Mudanças: suporte a WebSocket, 15% mais rápido
  
  • correnteza (v2.1.0 → v2.2.0)
    └ Mudanças: 3 bugs corrigidos, 1 breaking change
  
Execute 'harpia atualizar' para revisar mudanças antes de instalar.
```

---

## 3. Exemplos de Documentação

### 3.1 Página de Tutorial (Markdown)

```markdown
# Tutorial: Seu Primeiro Programa em Harpia

Bem-vindo! Neste guia, você criará um programa simples que
valida emails e exibe mensagens personalizadas.

## O que você vai aprender

- Estruturas básicas e tipos
- Padrão de matching (`caso`)
- Como trabalhar com `Resultado<T>` para tratamento de erros

## Pré-requisitos

- Harpia v1.2+ instalado (`harpia versao`)
- Editor de texto ou VS Code com extensão Harpia

## Passo 1: Criar um novo projeto

```bash
$ harpia novo meu_email_validator
$ cd meu_email_validator
```

Você verá uma estrutura como:

```
meu_email_validator/
├── harpia.toml
├── src/
│   └── main.hrp
└── testes/
```

## Passo 2: Escrever o código

Abra `src/main.hrp` e copie:

```harpia
usar raiz::imprimir

funcao validar_email(email: Texto) -> Booleano {
  retornar email contém "@" e email contém "."
}

funcao main() {
  email = "usuario@harpia.dev"
  
  se validar_email(email) {
    imprimir("✓ Email válido: " + email)
  } senao {
    imprimir("✗ Email inválido")
  }
}
```

## Passo 3: Executar

```bash
$ harpia executar
```

Você deve ver:

```
✓ Email válido: usuario@harpia.dev
```

## Próximos passos

- Aprenda sobre [Resultado<T></t> e tratamento de erros](./resultado.md)
- Explore [tipos personalizados](./tipos.md)
- Veja [exemplo de API HTTP](./api-http.md)

---

**Precisa de ajuda?** Digite `harpia ia` ou acesse nosso fórum.

```

**Estilo de escrita:**
- Ton caloroso mas preciso
- Títulos em markdown (`#`)
- Blocos de código com syntax highlight
- Progressão clara (passo-a-passo)
- Links internos contextualizados

### 3.2 Página de Referência de API

```markdown
## Módulo: `raiz` — Núcleo do Sistema

O módulo `raiz` fornece funções e tipos fundamentais para
qualquer programa Harpia.

### Funções

#### `imprimir(mensagem: Texto) -> Nada`

Exibe uma mensagem no console padrão.

**Exemplo:**
```harpia
imprimir("Olá, Harpia!")
// Saída: Olá, Harpia!
```

**Notas:**

- Adiciona quebra de linha automaticamente
- Suporta interpolação: `imprimir("Número: {42}")`
- Use `imprimir_erro()` para stderr

---

#### `debug(valor: T) -> T`

Exibe o valor para debugging e o retorna.

**Exemplo:**

```harpia
resultado = lista |> filtro(...) |> debug() |> mapa(...)
// Exibe: [1, 2, 3]
```

**Útil para:** Inspecionar fluxos de dados em pipes

---

### Tipos

#### `Resultado<T, E>`

Tipo que representa sucesso ou erro.

**Variantes:**

- `Resultado::ok(valor: T)` — sucesso
- `Resultado::erro(erro: E)` — falha

**Exemplo:**

```harpia
funcao dividir(a: Numero, b: Numero) -> Resultado<Numero> {
  se b == 0 {
    retornar Resultado::erro("Divisão por zero")
  }
  retornar Resultado::ok(a / b)
}
```

---

**Veja também:** [Tratamento de erros](./erros.md) | [Matching](./match.md)

```

**Estilo:**
- Hierarquia clara (headings aninhados)
- Exemplos de código comentados
- Links de "Veja também"
- Notas e avisos destacados

---

## 4. Exemplos de Comunicação Social/Comunidade

### 4.1 Post de Lançamento (Twitter/Bluesky)

```

🦅 Harpia v1.2.0 foi ao ar!

✨ O que tem de novo:
  • WebSocket nativo (copa)
  • Compilação 40% mais rápida
  • 25 bugs corrigidos
  • Docs em português 100% completas

Baixe agora: https://harpia.dev/download
Changelog: https://harpia.dev/v1.2.0

Muito obrigado à comunidade brasileira! 🇧🇷💚

#Harpia #LanguageDev #OpenSource

```

**Tom:**
- Entusiasmado mas profissional
- Emojis do tema (🦅 para Harpia)
- Links diretos
- Agradecimento à comunidade

### 4.2 Anúncio de Comunidade (Discord/Slack)

```

🦅 Bem-vindo à Comunidade Harpia!

Que bom ter você aqui! Somos um grupo de desenvolvedoras e
desenvolvedores apaixonados por código limpo e arquitetura
elegante — e que acreditam em tecnologia brasileira.

📌 Canais principais:
  • #geral — conversas sobre Harpia
  • #ajuda — perguntas (não há perguntas bobas!)
  • #showoff — compartilhe seus projetos
  • #contribuindo — guia para contribuir ao projeto

✨ Primeira vez? Leia nosso Código de Conduta: link

Dúvida técnica? Use 'harpia ia <sua pergunta></sua>' no CLI.

Vamos codar juntos! 🚀

```

**Tom:**
- Acolhedor e inclusivo
- Informativo mas conciso
- Referência a recursos disponíveis
- Convite à participação

### 4.3 Newsletter (Harpia Weekly)

```

═══════════════════════════════════════════════════════════
  HARPIA WEEKLY #47 — Semana de 15 de julho de 2026
═══════════════════════════════════════════════════════════

Olá, comunidade Harpia! 👋

Semana intensa no repo. Vamos aos destaques:

📰 NOTÍCIAS

[1] Harpia v1.2 em Release Candidate
    A compilação agora é 40% mais rápida graças a
    compilação incremental. Teste: harpia novo teste-v12

[2] João Silva fez PR monumental (DDD native patterns)
    680 linhas de código, 12 testes verdes.
    Parabéns, João! 🎉

🔧 MERGES NOTÁVEIS

  • Corrigido parsing de strings multilinhas (#456)
  • Adicionado suporte a destructuring em case (#421)
  • Melhorado erro HRP-2001 com dica automática (#398)

💡 TIP DA SEMANA

Você sabia? Pode usar pipes em testes também:

    teste "filtro deve remover nulos" {
      resultado = [1, nulo, 3] |> filtro(nao_nulo?)
      afirmar resultado == [1, 3]
    }

📚 RECURSO EM DESTAQUE

Novo tutorial: "Construindo uma API REST com Harpia"
Link: https://docs.harpia.dev/api-rest

👥 NÚMEROS DESTA SEMANA

  ⭐ 250 novos stars no GitHub
  🔀 18 PRs merged
  🐛 5 bugs corrigidos
  📝 7 docs adicionadas

🎤 VOCÊ TEM VOZ

Tem uma ideia legal para Harpia? Abra uma discussion em:
https://github.com/harpialang/harpia/discussions

Até semana que vem! 🦅

═══════════════════════════════════════════════════════════
  Curado com ❤️ pela Comunidade Harpia
═══════════════════════════════════════════════════════════

```

**Estrutura:**
- Header visual (emojis Harpia)
- Seções claras (Notícias, Merges, Tips, Números)
- Tom casual mas informativo
- Call-to-action no final

---

## 5. Exemplos de Assets Gráficos (Descrição)

### 5.1 Ícone de Erro (HRP-XXXX)

```

Descrição: Ícone mostra um "!" em círculo
  Cores: Terracota (#D45D34) em fundo Sombra da Floresta (#162913)
  Estilo: Geométrico, sem preenchimento (stroke 2px)
  Uso: Mensagens de erro, placeholders de aviso
  Tamanho: 16x16px (mínimo), 32x32px (padrão), 64x64px (grande)

```

### 5.2 Ícone de Sucesso

```

Descrição: Checkmark estilizado (simplificado)
  Cores: Broto de Ipê (#00A86B) em fundo transparente
  Estilo: Stroke 2.5px, suavizado
  Uso: Confirmações, validações bem-sucedidas
  Animação: Opcional - stroke animation 0.6s ease-in-out

```

### 5.3 Gradiente de Fundo (Landing Page)

```

Descrição: Gradiente diagonal sutil da Floresta ao Rio Profundo
  Ponto inicial: Sombra da Floresta (#162913) — canto superior esquerdo
  Ponto final: Rio Profundo (#171E26) — canto inferior direito
  Ângulo: 135 graus
  Transição: Linear, suavidade natural
  Sobreposição: Padrão geométrico em 10% opacity
  Uso: Backgrounds de página, cards, seções

```

---

## 6. Checklist de Revisão (QA/Brand)

Ao revisar qualquer material Harpia (código, docs, social, etc.):

### Cores ✓
- [ ] Apenas cores da paleta oficial usadas
- [ ] Contraste suficiente (4.5:1 para texto pequeno)
- [ ] Dark Mode é padrão (Light Mode opcional)

### Tipografia ✓
- [ ] Títulos em Outfit ou Inter (não serif)
- [ ] Código em JetBrains Mono ou Fira Code
- [ ] Escala tipográfica respeitada
- [ ] Sem mais de 2-3 tamanhos diferentes

### Tom e Voz ✓
- [ ] Autoridade amigável (não muito formal, não casual demais)
- [ ] Jargão técnico explicado ou em português
- [ ] Mensagens construtivas (não apenas problemas)
- [ ] Inclusivo; não pressupõe conhecimento avançado

### Logo e Símbolos ✓
- [ ] Logo é versão oficial (sem distorção)
- [ ] Monocromático para CLI/terminal
- [ ] Colorido com olho destacado em web
- [ ] Espaço livre respeitado

### Conteúdo ✓
- [ ] Nome é "Harpia" (não harpia-lang, HarpiaLang)
- [ ] Extensão `.hrp` usada consistentemente
- [ ] Exemplos de código são claros e educativos
- [ ] Links para docs estão ativos

### Acessibilidade ✓
- [ ] Alt text em imagens
- [ ] Labels claros em botões
- [ ] Navegação com teclado possível
- [ ] Não depende apenas de cor para entender

---

## 7. Template para Novos Materiais

### Template: README para Novo Repositório

```markdown
# [Nome do Repositório]

[1-2 linhas descrevendo o projeto]

## 🚀 Quick Start

```bash
# Comando rápido de instalação/setup
```

## 📚 Documentação

- [Guia Completo](docs/GUIA.md)
- [Exemplos](examples/)
- [FAQ](docs/FAQ.md)

## 🤝 Contribuindo

Adoramos contribuições! Veja [CONTRIBUINDO.md](CONTRIBUINDO.md)
para orientações.

## 📜 Licença

MIT — Veja [LICENCA.txt](LICENCA.txt)

## Comunidade

- 💬 [Discussões](https://github.com/harpialang/harpia/discussions)
- 🦅 [Site oficial](https://harpia.dev)

---

Feito com ❤️ pela Comunidade Harpia

```

---

**Última atualização:** 16 de julho de 2026

Para mais informações, veja o Guia de Marca completo: `harpia-brand-guide.md`
```
