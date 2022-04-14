variable clouds {
  type    = set(string)
  default = ["aws"]
}

resource multy_virtual_network vn {
  for_each = var.clouds

  name       = "vn_test2"
  cidr_block = "10.0.0.0/16"
  location   = "us_east"
  cloud      = each.key
}
