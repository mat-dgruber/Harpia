#!/usr/bin/env bash

set -euo pipefail

aplicar_estilo() {
    local cor="$1"
    local negrito="$2"
    local texto="$3"

    local estilo=""
    local remover_estilo="\033[0m"

    if [[ -t 1 && ${COLORTERM:-} == "truecolor" ]]; then
        case "$cor" in
        "vermelho") cor='\e[38;2;255;0;0m' ;;
        "verde") cor='\e[38;2;0;255;0m' ;;
        "amarelo") cor='\e[38;2;255;255;0m' ;;
        "branco") cor='\e[38;2;255;255;255m' ;;
        *) cor="" ;;
        esac
    else
        case "$cor" in
        "vermelho") cor='\033[0;31m' ;;
        "verde") cor='\033[0;32m' ;;
        "amarelo") cor='\033[0;33m' ;;
        "branco") cor='\033[0;2m' ;;
        *) cor="" ;;
        esac
    fi

    if [[ "$negrito" == "true" ]]; then
        estilo='\033[1m'
    fi

    printf "%b" "${estilo}${cor}${texto}${remover_estilo}"
}

log() {
    local nivel="$1"
    local mensagem="$2"

    local cor=""
    local negrito="true"

    case "$nivel" in
    "ERRO") cor="vermelho" ;;
    "INFO") cor="branco" ;;
    "AVISO") cor="amarelo" ;;
    "SUCESSO") cor="verde" ;;
    *) cor="branco"; negrito="false" ;;
    esac

    echo -n "[ "
    aplicar_estilo "$cor" "$negrito" "$nivel"
    echo -n " ]: "
    aplicar_estilo "$cor" "false" "$mensagem"
    echo

    if [[ "$nivel" == "ERRO" ]]; then
        exit 1
    fi
}

if [[ $# -gt 1 ]]; then
    log "ERRO" "Foi recebido mais argumentos do que o esperado. Caso deseje instalar uma versão específica, use por exemplo: v0.1.0."
fi

if ! command -v curl >/dev/null; then
    log "ERRO" "O comando curl é essencial para o processo, mas ele não pôde ser achado."
fi

case $(uname -sm) in
"Darwin x86_64") target="Darwin_x86_64" ;;
"Darwin arm64") target="Darwin_arm64" ;;
"Linux aarch64") target="Linux_arm64" ;;
*) target="Linux_x86_64" ;;
esac

sufixo=".tar.gz"

if ! command -v tar >/dev/null; then
    log "ERRO" "O comando tar é essencial para o processo, mas ele não pôde ser achado."
fi

GITHUB=${GITHUB-"https://github.com"}
repo_github="$GITHUB/mat-dgruber/Harpia"
arquivo_compactado="$target$sufixo"

