# Technical Reference: Trellis Architecture

## I. Fundamentos (Core Foundation)

Esta seção define os pilares arquiteturais, regras de design e estratégias que governam todo o projeto.

### 1. Definição Formal (Identity)

Tecnicamente, o Trellis é um **Reentrant Deterministic Finite Automaton (DFA) with Controlled Side-Effects**.

- **Reentrant**: O Engine pode ser serializado ("adormecido") e reidratado ("acordado") em qualquer estado estável sem perda de continuidade.
- **Deterministic**: Dado o mesmo Estado Inicial + Input + Resultado de Tools, o Engine *sempre* produzirá a mesma transição, eliminando "flaky workflows".
- **Managed Side-Effects**: Efeitos colaterais (IO, API calls) são delegados ao Host via *Syscalls* (`ActionCallTool`), garantindo que a lógica de transição permaneça pura e testável.

### 2. Arquitetura Hexagonal (Ports & Adapters)

O *Core* da Trellis não conhece banco de dados, não conhece HTTP e não conhece CLI. Ele define **Portas** (Interfaces) que o mundo externo deve satisfazer.
Essa arquitetura desacoplada torna o Trellis leve o suficiente para ser embutido em CLIs simples ou usado como biblioteca "low-level" dentro de frameworks de Agentes de IA maiores.

#### 2.1. Driver Ports (Entrada)

A API primária para interagir com o engine.

- `Engine.Render(state)`: Retorna a view (ações) para o estado atual e se é terminal.
- `Engine.Navigate(state, input)`: Computa o próximo estado dado um input.
- `Engine.Inspect()`: Retorna o grafo completo para visualização.

#### 2.2. Driven Ports (Saída)

As interfaces que o engine usa para buscar dados.

- `GraphLoader.GetNode(id)`: Abstração para carregar nós. O **Loam** implementa isso via adapter.
- `GraphLoader.ListNodes()`: Descoberta de nós para introspecção.

#### 2.2.1. Portas de Persistência (Store)

Interface experimental para "Durable Execution" (Sleep/Resume).

- `StateStore.Save(ctx, sessionID, state)`: Persiste o snapshot da execução.
- `StateStore.Load(ctx, sessionID)`: Hidrata uma sessão anterior.

#### 2.2.2. Session Manager (Orchestrator)

The `pkg/session` package acts as the orchestrator for state durability. It wraps the `StateStore` to add concurrency control (locking) and lifecycle management (atomic "Load or Create").

**Concurrency Strategy (Reference Counting)**:
To prevent memory leaks in high-traffic scenarios, the Manager uses a **Reference Counting** mechanism for session locks. Locks are created on demand and automatically deleted when the reference count drops to zero.

```mermaid
sequenceDiagram
    participant Caller
    participant Manager (Global)
    participant Entry (Ref)

    Caller->>Manager (Global): Acquire(ID)
    Manager (Global)->>Manager (Global): Lock Global -> Inc Ref -> Unlock Global
    Manager (Global)-->>Caller: Entry (Ref)

    Caller->>Entry (Ref): Lock()
    Note right of Caller: Critical Section
    Caller->>Entry (Ref): Unlock()

    Caller->>Manager (Global): Release(ID)
    Manager (Global)->>Manager (Global): Lock Global -> Dec Ref -> Del if 0 -> Unlock Global
```

#### 2.3. Diagrama de Arquitetura

```mermaid
graph TD
    Host[Host Application / CLI] -->|Driver Port| Engine
    MCP[MCP Client / Inspector] -->|Driver Port| Engine
    subgraph "Trellis Core"
        Engine[Engine - Runtime]
        Domain[Domain - Node, State]
    end
    Engine -->|Driven Port| Loader[GraphLoader Interface]
    Loader -.->|Adapter| Loam[Loam - File System]
    Loader -.->|Adapter| Mem[InMemory - Testing]
    
    Host -->|Uses| Store[StateStore Interface]
    Store -.->|Adapter| FileStore[FileStore - Local JSON]
    Store -.->|Adapter| Redis[Redis - Cloud]
    Store -.->|Adapter| Memory[In-Memory - Ephemeral]
```

### 3. Estrutura de Diretórios

```text
trellis/
├── cmd/
│   └── trellis/       # Entrypoint (CLI)
├── internal/          # Detalhes de implementação (Privado)
│   ├── adapters/      # Implementações (Loam, HTTP, MCP)
│   ├── presentation/  # TUI & Renderização
│   ├── runtime/       # Engine de execução
│   └── validator/     # Lógica de validação
├── pkg/               # Contratos Públicos (Safe to import)
│   ├── adapters/      # Adaptadores de Infraestrutura (Inmemory)
│   ├── domain/        # Core Domain (Node, State)
│   ├── ports/         # Interfaces (Driver & Driven)
│   ├── registry/      # Registro de Ferramentas
│   ├── runner/        # Loop de Execução e Handlers
│   └── session/       # Gerenciamento de Sessão e Locking
└── go.mod
```

### 4. Princípios de Design (Constraints)

Para evitar a "Complexidade Oculta", seguimos estas restrições:

#### 4.1. Logic-Data Decoupling

A lógica complexa **nunca** deve residir no grafo (Markdown).

- **Errado**: `condition: user.age > 18 && user.status == 'active'` (Exige parser complexo).
- **Correto**: `condition: is_adult_active` (O Host resolve e retorna bool).

> Veja [Interactive Inputs](../docs/guides/interactive_inputs.md) para detalhes sobre como o Host gerencia inputs.

