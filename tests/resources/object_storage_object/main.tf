variable "cloud" {
  type    = string
  default = "gcp"
}

resource multy_object_storage "obj_storage" {
  name       = "teststorage123multy"
  cloud      = var.cloud
  location   = "us_east_1"
  versioning = true
}

resource multy_object_storage_object "public_obj_storage" {
  name              = "hello-world.html"
  object_storage_id = multy_object_storage.obj_storage.id
  content_base64    = base64encode("<h1>hello world from ${var.cloud}</h1>")
  content_type      = "text/html"
  acl               = "public_read"
}

resource multy_object_storage_object "private_obj_storage" {
  name              = "super-secret-file.html"
  object_storage_id = multy_object_storage.obj_storage.id
  content_base64    = base64encode("<h1>goodbye world from ${var.cloud}</h1>")
  content_type      = "text/html"
}

output "object_url" {
  value = multy_object_storage_object.public_obj_storage.url
}
