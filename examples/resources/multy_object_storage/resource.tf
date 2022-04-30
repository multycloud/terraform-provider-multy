resource "multy_object_storage" "obj_storage" {
  name       = "dev-storage"
  cloud      = "aws"
  location   = "eu_west_1"
  versioning = true
}