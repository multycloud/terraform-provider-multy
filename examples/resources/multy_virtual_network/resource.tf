resource multy_virtual_network vn {
  for_each = ["aws", "azure"]

  name       = "dev-vn"
  cidr_block = "10.0.0.0/16"
  cloud      = each.key
  location   = "us_east_1"
}