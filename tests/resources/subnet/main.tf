terraform {
  required_providers {
    multy = {
      version = "0.0.1"
      source  = "hashicorp.com/dev/multy"
    }
  }
}

provider "multy" {
  api_key         = "secret-1"
  server_endpoint = "localhost:8000"
}

resource multy_virtual_network vn {
  name       = "test_subnet"
  cidr_block = "10.0.0.0/16"
  cloud      = "aws"
  location   = "us_east"
}

resource multy_subnet subnet {
  name               = "test_subnet"
  cidr_block         = "10.0.0.0/1"
  virtual_network_id = multy_virtual_network.vn.id
}


resource multy_subnet subnet_2 {
  name               = "test_subnet"
  cidr_block         = "10.0.1.0/1"
  virtual_network_id = multy_virtual_network.vn.id
}