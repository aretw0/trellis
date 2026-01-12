# Technical: Trellis Architecture

## 1. Arquitetura Hexagonal (Ports & Adapters)

O *Core* da Trellis não conhece banco de dados, não conhece HTTP e não conhece CLI. Ele define **Portas** (Interfaces) que o mundo externo deve satisfazer.
Essa arquitetura desacoplada torna o Trellis leve o suficiente para ser embutido em CLIs simples ou usado como biblioteca "low-level" dentro de frameworks de Agentes de IA maiores.

### 1.1. Driver Ports (Entrada)

A API primária para interagir com o engine.

- `Engine.Render(state)`: Retorna a view (ações) para o estado atual e se é terminal.
- `Engine.Navigate(state, input)`: Computa o próximo estado dado um input.
- `Engine.Inspect()`: Retorna o grafo completo para visualização.

### 1.2. Driven Ports (Saída)

As interfaces que o engine usa para buscar dados.

- `GraphLoader.GetNode(id)`: Abstração para carregar nós. O **Loam** implementa isso via adapter.
- `GraphLoader.ListNodes()`: Descoberta de nós para introspecção.

### 1.3. Diagrama de Arquitetura

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
```

### 1.4. Fluxo de Execução

```mermaid
sequenceDiagram
    participant Host
    participant Engine
    participant Loader
    
    Host->>Engine: Start()
    Engine->>Loader: GetNode("start")
    Loader-->>Engine: Node
    Engine-->>Host: State(Start)
    
    loop Game Loop
        Host->>Engine: Render(State)
        Engine-->>Host: Actions (View)
        Host->>Host: User Input / Logic
        Host->>Engine: Navigate(State, Input)
        Engine->>Engine: Evaluate Transitions
        Engine->>Loader: GetNode(NextID)
        Loader-->>Engine: Node
        Engine-->>Host: NewState
    end
```

## 2. Estrutura de Diretórios

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
│   └── runner/        # Loop de Execução e Handlers
└── go.mod
```

## 3. Integridade e Persistência

### 3.1. O Papel do Loam

O **Loam** atua como bibliotecário e camada de persistência.

- **Responsabilidade**: Garantir integridade e fornecer documentos normalizados (`DocumentModel`).
- **Trellis Adapter (`LoamLoader`)**: Converte `DocumentModel` para JSON/Structs que o Compiler entende.
- **Constraints**: Em modo de desenvolvimento, o Loam facilita o hot-reload e a leitura segura de arquivos.

### 1.5. Model Context Protocol (MCP) Adapter

Introduzido na v0.3.3 (Experimental), o adaptador MCP (`internal/adapters/mcp`) expõe o Trellis como um Servidor MCP.

- **Tools**:
  - `render_state`: Mapeia para `Engine.Render`.
  - `navigate`: Mapeia para `Engine.Navigate`.
  - `get_graph`: Mapeia para `Engine.Inspect`.
- **Resources**:
  - `trellis://graph`: Retorna o grafo via `Engine.Inspect`.
- **Transports**: Suporta `Stdio` (para agentes locais) e `SSE` (para remoto/debug).

### 3.2. Global Strict Serialization

O Trellis adota uma postura de "Strict Types" para garantir a determinística da máquina de estados.

#### O Problema do `float64`

Por padrão, decodificadores JSON em Go tratam números arbitrários como `float64`. Isso é catastrófico para IDs numéricos grandes ou timestamps.

#### A Solução

O Trellis força o modo estrito em **todos** os adaptadores. Isso garante que números sejam decodificados como `json.Number` ou `int64`, e que exista consistência entre JSON e YAML.

### 3.3. Fluxo de Dados e Serialização

O Trellis utiliza múltiplas camadas de serialização, o que explica a presença de diferentes tags (`json`, `yaml`, `mapstructure`) nas structs do domínio.

