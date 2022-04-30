resource "multy_virtual_network" "vn" {
  name       = "dev-vn"
  cidr_block = "10.0.0.0/16"
  cloud      = "aws"
  location   = "eu_west_1"
}

resource "multy_subnet" "subnet" {
  name               = "dev-subnet"
  cidr_block         = "10.0.10.0/24"
  virtual_network_id = multy_virtual_network.vn.id
}

resource "multy_virtual_machine" "vm" {
  name               = "dev-vm"
  size               = "micro"
  operating_system   = "linux"
  subnet_id          = multy_subnet.subnet.id
  generate_public_ip = true
  user_data_base64   = "echo 'Hello World'"
  public_ssh_key     = file("./ssh_key.pub")
  cloud              = "aws"
  location           = "eu_west_1"
}