resource "azurerm_servicebus_namespace" "azcagit_trigger" {
  name                = "sbcontainerapps"
  location            = azurerm_resource_group.platform.location
  resource_group_name = azurerm_resource_group.platform.name
  sku                 = "Standard"
}

resource "azuread_group" "azcagit_trigger" {
  display_name     = "azcagit-trigger"
  security_enabled = true
}

resource "azuread_group_member" "azcagit_trigger" {
  group_object_id  = azuread_group.azcagit_trigger.id
  member_object_id = data.azuread_client_config.current.object_id
}

resource "azurerm_role_assignment" "azcagit_trigger" {
  scope                = azurerm_servicebus_namespace.azcagit_trigger.id
  role_definition_name = "Azure Service Bus Data Sender"
  principal_id         = azuread_group.azcagit_trigger.object_id
}

resource "azurerm_servicebus_queue" "azcagit_trigger" {
  name                = "azcagit_trigger"
  namespace_id        = azurerm_servicebus_namespace.azcagit_trigger.id
  enable_partitioning = true
}

resource "azapi_resource" "dapr_azcagit_trigger" {
  type                      = "Microsoft.App/managedEnvironments/daprComponents@2022-03-01"
  name                      = "azcagit-trigger"
  parent_id                 = azapi_resource.managed_environment.id
  schema_validation_enabled = false

  body = jsonencode({
    properties = {
      componentType = "pubsub.azure.servicebus"
      version       = "v1"
      metadata = [
        {
          name      = "connectionString"
          secretRef = "sb-root-connectionstring"
        }
      ]
      secrets = [
        {
          name  = "sb-root-connectionstring"
          value = azurerm_servicebus_namespace.azcagit_trigger.default_primary_connection_string
        }
      ]
      scopes = [
        azapi_resource.container_app_azcagit.name
      ]
    }
  })

  response_export_values = ["properties"]
}
