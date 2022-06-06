resource "multy_kubernetes_node_pool" "node_pool" {
  cluster_id     = multy_kubernetes_cluster.cluster1.id
  name           = "node_pool"
  min_node_count = 1
  max_node_count = 3
  vm_size        = "general_medium"
  disk_size_gb   = 10
  subnet_id      = multy_subnet.subnet1.id
  labels         = { "os" : "multy" }
}
