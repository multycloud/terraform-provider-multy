variable "location" {
  type    = string
  default = "ireland"
}

resource multy_virtual_network vn {
  name       = "test_vm"
  cidr_block = "10.0.0.0/16"
  cloud      = "aws"
  location   = var.location
}

resource multy_subnet subnet {
  name               = "test_vm"
  cidr_block         = "10.0.10.0/24"
  virtual_network_id = multy_virtual_network.vn.id
}

resource "multy_network_security_group" nsg {
  name               = "test_vm"
  virtual_network_id = multy_virtual_network.vn.id
  cloud              = "aws"
  location           = var.location
  rule {
    protocol   = "tcp"
    priority   = 120
    from_port  = 22
    to_port    = 22
    cidr_block = "0.0.0.0/0"
    direction  = "both"
  }
  rule {
    protocol   = "tcp"
    priority   = 131
    from_port  = 443
    to_port    = 443
    cidr_block = "0.0.0.0/0"
    direction  = "both"
  }
  rule {
    protocol   = "tcp"
    priority   = 132
    from_port  = 80
    to_port    = 80
    cidr_block = "0.0.0.0/0"
    direction  = "both"
  }
  depends_on = [multy_subnet.subnet]
}

resource multy_virtual_machine vm {
  name               = "test_vm"
  size               = "micro"
  operating_system   = "linux"
  subnet_id          = multy_subnet.subnet.id
  generate_public_ip = true
  user_data          = "#!/bin/bash -xe\nsudo su;\nyum update -y; yum install -y httpd.x86_64; systemctl start httpd.service; systemctl enable httpd.service; touch /var/www/html/index.html; echo \"<h1>Hello from Multy on AWS</h1>\" > /var/www/html/index.html"
  public_ssh_key     = file("./ssh_key.pub")
  cloud              = "aws"
  location           = var.location

  depends_on = [multy_network_security_group.nsg]
}