#### 4.2. Strict Mode Compiler

O compilador deve ser implacável.

- Variáveis não declaradas resultam em erro de compilação.
- O objetivo é **Confiança Total**: Se compilou, não existem "Dead Ends" lógicos causados por typos.

#### 4.3. Convenção de Ponto de Entrada (Entry Point)

O Trellis segue a filosofia **Convention over Configuration** para o início do fluxo.

- **ID Obrigatório**: O fluxo SEMPRE começa no nó com ID `start`.
- **Resolução de Arquivo**: Por padrão, o `LoamLoader` busca por um arquivo chamado `start.md` (ou `start.json`) na raiz do diretório.
- **Sub-Grafos**: Ao pular para um sub-módulo (`jump_to: modules/auth`), o engine busca por `modules/auth/start.md`.

> **Nota**: Embora seja possível injetar um `State` inicial diferente via código (`engine.Navigate(ctx, customState, input)`), a CLI e os Runners padrão assumem `start` como entrypoint.

#### 4.4. Hot Reload Lifecycle (v0.6)

Com a introdução do `StateStore`, o ciclo de Hot Reload tornou-se "Stateful". Ao detectar uma mudança, o Engine é recarregado, mas o Runner tenta reidratar o estado anterior.

```mermaid
sequenceDiagram
    participant W as File Watcher
    participant C as CLI (RunWatch)
    participant E as Engine (New)
    participant S as SessionManager
    participant R as Runner

    Note over W, R: Loop de Desenvolvimento
    W->>C: Change Detected
    C->>E: Initialize New Engine
    alt Compile Error
        C->>C: Log & Wait for fix
    else Success
        C->>S: LoadOrStart(sessionID)
        S->>C: Return InitialState
        C->>C: Validate Node exists & Context
        C->>R: Run(Engine, InitialState)
    end
```

**Estratégias de Recuperação (Guardrails)**:

- **Missing Node**: Fallback para `start` se o nó atual for removido.
- **Validation Failure**: Pausa se novos `required_context` surgirem sem dados na sessão.
- **Type Mismatch**: Reseta o status de `WaitingForTool` se o nó mudar de tipo.

### 5. Design Constraints & Verified Limitations (v0.6)

This section maps the "Sober View" of the architectural trade-offs accepted in version 0.6.

#### 5.1. Session Concurrency (RefCounting Fragility)

To solve memory leaks without a heavy Garbage Collector, `pkg/session` uses **Reference Counting**:

- **Risk**: Depends strictly on `Acquire/Release` pairing. A developer error (phantom panic or missing defer) creates a permanent leak for that ID.
- **Bottleneck**: The `Manager` uses a **Global Mutex** (`mu`) to protect the lock map. In extreme concurrency (>100k Lock/Unlock ops/sec), this global lock becomes a contention point.
- **Decision**: Acceptable for CLI/Agent use cases. For high-scale SaaS, sharding the Manager would be required:
  - *Strategy*: Use `[ShardCount]*sync.Mutex` array.
  - *Logic*: `shardID = hash(sessionID) % ShardCount`.
  - *Benefit*: Reduces contention by factor of `ShardCount` (e.g., 256).

#### 5.2. Redis Lazy Indexing (Zombie Entries)

The Redis Adapter avoids background workers ("Serverless-friendliness"):

- **Mechanism**: The `List()` method cleans up expired entries from the ZSET Index.
- **Implication**: If `List()` is rarely called, the ZSET index may contain "Zombie Entries" (IDs whose actual keys have expired) until the next listing.
- **Cost**: `List()` incurs a write penalty (`ZREMRANGEBYSCORE`).

#### 5.3. FileStore Pruning (Manual Maintenance)

- **Constraint**: The local file store never deletes old `.json` sessions automatically.
- **Mitigation**: Rely on manual hygiene (`trellis session rm`) or external OS-level cron jobs. No auto-pruning logic exists to keep the binary simple.

### 6. Estratégia de Testes

Para garantir a estabilidade do Core enquanto o projeto evolui, definimos uma pirâmide de testes rígida:

#### 6.1. Níveis de Teste

1. **Core/Logic (Unit)**:
    - **Alvo**: `pkg/domain`, `internal/runtime`.
    - **Estilo**: Table-Driven Tests extensivos.
    - **Objetivo**: Garantir que a máquina de estados (Engine) transite corretamente dado um input, sem depender de IO ou Filesystem.
    - **Strict Serialization**: Inclui testes de regressão (`tests/serialization_test.go`) para garantir consistência de `json.Number` e tipos.

2. **Adapters (Contract Tests)**:
    - **Alvo**: `LoamLoader` vs `MemoryLoader`.
    - **Estilo**: Interface Compliance Tests.
    - **Objetivo**: O mesmo suite de testes deve rodar contra ambas as implementações para garantir que o comportamento seja idêntico. Se o `MemoryLoader` funciona, o `LoamLoader` deve funcionar igual.

3. **Integration (E2E)**:
    - **Alvo**: `cmd/trellis` (Blackbox).
    - **Estilo**: CLI Runners / JSON In-Out.
    - **Objetivo**: Simular um usuário real ou sistema externo interagindo com o binário. Verifica se a "cola" (Wiring) entre as camadas está funcionando.

4. **Snapshot Testing**:
    - **Alvo**: CLI Output / TUI.
    - **Objetivo**: Garantir que a experiência de cada pixel na tela não regrediu (útil para detectar quebras de layout).

---

## II. Mecânica do Core (Engine & IO)

