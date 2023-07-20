locals {
  git_user_string = "${var.git_config.username}:${var.git_config.secret}@" == ":@" ? "" : "${var.git_config.username}:${var.git_config.secret}@"
  git_full_url    = "https://${local.git_user_string}${var.git_config.url}"
}

resource "azuread_application" "azcagit" {
  display_name = "sp-${local.eln}-azcagit"
  owners       = var.aad_resource_owner_object_ids
}

resource "azuread_service_principal" "azcagit" {
  application_id = azuread_application.azcagit.application_id
  owners         = var.aad_resource_owner_object_ids
}

resource "azuread_application_password" "azcagit" {
  application_object_id = azuread_application.azcagit.object_id
  end_date              = timeadd(timestamp(), "87600h") # 10 years

  lifecycle {
    ignore_changes = [
      end_date
    ]
  }
}

resource "azurerm_role_assignment" "azcagit_platform" {
  scope                = azurerm_resource_group.platform.id
  role_definition_name = "Contributor"
  principal_id         = azuread_service_principal.azcagit.object_id
}

resource "azurerm_role_assignment" "azcagit_platform_metrics_publisher" {
  scope                = azurerm_resource_group.platform.id
  role_definition_name = "Monitoring Metrics Publisher"
  principal_id         = azuread_service_principal.azcagit.object_id
}

resource "azurerm_role_assignment" "azcagit_tenant" {
  scope                = azurerm_resource_group.tenant.id
  role_definition_name = "Contributor"
  principal_id         = azuread_service_principal.azcagit.object_id
}

resource "azapi_resource" "azcagit_schedule" {
  schema_validation_enabled = false

  type      = "Microsoft.App/jobs@2023-04-01-preview"
  name      = "azcagit-schedule"
  location  = azurerm_resource_group.platform.location
  parent_id = azurerm_resource_group.platform.id

  body = jsonencode({
    properties = {
      configuration = {
        replicaRetryLimit = 1
        replicaTimeout    = 600
        scheduleTriggerConfig = {
          cronExpression         = "*/5 * * * *"
          parallelism            = 1
          replicaCompletionCount = 1
        }
        secrets = [
          {
            name  = "git-url"
            value = local.git_full_url
          },
          {
            name  = "container-registry-password"
            value = azurerm_container_registry.tenant.admin_password
          },
          {
            name  = "azure-tenant-id"
            value = data.azurerm_client_config.current.tenant_id
          },
          {
            name  = "azure-client-id"
            value = azuread_application.azcagit.application_id
          },
          {
            name  = "azure-client-secret"
            value = azuread_application_password.azcagit.value
          },
        ]
        triggerType = "Schedule"
      }
      environmentId = azurerm_container_app_environment.this.id
      template = {
        containers = [
          {
            name  = "azcagit"
            image = "ghcr.io/xenitab/azcagit:${var.azcagit_version}"
            args = [
              "--resource-group-name", azurerm_resource_group.tenant.name,
              "--environment", var.environment,
              "--subscription-id", data.azurerm_client_config.current.subscription_id,
              "--managed-environment-id", azurerm_container_app_environment.this.id,
              "--key-vault-name", azurerm_key_vault.tenant_kv.name,
              "--own-resource-group-name", azurerm_resource_group.platform.name,
              "--container-registry-server", azurerm_container_registry.tenant.login_server,
              "--container-registry-username", azurerm_container_registry.tenant.admin_username,
              "--location", azurerm_resource_group.tenant.location,
              "--git-branch", var.git_config.branch,
              "--git-yaml-path", var.git_config.path,
              "--notifications-enabled"
            ]
            env = [
              {
                name      = "GIT_URL"
                secretRef = "git-url"
              },
              {
                name      = "CONTAINER_REGISTRY_PASSWORD"
                secretRef = "container-registry-password"
              },
              {
                name      = "AZURE_TENANT_ID"
                secretRef = "azure-tenant-id"
              },
              {
                name      = "AZURE_CLIENT_ID"
                secretRef = "azure-client-id"
              },
              {
                name      = "AZURE_CLIENT_SECRET"
                secretRef = "azure-client-secret"
              },
            ]

            resources = {
              cpu    = "0.25"
              memory = "0.5Gi"
            }
          }
        ]
      }
    }
  })
}

resource "azapi_resource" "azcagit_event" {
  schema_validation_enabled = false

  type      = "Microsoft.App/jobs@2023-04-01-preview"
  name      = "azcagit-event"
  location  = azurerm_resource_group.platform.location
  parent_id = azurerm_resource_group.platform.id

  body = jsonencode({
    properties = {
      configuration = {
        replicaRetryLimit = 1
        replicaTimeout    = 600
        eventTriggerConfig = {
          replicaCompletionCount : 1
          parallelism : 1
          scale : {
            maxExecutions : 1
            minExecutions : 0
            pollingInterval : 5
            rules : [
              {
                name = "azure-servicebus-queue-rule"
                type = "azure-servicebus"
                metadata = {
                  messageCount : "1"
                  namespace : azurerm_servicebus_namespace.azcagit_trigger.name
                  queueName : azurerm_servicebus_queue.azcagit_trigger.name
                }
                auth = [
                  {
                    secretRef        = "service-bus-connection-string"
                    triggerParameter = "connection"
                  }
                ]
              }
            ]
          }
        }
        secrets = [
          {
            name  = "git-url"
            value = local.git_full_url
          },
          {
            name  = "container-registry-password"
            value = azurerm_container_registry.tenant.admin_password
          },
          {
            name  = "azure-tenant-id"
            value = data.azurerm_client_config.current.tenant_id
          },
          {
            name  = "azure-client-id"
            value = azuread_application.azcagit.application_id
          },
          {
            name  = "azure-client-secret"
            value = azuread_application_password.azcagit.value
          },
          {
            name  = "service-bus-connection-string"
            value = azurerm_servicebus_namespace.azcagit_trigger.default_primary_connection_string
          },
        ]
        triggerType = "Event"
      }
      environmentId = azurerm_container_app_environment.this.id
      template = {
        containers = [
          {
            name  = "azcagit"
            image = "ghcr.io/xenitab/azcagit:${var.azcagit_version}"
            args = [
              "--resource-group-name", azurerm_resource_group.tenant.name,
              "--environment", var.environment,
              "--subscription-id", data.azurerm_client_config.current.subscription_id,
              "--managed-environment-id", azurerm_container_app_environment.this.id,
              "--key-vault-name", azurerm_key_vault.tenant_kv.name,
              "--own-resource-group-name", azurerm_resource_group.platform.name,
              "--container-registry-server", azurerm_container_registry.tenant.login_server,
              "--container-registry-username", azurerm_container_registry.tenant.admin_username,
              "--location", azurerm_resource_group.tenant.location,
              "--git-branch", var.git_config.branch,
              "--git-yaml-path", var.git_config.path,
              "--notifications-enabled"
            ]
            env = [
              {
                name      = "GIT_URL"
                secretRef = "git-url"
              },
              {
                name      = "CONTAINER_REGISTRY_PASSWORD"
                secretRef = "container-registry-password"
              },
              {
                name      = "AZURE_TENANT_ID"
                secretRef = "azure-tenant-id"
              },
              {
                name      = "AZURE_CLIENT_ID"
                secretRef = "azure-client-id"
              },
              {
                name      = "AZURE_CLIENT_SECRET"
                secretRef = "azure-client-secret"
              },
            ]

            resources = {
              cpu    = "0.25"
              memory = "0.5Gi"
            }
          }
        ]
      }
    }
  })
}
