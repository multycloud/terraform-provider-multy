terraform {
  required_providers {
    multy = {
      version = "0.0.1"
      source  = "hashicorp.com/dev/multy"
    }
  }
}

variable clouds {
  type    = list(string)
  default = ["aws", "azure"]
}

provider "multy" {
  api_key         = "multy_local"
  server_endpoint = "localhost:8000"
}

resource multy_virtual_network vn {
  for_each = toset(var.clouds)

  name       = "vn_test"
  cidr_block = "10.0.0.0/16"
  cloud      = each.key
  location   = "us_east"
}

output vn {
  value = multy_virtual_network.vn
}
