variable clouds {
  type    = set(string)
  default = ["aws"]
}

resource multy_vault v {
  name     = "multyvault"
  cloud    = "aws"
  location = "us_east"
}
resource "multy_vault_secret" s {
  vault_id = multy_vault.v.id
  name     = "api-key"
  value    = "secret-1"
}