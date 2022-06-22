variable "cloud" {
  type    = string
  default = "aws"
}

variable "location" {
  type    = string
  default = "eu_west_1"
}

resource multy_virtual_network vn {
  name       = "test-vm"
  cidr_block = "10.0.0.0/16"
  cloud      = var.cloud
  location   = var.location
}

resource multy_subnet subnet {
  name               = "test-vm"
  cidr_block         = "10.0.10.0/24"
  virtual_network_id = multy_virtual_network.vn.id
}

resource multy_virtual_machine vm {
  name            = "test-vm"
  size            = "general_micro"
  subnet_id       = multy_subnet.subnet.id
  image_reference = {
    os : "ubuntu"
    version : "20.04"
  }
  cloud    = var.cloud
  location = var.location
}