Esta seção detalha o funcionamento interno do engine, ciclo de vida e tratamento de dados.

### 7. Ciclo de Vida do Engine (Lifecycle)

O Engine segue um ciclo de vida estrito de **Resolve-Execute-Update** para garantir previsibilidade.

```mermaid
sequenceDiagram
    participant Host
    participant Engine
    participant Loader

    Host->>Engine: Render(State)
    Engine->>Loader: GetNode(ID)
    Loader-->>Engine: Node
    Engine->>Engine: Interpolate Content & Tool Args (Deep)
    Engine-->>Host: Actions (View/ToolCall)
    
    Host->>Host: User Input / Tool Result
    
    Host->>Engine: Navigate(State, Input)
    Engine->>Loader: GetNode(ID)
    Loader-->>Engine: Node
    
    rect rgba(77, 107, 138, 1)
        note right of Engine: Update Phase
        Engine->>Engine: Apply Input (save_to) -> NewState
    end
    
    rect rgba(82, 107, 56, 1)
        note right of Engine: Resolve Phase
        Engine->>Engine: Evaluate Conditions (Transitions)
    end
    
    Engine-->>Host: NextState (with new ID)
```

**Fases do Ciclo:**

1. **Render (View)**: Carrega o nó, aplica interpolação profunda (incluindo argumentos de ferramentas) e retorna as ações. O estado *não* muda.
2. **Navigate (Update)**:
    - **Update**: Aplica o input ao contexto da sessão (se `save_to` estiver definido).
    - **Resolve**: Avalia as condições de transição baseadas no novo contexto.
    - **Transition**: Retorna o novo estado apontando para o próximo nó.

#### 7.1. Hot Reload Lifecycle (v0.6)

No modo `watch`, o Runner orquestra o recarregamento do motor e a reidratação do estado usando um `SignalContext` hierárquico.

```mermaid
sequenceDiagram
    participant W as Watcher (fsnotify)
    participant O as Orchestrator (internal/cli)
    participant S as SignalContext
    participant R as Runner (pkg/runner)

    Note over W, R: Ciclo de Hot Reload (Signal-Aware)
    W->>O: Evento: file.md alterado
    O->>S: Cancel(Reload)
    S->>R: ctx.Done() propagado
    
    par Graceful Shutdown
        R->>R: Interrompe IO (Stdin Block)
        R-->>O: Retorna ctx.Err() (Reload)
    and UI Update
        O->>O: Log "Change detected in file.md"
    end
    
    O->>O: Aguarda estabilização (100ms)
    O->>S: NewSignalContext()
    O->>R: Nova Iteração: Run(newCtx, engine, state)
    R->>R: Resume at 'CurrentNode'
```

**Estratégias de Recuperação (Guardrails):**

1. **Node Tipo 'tool' → 'text'**: Se o estado salvo era `WaitingForTool`, mas o nó foi alterado para `text` (ou deletado), o motor reseta o status para `Active` para evitar travamentos.
2. **Erro de Sintaxe**: Se o arquivo alterado contiver erro de sintaxe, o Runner aguarda a próxima correção sem derrubar o processo e registra o erro via `logger.Error`.
3. **Session Scoping**: No modo `watch`, se nenhum ID de sessão for fornecido, um ID determinístico baseado no hash do caminho do repositório (`watch-<hash>`) é gerado para evitar colisões entre projetos.

### 8. Protocolo de Efeitos Colaterais (Side-Effect Protocol)

O protocolo de side-effects permite que o Trellis solicite a execução de código externo (ferramentas) de forma determinística e segura.

#### 8.1. Filosofia: "Syscalls" para a IA

O Trellis trata chamadas de ferramenta como "Chamadas de Sistema" (Syscalls). O Engine não executa a ferramenta; ele **pausa** e solicita ao Host que a execute.

1. **Intenção (Intent)**: O Engine renderiza um nó do tipo `tool` e emite uma ação `CALL_TOOL`.
2. **Suspensão (Yield)**: O Engine entra em estado `WaitingForTool`, aguardando o resultado.
3. **Dispatch**: O Host (CLI, Servidor HTTP, MCP) recebe a solicitação e executa a lógica (ex: chamar API, rodar script).
4. **Resumo (Resume)**: O Host chama `Navigate` passando o `ToolResult`. O Engine retoma a execução verificando transições baseadas nesse resultado.

#### 8.2. Ciclo de Vida da Chamada de Ferramenta

```mermaid
sequenceDiagram
    participant Engine
    participant Host
    participant External as "External API/Script"

    Note over Engine: Estado: Active (Node A)
    Engine->>Host: Render() -> ActionCallTool(ID="tool_1", Name="calc", Args={op:"add"})
    
    Note over Engine: Estado: WaitingForTool (Pending="tool_1")
    
    Host->>External: Executa Ferramenta (Async)
    External-->>Host: Retorna Resultado (ex: "42")
    
    Host->>Engine: Navigate(State, Input=ToolResult{ID="tool_1", Success=true, Result="42"})
    
    Note over Engine: Valida ID & Resume
    Engine->>Engine: Avalia Transições do Node A (ex: if input == "42")
    Engine->>Host: NewState (Node B)
```

#### 8.3. Universal Dispatcher

Graças a este desacoplamento, a mesma definição de grafo pode usar ferramentas implementadas de formas diferentes dependendo do adaptador:

- **CLI Runner**: Executa scripts locais (`.sh`, `.py`) ou funções Go embutidas.
- **MCP Server**: Repassa a chamada para um cliente MCP (ex: Claude Desktop, IDE).
- **HTTP Server**: Webhooks que notificam serviços externos (ex: n8n, Zapier).

