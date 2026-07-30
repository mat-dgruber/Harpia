# Especificação de Design: Harpia CLI - Fase C (Lockfile & Gerenciamento de Pacotes)

## Objetivo
Evoluir o sistema de gerenciamento de dependências da CLI do Harpia (`harpia pacote` / `harpia instalar`), introduzindo reprodutibilidade estrita de builds por meio de um arquivo de bloqueio de integridade (`harpia.lock` / `harpia.lock.json`), verificação de integridade via checksum SHA-256 e subcomandos completos de gerenciamento de pacotes (`adicionar`, `remover`, `listar`, `limpar`).

---

## Componentes

### 1. Formato e Resolução do Lockfile (`harpia.lock` / `harpia.lock.json`)
- **Responsabilidade**: Garantir instâncias de instalação idênticas e imutáveis em qualquer ambiente de desenvolvimento ou CI/CD.
- **Estrutura de Dados**:
  ```json
  {
    "versao_lock": "1.0",
    "pacotes": {
      "matematica": {
        "versao": "1.2.0",
        "url": "https://github.com/Harpia/pacotes/raw/main/matematica-1.2.0.zip",
        "sha256": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
      }
    }
  }
  ```
- **Suporte Duplo de Leitura**:
  - `parseLockfile()` realiza a leitura transparente do arquivo `harpia.lock` (formato declarativo Harpia/KV) ou `harpia.lock.json` (JSON estruturado).
  - A escrita/atualização salva prioritariamente em `harpia.lock` ou no formato já existente no projeto.

### 2. Verificação de Integridade por Checksum SHA-256
- **Responsabilidade**: Proteger projetos contra alterações maliciosas ou corrupção de arquivos de dependências no servidor remoto.
- **Funcionamento**:
  - Durante a extração em `baixarEExtrairPacote`, o hash SHA-256 do arquivo ZIP recebido via HTTP é calculado.
  - Se o pacote já constar em `harpia.lock`, o hash obtido é comparado com o travado. Em caso de divergência, a instalação é abortada emitindo o código de erro `HRP-LOCK-001: Falha na verificação de integridade SHA-256 do pacote`.

### 3. Subcomando `harpia pacote` e Subcomandos Filhos
- **Responsabilidade**: Prover interface completa para manuseio do ciclo de vida de dependências do Harpia.
- **Comandos**:
  - `harpia pacote adicionar <nome> [versao]`: Adiciona dependência em `pacote.json`/`pacote.hrp`, instala em `./pt_modulos/` e atualiza `harpia.lock`.
  - `harpia pacote remover <nome>`: Exclui a pasta em `./pt_modulos/`, remove a referência do manifesto e do lockfile.
  - `harpia pacote listar`: Lista as dependências instaladas e verifica seu status de integridade em relação ao lockfile.
  - `harpia pacote limpar`: Remove o cache local de pacotes temporários.
  - Alias unificado: `harpia instalar` executa a instalação do projeto com base em `pacote.json`/`pacote.hrp` respeitando as travas de `harpia.lock`.

---

## Testes Automatizados
- Testes unitários em `cmd/pacotes_test.go`:
  - `TestParseLockfileJSONeHRP`: Valida a desserialização de lockfiles em formato JSON e declarativo Harpia.
  - `TestVerificacaoIntegridadeSHA256`: Testa a rejeição de pacotes cujo hash difere do registrado no lockfile.
  - `TestComandoPacoteAdicionarERemover`: Valida a adição e remoção de pacotes com atualização do manifesto e do lockfile.
  - `TestInstalacaoDeterministicaComLockfile`: Garante que `harpia instalar` utiliza exatamente a versão e URL especificadas no lockfile.
