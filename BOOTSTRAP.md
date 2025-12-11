# BOOTSTRAP: Trellis

> "Faça uma coisa e faça bem feita. Trabalhe com fluxos de texto." - Filosofia Unix

**Trellis** é o "Cérebro Lógico" de um sistema de automação. Projetada como uma **Função Pura de Transição de Estado**, ela opera isolada de efeitos colaterais, processando apenas estruturas de dados e retornando intenções.

---

## 1. Filosofia e Identidade

### The Unix Way

Trellis não é um framework monolítico; é um **filtro**.

- **Input**: Estado Atual + Grafo de Decisão + Input do Usuário.
- **Processamento**: Determinação determinística do próximo passo.
- **Output**: Novo Estado + Ações Solicitadas.

### Arquitetura Hexagonal (Ports & Adapters)

O *Core* da Trellis não conhece banco de dados, não conhece HTTP e não conhece CLI. Ele define **Portas** (Interfaces) que o mundo externo deve satisfazer.

1. **Driver Ports (Entrada)**: Como o mundo fala com a Trellis (ex: `Evaluate(input)`).
2. **Driven Ports (Saída)**: O que a Trellis precisa do mundo (ex: `GraphLoader`, `ActionDispatcher`).

---

## 2. Integração com Loam (O Bibliotecário)

**[Loam](https://github.com/aretw0/loam)** é o nosso motor de persistência escolhido, mas a Trellis **não deve depender** diretamente dele em seu pacote `core`. Recomendamos o uso do `@mcp:github-mcp-server` para complementar informações sobre o Loam.

### Estratégia de Desacoplamento

A Trellis define uma interface para carregar definições de nós. O Loam é apenas um detalhe de implementação que satisfaz essa interface.

```go
// pkg/ports/loader.go
package ports

// GraphLoader é a porta que a Trellis usa para buscar definições.
// Note que a Trellis não sabe O QUE é "Loam", apenas que algo entrega bytes.
type GraphLoader interface {
    GetNode(id string) ([]byte, error)
}
```

**Como conectar (Wiring na `main`):**
A aplicação hospedeira (`cmd/app`) importará ambos (`trellis` e `loam`) e fará a ponte:

1. Instancia o `loam.Service`.
2. Cria um `LoamAdapter` que implementa `ports.GraphLoader`.
3. Injeta o adapter na `trellis.Engine`.

> **Nota sobre o Loam**: O Loam é um "Embedded Transactional Document Engine". Ele nos fornece atomicidade e versionamento (Git-backed). Use-o para garantir que as definições do grafo sejam robustas, mas mantenha a Trellis agnóstica a arquivos `.md` ou `.json` específicos até onde for possível (o Compiler resolve isso).

---

## 3. Escopo Funcional de Bootstrapping

Para começar com o pé direito, focaremos no **MVP (Minimum Viable Prototype)** que valida a arquitetura:

### Core Domain (`pkg/domain`)

- **`Node`**: A unidade básica de lógica (Wiki-style ou Logic-style).
- **`State`**: Onde estamos no grafo e quais variáveis temos memória `map[string]any`.
- **`Transition`**: A regra que leva de `Node A` -> `Node B`.

### Compiler (`internal/compiler`)

- Responsável por transformar `[]byte` (vindo do Loader/Loam) em `struct Node`.
- Validação estática de links (Dead Links Check).

### Runtime (`internal/runtime`)

- A função `Step`.
- Não executa side-effects! Retorna uma lista de `ActionRequest`.
- **Action Dispatching**: O Host recebe `ActionRequest {Type: "CLI", Payload: "Print..."}` e decide o que fazer.

---

## 4. Plano de Ação (Sequential Thinking)

Recomendamos o uso da ferramenta `@mcp:sequential-thinking` para implementar funcionalidades complexas, garantindo que cada decisão arquitetural seja validada antes de escrever código.

1. **Fase 1: The Core Types**
    - Definir as structs anêmicas e interfaces no `pkg/domain` e `pkg/ports`.
    - Garantir que não há imports externos.

2. **Fase 2: The Adapters (Loam & Memory)**
    - Implementar um `InMemoryLoader` para testes unitários rápidos.
    - Implementar o `LoamLoader` para integração real.

3. **Fase 3: The Engine**
    - Implementar a lógica de match de transição.
    - Implementar interpolação de variáveis simples.

---

## 5. Estrutura de Diretórios Sugerida

```text
trellis/
├── cmd/
│   └── trellis/       # Entrypoint (Wiring dos Ports e Adapters)
├── internal/          # Detalhes de implementação (Privado)
│   ├── compiler/      # Parsers (Markdown/YAML -> Domain)
│   ├── runtime/       # Step execution engine
│   └── adapters/      # Implementações (Loam, Memory)
├── pkg/               # Contratos Públicos (Safe to import)
│   ├── domain/        # Node, Transition, Action
│   └── ports/         # Interfaces (Driver & Driven)
└── go.mod
```

> **Nota sobre `internal`**: Adotaremos o diretório `internal` para garantir que a aplicação hospedeira (ou usuários da lib) não dependam de implementações concretas do *Compiler* ou *Runtime*, forçando o uso das interfaces definidas em `pkg/ports`.

---

> **Mantra do Projeto**: "A complexidade deve ser empurrada para as bordas. O núcleo permanece puro."