#### 8.4. Defining Tools in Loam

You can define available tools directly in the Node's frontmatter. This allows the Engine to be aware of the tool's schema (name, description, parameters) without needing hardcoded Go structs.

```yaml
type: text
tools:
  - name: get_weather
    description: Get current temperature
    parameters:
      type: object
      properties:
        city: { type: string }
---
The weather is...
```

#### 8.5. Reusable Tool Libraries (Polymorphic Design)

To support modularity, the `tools` key in Frontmatter is polymorphic. It accepts both inline definitions and string references to other files.

```yaml
tools:
  - name: local_tool         # Inline Definition
    description: ...
  - "modules/tools/math.md"  # Reference (Mixin)
```

##### Resolution Strategy

The `LoamLoader` implements a recursive resolution strategy with **Shadowing** (Last-Write-Wins).

```mermaid
flowchart TD
    Start([Resolve Tools]) --> Init[Init Visited Set]
    Init --> Iterate{Iterate Items}
    
    Iterate -->|String Import| CheckCycle{Cycle?}
    CheckCycle -- Yes --> Error(Error: Cycle Detected)
    CheckCycle -- No --> Load[Load Referenced File]
    Load --> Recurse[Recursive Resolve]
    Recurse --> MergeImport[Merge Imported Tools]
    MergeImport --> Iterate
    
    Iterate -->|Map Definition| Decode[Decode Inline Map]
    Decode --> MergeInline[Merge/Shadow Tool]
    MergeInline --> Iterate
    
    Iterate -- Done --> Flatten[Flatten Map to List]
    Flatten --> End([Return Tool List])
    
    style MergeInline stroke:#f66,stroke-width:2px,color:#f66
    style MergeImport stroke:#66f,stroke-width:2px,color:#66f
```

**Technical Constraints:**

1. **Polymorphism (`[]any`)**: The Loader accepts generic types to support this UX. This requires **manual schema validation** at runtime.
2. **Cycle Detection**: Recursive imports are guarded against infinite loops (`visited` set).
3. **Shadowing Policy**: Local definitions always override imported ones.

### 9. Runner & IO Architecture

The `Runner` serves as the bridge between the Core Engine and the outside world. It manages the execution loop, handles middleware, and delegates IO to an `IOHandler`.

It delegates signal handling to a dedicated **SignalManager** (`pkg/runner/signal_manager.go`) which ensures race-free context cancellation and signal resetting.

#### 9.1. Session Cycle

```mermaid
sequenceDiagram
    participant CLI
    participant Runner
    participant SessionManager
    participant Engine
    participant Store

    Note over CLI: PrintBanner() (Branding)
    CLI->>Runner: Run(sessionID)
    Runner->>SessionManager: LoadOrStart(sessionID)
    SessionManager->>Store: Load(sessionID)
    alt Session Exists
        Store-->>SessionManager: State
    else New Session
        SessionManager->>Engine: Start()
        Engine-->>SessionManager: State
        SessionManager->>Store: Save(InitialState)
    end
    SessionManager-->>Runner: State

    loop Execution Loop
        Runner->>Engine: Render(State)
        Engine-->>Runner: Actions (Text/Tools)
        Runner->>CLI: Output / Wait Input
        CLI-->>Runner: Input
        Runner->>Engine: Navigate(State, Input)
        Engine-->>Runner: NewState
        Runner->>Store: Save(NewState)
    end

    Note over CLI: logCompletion(nodeID)
    Note over CLI: handleExecutionError()
```

#### 9.2. Stateless & Async IO

Trellis supports two primary modes of operation:

1. **Text Mode** (`TextHandler`): For interactive TUI/CLI usage. Blocks on user input.
2. **JSON Mode** (`JSONHandler`): For headless automation and API integration.

**Key constraint for JSON Mode:**

- **Strict JSON Lines (JSONL)**: All inputs to the `JSONHandler` must be single-line JSON strings.
- **Async/Non-Blocking**: The handler reads from Stdin in a background goroutine, allowing cancellation (timeout/interrupt).
- **Signal Mapping (Context-Aware)**: The Runner monitors:
  - `signals.Context().Done()`: Explicit User Signal (SIGINT). Maps to `"interrupt"`.
  - `ctx.Done()` (Parent): External Orchestrator (Watch Reload). Treated as Clean Exit (no signal mapping).
  - `inputCtx.Done()` (Deadline): Maps to `"timeout"`.

#### 9.3. Semântica de Texto e Bloqueio

O comportamento de nós de texto segue a semântica de State Machine pura:

1. **Nodes de Texto (`type: text`)**: São, por padrão, **Non-Blocking (Pass-through)** para o Engine.
    - Se houver uma transição válida incondicional, transita *imediatamente*.
    - **Nota de UX**: A transição é imediata (Pass-through) em todos os modos. Se você deseja que o usuário leia o texto antes de continuar (pressione Enter), deve definir explicitamente `wait: true`.
2. **Pausas Explícitas**:
    - `wait: true`: Força pausa para input (ex: "Pressione Enter") em *ambos* os modos.
    - `type: question`: Pausa explícita aguardando resposta (hard step).

#### 9.4. Diagrama de Decisão (Input Logic)

