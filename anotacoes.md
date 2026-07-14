# Anotações e Especificações de Sintaxe do Portuscript

Este arquivo serve como referência rápida para o design de sintaxe da linguagem.

## 🔤 Filosofia Geral

Sintaxe baseada em C/JavaScript/Go (blocos delimitados por chaves `{}`), mas com condições limpas sem parênteses `()` (estilo Go/Rust), e nomenclatura/palavras-chave 100% em português brasileiro.

---

## 1. Variáveis e Constantes

```portuscript
var nome: Texto = "Portuscript"  // Var com tipo opcional
constante VERSAO = "1.0.0"       // Constante obrigatória e imutável
```

## 2. Estruturas de Controle (Sem parênteses nas condições)

```portuscript
se idade >= 18 {
    imprimir("Maior de idade")
} senao {
    imprimir("Menor de idade")
}

enquanto condicao {
    executarAlgo()
}
```

## 3. Funções e Parâmetros Avançados

```portuscript
funcao soma(a: Inteiro, b: Inteiro = 0) -> Inteiro {
    retorne a + b
}

// Chamando com parâmetros nomeados
var resultado = soma(a = 10, b = 5)
```

## 4. Classes e Orientação a Objetos (Herança Simples)

```portuscript
classe Animal {
    inicializar(self, nome: Texto) {
        self.nome = nome
    }

    falar(self) {
        retorne "Som genérico"
    }
}

classe Cachorro estende Animal {
    falar(self) {
        retorne "Au! Eu sou " + self.nome
    }
}

var pet = nova Cachorro("Rex")
imprimir(pet.falar()) // "Au! Rex"
```

## 5. Operador de Canal (Pipes `|>`)

```portuscript
// Sintaxe fluída para manipulação em cadeia
var textoFormatado = "  ola mundo  " |> removerEspacos |> maiusculo
// Equivalente a: maiusculo(removerEspacos("  ola mundo  "))
```

## 6. Tratamento de Erros

```portuscript
tente {
    var resultado = 10 / 0
} capture (erro: ErroDivisao) {
    imprimir("Erro capturado: " + erro.mensagem)
} finalmente {
    imprimir("Sempre executa")
}
```

## 7. Reatividade (Sinais e Estado Global)

```portuscript
// Sinais locais
var [contador, definirContador] = sinal(0)

efeito(funcao() {
    imprimir("O contador agora é: " + contador())
})

definirContador(1) // Dispara o efeito automaticamente

// Estado Global (Armazém)
var estadoGlobal = armazem({
    usuarioLogado: falso
})
```

## 8. Testes Nativos na Linguagem (TDD/SDD)

```portuscript
testar "deve somar dois numeros corretamente" {
    assegura(soma(2, 2) == 4)
}
```
