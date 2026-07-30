# Especificação de Design: Harpia CLI - Fase B (Build, Release, Cross-compilação e Self-Updater)

## Objetivo
Elevar a maturidade dos comandos de compilação, distribuição e atualização da CLI do Harpia (`harpia compilar`/`construir` e `harpia atualizar`/`atualize`), eliminando dependências de utilitários externos do SO (`curl`, `tar`, `unzip`), adicionando suporte nativo a cross-compilação e viabilizando a injeção de metadados de build (`ldflags`).

---

## Componentes

### 1. Injeção de Metadados de Build (`cmd/versao.go`)
- **Responsabilidade**: Exibir informações detalhadas da versão compilada da CLI.
- **Variáveis Globais**:
  - `Versao` (default: `"dev"`)
  - `GitCommit` (default: `"desconhecido"`)
  - `DataBuild` (default: `"desconhecida"`)
- **Funcionamento**: Permite injeção em tempo de compilação via `go build -ldflags "-X github.com/mat-dgruber/Harpia/cmd.Versao=v1.2.0 ..."`.
- **Formatação de Saída**: Suporta saída amigável padrão ou `--formato=json`.

### 2. Auto-atualização Atômica 100% Go Nativo (`cmd/atualizar.go`)
- **Responsabilidade**: Checar e instalar novas versões da CLI diretamente do GitHub Releases.
- **Funcionalidades**:
  - Alias unificados: `harpia atualizar` e `harpia atualize`.
  - Flag `--checar`: Consulta a API do GitHub (`/repos/mat-dgruber/Harpia/tags`) e informa se há atualização sem realizar o download.
  - Download streaming via stdlib `net/http` com indicador de progresso TTY.
  - Extração nativa na stdlib: `archive/zip` para Windows e `archive/tar` + `compress/gzip` para Unix (Linux/macOS).
  - Substituição atômica de executável em execução: renomeia o binário ativo para `.old` antes de gravar o novo (necessário para SOs com trava de arquivo em execução, como o Windows) e limpa o temporário.

### 3. Cross-Compilação & Build AOT (`cmd/compilar.go`)
- **Responsabilidade**: Compilar scripts Harpia para alvos Web, Nativo, WASM ou WASI, e fornecer o alias `harpia construir`.
- **Funcionalidades**:
  - Alias: `harpia construir` aponta para a mesma estrutura do comando de compilação.
  - Flags de Target de Sistema Operacional e Arquitetura: `--os` (`darwin`, `linux`, `windows`) e `--arch` (`amd64`, `arm64`, `386`), populando `GOOS` e `GOARCH` em compilações nativas AOT.
  - Isolamento de Streams: Redirecionamento de saída de erros e dados via `cmd.OutOrStdout()` e `cmd.ErrOrStderr()` com controle via `RunE`.

---

## Testes Automatizados
- Testes unitários em `cmd/atualizar_test.go`:
  - Validação de comparação de versões (`semver`).
  - Teste do servidor mock HTTP para simular a API de releases e download de pacotes ZIP e TAR.GZ.
  - Validação da flag `--checar`.
- Testes unitários em `cmd/compilar_test.go`:
  - Validação das flags de cross-compilação `--os` e `--arch`.
  - Verificação da execução limpa via `RunE`.
