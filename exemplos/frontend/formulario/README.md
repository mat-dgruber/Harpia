# Exemplo: Formulário de Contato

SPA que demonstra validação reativa de campos com Sinais e eventos `aoMudar`.

## Conceitos demonstrados

- `sinal()` por campo (nome, email, mensagem)
- `derivado()` para validação computada
- `aoMudar={...}` mapeado para `oninput` no DOM

## Como rodar

```bash
portuscript compilar --alvo=web --entrada=main.ptst --saida=dist
# Abra dist/index.html no browser
```
