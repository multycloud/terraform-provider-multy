terraform {
  required_providers {
    multy = {
      source = "multycloud/multy"
    }
  }
}

provider "multy" {
  api_key         = "XXX-YYY-ZZZ"
  aws             = {}
  azure           = {}
  server_endpoint = "localhost:8000"
}