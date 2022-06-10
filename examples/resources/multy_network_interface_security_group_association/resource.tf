resource "multy_network_interface" "nic" {
  cloud     = "aws"
  name      = "test-private-nic"
  subnet_id = multy_subnet.subnet.id
  location  = "eu_west_1"
}
resource "multy_network_security_group" "nsg" {
  name               = "test_nsg"
  virtual_network_id = multy_virtual_network.vn.id
  cloud              = "aws"
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