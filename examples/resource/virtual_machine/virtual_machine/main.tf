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
  public_ip        = false
  user_data        = "echo HelloWorld"
  ssh_key          = file("./ssh_key.pub")
  cloud            = "aws"
  location         = "ireland"
}