# Trellis

> "Faça uma coisa e faça bem feita. Trabalhe com fluxos de texto." - Filosofia Unix

**Trellis** é o "Cérebro Lógico" de um sistema de automação. Projetada como uma **Função Pura de Transição de Estado**, opera isolada de efeitos colaterais.

## Quick Start

### Instalação

```bash
git clone https://github.com/aretw0/trellis
cd trellis
go mod tidy
```

### Rodando o Golden Path (Demo)

```bash
# Geração dos dados de teste
go run ./cmd/gen-trail ./examples/golden-path

# Execução do Engine
go run ./cmd/trellis ./examples/golden-path
```

## Documentação

- [📖 Product Vision & Philosophy](./docs/PRODUCT.md)
- [🏗 Architecture & Technical Details](./docs/TECHNICAL.md)
- [📅 Roadmap & Planning](./docs/PLANNING.md)

## Estrutura

```text
trellis/
├── cmd/           # Entrypoints (trellis, gen-trail)
├── docs/          # Documentação do Projeto
├── internal/      # Implementação (Loam Adapter, Runtime)
└── pkg/           # Contratos Públicos (Domain, Ports)
```
