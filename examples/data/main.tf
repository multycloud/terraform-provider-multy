terraform {
  required_providers {
    multy = {
      version = "0.0.1"
      source  = "hashicorp.com/dev/multy"
    }
  }
}

provider "multy" {
  api_key  = "123"
  location = "eu_west_1"
}

data multy_virtual_network vn {
  id = "OGZkNTkyZmItYzVhMi00YTM1LWE2NjItZTFiOWJlYWM2OGVj"
}