```mermaid
flowchart TD
    Start([Engine.Render]) --> Content{Has Content?}
    Content -- Yes --> EmitRender[ActionRenderContent]
    Content -- No --> CheckInput
    EmitRender --> CheckInput

    CheckInput{Needs Input?}
    CheckInput -->|wait: true| YesInput
    CheckInput -->|type: question| YesInput
    CheckInput -->|input_type != nil| YesInput
    CheckInput -->|"Default (Pass-through)"| AutoNav[Navigate - State, Empty]

    YesInput --> EmitRequest[ActionRequestInput]
    EmitRequest --> Stop([Runner Pauses])

    AutoNav --> Result{Status?}
    Result -- Terminated --> Stop
    Result -- Active --> Next([Next State - Loop])
```

#### 9.5. Pattern: Stdin Pump (IO Safety)

Input handling in Go, especially with `os.Stdin`, is blocking by nature. To support **Timeouts** (cancelable reads) without blocking the main event loop or leaking "ghost readers" (race conditions where a stale goroutine eats input intended for the next step), `TextHandler` implements the **Stdin Pump** pattern.

- **Single Producer**: A persistent goroutine (`pump`) reads from the underlying Reader loop forever.
- **Buffered Channel**: Results (`string` or `error`) are sent to `inputChan`.
- **Select-based Consumption**: The `Input(ctx)` method listens to `inputChan`. If `ctx` times out, it stops listening, but the pump **remains active**, ready to serve the next call (type-ahead support).

```mermaid
flowchart LR
    Stdin[os.Stdin] -->|ReadString| Pump((Pump Goroutine))
    Pump -->|inputResult| Chan[inputChan]
    
    subgraph "Input(ctx) Call"
        Chan -->|Select| Consumer[Runner]
        Timer[Context Timeout] -.->|Cancel| Consumer
    end
    
    Consumer -->|Sanitized Input| Engine
```

> **Stewardship Note**: This pattern prevents multiple goroutines from fighting over `bufio.Reader`. The `Runner` automatically memoizes the handler instance to ensure that reusing a `Runner` instance also reuses the single Pump goroutine.

#### 9.6. Architectural Insight: Engine-bound vs Runner-bound

Para manter a arquitetura limpa, diferenciamos onde cada responsabilidade reside:

1. **Engine-bound (Passive/Logic)**:
    - *Exemplos*: `LifecycleHooks` (`OnNodeEnter`, `OnTransition`).
    - *Natureza*: O Engine apenas **emite** eventos sobre o que calculou. Ele não sabe quem está ouvindo e não espera resposta.
    - *Propósito*: Observabilidade pura.

2. **Runner-bound (Active/Control)**:
    - *Exemplos*: `StateStore` (Persistência), `ToolInterceptor` (Segurança), `SignalManager` (Interrupção).
    - *Natureza*: O Runner **orquestra** e decide se o fluxo deve continuar, pausar ou falhar.
    - *Propósito*: Controle do Ciclo de Vida e Integração com o Mundo Real (IO).

Essa separação garante que o Core permaneça uma Máquina de Estados Pura e Determinística, enquanto o Runner assume a responsabilidade pela "sujeira" (Timeouts, Discos, Sinais de SO).

#### 9.7. Estratégia de Persistência (Scope)

- **Workspace-first**: As sessões são armazenadas em `.trellis/sessions/` no diretório de trabalho atual.
- **Motivação**: Isolar sessões por projeto (como `.git` ou `.terraform`), facilitando o desenvolvimento e evitando colisões globais em ambientes multi-projeto.
- **Formato**: Arquivos JSON simples para facilitar inspeção e debugging manual ("Loam-ish").

#### 9.8. Session Management CLI (Chaos Control)

Para gerenciar o ciclo de vida dessas sessões persistentes, o Trellis expõe comandos administrativos ("Chaos Control"):

- **List (`ls`)**: Enumera sessões ativas no workspace.
- **Inspect**: Visualiza o Estado JSON puro (Current Node, Context, History) para debugging.
- **Remove (`rm`)**: Permite "matar" sessões travadas ou limpar o ambiente.

Essa camada é crucial para operações de longa duração, onde "desligar e ligar de novo" (resetar o processo) não é suficiente para limpar o estado.

> **Maintenance Note**: O FileStore não implementa *Auto-Pruning* (limpeza automática) de sessões antigas. Cabe ao administrador ou desenvolvedor executar `trellis session rm` periodicamente ou configurar scripts externos de limpeza (cron) se o diretório de sessões crescer excessivamente.

#### 10. Fluxo de Dados e Serialização

#### 10.1. Data Binding (SaveTo)

O Trellis adota a propriedade `save_to` para indicar a *intenção* de persistir a resposta de um nó no contexto da sessão.

```yaml
type: question
text: "Qual é o seu nome?"
save_to: "user_name" # Salva input em context["user_name"]
```

**Regras de Execução:**

1. **Precedência**: O valor é salvo no contexto *antes* de avaliar as transições.
2. **Imutabilidade**: O Engine realiza **Deep Copy** do Contexto a cada transição.
3. **Tipagem Preservada**: `save_to` armazena o input como recebido (`any`).

#### 9.2. Variable Interpolation

O Trellis adota uma arquitetura plugável para interpolação de variáveis via interface `Interpolator`.

- **Default Strategy**: Go Templates (`{{ .UserName }}`).
- **Legacy Strategy**: `strings.ReplaceAll` (`{{ Key }}`).

#### 10.3. Global Strict Serialization

O Trellis força **Strict Mode** em todos os adaptadores para resolver o problema do `float64` em JSON. Números são decodificados como `json.Number` ou `int64` para garantir integridade de IDs e timestamps.

#### 10.4. Data Contracts (Validation & Defaults)

