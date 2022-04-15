variable "cloud" {
  type    = string
  default = "aws"
}

resource multy_object_storage "obj_storage" {
  name       = "test-storage"
  cloud      = var.cloud
  location   = "ireland"
  versioning = true
}