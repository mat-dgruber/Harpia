#!/bin/zsh

# Rebuild automático do compilador Harpia e distribuição nos caminhos globais.
# Executa este script toda vez que você modificar arquivos Go do compilador
# (lexer, parser, hrp, cmd) para propagar as mudanças para o seu terminal.

set -e

COMPILADOR_DIR="/Users/matheus.diniz_1/Documents/GitHub/Harpia/Harpia"
BINARIOS_DESTINO=(
    "/Users/matheus.diniz_1/go/bin/harpia"
    "/usr/local/bin/harpia"
)

echo "🔧 Compilando Harpia a partir de: $COMPILADOR_DIR"
cd "$COMPILADOR_DIR"

go build -o ./harpia .

for destino in "${BINARIOS_DESTINO[@]}"; do
    dir_destino=$(dirname "$destino")
    if [ -w "$dir_destino" ]; then
        cp ./harpia "$destino"
        echo "✅ Atualizado: $destino"
    else
        echo "⚠️  Sem permissão de escrita em $dir_destino (tentando com sudo)"
        sudo cp ./harpia "$destino" && echo "✅ Atualizado (sudo): $destino"
    fi
done

# Limpa o cache do executável (zsh guarda hash para PATH rápido)
rehash 2>/dev/null || hash -r 2>/dev/null || true

echo ""
echo "✅ Compilador Harpia atualizado em todos os caminhos do sistema."
echo "💡 Confirme com: harpia --version"
