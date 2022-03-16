terraform {
  required_providers {
    multy = {
      version = "1.0.0"
      source  = "hashicorp.com/dev/multy"
    }
  }
}

provider "multy"{
  api_key = "123"
}

resource multy_virtual_network vn {
  name = "test"
  cidr_block = "10.0.0.0/20"
}