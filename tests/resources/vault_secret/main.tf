variable cloud {
  type    = string
  default = "azure"
}

resource multy_vault v {
  name     = "multyvault"
  cloud    = var.cloud
  location = "us_east_1"
}
resource "multy_vault_secret" s {
  vault_id = multy_vault.v.id
  name     = "api-key"
  value    = "secret-1"
}