1. **Entrada (Source)**: Arquivos `.md` (YAML Frontmatter) ou `.json`.
2. **Leitura (Loam)**: A biblioteca Loam usa `mapstructure` para decodificar YAML/JSON genérico em structs de DTO (`internal/dto`).
3. **Adaptação (Loader)**: O `LoamLoader` converte os DTOs em um novo JSON limpo, estritamente tipado para o domínio.
4. **Compilação (Engine)**: O `Compiler` lê esse JSON interno e popula as structs de Domínio (`pkg/domain`).

**Por que a mistura de tags?**
Para evitar a duplicação excessiva de structs, alguns tipos do Domínio (como `ToolCall`) são reutilizados nos DTOs.

- `json`: Usado pelo **Compiler** (interno) e pela API REST/MCP (externo).
- `mapstructure`: Usado pelo **Loam** para ler arquivos de disco.
- `yaml`: Usado apenas para documentação ou ferramentas de exportação futuras (não utilizado no load crítico).

## 4. Escalabilidade: Sub-Grafos e Namespaces (v0.4+)

Para escalar fluxos complexos, o Trellis suporta **Sub-Grafos** via organização de diretórios.

### 4.1. Semântica `jump_to` vs `to`

Tecnicamente, o Trellis Engine vê apenas IDs (`to_node_id`). O conceito de `jump_to` é **açúcar sintático** do adaptador Loam para clareza arquitetural.

- **`to`**: Indica uma transição local, dentro do mesmo contexto lógico ou arquivo.
- **`jump_to`**: Indica uma transição para um **Sub-Grafo** ou Módulo externo. É uma sinalização explícita de mudança de contexto.

```yaml
transitions:
  - text: "Checkout"
    jump_to: modules/checkout/start # Semântica: Mudança de Contexto
```

### 4.2. IDs Implícitos e Normalização

1. **Implicit IDs**: Arquivos em subdiretórios herdam o caminho como ID (ex: `modules/checkout/start`).
2. **Normalization**: O Adapter normaliza todos os IDs para usar `/` (forward slash), garantindo que fluxos criados no Windows rodem no Linux sem alterações.

## 5. Runner & IO Architecture

The `Runner` serves as the bridge between the Core Engine and the outside world. It manages the execution loop, handles middleware (like confirmation), and delegates IO to an `IOHandler`.

### Stateless & Async IO

Trellis supports two primary modes of operation:

1. **Text Mode** (`TextHandler`): For interactive TUI/CLI usage. Blocks on user input.
2. **JSON Mode** (`JSONHandler`): For headless automation and API integration.

**Key constraint for JSON Mode:**

- **Strict JSON Lines (JSONL)**: All inputs to the `JSONHandler`, including tool results, must be single-line JSON strings.
- **Async/Non-Blocking**: The handler reads from Stdin in a background goroutine. This allows the Engine to cancel wait operations (e.g. timeout or interrupt) without hanging on OS-level read syscalls.

## 6. Princípios de Design (Constraints)

Para evitar a "Complexidade Oculta", seguimos estas restrições:

### 5.1. Logic-Data Decoupling

A lógica complexa **nunca** deve residir no grafo (Markdown).

- **Errado**: `condition: user.age > 18 && user.status == 'active'` (Exige parser complexo).
- **Correto**: `condition: is_adult_active` (O Host resolve e retorna bool).

> Veja [Interactive Inputs](../docs/guides/interactive_inputs.md) para detalhes sobre como o Host gerencia inputs.

### 6.2. Strict Mode Compiler

O compilador deve ser implacável.

- Variáveis não declaradas resultam em erro de compilação.
- O objetivo é **Confiança Total**: Se compilou, não existem "Dead Ends" lógicos causados por typos.

## 7. Stateless Server Mode (v0.3.3+)

Introduzido na versão 0.3.3, o Trellis pode operar como um servidor HTTP stateless.

- **Contract-First**: A API é definida em `api/openapi.yaml`.
- **Implementação**: `internal/adapters/http` usa `chi` como roteador.
- **Endpoints de Dados**:
  - `GET /graph`: Retorna o grafo completo.
  - `POST /navigate`: Recebe `State` + `Input`, retorna `NewState`.
