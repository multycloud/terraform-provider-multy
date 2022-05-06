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
}
