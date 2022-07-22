variable "cloud" {
  type    = string
  default = "azure"
}

resource "multy_virtual_network" "vn" {
  name       = "vn-test2"
  cidr_block = "10.0.0.0/16"
  location   = "eu_west_1"
  cloud      = var.cloud
}
