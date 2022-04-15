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

resource multy_object_storage_object "public_obj_storage" {
  name              = "hello_world"
  object_storage_id = multy_object_storage.obj_storage.id
  content           = "<h1>hello world from ${var.cloud}</h1>"
  content_type      = "text/html"
  acl               = "public_read"
}

resource multy_object_storage_object "private_obj_storage" {
  name              = "super-secret-file"
  object_storage_id = multy_object_storage.obj_storage.id
  content           = "<h1>goodbye world from ${var.cloud}</h1>"
  content_type      = "text/html"
}

output "object_url" {
  value = multy_object_storage_object.public_obj_storage.url
}