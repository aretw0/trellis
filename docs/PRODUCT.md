# Produto: Trellis

> "Agentes de IA sem Estado são apenas chatbots criativos. Agentes com Estado são Sistemas de Automação."

**Trellis** é o "Neuro-Symbolic Backbone" (Espinha Dorsal Neuro-Simbólica) para Automação e Agentes de IA.

> **Analogy**: Assim como o **React** gerencia a complexidade da UI através de um DOM Virtual, o **Trellis** gerencia a complexidade do Comportamento do Agente através de uma Máquina de Estados Virtual.

Projetada como uma **Função Pura de Transição de Estado**, ela opera isolada de efeitos colaterais, garantindo que "O que deveria acontecer" sempre corresponda ao "O que aconteceu".

## Filosofia e Identidade

### Por que "Máquinas de Estado"?

Embora o **Loam** guarde os arquivos, é o **Trellis** que garante a integridade do processo. Agentes de IA não precisam apenas processar texto; precisam respeitar *regras*.

O Trellis transforma a "Conversa Probabilística" (LLM) em "Transição Determinística" (DFA). Ele garante que o sistema só transite de "Pagamento" para "Envio" se o pagamento for confirmado, independente do quão persuasivo o LLM seja.

### Arquitetura: The Pure Kernel

Trellis não é um framework monolítico; é um **Kernel Puro**.

- **Input**: Estado Atual + Intenção (LLM/User) + Regras.
- **Process**: Computação determinística da próxima transição.
- **Output**: Novo Estado Auditável.

### Estratégia: Limites de Design (Constraints)

Para manter o foco do projeto, definimos o que ele **NÃO É**:

1. **Não é uma Linguagem de Programação**: Não haverá loops complexos, definição de funções ou matemática arbitrária no Markdown.
2. **Não é um Template Engine Genérico**: Evitaremos recriar o Jinja2/Liquid. A lógica deve ser delegada, não embutida.
3. **Não é o Banco de Dados da Aplicação**: O Trellis persiste o **Estado de Execução** (Durable Execution), mas não substitui seu banco de dados de negócio. Ele armazena "Onde estou e contexto temporário", não "O histórico de pedidos da empresa".

### Escalabilidade & Organização

À medida que os fluxos crescem, a organização se torna crítica. O Trellis suporta **Sub-Grafos e Namespaces** nativamente. Isso permite que equipes dividam grandes fluxos monolíticos em pequenos módulos independentes (pastas), conectados via `jump_to`.

### Posicionamento: Critical Infrastructure

Para responder onde o Trellis quer chegar: **Ele deve ser a infraestrutura crítica que Toolmakers confiam.**

Seja para criar um CLI interativo, um bot de WhatsApp, ou um Agente Autônomo Enterprise, o problema é sempre o mesmo: *Garantir que o fluxo siga as regras*.

**O Trellis é o Guarda de Trânsito, não o Motor do Carro.**

- Ele apenas diz *para onde ir* (Próximo Nó).
- Ele não *dirige o carro* (não faz chamadas de API, não processa pagamentos).

### Durable Execution (Reality)

O Trellis é uma plataforma de **Durable Execution** (v0.6+). Ele permite que fluxos "durmam" por dias e acordem exatamente onde pararam (Time-Travel), habilitando padrões **SAGA** (transações longas com rollback automático).

Isso permite que ele seja usado tanto em um script bash simples quanto em um backend Go complexo para Telegram com requisitos de missão crítica.

## Trellis na Era da IA (Agentes)

O Trellis ocupa um espaço crítico na arquitetura de **Agentes de IA (LLMs)**, atuando como o padrão **"Deterministic Guardrails"**.

- **O Problema**: Agentes puramente baseados em LLM são criativos, mas não determinísticos. Eles "alucinam" fluxos, esquecem regras de negócio ou inventam descontos.
- **A Solução: Controle Híbrido (Criativo + Lógico)**:
  - **LLM**: Gera o texto, traduz a intenção do usuário e mantém a empatia.
  - **Trellis**: Fornece a estrutura rígida. Mantém o Estado e define as Transições permitidas.
  - **MemoryLoader**: Permite injetar grafos dinamicamente em memória, ideal para ambientes onde o fluxo é gerado on-the-fly.

> "Use o Trellis para garantir que o Agente siga o processo de compliance, e o LLM para garantir que ele seja educado enquanto faz isso."

## Related Guides

- [Running HTTP Server](./guides/running_http_server.md): Guide for using Trellis as a stateless backend for Agents.
- [Running MCP Server](./guides/running_mcp_server.md): Connect Trellis to Claude/Cursor as a Tool Provider.
- [Manual SAGA Pattern](./guides/manual_saga_pattern.md): How to implement transactions and undo logic.
- [Interactive Inputs](./guides/interactive_inputs.md): Patterns for Human-in-the-loop workflows.
