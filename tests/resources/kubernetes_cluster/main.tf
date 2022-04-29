variable cloud {
  type    = string
  default = "aws"
}

resource "multy_kubernetes_cluster" "cluster1" {
  cloud      = var.cloud
  location   = "us_east_1"
  name       = "multy_cluster_1"
  subnet_ids = [multy_subnet.subnet1.id, multy_subnet.subnet2.id]

  default_node_pool = {
    name                = "default"
    starting_node_count = 2
    min_node_count      = 1
    max_node_count      = 3
    vm_size             = "medium"
    disk_size_gb        = 10
    subnet_ids          = [multy_subnet.subnet1.id, multy_subnet.subnet2.id]
  }

  depends_on = [multy_route_table_association.subnet1]
}


resource "multy_kubernetes_cluster" "cluster2" {
  cloud      = var.cloud
  location   = "us_east_1"
  name       = "multy_cluster_2"
  subnet_ids = [multy_subnet.subnet1.id, multy_subnet.subnet2.id]

  default_node_pool = {
    name           = "default"
    min_node_count = 1
    max_node_count = 3
    vm_size        = "medium"
    disk_size_gb   = 10
    subnet_ids     = [multy_subnet.subnet1.id, multy_subnet.subnet2.id]
    labels         = { "os" : "multy" }
  }

  depends_on = [multy_route_table_association.subnet1]
}

resource "multy_virtual_network" "example_vn" {
  name       = "example_vn"
  cidr_block = "10.0.0.0/16"
  cloud      = var.cloud
  location   = "us_east_1"
}
resource "multy_subnet" "subnet1" {
  name               = "subnet1"
  cidr_block         = "10.0.1.0/24"
  virtual_network_id = multy_virtual_network.example_vn.id
  availability_zone  = 1
}
resource "multy_subnet" "subnet2" {
  name               = "subnet2"
  cidr_block         = "10.0.2.0/24"
  virtual_network_id = multy_virtual_network.example_vn.id
  availability_zone  = 2
}

resource multy_route_table rt {
  name               = "rta_test"
  virtual_network_id = multy_virtual_network.example_vn.id
  route {
    cidr_block  = "0.0.0.0/0"
    destination = "internet"
  }
}

resource multy_route_table_association subnet1 {
  route_table_id = multy_route_table.rt.id
  subnet_id      = multy_subnet.subnet1.id
}