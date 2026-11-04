# Política de Segurança

Obrigado por ajudar a manter o Harpia seguro. Este documento explica quais versões recebem correções, como reportar uma vulnerabilidade de forma privada e o que esperar depois do relato.

## Versões suportadas

Somente a linha 1.x mais recente recebe correções de segurança. Quando uma nova versão é publicada, as anteriores deixam de ser corrigidas: atualize para a mais recente.

| Versão                         | Suporte de segurança |
| ------------------------------ | -------------------- |
| 1.x mais recente (hoje `v1.0.2`; a próxima será `v1.1.0`) | Sim |
| `v1.0.1` e `v1.0.0`            | Não                  |
| Anteriores a `v1.0.0`          | Não                  |

## Como reportar uma vulnerabilidade

**Não abra uma issue pública, discussão ou pull request descrevendo a falha.** Isso expõe quem usa o Harpia antes de existir uma correção.

Use exclusivamente o relato privado do GitHub:

1. Acesse a aba **Security** do repositório e clique em **Report a vulnerability**; ou
2. Abra diretamente: <https://github.com/mat-dgruber/Harpia/security/advisories/new>

O relato fica visível apenas para você e para o mantenedor até a divulgação.

### O que incluir no relato

- Versão do Harpia (`harpia versao`) ou o hash do commit usado. Em binários antigos, que mostram `dev`, informe a tag da release baixada.
- Sistema operacional e arquitetura (ex.: macOS 15 arm64, Ubuntu 24.04 x86_64, Windows 11).
- Componente afetado (CLI, VM, stdlib, playground, `harpia servir`, `harpia instalar`, LSP, extensão VS Code).
- Passos para reproduzir, incluindo comandos e flags usados.
- Prova de conceito mínima (um script `.hrp` curto ou a requisição HTTP exata).
- Impacto observado e o que um atacante conseguiria fazer.
- Se possível, uma sugestão de correção ou mitigação.

## Prazos de resposta

O Harpia é mantido por uma única pessoa, então os prazos abaixo são metas realistas, não garantias contratuais:

- **Confirmação de recebimento:** em até 5 dias úteis.
- **Avaliação inicial** (validade, severidade e escopo): em até 14 dias.
- **Correção**, conforme a severidade:
  - Crítica: o mais rápido possível, com meta de até 30 dias.
  - Alta: até 60 dias.
  - Média e baixa: na próxima release planejada.

Você receberá atualizações pelo próprio advisory do GitHub ao longo do processo.

### Divulgação coordenada

- O prazo padrão de divulgação é de **90 dias** a partir do relato, ou antes, se a correção já tiver sido publicada.
- Se a correção precisar de mais tempo, o prazo pode ser estendido em comum acordo.
- O advisory é publicado junto com a versão corrigida e, quando aplicável, com um identificador CVE.
- O pesquisador recebe **crédito público** no advisory e nas notas de versão, se assim desejar. Também é possível permanecer anônimo.

## Escopo

> **Nota de versão.** O modo `harpia executar --sandbox` e o bind exclusivo em `127.0.0.1` do playground e do `harpia servir` chegam na **v1.1.0**. Até `v1.0.2`, esses servidores escutam em todas as interfaces e não existe modo sandbox para scripts. Nessas versões, não rode código não confiável e use o playground/`servir` apenas em redes confiáveis.

### Dentro do escopo

- Bypass do sandbox (a partir da v1.1.0: `harpia executar --sandbox`, `BloquearArquivos`/`BloquearRede`): qualquer forma de um script em sandbox ler ou escrever arquivos, acessar a rede, ler variáveis de ambiente ou executar processos.
- Execução remota de código no playground web ou no `harpia servir` **na configuração padrão** (a partir da v1.1.0, escutando só em `127.0.0.1`), incluindo ataques vindos de páginas web abertas no navegador (CSRF, DNS rebinding e afins).
- Path traversal ou zip slip no gerenciador de pacotes (`harpia instalar`), permitindo escrever fora do diretório de destino.
- Injeção de SQL pelo QueryBuilder da stdlib (`bd`) quando usado conforme a documentação.
- Falhas criptográficas na stdlib (`cripto`, `autenticacao`): validação incorreta de JWT, aceitação de `alg: none`, comparações que vazam tempo, aleatoriedade fraca.
- Vazamento de segredos (tokens, chaves, variáveis de ambiente) em logs, mensagens de erro ou artefatos gerados pelo Harpia.
- Extensão VS Code (`vscode-harpia`) ou LSP executando comandos ou código sem ação explícita do usuário, por exemplo ao apenas abrir um arquivo ou workspace.
- Cadeia de suprimentos do índice de pacotes: falhas que permitam instalar um pacote diferente do pedido, burlar a verificação de `sha256` ou adulterar o lockfile.

### Fora do escopo

- **Scripts executados sem `--sandbox` lendo arquivos, rodando processos ou acessando a rede.** Isso é o comportamento esperado de uma linguagem de programação, assim como rodar um script Python ou Node.js. Rodar código não confiável sem sandbox equivale a executá-lo diretamente na sua máquina.
- Negação de serviço causada por scripts do próprio usuário (laços infinitos, consumo de memória, recursão profunda).
- Playground, `harpia servir` ou outras ferramentas de desenvolvimento expostas de propósito na rede (por exemplo, com `harpia servir --host 0.0.0.0`, a partir da v1.1.0). Essas ferramentas não foram feitas para produção; expô-las é responsabilidade de quem as configura.
- Vulnerabilidades em dependências sem um caminho comprovadamente alcançável a partir do Harpia. Relatos apenas com saída de scanners automáticos, sem prova de impacto, também ficam fora.
- Problemas que exigem acesso prévio à máquina ou à conta do usuário.

Em caso de dúvida se algo está no escopo, reporte mesmo assim pelo canal privado.

## Boas práticas para quem usa o Harpia

- **Rode código não confiável com sandbox:** `harpia executar --sandbox script.hrp` (v1.1.0 ou mais recente). Sem a flag, o script tem as mesmas permissões que o seu usuário. Em versões anteriores, não execute código não confiável.
- **Não exponha o playground nem o `harpia servir` na rede.** A partir da v1.1.0, o padrão é `127.0.0.1`: mantenha-o. Em versões anteriores, use-os só em redes confiáveis. Para publicar aplicações, use um servidor de produção adequado.
- **Fixe suas dependências:** versione o lockfile (`harpia.lock`) e confira os hashes `sha256` dos pacotes antes de atualizar.
- **Mantenha o Harpia atualizado** na versão 1.x mais recente para receber as correções de segurança.
