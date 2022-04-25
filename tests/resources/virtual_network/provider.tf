terraform {
  required_providers {
    multy = {
      #      version = "0.0.1"
      #      source  = "hashicorp.com/dev/multy"
      source = "multycloud/multy"
    }
  }
}

provider "multy" {
  aws = {}
  #  server_endpoint = "localhost:8000"
}