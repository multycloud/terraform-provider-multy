# TODO
resource multy_virtual_network vn {
  name       = "dev-nsg"
  cidr_block = "10.0.0.0/16"
  cloud      = "aws"
  location   = "ireland"
}

resource "multy_network_security_group" nsg {
  name               = "dev-nsg"
  virtual_network_id = multy_virtual_network.vn.id
  cloud              = "aws"
  location           = "ireland"

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
