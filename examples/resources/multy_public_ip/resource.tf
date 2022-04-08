# TODO
resource "multy_virtual_network" "example_vn" {
  name       = "dev-vn"
  cidr_block = "10.0.0.0/16"
  cloud      = "aws"
  location   = "ireland"
}
resource "multy_subnet" "subnet" {
  name              = "dev-subnet"
  cidr_block        = "10.0.2.0/24"
  virtual_network   = example_vn
  availability_zone = 2
}
resource "multy_network_interface" "nic" {
  name      = "dev-nic"
  subnet_id = subnet
}
resource "multy_public_ip" "ip" {
  name                 = "dev-ip"
  network_interface_id = multy_network_interface.nic.id
}