resource "azurerm_resource_group" "platform" {
  name     = "rg-${local.eln}-platform"
  location = var.location
}

resource "azurerm_log_analytics_workspace" "this" {
  name                = "log-${local.eln}"
  location            = azurerm_resource_group.platform.location
  resource_group_name = azurerm_resource_group.platform.name
  sku                 = "PerGB2018"
  retention_in_days   = 30
}

resource "azurerm_virtual_network" "this" {
  name                = "vnet-${local.eln}"
  location            = azurerm_resource_group.platform.location
  resource_group_name = azurerm_resource_group.platform.name
  address_space       = ["10.0.0.0/16"]
}

resource "azurerm_subnet" "this" {
  name                 = "snet-${local.eln}-ca"
  resource_group_name  = azurerm_resource_group.platform.name
  virtual_network_name = azurerm_virtual_network.this.name
  address_prefixes     = ["10.0.0.0/20"]
}

resource "azapi_resource" "managed_environment" {
  type                      = "Microsoft.App/managedEnvironments@2022-03-01"
  name                      = "me-${local.eln}"
  parent_id                 = azurerm_resource_group.platform.id
  location                  = azurerm_resource_group.platform.location
  schema_validation_enabled = false

  body = jsonencode({
    properties = {
      internalLoadBalancerEnabled = false
      appLogsConfiguration = {
        destination = "log-analytics"
        logAnalyticsConfiguration = {
          customerId = azurerm_log_analytics_workspace.this.workspace_id
          sharedKey  = azurerm_log_analytics_workspace.this.primary_shared_key
        }
      }
      vnetConfiguration = {
        infrastructureSubnetId = azurerm_subnet.this.id
        internal               = false
      },
      zoneRedundant = true
    }
  })
}

resource "azurerm_servicebus_namespace" "azcagit_trigger" {
  name                = "sb${replace(local.eln, "-", "")}${var.unique_suffix}"
  location            = azurerm_resource_group.platform.location
  resource_group_name = azurerm_resource_group.platform.name
  sku                 = "Standard"
}

resource "azuread_group" "azcagit_trigger" {
  display_name     = "aad-${local.eln}"
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

resource "azurerm_servicebus_topic" "azcagit_trigger" {
  name         = "sbt-${local.eln}-trigger"
  namespace_id = azurerm_servicebus_namespace.azcagit_trigger.id

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
        "azcagit"
      ]
    }
  })
}
