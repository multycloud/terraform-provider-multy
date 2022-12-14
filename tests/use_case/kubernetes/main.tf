variable "clouds" {
  type    = set(string)
  default = ["aws", "azure", "gcp"]
}

variable "aws_region" {
  type    = string
  default = "us_west_1"
}

variable "region" {
  type    = string
  default = "eu_west_2"
}

resource "multy_kubernetes_cluster" "cluster1" {
  for_each = var.clouds

  cloud              = each.value
  location           = each.value == "aws" ? var.aws_region : var.region
  name               = "multy-cluster1"
  virtual_network_id = multy_virtual_network.example_vn[each.value].id
  service_cidr       = "10.100.0.0/16"

  default_node_pool = {
    name                = "default"
    #    starting_node_count = each.value == "aws" ? 2 : 1
    starting_node_count = 1
    min_node_count      = 1
    max_node_count      = 3
    disk_size_gb        = 30
    vm_size             = "general_medium"
    subnet_id           = multy_subnet.subnet1[each.value].id
    availability_zones  = [2]
  }

  depends_on = [multy_route_table_association.subnet1]
}

resource "multy_virtual_network" "example_vn" {
  for_each = var.clouds

  cloud      = each.value
  name       = "example-vn"
  cidr_block = "10.0.0.0/16"
  location   = each.value == "aws" ? var.aws_region : var.region
}
resource "multy_subnet" "subnet1" {
  for_each = var.clouds

  name               = "subnet1"
  cidr_block         = "10.0.1.0/24"
  virtual_network_id = multy_virtual_network.example_vn[each.key].id
}

resource "multy_route_table" "rt" {
  for_each = var.clouds

  name               = "rta-test"
  virtual_network_id = multy_virtual_network.example_vn[each.key].id
  route {
    cidr_block  = "0.0.0.0/0"
    destination = "internet"
  }
}

resource "multy_route_table_association" "subnet1" {
  for_each = var.clouds

  route_table_id = multy_route_table.rt[each.value].id
  subnet_id      = multy_subnet.subnet1[each.value].id
}
