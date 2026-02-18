---
page_title: "Popsink Provider"
subcategory: ""
description: |-
  The Popsink provider is used to interact with Popsink API resources to manage teams and connectors.
---

# Popsink Provider

The Popsink provider allows you to manage Popsink teams and connectors using Terraform.

Use the navigation to the left to read about the available resources.

## Schema

### Required

- `base_url` (String) The base URL for the Popsink API. Can also be set via the `POPSINK_BASE_URL` environment variable.
- `token` (String, Sensitive) The API token for authenticating with the Popsink API. Can also be set via the `POPSINK_TOKEN` environment variable.

### Optional

- `insecure` (Boolean) Skip TLS certificate verification. Can also be set via the `POPSINK_INSECURE` environment variable. Defaults to `false`.

## Authentication

The recommended way to configure the provider is using environment variables:

```bash
export POPSINK_BASE_URL="https://data-plane.example.com/api"
export POPSINK_TOKEN="your-api-token"
export POPSINK_INSECURE="true"  # optional, for self-signed certificates
```

Then in your Terraform configuration:

```hcl
provider "popsink" {}
```

You can also configure the provider directly:

```hcl
provider "popsink" {
  base_url = "https://data-plane.example.com/api"
  token    = var.popsink_token
  insecure = true
}
```

## Resources

- [popsink_team](resources/team.md) - Manage Popsink teams
- [popsink_connector](resources/connector.md) - Manage Popsink connectors (source/target)
