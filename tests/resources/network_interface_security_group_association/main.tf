variable "cloud" {
  type    = string
  default = "aws"
}

resource "multy_virtual_network" "example_vn" {
  name       = "rta_test"
  cidr_block = "10.0.0.0/16"
  location   = "eu_west_1"
  cloud      = var.cloud
}

resource "multy_subnet" "subnet1" {
  name               = "rta_test_s1"
  cidr_block         = "10.0.1.0/24"
  virtual_network_id = multy_virtual_network.example_vn.id
}

resource "multy_public_ip" "pip" {
  name     = "test-pip"
  location = "eu_west_1"
  cloud    = var.cloud
}

resource "multy_network_interface" "nic" {
  name         = "test_nic"
  subnet_id    = multy_subnet.subnet1.id
  public_ip_id = multy_public_ip.pip.id
  location     = "eu_west_1"
  cloud        = var.cloud
}

resource "multy_network_security_group" nsg {
  name               = "test_nsg"
  virtual_network_id = multy_virtual_network.example_vn.id
  cloud              = var.cloud
  location           = "eu_west_1"
  rule {
    protocol   = "tcp"
    priority   = 120
    from_port  = 22
    to_port    = 22
    cidr_block = "0.0.0.0/0"
    direction  = "both"
  }
}

resource "multy_network_interface_security_group_association" "nic-association" {
  network_interface_id = multy_network_interface.nic.id
  security_group_id    = multy_network_security_group.nsg.id
}
