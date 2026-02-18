# Terraform configuration for Popsink provider
terraform {
  required_providers {
    popsink = {
      source  = "popsink/popsink"
      version = "~> 1.0"
    }
  }
}

provider "popsink" {
  # Configuration will be read from environment variables:
  # POPSINK_BASE_URL (required)
  # POPSINK_TOKEN (required)
}
