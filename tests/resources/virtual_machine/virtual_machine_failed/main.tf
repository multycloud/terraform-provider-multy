variable "cloud" {
  type    = string
  default = "aws"
}

resource multy_virtual_network vn {
  name       = "test"
  cidr_block = "10.0.0.0/10"
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
  name               = "test_vm"
  size               = "micro"
  operating_system   = "xx"
  subnet_id          = multy_subnet.subnet.id
  public_ip_id       = "123"
  generate_public_ip = false
  cloud              = var.cloud
  location           = "eu_west_1"
}