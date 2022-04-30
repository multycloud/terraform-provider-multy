resource "multy_virtual_network" "vn" {
  name       = "dev-vn"
  cidr_block = "10.0.0.0/16"
  cloud      = "aws"
  location   = "us_east_1"
}

resource "multy_subnet" "subnet" {
  name               = "dev-subnet"
  cidr_block         = "10.0.10.0/24"
  virtual_network_id = multy_virtual_network.vn.id
}