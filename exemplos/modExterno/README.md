# Módulos de Extensões Externas (.so)

O Harpia oferece suporte nativo e maduro para carregamento dinâmico de **Plugins de Extensões** compilados em Go, permitindo estender e expor recursos complexos com a performance bruta da linguagem Go.

---

## 🏗️ Como Compilar

Para compilar o arquivo Go como uma biblioteca dinâmica compartilhada compatível com o interpretador, utilize o compilador do Go informando a flag de modo de construção de plugin `-buildmode=plugin`:

```bash
# Compila main.go gerando o plug-in dinâmico externos.so
go build -buildmode=plugin -o externos.so main.go

# Padrão genérico de compilação:
# go build -buildmode=plugin -o <nome_do_modulo>.so <arquivo_entrada>.go
```

> **Nota de Compatibilidade**:  
> Os plug-ins dinâmicos em Go (extensões `.so`) são suportados nativamente em plataformas baseadas em Unix (como Linux e macOS), mas possuem limitações operacionais no Windows.
