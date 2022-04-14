resource multy_object_storage "obj_storage" {
  name       = "test-storage"
  cloud      = "aws"
  location   = "ireland"
  versioning = true
}