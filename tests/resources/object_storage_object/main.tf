terraform {
  required_providers {
    multy = {
      version = "0.0.1"
      source  = "hashicorp.com/dev/multy"
    }
  }
}

provider "multy" {
  api_key         = "secret-1"
  server_endpoint = "localhost:8000"
}

resource multy_object_storage "obj_storage" {
  name     = "test-storage-123-multy"
  cloud    = "azure"
  location = "us_east"
}

resource multy_object_storage_object "obj_storage" {
  name              = "test-obj"
  object_storage_id = multy_object_storage.obj_storage.id
  content           = "abc"
  content_type      = "text/html"
}