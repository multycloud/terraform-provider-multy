terraform {
  required_providers {
    multy = {
      version = "0.0.1"
      source  = "hashicorp.com/dev/multy"
    }
  }
}

provider "multy" {
  aws             = {}
  api_key         = "aws-123-1"
  server_endpoint = "localhost:8000"
}