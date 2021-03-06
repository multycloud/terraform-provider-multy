resource "multy_network_interface" "public-nic" {
  name      = "test-public-nic"
  subnet_id = multy_subnet.subnet.id
  cloud     = "aws"
  location  = "eu_west_1"
}
resource "multy_public_ip" "ip" {
  name     = "test-ip"
  cloud    = "aws"
  location = "eu_west_1"
}