resource "azurerm_container_registry" "tenant" {
  name                = "acrtenantcontainerapps"
  resource_group_name = azurerm_resource_group.tenant.name
  location            = azurerm_resource_group.tenant.location
  sku                 = "Standard"
  admin_enabled       = true
}
