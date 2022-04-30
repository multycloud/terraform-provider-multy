variable "location" {
  type    = string
  default = "eu_west_1"
}

variable "cloud" {
  type    = string
  default = "aws"
}

resource "multy_virtual_network" "example_vn" {
  cloud      = var.cloud
  name       = "nic_test"
  cidr_block = "10.0.0.0/16"
  location   = var.location
}
resource "multy_subnet" "subnet" {
  name               = "nic_test"
  cidr_block         = "10.0.2.0/24"
  virtual_network_id = multy_virtual_network.example_vn.id
}
resource "multy_network_interface" "private-nic" {
  cloud     = var.cloud
  name      = "test-private-nic"
  subnet_id = multy_subnet.subnet.id
  location  = var.location
}