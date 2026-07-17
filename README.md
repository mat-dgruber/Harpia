# 🦅 Harpia

<p align="center">
  <img src="docs/assets/logo_harpia_refinado.jpg" alt="Logo Harpia" width="220" />
</p>

<p align="center">
  <strong>Linguagem de Programação Reativa, 100% Brasileira e Focada em Arquitetura Limpa</strong>
</p>

<p align="center">
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-00A86B.svg" alt="License" /></a>
  <a href="docs/BRAND_GUIDELINES.md"><img src="https://img.shields.io/badge/Diretrizes_de_Marca-Ver_Manual-F2A900.svg" alt="Brand Guidelines" /></a>
  <a href="ROADMAP.md"><img src="https://img.shields.io/badge/Desenvolvimento-Roadmap-D45D34.svg" alt="Roadmap" /></a>
</p>

---

## 📖 O que é a Harpia?

A **Harpia** é uma linguagem de programação brasileira moderna, focada em desenvolvimento ágil de ponta a ponta (Full Stack). Ela foi projetada para ir além do ensino de lógica de programação, permitindo a criação de sistemas profissionais de nível industrial — incluindo frontends SPA reativos, backends corporativos e APIs seguras — tudo utilizando a nossa língua nativa.

Representada pela imponente águia-real das Américas, a marca simboliza **soberania, precisão cirúrgica, força e foco absoluto**.

---

## ⚡ Filosofia e Conceito: O Método Ponytail

A linguagem é orientada pela **Filosofia Ponytail** (preguiçosa com o código redundante, atenta com a leitura e pragmática com a solução):

1. **YAGNI (You Aren't Gonna Need It):** Eliminar qualquer código ou feature que não seja estritamente necessário.
2. **Modularização sem Cíclicos:** O compilador barra dependências cíclicas estaticamente antes da execução.
3. **Erros Didáticos e IA:** Mensagens estruturadas (`HRP-XXXX`) com dicas em português e suporte a explicações guiadas por IA local (`harpia erro explicar`).
4. **Legibilidade Semântica:** Uso do operador de canais (`|>`) e termos nativos do ecossistema brasileiro (como `raiz`, `ninho`, `copa` e `correnteza` para as bibliotecas padrão).

---

## 🏗️ Clean Architecture e DDD Nativo

A Harpia foi desenhada sob a premissa de estruturar projetos robustos por padrão. Ao iniciar novos projetos corporativos, o compilador gera a árvore organizacional baseada em **DDD (Domain-Driven Design)** e **Clean Architecture**:

```text
meu-app/
├── dependencias.json        -> Manifesto de dependências e configurações
├── main.hrp                 -> Ponto de entrada (Bootstrapper)
├── dominio/                 -> Regras de Negócio Isoladas
│   ├── modelos/             -> Entidades e Objetos de Valor (ex: usuario.hrp)
│   └── servicos/            -> Validações e regras de domínio (ex: validador.hrp)
├── infra/                   -> Detalhes de Tecnologia (Banco de Dados, APIs)
│   ├── bd/                  -> Conexões e repositórios (ex: sqlite.hrp)
│   └── api/                 -> Clientes de requisição externa
├── web/                     -> Camada de Apresentação (Frontend SPA)
│   ├── rotas/               -> Páginas com File-system Routing (ex: index.hrp)
│   ├── componentes/         -> Componentes UI reutilizáveis (ex: botao.hrp)
│   └── estilos/             -> Folhas de estilo locais ou globais
└── testes/                  -> Camada de Testes Automatizados
```

---

## ✨ Características Principais & Performance

- **Direct-Threaded JIT VM:** Bytecodes dinamicamente traduzidos em chamadas Go nativas em tempo de execução, otimizando o loop de decodificação.
- **Pool de Alocação Eden:** Pré-boxeamento de inteiros curtos (de `-100` a `2000`) em $O(1)$ para aniquilar pressões desnecessárias do Garbage Collector.
- **Reatividade Nativa (SPA):** Transpilação reativa eficiente para a web (`--alvo=web`) baseada em Sinais (`var [contador, definirContador] = sinal(0)`), Efeitos e Estado Global.
- **Estilização Nativa:** Blocos de estilo CSS integrados nativamente e classes utilitárias na estrutura de marcação.
- **Contrato RPC Automático:** Comunicação simplificada entre o Front-end e o Back-end sem a necessidade de APIs manuais complexas.

---

## 🛠️ Caixa de Ferramentas (CLI)

O interpretador de linha de comando da Harpia disponibiliza utilitários completos:

- **`harpia`**: Inicia o REPL ou TUI gráfica de depuração com inspetor de memória e VM ativo.
- **`harpia executar [arquivo.hrp]`**: Roda o script de forma instantânea.
- **`harpia compilar [entrada.hrp] --alvo=web`**: Transpila o frontend para `/dist` gerando o build estático.
- **`harpia testar [caminho]`**: Executa testes de unidade e integração declarados diretamente no código com o bloco `testar`.
- **`harpia diagramar`**: Mapeia as relações de importações e cospe um diagrama em sintaxe Mermaid, alertando se houver violações de arquitetura limpa.

---

## 📚 Documentação & Guias

Para mais informações sobre as regras de desenvolvimento do projeto:

- 🎨 **[Diretrizes de Marca & Identidade Visual](docs/BRAND_GUIDELINES.md)**
- 💡 **[Exemplos Práticos de Aplicação e Código](docs/harpia-brand-examples.md)**
- 🚀 **[Guia de Contribuição](CONTRIBUTING.md)**
- 🗺️ **[Roadmap de Evolução](ROADMAP.md)**

---

<p align="center">
  <img src="docs/assets/fundo_harpia.jpg" alt="Wallpaper Harpia" width="100%" />
</p>

<p align="center">
  Feito com ❤️ pela Comunidade Brasileira de Programação 🇧🇷
</p>
