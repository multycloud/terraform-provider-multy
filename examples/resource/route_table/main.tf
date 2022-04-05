terraform {
  required_providers {
    multy = {
      version = "0.0.1"
      source  = "hashicorp.com/dev/multy"
    }
  }
}

provider "multy" {
  api_key         = "multy_local"
  server_endpoint = "localhost:8000"
}

resource "multy_virtual_network" "example_vn" {
  name       = "rta_test"
  cidr_block = "10.0.0.0/16"
  location   = "ireland"
  cloud      = "aws"
}

resource "multy_subnet" "subnet" {
  name               = "rta_test_s1"
  cidr_block         = "10.0.1.0/24"
  virtual_network_id = multy_virtual_network.example_vn.id
}

resource multy_route_table rt {
  name               = "rt_test"
  virtual_network_id = multy_virtual_network.example_vn.id
  route {
    cidr_block  = "0.0.0.0/0"
    destination = "internet"
  }
  depends_on = [multy_subnet.subnet]
}
