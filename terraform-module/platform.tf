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
  address_space       = [var.network.virtual_network_address_space]
}

resource "azurerm_subnet" "this" {
  name                 = "snet-${local.eln}-ca"
  resource_group_name  = azurerm_resource_group.platform.name
  virtual_network_name = azurerm_virtual_network.this.name
  address_prefixes     = [var.network.subnet_address_prefix]
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

resource "azurerm_storage_account" "this" {
  name                     = "sa${replace(local.eln, "-", "")}${var.unique_suffix}"
  resource_group_name      = azurerm_resource_group.platform.name
  location                 = azurerm_resource_group.platform.location
  account_tier             = "Premium"
  account_kind             = "FileStorage"
  account_replication_type = var.storage_configuration.account_replication_type
}

resource "azurerm_storage_share" "this" {
  name                 = "containerapps"
  storage_account_name = azurerm_storage_account.this.name
  quota                = var.storage_configuration.share_quota
}

resource "azurerm_container_app_environment_storage" "this" {
  name                         = "storage"
  container_app_environment_id = azurerm_container_app_environment.this.id
  account_name                 = azurerm_storage_account.this.name
  share_name                   = azurerm_storage_share.this.name
  access_key                   = azurerm_storage_account.this.primary_access_key
  access_mode                  = "ReadWrite"
}

resource "azurerm_servicebus_namespace" "azcagit_trigger" {
  name                = "sb${replace(local.eln, "-", "")}${var.unique_suffix}"
  location            = azurerm_resource_group.platform.location
  resource_group_name = azurerm_resource_group.platform.name
  sku                 = "Basic"
}

resource "azuread_group" "azcagit_trigger" {
  display_name     = "aad-${local.eln}-trigger"
  security_enabled = true
  owners           = var.aad_resource_owner_object_ids
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

resource "azurerm_role_assignment" "azcagit_trigger_receiver" {
  scope                = azurerm_servicebus_namespace.azcagit_trigger.id
  role_definition_name = "Azure Service Bus Data Receiver"
  principal_id         = azuread_service_principal.azcagit.object_id
}

resource "azurerm_servicebus_queue" "azcagit_trigger" {
  name         = "sbq-${local.eln}-trigger"
  namespace_id = azurerm_servicebus_namespace.azcagit_trigger.id

  enable_partitioning = true
}


resource "azurerm_cosmosdb_account" "this" {
  name                      = "ca-${local.eln}"
  location                  = azurerm_resource_group.platform.location
  resource_group_name       = azurerm_resource_group.platform.name
  offer_type                = "Standard"
  kind                      = "GlobalDocumentDB"
  enable_automatic_failover = false

  consistency_policy {
    consistency_level = "Session"
  }

  geo_location {
    location          = azurerm_resource_group.platform.location
    failover_priority = 0
  }

  capabilities {
    name = "EnableServerless"
  }
}

resource "azurerm_cosmosdb_sql_database" "this" {
  name                = "azcagit"
  resource_group_name = azurerm_resource_group.platform.name
  account_name        = azurerm_cosmosdb_account.this.name
}

# Cosmos DB Built-in Data Contributor (https://learn.microsoft.com/en-us/azure/cosmos-db/how-to-setup-rbac#built-in-role-definitions)
# Actions:
#   - Microsoft.DocumentDB/databaseAccounts/readMetadata
#   - Microsoft.DocumentDB/databaseAccounts/sqlDatabases/containers/*
#   - Microsoft.DocumentDB/databaseAccounts/sqlDatabases/containers/items/*
data "azurerm_cosmosdb_sql_role_definition" "data_contributor" {
  resource_group_name = azurerm_resource_group.platform.name
  account_name        = azurerm_cosmosdb_account.this.name
  role_definition_id  = "00000000-0000-0000-0000-000000000002"
}

resource "random_uuid" "azcagit_azurerm_cosmosdb_sql_role_assignment" {}

resource "azurerm_cosmosdb_sql_role_assignment" "azcagit" {
  name                = random_uuid.azcagit_azurerm_cosmosdb_sql_role_assignment.result
  resource_group_name = azurerm_resource_group.platform.name
  account_name        = azurerm_cosmosdb_account.this.name
  scope               = azurerm_cosmosdb_account.this.id
  role_definition_id  = data.azurerm_cosmosdb_sql_role_definition.data_contributor.id
  principal_id        = azuread_service_principal.azcagit.object_id
}

resource "random_uuid" "current_user_azurerm_cosmosdb_sql_role_assignment" {}

resource "azurerm_cosmosdb_sql_role_assignment" "current_user" {
  for_each = {
    for s in ["current"] :
    s => s
    if var.add_permissions_to_current_user
  }

  name                = random_uuid.current_user_azurerm_cosmosdb_sql_role_assignment.result
  resource_group_name = azurerm_resource_group.platform.name
  account_name        = azurerm_cosmosdb_account.this.name
  scope               = azurerm_cosmosdb_account.this.id
  role_definition_id  = data.azurerm_cosmosdb_sql_role_definition.data_contributor.id
  principal_id        = data.azuread_client_config.current.object_id
}

resource "azurerm_cosmosdb_sql_container" "app_cache" {
  name                  = "app-cache"
  resource_group_name   = azurerm_resource_group.platform.name
  account_name          = azurerm_cosmosdb_account.this.name
  database_name         = azurerm_cosmosdb_sql_database.this.name
  partition_key_path    = "/name"
  partition_key_version = 1
  default_ttl           = 3600

  indexing_policy {
    indexing_mode = "consistent"

    included_path {
      path = "/*"
    }
  }

  unique_key {
    paths = ["/name"]
  }
}


resource "azurerm_cosmosdb_sql_container" "job_cache" {
  name                  = "job-cache"
  resource_group_name   = azurerm_resource_group.platform.name
  account_name          = azurerm_cosmosdb_account.this.name
  database_name         = azurerm_cosmosdb_sql_database.this.name
  partition_key_path    = "/name"
  partition_key_version = 1
  default_ttl           = 3600

  indexing_policy {
    indexing_mode = "consistent"

    included_path {
      path = "/*"
    }
  }

  unique_key {
    paths = ["/name"]
  }
}

resource "azurerm_cosmosdb_sql_container" "notification_cache" {
  name                  = "notification-cache"
  resource_group_name   = azurerm_resource_group.platform.name
  account_name          = azurerm_cosmosdb_account.this.name
  database_name         = azurerm_cosmosdb_sql_database.this.name
  partition_key_path    = "/name"
  partition_key_version = 1
  default_ttl           = 3600

  indexing_policy {
    indexing_mode = "consistent"

    included_path {
      path = "/*"
    }
  }

  unique_key {
    paths = ["/name"]
  }
}
