# Reload Demo

Este exemplo demonstra as capacidades de **Stateful Hot Reload** do Trellis.

## Como Testar

1. **Inicie o Hot Reload**:

    ```bash
    go run ./cmd/trellis run --watch --session demo-reload ./examples/reload-demo
    ```

2. **Siga o Fluxo**:
    Responda o seu nome no primeiro passo.

3. **Mude o Arquivo**:
    Abra `examples/reload-demo/step2.md` e mude o texto.
    Observe que o Trellis recarrega **sem perguntar seu nome de novo**.

4. **Teste o Guardrail (Missing Node)**:
    Delete o arquivo `examples/reload-demo/step2.md`.
    O Trellis deve avisar que o nó sumiu e voltar para o `start`.

5. **Teste o Guardrail (Syntax Error)**:
    Introduza um erro de YAML no `start.md`.
    O Trellis manterá a última versão válida rodando enquanto você não consertar.

6. **Teste o Guardrail (Type Change)**:
    Mude o `step2.md` para `type: tool`.
    O estado será resetado para ativo se ele estiver esperando uma resposta.
