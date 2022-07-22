variable "cloud" {
  type    = string
  default = "gcp"
}

resource multy_virtual_network vn {
  name       = "test-subnet"
  cidr_block = "10.0.0.0/16"
  cloud      = var.cloud
  location   = "us_east_1"
}

resource multy_subnet subnet {
  name               = "test-subnet"
  cidr_block         = "10.0.1.0/24"
  virtual_network_id = multy_virtual_network.vn.id
}


resource multy_subnet subnet_2 {
  name               = "test-subnet2"
  cidr_block         = "10.0.2.0/24"
  virtual_network_id = multy_virtual_network.vn.id
}
