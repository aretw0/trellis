---
to: end
---

# 🔐 Secure Storage Demo

I have injected sensitive data into the context:

- API Key: {{ .api_key }}
- Password: {{ .password }}

This data is available in memory for me to use.

But when I am saved to disk, I will be **PII Masked** AND **Encrypted**.

Goodbye!
