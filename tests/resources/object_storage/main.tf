terraform {
  required_providers {
    multy = {
      version = "0.0.1"
      source  = "hashicorp.com/dev/multy"
    }
  }
}

provider "multy" {
  api_key = "multy_local"
}

resource multy_object_storage "obj_storage" {
  name       = "test-storage"
  cloud      = "aws"
  location   = "ireland"
  versioning = true
}