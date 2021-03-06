resource "multy_virtual_network" "example_vn" {
  cloud      = "aws"
  name       = "test-nic"
  cidr_block = "10.0.0.0/16"
  location   = "eu_west_1"
}
resource "multy_subnet" "subnet" {
  name               = "test-nic"
  cidr_block         = "10.0.2.0/24"
  virtual_network_id = multy_virtual_network.example_vn.id
}
resource "multy_public_ip" "ip" {
  name     = "test-ip"
  cloud    = "aws"
  location = "eu_west_1"
}
resource "multy_network_interface" "private-nic" {
  cloud        = "aws"
  name         = "test-private-nic"
  subnet_id    = multy_subnet.subnet.id
  public_ip_id = multy_public_ip.ip.id
  location     = "eu_west_1"
}
