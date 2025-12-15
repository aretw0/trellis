# Product: Trellis

> "Faça uma coisa e faça bem feita. Trabalhe com fluxos de texto." - Filosofia Unix

**Trellis** é o "Cérebro Lógico" de um sistema de automação. Projetada como uma **Função Pura de Transição de Estado**, ela opera isolada de efeitos colaterais, processando apenas estruturas de dados e retornando intenções.

## Filosofia e Identidade

### Por que "Fluxos de Texto"?

Embora o **Loam** guarde os arquivos, é o **Trellis** que dá *vida* e *fluxo* a eles. Uma conversa não é estática; é um fluxo de intenções textuais. O Trellis é o filtro que transforma esse texto bruto em próximo passo lógico.

### O "Unix Way"

Trellis não é um framework monolítico; é um **filtro**.

- **Input**: Estado Atual + Grafo de Decisão + Input do Usuário.
- **Processamento**: Determinação determinística do próximo passo.
- **Output**: Novo Estado + Ações Solicitadas.

### Estratégia: O que o Trellis NÃO É

Para manter a sanidade do projeto, definimos limites claros:

1. **Não é uma Linguagem de Programação**: Não haverá loops complexos, definição de funções ou matemática arbitrária no Markdown.
2. **Não é um Template Engine Genérico**: Evitaremos recriar o Jinja2/Liquid. A lógica deve ser delegada, não embutida.
3. **Não é um Banco de Dados**: O Trellis consome estado, mas não gerencia persistência complexa (isso é trabalho do Loam ou do Host).

### Visão de Futuro: The Toolmaker's Tool

Para responder onde o Trellis quer chegar: **Ele deve ser a ferramenta que os criadores de ferramentas usam.**

Seja para criar um CLI interativo, um bot de WhatsApp, ou um wizard de instalação, o problema é sempre o mesmo: *Gerenciar o fluxo de conversa*. O Trellis resolve isso de forma agnóstica.

**O Trellis é o Guarda de Trânsito, não o Motor do Carro.**

- Ele apenas diz *para onde ir* (Próximo Nó).
- Ele não *dirige o carro* (não faz chamadas de API, não processa pagamentos).

Isso permite que ele seja usado tanto em um script bash simples quanto em um backend Go complexo para Telegram.

## Trellis na Era da IA (Agentes)

O Trellis ocupa um espaço crítico na arquitetura de **Agentes de IA (LLMs)**, atuando como o padrão **"Deterministic Guardrails"**.

- **O Problema**: Agentes puramente baseados em LLM são criativos, mas não determinísticos. Eles "alucinam" fluxos, esquecem regras de negócio ou inventam descontos.
- **A Solução Híbrida (Neuro-Symbolic)**:
  - **LLM (Cérebro Criativo)**: Gera o texto, traduz a intenção do usuário, mantem a empatia.
  - **Trellis (Espinha Dorsal Lógica)**: Mantém o Estado e define as Transições permitidas.
  - **MemoryLoader**: Permite injetar grafos dinamicamente em memória, ideal para ambientes serverless ou onde o fluxo é gerado on-the-fly.

> "Use o Trellis para garantir que o Agente siga o processo de compliance, e o LLM para garantir que ele seja educado enquanto faz isso."
