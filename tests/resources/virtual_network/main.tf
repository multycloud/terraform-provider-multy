variable "cloud" {
  type    = string
  default = "aws"
}

resource "multy_virtual_network" "vn" {
  name       = "vn_test2"
  cidr_block = "10.0.0.0/16"
  location   = "eu_west_2"
  cloud      = var.cloud
}
