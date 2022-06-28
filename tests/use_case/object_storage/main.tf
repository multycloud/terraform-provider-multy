terraform {
  required_providers {
    multy = {
      source = "multycloud/multy"
    }
  }
}

provider "multy" {
  api_key         = "XXX-YYY-ZZZ"
  aws             = {}
  azure           = {}
  server_endpoint = "localhost:8000"
}

variable "clouds" {
  type    = set(string)
  default = ["aws", "azure"]
}

resource "random_string" "obj_suffix" {
  length  = 4
  special = false
  upper   = false
}

resource "multy_object_storage" "obj_storage" {
  for_each   = var.clouds
  name       = "multy-test-${random_string.obj_suffix.result}"
  cloud      = each.key
  location   = "us_east_1"
  versioning = true
}

resource "multy_object_storage_object" "public_obj_storage" {
  for_each          = var.clouds
  name              = "hello_world"
  object_storage_id = multy_object_storage.obj_storage[each.key].id
  content_base64    = base64encode("<h1>hello world from ${each.key}</h1>")
  content_type      = "text/html"
  acl               = "public_read"
}

resource "multy_object_storage_object" "private_obj_storage" {
  for_each          = var.clouds
  name              = "super-secret-file"
  object_storage_id = multy_object_storage.obj_storage[each.key].id
  content_base64    = base64encode("<h1>goodbye world from ${each.key}</h1>")
  content_type      = "text/html"
}

output "aws_object_url" {
  value = multy_object_storage_object.public_obj_storage["aws"].url
}

output "azure_object_url" {
  value = multy_object_storage_object.public_obj_storage["azure"].url
}