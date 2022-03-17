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
  location = "ireland"
}

resource multy_virtual_network vn {
  name       = "test"
  cidr_block = "10.0.0.0/16"
}

output vn {
  value = multy_virtual_network.vn
}