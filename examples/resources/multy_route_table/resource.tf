# TODO
resource "multy_virtual_network" "vn" {
  name       = "dev-vn"
  cidr_block = "10.0.0.0/16"
  location   = "ireland"
  cloud      = "aws"
}

resource "multy_route_table" "rt" {
  name               = "dev-rt"
  virtual_network_id = multy_virtual_network.vn.id
  route {
    cidr_block  = "0.0.0.0/0"
    destination = "internet"
  }
}
