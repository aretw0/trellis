# Padrões de Documentação e Engenharia

Este documento define os padrões para colaboração, escrita e manutenção do projeto Trellis.

## 1. Política de Idiomas (Language Policy)

O projeto adota uma abordagem **bilingue e contextual** para maximizar tanto o alcance global quanto a velocidade de engenharia.

### A. Inglês (Interface Pública)

**Público-alvo:** Usuários, Contribuidores Open Source, Consumidores da Lib.

Documentos que explicam *como usar* ou *o que é* o projeto devem ser estritamente em **Inglês**.

* **Locais:**
  * `README.md` (Raiz e subdiretórios públicos como `examples/`)
  * `docs/*.md` (Guias de usuário, tutoriais)
  * **Código:** Comentários, GoDocs, nomes de variáveis/funções.
  * **CLI:** Mensagens de ajuda (`--help`), logs padrão.
  * **Testes:** `TESTING.md` (pois orienta contribuidores externos).

### B. Português (Engenharia Interna)

**Público-alvo:** Core Team, Arquitetos, Mantenedores.

Documentos que explicam *como funciona internamente*, *por que decidimos assim* e *o que faremos a seguir* devem ser em **Português**. Isso reduz a carga cognitiva durante o design complexo.

* **Locais:**
  * `docs/TECHNICAL.md` (Arquitetura profunda)
  * `docs/PLANNING.md` (Roadmap e tarefas)
  * `docs/DECISIONS.md` (Log de decisões arquiteturais - ADRs)
  * `docs/INSIGHTS.md` (Reflexões abstratas)

---

## 2. Estrutura de Documentação

Para manter a raiz limpa, movemos a documentação para `docs/` sempre que possível, referenciando-a no `README.md` principal.

* `docs/reference/`: Especificações técnicas (ex: Sintaxe de Nós).
* `docs/guides/`: Tutoriais e HOWTOs.
* `docs/architecture/`: Detalhes internos (em PT).

### Exceções (Arquivos na Raiz)

Alguns arquivos padrão da comunidade devem permanecer na raiz para descoberta automática:

* `README.md`
* `LICENSE`
* `CONTRIBUTING.md` (Se houver)
* `go.mod`, `go.sum`
* `Makefile`
