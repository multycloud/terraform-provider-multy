terraform {
  required_providers {
    multy = {
      version = "1.0.0"
      source  = "hashicorp.com/dev/multy"
    }
  }
}

provider "multy" {
  api_key = "123"
}