terraform {
  required_providers {
    multy = {
      source = "multycloud/multy"
    }
  }
}

variable clouds {
  type    = list(string)
  default = ["aws"]
}

provider "multy" {
  aws     = {}
  api_key = "secret-2"
}

resource multy_virtual_network vn {
  for_each = toset(var.clouds)

  name       = "vn_test2"
  cidr_block = "10.0.0.0/16"
  cloud      = each.key
  location   = "us_east"
}


output vn {
  value = multy_virtual_network.vn
}
