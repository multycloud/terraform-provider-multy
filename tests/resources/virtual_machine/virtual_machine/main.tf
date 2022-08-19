variable "cloud" {
  type    = string
  default = "gcp"
}

variable "location" {
  type    = string
  default = "eu_west_1"
}

resource multy_virtual_network vn {
  name       = "test-vm"
  cidr_block = "10.0.0.0/16"
  cloud      = var.cloud
  location   = var.location
}

resource multy_subnet subnet {
  name               = "test-vm"
  cidr_block         = "10.0.10.0/24"
  virtual_network_id = multy_virtual_network.vn.id
}

resource "multy_network_security_group" nsg {
  name               = "test-vm"
  virtual_network_id = multy_virtual_network.vn.id
  cloud              = var.cloud
  location           = var.location
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

resource "multy_route_table" "rt" {
  name               = "rt"
  virtual_network_id = multy_virtual_network.vn.id
  route {
    cidr_block  = "0.0.0.0/0"
    destination = "internet"
  }
}

resource multy_route_table_association rta1 {
  route_table_id = multy_route_table.rt.id
  subnet_id      = multy_subnet.subnet.id
}

resource multy_virtual_machine vm {
  name            = "test-vm"
  size            = "general_micro"
  image_reference = {
    os      = "ubuntu"
    version = "20.04"
  }
  subnet_id          = multy_subnet.subnet.id
  generate_public_ip = true
  user_data_base64   = base64encode(<<-EOF
      #!/bin/bash -xe
      sudo su
      apt update -y && apt install -y apache2
      systemctl enable apache2
      touch /var/www/html/index.html
      echo "<h1>Hello from Multy on ${var.cloud}</h1>" > /var/www/html/index.html
    EOF
  )
  cloud                      = var.cloud
  location                   = var.location
  network_security_group_ids = [multy_network_security_group.nsg.id]
  public_ssh_key             = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQDQJRMv5ix5QPfs3+t6mLcXvtBuMdxHc5aitbWE4myzVf33UttLI4td1Q2WZ+kfn5SwiF7b9YowqWlM0kiIFGoJhcBwN0Mq4TOcGmn5Hidl94Rf8xzk88+W0LJEd+JEKY4czZNNgdWhhsxuZgt9P4NdzONqFC2XL5AggLau7SdDVV9JCHHI+dw1C0FLq1Y/5Ga7rJN2Zm7vmT4I/tCPEheEDYN2MH2ClKgQf4Ni2KoiHLxvrbBmcXOuknn+/yjN+dpiAncQFnMjykV5lKMnXFm6u43KlMLpr/XKKmdaLDZWBVaNRdPeqt2FiWisGFowAuUsXWUhSBcwXArZbWRbc0rp+ASh9is2fC9BvkOcCdMZTqJmoXFib23KaSKi1wZhGOPB93YJvZ/Gnr6n0GHWryqPJ6QvijmN20zYJq1j4ZUrNsxLgIMkGNMtCId+pbJ1amn1DM7X2fT99k+rcycTIuI83OEASs49hJfty0AM1ChSxvkm0ZAZQcFsdUnuAAyhXgU= root@Joaos-MacBook-Pro.local"
  depends_on = [multy_route_table_association.rta1]
}
