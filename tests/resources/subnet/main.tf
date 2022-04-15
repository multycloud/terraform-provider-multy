variable "cloud" {
  type    = string
  default = "aws"
}

resource multy_virtual_network vn {
  name       = "test_subnet"
  cidr_block = "10.0.0.0/16"
  cloud      = var.cloud
  location   = "us_east"
}

resource multy_subnet subnet {
  name               = "test_subnet"
  cidr_block         = "10.0.1.0/24"
  virtual_network_id = multy_virtual_network.vn.id
}


resource multy_subnet subnet_2 {
  name               = "test_subnet"
  cidr_block         = "10.0.2.0/24"
  virtual_network_id = multy_virtual_network.vn.id
}
