terraform {
  required_providers {
    multy = {
      version = "0.0.1"
      source  = "hashicorp.com/dev/multy"
    }
  }
}

provider "multy" {
  api_key         = "secret-1"
  server_endpoint = "localhost:8000"
  aws             = {}
}
resource "multy_database" "example_db" {
  cloud          = "aws"
  location       = "us_east"
  storage_gb     = 10
  name           = "exampledbmulty"
  engine         = "mysql"
  engine_version = "5.7"
  username       = "multyadmin"
  password       = "multy$Admin123!"
  size           = "micro"
  subnet_ids     = [multy_subnet.subnet1.id, multy_subnet.subnet2.id]
}
resource "multy_virtual_network" "example_vn" {
  name       = "example_vn"
  cidr_block = "10.0.0.0/16"
  cloud      = "aws"
  location   = "us_east"
}
resource "multy_subnet" "subnet1" {
  name               = "subnet1"
  cidr_block         = "10.0.1.0/24"
  virtual_network_id = multy_virtual_network.example_vn.id
  availability_zone  = 1
}
resource "multy_subnet" "subnet2" {
  name               = "subnet2"
  cidr_block         = "10.0.2.0/24"
  virtual_network_id = multy_virtual_network.example_vn.id
  availability_zone  = 2
}