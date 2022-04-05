resource multy_virtual_network vn {
  for_each = ["aws", "azure"]

  name       = "vn_test2"
  cidr_block = "10.0.0.0/16"
  cloud      = each.key
  location   = "us_east"
}