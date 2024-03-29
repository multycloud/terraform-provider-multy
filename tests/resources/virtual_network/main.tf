variable "cloud" {
  type    = string
  default = "aws"
}

resource "multy_virtual_network" "vn" {
  name       = "vn-test"
  cidr_block = "10.0.0.0/16"
  location   = "eu_west_1"
  cloud      = var.cloud
}
