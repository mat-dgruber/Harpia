#!/usr/bin/env bash
# ponytail: script de instalação local minimalista
set -e
cd "$(dirname "$0")"
go build -o harpia main.go
mkdir -p "$HOME/.harpia/bin"
cp harpia "$HOME/.harpia/bin/"
echo "Harpia compilado e instalado globalmente com sucesso!"