- **Endpoints de Gerenciamento**:
  - `GET /health`: Liveness probe.
  - `GET /info`: Metadados e versões.

> Para um guia prático, veja [Running HTTP Server](../docs/guides/running_http_server.md).

## 8. Real-Time & Events (SSE)

Para suportar experiências dinâmicas (como **Hot-Reload** no navegador), o Trellis utiliza **Server-Sent Events (SSE)**.

### 7.1. Por que SSE e não WebSockets?

- **Simplicidade**: SSE usa HTTP padrão (`Content-Type: text/event-stream`). Não requer upgrade de protocolo ou handshake complexo.
- **Unidirecional**: O Trellis é o "State of Truth". O cliente apenas reage a mudanças (ex: arquivo salvo no disco -> notifica cliente -> cliente recarrega). Para enviar dados (inputs), o cliente continua usando `POST /navigate` (HTTP padrão).
- **Reconexão Nativa**: O objeto `EventSource` do navegador gerencia reconexões automaticamente.

### 7.2. Fluxo de Hot-Reload

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

### 7.3. Event Source & Caveats

O Trellis desacopla a fonte de eventos do transporte SSE:

1. **Fonte (Source)**: O `LoamLoader` utiliza `fsnotify` (via biblioteca Loam) para monitorar o sistema de arquivos.
    - *Nota*: Atualmente, o evento emitido é genérico (`"reload"`), sinalizando que *alguma coisa* mudou.
    - *Futuro*: A interface `Watch` retorna `chan string`, permitindo eventos granulares como `update:docs/start.md` ou `delete:modules/auth.json`.

2. **Transporte**: O handler SSE (`Server.SubscribeEvents`) apenas repassa as strings recebidas pelo canal para o cliente HTTP.

> **Limitações Atuais (v0.3.3)**:
>
> - Não há distinção de qual arquivo mudou no payload do evento SSE (apenas `data: reload`).

## 9. Protocolo de Efeitos Colaterais (Side-Effect Protocol)

Introduzido na v0.4.0, o protocolo de side-effects permite que o Trellis solicite a execução de código externo (ferramentas) de forma determinística e segura.

### 8.1. Filosofia: "Syscalls" para a IA

O Trellis trata chamadas de ferramenta como "Chamadas de Sistema" (Syscalls). O Engine não executa a ferramenta; ele **pausa** e solicita ao Host que a execute.

1. **Intenção (Intent)**: O Engine renderiza um nó do tipo `tool` e emite uma ação `CALL_TOOL`.
2. **Suspensão (Yield)**: O Engine entra em estado `WaitingForTool`, aguardando o resultado.
3. **Dispatch**: O Host (CLI, Servidor HTTP, MCP) recebe a solicitação e executa a lógica (ex: chamar API, rodar script).
4. **Resumo (Resume)**: O Host chama `Navigate` passando o `ToolResult`. O Engine retoma a execução verificando transições baseadas nesse resultado.

### 8.2. Ciclo de Vida da Chamada de Ferramenta

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

### 8.3. Universal Dispatcher

Graças a este desacoplamento, a mesma definição de grafo pode usar ferramentas implementadas de formas diferentes dependendo do adaptador:

- **CLI Runner**: Executa scripts locais (`.sh`, `.py`) ou funções Go embutidas.
- **MCP Server**: Repassa a chamada para um cliente MCP (ex: Claude Desktop, IDE).
- **HTTP Server**: Webhooks que notificam serviços externos (ex: n8n, Zapier).

### 8.4. Limitações Conhecidas

1. **Interpolação de Strings (Legado)**:
   - Até v0.4.0, era utilizado `strings.ReplaceAll` (`{{ key }}`).
   - **v0.4.1+**: Suportado `Interpolator` Interface (Default: Go Templates `{{ .key }}`).
   - **Nota**: A compatibilidade com sintaxe antiga é mantida via `LegacyInterpolator` opcional.

