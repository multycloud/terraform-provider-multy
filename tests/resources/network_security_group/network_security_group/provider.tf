terraform {
  required_providers {
    multy = {
      version = "0.0.1"
      source  = "hashicorp.com/dev/multy"
    }
  }
}

provider "multy" {
  aws = {}
  azure = {}
  gcp = {project = "multy-project"}
  api_key         = "secret-1"
  server_endpoint = "localhost:8000"
}
