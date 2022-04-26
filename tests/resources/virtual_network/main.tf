variable "cloud" {
  type    = set(string)
  default = ["aws", "azure"]
}

resource "multy_virtual_network" "vn" {
  for_each   = var.cloud
  name       = "vn_test2"
  cidr_block = "10.0.0.0/16"
  location   = "us_east"
  cloud      = each.key
}
