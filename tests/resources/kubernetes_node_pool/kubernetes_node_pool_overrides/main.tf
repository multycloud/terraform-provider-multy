variable cloud {
  type    = string
  default = "azure"
}

resource "multy_kubernetes_cluster" "cluster1" {
  cloud              = var.cloud
  location           = "us_east_1"
  name               = "multy-cluster1"
  virtual_network_id = multy_virtual_network.example_vn.id

  default_node_pool = {
    name                = "default"
    starting_node_count = 3
    min_node_count      = 3
    max_node_count      = 3
    vm_size             = "general_medium"
    disk_size_gb        = 10
    subnet_id           = multy_subnet.subnet1.id
  }

  depends_on = [multy_route_table_association.subnet1]
}


resource "multy_kubernetes_node_pool" "node_pool" {
  cluster_id     = multy_kubernetes_cluster.cluster1.id
  name           = "pool"
  min_node_count = 3
  max_node_count = 3
  vm_size        = "general_medium"
  disk_size_gb   = 10
  subnet_id      = multy_subnet.subnet1.id
  labels         = { "os" : "multy" }

  aws_overrides = {
    instance_types = ["t4g.micro"]
  }
  azure_overrides = {
    vm_size = "Standard_A1_v2"
  }
}

resource "multy_virtual_network" "example_vn" {
  name       = "example-vn"
  cidr_block = "10.0.0.0/16"
  cloud      = var.cloud
  location   = "us_east_1"
}
resource "multy_subnet" "subnet1" {
  name               = "subnet1"
  cidr_block         = "10.0.1.0/24"
  virtual_network_id = multy_virtual_network.example_vn.id
}

resource multy_route_table rt {
  name               = "rta-test"
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