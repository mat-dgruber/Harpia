#!/usr/bin/env bash
# ==============================================================================
# Script de Compilação e Instalação do Harpia com Atualização Automática do PATH
# ==============================================================================

set -e

DIR_ATUAL="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BIN_NOME="harpia"
INSTALACAO_DIR="$HOME/.harpia/bin"
BIN_DESTINO="$INSTALACAO_DIR/$BIN_NOME"

echo "🦅 Compilando o CLI do Harpia..."
cd "$DIR_ATUAL"
go build -o "$BIN_NOME" main.go

echo "📁 Garantindo diretório de instalação em $INSTALACAO_DIR..."
mkdir -p "$INSTALACAO_DIR"

echo "🚚 Instalando binário compilado em $BIN_DESTINO..."
mv "$BIN_NOME" "$BIN_DESTINO"
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