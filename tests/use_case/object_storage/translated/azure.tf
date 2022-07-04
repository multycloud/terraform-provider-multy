provider "azurerm" {
  features {}
}


resource "azurerm_resource_group" "rg1" {
  name     = "rg1"
  location = "eastus"
}
resource "azurerm_storage_account" "obj_storage_azure" {
  resource_group_name             = azurerm_resource_group.rg1.name
  name                            = "multy-test-${random_string.obj_suffix.result}"
  location                        = "eastus"
  account_tier                    = "Standard"
  account_replication_type        = "GZRS"
  allow_nested_items_to_be_public = true
  blob_properties {
    versioning_enabled = false
  }
}
resource "azurerm_storage_container" "obj_storage_azure_public" {
  name                  = "public"
  storage_account_name  = azurerm_storage_account.obj_storage_azure.name
  container_access_type = "blob"
}
resource "azurerm_storage_container" "obj_storage_azure_private" {
  name                  = "private"
  storage_account_name  = azurerm_storage_account.obj_storage_azure.name
  container_access_type = "private"
}

resource "azurerm_storage_blob" "file1_public_azure" {
  name                   = "hello_world"
  storage_account_name   = azurerm_storage_account.obj_storage_azure.name
  storage_container_name = azurerm_storage_container.obj_storage_azure_public.name
  type                   = "Block"
  source_content         = "<h1>hello world from azure</h1>"
  content_type           = "text/html"
}

output "azure_object_url" {
  value = "https://${azurerm_storage_account.obj_storage_azure.name}.blob.core.windows.net/public/${azurerm_storage_blob.file1_public_azure.name}"
}