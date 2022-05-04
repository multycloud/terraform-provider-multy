resource "multy_vault" "web_app_vault" {
  name     = "web-app-vault-test"
  cloud    = "azure"
  location = "ireland"
}


resource "multy_vault_access_policy" "kv_ap" {
  vault_id = multy_vault.web_app_vault.id
  identity = multy_virtual_machine.vm.identity
  access   = "owner"
}
