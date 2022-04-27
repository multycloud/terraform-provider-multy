resource multy_vault vault {
  name     = "test-vault"
  cloud    = "azure"
  location = "us_east"
}
resource "multy_vault_secret" s {
  vault_id = multy_vault.vault.id
  name     = "api-key"
  value    = "secret-1"
}