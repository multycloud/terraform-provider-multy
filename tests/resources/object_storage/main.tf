variable "cloud" {
  type    = string
  default = "gcp"
}

resource multy_object_storage "obj_storage" {
  name       = "multytestst774"
  cloud      = var.cloud
  location   = "eu_west_1"
  versioning = true
}