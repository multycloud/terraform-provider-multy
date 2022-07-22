variable cloud {
  type    = string
  default = "aws"
}

resource "multy_database" "example_db" {
  cloud          = var.cloud
  location       = "us_east_1"
  storage_gb     = 10
  name           = "exampledbmulty"
  engine         = "mysql"
  engine_version = "5.7"
  username       = "multyadmin"
  password       = "multy$Admin123!"
  size           = "micro"
  subnet_id      = multy_subnet.subnet1.id
}
resource "multy_virtual_network" "example_vn" {
  name       = "example-vn"
  cidr_block = "10.0.0.0/16"
  cloud      = var.cloud
  location   = "us_east_1"
}
resource "multy_subnet" "subnet1" {
  name               = "subnet1"
  cidr_block         = "10.0.1.0/24"
  virtual_network_id = multy_virtual_network.example_vn.id
}