if [[ $# = 0 ]]; then
    harpia_uri=$repo_github/releases/latest/download/$arquivo_compactado
else
    harpia_uri=$repo_github/releases/download/$1/$arquivo_compactado
fi

raiz_harpia="${RAIZ_HARPIA:-$HOME/.harpia}"
diretorio_binario="$raiz_harpia/bin"
executavel="$diretorio_binario/harpia"

if [[ ! -d $diretorio_binario ]]; then
    mkdir -p "$diretorio_binario" ||
        log "ERRO" "Falha ao criar o diretório de instalação \"$diretorio_binario\""
fi

arquivo_temp=$(mktemp -t harpia.XXXXXXXXXX)
trap 'rm -f "$arquivo_temp"' EXIT

log "DEBUG" "Iniciando download do arquivo compactado"
curl --fail --location --progress-bar --output "$arquivo_temp" "$harpia_uri" ||
    log "ERRO" "Falha ao baixar o Harpia de \"$harpia_uri\""

log "DEBUG" "Iniciando descompactação"
tar -xf "$arquivo_temp" -C "$diretorio_binario"
log "SUCESSO" "Parece que a descompactação foi um sucesso"

chmod +x "$executavel"

log "SUCESSO" "Parabéns, agora você tem o Harpia disponível em \033[1m$executavel\033[0m"

refresh_command=""
if [[ ":$PATH:" == *":$diretorio_binario:"* ]]; then
	log "INFO" "Agora você pode usar o comando 'harpia --help' para ter um guia de comandos"
else
	case $(basename "$SHELL") in
    fish)
        commands=(
            "set --export DIRETORIO_HARPIA $raiz_harpia"
            "set --export PATH \$DIRETORIO_HARPIA/bin \$PATH"
        )
        fish_config=$HOME/.config/fish/config.fish
        if [[ -w $fish_config ]]; then
            if grep -q "configurações harpia" "$fish_config" 2>/dev/null; then
                log "INFO" "O caminho já está configurado em \"$fish_config\""
                refresh_command="source $fish_config"
            else
                cp "$fish_config" "${fish_config}.harpia.bak"
                if ! {
                    echo -e '\n# configurações harpia'
                    for command in "${commands[@]}"; do echo "$command"; done
                } >>"$fish_config"; then
                    mv "${fish_config}.harpia.bak" "$fish_config"
                    log "ERRO" "Falha ao gravar no arquivo \"$fish_config\". Backup restaurado."
                fi
                log "INFO" "Adicionado o caminho \"$diretorio_binario\" ao \$PATH em \"$fish_config\" (backup criado em \"${fish_config}.harpia.bak\")"
                refresh_command="source $fish_config"
            fi
        else
            log "AVISO" "Adicione manualmente os comandos ao $fish_config:"
            for command in "${commands[@]}"; do log "INFO" "  $command"; done
        fi
        ;;
    zsh)
        commands=(
            "export DIRETORIO_HARPIA=$raiz_harpia"
            "export PATH=\"\$DIRETORIO_HARPIA/bin:\$PATH\""
        )
        zsh_config=$HOME/.zshrc
        if [[ -w $zsh_config ]]; then
            if grep -q "configurações harpia" "$zsh_config" 2>/dev/null; then
                log "INFO" "O caminho já está configurado em \"$zsh_config\""
                refresh_command="source $zsh_config"
            else
                cp "$zsh_config" "${zsh_config}.harpia.bak"
                if ! {
                    echo -e '\n# configurações harpia'
                    for command in "${commands[@]}"; do echo "$command"; done
                } >>"$zsh_config"; then
                    mv "${zsh_config}.harpia.bak" "$zsh_config"
                    log "ERRO" "Falha ao gravar no arquivo \"$zsh_config\". Backup restaurado."
                fi
                log "INFO" "Adicionado o caminho \"$diretorio_binario\" ao \$PATH em \"$zsh_config\" (backup criado em \"${zsh_config}.harpia.bak\")"
                refresh_command="exec $SHELL"
            fi
        else
            log "AVISO" "Adicione manualmente os comandos ao $zsh_config:"
            for command in "${commands[@]}"; do log "INFO" "  $command"; done
        fi
        ;;
    bash|*)
        commands=(
            "export DIRETORIO_HARPIA=$raiz_harpia"
            "export PATH=\$DIRETORIO_HARPIA/bin:\$PATH"
        )
        bash_configs=("$HOME/.bashrc" "$HOME/.bash_profile")
        if [[ ${XDG_CONFIG_HOME:-} ]]; then
            bash_configs+=("$XDG_CONFIG_HOME/.bash_profile" "$XDG_CONFIG_HOME/.bashrc")
        fi
        set_manually=true
        for bash_config in "${bash_configs[@]}"; do
            if [[ -w $bash_config ]]; then
                if grep -q "configurações harpia" "$bash_config" 2>/dev/null; then
                    log "INFO" "O caminho já está configurado em \"$bash_config\""
                    refresh_command="source $bash_config"
                    set_manually=false
                    break
                else
                    cp "$bash_config" "${bash_config}.harpia.bak"
                    if ! {
                        echo -e '\n# configurações harpia'
                        for command in "${commands[@]}"; do echo "$command"; done
                    } >>"$bash_config"; then
                        mv "${bash_config}.harpia.bak" "$bash_config"
                        log "ERRO" "Falha ao gravar no arquivo \"$bash_config\". Backup restaurado."
                    fi
                    log "INFO" "Adicionado o caminho \"$diretorio_binario\" ao \$PATH em \"$bash_config\" (backup criado em \"${bash_config}.harpia.bak\")"
                    refresh_command="source $bash_config"
                    set_manually=false
                    break
                fi
            fi
        done
        if [[ $set_manually = true ]]; then
            log "AVISO" "Adicione manualmente os comandos ao seu arquivo de configuração de shell:"
            for command in "${commands[@]}"; do log "INFO" "  $command"; done
        fi
        ;;
    esac
fi

echo
log "INFO" "Para um bom início, execute:"
echo
if [[ $refresh_command ]]; then
    log "INFO" "  $refresh_command"
fi
log "INFO" "  harpia --help"
echo
log "SUCESSO" "Finalmente chegamos ao fim"
log "INFO" "Considere também deixar uma estrelinha no nosso repositório $repo_github"