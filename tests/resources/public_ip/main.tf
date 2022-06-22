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
  name       = "example-vn"
  cidr_block = "10.0.0.0/16"
  location   = var.location
}
resource "multy_subnet" "subnet" {
  name               = "subnet"
  cidr_block         = "10.0.2.0/24"
  virtual_network_id = multy_virtual_network.example_vn.id
  availability_zone  = 2
}
resource "multy_network_interface" "public-nic" {
  cloud     = var.cloud
  name      = "test-public-nic"
  subnet_id = multy_subnet.subnet.id
  location  = var.location
}
resource "multy_network_interface" "private-nic" {
  cloud     = var.cloud
  name      = "test-private-nic"
  subnet_id = multy_subnet.subnet.id
  location  = var.location
}
resource "multy_public_ip" "ip" {
  cloud    = var.cloud
  name     = "test-ip"
  location = var.location
}