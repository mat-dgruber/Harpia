#!/usr/bin/env python3
"""Gera llms.txt (índice, formato llmstxt.org) e llms-full.txt (conteúdo completo) a partir
dos documentos do repositório, para que não fiquem cópias manuais desatualizadas do Manual.

Uso (na pasta Harpia/): python3 scripts/gerar_llms.py
"""
import pathlib

RAIZ = pathlib.Path(__file__).resolve().parent.parent
BRUTO = "https://raw.githubusercontent.com/mat-dgruber/Harpia/main/"
PORTAL = "https://harpia-lang.web.app"

# (arquivo, descrição) na ordem em que entram no llms-full.txt
DOCUMENTOS = [
    ("README.md", "Visão geral, instalação e primeiros passos"),
    ("Manual.md", "Manual de referência: sintaxe, CLI, stdlib, frontend, VM e segurança"),
    ("CHANGELOG.md", "Mudanças por versão e notas de migração"),
    ("SECURITY.md", "Política de segurança e como reportar vulnerabilidades"),
]
ADRS = sorted((RAIZ / "docs" / "adrs").glob("*.md"))

CABECALHO = """# Harpia

> Harpia é uma linguagem de programação em português, com VM própria escrita em Go, biblioteca padrão
> completa (HTTP, banco de dados, IA, cripto) e frontend reativo com sinais e JSX transpilado para JavaScript.

- Arquivos `.hrp`; imports no formato `de "modulo" importe nome`.
- Blocos com chaves, condições sem parênteses: `se x > 1 { ... }`; `;` é opcional.
- Funções globais como `imprimir`, `tamanho`, `tipo`, `sequencia`, `sinal` não precisam de importação.
- Erros: `tente { } capture (erro) { } finalmente { }`; para sinalizar, `assegura(condicao, "mensagem")`.
"""


def gerar_indice():
    linhas = [CABECALHO, "## Documentação", ""]
    linhas += [f"- [{arq}]({BRUTO}{arq}): {desc}" for arq, desc in DOCUMENTOS]
    linhas += ["", "## Decisões de arquitetura", ""]
    linhas += [f"- [{a.stem}]({BRUTO}docs/adrs/{a.name})" for a in ADRS]
    linhas += ["", "## Optional", "", f"- [Portal com referência da API e tutoriais]({PORTAL}/api)", ""]
    return "\n".join(linhas)


def gerar_completo():
    partes = [CABECALHO]
    for arq in [d for d, _ in DOCUMENTOS] + [str(a.relative_to(RAIZ)) for a in ADRS]:
        partes.append(f"\n\n---\n\n<!-- fonte: {arq} -->\n\n" + (RAIZ / arq).read_text(encoding="utf-8").strip())
    return "".join(partes) + "\n"


if __name__ == "__main__":
    (RAIZ / "llms.txt").write_text(gerar_indice(), encoding="utf-8")
    (RAIZ / "llms-full.txt").write_text(gerar_completo(), encoding="utf-8")
    print("llms.txt e llms-full.txt gerados")
