# Trellis Configuration

Este documento descreve flags de CLI, variaveis de ambiente e convencoes de runtime usadas para configurar o Trellis.

## Flags de CLI

O Trellis registra as flags no comando raiz, mas nem todas sao usadas por todos os subcomandos.

### Flags usadas pelo `run`

| Flag | Tipo | Padrao | Descricao |
| --- | --- | --- | --- |
| `--dir` | string | `.` | Diretorio do projeto Trellis. Em `trellis run`, um argumento posicional tambem define o diretorio. |
| `--headless` | bool | `false` | Executa sem prompts interativos (IO estrito). |
| `--json` | bool | `false` | Ativa modo NDJSON de entrada/saida. |
| `--debug` | bool | `false` | Ativa logs detalhados (hooks de observabilidade). |
| `--context`, `-c` | string | `""` | Contexto inicial em JSON (ex: `'{"user":"Alice"}'`). |
| `--session`, `-s` | string | `""` | ID de sessao para execucao duravel (retoma se existir). |
| `--watch`, `-w` | bool | `false` | Hot-reload. Nao pode ser usado com `--headless`. |
| `--fresh` | bool | `false` | Inicia com sessao limpa (remove dados existentes). |
| `--redis-url` | string | `""` | URL Redis para estado distribuido e locking. |
| `--tools` | string | `tools.yaml` | Caminho do registry de tools. Se `tools.yaml` existir no repo, ele e usado automaticamente. |
| `--unsafe-inline` | bool | `false` | Permite execucao inline de scripts no frontmatter. |

### Flags usadas pelo `graph`

| Flag | Tipo | Padrao | Descricao |
| --- | --- | --- | --- |
| `--dir` | string | `.` | Diretorio do projeto Trellis. Um argumento posicional tambem define o diretorio. |
| `--session` | string | `""` | Sobrepoe o grafo com historico e no atual da sessao. |

### Flags usadas pelo `validate`

| Parametro | Tipo | Padrao | Descricao |
| --- | --- | --- | --- |
| argumento posicional | string | CWD | Diretorio do projeto. `validate` nao usa `--dir`. |

### Exemplos

Rodar um fluxo com contexto inicial:

```bash
trellis run ./examples/tour --context '{"user":"Alice"}'
```

Rodar em watch com debug:

```bash
trellis run ./examples/tour --watch --debug
```

Rodar com sessao Redis:

```bash
trellis run ./examples/tour --session demo --redis-url redis://localhost:6379
```

Exportar grafo com overlay de sessao:

```bash
trellis graph ./examples/tour --session demo
```

## Persistencia de Sessao

- Se `--session` for informado, as sessoes sao armazenadas em `.trellis/sessions` por padrao.
- Com `--redis-url`, o Trellis usa Redis para estado e locks distribuidos.
- `--fresh` remove a sessao antes de iniciar.

## Sanitizacao de Input

O Trellis sanitiza a entrada do usuario impondo limite de tamanho e validacao UTF-8.

| Env Var | Padrao | Descricao |
| --- | --- | --- |
| `TRELLIS_MAX_INPUT_SIZE` | `4096` | Tamanho maximo em bytes antes de rejeitar o input. |
