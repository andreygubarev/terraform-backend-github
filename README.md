# Terraform HTTP Backend with GitHub

Terraform HTTP backend that uses Github as a storage backend.

# Usage

```yaml
version: '3'

services:
  terraform-backend-github:
    image: ghcr.io/andreygubarev/terraform-backend-github:0.1.0
    environment:
      - GITHUB_TOKEN=${GITHUB_TOKEN}
    ports:
      - 8080:8080
    restart: unless-stopped
```

# Motivation

# Reference

- https://github.com/andreygubarev/terraform-backend-github
- https://developer.hashicorp.com/terraform/language/settings/backends/http
- https://github.com/plumber-cd/terraform-backend-git/
