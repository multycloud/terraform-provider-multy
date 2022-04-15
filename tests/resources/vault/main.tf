variable "clouds" {
  type    = set(string)
  default = ["aws"]
}

resource "multy_vault" "v" {
  name     = "multyvault"
  cloud    = "aws"
  location = "us_east"
}
