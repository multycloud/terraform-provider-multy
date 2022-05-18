terraform {
  required_providers {
    multy = {
      version = "0.0.1"
      source  = "hashicorp.com/dev/multy"
    }
  }
}

provider "multy" {
  azure           = {}
  aws             = {}
  api_key         = "test"
  server_endpoint = "localhost:8000"
}
