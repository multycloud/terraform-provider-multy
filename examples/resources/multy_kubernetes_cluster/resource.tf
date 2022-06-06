resource "multy_kubernetes_cluster" "cluster1" {
  cloud              = var.cloud
  location           = "us_east_1"
  name               = "multy_cluster_1"
  virtual_network_id = multy_virtual_network.vn.id

  default_node_pool = {
    name                = "default"
    starting_node_count = 2
    min_node_count      = 1
    max_node_count      = 3
    vm_size             = "general_medium"
    disk_size_gb        = 10
    subnet_id           = multy_subnet.subnet1.id
  }
}