**Standard Serialization (Snake Case):**

To ensure interoperability, the Engine serializes its state to JSON using `snake_case` keys (e.g., `current_node_id`, `pending_tool_call`), regardless of the internal Go struct naming. This allows for cleaner integration with external tools and manual session inspection.

**Fail Fast (Required Context):**

Nós servem como fronteiras de dados e podem impor contratos de execução:

**Fail Fast (Required Context):**

```yaml
required_context:
  - user_id
  - api_key
```

Se uma chave estiver faltando, o Engine aborta a execução com `ContextValidationError`.

**Valores Padrão (Mocking):**

Nós (convencionalmente `start`) podem definir valores de fallback para simplificar o desenvolvimento local:

```yaml
default_context:
  api_url: "http://localhost:8080"
```

#### 10.5. Initial Context Injection (Seed State)

Para facilitar testes automatizados e integração, o Trellis permite injetar o estado inicial.

- **API**: `Engine.Start(ctx, initialData map[string]any)`
- **CLI**: Flag `--context '{"user": "Alice"}'`
- **Configuração**: Use `trellis.WithEntryNode("custom_start")` para sobrescrever o ponto de entrada padrão ("start").
- **Precedência**: `initialData` (Runtime) > `default_context` (File).
- **Uso**: Dados injetados estão disponíveis imediatamente para interpolação (`{{.user}}`) no nó de entrada.

---

## III. Funcionalidades Estendidas (System Features)

Recursos avançados para escalabilidade, segurança e integração.

### 11. Escalabilidade: Sub-Grafos e Namespaces

Para escalar fluxos complexos, o Trellis suporta **Sub-Grafos** via organização de diretórios.

#### 11.1. Semântica `jump_to` vs `to`

- **`to`**: Transição local (mesmo arquivo/contexto).
- **`jump_to`**: Transição para um **Sub-Grafo** ou Módulo externo (mudança de contexto).

#### 11.2. IDs Implícitos e Normalização

- **Implicit IDs**: Arquivos em subdiretórios herdam o caminho como ID (ex: `modules/checkout/start`).
- **Normalization**: O Adapter normaliza todos os IDs para usar `/` (forward slash).

#### 11.3. Syntactic Sugar: Options

Atalho para menus de escolha simples.

- **Options**: Correspondência exata de texto. Avaliadas PRIMEIRO.
- **Transitions**: Lógica genérica. Avaliadas DEPOIS.

### 12. Segurança e Policies

#### 12.1. Interceptors (Safety Middleware)

Para mitigar riscos de execução arbitrária, o Runner aceita interceptadores:

```go
type ToolInterceptor func(ctx, call) (allowed bool, result ToolResult, err error)
```

- **ConfirmationMiddleware**: Solicita confirmação explícita (`[y/N]`). O trecho `metadata.confirm_msg` no nó pode personalizar o alerta.
- **AutoApproveMiddleware**: Para modo Headless/Automação.

#### 12.2. Error Handling (on_error)

Mecanismo robusto para recuperação de falhas em ferramentas.

- Se `ToolResult.IsError` for true:
  - O Engine **PULA** o `save_to` (evita context poisoning).
  - O Engine busca transição `on_error` ou `on_error: "retry"`.
  - Se houver `on_error`: Transita para o nó de recuperação.
  - Se não houver handler, o erro sobe (Panic/Fatal).

#### 12.3. Controle de Execução (Signals & Timeouts)

Mecanismos para controle de fluxo assíncrono e limites de execução.

**Timeouts (Sinal de Sistema):**

- **Definição**: `timeout: 30s` (declarativo no nó).
- **Mapping**: O Runner mapeia `context.DeadlineExceeded` automaticamente para o sinal `"timeout"`.
- **Fail Fast**: Se o sinal `"timeout"` não for tratado via `on_signal`, o Runner encerra a execução com erro (`timeout exceeded`). Isso evita loops infinitos ou estados zumbis.

**Sinais Globais (Interrupções):**

- **API**: `POST /signal` (e.g., Interrupt, Shutdown).
- **Handlers**: O Engine verifica `on_signal` no estado atual.
- **Workflow**:
  1. Runner detecta sinal via Context ou Input.
  2. Engine tenta criar transição (`on_signal`).
  3. **Sucesso**: Fluxo continua no novo estado.
  4. **Falha (Unhandled/Loader Error)**: Runner aborta execução (`fail fast`).

```mermaid
flowchart TD
    Start([Execute Tool]) --> Result{Result.IsError?}
    Result -- No --> Success[Apply save_to & Transitions]
    Result -- Yes --> HasHandler{Has on_error?}
    
    HasHandler -- Yes --> Recovery([Transition to on_error Node])
    HasHandler -- No --> FailFast{{STOP: UnhandledToolError}}
    
    style FailFast fill:#f00,stroke:#333,color:#fff
    style Recovery fill:#6f6,stroke:#333,color:#000
```

#### 12.4. System Context Namespace

O namespace `sys.*` é reservado no Engine.

- **Read-Only**: Templates podem ler (`{{ .sys.error }}`).
- **Write-Protected**: `save_to` não pode escrever em `sys` (proteção contra injeção).

#### 12.5. Global Signals (Interrupts)

O Trellis suporta a conversão de sinais do sistema operacional (ex: `Ctrl+C` / `SIGINT`) em transições de estado.

- **`on_signal`**: Define um mapa de sinais para nós de destino.
- **Engine.Signal**: Método que dispara a transição.

```yaml
type: text
wait: true
on_signal:
  interrupt: confirm_exit
```

