# TODO
resource multy_object_storage "obj_storage" {
  name       = "dev-storage"
  cloud      = "aws"
  location   = "ireland"
  versioning = true
}