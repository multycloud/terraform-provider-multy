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

resource multy_subnet subnet {
  name               = "test_subnet"
  cidr_block         = "10.0.10.0/24"
  virtual_network_id = multy_virtual_network.vn.id
}