Se o sinal "interrupt" for recebido enquanto o nó estiver ativo, o Engine transitará para `confirm_exit` em vez de encerrar o processo.

> **Consistency Note**: Quando um sinal dispara uma transição, o evento `OnNodeLeave` é emitido para o nó interrompido, mantendo a consistência do ciclo de vida.

#### 12.6. Extensibility: Signals & Contexts

O mecanismo de `on_signal` é a base para extensibilidade do fluxo via eventos:

- **System Contexts (Timeouts)**: `on_signal: { timeout: "retry_node" }`. (Implementado)
- **External Signals (Webhooks)**: `on_signal: { payment_received: "success" }`. Disparado via `POST /signal`. (Implementado)
- **Payload injection (Future)**: Injeção de dados junto com o sinal (ex: webhook payload -> `context.webhook_data`).

```mermaid
flowchart TD
    Start([User Input]) --> Wait{Waiting Input?}
    Wait -- Ctrl+C / Timeout --> Sig[SignalManager: Capture Signal]
    InputAPI([API / Webhook]) -.-> Sig
    Sig --> Engine[Engine.Signal]
    
    Engine --> Handled{Has on_signal?}
    Handled -- Yes --> Leave[Emit OnNodeLeave]
    Leave --> Transition[Transition to Target Node]
    
    Transition --> Reset[SignalManager: Reset Context]
    Reset --> Resume([Resume Execution])
    
    Handled -- No --> Exit{{Fail Fast: Exit Process}}
    
    style Sig fill:#783578,stroke:#333
    style Reset fill:#4a4a7d,stroke:#333
    style Exit fill:#f00,stroke:#333,color:#fff
```

#### 12.7. Input Sanitization & Limits

To ensure robust operation in production (especially in shared-memory environments like Kubernetes Pods), Trellis enforces limits on user input at the Runner layer. This applies globally to **all adapters** (CLI, HTTP, MCP).

- **Max Input Size**: Defaults to 4KB. Configurable via `TRELLIS_MAX_INPUT_SIZE`.
- **Control Characters**: Automatically strips dangerous ANSI/Control codes to prevent log poisoning.
- **Behavior**: Inputs exceeding the limit are **Rejected** (returning an error) rather than silently truncated, preserving state integrity ("Deterministic State").

See [Deployment Strategies](../docs/guides/deployment_strategies.md) for provisioning advice.

### 13. Adapters & Interfaces

#### 13.1. Camada de Apresentação

Responsável por converter visualmente o grafo e estados.

- **Trellis Graph**: Gera diagramas Mermaid.
  - **Start/Root** (`(( ))`): Nó inicial ou com ID "start".
  - **Question/Input** (`[/ /]`): Nós que exigem interação do usuário.
  - **Tool/Action** (`[[ ]]`): Nós que executam efeitos colaterais.
  - **Default** (`[ ]`): Nós de texto simples ou lógica interna.
  - **Timeouts** (`⏱️`): Anotação visual no label do nó.

**Arestas e Transições:**

- **Fluxo Normal** (`-->`): Transições padrão.
- **Salto de Módulo** (`-.->`): Transições entre arquivos (`jump_to`).
- **Sinais/Interrupções** (`-. ⚡ .->`): Transições disparadas por `on_signal`.

#### 13.1.1. Visual Debug Strategies (Visualizing State)

A flag `--session <id>` permite sobrepor o estado de uma sessão ao grafo estático.

**Implementação Atual (v0.6 - "Heatmap"):**

- **Modelo**: Conjunto de Nós Visitados (Set).
- **Estilo**: Nós visitados ficam azuis; nó atual fica amarelo.
- **Limitação (Caveat)**: Não representa a **ordem** nem a **frequência** de visita.
  - Se o fluxo fez `A -> B -> A -> C`, o grafo mostra `A`, `B` e `C` pintados.
  - Não é possível distinguir se o usuário veio de `B` ou `Start`.
  - Loops aparecem achatados.

**Evolução Futura (Vision):**

Para debugging forense de falhas complexas (Saga/Loops), o modelo visual precisará evoluir:

1. **Numbered Path (Badges)**: Adicionar badges (ex: `🔴 #1, #3`) aos nós para indicar a ordem da sequência de passos.
2. **Edge Highlighting**: Pintar as **arestas** percorridas. Desafio técnico: Mermaid não facilita ID em arestas.
3. **Sequence Diagram Export**: Para fluxos lineares longos, um Diagrama de Sequência (`sequenceDiagram`) pode ser mais legível que um Flowchart, mostrando temporalidade no eixo Y.
4. **Interactive Scrubbing**: Ferramenta Web (HTML/JS) que permite "tocar" o histórico (Previous/Next), iluminando o caminho passo-a-passo.

> **Decisão Sóbria**: Mantivemos a v0.6 simples (Heatmap) pois resolve 80% dos casos ("Onde parei?" e "Por onde passei?") sem complexidade de renderização dinâmica. É uma ferramenta de **Orientação**, não de **Perícia**.

#### 13.2. HTTP Server (Stateless)

Adaptador REST API (`internal/adapters/http`).

- **Endpoints**: `POST /navigate`, `GET /graph`.
- **SSE (Server-Sent Events)**: Endpoint `/events` notifica clientes sobre mudanças (Hot-Reload).
  - Fonte: `fsnotify` (via Loam).
  - Transporte: `text/event-stream`.

