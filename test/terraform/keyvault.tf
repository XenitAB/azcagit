resource "azurerm_key_vault" "tenant_kv" {
  name                       = "kvcontainerapps"
  location                   = azurerm_resource_group.tenant.location
  resource_group_name        = azurerm_resource_group.tenant.name
  tenant_id                  = data.azurerm_client_config.current.tenant_id
  soft_delete_retention_days = 7
  purge_protection_enabled   = true

  sku_name = "standard"
}

resource "azurerm_key_vault_access_policy" "tenant_azcagit" {
  key_vault_id = azurerm_key_vault.tenant_kv.id
  tenant_id    = data.azurerm_client_config.current.tenant_id
  object_id    = azuread_service_principal.azcagit.object_id

  secret_permissions = [
    "Get",
    "List"
  ]
}

resource "azurerm_key_vault_access_policy" "tenant_current" {
  key_vault_id = azurerm_key_vault.tenant_kv.id
  tenant_id    = data.azurerm_client_config.current.tenant_id
  object_id    = data.azuread_client_config.current.object_id

  secret_permissions = [
    "Backup",
    "Delete",
    "Get",
    "List",
    "Purge",
    "Recover",
    "Restore",
    "Set"
  ]
}

resource "azurerm_key_vault_secret" "example_mssql_secret" {
  name         = "mssql-connection-string"
  value        = "foobar"
  key_vault_id = azurerm_key_vault.tenant_kv.id
}
