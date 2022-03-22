terraform {
  required_providers {
    multy = {
      version = "0.0.1"
      source  = "hashicorp.com/dev/multy"
    }
  }
}

provider "multy" {
  api_key  = "123"
  location = "us_east"
}

resource multy_virtual_network vn {
  name       = "test2"
  cidr_block = "10.0.0.0/16"
  location   = "ireland"
  cloud      = "aws"
}

output vn {
  value = multy_virtual_network.vn
}