```mermaid
sequenceDiagram
    participant Dev as Developer
    participant Loam as Loam (FileWatcher)
    participant Server as HTTP Server
    participant Client as Browser (UI)

    Client->>Server: GET /events (Inscreve-se)
    Server-->>Client: 200 OK (Stream aberto)
    
    Dev->>Loam: Salva arquivo MD/JSON
    Loam->>Server: Notifica FileChangeEvent
    Server->>Client: Envia SSE: "reload"
    Client->>Client: window.location.reload()
    Client->>Server: GET /render (Busca estado atualizado)
```

#### 13.3. MCP Adapter

Expõe o Trellis como um servidor MCP (Model Context Protocol).

- **Tools Expostas**: `navigate`, `render_state`.

#### 13.4. Redis Persistence Model

To support scalable persistent sessions, the Redis adapter implements a specialized indexing strategy.

- **Storage**: Sessions are stored as JSON blobs in strict keys (`trellis:session:<id>`) with an optional TTL.
- **Indexing**: A `ZSET` (`trellis:session:index`) tracks active sessions using the expiration timestamp as the score.
- **Lazy Cleanup**: The `List()` operation performs maintenance. It removes expired entries from the index (*ZREMRANGEBYSCORE*) before returning valid sessions.

> **Trade-off**: This design keeps the adapter stateless (no background workers required), aligning with Serverless architectures. However, it means `List()` incurs a write cost. For high-throughput environments requiring read-only listing, this behavior can be disabled in favor of an external garbage collector (Future Work).

```mermaid
sequenceDiagram
    participant App
    participant Adapter
    participant Redis(ZSET)
    participant Redis(Key)

    App->>Adapter: Save(session)
    Adapter->>Redis(Key): SET session JSON (TTL)
    Adapter->>Redis(ZSET): ZADD index (Score=Now+TTL)
    
    App->>Adapter: List()
    Adapter->>Redis(ZSET): ZREMRANGE (Score < Now)
    Adapter->>Redis(ZSET): ZRANGE (All)
    Redis(ZSET)-->>Adapter: [session_ids]
    Adapter-->>App: [session_ids]
```

### 14. Segurança & Privacidade (Privacy Hooks)

O Trellis oferece suporte camadas de middleware para garantir conformidade com políticas de segurança (Encryption at Rest) e privacidade (PII Sanitization).

#### 14.1. Envelope Encryption (At Rest)

Para proteger o estado da sessão (que pode conter chaves de API e dados do usuário) em armazenamento não confiável (como disco ou REDIS compartilhado), utilizamos o **Envelope Pattern**.

O middleware criptografa todo o estado da sessão e o armazena dentro de um "Estado Envelope" opaco.

```mermaid
graph LR
    Engine[Engine State] -->|Plain JSON| Middleware[Encryption Middleware]
    Middleware -->|AES-GCM (Key A)| Cipher[Ciphertext Blob]
    Cipher -->|Wrap| Envelope[Envelope State]
    Envelope -->|Save| Store[Storage Adapter]
    
    subgraph "Envelope State"
        Ctx["__encrypted__: <base64>"]
    end
```

- **Key Rotation**: O middleware suporta rotação de chaves sem downtime. Ao carregar, ele tenta a chave ativa; se falhar, tenta chaves de fallback sequencialmente. Ao salvar, sempre re-encripta com a chave ativa mais recente.

#### 14.2. PII Sanitization (Compliance)

Um middleware separado permite a sanitização de dados sensíveis (Personally Identifiable Information) antes da persistência.

- **Deep Masking**: Percorre recursivamente o mapa de contexto e substitui valores de chaves sensíveis (ex: `password`, `ssn`, `api_key`) por `***`.
- **Imutabilidade em Memória**: A sanitização ocorre em uma **cópia profunda** (Deep Copy) do estado antes de salvar. O estado em memória usado pelo Engine permanece intacto para execução contínua.
- **Caveat**: Se o processo falhar e for reiniciado, os dados persistidos estarão mascarados (`***`), o que pode impedir a retomada se o fluxo depender desses dados. Use este middleware para Compliance de Logs ou quando a durabilidade do dado sensível não for crítica.
- **Resources**: `trellis://graph`.

### 14. Observability

Trellis fornece **Lifecycle Hooks** para instrumentação externa.

- **Hooks**: `OnNodeEnter`, `OnNodeLeave`, `OnToolReturn`, etc.
- **Log Standard**: Events use consistent keys.
  - `node_id`: ID of the node.
  - `tool_name`: Name of the tool (never empty).
  - `type`: Event type (enter, leave, tool_call).
- **Integração**: Pode ser usado com `log/slog` e `Prometheus` sem acoplar essas dependências ao Core (ex: `examples/structured-logging`).

#### 14.1 Diagrama de Eventos (Lifecycle Hooks)

O diagrama abaixo ilustra onde cada evento é emitido durante o ciclo `Navigate`:

```mermaid
sequenceDiagram
    participant Host
    participant Engine
    participant Hooks

    Note over Engine: Start Navigation (Node A)
    
    Engine->>Hooks: Emit OnNodeEnter(A)
    Engine->>Engine: Render Content
    
    alt is Tool Call
        Engine->>Hooks: Emit OnToolCall(ToolSpec)
        Engine-->>Host: ActionCallTool
        Host->>Host: Execute Tool
        Host->>Engine: Return Result
        Engine->>Hooks: Emit OnToolReturn(Result)
    end

    Engine->>Engine: Update Context (save_to)
    
    Engine->>Hooks: Emit OnNodeLeave(A)
    
    Engine->>Engine: Resolve Transition -> Node B
```
