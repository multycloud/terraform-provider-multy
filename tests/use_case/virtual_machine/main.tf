terraform {
  required_providers {
    multy = {
      version = "0.0.1"
      source  = "hashicorp.com/dev/multy"
    }
  }
}

provider "multy" {
  server_endpoint = "localhost:8000"
  aws             = {}
  azure           = {}
  gcp             = { project = "multy-project" }
  api_key         = "goncalo"
}

variable "location" {
  type    = string
  default = "eu_west_1"
}

variable "clouds" {
  type    = set(string)
  default = ["aws", "azure", "gcp"]
}

variable "vm_clouds" {
  type    = set(string)
  default = []
}

variable "db_cloud" {
  type    = string
  default = "aws"
}

resource multy_virtual_network vn {
  for_each   = var.clouds
  name       = "web-app-vn"
  cidr_block = "10.0.0.0/16"
  cloud      = each.key
  location   = var.location
}

resource multy_subnet public_subnet {
  for_each           = var.clouds
  name               = "web-app-public-subnet"
  cidr_block         = "10.0.10.0/24"
  virtual_network_id = multy_virtual_network.vn[each.key].id
}

resource multy_subnet db_subnet {
  for_each           = var.clouds
  name               = "web-app-db-subnet"
  cidr_block         = "10.0.11.0/24"
  virtual_network_id = multy_virtual_network.vn[each.key].id
}

resource "multy_route_table" "rt" {
  for_each           = var.clouds
  name               = "web-app-rt"
  virtual_network_id = multy_virtual_network.vn[each.key].id
  route {
    cidr_block  = "0.0.0.0/0"
    destination = "internet"
  }
}

resource multy_route_table_association rta1 {
  for_each       = var.clouds
  route_table_id = multy_route_table.rt[each.key].id
  subnet_id      = multy_subnet.public_subnet[each.key].id
}
resource multy_route_table_association rta2 {
  for_each       = var.clouds
  route_table_id = multy_route_table.rt[each.key].id
  subnet_id      = multy_subnet.db_subnet[each.key].id
}

resource "multy_network_security_group" nsg {
  for_each           = var.clouds
  name               = "web-app-nsg"
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
    priority   = 133
    from_port  = 80
    to_port    = 80
    cidr_block = "0.0.0.0/0"
    direction  = "both"
  }
  rule {
    protocol   = "tcp"
    priority   = 132
    from_port  = 4000
    to_port    = 4000
    cidr_block = "0.0.0.0/0"
    direction  = "both"
  }
  rule {
    protocol   = "tcp"
    priority   = 135
    from_port  = 3306
    to_port    = 3306
    cidr_block = "0.0.0.0/0"
    direction  = "both"
  }
}

resource multy_virtual_machine vm {
  for_each = var.vm_clouds

  name            = "web-app-vm2"
  size            = "general_large"
  image_reference = {
    os      = "ubuntu"
    version = "18.04"
  }
  subnet_id          = multy_subnet.public_subnet[each.key].id
  generate_public_ip = true
  user_data_base64   = base64encode(templatefile("./${each.key}_init.sh", {
    db_host_secret_name : multy_database.example_db.hostname,
    db_username_secret_name : multy_database.example_db.username,
    db_password_secret_name : multy_database.example_db.password,
  }))
  network_security_group_ids = [multy_network_security_group.nsg[each.key].id]

  public_ssh_key = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQC4vxp+ZIuSTFfBU5XjCwoOr/Pz7sfEBkcjiV9lm8tcpC7HTjlN7VPaQvHiVRyKOdK5/I3CcE3jWi8rrkxmDeFxvLgmJLMrd1zANOlWmrr4+6DVPIGPnGru+p+/A+/fEO1sKb5W/2Zrpo7IrhpWI8hKHPN26OLZyze9vWLy+4aHeIyDPGzNcNQ8MUzvF5s/YuZj8s9Nx6xMNVwLzm1lPX37SGpg2p5mcGSShV5OQp9keYgMHfbpsloVU5Yq6yaZsL/nq0n0h84znLBT4WfNvtc8qN2SFTsOEDsFJ3ulIp2pvrstllujzM92KMl+ONpeDs0yVlMaVTth4CfC0ZY9FOR7PBZ0UWeHcXtZC5UUHcMyK6x/wUROt8WVaKpXHUKWhOxAROEnzMc5F/IRmw2h61JEf8764EmcAvFxL8GzzgQhWKr6I61TJZQYNOq+D5gLqHR567Ty8eXRhgF9EqKMKb6vGgIJ7iUpYi5UDpLJObHREwCDOk2a8AhD8QFld3VXXQGcb36b/ODrDlIP3AHl+7LNmgZFy9+6MFg/F92nPj5B971oG1e2nUWHna9Mp4wyJJQk5RK0DhQ+dxZnmrs3NKBXo58o1CXD6PLGZy0fhvXeBRqB/QbaexeBPviQuNv6gfKnX5AchcgrofK32nTsgA95tXoSu1Ci/4Ea0bLs0xkPSQ== joaocoelho@Joao-MBP"
  cloud          = each.key
  location       = var.location

  #  aws_overrides = {
  #    instance_type = "c4.8xlarge"
  #  }

  depends_on = [multy_network_security_group.nsg]
}
resource "multy_database" "example_db" {
  storage_gb     = 10
  name           = "exampledbmulty"
  engine         = "mysql"
  engine_version = "5.7"
  username       = "multyadmin"
  password       = "multy-Admin123!"
  size           = "micro"
  subnet_id      = multy_subnet.db_subnet[var.db_cloud].id
  cloud          = var.db_cloud
  location       = var.location

  depends_on = [multy_route_table_association.rta2]
}
output "endpoint" {
  value = {
  for k, vm in multy_virtual_machine.vm : k => "http://${vm.public_ip}:4000"
  }
}
