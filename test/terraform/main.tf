terraform {
  required_providers {
    azuread = {
      version = "2.19.1"
      source  = "hashicorp/azuread"
    }
    azurerm = {
      version = "3.8.0"
      source  = "hashicorp/azurerm"
    }
    azapi = {
      version = "0.3.0"
      source  = "Azure/azapi"
    }
  }
}

provider "azapi" {
}

provider "azurerm" {
  features {}
}

data "azurerm_client_config" "current" {}

resource "azurerm_resource_group" "platform" {
  name     = "rg-aca-platform"
  location = "west europe"
}

resource "azurerm_log_analytics_workspace" "this" {
  name                = "la-container-apps"
  location            = azurerm_resource_group.platform.location
  resource_group_name = azurerm_resource_group.platform.name
  sku                 = "PerGB2018"
  retention_in_days   = 30
}

resource "azurerm_virtual_network" "this" {
  name                = "vnet-container-apps"
  location            = azurerm_resource_group.platform.location
  resource_group_name = azurerm_resource_group.platform.name
  address_space       = ["10.0.0.0/16"]
}

resource "azurerm_subnet" "this" {
  name                 = "subnet-container-apps"
  resource_group_name  = azurerm_resource_group.platform.name
  virtual_network_name = azurerm_virtual_network.this.name
  address_prefixes     = ["10.0.0.0/20"]
}

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

resource "azapi_resource" "managed_environment" {
  type                      = "Microsoft.App/managedEnvironments@2022-03-01"
  name                      = "me-container-apps"
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

  response_export_values = ["properties"]
}

# output "managed_environment" {
#   value = jsondecode(azapi_resource.managed_environment.output).properties
# }

resource "azapi_resource" "dapr_blob" {
  type      = "Microsoft.App/managedEnvironments/daprComponents@2022-03-01"
  name      = "blob"
  parent_id = azapi_resource.managed_environment.id
  #   location                  = azurerm_resource_group.platform.location
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

resource "azapi_resource" "dapr_sb" {
  type      = "Microsoft.App/managedEnvironments/daprComponents@2022-03-01"
  name      = "sb"
  parent_id = azapi_resource.managed_environment.id
  #   location                  = azurerm_resource_group.platform.location
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
          value = azurerm_servicebus_namespace.this.default_primary_connection_string
        }
      ]
      scopes = [
        azapi_resource.container_app_azcagit.name
      ]
    }
  })

  response_export_values = ["properties"]
}


resource "azurerm_container_registry" "this" {
  name                = "acrcontainerapps"
  resource_group_name = azurerm_resource_group.platform.name
  location            = azurerm_resource_group.platform.location
  sku                 = "Standard"
  admin_enabled       = true
}

resource "azurerm_servicebus_namespace" "this" {
  name                = "sbcontainerapps"
  location            = azurerm_resource_group.platform.location
  resource_group_name = azurerm_resource_group.platform.name
  sku                 = "Standard"
}

resource "azurerm_servicebus_queue" "this" {
  name                = "azcagit_trigger"
  namespace_id        = azurerm_servicebus_namespace.this.id
  enable_partitioning = true
}

resource "azurerm_resource_group" "tenant" {
  name     = "rg-aca-tenant"
  location = "west europe"
}
