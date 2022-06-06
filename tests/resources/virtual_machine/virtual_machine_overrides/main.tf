variable "cloud" {
  type    = string
  default = "azure"
}

variable "location" {
  type    = string
  default = "eu_west_1"
}

resource multy_virtual_network vn {
  name       = "test_vm"
  cidr_block = "10.0.0.0/16"
  cloud      = var.cloud
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
  cloud              = var.cloud
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
  name            = "test_vm"
  size            = "general_micro"
  image_reference = {
    os      = "cent_os"
    version = "8.2"
  }
  subnet_id          = multy_subnet.subnet.id
  generate_public_ip = true
  user_data_base64   = base64encode("#!/bin/bash -xe\nsudo su;\nyum update -y; yum install -y httpd.x86_64; systemctl start httpd.service; systemctl enable httpd.service; touch /var/www/html/index.html; echo \"<h1>Hello from Multy on AWS</h1>\" > /var/www/html/index.html")
  public_ssh_key     = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQCf3a02CbBVs6w3QVsf5yZ+WU+AAVpP86SufnMsSOV29DNXKmAGsB16jqJYq+znqDFTscOmf8WkR/AEKDwU+Q9auvBIWKtwB8aUrd5hCTC0EhC/2322PsOoOs0fEOki39xbaF9vWRXKPES/GM7lHR3xV5TFB4GBiq12mH7ALhHbcAjafxf+/Q3PzCYeJxRDSl7RvjihiMoOgjK9jy1DqlVLgOJUQuLgwxv1Nm1EwVygi5czBoYFXhDGszOuq4xpq8rUBTIGEczMn7glVLIyAIADLUkD0x+frjamI6I3BX1yn9GfJ3BPa8vC5GXsWnLelLeMg5SX8AiB4MfpTirQuvFeMfGPvFvKK6YwcuVHPDYd2/oisIf/wFlmjxXoTA1LEdH7o5/C5swIisEpppcaIO7F0v7gJwEdktpORzSxZEIirYGf8eTrmz2Mx3GH/vGUbUhJtwazx/7Lnv6FZH0ncqlV4DX0BCQZi3AHGWcPcFW/sGTv8EAS8PCQUZdnEptZLI8= joao@Joaos-MB"
  cloud              = var.cloud
  location           = var.location

  depends_on = [multy_network_security_group.nsg]

  aws_overrides = {
    instance_type = "t4g.micro"
  }
  azure_overrides = {
    size = "Standard_A1_v2"
  }
}