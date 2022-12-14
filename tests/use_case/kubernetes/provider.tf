terraform {
  required_providers {
    multy = {
      version = "0.0.1"
      source  = "hashicorp.com/dev/multy"
    }
  }
}

provider "multy" {
  api_key         = "joao"
  server_endpoint = "localhost:8000"
  aws             = {}
  azure           = {}
  gcp             = {project="multy-project"}
}
