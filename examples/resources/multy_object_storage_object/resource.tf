resource "multy_object_storage" "obj_storage" {
  name     = "test-storage-123-multy"
  cloud    = "azure"
  location = "us_east"
}

resource "multy_object_storage_object" "obj_storage" {
  name              = "index.html"
  object_storage_id = multy_object_storage.obj_storage.id
  content           = "<h1>Hello World</h1>"
  content_type      = "text/html"
}