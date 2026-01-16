# Guia: Gerenciamento de Sessão (Chaos Control)

O Trellis suporta **Durable Execution**, o que significa que o estado do seu fluxo pode sobreviver entre reinicializações do processo. Isso é gerenciado através de **Sessões Persistentes**.

Este guia explica como usar a CLI para gerenciar, inspecionar e limpar essas sessões.

## 1. O que é uma Sessão?

Uma sessão é um arquivo JSON armazenado em `.trellis/sessions/<session-id>.json`. Ele contém:

- **Current Node**: Onde o usuário parou.
- **Context**: Variáveis coletadas (`save_to`).
- **History**: Trilha de nós visitados.

## 2. Criando uma Sessão

Para criar ou retomar uma sessão, use a flag `--session` ao rodar um fluxo:

```bash
trellis run --session my-experiment ./examples/tour
```

Se a sessão `my-experiment` já existir, o Trellis a retomará do ponto exato onde parou. Se não, criará uma nova.

## 3. Listando Sessões (`ls`)

Para ver quais sessões estão ativas no seu projeto atual:

```bash
trellis session ls
```

**Saída:**

```text
Active Sessions:
- my-experiment
- dev-test-01
```

## 4. Inspecionando o Estado (`inspect`)

Para debugar variáveis ou entender por que um fluxo travou, você pode visualizar o JSON bruto do estado:

```bash
trellis session inspect my-experiment
```

**Saída:**

```json
{
  "CurrentNodeID": "ask_name",
  "Status": "active",
  "Context": {
    "count": 42
  },
  "History": ["start", "welcome"]
}
```

## 5. Limpando Sessões (`rm`)

Para remover uma sessão (reseta o estado para a próxima execução):

```bash
trellis session rm my-experiment
```

Você pode remover múltiplas sessões de uma vez:

```bash
trellis session rm s1 s2 s3
```

## Boas Práticas

- **Gitignore**: Certifique-se de adicionar `.trellis/` ao seu `.gitignore` para não commitar sessões de teste.
- **Ambientes**: Use IDs descritivos para evitar colisão (ex: `user-123-prod`).
- **Debugging**: Use `inspect` para verificar se `save_to` salvou os dados corretamente sem precisar adicionar "Print Debugging" no fluxo.

## 6. Live Coding com Hot Reload (`--watch`)

O Trellis permite combinar `--watch` e `--session` para uma experiência de desenvolvimento fluida.

```bash
trellis run --watch --session dev-session ./my-flow
```

**Benefícios:**

- **Preservação de Contexto**: Se você estiver no 10º passo de um fluxo complexo e mudar um texto, o Trellis recarrega o arquivo e te mantém no 10º passo, com todas as variáveis anteriores preservadas.
- **Guardrails de Segurança**: Se o novo código introduzir um erro de sintaxe ou remover o nó atual, o Trellis se recupera graciosamente (voltando para o início ou mantendo a versão anterior estável).
