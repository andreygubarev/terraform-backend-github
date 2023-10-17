# Terraform HTTP Backend with GitHub

Terraform HTTP backend that uses Github as a storage backend.

# Usage

Run the container using the following docker-compose snippet:

```yaml
version: '3'

services:
  terraform-backend-github:
    image: ghcr.io/andreygubarev/terraform-backend-github:0.1.1
    environment:
      - GITHUB_TOKEN=${GITHUB_TOKEN}
    ports:
      - 8080:8080
    restart: unless-stopped
```

Configure the Terraform HTTP backend in your Terraform configuration:

```hcl
terraform {
  backend "http" {
    address = "http://localhost:8080/andreygubarev/terraform-backend/state/terraform.tfstate"
    lock_address = "http://localhost:8080/andreygubarev/terraform-backend/state/terraform.tfstate.lock"
    unlock_address = "http://localhost:8080/andreygubarev/terraform-backend/state/terraform.tfstate.lock"
  }
}
```

# Motivation

# Reference

- https://github.com/andreygubarev/terraform-backend-github
- https://developer.hashicorp.com/terraform/language/settings/backends/http
- https://github.com/plumber-cd/terraform-backend-git/
