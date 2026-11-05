#!/usr/bin/env bash
# ==============================================================================
# Script de Compilação e Instalação do Harpia com Atualização Automática do PATH
# ==============================================================================

set -e

DIR_ATUAL="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BIN_NOME="harpia"
INSTALACAO_DIR="$HOME/.harpia/bin"
BIN_DESTINO="$INSTALACAO_DIR/$BIN_NOME"

REPO="mat-dgruber/Harpia"

echo "📁 Garantindo diretório de instalação em $INSTALACAO_DIR..."
mkdir -p "$INSTALACAO_DIR"

if [ -f "$DIR_ATUAL/main.go" ]; then
    # Dentro do código-fonte: compila localmente.
    echo "🦅 Compilando o CLI do Harpia..."
    cd "$DIR_ATUAL"
    go build -o "$BIN_NOME" main.go
    echo "🚚 Instalando binário compilado em $BIN_DESTINO..."
    mv "$BIN_NOME" "$BIN_DESTINO"
else
    # Executado via 'curl ... | bash': baixa o binário da última release e confere o SHA-256.
    SO="$(uname -s)"     # Darwin | Linux
    ARQ="$(uname -m)"    # arm64 | x86_64 | aarch64
    [ "$ARQ" = "aarch64" ] && ARQ="arm64"
    ARQUIVO="${SO}_${ARQ}.tar.gz"

    TAG="$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" | sed -n 's/.*"tag_name": *"\([^"]*\)".*/\1/p' | head -n 1)"
    if [ -z "$TAG" ]; then
        echo "❌ Não foi possível descobrir a última versão do Harpia." >&2
        exit 1
    fi
    VERSAO="${TAG#v}"
    BASE="https://github.com/$REPO/releases/download/$TAG"
    TMP="$(mktemp -d)"
    trap 'rm -rf "$TMP"' EXIT

    echo "⬇️  Baixando Harpia $TAG ($ARQUIVO)..."
    curl -fsSL "$BASE/$ARQUIVO" -o "$TMP/$ARQUIVO"
    curl -fsSL "$BASE/Harpia_${VERSAO}_checksums.txt" -o "$TMP/checksums.txt"

    ESPERADO="$(grep " ${ARQUIVO}\$" "$TMP/checksums.txt" | awk '{print $1}')"
    if command -v sha256sum >/dev/null 2>&1; then
        OBTIDO="$(sha256sum "$TMP/$ARQUIVO" | awk '{print $1}')"
    else
        OBTIDO="$(shasum -a 256 "$TMP/$ARQUIVO" | awk '{print $1}')"
    fi
    if [ -z "$ESPERADO" ] || [ "$ESPERADO" != "$OBTIDO" ]; then
        echo "❌ Falha na verificação SHA-256 de $ARQUIVO (esperado: ${ESPERADO:-ausente}, obtido: $OBTIDO)." >&2
        exit 1
    fi
    echo "🔐 SHA-256 conferido."

    tar -xzf "$TMP/$ARQUIVO" -C "$TMP"
    echo "🚚 Instalando binário em $BIN_DESTINO..."
    mv "$TMP/$BIN_NOME" "$BIN_DESTINO"
fi
chmod +x "$BIN_DESTINO"

# Atualização automática do PATH no shell
SHELL_PROFILE=""
if [ -n "$ZSH_VERSION" ] || [ "$(basename "$SHELL")" = "zsh" ]; then
    SHELL_PROFILE="$HOME/.zshrc"
elif [ -n "$BASH_VERSION" ] || [ "$(basename "$SHELL")" = "bash" ]; then
    if [ -f "$HOME/.bash_profile" ]; then
        SHELL_PROFILE="$HOME/.bash_profile"
    else
        SHELL_PROFILE="$HOME/.bashrc"
    fi
fi

EXPORT_LINE="export PATH=\"\$HOME/.harpia/bin:\$PATH\""

if [[ ":$PATH:" != *":$INSTALACAO_DIR:"* ]]; then
    echo "⚙️ Adicionando $INSTALACAO_DIR ao PATH..."
    if [ -n "$SHELL_PROFILE" ]; then
        if ! grep -qs "$INSTALACAO_DIR" "$SHELL_PROFILE"; then
            echo "" >> "$SHELL_PROFILE"
            echo "# Harpia Programming Language" >> "$SHELL_PROFILE"
            echo "$EXPORT_LINE" >> "$SHELL_PROFILE"
            echo "✅ Linha adicionada a $SHELL_PROFILE!"
        else
            echo "ℹ️ O caminho já consta em $SHELL_PROFILE."
        fi
    else
        echo "⚠️ Não foi possível detectar o arquivo de perfil do shell (~/.zshrc / ~/.bashrc)."
        echo "Adicione manualmente a seguinte linha ao seu shell:"
        echo "  $EXPORT_LINE"
    fi
else
    echo "✅ O caminho $INSTALACAO_DIR já está presente no PATH!"
fi

echo ""
echo "🎉 Harpia instalado com sucesso!"
echo "📍 Local do binário: $BIN_DESTINO"
echo "💡 Para carregar as alterações imediatamente no terminal atual, execute:"
if [ -n "$SHELL_PROFILE" ]; then
    echo "   source $SHELL_PROFILE"
fi