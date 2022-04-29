variable "cloud" {
  type    = string
  default = "aws"
}

resource "multy_vault" "v" {
  name     = "multyvault"
  cloud    = var.cloud
  location = "us_east_1"
}
