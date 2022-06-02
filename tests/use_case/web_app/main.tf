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
  server_endpoint = "localhost:8000"
  aws             = {}
  azure           = {}
  api_key         = "xx"
}

variable "location" {
  type    = string
  default = "eu_west_1"
}

variable "clouds" {
  type    = set(string)
  default = ["aws", "azure"]
}

variable "vm_clouds" {
  type    = set(string)
  default = ["aws", "azure"]
}

variable "db_cloud" {
  type    = string
  default = "aws"
}

resource multy_virtual_network vn {
  for_each   = var.clouds
  name       = "web_app_vn"
  cidr_block = "10.0.0.0/16"
  cloud      = each.key
  location   = var.location
}

resource multy_subnet public_subnet {
  for_each           = var.clouds
  name               = "web_app_public_subnet"
  cidr_block         = "10.0.10.0/24"
  virtual_network_id = multy_virtual_network.vn[each.key].id
}

resource multy_subnet private_subnet {
  for_each           = var.clouds
  name               = "web_app_private_subnet"
  cidr_block         = "10.0.11.0/24"
  virtual_network_id = multy_virtual_network.vn[each.key].id
  availability_zone  = 1
}

resource multy_subnet private_subnet2 {
  for_each           = var.clouds
  name               = "web_app_private_subnet2"
  cidr_block         = "10.0.12.0/24"
  virtual_network_id = multy_virtual_network.vn[each.key].id
  availability_zone  = 2
}

resource "multy_route_table" "rt" {
  for_each           = var.clouds
  name               = "web_app_rt"
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
  subnet_id      = multy_subnet.private_subnet[each.key].id
}
resource multy_route_table_association rta3 {
  for_each       = var.clouds
  route_table_id = multy_route_table.rt[each.key].id
  subnet_id      = multy_subnet.private_subnet2[each.key].id
}

resource "multy_network_security_group" nsg {
  for_each           = var.clouds
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
}

resource multy_virtual_machine vm {
  #  for_each        = var.clouds
  for_each        = var.vm_clouds
  name            = "web_app_vm"
  size            = each.key == "azure" ? "large" : "micro"
  image_reference = {
    os      = "ubuntu"
    version = "18.04"
  }
  subnet_id          = multy_subnet.public_subnet[each.key].id
  generate_public_ip = true
  user_data_base64   = base64encode(templatefile("./${each.key}_init.sh", {
    vault_name : multy_vault.web_app_vault[each.key].name,
    db_host_secret_name : multy_vault_secret.db_host[each.key].value,
    db_username_secret_name : multy_vault_secret.db_username[each.key].value,
    db_password_secret_name : multy_vault_secret.db_password[each.key].value,
  }))
  network_security_group_ids = [multy_network_security_group.nsg[each.key].id]

  public_ssh_key = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQCf3a02CbBVs6w3QVsf5yZ+WU+AAVpP86SufnMsSOV29DNXKmAGsB16jqJYq+znqDFTscOmf8WkR/AEKDwU+Q9auvBIWKtwB8aUrd5hCTC0EhC/2322PsOoOs0fEOki39xbaF9vWRXKPES/GM7lHR3xV5TFB4GBiq12mH7ALhHbcAjafxf+/Q3PzCYeJxRDSl7RvjihiMoOgjK9jy1DqlVLgOJUQuLgwxv1Nm1EwVygi5czBoYFXhDGszOuq4xpq8rUBTIGEczMn7glVLIyAIADLUkD0x+frjamI6I3BX1yn9GfJ3BPa8vC5GXsWnLelLeMg5SX8AiB4MfpTirQuvFeMfGPvFvKK6YwcuVHPDYd2/oisIf/wFlmjxXoTA1LEdH7o5/C5swIisEpppcaIO7F0v7gJwEdktpORzSxZEIirYGf8eTrmz2Mx3GH/vGUbUhJtwazx/7Lnv6FZH0ncqlV4DX0BCQZi3AHGWcPcFW/sGTv8EAS8PCQUZdnEptZLI8= joao@Joaos-MB"
  cloud          = each.key
  location       = var.location

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
  subnet_ids     = [multy_subnet.private_subnet[var.db_cloud].id, multy_subnet.private_subnet2[var.db_cloud].id]
  cloud          = var.db_cloud
  location       = var.location

  depends_on = [multy_route_table_association.rta2, multy_route_table_association.rta3]
}

resource "multy_vault" "web_app_vault" {
  for_each = var.clouds
  name     = "web-app-vault-test"
  cloud    = each.key
  location = var.location
}
resource "multy_vault_secret" "db_host" {
  for_each = var.clouds
  name     = "db-host"
  vault_id = multy_vault.web_app_vault[each.key].id
  value    = multy_database.example_db.hostname
}
resource "multy_vault_secret" "db_username" {
  for_each = var.clouds
  name     = "db-username"
  vault_id = multy_vault.web_app_vault[each.key].id
  value    = multy_database.example_db.username
}
resource "multy_vault_secret" "db_password" {
  for_each = var.clouds
  name     = "db-password"
  vault_id = multy_vault.web_app_vault[each.key].id
  value    = multy_database.example_db.password
}
resource "multy_vault_access_policy" "kv_ap" {
  for_each = var.vm_clouds
  vault_id = multy_vault.web_app_vault[each.key].id
  identity = multy_virtual_machine.vm[each.key].identity
  access   = "owner"
}
resource multy_object_storage app {
  name     = "multy-web-app-test"
  cloud    = var.db_cloud
  location = var.location
}
output "endpoint" {
  value = {
  for k, vm in multy_virtual_machine.vm : k => "http://${vm.public_ip}:4000"
  }
}