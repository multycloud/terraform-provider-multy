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

multy "virtual_network" "example_vn" {
  name       = "example_vn"
  cidr_block = "10.0.0.0/16"
  cloud      = "aws"
  location   = "ireland"
}
multy "subnet" "subnet" {
  name              = "subnet"
  cidr_block        = "10.0.2.0/24"
  virtual_network   = example_vn
  availability_zone = 2
}
multy "network_interface" "public-nic" {
  name      = "test-public-nic"
  subnet_id = subnet
}
multy "network_interface" "private-nic" {
  name      = "test-private-nic"
  subnet_id = subnet
}
multy "public_ip" "ip" {
  name                 = "test-ip"
  network_interface_id = public-nic
  location             = "ireland"
  cloud                = "aws"
}