terraform {
  required_providers {
    multy = {
      version = "0.0.1"
      source  = "hashicorp.com/dev/multy"
    }
  }
}

provider "multy" {
  api_key = "multy_local"
}

resource "multy_virtual_network" "example_vn" {
  name       = "nic_test"
  cidr_block = "10.0.0.0/16"
  cloud      = "aws"
  location   = "ireland"
}
resource "multy_subnet" "subnet" {
  name            = "nic_test"
  cidr_block      = "10.0.2.0/24"
  virtual_network = multy_virtual_network.example_vn.id
}
resource "multy_network_interface" "nic" {
  name      = "nic_test"
  subnet_id = multy_subnet.subnet.id
}