variable "cloud" {
  type    = string
  default = "aws"
}


resource multy_virtual_network vn {
  name       = "test"
  cidr_block = "10.0.0.0/16"
  cloud      = var.cloud
  location   = "eu_west_1"
}

resource multy_subnet subnet {
  name               = "test_subnet"
  cidr_block         = "10.0.10.0/24"
  virtual_network_id = multy_virtual_network.vn.id
  location           = "eu_west_1"
}

resource multy_virtual_machine vm {
  name         = "test_vm"
  size         = "micro"
  subnet_id    = multy_subnet.subnet.id
  public_ip_id = "123"
  ssh_key      = file("./ssh_key")
  cloud        = var.cloud
  location     = "eu_west_1"
}

resource "multy_network_security_group" nsg {
  name               = "test-nsg"
  virtual_network_id = multy_virtual_network.vn.id
  cloud              = var.cloud
  location           = "eu_west_1"
  rule {
    protocol   = "xx"
    priority   = "120"
    from_port  = "-9"
    to_port    = "9"
    cidr_block = "0.0.0.0"
    direction  = "down"
  }
}

resource "multy_network_security_group" nsg_empty {
  name               = "nsg_empty"
  virtual_network_id = multy_virtual_network.vn.id
  cloud              = var.cloud
  location           = "eu_west_1"
}
