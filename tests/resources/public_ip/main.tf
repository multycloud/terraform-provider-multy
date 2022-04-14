resource "multy_virtual_network" "example_vn" {
  name       = "example_vn"
  cidr_block = "10.0.0.0/16"
  cloud      = "aws"
  location   = "ireland"
}
resource "multy_subnet" "subnet" {
  name              = "subnet"
  cidr_block        = "10.0.2.0/24"
  virtual_network   = example_vn
  availability_zone = 2
}
resource "multy_network_interface" "public-nic" {
  name      = "test-public-nic"
  subnet_id = subnet
}
resource "multy_network_interface" "private-nic" {
  name      = "test-private-nic"
  subnet_id = subnet
}
resource "multy_public_ip" "ip" {
  name                 = "test-ip"
  network_interface_id = public-nic
  location             = "ireland"
  cloud                = "aws"
}