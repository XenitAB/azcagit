resource "azurerm_storage_account" "this" {
  name                     = "sacontainerapps"
  resource_group_name      = azurerm_resource_group.platform.name
  location                 = azurerm_resource_group.platform.location
  account_tier             = "Standard"
  account_replication_type = "ZRS"
}

resource "azurerm_storage_container" "this" {
  name                  = "state"
  storage_account_name  = azurerm_storage_account.this.name
  container_access_type = "private"
}

resource "azapi_resource" "dapr_blob" {
  type                      = "Microsoft.App/managedEnvironments/daprComponents@2022-03-01"
  name                      = "blob"
  parent_id                 = azapi_resource.managed_environment.id
  schema_validation_enabled = false

  body = jsonencode({
    properties = {
      componentType = "state.azure.blobstorage"
      version       = "v1"
      metadata = [
        {
          name      = "accountName"
          secretRef = "account-name"
        },
        {
          name      = "accountKey"
          secretRef = "account-key"
        },
        {
          name      = "containerName"
          secretRef = "container-name"
        }
      ]
      secrets = [
        {
          name  = "account-name"
          value = azurerm_storage_account.this.name
        },
        {
          name  = "account-key"
          value = azurerm_storage_account.this.primary_access_key
        },
        {
          name  = "container-name"
          value = azurerm_storage_container.this.name
        }
      ]
    }
  })

  response_export_values = ["properties"]
}
