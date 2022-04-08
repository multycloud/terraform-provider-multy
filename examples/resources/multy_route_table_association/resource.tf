# TODO
resource "multy_virtual_network" "vn" {
  name       = "dev-vn"
  cidr_block = "10.0.0.0/16"
  location   = "ireland"
  cloud      = "aws"
}

resource "multy_subnet" "subnet" {
  name            = "dev-public-subnet"
  cidr_block      = "10.0.1.0/24"
  virtual_network = multy_virtual_network.vn.id
}

resource multy_route_table rt {
  name               = "dev-rt"
  virtual_network_id = multy_virtual_network.vn.id
  route {
    cidr_block  = "0.0.0.0/0"
    destination = "internet"
  }
}

resource multy_route_table_association rta {
  route_table_id = multy_route_table.rt.id
  subnet_id      = multy_subnet.subnet.id
}