terraform {
  required_providers {
    multy = {
      version = "0.0.1"
      source  = "hashicorp.com/dev/multy"
    }
  }
}

provider "multy" {
  api_key  = "multy_local"
  location = "us_east"
}

resource multy_virtual_network vn {
  name       = "vn_test4"
  cidr_block = ""
  cloud      = "aws"
}

output vn {
  value = multy_virtual_network.vn
}
