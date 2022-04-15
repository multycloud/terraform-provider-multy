variable "cloud" {
  type    = string
  default = "aws"
}

resource multy_object_storage "obj_storage" {
  name       = "test-storage-123-multy"
  cloud      = var.cloud
  location   = "us_east"
  versioning = true
}

resource multy_object_storage_object "obj_storage" {
  name              = "test-obj"
  object_storage_id = multy_object_storage.obj_storage.id
  content           = "<h1>hello world</h1>"
  content_type      = "text/html"
}