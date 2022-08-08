terraform {
  required_providers {
    multy = {
      version = "0.0.1"
      source  = "hashicorp.com/dev/multy"
    }
  }
}

provider "multy" {
  gcp = {
    project = "multy-project"
  }
  aws             = {}
  azure           = {}
  api_key         = "goncalo"
  server_endpoint = "localhost:8000"
}
