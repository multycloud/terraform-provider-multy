resource "random_password" "password" {
  length = 16
}

resource "multy_database" "example_db" {
  cloud          = "aws"
  location       = "us_east_1"
  storage_gb     = 10
  name           = "multydb"
  engine         = "mysql"
  engine_version = "5.7"
  username       = "multyadmin"
  password       = random_password.password.result
  size           = "micro"
  subnet_ids     = [multy_subnet.subnet1.id, multy_subnet.subnet2.id]
}