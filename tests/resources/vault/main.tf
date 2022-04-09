terraform {
  required_providers {
    multy = {
      version = "0.0.1"
      source  = "hashicorp.com/dev/multy"
    }
  }
}

variable clouds {
  type    = list(string)
  default = ["aws"]
}

provider "multy" {
  aws             = {}
  api_key         = "secret-1"
  server_endpoint = "localhost:8000"
}

resource multy_vault v {
  name     = "multyvault"
  cloud    = "aws"
  location = "us_east"
}