2. **Bloqueio de I/O (JSON Adapter)**:
   - Em modo headless via Stdin/Stdout, o Engine pode bloquear se o Host não enviar a resposta da ferramenta imediatamente.
   - O Runner utiliza pipes padrão que podem bloquear se não consumidos corretamente pelo Host.

3. **Semântica de Texto e Bloqueio (UX)**:
   - Atualmente, o `TextHandler` (CLI) assume que qualquer nó `type: text` com renderização exige pausa para leitura (espera `Enter`).
   - Isso impede "Pass-through Nodes" (ex: Templates) que apenas mostram dados e avançam.
   - **Plano (v0.5)**: Tornar `text` não-bloqueante por padrão e introduzir `type: prompt` para pausas explícitas.

### 8.5. Segurança e Policies (Interceptor)

Para mitigar riscos de execução arbitrária, introduzimos o padrão **Interceptor** no Runner:

```go
type ToolInterceptor func(ctx, call) (allowed bool, result ToolResult, err error)
```

- **ConfirmationMiddleware**: Padrão para modo interativo. Intercepta a chamada e solicita confirmação explícita (`[y/N]`) ao usuário antes de permitir a execução.
- **AutoApproveMiddleware**: Padrão para modo Headless/Automação.

### 8.6. Metadata-Driven Safety (v0.4.1+)

Além da confirmação padrão, o Trellis permite que o autor do fluxo personalize a mensagem de segurança via metadados.

```yaml
type: tool
tool_call:
  name: delete_database
metadata:
  confirm_msg: "⚠️  DANGER: This will destroy production data. Are you sure?"
```

O `ConfirmationMiddleware` detecta o campo `confirm_msg` e o utiliza no prompt, permitindo alertas contextuais ricos.

### 8.7. Protocolo de Mensagens de Sistema (System Messages)

Para permitir que o sistema se comunique com o usuário fora do fluxo principal (sem ser conteúdo de nó), a v0.4.1 introduziu `ActionSystemMessage` (`SYSTEM_MESSAGE`).

- **Finalidade**: Logs, feedback de status ("Salvando...", "Executando ferramenta..."), avisos de erro não-fatais.
- **Semântica**: Notificação Unidirecional (Fire-and-forget). O cliente deve exibir a mensagem mas não precisa responder.
- **Formato**:

  ```json
  [
    {
      "Type": "SYSTEM_MESSAGE",
      "Payload": "Tool 'calc' returned 42"
    }
  ]
  ```

- **Diferença para RenderContent**: `RenderContent` é *Conteúdo do Nó* (parte da narrativa). `SystemMessage` é *Metadado da Infraestrutura*.

## 10. Variable Interpolation (v0.4.1+)

A partir da v0.4.1, o Trellis adota uma arquitetura plugável para interpolação de variáveis.

### 9.1. Interpolator Interface

O motor define a interface `Interpolator` em `pkg/runtime`:

```go
type Interpolator func(ctx context.Context, templateStr string, data any) (string, error)
```

Isso permite que consumidores da biblioteca (Hosts) injetem sua própria lógica de template (ex: Mustache, Jinja2, Lua) se desejarem.

### 9.2. Default Strategy: Go Templates

A implementação padrão (`DefaultInterpolator`) utiliza a biblioteca nativa `text/template` do Go.

- **Sintaxe**: Standard Go Templates (`{{ .UserName }}`, `{{ if .IsVIP }}...{{ end }}`).
- **Robustez**: Suporta acesso a campos aninhados, condicionais, loops e pipes.
- **Segurança**: Executa em contexto isolado, mas requer cuidado ao renderizar HTML (use `html/template` no Host se necessário, o Trellis foca em Texto/Dados).

### 9.3. Legacy Strategy

Para facilitar a migração, o Trellis fornece `LegacyInterpolator`, que mantém o comportamento antigo de `strings.ReplaceAll` com a sintaxe `{{ Key }}` (sem ponto).

Para usar, basta configurar o Engine:

```go
engine, err := trellis.New(dir, trellis.WithInterpolator(runtime.LegacyInterpolator))
```
