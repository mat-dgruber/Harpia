# Changelog

Todas as mudanças notáveis do Harpia são documentadas aqui.

O formato segue [Keep a Changelog](https://keepachangelog.com/pt-BR/1.1.0/) e o versionamento segue [SemVer 2.0.0](https://semver.org/lang/pt-BR/).

## [1.1.0] — 2026-11-05

Rodada de correções do review geral de segurança e arquitetura (`docs/ai/reports/2026-11-04-review-max-geral-seguranca-arquitetura.md`). Compatível com a 1.0.x.

### Segurança
- **Playground e `harpia servir`** escutam só em `127.0.0.1` (`servir --host` para expor de propósito), recusam origem externa, DNS rebinding e CSRF (`/api/executar` exige JSON da própria origem), rodam em sandbox com tempo limite e não servem dotfiles (`.env`, `.git/`).
- **`harpia executar --sandbox`** (novo): bloqueia arquivos, processos, rede e variáveis de ambiente.
- **Sandbox *fail-closed*:** a permissão é resolvida pelo módulo chamador (não pelo contexto global) e vale para `arquivos`, `sistema`, `http`, `pdf`, `bd`, `ia`, `soquete`, `email` e `ffi`.
- **`sistema.executarCodigoHarpia`** (usado pelo playground do portal) roda isolado: `--sandbox`, diretório temporário, ambiente mínimo, 10s e 1 MB de saída.
- **Gerenciador de pacotes:** nome validado (path traversal), timeout, limite de download e de descompressão (zip bomb), sem setuid, https obrigatório, índice oficial em `mat-dgruber/harpia-pacotes` com `sha256` obrigatório (`HARPIA_REGISTRO` para espelhos).
- **Servidor HTTP:** erros e pânicos não vazam ao cliente (Problem Details RFC 9457 com `trace_id`); `requisitar` limita a resposta a 10 MB.
- **Cripto:** `cifrar` deriva a chave com scrypt e sal (envelope `v2:`); `decifrar` continua lendo o formato antigo.
- **Console `/docs`:** token e histórico de requisições em `sessionStorage`.
- **Toolchain Go 1.26.10** e dependências atualizadas: `govulncheck` 31 → 0 vulnerabilidades alcançáveis.

### Corrigido
- Parser travava (estouro de pilha) com o operador `!` e com `funcao f(` sem `)`.
- `se x { a = 1 }` (instrução terminada por `}`) era erro de sintaxe.
- Operador `**` não funcionava (existia só no parser).
- Recursão infinita derrubava o processo: agora `ErroDePilha` acima de 2000 chamadas.
- Overflow de `Inteiro` dava a volta em silêncio: agora `ErroDeLimite`.
- Handlers HTTP concorrentes corrompiam mapas globais (até queda fatal): execução serializada por servidor (ADR 0002).
- `/readyz` ficava em 503 para sempre depois de um único erro; rotas de mesma força eram escolhidas ao acaso; middleware podia rodar duas vezes.
- Função `assincrono` chamada com o contexto encerrado derrubava o processo.
- Fallback VM → interpretador comparava texto de erro (podia reexecutar o programa).
- `harpia executar sub/app.hrp` falhava sempre; `--assistir` rodava duas execuções em paralelo (agora reinicia por subprocesso).
- `transacao` não fazia rollback em pânico; `executar` devolvia id antigo em `UPDATE`.
- Módulos `autenticacao`, `fila` e `seguranca` não eram importáveis pelo CLI.
- Import em pasta com `.so` no nome (`proj.solar/`) falhava.
- `harpia versao` sempre mostrava `dev`; `atualizar` escolhia a tag errada e anunciava uma atualização que não fazia.
- `email.enviar`: `MAIL FROM:<<nil>>`, vários destinatários num único `RCPT`, assunto UTF-8 corrompido.
- Build web terminava com sucesso mesmo com erro de sintaxe em rota ou layout.
- Instalação via `curl ... | bash` do README dava 404 e só funcionava com o código-fonte: agora baixa a última release e confere o SHA-256.
- Escapes em textos (`\n`, `\t`, `\"`, `\\`, `\{`) não eram decodificados.
- `1 == 1.0` era assimétrico e `Decimal` não aceitava `<`, `<=`, `>`, `>=`.
- `imprimir(tipo(x))` mostrava a struct Go em vez do nome do tipo.
- Pânico numa função nativa derrubava o processo: agora vira `RuntimeErro` capturável. `extende`, `pop` e `junta` com argumentos comuns não quebram mais.
- `retentativa.executar` devolve o resultado da função; `esquema.criar` aceita tipos em minúsculas.
- `mapa.chaves()`, `valores()` e `atualizar()` falhavam ("recebeu Modulo"); `"chave" em mapa` não funcionava.
- `imprimir([1, 2])` mostrava `1 2` e o repr de `[[1, 2]]` saía achatado.
- `app.obter(rota, middleware, handler)` respondia 500: middlewares por rota agora recebem `(req, res, proximo)`.
- `para ... em` na VM quebrava sobre Mapa, `sequencia()` e Texto.
- Frontend: toda string interpolada aparecia com aspas sobrando; `\` em texto JSX gerava JS inválido; o runtime web lançava `TypeError` em todo elemento com `href`/`src` e não renderizava o componente.
- Comentário no fim da linha (`var x = 1 # nota`) quebrava a sintaxe.
- `assegura(cond, "msg")` com parênteses perdia a mensagem.
- `sinal` no backend ignorava o definidor.
- `harpia checar` acusava "já declarada" em variáveis com nome de módulo (`bd`, `tarefas`, `doc`).

### Adicionado
- Pipe `|>` pode continuar na linha seguinte.

### Adicionado
- `onde(col, "==", v)` como sinônimo de `"="`; encadeamento fluente com o ponto no início da linha.
- `app.cors([...])` e `HARPIA_DOCS=0` no servidor HTTP; `executarDetalhado` em conexões SQL.
- `sistema.executar(cmd, args, {tempoLimite})` devolve também `codigo`.
- CI: job `govulncheck` e fuzz noturno do parser; `SECURITY.md`.

### Alterado
- Documentação padronizada na sintaxe de importação suportada: `de "modulo" importe nome`.
