# Architecture Proposal: Trellis DSL Compiler & Macro Nodes

**Status**: 🚧 **PROPOSAL** (Draft)
**Date**: 2026-02-19
**Related Concepts**: DX, Macro Nodes, Compilation

## 1. Contexto

O Trellis Core é arquitetado como um **DFA (Deterministic Finite Automaton)**. Isso garante robustez, auditabilidade e "flaky-free workflows".

No entanto, essa pureza traz um custo de **Verbosidade**. Representar fluxos lineares simples (Pergunta -> Resposta -> Ação) exige múltiplos nós explícitos no grafo, o que pode prejudicar a Experiência do Desenvolvedor (DX).

## 2. Inspiração: Colang 2.0 (NeMo Guardrails)

A análise do **Colang 2.0** (NVIDIA Ace/NeMo) demonstra o valor de uma sintaxe densa para fluxos de conversação.

* **Compactness**: O Colang permite definir interações de múltiplos turnos em poucas linhas (`match` -> `send`).
* **Abstração**: O desenvolvedor foca no *fluxo*, não nos *estados*.

## 3. Solução: Macro Nodes (`type: flow`)

Para alcançar a expressividade de uma DSL dedicada sem abandonar a arquitetura DFA do Trellis, introduzimos o conceito de **Macro Nodes**.

### 3.1. Proposta de Sintaxe (YAML Sugar)

Em vez de declarar nós individuais, o desenvolvedor declara um "Fluxo":

```yaml
# Trellis (Future v0.8 Concept)
type: flow
steps:
  - match: "Oi"          # Expande para: Node WaitingForInput + Condition
  - say: "Olá"           # Expande para: Node Text Pass-through
  - await: "payment_tool" # Expande para: Node WaitingForTool
```

## 4. Estratégia Arquitetural: The "Graph Compiler" (Lowering Phase)

Para garantir consistência entre diferentes adapters (File, Redis, Database) e manter o Engine simples, implementamos uma fase de compilação.

### 4.1. O Processo de Lowering

A expansão de macros não ocorre no Loader (que apenas lê arquivos), mas sim num **Preprocessor/Compiler** que roda antes do Engine.

1. **Source**: O Loader carrega o grafo "Sintático" (contendo `type: flow` e outras abstrações de alto nível).
2. **Compiler**: O Compilador varre o grafo, identifica macro-nodes e realiza **Inlining**.
    * Um nó `flow` é explodido em múltiplos nós atômicos de DFA (`node_0` -> `node_1` -> `node_2`).
    * Links e referências são reescritos para apontar para os novos IDs gerados.
3. **Engine**: Recebe e executa apenas o grafo "Binário" (Expandido).

### 4.2. Benefícios

* **Zero Runtime Cost**: O Engine continua sendo uma máquina de estados simples e rápida.
* **Adapter Agnostic**: Se o grafo vier de um DB SQL ou Redis, o compilador funciona igual.
* **Simplicidade de Teste**: Podemos testar o Compilador isoladamente (Input Macro -> Output Nodes) e o Engine isoladamente (Input Nodes -> Transitions).
