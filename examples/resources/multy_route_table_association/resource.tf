resource "multy_virtual_network" "example_vn" {
  name       = "rta_test"
  cidr_block = "10.0.0.0/16"
  location   = "eu_west_1"
  cloud      = "aws"
}

resource "multy_subnet" "subnet1" {
  name               = "rta_test_s1"
  cidr_block         = "10.0.1.0/24"
  virtual_network_id = multy_virtual_network.example_vn.id
}

resource "multy_route_table" "rt" {
  name               = "rta_test"
  virtual_network_id = multy_virtual_network.example_vn.id
  route {
    cidr_block  = "0.0.0.0/0"
    destination = "internet"
  }
}

resource "multy_route_table_association" "subnet1" {
  route_table_id = multy_route_table.rt.id
  subnet_id      = multy_subnet.subnet1.id
}
