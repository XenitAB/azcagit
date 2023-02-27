resource "azurerm_resource_group" "platform" {
  name     = "rg-${local.eln}-platform"
  location = var.location
}

resource "azurerm_role_assignment" "current_platform_metrics_publisher" {
  for_each = {
    for s in ["current"] :
    s => s
    if var.add_permissions_to_current_user
  }

  scope                = azurerm_resource_group.platform.id
  role_definition_name = "Monitoring Metrics Publisher"
  principal_id         = data.azuread_client_config.current.object_id
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

# FIXME: Add zone_redundant when supported: https://github.com/hashicorp/terraform-provider-azurerm/issues/20538
resource "azurerm_container_app_environment" "this" {
  name                           = "me-${local.eln}"
  location                       = azurerm_resource_group.platform.location
  resource_group_name            = azurerm_resource_group.platform.name
  log_analytics_workspace_id     = azurerm_log_analytics_workspace.this.id
  infrastructure_subnet_id       = azurerm_subnet.this.id
  internal_load_balancer_enabled = false
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
  for_each = {
    for s in ["current"] :
    s => s
    if var.add_permissions_to_current_user
  }

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

resource "azurerm_container_app_environment_dapr_component" "azcagit_trigger" {
  name                         = "azcagit-trigger"
  container_app_environment_id = azurerm_container_app_environment.this.id
  component_type               = "pubsub.azure.servicebus"
  version                      = "v1"
  scopes                       = ["azcagit"]

  secret {
    name  = "sb-root-connectionstring"
    value = azurerm_servicebus_namespace.azcagit_trigger.default_primary_connection_string
  }

  metadata {
    name        = "connectionString"
    secret_name = "sb-root-connectionstring"
  }
}

