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
  clouds   = ["aws"]
  location = "us_east"
}

resource multy_virtual_network vn {
  name       = "test"
  cidr_block = "10.0.0.0/16"
  location   = "ireland"
  clouds     = []
}

output vn {
  value = multy_virtual_network.vn
}