resource "multy_virtual_network" "vn" {
  name       = "test_nsg"
  cidr_block = "10.0.0.0/16"
  cloud      = "azure"
  location   = "eu_west_1"
}

resource "multy_network_security_group" "nsg" {
  name               = "test_nsg"
  virtual_network_id = multy_virtual_network.vn.id
  cloud              = "azure"
  location           = "eu_west_1"
  rule {
    protocol   = "tcp"
    priority   = 120
    from_port  = 22
    to_port    = 22
    cidr_block = "0.0.0.0/0"
    direction  = "both"
  }
  rule {
    protocol   = "tcp"
    priority   = 130
    from_port  = 443
    to_port    = 444
    cidr_block = "0.0.0.0/0"
    direction  = "both"
  }
}
