# Insights: Structure, Web, and Hypermedia

Captured on: 2026-02-18
Context: v0.7.12 Planning

## 1. The Entrypoint Convention: `start` vs `main` vs `index`

We analyzed three conventions for naming the entry node of a flow:

* **`start` (Action/Flow-centric):**
  * *Semantics:* "Begin the process."
  * *Best for:* Wizards, installation scripts, interactive flows.
  * *Current Status:* The default in Trellis.

* **`main` (Program-centric):**
  * *Semantics:* "The entry point of the executable."
  * *Best for:* Building CLIs, logic-heavy applications (akin to C/Go/Rust).
  * *Status:* Supported as fallback in v0.7.12.

* **`index` (Location/Web-centric):**
  * *Semantics:* "The default document of this directory."
  * *Best for:* Content-heavy sites, Wikis, documentation viewers.
  * *Status:* Not yet supported, but relevant if Trellis moves towards web serving.

**Conclusion:**
Trellis is a chameleon. It can be a CLI (Program), a Flow (Action), or potentially a Site (Location). Supporting a fallback chain (`start` -> `main` -> `DirectoryName` -> `index`) allows the user to choose the semantic that matches their domain.

## 2. Trellis vs. Astro (The Web Server Potential)

The comparison to [Astro](https://astro.build) is pertinent:

* **File-System Routing:** Astro maps `src/pages/about.astro` to `/about`. Trellis maps `repo/about.md` to the `about` node.
* **Component vs. State:**
  * Astro renders components to HTML.
  * Trellis currently renders "Nodes" to JSON/TUI.
* **The Convergence:**
    If Trellis were to act as a web server, it would effectively be a **"Stateful Web Server"**.
  * Unlike a standard web server (stateless request/response), a Trellis session maintains continuity.
  * *URL Structure:* `GET /session/:id/node/step-2`.

**Trellis as a "Hypermedia Engine":**
Trellis is closer to the [HTMX](https://htmx.org) philosophy or true REST (HATEOAS) than React/Vue. The state *is* the resource.

## 3. Markdown Links as Transitions (The "Wikilink" Graph)

**The Insight:**
Currently, we define structure in YAML/JSON:

```yaml
transitions:
  - to: next_step
    label: "Go Next"
```

But Markdown acts as a natural graph definition language:

```markdown
# Welcome
Would you like to [Log In](./login.md) or [Register](./register.md)?
```

**Proposal: "Implicit Transitions"**

* **Parser Logic:** Scan the Rendered Markdown for standard links (`[Label](Target)`).
* **Transformation:** Convert these links into runtime `ActionRequests`.
* **Benefit:** Zero-config navigation. The text *is* the interface.

**Challenges:**

* **Ambiguity:** Does clicking a link just "navigate" or does it submit data?
* **Parameters:** How to pass flags/inputs via a simple link? (Maybe query params: `[Buy](./buy.md?qty=1)`)
* **Validation:** Ensuring the target file actually exists (ID collision check helps here).

**Roadmap Implication:**
This moves Trellis from a "Configured State Machine" to a "Hypertext State Machine". It drastically lowers the barrier to entry—users just write Markdown.
