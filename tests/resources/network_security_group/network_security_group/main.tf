variable "location" {
  type    = string
  default = "eu_west_1"
}

variable "cloud" {
  type    = string
  default = "aws"
}

resource multy_virtual_network vn {
  name       = "test-nsg"
  cidr_block = "10.0.0.0/16"
  cloud      = var.cloud
  location   = var.location
}

resource "multy_network_security_group" nsg {
  name               = "test-nsg"
  virtual_network_id = multy_virtual_network.vn.id
  cloud              = var.cloud
  location           = var.location
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
    priority   = 131
    from_port  = 443
    to_port    = 444
    cidr_block = "0.0.0.0/0"
    direction  = "both"
  }
}
