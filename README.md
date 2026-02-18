# Terraform Provider for Popsink

[![Build Status](https://github.com/Popsink/popsink-terraform-provider/workflows/test/badge.svg)](https://github.com/Popsink/popsink-terraform-provider/actions)
[![Go Version](https://img.shields.io/github/go-mod/go-version/Popsink/popsink-terraform-provider)](https://golang.org/)

The Popsink Terraform Provider allows you to manage Popsink teams and connectors (source and target) using Infrastructure as Code.

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.25

## Using the Provider

### Installation

Add the provider to your Terraform configuration:

```hcl
terraform {
  required_providers {
    popsink = {
      source  = "popsink/popsink"
      version = "~> 1.0"
    }
  }
}

provider "popsink" {
  # Configuration is read from environment variables:
  # POPSINK_BASE_URL, POPSINK_TOKEN, POPSINK_INSECURE
}
```

### Quick Example

```hcl
resource "popsink_team" "example" {
  name        = "data-engineering"
  description = "Data engineering team"
}

resource "popsink_connector" "kafka_source" {
  name           = "kafka-source"
  connector_type = "KAFKA_SOURCE"
  team_id        = popsink_team.example.id

  json_configuration = jsonencode({
    bootstrap_servers = "kafka.example.com:9092"
    security_protocol = "SASL_SSL"
    sasl_mechanism    = "SCRAM-SHA-256"
    sasl_username     = "admin"
    sasl_password     = var.kafka_password
    topic             = "orders"
  })
}
```

## Documentation

- **Resources**: See [docs/resources/](./docs/resources/) for detailed documentation on each resource
  - [popsink_team](./docs/resources/team.md) - Teams
  - [popsink_connector](./docs/resources/connector.md) - Source and target connectors (7 types supported)

- **Examples**: See [examples/](./examples/) for complete working configurations (teams + connectors)

## Development

```bash
make build     # Build the provider (runs lint first)
make test      # Run unit tests
make testacc   # Run acceptance tests
make install   # Install locally for testing
make fmt       # Format code
make lint      # Run linter
make clean     # Clean build artifacts
```

### Testing locally with `terraform apply`

1. Build and install the provider:

```bash
make install
```

2. Set environment variables and run:

```bash
export POPSINK_BASE_URL="https://data-plane.ppsk.localhost/api"
export POPSINK_TOKEN="your-token"
export POPSINK_INSECURE="true"  # for self-signed certificates

cd examples/
terraform init
terraform plan
terraform apply
```
