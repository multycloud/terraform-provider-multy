terraform {
  required_providers {
    multy = {
      version = "0.0.1"
      source  = "hashicorp.com/dev/multy"
    }
  }
}

provider "multy" {
  api_key = "123"
}

resource multy_virtual_network vn {
  name       = "test"
  cidr_block = "10.0.0.0/16"
  cloud      = "aws"
  location   = "ireland"
}

resource multy_subnet subnet {
  name               = "test_subnet"
  cidr_block         = "10.0.10.0/24"
  virtual_network_id = multy_virtual_network.vn.id
  cloud              = "aws"
  location           = "ireland"
}

resource multy_virtual_machine vm {
  name             = "test_vm"
  size             = "micro"
  operating_system = "linux"
  subnet_id        = multy_subnet.subnet.id
  public_ip_id     = "123"
  ssh_key          = file("./ssh_key")
  cloud            = "aws"
  location         = "ireland"
}

resource "multy_network_security_group" nsg {
  name               = "test-nsg"
  virtual_network_id = multy_virtual_network.vn.id
  cloud              = "aws"
  location           = "ireland"
  rule {
    protocol   = "tcp"
    priority   = "120"
    from_port  = "22"
    to_port    = "22"
    cidr_block = "0.0.0.0/0"
    direction  = "bOTH"
  }
  rule {
    protocol   = "tcp"
    priority   = "140"
    from_port  = "443"
    to_port    = "443"
    cidr_block = "0.0.0.0/0"
    direction  = "both"
  }
}
