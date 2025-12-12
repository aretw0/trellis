# Technical: Trellis Architecture

## Arquitetura Hexagonal (Ports & Adapters)

O *Core* da Trellis não conhece banco de dados, não conhece HTTP e não conhece CLI. Ele define **Portas** (Interfaces) que o mundo externo deve satisfazer.

### 1. Driver Ports (Entrada)

A API primária para interagir com o engine.

- `Engine.Step(state, input)`: Executa um ciclo de transição.

### 2. Driven Ports (Saída)

As interfaces que o engine usa para buscar dados.

- `GraphLoader.GetNode(id)`: Abstração para carregar nós. O **Loam** implementa isso via adapter.

## Estrutura de Diretórios e Decisões

```text
trellis/
├── cmd/
│   └── trellis/       # Entrypoint (Wiring dos Ports e Adapters)
├── internal/          # Detalhes de implementação (Privado)
│   ├── adapters/      # Implementações (Loam, Memory) e Loaders
│   └── runtime/       # Engine de execução
├── pkg/               # Contratos Públicos (Safe to import)
│   ├── domain/        # Node, Transition, Action (Structs puras)
│   └── ports/         # Interfaces (Driver & Driven)
└── go.mod
```

### O Papel do Loam

O **Loam** atua como bibliotecário e camada de persistência.

- **Responsabilidade**: Garantir integridade e fornecer documentos normalizados (`DocumentModel`).
- **Trellis Adapter (`LoamLoader`)**: Converte `DocumentModel` para JSON/Structs que o Compiler entende.

### Estratégia de Compilação

O Compiler do Trellis valida estaticamente:

- Integridade do JSON/Markdown.
- Links mortos (em breve).
