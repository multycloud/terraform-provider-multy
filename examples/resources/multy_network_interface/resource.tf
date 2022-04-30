# TODO
resource "multy_virtual_network" "example_vn" {
  name       = "dev-nic"
  cidr_block = "10.0.0.0/16"
  cloud      = "aws"
  location   = "eu_west_1"
}

resource "multy_subnet" "subnet" {
  name            = "dev-nic"
  cidr_block      = "10.0.2.0/24"
  virtual_network = multy_virtual_network.example_vn.id
}

resource "multy_network_interface" "nic" {
  name      = "dev-nic"
  subnet_id = multy_subnet.subnet.id
}