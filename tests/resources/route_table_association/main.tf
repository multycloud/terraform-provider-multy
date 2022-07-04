variable "cloud" {
  type    = string
  default = "aws"
}

resource "multy_virtual_network" "example_vn" {
  name       = "rta-test"
  cidr_block = "10.0.0.0/16"
  location   = "eu_west_1"
  cloud      = var.cloud
}

resource "multy_subnet" "subnet1" {
  name               = "rta-test-s1"
  cidr_block         = "10.0.1.0/24"
  virtual_network_id = multy_virtual_network.example_vn.id
}

resource "multy_subnet" "subnet2" {
  name               = "rta-test-s2"
  cidr_block         = "10.0.2.0/24"
  virtual_network_id = multy_virtual_network.example_vn.id
}

resource multy_route_table rt {
  name               = "rta-test"
  virtual_network_id = multy_virtual_network.example_vn.id
  route {
    cidr_block  = "0.0.0.0/0"
    destination = "internet"
  }
}

resource multy_route_table_association subnet1 {
  route_table_id = multy_route_table.rt.id
  subnet_id      = multy_subnet.subnet1.id
}