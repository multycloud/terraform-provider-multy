terraform {
  required_providers {
    multy = {
      #      source = "multycloud/multy"
      version = "0.0.1"
      source  = "hashicorp.com/dev/multy"
    }
  }
}

provider "multy" {
  api_key         = "secret-2"
  server_endpoint = "localhost:8000"
  aws             = {}
}

variable "location" {
  type    = string
  default = "ireland"
}

variable "clouds" {
  type    = list(string)
  default = ["aws"]
}

resource multy_virtual_network vn {
  for_each   = toset(var.clouds)
  name       = "web_app_vn"
  cidr_block = "10.0.0.0/16"
  cloud      = each.key
  location   = var.location
}

resource multy_subnet public_subnet {
  for_each           = toset(var.clouds)
  name               = "web_app_public_subnet"
  cidr_block         = "10.0.10.0/24"
  virtual_network_id = multy_virtual_network.vn[each.key].id
}

resource multy_subnet private_subnet {
  for_each           = toset(var.clouds)
  name               = "web_app_private_subnet"
  cidr_block         = "10.0.11.0/24"
  virtual_network_id = multy_virtual_network.vn[each.key].id
}

resource multy_subnet private_subnet2 {
  for_each           = toset(var.clouds)
  name               = "web_app_private_subnet2"
  cidr_block         = "10.0.12.0/24"
  virtual_network_id = multy_virtual_network.vn[each.key].id
}

resource "multy_route_table" "rt" {
  for_each           = toset(var.clouds)
  name               = "web_app_rt"
  virtual_network_id = multy_virtual_network.vn[each.key].id
  cloud              = each.key
  route {
    cidr_block  = "0.0.0.0/0"
    destination = "internet"
  }
}

resource multy_route_table_association rta3 {
  route_table_id = multy_route_table.rt.id
  subnet_id      = multy_subnet.public_subnet
}

resource "multy_network_security_group" nsg {
  for_each           = toset(var.clouds)
  name               = "web_app_nsg"
  virtual_network_id = multy_virtual_network.vn[each.key].id
  cloud              = each.key
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
}

resource "multy_database" "example_db" {
  for_each       = toset(var.clouds)
  storage_gb     = 10
  name           = "exampledbmulty"
  engine         = "mysql"
  engine_version = "5.7"
  username       = "multyadmin"
  password       = "multy$Admin123!"
  size           = "micro"
  subnet_ids     = [multy_subnet.private_subnet[each.key].id, multy_subnet.private_subnet2[each.key].id]
  cloud          = each.key
  location       = var.location
}

resource multy_virtual_machine vm {
  for_each         = toset(var.clouds)
  name             = "web_app_vm"
  size             = "micro"
  operating_system = "linux"
  subnet_id        = multy_subnet.public_subnet[each.key].id
  public_ip        = true
  user_data        = base64encode(file("./${each.key}_init.sh"))
  public_ssh_key   = file("./ssh_key.pub")
  cloud            = each.key
  location         = var.location

  depends_on = [multy_network_security_group.nsg]
}