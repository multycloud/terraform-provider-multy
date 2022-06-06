variable "cloud" {
  type    = string
  default = "aws"
}


resource multy_virtual_network vn {
  name       = "test"
  cidr_block = "10.0.0.0/16"
  cloud      = var.cloud
  location   = "eu_west_1"
}

resource multy_subnet subnet {
  name               = "test_subnet"
  cidr_block         = "10.0.10.0/24"
  virtual_network_id = multy_virtual_network.vn.id
  location           = "eu_west_1"
}

resource multy_virtual_machine vm {
  name         = "test_vm"
  size         = "general_micro"
  subnet_id    = multy_subnet.subnet.id
  public_ip_id = "123"
  ssh_key      = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQDFRQk+HkW4QXy1EdEd6BcCQcaT8pb/ySF98GvbXFTP/qZEnzl074SaBzefMP0zZi3N5vQD6tBWe/uxpZUKsHqkti+l6S3eR8Ols0E7jSpbLvfV+cBeNle7bdzH76V0SjUc3xEkAZNLrcKTNQgnot69ChE/Z5URwL1dMeD8GXATVtSH/AvGat3PSexkL75rWbCBXXmr+5/Re8kLSqYPf6WsLUbI6rIp3Okd1Kmo8pIHq9fqm/B9HSJjOXl08G2RC2H02+HIzRc6AIIqFBbPTQwjw5VEHaZiUC7tl5S117CpAx8oiv8njjR6+sNfEocjaPYl9ks/cVmpY1jCtEiP/5rBmfTSaBVm1BqAqbyLt+H2j7E/IzJBT1SWSy/tlk7r/E32b+JXCLfytNkoOlX7v3PrY9gy8927+4n0rmkLAHcglpXt93/Qqy81fv/QMmhLsnxL6JFrlvjx1X5GIiHvid3AG3K9Pm925whxMNN3HOLHxQPHghtB2iCgiv0DpU9sLjs= joao@Joaos-MBP"
  cloud        = var.cloud
  location     = "eu_west_1"
}

resource "multy_network_security_group" nsg {
  name               = "test-nsg"
  virtual_network_id = multy_virtual_network.vn.id
  cloud              = var.cloud
  location           = "eu_west_1"
  rule {
    protocol   = "xx"
    priority   = "120"
    from_port  = "-9"
    to_port    = "9"
    cidr_block = "0.0.0.0"
    direction  = "down"
  }
}

resource "multy_network_security_group" nsg_empty {
  name               = "nsg_empty"
  virtual_network_id = multy_virtual_network.vn.id
  cloud              = var.cloud
  location           = "eu_west_1"
}
