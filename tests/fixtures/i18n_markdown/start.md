---
id: start
type: format
save_to: name
wait: true
messages:
  en:
    - text: "# Welcome to Trellis\n"
    - text: "Hello, **{{ .name }}**! You are a new user.\n"
      condition: "input == 'new'"
    - text: "Welcome back, **{{ .name }}**!\n"
      condition: "input != 'new'"
    - text: "\nWhat is your name?"
  pt:
    - text: "# Bem-vindo ao Trellis\n"
    - text: "Olá, **{{ .name }}**! Você é um novo usuário.\n"
      condition: "input == 'new'"
    - text: "Bem-vindo de volta, **{{ .name }}**!\n"
      condition: "input != 'new'"
    - text: "\nQual o seu nome?"
transitions:
  - to: markdown_showcase
---

# Fallback Welcome
What is your name?
