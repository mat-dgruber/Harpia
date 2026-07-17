## 📝 Descrição

Forneça um resumo claro e conciso das alterações propostas neste Pull Request e os motivos para sua implementação.

Link para a Issue relacionada (se aplicável): #

---

## 🛠️ Tipo de Alteração

Marque a opção apropriada:

- [ ] 🐛 Correção de Bug (Bug Fix)
- [ ] ✨ Nova Funcionalidade (Feature)
- [ ] ⚡ Otimização de Performance
- [ ] 📚 Atualização de Documentação / Manual
- [ ] 🧪 Adição ou Ajuste de Testes
- [ ] 🔧 Refatoração ou Ajustes de Tooling / CI/CD

---

## 🏗️ Clean Architecture & DDD Checklist

Por favor, verifique se suas alterações cumprem as regras de arquitetura do Harpia:

- [ ] As regras de negócio puras foram mantidas isoladas na camada `/dominio`?
- [ ] Modificações de persistência, APIs ou integrações com sistemas externos ficaram na camada `/infra`?
- [ ] A camada visual de visualização/SPA foi adicionada na pasta `/web`?
- [ ] Não há importações cíclicas ou violações de dependências entre as camadas (ex: domínio importando infraestrutura)?

---

## 🧪 Checklist de Testes

- [ ] Todos os testes nativos do Harpia foram executados (`harpia testar`) e passaram?
- [ ] Todos os testes do compilador/runtime Go foram executados (`go test ./...`) e passaram sem erros?
- [ ] Foram adicionados novos testes nativos cobrindo o bug ou a funcionalidade adicionada?

---

## 📖 Checklist de Documentação

- [ ] A documentação inline foi inserida usando comentários com três barras (`///`) nos novos métodos/funções?
- [ ] O manual técnico (`Manual.md`) foi devidamente atualizado se houver novas sintaxes ou stdlibs?
- [ ] Se aplicável, a alteração foi registrada no `ROADMAP.md`?
