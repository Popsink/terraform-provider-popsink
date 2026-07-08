---
page_title: "popsink_env Resource - Popsink"
subcategory: ""
description: |-
  Manages a Popsink environment.
---

# popsink_env (Resource)

Manages a Popsink environment. An environment is the foundational namespace of
the data-plane: [teams](team.md) carry an `env_id`, and all connectors and
pipelines are scoped through teams to an environment. Every environment requires
a Kafka broker retention configuration.

## Example Usage

### Plaintext broker

```hcl
resource "popsink_env" "development" {
  name = "development"

  retention_configuration = {
    bootstrap_server = "kafka.internal:9092"
  }
}
```

### SASL/SSL broker with credentials

```hcl
resource "popsink_env" "production" {
  name = "production"

  retention_configuration = {
    bootstrap_server  = "kafka.example.com:9093"
    security_protocol = "SASL_SSL"
    sasl_mechanism    = "SCRAM-SHA-256"
    sasl_username     = var.broker_username
    sasl_password     = var.broker_password
  }
}
```

## Argument Reference

- `name` (Required) - The name of the environment.
- `retention_configuration` (Required) - Kafka broker retention configuration. Fields:
  - `bootstrap_server` (Required) - Kafka bootstrap server as `host:port`.
  - `security_protocol` (Optional) - One of `PLAINTEXT`, `SASL_PLAINTEXT`, `SASL_SSL`, `SSL`. Defaults to `PLAINTEXT` server-side when omitted.
  - `sasl_mechanism` (Optional) - One of `OAUTHBEARER`, `PLAIN`, `SCRAM-SHA-256`, `SCRAM-SHA-512`, `GSSAPI`. Defaults to `PLAIN` server-side when omitted.
  - `sasl_username` (Optional) - SASL username.
  - `sasl_password` (Optional, Sensitive) - SASL password.
  - `ca_cert` (Optional, Sensitive) - CA certificate (PEM).
  - `cert` (Optional, Sensitive) - Client certificate (PEM).
  - `key` (Optional, Sensitive) - Client key (PEM).
  - `group_id` (Optional) - Consumer group ID.

## Attribute Reference

- `id` - The unique identifier of the environment.

## Credential read-back behavior

The data-plane read endpoint (`GET /envs/{id}`) returns only the non-credential
broker fields (`bootstrap_server`, `security_protocol`, `sasl_mechanism`,
`group_id`); the credential fields (`sasl_username`, `sasl_password`, `ca_cert`,
`cert`, `key`) are never returned. The provider therefore keeps the configured
credential values in state and refreshes only the non-credential fields on read.
A consequence is that out-of-band changes to credentials are not detected by
value comparison. Supply credentials through sensitive input variables.

## Import

Environments can be imported using their ID:

```bash
terraform import popsink_env.production <env-uuid>
```
