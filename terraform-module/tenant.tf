resource "azurerm_resource_group" "tenant" {
  name     = "rg-${local.eln}-tenant"
  location = var.location
}

resource "azurerm_container_registry" "tenant" {
  name                = "cr${replace(local.eln, "-", "")}${var.unique_suffix}"
  resource_group_name = azurerm_resource_group.tenant.name
  location            = azurerm_resource_group.tenant.location
  sku                 = "Standard"
  admin_enabled       = true
}

resource "azurerm_key_vault" "tenant_kv" {
  name                     = "kv${replace(local.eln, "-", "")}${var.unique_suffix}"
  location                 = azurerm_resource_group.tenant.location
  resource_group_name      = azurerm_resource_group.tenant.name
  tenant_id                = data.azurerm_client_config.current.tenant_id
  purge_protection_enabled = false

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
  for_each = {
    for s in ["current"] :
    s => s
    if var.add_permissions_to_current_user
  }

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
